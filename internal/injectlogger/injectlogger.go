package injectlogger

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/logger"
)

func New(ctx context.Context) interceptors.Interceptors {
	return &interceptor{
		ctx: ctx,
	}
}

type interceptor struct {
	ctx context.Context
}

func (i *interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log := logger.C(i.ctx)
		return handler(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), req)
	}
}

func (i *interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		log := logger.C(i.ctx)
		return handler(srv, interceptors.NewContextServerStream(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), ss))
	}
}

func (i *interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log := logger.C(i.ctx)
		return invoker(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), method, req, reply, cc, opts...)
	}
}

func (i *interceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log := logger.C(i.ctx)
		return streamer(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), desc, cc, method, opts...)
	}
}
