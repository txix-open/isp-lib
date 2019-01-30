package http

import "github.com/valyala/fasthttp"

type Ctx struct {
	*fasthttp.RequestCtx
	m                  map[string]interface{}
	mappedRequestBody  interface{}
	mappedResponseBody interface{}
	err                error
	action             string
}

func (c *Ctx) Put(key string, value interface{}) {
	c.m[key] = value
}

func (c *Ctx) Get(key string) interface{} {
	return c.m[key]
}

func (c *Ctx) GetInt32(key string) (int32, bool) {
	if val, ok := c.m[key]; !ok {
		return 0, false
	} else if i, ok := val.(int32); ok {
		return i, true
	} else {
		return 0, false
	}
}

func (c *Ctx) MappedRequestBody() interface{} {
	return c.mappedRequestBody
}

func (c *Ctx) MappedResponseBody() interface{} {
	return c.mappedResponseBody
}

func (c *Ctx) Error() error {
	return c.err
}

func (c *Ctx) Action() string {
	return c.action
}
