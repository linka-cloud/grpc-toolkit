package grpc

import (
	"go.linka.cloud/grpc/transport"
	"google.golang.org/grpc"
)

var (
	_ transport.Transport = &grpc.Server{}
)
