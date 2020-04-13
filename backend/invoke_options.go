package backend

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type InvokeOption func(opts *invokeOpts)

type invokeOpts struct {
	md       metadata.MD
	timeout  time.Duration
	ctx      context.Context
	callOpts []grpc.CallOption
}

func WithTimeout(timeout time.Duration) InvokeOption {
	return func(opts *invokeOpts) {
		opts.timeout = timeout
	}
}

func WithMetadata(md metadata.MD) InvokeOption {
	return func(opts *invokeOpts) {
		opts.md = md
	}
}

func WithCallOptions(callOpts ...grpc.CallOption) InvokeOption {
	return func(opts *invokeOpts) {
		opts.callOpts = callOpts
	}
}

func WithContext(ctx context.Context) InvokeOption {
	return func(opts *invokeOpts) {
		opts.ctx = ctx
	}
}

func defaultInvokeOpts() *invokeOpts {
	return &invokeOpts{
		md:      metadata.Pairs(),
		timeout: 15 * time.Second,
		ctx:     context.Background(),
	}
}
