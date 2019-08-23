package schema

import (
	"github.com/integration-system/jsonschema"
	"github.com/mohae/deepcopy"
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
	cache := make(jsonschema.Definitions)
	s.Type = dereferenceType(s.Type, s.Definitions, cache)
	s.Definitions = nil
	return s
}

func dereferenceType(t *jsonschema.Type, definitions, cache jsonschema.Definitions) *jsonschema.Type {
	toDeref := t
	ref := strings.TrimPrefix(t.Ref, "#/definitions/")
	if ref != "" {
		if dereferenced, ok := cache[ref]; ok {
			copied := &(*dereferenced)
			copied.Title = t.Title
			copied.Description = t.Description
			copied.Default = t.Default
			return copied
		}

		def, ok := definitions[ref]
		if !ok {
			return t
		}
		def = deepcopy.Copy(def).(*jsonschema.Type)
		toDeref = def
	}

	if toDeref.Items != nil {
		toDeref.Items = dereferenceType(toDeref.Items, definitions, cache)
	}
	if toDeref.AdditionalItems != nil {
		toDeref.AdditionalItems = dereferenceType(toDeref.AdditionalItems, definitions, cache)
	}
	if toDeref.Not != nil {
		toDeref.Not = dereferenceType(toDeref.Not, definitions, cache)
	}
	toDeref.Properties = derefMap(toDeref.Properties, definitions, cache)
	toDeref.PatternProperties = derefMap(toDeref.PatternProperties, definitions, cache)
	toDeref.Dependencies = derefMap(toDeref.Dependencies, definitions, cache)
	toDeref.OneOf = derefArray(toDeref.OneOf, definitions, cache)
	toDeref.AllOf = derefArray(toDeref.AllOf, definitions, cache)
	toDeref.AnyOf = derefArray(toDeref.AnyOf, definitions, cache)

	if ref != "" {
		cache[ref] = toDeref

		copied := &(*toDeref)
		copied.Title = t.Title
		copied.Description = t.Description
		copied.Default = t.Default
		return copied
	}

	return toDeref
}

func derefArray(arr []*jsonschema.Type, definitions, cache jsonschema.Definitions) []*jsonschema.Type {
	for i, t := range arr {
		arr[i] = dereferenceType(t, definitions, cache)
	}
	return arr
}

func derefMap(m map[string]*jsonschema.Type, definitions, cache jsonschema.Definitions) map[string]*jsonschema.Type {
	for key, t := range m {
		m[key] = dereferenceType(t, definitions, cache)
	}
	return m
}
