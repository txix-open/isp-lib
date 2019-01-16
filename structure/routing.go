package structure

import (
	"bytes"
	"encoding/json"
)

type AddressConfiguration struct {
	Port string `json:"port"`
	IP   string `json:"ip"`
}

type ModuleInfo struct {
	ModuleName  string   `json:"moduleName"`
	Version     string   `json:"version"`
	LibVersion  string   `json:"libVersion"`
	AwaitEvents []string `json:"awaitEvents"`
}

func (addressConfiguration *AddressConfiguration) GetAddress() string {
	return addressConfiguration.IP + ":" + addressConfiguration.Port
}

type RoutingConfig []BackendDeclaration

type EndpointConfig struct {
	Path           string `valid:"required~Required" json:"path"`
	Inner          bool   `json:"inner"`
	IgnoreOnRouter bool   `json:"ignoreOnRouter"`
}

type BackendDeclaration struct {
	ModuleName string               `json:"moduleName"`
	Version    string               `json:"version"`
	LibVersion string               `json:"libVersion"`
	Endpoints  []EndpointConfig     `json:"endpoints"`
	Address    AddressConfiguration `json:"address"`
}

func (backedConfig *BackendDeclaration) IsIPAndPortEqual(ip string, port string) bool {
	return backedConfig.Address.IP == ip && backedConfig.Address.Port == port
}

func (backedConfig *BackendDeclaration) IsAddressEquals(address AddressConfiguration) bool {
	return backedConfig.Address.IP == address.IP && backedConfig.Address.Port == address.Port
}

func (backedConfig *BackendDeclaration) IsPathsEqual(paths []EndpointConfig) bool {
	newBytes, err := json.Marshal(paths)
	if err != nil {
		return false
	}
	oldBytes, err := json.Marshal(backedConfig.Endpoints)
	if err != nil {
		return false
	}
	return bytes.Equal(newBytes, oldBytes)
}

func (cfg *RoutingConfig) AddAddressOrUpdate(backendConfig BackendDeclaration) bool {
	exists := false
	changed := false
	for i, v := range *cfg {
		if v.IsAddressEquals(backendConfig.Address) {
			if !v.IsPathsEqual(backendConfig.Endpoints) {
				(*cfg)[i] = backendConfig
				changed = true
			}
			exists = true
		}
	}
	if !exists {
		*cfg = append(*cfg, backendConfig)
		changed = true
	}
	return changed
}

func (cfg RoutingConfig) ToJSON() ([]byte, error) {
	return json.Marshal(cfg)
}
