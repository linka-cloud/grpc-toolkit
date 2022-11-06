package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc/interceptors"
	"go.linka.cloud/grpc/service"
)

type Registerer interface {
	Register(svc service.Service)
}

type Interceptors interface {
	ServerInterceptors
	ClientInterceptors
}

type ServerInterceptors interface {
	Registerer
	interceptors.ServerInterceptors
	prometheus.Collector
	EnableHandlingTimeHistogram(opts ...grpc_prometheus.HistogramOption)

	EnableClientHandlingTimeHistogram(opts ...grpc_prometheus.HistogramOption)
	EnableClientStreamReceiveTimeHistogram(opts ...grpc_prometheus.HistogramOption)
	EnableClientStreamSendTimeHistogram(opts ...grpc_prometheus.HistogramOption)
}

type ClientInterceptors interface {
	interceptors.ClientInterceptors
}

type metrics struct {
	s *grpc_prometheus.ServerMetrics
	c *grpc_prometheus.ClientMetrics
}

func (m *metrics) EnableHandlingTimeHistogram(opts ...grpc_prometheus.HistogramOption) {
	if m.s != nil {
		if m.s == grpc_prometheus.DefaultServerMetrics {
			grpc_prometheus.EnableHandlingTimeHistogram(opts...)
		} else {
			m.s.EnableHandlingTimeHistogram(opts...)
		}
	}
}

func (m *metrics) EnableClientHandlingTimeHistogram(opts ...grpc_prometheus.HistogramOption) {
	if m.c != nil {
		if m.c == grpc_prometheus.DefaultClientMetrics {
			grpc_prometheus.EnableClientHandlingTimeHistogram(opts...)
		} else {
			m.c.EnableClientHandlingTimeHistogram(opts...)
		}
	}
}

func (m *metrics) EnableClientStreamReceiveTimeHistogram(opts ...grpc_prometheus.HistogramOption) {
	if m.c != nil {
		if m.c == grpc_prometheus.DefaultClientMetrics {
			grpc_prometheus.EnableClientStreamReceiveTimeHistogram(opts...)
		} else {
			m.c.EnableClientStreamReceiveTimeHistogram(opts...)
		}
	}
}

func (m *metrics) EnableClientStreamSendTimeHistogram(opts ...grpc_prometheus.HistogramOption) {
	if m.c != nil {
		if m.c == grpc_prometheus.DefaultClientMetrics {
			grpc_prometheus.EnableClientStreamSendTimeHistogram(opts...)
		} else {
			m.c.EnableClientStreamSendTimeHistogram(opts...)
		}
	}
}

func (m *metrics) Describe(descs chan<- *prometheus.Desc) {
	if m.s != nil {
		m.s.Describe(descs)
	}
}

func (m *metrics) Collect(c chan<- prometheus.Metric) {
	if m.s != nil {
		m.s.Collect(c)
	}
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
