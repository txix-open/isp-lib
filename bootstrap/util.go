package bootstrap

import (
	"encoding/json"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/golang-socketio"
	"net"
	"reflect"
)

func UnmarshalAddressListAndThen(event string, f func([]structure.AddressConfiguration)) func(*gosocketio.Channel, string) error {
	return func(_ *gosocketio.Channel, data string) error {
		logger.Infof("--- Got event: %s message: %s", event, data)

		list := make([]structure.AddressConfiguration, 0)
		if err := json.Unmarshal([]byte(data), &list); err != nil {
			logger.Error(err)
			return err
		} else {
			f(list)
		}
		return nil
	}
}

func assertSingleParamFunc(rt reflect.Type, expectingType string) {
	if rt.Kind() != reflect.Func ||
		rt.NumIn() != 1 ||
		rt.In(0).String() != expectingType {
		logger.Fatalf("Expecting function with one parameter of '%s' type, received '%s'", expectingType, rt.In(0).String())
	}
}

func assertTwoParamFunc(rt reflect.Type, expectingType string) {
	if rt.Kind() != reflect.Func ||
		rt.NumIn() != 2 ||
		rt.In(0).String() != expectingType ||
		rt.In(1).String() != expectingType {
		logger.Fatalf("Expecting function with two '%s' parameters", expectingType)
	}
}

func must(err error) {
	if err != nil {
		logger.Fatal(err)
	}
}

func makeAddressListConsumer(event string, c chan connectEvent) func([]structure.AddressConfiguration) {
	return func(list []structure.AddressConfiguration) {
		c <- connectEvent{event: event, addressList: list}
	}
}

func getOutboundIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}
