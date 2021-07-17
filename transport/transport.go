package transport

import (
	"net"

	"google.golang.org/grpc"
)

type Transport interface {
	grpc.ServiceRegistrar
	RegisterService(sd *grpc.ServiceDesc, ss interface{})
	Serve(lis net.Listener) error
	Stop()
	GracefulStop()
}
