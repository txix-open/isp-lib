package http

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc/status"
)

const (
	POST = "POST"
	GET  = "GET"
)

type ErrorResponse struct {
	StatusCode int
	Status     string
	Body       string
}

func (r ErrorResponse) Error() string {
	return fmt.Sprintf("statusCode:%d  status:%s  body:%s", r.StatusCode, r.Status, r.Body)
}

func (r ErrorResponse) ToGrpcError() error {
	st, _ := status.
		New(HttpStatusToCode(r.StatusCode), r.Status).
		WithDetails(&structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
			Fields: map[string]*structpb.Value{"response": {Kind: &structpb.Value_StringValue{StringValue: r.Body}}},
		}}})
	return st.Err()
}

type RestClient interface {
	Invoke(method, uri string, headers map[string]string, requestBody, responsePtr interface{}) error
	InvokeWithoutHeaders(method, uri string, requestBody, responsePtr interface{}) error
	Post(uri string, requestBody, responsePtr interface{}) error
	Get(uri string, responsePtr interface{}) error
	InvokeWithDynamicResponse(method, uri string, headers map[string]string, requestBody interface{}) (interface{}, error)
}
