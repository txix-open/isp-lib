package endpoint

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/integration-system/isp-lib/v3/grpc"
	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recovery() Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (msg *isp.Message, err error) {
			defer func() {
				if r := recover(); r != nil {
					recovered, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", recovered)
					}
					stack := make([]byte, 4<<10)
					length := runtime.Stack(stack, false)
					err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				}
			}()
			return next(ctx, msg)
		}
	}
}

func ErrorLogger(logger log.Logger) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			result, err := next(ctx, message)
			if err != nil {
				logger.Error(ctx, err)
			}
			return result, err
		}
	}
}

type Validator interface {
	Validate(ctx context.Context, value interface{}) (bool, map[string]string)
}

func RequestBodyValidationMiddleware(validator Validator) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			requestBody := RequestBodyFromContext(ctx)
			if requestBody == nil {
				return next(ctx, message)
			}
			ok, errors := validator.Validate(ctx, requestBody)
			if !ok {
				descriptions := make([]string, 0, len(errors))
				for field, err := range errors {
					descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
				}
				err := status.Errorf(codes.InvalidArgument, "invalid request body: %v", strings.Join(descriptions, ";"))
				return nil, err
			}
			return next(ctx, message)
		}
	}
}

func RequestId() Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			//todo
		}
	}
}
