package bootstrap

import (
	"encoding/json"
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"reflect"
)

func handleRemoteConfiguration(remoteConfigChan chan string, event string) func(c *gosocketio.Channel, data string) error {
	return func(c *gosocketio.Channel, data string) error {
		logger.Infof("--- Got event: %s message: %s", event, data)
		remoteConfigChan <- data
		return nil
	}
}

func handleError(onSocketErrorReceive *reflect.Value, event string) func(c *gosocketio.Channel, args map[string]interface{}) error {
	return func(c *gosocketio.Channel, args map[string]interface{}) error {
		logger.Infof("--- Got event: %s message: %s", event, args)
		callFunc(onSocketErrorReceive, args)
		return nil
	}
}

func handleConfigError(onConfigErrorReceive *reflect.Value, event string) func(c *gosocketio.Channel, args string) error {
	return func(c *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s message: %s", event, args)
		callFunc(onConfigErrorReceive, args)
		return nil
	}
}

func handleRoutes(routesChan chan structure.RoutingConfig, event string) func(c *gosocketio.Channel, args string) error {
	return func(c *gosocketio.Channel, args string) error {
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
			routesChan <- routes
			return nil
		} else {
			logger.Warn("Received invalid route configuration", err)
			return err
		}
	}
}
