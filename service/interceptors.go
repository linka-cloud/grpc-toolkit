package service

import (
	"context"
	"fmt"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	insecure2 "google.golang.org/grpc/credentials/insecure"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/interceptors/metadata"
)

func md(opts *options) interceptors.Interceptors {
	var pairs []string
	if opts.name != "" {
		pairs = append(pairs, "grpc-service-name", opts.name)
	}
	if opts.version != "" {
		pairs = append(pairs, "grpc-service-version", opts.version)
	}
	if len(pairs) != 0 {
		return metadata.NewInterceptors(pairs...)
	}
	return nil
}

func (s *service) wrapCC() grpc.ClientConnInterface {
	c, err := grpc.NewClient("internal", grpc.WithTransportCredentials(insecure2.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("failed to create fake grpc client: %v", err))
	}
	w := &client{ch: s.inproc, c: c}
	if len(s.opts.unaryClientInterceptors) != 0 {
		w.ui = grpc_middleware.ChainUnaryClient(s.opts.unaryClientInterceptors...)
	}
	if len(s.opts.streamClientInterceptors) != 0 {
		w.si = grpc_middleware.ChainStreamClient(s.opts.streamClientInterceptors...)
	}
	return w
}

type client struct {
	ui grpc.UnaryClientInterceptor
	si grpc.StreamClientInterceptor
	ch *inprocgrpc.Channel
	c  *grpc.ClientConn
}

func (c *client) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	if c.ui != nil {
		return c.ui(ctx, method, args, reply, c.c, c.invoke, opts...)
	}
	return c.ch.Invoke(ctx, method, args, reply, opts...)
}

func (c *client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if c.si != nil {
		return c.si(ctx, desc, c.c, method, c.stream, opts...)
	}
	return c.ch.NewStream(ctx, desc, method, opts...)
}

func (c *client) invoke(ctx context.Context, method string, req, reply any, _ *grpc.ClientConn, opts ...grpc.CallOption) error {
	return c.ch.Invoke(ctx, method, req, reply, opts...)
}

func (c *client) stream(ctx context.Context, desc *grpc.StreamDesc, _ *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.ch.NewStream(ctx, desc, method, opts...)
}
