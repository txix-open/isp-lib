package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"net"
	"reflect"
	"strings"
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

func callFunc(f *reflect.Value, args ...interface{}) {
	values := make([]reflect.Value, len(args))
	for i, v := range args {
		values[i] = reflect.ValueOf(v)
	}

	f.Call(values)
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

func getJsonModuleDeclaration(moduleInfo ModuleInfo) ([]byte, error) {
	endpoints := backend.GetEndpoints(moduleInfo.ModuleName, moduleInfo.Handlers...)
	addr := moduleInfo.GrpcOuterAddress.IP
	hasSchema := strings.Contains(addr, "http://")
	if hasSchema {
		addr = strings.Replace(addr, "http://", "", -1)
	}
	if addr == "" {
		ip, err := getOutboundIp()
		if err != nil {
			logger.Warn(err)
		} else {
			if hasSchema {
				ip = fmt.Sprintf("http://%s", ip)
			}
			moduleInfo.GrpcOuterAddress.IP = ip
		}
	}
	declaration := structure.BackendDeclaration{
		ModuleName: moduleInfo.ModuleName,
		Version:    moduleInfo.ModuleVersion,
		Address:    moduleInfo.GrpcOuterAddress,
		LibVersion: LibraryVersion,
		Endpoints:  endpoints,
	}

	return json.Marshal(declaration)
}
