package backend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/integration-system/isp-lib/v2/proto/stubs"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"

	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/structure"
	"google.golang.org/grpc"
)

const (
	defaultConnsPerAddress = 3
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

	var resolvedAddrs []resolver.Address
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

	ctx, cancel := context.WithTimeout(context.Background(), options.timeout)
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer cancel()

	msg, err := toBytes(requestBody)
	if err != nil {
		return err
	}

	res, err := rc.ispConn.Request(ctx, msg, options.callOpts...)
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

	streamClient, err := rc.ispConn.RequestStream(ctx)
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

// NewRxGrpcClient is incompatible with grpc.WithBlock() option
func NewRxGrpcClient(opts ...RxOption) *RxGrpcClient {
	rxGrpcClient := &RxGrpcClient{}
	for _, o := range opts {
		o(rxGrpcClient)
	}
	if rxGrpcClient.connsPerAddress <= 0 {
		rxGrpcClient.connsPerAddress = defaultConnsPerAddress
	}

	manualResolver, cleanup := manual.GenerateAndRegisterManualResolver()
	// unregister global resolver because we use resolver locally
	defer cleanup()
	rxGrpcClient.resolver = manualResolver

	dialOpts := append(rxGrpcClient.options,
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithResolvers(manualResolver),
	)
	serverAddr := manualResolver.Scheme() + ":///"
	conn, err := grpc.Dial(serverAddr, dialOpts...)

	// only if invalid options
	if err != nil {
		panic(err)
	}

	rxGrpcClient.conn = conn
	rxGrpcClient.ispConn = isp.NewBackendServiceClient(conn)

	return rxGrpcClient
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
