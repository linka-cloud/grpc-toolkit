package transport

import (
	"net"

	"google.golang.org/grpc"
)

type Transport interface {
	grpc.ServiceRegistrar
	Serve(lis net.Listener) error
	Stop()
	GracefulStop()
}
