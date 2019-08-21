package schema

import (
	"github.com/integration-system/jsonschema"
	"strings"
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

func DereferenceSchema(s Schema) Schema {
	for k, def := range s.Definitions {
		s.Definitions[k] = replaceRef(def, s.Definitions)
	}

	s.Type = replaceRef(s.Type, s.Definitions)
	return s
}

func replaceRef(def *jsonschema.Type, definitions jsonschema.Definitions) *jsonschema.Type {
	if def != nil {
		if def.Ref != "" {
			tp := strings.TrimPrefix(def.Ref, "#/definitions/")
			def = definitions[tp]
		}
		if def.Items != nil && def.Items.Ref != "" {
			tp := strings.TrimPrefix(def.Items.Ref, "#/definitions/")
			def.Items = definitions[tp]
		}
		for k, patternProp := range def.PatternProperties {
			if patternProp.Ref != "" {
				tp := strings.TrimPrefix(patternProp.Ref, "#/definitions/")
				def.PatternProperties[k] = definitions[tp]
			}
		}
		for k, dep := range def.Dependencies {
			if dep.Ref != "" {
				tp := strings.TrimPrefix(dep.Ref, "#/definitions/")
				def.Dependencies[k] = definitions[tp]
			}
		}
		for k, v := range def.AllOf {
			if v.Ref != "" {
				tp := strings.TrimPrefix(v.Ref, "#/definitions/")
				def.AllOf[k] = definitions[tp]
			}
		}
		for k, v := range def.AnyOf {
			if v.Ref != "" {
				tp := strings.TrimPrefix(v.Ref, "#/definitions/")
				def.AnyOf[k] = definitions[tp]
			}
		}
		for k, v := range def.OneOf {
			if v.Ref != "" {
				tp := strings.TrimPrefix(v.Ref, "#/definitions/")
				def.OneOf[k] = definitions[tp]
			}
		}
		replaceRef(def.Not, definitions)
		replaceRef(def.Media, definitions)
		for k, prop := range def.Properties {
			def.Properties[k] = replaceRef(prop, definitions)
		}
	}
	return def
}
