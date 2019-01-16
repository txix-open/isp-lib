package backend

import (
	"errors"
	"github.com/integration-system/isp-lib/http"
	"github.com/integration-system/isp-lib/utils"
	"strconv"
	"sync"
)

var (
	ErrNoWellKnownConverters = errors.New("No well known converters")
)

type InternalHttpClient struct {
	next       int
	addrList   []string
	length     int
	mu         sync.Mutex
	restClient http.RestClient
}

func (bc *InternalHttpClient) Invoke(method string, callerId int, requestBody, responsePointer interface{}) error {
	addr := bc.nextAddr()
	if addr == "" {
		return ErrNoWellKnownConverters
	}
	headers := map[string]string{utils.ApplicationIdHeader: strconv.Itoa(callerId)}
	return bc.restClient.Invoke(http.POST, addr+method, headers, requestBody, responsePointer)
}

func (bc *InternalHttpClient) nextAddr() string {
	if bc.length == 0 {
		return ""
	}
	if bc.length == 1 {
		return bc.addrList[0]
	}

	bc.mu.Lock()
	sc := bc.addrList[bc.next]
	bc.next = (bc.next + 1) % bc.length
	bc.mu.Unlock()
	return sc
}

func NewHttpClientV2(converterAddr string) *InternalHttpClient {
	return NewHttpClientV3([]string{converterAddr})
}

func NewHttpClientV3(convertersAddrList []string) *InternalHttpClient {
	return &InternalHttpClient{
		next:       0,
		addrList:   convertersAddrList,
		length:     len(convertersAddrList),
		restClient: http.NewJsonRestClient(),
	}
}
