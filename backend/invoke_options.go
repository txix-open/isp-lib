package backend

import (
	"google.golang.org/grpc/metadata"
	"time"
)

type InvokeOption func(opts *invokeOpts)

type invokeOpts struct {
	md      metadata.MD
	timeout time.Duration
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

func defaultInvokeOpts() *invokeOpts {
	return &invokeOpts{
		md:      metadata.Pairs(),
		timeout: 15 * time.Second,
	}
}
