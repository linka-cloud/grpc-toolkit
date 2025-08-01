package recovery

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

func NewInterceptors(opts ...grpc_recovery.Option) interceptors.ServerInterceptors {
	return &recovery{opts: opts}
}

type recovery struct {
	opts []grpc_recovery.Option
}

func (i *recovery) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(i.opts...)
}

func (i *recovery) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_recovery.StreamServerInterceptor(i.opts...)
}
