package recovery

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

type interceptors struct {
	opts grpc_recovery.Option
}

func (i *interceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(i.opts)
}

func (i *interceptors) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_recovery.StreamServerInterceptor(i.opts)
}

func (i *interceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	panic("not implemented")
}

func (i *interceptors) StreamClientInterceptor() grpc.StreamClientInterceptor {
	panic("not implemented")
}



