package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc/interceptors"
	"go.linka.cloud/grpc/service"
)

type Registerer interface {
	Register(svc service.Service)
}

type Interceptors interface {
	interceptors.Interceptors
	Registerer
}

type ServerInterceptors interface {
	interceptors.ServerInterceptors
	Registerer
}

type ClientInterceptors interface {
	interceptors.ClientInterceptors
}

type metrics struct {
	s *grpc_prometheus.ServerMetrics
	c *grpc_prometheus.ClientMetrics
}

func (m *metrics) Register(svc service.Service) {
	if m.s != nil {
		m.s.InitializeMetrics(svc)
	}
}

func NewInterceptors(opts ...grpc_prometheus.CounterOption) Interceptors {
	s := grpc_prometheus.NewServerMetrics(opts...)
	c := grpc_prometheus.NewClientMetrics(opts...)
	return &metrics{s: s, c: c}
}

func NewServerInterceptors(opts ...grpc_prometheus.CounterOption) ServerInterceptors {
	s := grpc_prometheus.NewServerMetrics(opts...)
	return &metrics{s: s}
}

func NewClientInterceptors(opts ...grpc_prometheus.CounterOption) ClientInterceptors {
	c := grpc_prometheus.NewClientMetrics(opts...)
	return &metrics{c: c}
}

func DefaultInterceptors() Interceptors {
	return &metrics{s: grpc_prometheus.DefaultServerMetrics, c: grpc_prometheus.DefaultClientMetrics}
}

func DefaultServerInterceptors() ServerInterceptors {
	return &metrics{s: grpc_prometheus.DefaultServerMetrics}
}

func DefaultClientInterceptors() ClientInterceptors {
	return &metrics{c: grpc_prometheus.DefaultClientMetrics}
}

func (m *metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return m.s.UnaryServerInterceptor()
}

func (m *metrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return m.s.StreamServerInterceptor()
}

func (m *metrics) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return m.c.UnaryClientInterceptor()
}

func (m *metrics) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return m.c.StreamClientInterceptor()
}
