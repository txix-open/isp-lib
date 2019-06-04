package backend

import (
	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-lib/structure"
	"google.golang.org/grpc"
	"sort"
	"sync"
	"time"
)

type RxOption func(rc *RxGrpcClient)

type RxGrpcClient struct {
	*InternalGrpcClient
	active          bool
	lock            sync.RWMutex
	lastRoutersList []string
	eh              errorHandler
	options         []grpc.DialOption

	metricIntercept func(method string, dur time.Duration, err error)
}

func (rc *RxGrpcClient) ReceiveAddressList(list []structure.AddressConfiguration) bool {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if !rc.active {
		return true
	}

	addrList := make([]string, len(list))
	for i, addr := range list {
		addrList[i] = addr.GetAddress()
	}
	sort.Strings(addrList)

	if !cmp.Equal(rc.lastRoutersList, addrList) {
		if rc.InternalGrpcClient != nil {
			rc.InternalGrpcClient.CloseQuietly()
			rc.InternalGrpcClient = nil
		}
		if c, err := NewGrpcClientV2(addrList, rc.eh, rc.options...); err != nil {
			if rc.eh != nil {
				rc.eh(err)
			}
		} else {
			c.WithMetric(rc.metricIntercept)
			rc.InternalGrpcClient = c
		}
		rc.lastRoutersList = addrList
	}

	return rc.InternalGrpcClient != nil && len(rc.clients) > 0
}

func (rc *RxGrpcClient) Close() {
	rc.lock.Lock()

	if rc.InternalGrpcClient != nil {
		rc.InternalGrpcClient.CloseQuietly()
		rc.InternalGrpcClient = nil
		rc.active = false
	}

	rc.lock.Unlock()
}

func (rc *RxGrpcClient) Visit(visitor func(c *InternalGrpcClient) error) error {
	rc.lock.RLock()
	defer rc.lock.RUnlock()

	if rc.InternalGrpcClient == nil {
		return ErrNoAliveConnections
	}
	return visitor(rc.InternalGrpcClient)
}

func NewRxGrpcClient(opts ...RxOption) *RxGrpcClient {
	rxGrpcClient := &RxGrpcClient{active: true}
	for _, o := range opts {
		o(rxGrpcClient)
	}
	return rxGrpcClient
}

func WithDialOptions(opts ...grpc.DialOption) RxOption {
	return func(rc *RxGrpcClient) {
		rc.options = opts
	}
}

func WithDialingErrorHandler(eh errorHandler) RxOption {
	return func(rc *RxGrpcClient) {
		rc.eh = eh
	}
}

func WithMetric(catchMetric func(method string, dur time.Duration, err error)) RxOption {
	return func(rc *RxGrpcClient) {
		rc.metricIntercept = catchMetric
	}
}
