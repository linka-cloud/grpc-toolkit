package defaulter

import (
	"context"

	"google.golang.org/grpc"

	"go.linka.cloud/grpc/interceptors"
)

type interceptor struct{}

func NewInterceptors() interceptors.Interceptors {
	return &interceptor{}
}

func defaults(v interface{}) {
	if d, ok := v.(interface{ Default() }); v != nil && ok {
		d.Default()
	}
}

func (i interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defaults(req)
		return handler(ctx, req)
	}
}

func (i interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		defaults(req)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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

func (s *recvWrapper) RecvMsg(m interface{}) error {
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
	return func(srv interface{}, stream grpc.ServerStream) error {
		return s.handler(srv, s)
	}
}

func (s *sendWrapper) SendMsg(m interface{}) error {
	defaults(m)
	return s.ServerStream.SendMsg(m)
}
