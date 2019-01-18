package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/metric"
	"github.com/integration-system/isp-lib/socket"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"github.com/mohae/deepcopy"
	"github.com/thecodeteam/goodbye"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	LibraryVersion = "0.7.0"
)

type RoutesDeclarator interface {
	DeclareRoutes()
}

type socketConfigProducer func(localConfigPtr interface{}) socket.SocketConfiguration

type shutdownHandler func(ctx context.Context, sig os.Signal)

type SocketConsumer func(sockClient *gosocketio.Client)

type moduleInfoProducer func(localConfigPtr interface{}) ModuleInfo

type addressListConsumer func(list []structure.AddressConfiguration) bool

type routesConsumer func(routes structure.RoutingConfig) bool

type declaratorAcquirer func(dec RoutesDeclarator)

type declarator struct {
	f func(eventType string)
}

func (d *declarator) DeclareRoutes() {
	if d != nil {
		d.f(utils.ModuleUpdateRoutes)
	}
}

type connectEvent struct {
	event       string
	addressList []structure.AddressConfiguration
}

type connectConsumer struct {
	consumer    addressListConsumer
	mustConnect bool
}

type ModuleInfo struct {
	ModuleName       string
	ModuleVersion    string
	GrpcOuterAddress structure.AddressConfiguration
	Handlers         []interface{}
}

type ModuleRequirements struct {
	RequiredModules []string
	RequireRoutes   bool
}

func (r ModuleRequirements) IsEmpty() bool {
	return len(r.RequiredModules) == 0 && !r.RequireRoutes
}

type bootstrap2 struct {
	info ModuleInfo

	localConfigPtr   interface{}
	localConfigType  string
	remoteConfigPtr  interface{}
	remoteConfigType string

	onLocalConfigLoad     *reflect.Value
	onRemoteConfigReceive *reflect.Value
	onSocketErrorReceive  *reflect.Value
	onConfigErrorReceive  *reflect.Value
	onRoutesReceive       routesConsumer
	onLocalConfigChange   interface{}
	onShutdown            shutdownHandler

	requiredModules  map[string]*connectConsumer
	connectedModules map[string][]string

	socketCfgProducer  socketConfigProducer
	moduleInfoProducer moduleInfoProducer
	declaratorAcquirer declaratorAcquirer

	remoteConfigChan chan interface{}
	routesChan       chan structure.RoutingConfig
	connectEventChan chan connectEvent
	exitChan         chan struct{}
	disconnectChan   chan struct{}

	subs                     map[string]interface{}
	client                   *gosocketio.Client
	ready                    bool
	lastFailedConnectionTime time.Time
}

/**
 * Add an event listener for the moment when the local config for an application loaded
 */
func (b *bootstrap2) OnLocalConfigLoad(f interface{}) *bootstrap2 {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, b.localConfigType)
	b.onLocalConfigLoad = &rv
	return b
}

/**
 * Add an event listener for the moment when the local config for current application changed
 */
func (b *bootstrap2) OnLocalConfigChange(f interface{}) *bootstrap2 {
	rt := reflect.TypeOf(f)
	assertTwoParamFunc(rt, b.localConfigType)
	b.onLocalConfigChange = f
	return b
}

/**
 * Add an event listener for the moment when the error message receive
 */
func (b *bootstrap2) OnSocketErrorReceive(f interface{}) *bootstrap2 {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, reflect.TypeOf(map[string]interface{}{}).String())
	b.onSocketErrorReceive = &rv
	return b
}

/**
 * Add an event listener for the moment when the config error message receive
 */
func (b *bootstrap2) OnConfigErrorReceive(f interface{}) *bootstrap2 {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, "string")
	b.onConfigErrorReceive = &rv
	return b
}

/**
 * Add an event listener for the moment when an application received its configuration
 */
func (b *bootstrap2) OnRemoteConfigReceive(f interface{}) *bootstrap2 {
	if b.remoteConfigType == "" {
		logger.Fatal("Remote config type is undefined.")
		return nil
	}
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertTwoParamFunc(rt, b.remoteConfigType)
	b.onRemoteConfigReceive = &rv
	return b
}

/**
 * Add a hook for executing a code when an application is ready to be ended
 */
func (b *bootstrap2) OnShutdown(f shutdownHandler) *bootstrap2 {
	b.onShutdown = f
	return b
}

func (b *bootstrap2) OnSocketEvent(event string, f interface{}) *bootstrap2 {
	b.subs[event] = f
	return b
}

/**
 * Specify the socket builder function that creates a socket configuration
 */
func (b *bootstrap2) SocketConfiguration(f socketConfigProducer) *bootstrap2 {
	b.socketCfgProducer = f
	return b
}

func (b *bootstrap2) AcquireDeclarator(f declaratorAcquirer) *bootstrap2 {
	b.declaratorAcquirer = f
	return b
}

func (b *bootstrap2) DeclareMe(f moduleInfoProducer) *bootstrap2 {
	b.moduleInfoProducer = f
	return b
}

func (b *bootstrap2) RequireRoutes(f routesConsumer) *bootstrap2 {
	b.onRoutesReceive = f
	return b
}

func (b *bootstrap2) RequireModule(moduleName string, consumer addressListConsumer, mustConnect bool) *bootstrap2 {
	b.requiredModules[utils.ModuleConnected(moduleName)] = &connectConsumer{mustConnect: mustConnect, consumer: consumer}
	return b
}

func (b *bootstrap2) Run() {
	ctx := b.initShutdownHandler()
	defer goodbye.Exit(ctx, 0)

	defer func() {
		err := recover()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	b.initLocalConfig()
	b.initModuleInfo()
	b.initSocketConnection()
	b.initStatusMetrics()

	if b.declaratorAcquirer != nil {
		b.declaratorAcquirer(&declarator{b.sendDeclaration})
	}

	b.sendRemoteConfigSchema()

	b.ready = false

	remoteConfigReady, requiredModulesReady, routesReady, currentConnectedModules := b.initialState()
	remoteConfigTimeoutChan := time.After(3 * time.Second)
	neverTriggerChan := make(chan time.Time)
	initChan := make(chan struct{}, 1)
	for {
		if !b.ready && remoteConfigReady && requiredModulesReady && routesReady {
			initChan <- struct{}{}
		}

		select {
		case newRemoteConfig := <-b.remoteConfigChan:
			old := b.remoteConfigPtr
			if b.onRemoteConfigReceive != nil {
				b.callOnRemoteConfigReceive(newRemoteConfig, old)
			}
			remoteConfigReady = true
			b.remoteConfigPtr = newRemoteConfig
			if !b.ready {
				b.sendRequirements()
			}
			remoteConfigTimeoutChan = neverTriggerChan
		case <-remoteConfigTimeoutChan:
			logger.Warn("Remote config isn't received")
			remoteConfigTimeoutChan = time.After(3 * time.Second)
		case routers := <-b.routesChan:
			if b.onRoutesReceive != nil {
				routesReady = b.onRoutesReceive(routers)
			}
		case e := <-b.connectEventChan:
			if c, ok := b.requiredModules[e.event]; ok {
				if ok := c.consumer(e.addressList); ok {
					currentConnectedModules[e.event] = true
				}

				ok := true
				for e, consumer := range b.requiredModules {
					val := currentConnectedModules[e]
					if !val && consumer.mustConnect {
						ok = false
						break
					}
				}
				requiredModulesReady = ok

				addrList := make([]string, 0, len(e.addressList))
				if currentConnectedModules[e.event] {
					for _, addr := range e.addressList {
						addrList = append(addrList, addr.GetAddress())
					}
				}
				b.connectedModules[e.event] = addrList
			}
		case <-initChan:
			b.ready = true
			b.sendModuleReady()
		case <-b.disconnectChan:
			b.ready = false
			remoteConfigReady, requiredModulesReady, routesReady, currentConnectedModules = b.initialState()
		case <-b.exitChan:
			return
		}

	}
}

func (b *bootstrap2) initShutdownHandler() context.Context {
	ctx := context.Background()

	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		logger.Info(logger.FmtAlertMsg("module shutting down now"))

		if b.client != nil {
			b.client.Close()
		}

		if b.onShutdown != nil {
			b.onShutdown(ctx, sig)
		}

		logger.Info(logger.FmtAlertMsg("module has gracefully shut down"))

		close(b.exitChan)
	})

	return ctx
}

func (b *bootstrap2) initLocalConfig() {
	if b.onLocalConfigChange != nil {
		config.OnConfigChange(b.onLocalConfigChange)
	}
	b.localConfigPtr = config.InitConfigV2(b.localConfigPtr, false)
	if b.onLocalConfigLoad != nil {
		b.onLocalConfigLoad.Call([]reflect.Value{reflect.ValueOf(b.localConfigPtr)})
	}
}

func (b *bootstrap2) initModuleInfo() {
	b.info = b.moduleInfoProducer(config.Get())
}

func (b *bootstrap2) initSocketConnection() {
	if b.socketCfgProducer == nil {
		logger.Fatal("Socket configuration is not specified. Call 'SocketConfiguration' first")
		return
	}

	socketConfig := b.socketCfgProducer(b.localConfigPtr)
	builder := gosocketio.NewClientBuilder().
		EnableReconnection().
		ReconnectionTimeout(3*time.Second).
		OnReconnectionError(func(err error) {
			logger.Warnf("SocketIO reconnection error: %v", err)
			b.lastFailedConnectionTime = time.Now()
		}).
		On(gosocketio.OnDisconnection, func(arg interface{}) error {
			logger.Warn("SocketIO disconnected")
			b.lastFailedConnectionTime = time.Now()
			b.disconnectChan <- struct{}{}
			return nil
		}, nil)
	connectionString := socketConfig.GetConnectionString()
	b.client = builder.BuildToConnect(connectionString)

	if b.onSocketErrorReceive != nil {
		b.subscribeToErrorReceive()
	}
	if b.onConfigErrorReceive != nil {
		b.subscribeToConfigErrorReceive()
	}
	if b.remoteConfigPtr != nil {
		b.subscribeToReceiveRemoteConfig(utils.ConfigSendConfigWhenConnected)
		b.subscribeToReceiveRemoteConfig(utils.ConfigSendConfigChanged)
		b.subscribeToReceiveRemoteConfig(utils.ConfigSendConfigOnRequest)
	}
	if b.onRoutesReceive != nil {
		b.subscribeToRoutesReceive(utils.ConfigSendRoutesChanged)
		b.subscribeToRoutesReceive(utils.ConfigSendRoutesWhenConnected)
		b.subscribeToRoutesReceive(utils.ConfigSendRoutesOnRequest)
	}
	for e := range b.requiredModules {
		must(b.client.On(e, UnmarshalAddressListAndThen(e, makeAddressListConsumer(e, b.connectEventChan))))
	}
	for e, f := range b.subs {
		evt := e
		must(b.client.On(evt, f))
	}

	err := b.client.Dial()
	for err != nil {
		logger.Warnf("Could not connect to SocketIO: %v", err)
		b.lastFailedConnectionTime = time.Now()

		select {
		case <-b.exitChan:
			return
		case <-time.After(3 * time.Second):

		}
		err = b.client.Dial()
	}
}

func (b *bootstrap2) initStatusMetrics() {
	metric.InitStatusChecker("config-websocket", func() interface{} {
		socketConfig := b.socketCfgProducer(b.localConfigPtr)
		uri := fmt.Sprintf("%s:%s", socketConfig.Host, socketConfig.Port)
		status := true
		if b.client == nil || !b.client.IsAlive() {
			status = false
		}
		lastFailedConnectionMsAgo := time.Duration(0)
		if !b.lastFailedConnectionTime.IsZero() {
			lastFailedConnectionMsAgo = time.Now().Sub(b.lastFailedConnectionTime) / 1e6
		}
		return map[string]interface{}{
			"connected":                 status,
			"lastFailedConnectionMsAgo": lastFailedConnectionMsAgo,
			"address":                   uri,
			"moduleReady":               b.ready,
		}
	})

	for k := range b.requiredModules {
		moduleName := strings.Replace(k, "_"+utils.ModuleConnectionSuffix, "", -1)
		keyCopy := k
		metric.InitStatusChecker(fmt.Sprintf("%s-grpc", moduleName), func() interface{} {
			addrList, ok := b.connectedModules[keyCopy]
			if ok {
				return addrList
			} else {
				return []string{}
			}
		})
	}
}

func (b *bootstrap2) subscribeToReceiveRemoteConfig(event string) {
	must(b.client.On(event, func(c *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s message: %s", event, args)

		oldConfig := deepcopy.Copy(b.remoteConfigPtr)
		newRemoteConfig := config.InitRemoteConfig(oldConfig, args)
		b.remoteConfigChan <- newRemoteConfig
		return nil
	}))
}

func (b *bootstrap2) subscribeToErrorReceive() {
	must(b.client.On(utils.ErrorConnection, func(c *gosocketio.Channel, args map[string]interface{}) error {
		logger.Infof("--- Got event: %s message: %s", utils.ErrorConnection, args)

		b.onSocketErrorReceive.Call([]reflect.Value{reflect.ValueOf(args)})
		return nil
	}))
}

func (b *bootstrap2) subscribeToConfigErrorReceive() {
	must(b.client.On(utils.ConfigError, func(c *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s message: %s", utils.ConfigError, args)

		b.onConfigErrorReceive.Call([]reflect.Value{reflect.ValueOf(args)})
		return nil
	}))
}

func (b *bootstrap2) subscribeToRoutesReceive(event string) {
	must(b.client.On(event, func(c *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s", event)

		routes := structure.RoutingConfig{}
		err := json.Unmarshal([]byte(args), &routes)
		if err != nil {
			logger.Warnf("Received invalid json payload, %s", err)
			return err
		}

		if err := utils.Validate(routes); err == nil {
			logger.Debugf("Routes received: %s", args)
			for _, v := range routes {
				logger.Infof("Routes received: %d, module: %s, version: %s, address: %s",
					len(v.Endpoints),
					v.ModuleName,
					v.Version,
					v.Address.GetAddress(),
				)
			}
			b.routesChan <- routes
			return nil
		} else {
			logger.Warn("Received invalid route configuration", err)
			return err
		}
	}))
}

func (b *bootstrap2) sendRequirements() {
	requiredModules := make([]string, 0, len(b.requiredModules))
	for evt := range b.requiredModules {
		requiredModules = append(requiredModules, evt)
	}

	requirements := ModuleRequirements{
		RequiredModules: requiredModules,
		RequireRoutes:   b.onRoutesReceive != nil,
	}

	if !requirements.IsEmpty() {
		logger.Infof("%s: %v", utils.ModuleSendRequirements, requirements)
		if s, err := json.Marshal(requirements); err != nil {
			logger.Warn("Could not serialize requirements to JSON", err)
		} else if err := b.client.Emit(utils.ModuleSendRequirements, string(s)); err != nil {
			logger.Warn("Could not send requirements", err)
		}
	}
}

func (b *bootstrap2) sendDeclaration(eventType string) {
	bytes, err := b.getSerializedDeclaration()
	if err != nil {
		logger.Warn("Could not serialize declaration to JSON", err)
		return
	}

	logger.Debugf("MODULE_DECLARATION: %s", string(bytes))
	logger.Info(eventType)
	if err := b.client.Emit(eventType, string(bytes)); err != nil {
		logger.Warn("Could not send declaration", err)
	}
}

func (b *bootstrap2) sendRemoteConfigSchema() {
	s := schema.GenerateConfigSchema(b.remoteConfigPtr)
	req := schema.ConfigSchema{Version: b.info.ModuleVersion, Schema: s}
	if bytes, err := json.Marshal(req); err != nil {
		logger.Error("Could not serialize config schema to JSON", err)
	} else if err := b.client.Emit(utils.ModuleSendConfigSchema, string(bytes)); err != nil {
		logger.Error("Could not send config schema", err)
	}
}

func (b *bootstrap2) sendModuleReady() {
	b.sendDeclaration(utils.ModuleReady)
}

func (b *bootstrap2) callOnRemoteConfigReceive(newRemoteConfig, oldRemoteConfig interface{}) {
	oldCfg := reflect.ValueOf(oldRemoteConfig)
	newCfg := reflect.ValueOf(newRemoteConfig)
	b.onRemoteConfigReceive.Call([]reflect.Value{newCfg, oldCfg})
}

func (b *bootstrap2) initialState() (remoteConfigReady, requiredModulesReady, routesReady bool, currentConnectedModules map[string]bool) {
	remoteConfigReady = false
	currentConnectedModules = make(map[string]bool)
	for evt, c := range b.requiredModules {
		if !c.mustConnect {
			currentConnectedModules[evt] = true
		}
	}
	requiredModulesReady = len(b.requiredModules) == len(currentConnectedModules)
	routesReady = b.onRoutesReceive == nil
	return
}

func (b *bootstrap2) getSerializedDeclaration() ([]byte, error) {
	moduleInfo := b.moduleInfoProducer(config.Get())
	b.info = moduleInfo

	endpoints := backend.GetEndpoints(moduleInfo.ModuleName, moduleInfo.Handlers...)
	addr := moduleInfo.GrpcOuterAddress.IP
	hasSchema := strings.Contains(addr, "http://")
	if hasSchema {
		addr = strings.Replace(addr, "http://", "", -1)
	}
	if addr == "" {
		ip, err := getOutboundIp()
		if err != nil {
			logger.Warn(err)
		} else {
			if hasSchema {
				ip = fmt.Sprintf("http://%s", ip)
			}
			moduleInfo.GrpcOuterAddress.IP = ip
		}
	}
	declaration := structure.BackendDeclaration{
		ModuleName: moduleInfo.ModuleName,
		Version:    moduleInfo.ModuleVersion,
		Address:    moduleInfo.GrpcOuterAddress,
		LibVersion: LibraryVersion,
		Endpoints:  endpoints,
	}

	return json.Marshal(declaration)
}

func ServiceBootstrap(localConfigPtr, remoteConfigPtr interface{}) *bootstrap2 {
	if localConfigPtr == nil || reflect.TypeOf(localConfigPtr).Kind() != reflect.Ptr {
		logger.Fatal("Expecting not nil pointer to local config struct")
		return nil
	}
	if remoteConfigPtr != nil && reflect.TypeOf(remoteConfigPtr).Kind() != reflect.Ptr {
		logger.Fatal("Expecting not nil pointer to remote config struct")
		return nil
	}
	b := &bootstrap2{
		localConfigPtr:   localConfigPtr,
		localConfigType:  reflect.TypeOf(localConfigPtr).String(),
		requiredModules:  make(map[string]*connectConsumer),
		subs:             make(map[string]interface{}),
		connectedModules: make(map[string][]string),
		remoteConfigChan: make(chan interface{}),
		connectEventChan: make(chan connectEvent),
		exitChan:         make(chan struct{}),
		routesChan:       make(chan structure.RoutingConfig),
		disconnectChan:   make(chan struct{}),
	}
	if remoteConfigPtr != nil {
		b.remoteConfigPtr = remoteConfigPtr
		b.remoteConfigType = reflect.TypeOf(remoteConfigPtr).String()
	}

	return b
}
