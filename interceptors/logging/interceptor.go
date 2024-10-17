package logging

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/logger"
)

func New(ctx context.Context, opts ...logging.Option) interceptors.Interceptors {
	log := logger.C(ctx)
	return &interceptor{
		log: logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
			switch level {
			case logging.LevelDebug:
				log.WithReportCaller(true, 2).WithFields(fields...).Debug(msg)
			case logging.LevelInfo:
				log.WithReportCaller(true, 2).WithFields(fields...).Info(msg)
			case logging.LevelWarn:
				log.WithReportCaller(true, 2).WithFields(fields...).Warn(msg)
			case logging.LevelError:
				log.WithReportCaller(true, 2).WithFields(fields...).Error(msg)
			}
		}),
		opts: opts,
	}
}

type interceptor struct {
	log  logging.Logger
	opts []logging.Option
}

func (i *interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_middleware.ChainUnaryServer(
		logging.UnaryServerInterceptor(i.log, i.opts...),
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			log := logger.C(ctx)
			return handler(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), req)
		},
	)
}

func (i *interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_middleware.ChainStreamServer(
		logging.StreamServerInterceptor(i.log, i.opts...),
		func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			ctx := ss.Context()
			log := logger.C(ctx)
			return handler(srv, interceptors.NewContextServerStream(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), ss))
		},
	)
}

func (i *interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return grpc_middleware.ChainUnaryClient(
		logging.UnaryClientInterceptor(i.log, i.opts...),
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			log := logger.C(ctx)
			return invoker(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), method, req, reply, cc, opts...)
		},
	)
}

func (i *interceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return grpc_middleware.ChainStreamClient(
		logging.StreamClientInterceptor(i.log, i.opts...),
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			log := logger.C(ctx)
			return streamer(logger.Set(ctx, log.WithFields(logging.ExtractFields(ctx)...)), desc, cc, method, opts...)
		},
	)
}
