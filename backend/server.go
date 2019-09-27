package backend

import (
	"errors"
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

var (
	grpcAddress *structure.AddressConfiguration
	server      *GrpcServer
	lock        = sync.Mutex{}
)

type GrpcServer struct {
	*grpc.Server
	service *DefaultService
}

func StartBackendGrpcServer(addr structure.AddressConfiguration, service *DefaultService, opt ...grpc.ServerOption) {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		log.Fatalf(stdcodes.ModuleGrpcServiceStartError, "grpc service has already started on %v", grpcAddress.GetAddress())
		return
	}

	grpcAddress = &addr

	var ln net.Listener
	var err error
	for ln, err = net.Listen("tcp", grpcAddress.GetAddress()); err != nil; {
		log.Errorf(stdcodes.ModuleGrpcServiceStartError, "open grpc port: %v, err: %v, retry after 3 second...", grpcAddress, err)
		time.Sleep(time.Second * 3)
	}

	StartBackendGrpcServerOn(addr, ln, service, opt...)
}

func StartBackendGrpcServerOn(addr structure.AddressConfiguration, ln net.Listener, service *DefaultService, opt ...grpc.ServerOption) {
	grpcAddress = &addr

	grpcServer := grpc.NewServer(opt...)
	isp.RegisterBackendServiceServer(grpcServer, service)
	server = &GrpcServer{grpcServer, service}

	go func() {
		log.Infof(stdcodes.ModuleGrpcServiceStart, "start grpc service on %v", grpcAddress.GetAddress())
		if err := server.Serve(ln); err != nil {
			log.Fatal(stdcodes.ModuleGrpcServiceStartError, err)
		} else {
			log.Infof(stdcodes.ModuleGrpcServiceManualShutdown, "shutdown grpc service on %v", grpcAddress.GetAddress())
		}
	}()
}

func StopGrpcServer() {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		server.GracefulStop()
		server = nil
		grpcAddress = nil
	}
}

func UpdateHandlers(methodPrefix string, handlersStructs ...interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		funcs, streams, err := resolveHandlers(methodPrefix, handlersStructs...)
		if err != nil {
			return err
		}
		server.service.functions = funcs
		server.service.streamConsumers = streams
		return nil
	} else {
		return errors.New("grpc server not initialized")
	}
}

func ServerIsInitialized() bool {
	lock.Lock()
	defer lock.Unlock()

	return server != nil
}

func checkPortIsFree(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	} else {
		ln.Close()
		return true
	}
}
