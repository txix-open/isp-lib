package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/integration-system/isp-lib/v3/grpc"
	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/integration-system/isp-lib/v3/requestid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
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
			return next(ctx, message)
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

func RequestId() Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errors.New("metadata expected in context")
			}
			values := md.Get(grpc.RequestIdHeader)
			requestId := ""
			if len(values) > 0 {
				requestId = values[0]
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))

			return next(ctx, message)
		}
	}
}

type LevelLogger interface {
	Log(ctx context.Context, level log.Level, message interface{}, fields ...log.Field)
}

func BodyLogger(logger LevelLogger, level log.Level) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			logger.Log(ctx, level, "request body", log.Any("requestBody", json.RawMessage(message.GetBytesBody())))

			response, err := next(ctx, message)
			if err == nil {
				logger.Log(ctx, level, "response body", log.Any("responseBody", json.RawMessage(response.GetBytesBody())))
			}

			return response, err
		}
	}
}
