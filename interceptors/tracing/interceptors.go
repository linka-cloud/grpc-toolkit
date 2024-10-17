package tracing

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

type tracing struct {
	opts []otelgrpc.Option
}

func NewInterceptors(opts ...otelgrpc.Option) interceptors.Interceptors {
	return tracing{opts: opts}
}

func NewClientInterceptors(opts ...otelgrpc.Option) interceptors.ClientInterceptors {
	return tracing{opts: opts}
}

func (t tracing) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otelgrpc.UnaryClientInterceptor(t.opts...)
}

func (t tracing) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return otelgrpc.StreamClientInterceptor(t.opts...)
}

func NewServerInterceptors(opts ...otelgrpc.Option) interceptors.ServerInterceptors {
	return tracing{opts: opts}
}

func (t tracing) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(t.opts...)
}

func (t tracing) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor(t.opts...)
}
