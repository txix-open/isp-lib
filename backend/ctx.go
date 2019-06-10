package backend

import "google.golang.org/grpc/metadata"

type RequestCtx interface {
	Method() string
	Metadata() metadata.MD
	RequestBody() []byte
	ResponseBody() []byte
	MappedRequest() interface{}
	MappedResponse() interface{}
	Error() error
}

type ctx struct {
	method         string
	md             metadata.MD
	requestBody    []byte
	responseBody   []byte
	mappedRequest  interface{}
	mappedResponse interface{}
	err            error
}

func (c *ctx) Method() string {
	return c.method
}

func (c *ctx) Metadata() metadata.MD {
	return c.md
}

func (c *ctx) RequestBody() []byte {
	return c.requestBody
}

func (c *ctx) ResponseBody() []byte {
	return c.responseBody
}

func (c *ctx) MappedRequest() interface{} {
	return c.mappedRequest
}

func (c *ctx) MappedResponse() interface{} {
	return c.mappedResponse
}

func (c *ctx) Error() error {
	return c.err
}

func newCtx() *ctx {
	return &ctx{}

}
