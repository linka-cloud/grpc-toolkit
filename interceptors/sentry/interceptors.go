package sentry

import (
	"google.golang.org/grpc"

	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

type interceptor struct {
	opts []grpc_sentry.Option
}

func NewInterceptors(option ...grpc_sentry.Option) interceptors.Interceptors {
	return &interceptor{opts: option}
}

func (i *interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_sentry.UnaryServerInterceptor(i.opts...)
}

func (i *interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_sentry.StreamServerInterceptor(i.opts...)
}

func (i *interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return grpc_sentry.UnaryClientInterceptor(i.opts...)
}

func (i *interceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return grpc_sentry.StreamClientInterceptor(i.opts...)
}
