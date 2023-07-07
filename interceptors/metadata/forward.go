package metadata

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

func NewForwardInterceptors() interceptors.ServerInterceptors {
	return &forward{}
}

type forward struct{}

func (f *forward) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = metadata.NewOutgoingContext(ctx, md.Copy())
		}
		return handler(ctx, req)
	}
}

func (f *forward) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md1, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(srv, ss)
		}
		o := md1.Copy()
		if md2, ok := metadata.FromOutgoingContext(ctx); ok {
			o = metadata.Join(o, md2.Copy())
		}
		return handler(srv, interceptors.NewContextServerStream(metadata.NewOutgoingContext(ctx, o), ss))
	}
}
