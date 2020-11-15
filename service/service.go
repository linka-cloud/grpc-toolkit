package service

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"gitlab.bertha.cloud/partitio/lab/grpc/registry"
	"gitlab.bertha.cloud/partitio/lab/grpc/registry/noop"
	"gitlab.bertha.cloud/partitio/lab/grpc/utils/addr"
	"gitlab.bertha.cloud/partitio/lab/grpc/utils/backoff"
	net2 "gitlab.bertha.cloud/partitio/lab/grpc/utils/net"
)

type Service interface {
	Options() Options
	Server() *grpc.Server
	DB() *gorm.DB
	Start() error
	Stop() error
	Close() error
	Cmd() *cobra.Command
}

func New(opts ...Option) (Service, error) {
	return newService(opts...)
}

type service struct {
	cmd     *cobra.Command
	opts    *options
	cancel  context.CancelFunc
	server  *grpc.Server
	list    net.Listener
	mu      sync.Mutex
	running bool

	id     string
	regSvc *registry.Service
	closed chan struct{}
}

func newService(opts ...Option) (*service, error) {
	cmd.ParseFlags(os.Args)
	s := &service{
		opts: parseFlags(NewOptions()),
		cmd:  cmd,
		id:   uuid.New().String(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range opts {
		f(s.opts)
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
	var err error
	s.list, err = net.Listen("tcp", s.opts.address)
	if err != nil {
		return nil, err
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
	gopts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(s.opts.serverInterceptors...),
		),
	}
	if s.opts.tlsConfig != nil {
		gopts = append(gopts, grpc.Creds(credentials.NewTLS(s.opts.tlsConfig)))
	}
	s.server = grpc.NewServer(append(gopts, s.opts.serverOpts...)...)
	if s.opts.reflection {
		reflection.Register(s.server)
	}
	return s, nil
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) DB() *gorm.DB {
	return s.opts.db
}

func (s *service) Server() *grpc.Server {
	return s.server
}

func (s *service) Cmd() *cobra.Command {
	return s.cmd
}

func (s *service) register() error {
	const (
		defaultRegisterInterval = time.Second * 30
		defaultRegisterTTL      = time.Second * 90
	)
	regFunc := func(service *registry.Service) error {
		var regErr error

		for i := 0; i < 3; i++ {
			// set the ttl
			rOpts := []registry.RegisterOption{registry.RegisterTTL(defaultRegisterTTL)}
			// attempt to register
			if err := s.opts.Registry().Register(service, rOpts...); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(backoff.Do(i + 1))
				continue
			}
			// success so nil error
			regErr = nil
			break
		}

		return regErr
	}

	var err error
	var advt, host, port string

	//// check the advertise address first
	//// if it exists then use it, otherwise
	//// use the address
	//if len(config.Advertise) > 0 {
	//	advt = config.Advertise
	//} else {
		advt = s.opts.address
	//}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = s.opts.address
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// register service
	node := &registry.Node{
		Id:      s.opts.name + "-" + s.id,
		Address: net2.HostPort(addr, port),
	}

	s.regSvc = &registry.Service{
		Name: s.opts.name,
		Version:   s.opts.version,
		Nodes: []*registry.Node{node},
	}

	// register the service
	if err := regFunc(s.regSvc); err != nil {
		return err
	}

	return nil
}

func (s *service) run() error {
	s.mu.Lock()
	s.closed = make(chan struct{})
	for i := range s.opts.beforeStart {
		if err := s.opts.beforeStart[i](); err != nil {
			s.mu.Unlock()
			return err
		}
	}
	s.opts.address = s.list.Addr().String()

	if err := s.register(); err != nil {
		return err
	}
	s.running = true
	errs := make(chan error)
	go func() {
		errs <- s.server.Serve(s.list)
	}()
	for i := range s.opts.afterStart {
		if err := s.opts.afterStart[i](); err != nil {
			s.mu.Unlock()
			s.Stop()
			return err
		}
	}
	s.mu.Unlock()
	if err :=  <-errs; err != nil{
		logrus.Error(err)
		return err
	}
	return nil
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
	s.server.GracefulStop()
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

func (s *service) Close() error {
	err := multierr.Combine(s.Stop())
	if s.opts.db != nil {
		err = multierr.Append(s.opts.db.Close(), err)
	}
	<-s.closed
	return err
}
