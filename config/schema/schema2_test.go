package schema

import (
	"encoding/json"
	"fmt"
	"github.com/integration-system/jsonschema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateConfigSchema(t *testing.T) {
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
	obj2 := want.Definitions["Obj2"]
	obj2.Title = "title"
	obj2.Description = "desc"
	obj2.Properties["obj3"] = want.Definitions["Obj3"]
	want.Definitions["Obj"].Properties["obj2"] = obj2
	want.Type.Properties["sliceObj3"].Items = want.Definitions["Obj3"]
	want.Type.Properties["mapStringObj3"].PatternProperties[".*"] = want.Definitions["Obj3"]
	want.Type.Properties["d"] = want.Definitions["Obj"]
	want.Definitions = nil

	got := DereferenceSchema(ref.Reflect(config))

	if !assert.Equal(want, got) {
		json1, _ := json.MarshalIndent(got, "", "\t")
		fmt.Println(string(json1))
	}
}
