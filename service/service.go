package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	greflect "google.golang.org/grpc/reflection"

	"go.linka.cloud/grpc/logger"
	"go.linka.cloud/grpc/registry"
	"go.linka.cloud/grpc/registry/noop"
)

type Service interface {
	greflect.GRPCServer

	Options() Options
	Start() error
	Stop() error
	Close() error
}

func New(opts ...Option) (Service, error) {
	return newService(opts...)
}

type service struct {
	opts   *options
	cancel context.CancelFunc

	server  *grpc.Server
	mu      sync.Mutex
	running bool

	// inproc Channel is used to serve grpc gateway
	inproc   *inprocgrpc.Channel
	services map[string]*serviceInfo

	id     string
	regSvc *registry.Service
	closed chan struct{}
}

func newService(opts ...Option) (*service, error) {
	s := &service{
		opts:     NewOptions(),
		id:       uuid.New().String(),
		inproc:   &inprocgrpc.Channel{},
		services: make(map[string]*serviceInfo),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range opts {
		f(s.opts)
	}

	md := md(s.opts)
	if md != nil {
		s.opts.unaryServerInterceptors = append([]grpc.UnaryServerInterceptor{md.UnaryServerInterceptor()}, s.opts.unaryServerInterceptors...)
		s.opts.unaryClientInterceptors = append([]grpc.UnaryClientInterceptor{md.UnaryClientInterceptor()}, s.opts.unaryClientInterceptors...)
		s.opts.streamServerInterceptors = append([]grpc.StreamServerInterceptor{md.StreamServerInterceptor()}, s.opts.streamServerInterceptors...)
		s.opts.streamClientInterceptors = append([]grpc.StreamClientInterceptor{md.StreamClientInterceptor()}, s.opts.streamClientInterceptors...)
	}

	if s.opts.mux == nil {
		s.opts.mux = http.NewServeMux()
	}
	if s.opts.error != nil {
		return nil, s.opts.error
	}
	s.opts.ctx, s.cancel = context.WithCancel(s.opts.ctx)
	go func() {
		for {
			select {
			case <-s.opts.ctx.Done():
				s.Stop()
			}
		}
	}()
	if s.opts.registry == nil {
		s.opts.registry = noop.New()
	}

	if err := s.opts.parseTLSConfig(); err != nil {
		return nil, err
	}

	ui := grpcmiddleware.ChainUnaryServer(s.opts.unaryServerInterceptors...)
	s.inproc = s.inproc.WithServerUnaryInterceptor(ui)

	si := grpcmiddleware.ChainStreamServer(s.opts.streamServerInterceptors...)
	s.inproc = s.inproc.WithServerStreamInterceptor(si)

	gopts := []grpc.ServerOption{
		grpc.StreamInterceptor(si),
		grpc.UnaryInterceptor(ui),
	}
	s.server = grpc.NewServer(append(gopts, s.opts.serverOpts...)...)
	if s.opts.reflection {
		greflect.Register(s.server)
	}
	if s.opts.health {
		s.registerService(&grpc_health_v1.Health_ServiceDesc, health.NewServer())
	}
	if err := s.gateway(s.opts.gatewayOpts...); err != nil {
		return nil, err
	}
	if err := s.reactApp(); err != nil {
		return nil, err
	}
	// we do not configure grpc web here as the grpc handlers are not yet registered
	return s, nil
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) run() error {
	s.mu.Lock()
	s.closed = make(chan struct{})

	// configure grpc web now that we are ready to go
	if err := s.grpcWeb(s.opts.grpcWebOpts...); err != nil {
		return err
	}

	network := "tcp"
	if strings.HasPrefix(s.opts.address, "unix://") {
		network = "unix"
		s.opts.address = strings.TrimPrefix(s.opts.address, "unix://")
	}

	if s.opts.lis == nil {
		lis, err := net.Listen(network, s.opts.address)
		if err != nil {
			return err
		}
		if s.opts.tlsConfig != nil {
			lis = tls.NewListener(lis, s.opts.tlsConfig)
		}
		s.opts.lis = lis
		s.opts.address = lis.Addr().String()
	} else {
		s.opts.address = s.opts.lis.Addr().String()
	}

	mux := cmux.New(s.opts.lis)
	mux.SetReadTimeout(5 * time.Second)

	gLis := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	hList := mux.Match(cmux.Any())

	for i := range s.opts.beforeStart {
		if err := s.opts.beforeStart[i](); err != nil {
			s.mu.Unlock()
			return err
		}
	}

	if err := s.register(); err != nil {
		return err
	}
	s.running = true

	errs := make(chan error, 3)

	if reflect.DeepEqual(s.opts.cors, cors.Options{}) {
		s.opts.cors = cors.Options{
			AllowedHeaders: []string{"*"},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
				http.MethodHead,
			},
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		}
	}
	hServer := &http.Server{
		Handler: alice.New(s.opts.middlewares...).Then(cors.New(s.opts.cors).Handler(s.opts.mux)),
	}
	if s.opts.Gateway() || s.opts.grpcWeb || s.opts.hasReactUI {
		go func() {
			errs <- hServer.Serve(hList)
			hServer.Shutdown(s.opts.ctx)
		}()
	}
	go func() {
		errs <- s.server.Serve(gLis)
	}()

	go func() {
		if err := mux.Serve(); err != nil {
			// TODO(adphi): find more elegant solution
			if ignoreMuxError(err) {
				errs <- nil
				return
			}
			errs <- err
			return
		}
		errs <- nil
	}()
	for i := range s.opts.afterStart {
		if err := s.opts.afterStart[i](); err != nil {
			s.mu.Unlock()
			s.Stop()
			return err
		}
	}
	s.mu.Unlock()
	sigs := s.notify()
	select {
	case sig := <-sigs:
		fmt.Println()
		logger.C(s.opts.ctx).Warnf("received %v", sig)
		return s.Close()
	case err := <-errs:
		if err != nil && !ignoreMuxError(err) {
			logger.C(s.opts.ctx).Error(err)
			return err
		}
		return nil
	}
}

func (s *service) Start() error {
	return s.run()
}

func (s *service) Stop() error {
	log := logger.C(s.opts.ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	for i := range s.opts.beforeStop {
		if err := s.opts.beforeStop[i](); err != nil {
			return err
		}
	}
	if err := s.opts.registry.Deregister(s.regSvc); err != nil {
		log.Errorf("failed to deregister service: %v", err)
	}
	defer close(s.closed)
	sigs := s.notify()
	done := make(chan struct{})
	go func() {
		defer close(done)
		// TODO(adphi): find a better solution
		defer func() {
			// catch: Drain() is not implemented
			recover()
		}()
		log.Warn("shutting down gracefully")
		s.server.GracefulStop()
	}()
	select {
	case sig := <-sigs:
		fmt.Println()
		log.Warnf("received %v", sig)
		log.Warn("forcing shutdown")
		s.server.Stop()
	case <-done:
	}
	s.running = false
	s.cancel()
	for i := range s.opts.afterStop {
		if err := s.opts.afterStop[i](); err != nil {
			return err
		}
	}
	log.Info("server stopped")
	return nil
}

func (s *service) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registerService(desc, impl)
}

// serviceInfo wraps information about a service. It is very similar to
// ServiceDesc and is constructed from it for internal purposes.
type serviceInfo struct {
	// Contains the implementation for the methods in this service.
	serviceImpl interface{}
	methods     map[string]*grpc.MethodDesc
	streams     map[string]*grpc.StreamDesc
	mdata       interface{}
}

func (s *service) registerService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
	s.inproc.RegisterService(sd, ss)

	if _, ok := s.services[sd.ServiceName]; ok {
		logger.C(s.opts.ctx).Fatalf("grpc: Service.RegisterService found duplicate service registration for %q", sd.ServiceName)
	}
	info := &serviceInfo{
		serviceImpl: ss,
		methods:     make(map[string]*grpc.MethodDesc),
		streams:     make(map[string]*grpc.StreamDesc),
		mdata:       sd.Metadata,
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		info.methods[d.MethodName] = d
	}
	for i := range sd.Streams {
		d := &sd.Streams[i]
		info.streams[d.StreamName] = d
	}
	s.services[sd.ServiceName] = info
}

func (s *service) GetServiceInfo() map[string]grpc.ServiceInfo {
	ret := make(map[string]grpc.ServiceInfo)
	for n, srv := range s.services {
		methods := make([]grpc.MethodInfo, 0, len(srv.methods)+len(srv.streams))
		for m := range srv.methods {
			methods = append(methods, grpc.MethodInfo{
				Name:           m,
				IsClientStream: false,
				IsServerStream: false,
			})
		}
		for m, d := range srv.streams {
			methods = append(methods, grpc.MethodInfo{
				Name:           m,
				IsClientStream: d.ClientStreams,
				IsServerStream: d.ServerStreams,
			})
		}

		ret[n] = grpc.ServiceInfo{
			Methods:  methods,
			Metadata: srv.mdata,
		}
	}
	return ret
}

func (s *service) Close() error {
	err := multierr.Combine(s.Stop())
	<-s.closed
	return err
}

func (s *service) notify() <-chan os.Signal {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT)
	return sigs
}

func ignoreMuxError(err error) bool {
	if err == nil {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "mux: server closed")
}
