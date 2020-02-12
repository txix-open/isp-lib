package bootstrap

import (
	"encoding/json"
	"reflect"

	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
)

func UnmarshalAddressListAndThen(event string, f func([]structure.AddressConfiguration)) func([]byte) {
	return func(data []byte) {
		list := make([]structure.AddressConfiguration, 0)
		if err := json.Unmarshal(data, &list); err != nil {
			log.WithMetadata(log.Metadata{"event": event, "data": string(data)}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid address list: %v", err)
		} else {
			log.WithMetadata(log.Metadata{"event": event, "data": string(data)}).
				Info(stdcodes.ConfigServiceReceiveRequiredModuleAddress, "received required module address list")
			f(list)
		}
	}
}

func handleRemoteConfiguration(remoteConfigChan chan []byte, event string) func([]byte) {
	return func(data []byte) {
		log.WithMetadata(log.Metadata{"config": string(data)}).
			Info(stdcodes.ConfigServiceReceiveConfiguration, "received remote config")
		remoteConfigChan <- data
	}
}

func handleError(onSocketErrorReceive *reflect.Value, event string) func([]byte) {
	return func(data []byte) {
		var args map[string]interface{}
		_ = json.Unmarshal(data, &args)
		callFunc(onSocketErrorReceive, args)
	}
}

func handleConfigError(onConfigErrorReceive *reflect.Value, event string) func([]byte) {
	return func(data []byte) {
		callFunc(onConfigErrorReceive, string(data))
	}
}

func handleRoutes(routesChan chan structure.RoutingConfig, event string) func([]byte) {
	return func(data []byte) {
		routes := structure.RoutingConfig{}
		err := json.Unmarshal(data, &routes)
		if err != nil {
			log.WithMetadata(log.Metadata{"event": event, "data": string(data)}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid routes list: %v", err)
			return
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
		} else {
			log.WithMetadata(log.Metadata{"event": event, "data": string(data)}).
				Errorf(stdcodes.ConfigServiceInvalidDataReceived, "received invalid routes list: %v", err)
		}
	}
}
