package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

type ServerInterceptors interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

type ClientInterceptors interface {
	UnaryClientInterceptor() grpc.UnaryClientInterceptor
	StreamClientInterceptor() grpc.StreamClientInterceptor
}

type Interceptors interface {
	ServerInterceptors
	ClientInterceptors
}

func NewContextServerStream(ctx context.Context, ss grpc.ServerStream) grpc.ServerStream {
	return &ContextWrapper{ServerStream: ss, ctx: ctx}
}

type ContextWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *ContextWrapper) Context() context.Context {
	return w.ctx
}
