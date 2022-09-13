package backend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/integration-system/isp-lib/v2/isp"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"google.golang.org/grpc/status"

	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/structure"
	"google.golang.org/grpc"
)

const (
	defaultConnsPerAddress = 1
	resolverScheme         = "isp"
	resolverUrl            = resolverScheme + ":///"
)

type GrpcClient interface {
	ReceiveAddressList([]structure.AddressConfiguration) bool
	Invoke(method string, callerId int, requestBody, responsePointer interface{}, opts ...InvokeOption) error
	InvokeStream(method string, callerId int, consumer streaming.StreamConsumer) error
	Conn() isp.BackendServiceClient
	Close() error
}

var _ GrpcClient = (*RxGrpcClient)(nil)

type RxGrpcClient struct {
	options         []grpc.DialOption
	connsPerAddress int

	conn     *grpc.ClientConn
	ispConn  isp.BackendServiceClient
	resolver *manual.Resolver
}

func (rc *RxGrpcClient) ReceiveAddressList(list []structure.AddressConfiguration) bool {
	if len(list) == 0 {
		return true
	}

	resolvedAddrs := make([]resolver.Address, 0, len(list)*rc.connsPerAddress)
	for j := 1; j < rc.connsPerAddress+1; j++ {
		for i := 0; i < len(list); i++ {
			addr := list[i].GetAddress()
			resolvedAddrs = append(resolvedAddrs, resolver.Address{
				Addr:       addr,
				ServerName: fmt.Sprintf("%s_%d", addr, j),
			})
		}
	}
	rc.resolver.UpdateState(resolver.State{Addresses: resolvedAddrs})

	return true
}

func (rc *RxGrpcClient) Invoke(method string, callerId int, requestBody, responsePointer interface{}, opts ...InvokeOption) error {
	options := defaultInvokeOpts()
	for _, opt := range opts {
		opt(options)
	}

	md := options.md
	md.Set(utils.ProxyMethodNameHeader, method)
	md.Set(utils.ApplicationIdHeader, strconv.Itoa(callerId))

	ctx, cancel := context.WithTimeout(options.ctx, options.timeout)
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer cancel()

	msg, err := toBytes(requestBody)
	if err != nil {
		return err
	}

	var res *isp.Message
	err = retryUnavailable(func() (err error) {
		res, err = rc.ispConn.Request(ctx, msg, options.callOpts...)
		return
	})
	if err != nil {
		return err
	}

	if responsePointer != nil {
		return readBody(res, responsePointer)
	}

	return nil
}

func (rc *RxGrpcClient) InvokeStream(method string, callerId int, consumer streaming.StreamConsumer) error {
	md := metadata.Pairs(
		utils.ProxyMethodNameHeader, method,
		utils.ApplicationIdHeader, strconv.Itoa(callerId),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer cancel()

	var streamClient isp.BackendService_RequestStreamClient
	err := retryUnavailable(func() (err error) {
		streamClient, err = rc.ispConn.RequestStream(ctx)
		return
	})
	if err != nil {
		return err
	}

	return consumer(streamClient, md)
}

func (rc *RxGrpcClient) Conn() isp.BackendServiceClient {
	return rc.ispConn
}

func (rc *RxGrpcClient) Close() error {
	return rc.conn.Close()
}

func retryUnavailable(f func() error) error {
	bf := backoff.WithMaxRetries(new(backoff.ZeroBackOff), 2)
	var responseErr error
	err := backoff.Retry(func() error {
		err := f()
		switch status.Code(err) {
		case codes.Unavailable:
			return err
		default:
			responseErr = err
			return nil
		}
	}, bf)

	if responseErr != nil {
		return responseErr
	}
	return err
}

// NewRxGrpcClient is incompatible with grpc.WithBlock() option
func NewRxGrpcClient(opts ...RxOption) *RxGrpcClient {
	client := &RxGrpcClient{}
	for _, o := range opts {
		o(client)
	}
	if client.connsPerAddress <= 0 {
		client.connsPerAddress = defaultConnsPerAddress
	}

	client.resolver = manual.NewBuilderWithScheme(resolverScheme)
	dialOpts := append(client.options,
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithResolvers(client.resolver),
	)
	conn, err := grpc.Dial(resolverUrl, dialOpts...)

	// only if invalid options
	if err != nil {
		panic(err)
	}

	client.conn = conn
	client.ispConn = isp.NewBackendServiceClient(conn)

	return client
}

type RxOption func(rc *RxGrpcClient)

func WithDialOptions(opts ...grpc.DialOption) RxOption {
	return func(rc *RxGrpcClient) {
		rc.options = opts
	}
}

func WithConnectionsPerAddress(factor int) RxOption {
	return func(rc *RxGrpcClient) {
		rc.connsPerAddress = factor
	}
}
