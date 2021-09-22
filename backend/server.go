package backend

import (
	"net"
	"sync"
	"time"

	"github.com/integration-system/isp-lib/v2/isp"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	server *GrpcServer
	lock   = sync.Mutex{}
)

func newBackendGrpcServer(listener net.Listener, service *DefaultService, opt ...grpc.ServerOption) *GrpcServer {
	grpcServer := grpc.NewServer(opt...)
	isp.RegisterBackendServiceServer(grpcServer, service)
	srv := &GrpcServer{
		Server:   grpcServer,
		service:  service,
		listener: listener,
	}

	return srv
}

type GrpcServer struct {
	*grpc.Server
	listener net.Listener
	service  *DefaultService
}

func (s *GrpcServer) Start() {
	log.Infof(stdcodes.ModuleGrpcServiceStart, "start grpc service on %s", s.listener.Addr().String())
	if err := s.Serve(s.listener); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf(stdcodes.ModuleGrpcServiceStartError, "grpc serve: %v", err)
	} else {
		log.Infof(stdcodes.ModuleGrpcServiceManualShutdown, "shutdown grpc service on %s", s.listener.Addr().String())
	}
}

func (s *GrpcServer) UpdateHandlers(methodPrefix string, handlersStructs ...interface{}) error {
	funcs, streams, err := resolveHandlers(methodPrefix, handlersStructs...)
	if err != nil {
		return err
	}
	s.service.functions = funcs
	s.service.streamConsumers = streams

	return nil
}

func StartBackendGrpcServer(addr structure.AddressConfiguration, service *DefaultService, opt ...grpc.ServerOption) {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		log.Fatalf(stdcodes.ModuleGrpcServiceStartError, "grpc service has already started on %v", addr.GetAddress())
	}

	var ln net.Listener
	var err error
	for ln, err = net.Listen("tcp", addr.GetAddress()); err != nil; {
		log.Errorf(stdcodes.ModuleGrpcServiceStartError, "open grpc port: %s, err: %v, retry after 3 second...", addr.GetAddress(), err)
		time.Sleep(time.Second * 3)
	}

	server = newBackendGrpcServer(ln, service, opt...)
	go server.Start()
}

func StartBackendGrpcServerOn(addr structure.AddressConfiguration, ln net.Listener, service *DefaultService, opt ...grpc.ServerOption) {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		log.Fatalf(stdcodes.ModuleGrpcServiceStartError, "grpc service has already started on %v", addr.GetAddress())
	}

	server = newBackendGrpcServer(ln, service, opt...)
	go server.Start()
}

func StopGrpcServer() {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		server.GracefulStop()
		server = nil
	}
}

func UpdateHandlers(methodPrefix string, handlersStructs ...interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	if server != nil {
		return server.UpdateHandlers(methodPrefix, handlersStructs)
	}

	return errors.New("grpc server not initialized")
}

func ServerIsInitialized() bool {
	lock.Lock()
	defer lock.Unlock()

	return server != nil
}
