package schema

import (
	"github.com/integration-system/jsonschema"
	"sync"
)

type customSchema struct {
	mx        sync.Mutex
	mapSchema map[string]func(t *jsonschema.Type)
}

var Custom = &customSchema{
	mapSchema: make(map[string]func(t *jsonschema.Type)),
}

func (c *customSchema) Create(name string, f func(t *jsonschema.Type)) {
	c.mx.Lock()
	c.mapSchema[name] = f
	c.mx.Unlock()
}

func (c *customSchema) Remove(name string) {
	c.mx.Lock()
	delete(c.mapSchema, name)
	c.mx.Unlock()
}

func (c *customSchema) getFunctionByName(name string) (f func(t *jsonschema.Type)) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.mapSchema[name]
}
