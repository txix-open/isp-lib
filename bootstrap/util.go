package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	etp "github.com/integration-system/isp-etp-go/v2/client"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	"github.com/integration-system/isp-log/stdcodes"
	"nhooyr.io/websocket"
)

func assertSingleParamFunc(rt reflect.Type, expectingType string) {
	if rt.Kind() != reflect.Func ||
		rt.NumIn() != 1 ||
		rt.In(0).String() != expectingType {
		panic(fmt.Errorf("expecting function with one parameter of '%s' type, received '%s'", expectingType, rt.In(0).String()))
	}
}

func assertTwoParamFunc(rt reflect.Type, expectingType string) {
	if rt.Kind() != reflect.Func ||
		rt.NumIn() != 2 ||
		rt.In(0).String() != expectingType ||
		rt.In(1).String() != expectingType {
		panic(fmt.Errorf("expecting function with two '%s' parameters", expectingType))
	}
}

func callFunc(f *reflect.Value, args ...interface{}) {
	values := make([]reflect.Value, len(args))
	for i, v := range args {
		values[i] = reflect.ValueOf(v)
	}

	f.Call(values)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func makeAddressListConsumer(module string, c chan connectEvent) func([]structure.AddressConfiguration) {
	return func(list []structure.AddressConfiguration) {
		c <- connectEvent{module: module, addressList: list}
	}
}

func getOutboundIp(target string) (string, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}

func getWsUrl(host string, port string, secure bool, params map[string]string) string {
	var prefix string
	if secure {
		prefix = "wss://"
	} else {
		prefix = "ws://"
	}
	etpUrl := "/isp-etp/"
	connectionString := prefix + host + ":" + port + etpUrl
	if len(params) > 0 {
		vals := url.Values{}
		for k, v := range params {
			vals.Add(k, v)
		}
		connectionString += "?" + vals.Encode()
	}
	return connectionString
}

func parseConfigServiceAddresses(rawHosts, rawPorts string) ([]structure.AddressConfiguration, error) {
	hosts := strings.Split(rawHosts, ";")
	ports := strings.Split(rawPorts, ";")
	if len(hosts) != len(ports) {
		return nil, fmt.Errorf("different number of hosts/ports: %d/%d", len(hosts), len(ports))
	}
	addrs := make([]structure.AddressConfiguration, len(hosts))
	for i := 0; i < len(hosts); i++ {
		port, err := strconv.Atoi(ports[i])
		if err != nil {
			return nil, err
		}
		addrs[i] = structure.AddressConfiguration{
			IP:   hosts[i],
			Port: strconv.Itoa(port),
		}
	}

	return addrs, nil
}

func makeWebsocketConnectionStrings(sc structure.SocketConfiguration, addrs []structure.AddressConfiguration) []string {
	connStrings := make([]string, len(addrs))
	for _, addr := range addrs {
		connectionString := getWsUrl(
			addr.IP,
			addr.Port,
			sc.Secure,
			sc.UrlParams,
		)
		connStrings = append(connStrings, connectionString)
	}

	return connStrings
}

type ackEventMsg struct {
	event string
	data  interface{}
	err   error
}

func (m *ackEventMsg) info() (int, string) {
	switch m.event {
	case utils.ModuleSendConfigSchema:
		return stdcodes.ConfigServiceSendConfigSchema, "send config schema and default config"
	case utils.ModuleSendRequirements:
		return stdcodes.ConfigServiceSendRequirements, "send module requirements"
	case utils.ModuleReady:
		return stdcodes.ConfigServiceSendModuleReady, "send module declaration"
	}
	return 0, "INVALID ackEvent message event"
}

func ackEvent(client etp.Client, event string, data interface{}, bf backoff.BackOff) ackEventMsg {
	msg := ackEventMsg{event: event, data: data}
	bytes, err := json.Marshal(data)
	if err != nil {
		msg.err = fmt.Errorf("marshal payload to json: %v", err)
		return msg
	}

	var response []byte
	var connClosedErr error
	ack := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), ackMaxTimeout)
		defer cancel()
		response, err = client.EmitWithAck(ctx, event, bytes)
		if errors.As(err, &websocket.CloseError{}) {
			connClosedErr = err
			return nil
		} else if err != nil {
			return err
		} else if string(response) != utils.WsOkResponse {
			return fmt.Errorf("with invalid response: %s", response)
		}
		return nil
	}
	err = backoff.Retry(ack, bf)
	if connClosedErr != nil {
		err = connClosedErr
	}
	if err != nil {
		msg.err = fmt.Errorf("ack event to config service: %v", err)
		return msg
	}

	return msg
}

func getDefaultBackoff(ctx context.Context) backoff.BackOff {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = ackMaxTotalRetryTime
	backOff.RandomizationFactor = ackRetryRandomizationFactor
	bf := backoff.WithContext(backOff, ctx)
	return bf
}

type RoundRobinStrings struct {
	strings []string
	index   int
}

func (u *RoundRobinStrings) Get() string {
	if u.index == -1 {
		var random = rand.New(rand.NewSource(time.Now().UnixNano()))
		u.index = random.Intn(len(u.strings))
	} else {
		u.index += 1
		if u.index > len(u.strings)-1 {
			u.index = 0
		}
	}
	return u.strings[u.index]
}

func NewRoundRobinStrings(urls []string) *RoundRobinStrings {
	return &RoundRobinStrings{
		strings: urls,
		index:   -1,
	}
}
