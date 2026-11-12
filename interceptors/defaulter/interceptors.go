package defaulter

import (
	"context"

	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

type interceptor struct{}

func NewInterceptors() interceptors.Interceptors {
	return &interceptor{}
}

func defaults(v any) {
	if d, ok := v.(interface{ Default() }); v != nil && ok {
		d.Default()
	}
}

func (i interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defaults(req)
		return handler(ctx, req)
	}
}

func (i interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		defaults(req)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{ServerStream: stream}
		return handler(srv, wrapper)
	}
}

func (i interceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		desc.Handler = (&sendWrapper{handler: desc.Handler}).Handler()
		return streamer(ctx, desc, cc, method)
	}
}

type recvWrapper struct {
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	defaults(m)
	return nil
}

type sendWrapper struct {
	grpc.ServerStream
	handler grpc.StreamHandler
}

func (s *sendWrapper) Handler() grpc.StreamHandler {
	return func(srv any, stream grpc.ServerStream) error {
		return s.handler(srv, s)
	}
}

func (s *sendWrapper) SendMsg(m any) error {
	defaults(m)
	return s.ServerStream.SendMsg(m)
}
