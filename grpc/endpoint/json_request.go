package endpoint

import (
	"context"
	"reflect"

	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/integration-system/isp-lib/v3/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JsonRequestExtractor struct {

}

func (j JsonRequestExtractor) Extract(ctx context.Context, message *isp.Message, type_ reflect.Type) (reflect.Value, error) {
	instance := reflect.New(type_)
	iface := instance.Interface()
	err := json.Unmarshal(message.GetBytesBody(), iface)
	if err != nil {
		return reflect.Value{}, status.Errorf(codes.InvalidArgument, "unmarshal request body: %v", err)
	}
	return instance.Elem(), nil
}


