package utils

import (
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
	ConfigLogRotation             = "CONFIG:LOG_ROTATION"
	ConfigRestart                 = "CONFIG:MODULE_RESTART"

	ModuleReady            = "MODULE:READY"
	ModuleSendRequirements = "MODULE:SEND_REQUIREMENTS"
	ModuleUpdateRoutes     = "MODULE:UPDATE_ROUTES"
	ModuleSendConfigSchema = "MODULE:SEND_CONFIG_SCHEMA"

	ModuleConnectionSuffix = "MODULE_CONNECTED"

	ModuleNameGetParamKey = "module_name"
	// Deprecated: no longer used.
	InstanceUuidGetParamKey = "instance_uuid"
	WsOkResponse            = "ok"
)

func ModuleConnected(moduleName string) string {
	return fmt.Sprintf("%s_%s", moduleName, ModuleConnectionSuffix)
}

// ParseParameters always return empty instanceUUID as it was deprecated.
func ParseParameters(queryRaw string) (instanceUUID string, moduleName string, error error) {
	parsedParams, _ := url.ParseQuery(queryRaw)
	moduleName = parsedParams.Get(ModuleNameGetParamKey)
	if moduleName == "" {
		err := fmt.Errorf("not received all get parameters, %s: %s",
			ModuleNameGetParamKey,
			moduleName,
		)
		return "", "", err
	}
	return "", moduleName, nil
}
