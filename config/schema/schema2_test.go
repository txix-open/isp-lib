package schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/integration-system/isp-lib/v2/internal/testdata/testone"
	"github.com/integration-system/isp-lib/v2/internal/testdata/testtwo"
	"github.com/integration-system/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestGenerateConfigSchema(t *testing.T) {
	const packageName = "schema"

	assert := assert.New(t)
	type Obj3 struct {
		C string
	}
	type Obj2 struct {
		B    string
		Obj3 Obj3
	}
	type Obj struct {
		Str  string
		Obj2 Obj2 `schema:"title,desc"`
	}
	type Config struct {
		A             string
		B             int
		C             bool
		SliceObj3     []Obj3
		MapStringObj3 map[string]Obj3
		D             Obj
	}

	config := Config{
		A: "test1",
		B: 1,
		C: true,
		D: Obj{
			Str:  "test1",
			Obj2: Obj2{B: "bstring", Obj3: Obj3{C: "somestr"}},
		},
		SliceObj3: []Obj3{{C: "test"}},
		MapStringObj3: map[string]Obj3{
			"s1": {C: "test1"},
			"s2": {C: "test2"},
		},
	}
	ref := jsonschema.Reflector{
		FieldNameReflector: GetNameAndRequiredFlag,
		FieldReflector:     SetProperties,
		ExpandedStruct:     true,
	}
	want := Schema(ref.Reflect(config))
	obj2 := want.Definitions[packageName+"Obj2"]
	obj2.Title = "title"
	obj2.Description = "desc"
	obj2.Properties["obj3"] = want.Definitions[packageName+"Obj3"]
	want.Definitions[packageName+"Obj"].Properties["obj2"] = obj2
	want.Type.Properties["sliceObj3"].Items = want.Definitions[packageName+"Obj3"]
	want.Type.Properties["mapStringObj3"].PatternProperties[".*"] = want.Definitions[packageName+"Obj3"]
	want.Type.Properties["d"] = want.Definitions[packageName+"Obj"]
	want.Definitions = nil

	got := DereferenceSchema(ref.Reflect(config))

	if !assert.Equal(want, got) {
		json1, _ := json.MarshalIndent(got, "", "\t")
		fmt.Println(string(json1))
	}
}

func TestDefinitionCrossPackageCollision(t *testing.T) {
	type СommonCfg struct {
		One testone.Config
		Two testtwo.Config
	}

	cfg := &СommonCfg{
		One: testone.Config{Aa: "str"},
		Two: testtwo.Config{Bb: 5},
	}
	s := GenerateConfigSchema(cfg)
	json1, _ := json.MarshalIndent(s, "", "\t")

	var data map[string]interface{}
	_ = json.Unmarshal(json1, &data)
	if !assert.Equal(t, len(data["definitions"].(map[string]interface{})), 2) {
		fmt.Println(string(json1))
	}

	refs := make([]string, 0, 2)
	for _, prop := range s.Type.Properties {
		refs = append(refs, prop.Ref)
	}
	if !assert.Equal(t, len(refs), 2) {
		fmt.Println(string(json1))
	}
	if !assert.Equal(t, (refs[0] != refs[1]), true) {
		fmt.Println(string(json1))
	}
}
