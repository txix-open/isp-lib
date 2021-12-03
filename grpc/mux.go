package grpc

import (
	"context"

	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	ProxyMethodNameHeader = "proxy_method_name"
)

type HandlerFunc func(ctx context.Context, message *isp.Message) (*isp.Message, error)

type Mux struct {
	unaryHandlers map[string]HandlerFunc
}

func NewMux() *Mux {
	return &Mux{
		unaryHandlers: make(map[string]HandlerFunc),
	}
}

func (m *Mux) Handle(endpoint string, handler HandlerFunc) *Mux {
	_, ok := m.unaryHandlers[endpoint]
	if ok {
		panic(errors.Errorf("handler for endpoint %v is already provided", endpoint))
		return nil
	}
	m.unaryHandlers[endpoint] = handler
	return m
}

func (m *Mux) Request(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("medata is expected in context")
	}
	values := md[ProxyMethodNameHeader]
	if len(values) == 0 {
		return nil, status.Errorf(codes.DataLoss, "metadata [%s] is required", ProxyMethodNameHeader)
	}
	endpoint := values[0]
	ctx = log.ToContext(ctx, log.String("endpoint", endpoint))
	handler, ok := m.unaryHandlers[endpoint]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "handler not found for endpoint %s", endpoint)
	}
	return handler(ctx, message)
}

func (m *Mux) RequestStream(_ isp.BackendService_RequestStreamServer) error {
	return status.Errorf(codes.Unimplemented, "service is not support stream rpc")
}
