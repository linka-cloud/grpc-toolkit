package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"

	"go.linka.cloud/grpc-toolkit/registry/noop"
)

type Client interface {
	grpc.ClientConnInterface
}

func New(opts ...Option) (Client, error) {
	c := &client{opts: &options{}}
	for _, o := range opts {
		o(c.opts)
	}
	if c.opts.registry == nil {
		c.opts.registry = noop.New()
	}
	resolver.Register(c.opts.registry.ResolverBuilder())
	c.pool = newPool(DefaultPoolSize, DefaultPoolTTL, DefaultPoolMaxIdle, DefaultPoolMaxStreams)
	if c.opts.tlsConfig == nil && c.opts.Secure() {
		c.opts.tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if c.opts.tlsConfig != nil {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(c.opts.tlsConfig)))
	}
	if !c.opts.secure && c.opts.tlsConfig == nil {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if len(c.opts.unaryInterceptors) > 0 {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.opts.unaryInterceptors...)))
	}
	if len(c.opts.streamInterceptors) > 0 {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(c.opts.streamInterceptors...)))
	}
	switch {
	case c.opts.addr == "":
		c.addr = fmt.Sprintf("%s:///%s", c.opts.registry.String(), c.opts.name)
	case strings.HasPrefix(c.opts.addr, "tcp://"):
		c.addr = strings.Replace(c.opts.addr, "tcp://", "", 1)
	case strings.HasPrefix(c.opts.addr, "unix:///"):
		c.addr = c.opts.addr
	case strings.HasPrefix(c.opts.addr, "unix://"):
		c.addr = strings.Replace(c.opts.addr, "unix://", "unix:", 1)
	default:
		c.addr = c.opts.addr
	}
	if c.opts.version != "" && c.opts.addr == "" {
		c.addr = c.addr + ":" + strings.TrimSpace(c.opts.version)
	}
	return c, nil
}

type client struct {
	addr string
	pool *pool
	opts *options
}

func (c *client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	pc, err := c.pool.getConn(c.addr, c.opts.dialOptions...)
	if err != nil {
		return err
	}
	return pc.Invoke(ctx, method, args, reply, opts...)
}

func (c *client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	pc, err := c.pool.getConn(c.addr, c.opts.dialOptions...)
	if err != nil {
		return nil, err
	}
	return pc.NewStream(ctx, desc, method, opts...)
}
