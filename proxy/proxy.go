// Copyright 2021 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package proxy

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.linka.cloud/grpc-toolkit/service"
)

// New sets up a simple proxy that forwards all requests to dst.
func New(dst grpc.ClientConnInterface, opts ...service.Option) (service.Service, error) {
	opts = append(opts, WithDefault(dst))
	// Set up the proxy server and then serve from it like in step one.
	return service.New(opts...)
}

// WithDefault returns a grpc.UnknownServiceHandler with a DefaultDirector.
func WithDefault(cc grpc.ClientConnInterface) service.Option {
	return service.WithGRPCServerOpts(grpc.UnknownServiceHandler(TransparentHandler(DefaultDirector(cc))))
}

// DefaultDirector returns a very simple forwarding StreamDirector that forwards all
// calls.
func DefaultDirector(cc grpc.ClientConnInterface) StreamDirector {
	return func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = metadata.NewOutgoingContext(ctx, md.Copy())
		return ctx, cc, nil
	}
}

func With(director StreamDirector) service.Option {
	return service.WithGRPCServerOpts(grpc.UnknownServiceHandler(TransparentHandler(director)))
}
