package backend

import (
	"errors"
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNoAliveConnections = errors.New("No alive connections")
)

type errorHandler func(err error)

type client struct {
	cc *grpc.ClientConn
	isp.BackendServiceClient
}

type InternalGrpcClient struct {
	next    int
	clients []*client
	length  int
	mu      *sync.Mutex
}

func (bc *InternalGrpcClient) Invoke(method string, callerId int, requestBody, responsePointer interface{}, mdPairs ...string) error {
	md := metadata.Pairs(
		utils.ProxyMethodNameHeader, method,
		utils.ApplicationIdHeader, strconv.Itoa(callerId),
	)
	if len(mdPairs) > 0 {
		md = metadata.Join(md, metadata.Pairs(mdPairs...))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer cancel()

	c := bc.nextConn()
	msg, err := toBytes(requestBody, ctx)
	if err != nil {
		return err
	}
	res, err := c.Request(ctx, msg)
	if err != nil {
		return err
	}
	if responsePointer != nil {
		return readBody(res, responsePointer)
	}

	return nil
}

func (bc *InternalGrpcClient) Close(errorHandler errorHandler) {
	for _, c := range bc.clients {
		if c == nil {
			continue
		}
		if err := c.cc.Close(); err != nil && errorHandler != nil {
			errorHandler(err)
		}
	}
}

func (bc *InternalGrpcClient) CloseQuietly() {
	bc.Close(nil)
}

func (bc *InternalGrpcClient) nextConn() isp.BackendServiceClient {
	if bc.length == 1 {
		return bc.clients[0]
	}

	bc.mu.Lock()
	sc := bc.clients[bc.next]
	bc.next = (bc.next + 1) % bc.length
	bc.mu.Unlock()
	return sc
}

func NewGrpcClient(addr string, options ...grpc.DialOption) (*InternalGrpcClient, error) {
	var e1 error
	c, e2 := NewGrpcClientV2([]string{addr}, func(err error) {
		e1 = err
	}, options...)
	if e1 != nil {
		return nil, e1
	} else if e2 != nil {
		return nil, e2
	} else {
		return c, nil
	}
}

func NewGrpcClientV2(addrList []string, errorHandler errorHandler, options ...grpc.DialOption) (*InternalGrpcClient, error) {
	clients := make([]*client, 0)
	for _, addr := range addrList {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		cc, err := grpc.DialContext(ctx, addr, options...)
		if err != nil {
			if errorHandler != nil {
				errorHandler(err)
			}
		} else {
			sc := isp.NewBackendServiceClient(cc)
			clients = append(clients, &client{cc, sc})
		}
	}
	if len(clients) == 0 {
		return nil, ErrNoAliveConnections
	}
	return &InternalGrpcClient{
		clients: clients,
		mu:      &sync.Mutex{},
		next:    0,
		length:  len(clients),
	}, nil
}
