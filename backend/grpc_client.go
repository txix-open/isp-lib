package backend

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/integration-system/isp-lib/v2/proto/stubs"
	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoAliveConnections = errors.New("no alive connections")
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

	metricIntercept func(method string, dur time.Duration, err error)
}

func (client *InternalGrpcClient) doInvoke(method string, callerId int, requestBody interface{}, responseHandler func(*isp.Message, time.Time) error, opts ...InvokeOption) error {
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

	start := time.Now()

	c := client.nextConn()
	msg, err := toBytes(requestBody)
	if err != nil {
		return client.throwMetric(method, start, err)
	}
	if res, err := c.Request(ctx, msg); err != nil {
		return client.throwMetric(method, start, err)
	} else {
		return responseHandler(res, start)
	}

}

func (client *InternalGrpcClient) InvokeWithDynamicStruct(method string, callerId int, requestBody interface{}, opts ...InvokeOption) (interface{}, error) {
	var (
		resp interface{}
		t    time.Time
	)
	if err := client.doInvoke(method, callerId, requestBody, func(res *isp.Message, start time.Time) error {
		t = start
		bytes := res.GetBytesBody()
		if len(bytes) != 0 {
			switch bytes[0] {
			case '{':
				resp = make(map[string]interface{}, 0)
				if err := json.Unmarshal(bytes, &resp); err != nil {
					return err
				}
			case '[':
				resp = make([]interface{}, 0)
				if err := json.Unmarshal(bytes, &resp); err != nil {
					return err
				}
			default:
				resp = map[string]string{"response": string(bytes)}
			}
		}
		return nil
	}, opts...); err != nil {
		return nil, err
	} else {
		return resp, client.throwMetric(method, t, nil)
	}
}

func (client *InternalGrpcClient) Invoke(method string, callerId int, requestBody, responsePointer interface{}, opts ...InvokeOption) error {
	return client.doInvoke(method, callerId, requestBody, func(res *isp.Message, start time.Time) error {
		if responsePointer != nil {
			err := readBody(res, responsePointer)
			return client.throwMetric(method, start, err)
		}
		return client.throwMetric(method, start, nil)
	}, opts...)
}

func (client *InternalGrpcClient) InvokeStream(method string, callerId int, consumer streaming.StreamConsumer) error {
	md := metadata.Pairs(
		utils.ProxyMethodNameHeader, method,
		utils.ApplicationIdHeader, strconv.Itoa(callerId),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer cancel()

	conn := client.nextConn()
	streamClient, err := conn.RequestStream(ctx)
	if err != nil {
		return err
	}

	return consumer(streamClient, md)
}

func (client *InternalGrpcClient) Close(errorHandler errorHandler) {
	for _, c := range client.clients {
		if c == nil {
			continue
		}
		if err := c.cc.Close(); err != nil && errorHandler != nil {
			errorHandler(err)
		}
	}
}

func (client *InternalGrpcClient) CloseQuietly() {
	client.Close(nil)
}

func (client *InternalGrpcClient) WithMetric(catchMetric func(method string, dur time.Duration, err error)) *InternalGrpcClient {
	client.metricIntercept = catchMetric
	return client
}

func (client *InternalGrpcClient) Conn() (isp.BackendServiceClient, error) {
	if client.length == 0 {
		return nil, ErrNoAliveConnections
	}
	return client.nextConn(), nil
}

func (client *InternalGrpcClient) nextConn() isp.BackendServiceClient {
	if client.length == 1 {
		return client.clients[0]
	}

	client.mu.Lock()
	sc := client.clients[client.next]
	client.next = (client.next + 1) % client.length
	client.mu.Unlock()
	return sc
}

func (client *InternalGrpcClient) throwMetric(method string, start time.Time, err error) error {
	if client.metricIntercept != nil {
		client.metricIntercept(method, time.Since(start), err)
	}
	return err
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
