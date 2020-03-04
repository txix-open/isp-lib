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

const (
	methodPath    = "some/path"
	droppedAnswer = "dropper"
)

func TestNewRxGrpcClient_Balancing(t *testing.T) {
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
	addrs, _ := setupServers(serversCount)
	cli.ReceiveAddressList(addrs)
	time.Sleep(50 * time.Millisecond)

	answersMap := makeRequests(cli, requestsCount)

	const reqPerServer = requestsCount / serversCount
	assert.Len(t, answersMap, serversCount)
	for _, v := range answersMap {
		assert.Equal(t, reqPerServer, v)
	}
}

func TestNewRxGrpcClient_HandleUnavailableErrors(t *testing.T) {
	cli := NewRxGrpcClient(
		WithDialOptions(
			grpc.WithInsecure(),
		),
	)

	const (
		serversCount  = 5
		requestsCount = 10000
	)
	addrs, servers := setupServers(serversCount)
	cli.ReceiveAddressList(addrs)

	go func() {
		servers[0].Stop()
		time.Sleep(5 * time.Millisecond)
		servers[serversCount-1].Stop()
		time.Sleep(100 * time.Millisecond)
		servers[serversCount-2].GracefulStop()
	}()

	answersMap := makeRequests(cli, requestsCount)
	assert.Equal(t, 0, answersMap[droppedAnswer])
}

func makeRequests(cli *RxGrpcClient, requestsCount int) map[string]int {
	answersMap := make(map[string]int)
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
				answer = droppedAnswer
			}
			answersMapLock.Lock()
			answersMap[answer] = answersMap[answer] + 1
			answersMapLock.Unlock()
		}()
	}

	wg.Wait()
	return answersMap
}

func setupServers(count int) ([]structure.AddressConfiguration, []*GrpcServer) {
	addrs := make([]structure.AddressConfiguration, count)
	servers := make([]*GrpcServer, count)

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

		srv := newBackendGrpcServer(l, service)
		go srv.Start()
		servers[i] = srv
	}

	return addrs, servers
}
