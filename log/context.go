package log

import (
	"context"

	"go.uber.org/zap"
)

type contextLogKey int

var (
	contextKey = contextLogKey(-1)
)

func ContextLogValues(ctx context.Context) []zap.Field {
	if value, ok := ctx.Value(contextKey).([]zap.Field); ok {
		return value
	}
	return nil
}

func ToContext(ctx context.Context, kvs ...zap.Field) context.Context {
	existedValues := append(ContextLogValues(ctx), kvs...)
	return context.WithValue(ctx, contextKey, existedValues)
}
