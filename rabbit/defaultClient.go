package rabbit

import (
	"github.com/integration-system/isp-lib/logger"
	"sync"
	"time"
)

var (
	client *Client
	lock   sync.Mutex
)

func CloseDefaultClient() {
	lock.Lock()
	defer lock.Unlock()

	if client != nil {
		client.Close(func(err error) {
			logger.Warnf("Close rabbit client err: %v", err)
		})
		client = nil
	}
}

func IsInitialized() bool {
	return client != nil
}

func GetDefaultClient() *Client {
	if !IsInitialized() {
		logger.Fatal("Rabbit client has not initialized")
	}

	return client
}

func InitDefaultClient(cfg RabbitConfig) *Client {
	lock.Lock()
	defer lock.Unlock()

	connected := false
	for !connected {
		c, err := MakeClient(cfg)
		if err != nil {
			logger.Warnf("Could not connect to rabbit: %v", err)
			time.Sleep(cfg.reconnectionTimeout())
		} else {
			client = c
			connected = true
		}
	}

	return client
}
