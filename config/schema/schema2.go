package schema

import (
	"github.com/integration-system/jsonschema"
)

const (
	schemaTag = "schema"
)

type Schema *jsonschema.Schema

type ConfigSchema struct {
	Version       string                 `json:"version"`
	Schema        Schema                 `json:"schema"`
	DefaultConfig map[string]interface{} `json:"config"`
}

func GenerateConfigSchema(cfgPtr interface{}) Schema {
	ref := jsonschema.Reflector{
		FieldNameReflector: GetNameAndRequiredFlag,
		FieldReflector:     SetProperties,
		ExpandedStruct:     true,
	}
	s := ref.Reflect(cfgPtr)
	s.Title = "Remote config"
	s.Version = ""
	return s
}
