package schema

import (
	"github.com/integration-system/jsonschema"
	"reflect"
	"sync"
)

type generator func(field reflect.StructField, t *jsonschema.Type)

type customSchema struct {
	mx        sync.RWMutex
	mapSchema map[string]generator
}

var CustomGenerators = &customSchema{
	mapSchema: make(map[string]generator),
}

func (c *customSchema) Register(name string, f func(t *jsonschema.Type)) {
	c.mx.Lock()
	c.mapSchema[name] = f
	c.mx.Unlock()
}

func (c *customSchema) Remove(name string) {
	c.mx.Lock()
	delete(c.mapSchema, name)
	c.mx.Unlock()
}

func (c *customSchema) getGeneratorByName(name string) generator {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.mapSchema[name]
}
