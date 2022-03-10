package service

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mdInterceptors struct {
	k, v string
}

func (i mdInterceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := grpc.SetHeader(ctx, metadata.Pairs(i.k, i.v)); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i mdInterceptors) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := grpc.SetHeader(ss.Context(), metadata.Pairs(i.k, i.v)); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (i mdInterceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := grpc.SetHeader(ctx, metadata.Pairs(i.k, i.v)); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i mdInterceptors) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if err := grpc.SetHeader(ctx, metadata.Pairs(i.k, i.v)); err != nil {
			return nil, err
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
