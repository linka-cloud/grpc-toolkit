package grpc

import (
	"google.golang.org/grpc"

	"go.linka.cloud/grpc/transport"
)

var (
	_ transport.Transport = &grpc.Server{}
)

func New() transport.Transport {
	return grpc.NewServer()
}
