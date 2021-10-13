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
	"github.com/jinzhu/gorm"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	greflect "google.golang.org/grpc/reflection"

	"go.linka.cloud/grpc/registry"
	"go.linka.cloud/grpc/registry/noop"
)

type Service interface {
	Options() Options
	DB() *gorm.DB
	Start() error
	Stop() error
	Close() error
	Cmd() *cobra.Command

	RegisterService(desc *grpc.ServiceDesc, impl interface{})
}

func New(opts ...Option) (Service, error) {
	return newService(opts...)
}

type service struct {
	cmd    *cobra.Command
	opts   *options
	cancel context.CancelFunc

	server  *grpc.Server
	mu      sync.Mutex
	running bool

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	id     string
	regSvc *registry.Service
	closed chan struct{}
}

func newService(opts ...Option) (*service, error) {
	if err := cmd.ParseFlags(os.Args); err != nil {
		return nil, err
	}
	s := &service{
		opts:   parseFlags(NewOptions()),
		cmd:    cmd,
		id:     uuid.New().String(),
		inproc: &inprocgrpc.Channel{},
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range opts {
		f(s.opts)
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
	cmd.Use = s.opts.name
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Use == "" {
			cmd.Use = os.Args[0]
		}
		return s.run()
	}
	ui := grpcmiddleware.ChainUnaryServer(s.opts.serverInterceptors...)
	s.inproc = s.inproc.WithServerUnaryInterceptor(ui)

	si := grpcmiddleware.ChainStreamServer(/*TODO(adphi): add to options*/)
	s.inproc = s.inproc.WithServerStreamInterceptor(si)

	gopts := []grpc.ServerOption{
		grpc.StreamInterceptor(si),
		grpc.UnaryInterceptor(ui),
	}
	s.server = grpc.NewServer(append(gopts, s.opts.serverOpts...)...)
	if s.opts.reflection {
		greflect.Register(s.server)
	}
	if err := s.gateway(s.opts.gatewayOpts...); err != nil {
		return nil, err
	}
	// we do not configure grpc web here as the grpc handlers are not yet registered
	return s, nil
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) DB() *gorm.DB {
	return s.opts.db
}

func (s *service) Cmd() *cobra.Command {
	return s.cmd
}

func (s *service) run() error {
	s.mu.Lock()
	s.closed = make(chan struct{})

	// configure grpc web now that we are ready to go
	if err := s.grpcWeb(s.opts.grpcWebOpts...); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", s.opts.address)
	if err != nil {
		return err
	}
	if s.opts.tlsConfig != nil {
		lis = tls.NewListener(lis, s.opts.tlsConfig)
	}

	s.opts.address = lis.Addr().String()

	mux := cmux.New(lis)
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
			AllowedHeaders:   []string{"*"},
			AllowedMethods:   []string{
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
	if s.opts.Gateway() || s.opts.grpcWeb {
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
		logrus.Warnf("received %v", sig)
		return s.Close()
	case err := <-errs:
		if err != nil && !ignoreMuxError(err){
			logrus.Error(err)
			return err
		}
		return nil
	}
}

func (s *service) Start() error {
	return s.cmd.Execute()
}

func (s *service) Stop() error {
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
		logrus.Errorf("failed to deregister service: %v", err)
	}
	defer close(s.closed)
	sigs := s.notify()
	done := make(chan struct{})
	go func() {
		logrus.Warn("shutting down gracefully")
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case sig := <-sigs:
		fmt.Println()
		logrus.Warnf("received %v", sig)
		logrus.Warn("forcing shutdown")
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
	logrus.Info("server stopped")
	return nil
}

func (s *service) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.inproc.RegisterService(desc, impl)
}

func (s *service) Close() error {
	err := multierr.Combine(s.Stop())
	if s.opts.db != nil {
		err = multierr.Append(s.opts.db.Close(), err)
	}
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
