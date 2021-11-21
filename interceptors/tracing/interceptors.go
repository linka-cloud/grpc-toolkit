package tracing

import (
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc/interceptors"
)

type tracing struct {
	opts []otgrpc.Option
}

func NewInterceptors(opts ...otgrpc.Option) interceptors.Interceptors {
	return tracing{opts: opts}
}

func NewClientInterceptors(opts ...otgrpc.Option) interceptors.ClientInterceptors {
	return tracing{opts: opts}
}

func (t tracing) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer(), t.opts...)
}

func (t tracing) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer(), t.opts...)
}

func NewServerInterceptors(opts ...otgrpc.Option) interceptors.ServerInterceptors {
	return tracing{opts: opts}
}

func (t tracing) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer(), t.opts...)
}

func (t tracing) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer(), t.opts...)
}
