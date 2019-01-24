package utils

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	ErrorConnection = "ERROR_CONNECTION"
	ConfigError     = "ERROR_CONFIG"

	ConfigSendConfigWhenConnected = "CONFIG:SEND_CONFIG_WHEN_CONNECTED"
	ConfigSendConfigChanged       = "CONFIG:SEND_CONFIG_CHANGED"
	ConfigSendConfigOnRequest     = "CONFIG:SEND_CONFIG_ON_REQUEST"

	ConfigSendRoutesWhenConnected = "CONFIG:SEND_ROUTES_WHEN_CONNECTED"
	ConfigSendRoutesChanged       = "CONFIG:SEND_ROUTES_CHANGED"
	ConfigSendRoutesOnRequest     = "CONFIG:SEND_ROUTES_ON_REQUEST"

	ModuleReady            = "MODULE:READY"
	ModuleSendRequirements = "MODULE:SEND_REQUIREMENTS"
	ModuleUpdateRoutes     = "MODULE:UPDATE_ROUTES"
	ModuleSendConfigSchema = "MODULE:SEND_CONFIG_SCHEMA"

	ModuleConnectionSuffix = "MODULE_CONNECTED"
)

func ModuleConnected(moduleName string) string {
	return fmt.Sprintf("%s_%s", moduleName, ModuleConnectionSuffix)
}

func ParseParameters(queryRaw string) (instanceUUID string, moduleName string, error error) {
	parsedParams, _ := url.ParseQuery(queryRaw)
	moduleName = parsedParams.Get("module_name")
	instanceUuid := parsedParams.Get("instance_uuid")
	if moduleName == "" || instanceUuid == "" || !IsValidUUID(instanceUuid) {
		err := fmt.Sprintf("SocketIO not received all parameters, module_name: %s, instance_uuid: %s",
			moduleName,
			instanceUuid,
		)
		return "", "", errors.New(err)
	}
	return instanceUuid, moduleName, nil
}
