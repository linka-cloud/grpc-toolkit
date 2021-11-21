package interceptors

import (
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
