package metrics

import (
	"context"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

type ExemplarFromCtxFunc func(ctx context.Context) prometheus.Labels

type Option func(*options)

func WithCounterOptons(opts ...grpc_prometheus.CounterOption) Option {
	return func(o *options) {
		o.copts = append(o.copts, opts...)
	}
}

func WithHandlingTimeHistogram(opts ...grpc_prometheus.HistogramOption) Option {
	return func(o *options) {
		o.hopts = append(o.hopts, opts...)
	}
}

func WithHistogramOpts(opts ...grpc_prometheus.HistogramOption) Option {
	return func(o *options) {
		o.hopts = append(o.hopts, opts...)
	}
}

func WithExemplarFromContext(fn ExemplarFromCtxFunc) Option {
	return func(o *options) {
		o.fn = fn
	}
}

func WithRegisterer(reg prometheus.Registerer) Option {
	return func(o *options) {
		o.reg = reg
	}
}

type options struct {
	copts []grpc_prometheus.CounterOption
	hopts []grpc_prometheus.HistogramOption
	fn    func(ctx context.Context) prometheus.Labels
	reg   prometheus.Registerer
}

func (o *options) apply(opts ...Option) *options {
	for _, v := range opts {
		v(o)
	}
	if o.reg == nil {
		o.reg = prometheus.DefaultRegisterer
	}
	return o
}
