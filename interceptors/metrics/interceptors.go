package metrics

import (
	"context"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/service"
)

func DefaultExemplarFromCtx(ctx context.Context) prometheus.Labels {
	if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
		return prometheus.Labels{"traceID": span.TraceID().String()}
	}
	return nil
}

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
}

type ClientInterceptors interface {
	interceptors.ClientInterceptors
}

type metrics struct {
	s *grpc_prometheus.ServerMetrics
	c *grpc_prometheus.ClientMetrics
	o *options
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

func NewInterceptors(opts ...Option) Interceptors {
	o := (&options{}).apply(opts...)
	s := grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerCounterOptions(o.copts...),
		grpc_prometheus.WithServerHandlingTimeHistogram(o.hopts...),
	)
	c := grpc_prometheus.NewClientMetrics(
		grpc_prometheus.WithClientCounterOptions(o.copts...),
		grpc_prometheus.WithClientHandlingTimeHistogram(o.hopts...),
	)
	m := &metrics{s: s, c: c, o: o}
	o.reg.MustRegister(m)
	return m
}

func NewServerInterceptors(opts ...Option) ServerInterceptors {
	o := (&options{}).apply(opts...)
	s := grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerCounterOptions(o.copts...),
		grpc_prometheus.WithServerHandlingTimeHistogram(o.hopts...),
	)
	m := &metrics{s: s, o: o}
	o.reg.MustRegister(m)
	return m
}

func NewClientInterceptors(opts ...Option) ClientInterceptors {
	o := (&options{}).apply(opts...)
	c := grpc_prometheus.NewClientMetrics(
		grpc_prometheus.WithClientCounterOptions(o.copts...),
		grpc_prometheus.WithClientHandlingTimeHistogram(o.hopts...),
	)
	m := &metrics{c: c, o: o}
	o.reg.MustRegister(m)
	return m
}

func (m *metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return m.s.UnaryServerInterceptor(grpc_prometheus.WithExemplarFromContext(m.o.fn))
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
