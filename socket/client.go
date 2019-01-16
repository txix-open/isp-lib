package socket

import (
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/logger"
	"strconv"
	"time"
)

type SocketConfiguration struct {
	Host             string
	Port             string
	Path             string
	Secure           bool
	UrlParams        map[string]string
	ConnectionString string
}

func (sc *SocketConfiguration) GetConnectionString() string {
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

var socketClient *gosocketio.Client

func GetClient() *gosocketio.Client {
	if socketClient == nil || !socketClient.IsAlive() {
		errorMessage := "SocketIO client isn't alive"
		logger.Fatalf(errorMessage)
	}
	return socketClient
}

func InitClient(socketConfig SocketConfiguration, subscriptions func(client *gosocketio.Client)) *gosocketio.Client {
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
	connectionString := socketConfig.GetConnectionString()
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
