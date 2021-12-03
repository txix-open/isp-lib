package endpoint

import (
	"context"

	"github.com/integration-system/isp-lib/v3/grpc/isp"
)

func ContextParam() ParamMapper {
	return ParamMapper{
		Type:    "context.Context",
		Builder: func(ctx context.Context, message *isp.Message) (interface{}, error) {
			return ctx, nil
		},
	}
}
