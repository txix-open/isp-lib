package endpoint

import (
	"context"
)

type requestBodyContextKey int

var (
	requestBodyContextKeyValue = requestBodyContextKey(-1)
)

func RequestBodyFromContext(ctx context.Context) interface{} {
	return ctx.Value(requestBodyContextKeyValue)
}

func RequestBodyToContext(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, requestBodyContextKeyValue, value)
}

