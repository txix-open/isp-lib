package structure

import (
	"bytes"
	"encoding/json"
	"path"
)

type ModuleInfo struct {
	ModuleName  string   `json:"moduleName"`
	Version     string   `json:"version"`
	LibVersion  string   `json:"libVersion"`
	AwaitEvents []string `json:"awaitEvents"`
}

type RoutingConfig []BackendDeclaration

// Deprecated: use EndpointDescriptor instead
type EndpointConfig struct {
	Path           string `valid:"required~Required" json:"path"`
	Inner          bool   `json:"inner"`
	IgnoreOnRouter bool   `json:"ignoreOnRouter"`
}

type EndpointDescriptor struct {
	Path             string `valid:"required~Required" json:"path"`
	Inner            bool   `json:"inner"`
	UserAuthRequired bool
	Extra            map[string]interface{}
	Handler          interface{} `json:"-"`
}

type ModuleDependency struct {
	Name     string
	Required bool
}

func DescriptorsWithPrefix(prefix string, descriptors []EndpointDescriptor) []EndpointDescriptor {
	for i, descriptor := range descriptors {
		descriptor.Path = path.Join(prefix, descriptor.Path)
		descriptors[i] = descriptor
	}

	return descriptors
}

type BackendDeclaration struct {
	ModuleName      string               `json:"moduleName"`
	Version         string               `json:"version"`
	LibVersion      string               `json:"libVersion"`
	Endpoints       []EndpointDescriptor `json:"endpoints"`
	RequiredModules []ModuleDependency   `json:"requiredModules"`
	Address         AddressConfiguration `json:"address"`
}

func (backedConfig *BackendDeclaration) IsIPAndPortEqual(ip string, port string) bool {
	return backedConfig.Address.IP == ip && backedConfig.Address.Port == port
}

func (backedConfig *BackendDeclaration) IsAddressEquals(address AddressConfiguration) bool {
	return backedConfig.Address.IP == address.IP && backedConfig.Address.Port == address.Port
}

func (backedConfig *BackendDeclaration) IsPathsEqual(paths []EndpointDescriptor) bool {
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
