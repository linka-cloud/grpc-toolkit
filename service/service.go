package service

import (
	"context"
	"net"
	"os"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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
}

func newService(opts ...Option) (*service, error) {
	cmd.ParseFlags(os.Args)
	s := &service{
		opts: parseFlags(NewOptions()),
		cmd:  cmd,
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
		grpc.Creds(credentials.NewTLS(s.opts.tlsConfig)),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(s.opts.serverInterceptors...),
		),
	}
	if s.opts.tlsConfig != nil {
		gopts = append(gopts)
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

func (s *service) run() error {
	s.mu.Lock()
	for i := range s.opts.beforeStart {
		if err := s.opts.beforeStart[i](); err != nil {
			return err
		}
	}
	s.running = true
	s.opts.address = s.list.Addr().String()
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
	return <-errs
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
	s.server.GracefulStop()
	s.cancel()
	s.running = false
	for i := range s.opts.afterStop {
		if err := s.opts.afterStop[i](); err != nil {

		}
	}
	return nil
}

func (s *service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := multierr.Combine(s.Stop())
	if s.opts.db != nil {
		err = multierr.Append(s.opts.db.Close(), err)
	}
	return err
}
