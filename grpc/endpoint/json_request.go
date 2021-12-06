package endpoint

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/integration-system/isp-lib/v3/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator interface {
	Validate(value interface{}) (bool, map[string]string)
}

type JsonRequestExtractor struct {
	validator Validator
}

func (j JsonRequestExtractor) Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := json.Unmarshal(message.GetBytesBody(), instance.Interface())
	if err != nil {
		return reflect.Value{}, status.Errorf(codes.InvalidArgument, "unmarshal request body: %v", err)
	}

	elem := instance.Elem()

	ok, details := j.validator.Validate(elem.Interface())
	if !ok {
		descriptions := make([]string, 0, len(details))
		for field, err := range details {
			descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
		}
		err := status.Errorf(codes.InvalidArgument, "invalid request body: %v", strings.Join(descriptions, ";"))
		return reflect.Value{}, err
	}

	return elem, nil
}
