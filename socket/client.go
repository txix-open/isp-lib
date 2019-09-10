package socket

import (
	"errors"
	"fmt"
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"strconv"
	"strings"
	"time"
)

var socketClient *gosocketio.Client

func GetClient() *gosocketio.Client {
	if socketClient == nil || !socketClient.IsAlive() {
		errorMessage := "SocketIO client isn't alive"
		logger.Fatalf(errorMessage)
	}
	return socketClient
}

func InitClient(socketConfig structure.SocketConfiguration, subscriptions func(client *gosocketio.Client)) *gosocketio.Client {
	builder := gosocketio.NewClientBuilder().
		EnableReconnection().
		ReconnectionTimeout(3*time.Second).
		OnReconnectionError(func(err error) {
			logger.Warnf("SocketIO reconnection error: %v", err)
		}).
		On(gosocketio.OnDisconnection, func(arg interface{}) error {
			logger.Warn("SocketIO disconnected")
			return nil
		}, nil)

	if subscriptions != nil {
		subscriptions(builder.UnsafeClient())
	}
	connectionString := GetConnectionString(socketConfig)
	client := builder.BuildToConnect(connectionString)

	err := client.Dial()
	for err != nil {
		time.Sleep(3 * time.Second)
		err = client.Dial()
		logger.Warnf("Could not connect to SocketIO: %v", err)
	}
	socketClient = client

	return socketClient
}

func GetConnectionString(sc structure.SocketConfiguration) string {
	connectionString := sc.ConnectionString
	port, _ := strconv.Atoi(sc.Port)
	if connectionString == "" {
		connectionString = gosocketio.GetUrl(
			sc.Host,
			port,
			sc.Secure,
			sc.UrlParams,
		)
	}
	return connectionString
}

func GetConnectionStrings(sc structure.SocketConfiguration) ([]string, error) {
	hosts := strings.Split(sc.Host, ";")
	ports := strings.Split(sc.Port, ";")
	if len(hosts) != len(ports) {
		return nil, errors.New(fmt.Sprintf("Different number of hosts/ports: %n/%n", len(hosts), len(ports)))
	}
	connStrings := make([]string, len(hosts))
	for i := 0; i < len(hosts); i++ {
		port, err := strconv.Atoi(ports[i])
		if err != nil {
			return nil, err
		}
		connectionString := gosocketio.GetUrl(
			hosts[i],
			port,
			sc.Secure,
			sc.UrlParams,
		)
		connStrings[i] = connectionString
	}
	return connStrings, nil
}
