package backend

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/structure"
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
		logger.Fatal("Grpc server has already started at", grpcAddress.GetAddress())
	}

	grpcAddress = &addr

	var ln net.Listener
	var err error
	for ln, err = net.Listen("tcp", grpcAddress.GetAddress()); err != nil; {
		time.Sleep(time.Second * 3)
		logger.Warnf("Error grpc connection: %v, try again, err: %v", grpcAddress, err)
	}

	StartBackendGrpcServerOn(addr, ln, service, opt...)
}

func StartBackendGrpcServerOn(addr structure.AddressConfiguration, ln net.Listener, service *DefaultService, opt ...grpc.ServerOption) {
	grpcAddress = &addr

	grpcServer := grpc.NewServer(opt...)
	isp.RegisterBackendServiceServer(grpcServer, service)
	server = &GrpcServer{grpcServer, service}

	go func() {
		logger.Infof("Start backend grpc server on %s", grpcAddress.GetAddress())
		if err := server.Serve(ln); err != nil {
			logger.Warnf("Grpc backend server shutdown with error: %v", err)
		} else {
			logger.Info("Grpc backend server shutdown")
		}
	}()
}

func StopGrpcServer() {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		server.GracefulStop()
		/*for !checkPortIsFree(grpcAddress.Port) {
			stopCounter++
			time.Sleep(time.Second * time.Duration(stopCounter))
			logger.Warnf("Wait for free port for new grpc connection, address: %v", grpcAddress)
			if stopCounter > 4 {
				logger.Warnf("Hard stop grpc server, address: %v", grpcAddress)
				server.Stop()
			}
		}
		time.Sleep(time.Second * 3)
		if !checkPortIsFree(grpcAddress.Port) {
			logger.Error("Hard stop grpc server, address: %v", grpcAddress)
			panic(errors.New("Grpc server error"))
		}*/
		server = nil
		grpcAddress = nil
	}
}

func UpdateHandlers(methodPrefix string, handlersStructs ...interface{}) {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		funcs, streams := resolveHandlers(methodPrefix, handlersStructs...)
		server.service.functions = funcs
		server.service.streamConsumers = streams
	} else {
		logger.Fatal("Grpc server not initialized")
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
