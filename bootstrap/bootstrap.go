package bootstrap

import (
	"errors"
	"github.com/integration-system/isp-lib/utils"
	"reflect"
)

const (
	LibraryVersion = "1.7.0"
)

type bootstrapConfiguration struct {
	localConfigPtr   interface{}
	localConfigType  string
	remoteConfigPtr  interface{}
	remoteConfigType string

	defaultRemoteConfigPath string

	onLocalConfigLoad     *reflect.Value
	onRemoteConfigReceive *reflect.Value
	onSocketErrorReceive  *reflect.Value
	onConfigErrorReceive  *reflect.Value
	onRoutesReceive       routesConsumer
	onLocalConfigChange   interface{}
	onShutdown            shutdownHandler

	requiredModules  map[string]*connectConsumer
	connectedModules map[string][]string

	makeSocketConfig   socketConfigProducer
	makeModuleInfo     moduleInfoProducer
	declaratorAcquirer declaratorAcquirer

	subs map[string]interface{}
}

/**
 * Add an event listener for the moment when the local config for an application loaded
 */
func (cfg *bootstrapConfiguration) OnLocalConfigLoad(f interface{}) *bootstrapConfiguration {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, cfg.localConfigType)
	cfg.onLocalConfigLoad = &rv
	return cfg
}

/**
 * Add an event listener for the moment when the local config for current application changed
 */
func (cfg *bootstrapConfiguration) OnLocalConfigChange(f interface{}) *bootstrapConfiguration {
	rt := reflect.TypeOf(f)
	assertTwoParamFunc(rt, cfg.localConfigType)
	cfg.onLocalConfigChange = f
	return cfg
}

/**
 * Add an event listener for the moment when the error message receive
 */
func (cfg *bootstrapConfiguration) OnSocketErrorReceive(f interface{}) *bootstrapConfiguration {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, reflect.TypeOf(map[string]interface{}{}).String())
	cfg.onSocketErrorReceive = &rv
	return cfg
}

/**
 * Add an event listener for the moment when the config error message receive
 */
func (cfg *bootstrapConfiguration) OnConfigErrorReceive(f interface{}) *bootstrapConfiguration {
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertSingleParamFunc(rt, "string")
	cfg.onConfigErrorReceive = &rv
	return cfg
}

/**
 * Add an event listener for the moment when an application received its configuration
 */
func (cfg *bootstrapConfiguration) OnRemoteConfigReceive(f interface{}) *bootstrapConfiguration {
	if cfg.remoteConfigType == "" {
		panic(errors.New("remote config type is undefined"))
		return nil
	}
	rv, rt := reflect.ValueOf(f), reflect.TypeOf(f)
	assertTwoParamFunc(rt, cfg.remoteConfigType)
	cfg.onRemoteConfigReceive = &rv
	return cfg
}

/**
 * Add a hook for executing a code when an application is ready to be ended
 */
func (cfg *bootstrapConfiguration) OnShutdown(f shutdownHandler) *bootstrapConfiguration {
	cfg.onShutdown = f
	return cfg
}

// subscribe to web-socket event
func (cfg *bootstrapConfiguration) OnSocketEvent(event string, f interface{}) *bootstrapConfiguration {
	cfg.subs[event] = f
	return cfg
}

/**
 * Specify the socket builder function that creates a socket configuration
 */
func (cfg *bootstrapConfiguration) SocketConfiguration(f socketConfigProducer) *bootstrapConfiguration {
	cfg.makeSocketConfig = f
	return cfg
}

// set callback function which receive module declarator on startup
func (cfg *bootstrapConfiguration) AcquireDeclarator(f declaratorAcquirer) *bootstrapConfiguration {
	cfg.declaratorAcquirer = f
	return cfg
}

// provides callback function which return base module information
func (cfg *bootstrapConfiguration) DeclareMe(f moduleInfoProducer) *bootstrapConfiguration {
	cfg.makeModuleInfo = f
	return cfg
}

// module is in not ready state until received routes from config-service
func (cfg *bootstrapConfiguration) RequireRoutes(f routesConsumer) *bootstrapConfiguration {
	cfg.onRoutesReceive = f
	return cfg
}

// module is in not ready state until establish grpc connection with required modules
func (cfg *bootstrapConfiguration) RequireModule(moduleName string, consumer addressListConsumer, mustConnect bool) *bootstrapConfiguration {
	cfg.requiredModules[utils.ModuleConnected(moduleName)] = &connectConsumer{mustConnect: mustConnect, consumer: consumer}
	return cfg
}

// add path to remote config module
func (cfg *bootstrapConfiguration) DefaultRemoteConfigPath(path string) *bootstrapConfiguration {
	cfg.defaultRemoteConfigPath = path
	return cfg
}

// starts module, block until interruption
func (cfg *bootstrapConfiguration) Run() {
	makeRunner(*cfg).run()
}

// entry point to describe module
func ServiceBootstrap(localConfigPtr, remoteConfigPtr interface{}) *bootstrapConfiguration {
	if localConfigPtr == nil || reflect.TypeOf(localConfigPtr).Kind() != reflect.Ptr {
		panic(errors.New("expecting not nil pointer to local config struct"))
		return nil
	}
	if remoteConfigPtr != nil && reflect.TypeOf(remoteConfigPtr).Kind() != reflect.Ptr {
		panic(errors.New("expecting not nil pointer to remote config struct"))
		return nil
	}
	b := &bootstrapConfiguration{
		localConfigPtr:   localConfigPtr,
		localConfigType:  reflect.TypeOf(localConfigPtr).String(),
		requiredModules:  make(map[string]*connectConsumer),
		subs:             make(map[string]interface{}),
		connectedModules: make(map[string][]string),
	}
	if remoteConfigPtr != nil {
		b.remoteConfigPtr = remoteConfigPtr
		b.remoteConfigType = reflect.TypeOf(remoteConfigPtr).String()
	}

	return b
}
