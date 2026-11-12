package iface

import (
	"context"

	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

type UnaryInterceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
}

type StreamInterceptor interface {
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

type iface struct{}

func New() interceptors.ServerInterceptors {
	return &iface{}
}

func (s iface) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if i, ok := info.Server.(UnaryInterceptor); ok {
			return i.UnaryServerInterceptor()(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}

func (s iface) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if i, ok := srv.(StreamInterceptor); ok {
			return i.StreamServerInterceptor()(srv, ss, info, handler)
		}
		return handler(srv, ss)
	}
}
