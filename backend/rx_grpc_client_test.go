package backend

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const methodPath = "some/path"

func TestNewRxGrpcClient2(t *testing.T) {
	cli := NewRxGrpcClient(
		WithDialOptions(
			grpc.WithInsecure(),
		),
		WithConnectionsPerAddress(2),
	)

	const (
		serversCount  = 5
		requestsCount = 1000
	)
	addrs := setupServers(serversCount)
	cli.ReceiveAddressList(addrs)
	time.Sleep(50 * time.Millisecond)

	answersMap := make(map[string]int, serversCount)
	answersMapLock := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for i := 0; i < requestsCount; i++ {
		wg.Add(1)
		ii := i
		go func() {
			defer wg.Done()
			var answer string
			err := cli.Invoke(methodPath, ii, nil, &answer)
			if err != nil {
				answer = "dropped"
			}
			answersMapLock.Lock()
			answersMap[answer] = answersMap[answer] + 1
			answersMapLock.Unlock()
		}()
	}
	wg.Wait()

	reqPerServer := requestsCount / serversCount
	assert.Len(t, answersMap, serversCount)
	for _, v := range answersMap {
		assert.Equal(t, reqPerServer, v)
	}
}

func setupServers(count int) []structure.AddressConfiguration {
	addrs := make([]structure.AddressConfiguration, count)

	for i := 0; i < count; i++ {
		answer := strconv.Itoa(i)
		descriptors := []structure.EndpointDescriptor{
			{
				Path: methodPath,
				Handler: func() (string, error) {
					return answer, nil
				},
			},
		}
		service := NewDefaultService(descriptors)

		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}

		port := strings.Split(l.Addr().String(), ":")[1]
		addr := structure.AddressConfiguration{IP: "127.0.0.1", Port: port}
		addrs[i] = addr
		StartBackendGrpcServerOn(addr, l, service)
	}

	return addrs
}
