package bootstrap

import (
	"encoding/json"
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"reflect"
)

func UnmarshalAddressListAndThen(event string, f func([]structure.AddressConfiguration)) func(*gosocketio.Channel, string) error {
	return func(_ *gosocketio.Channel, data string) error {
		list := make([]structure.AddressConfiguration, 0)
		if err := json.Unmarshal([]byte(data), &list); err != nil {
			log.WithMetadata(log.Metadata{"event": event, "data": data}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid address list: %v", err)
			return err
		} else {
			log.WithMetadata(log.Metadata{"event": event, "data": data}).
				Info(stdcodes.ConfigServiceReceiveRequiredModuleAddress, "received required module address list")
			f(list)
		}
		return nil
	}
}

func handleRemoteConfiguration(remoteConfigChan chan string, event string) func(c *gosocketio.Channel, data string) error {
	return func(c *gosocketio.Channel, data string) error {
		log.WithMetadata(log.Metadata{"config": data}).
			Info(stdcodes.ConfigServiceReceiveConfiguration, "received remote config")
		remoteConfigChan <- data
		return nil
	}
}

func handleError(onSocketErrorReceive *reflect.Value, event string) func(c *gosocketio.Channel, args map[string]interface{}) error {
	return func(c *gosocketio.Channel, args map[string]interface{}) error {
		callFunc(onSocketErrorReceive, args)
		return nil
	}
}

func handleConfigError(onConfigErrorReceive *reflect.Value, event string) func(c *gosocketio.Channel, args string) error {
	return func(c *gosocketio.Channel, args string) error {
		callFunc(onConfigErrorReceive, args)
		return nil
	}
}

func handleRoutes(routesChan chan structure.RoutingConfig, event string) func(c *gosocketio.Channel, args string) error {
	return func(c *gosocketio.Channel, data string) error {
		routes := structure.RoutingConfig{}
		err := json.Unmarshal([]byte(data), &routes)
		if err != nil {
			log.WithMetadata(log.Metadata{"event": event, "data": data}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid routes list: %v", err)
			return err
		}

		if err := utils.Validate(routes); err == nil {
			totalModules := len(routes)
			totalEndpoints := 0
			for _, v := range routes {
				totalEndpoints += len(v.Endpoints)
			}
			log.WithMetadata(log.Metadata{"total_modules": totalModules, "total_endpoints": totalEndpoints}).
				Info(stdcodes.ConfigServiceReceiveRoutes, "received routes")
			routesChan <- routes
			return nil
		} else {
			log.WithMetadata(log.Metadata{"event": event, "data": data}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid routes list: %v", err)
			return err
		}
	}
}
