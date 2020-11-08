package client

import (
	"crypto/tls"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"

	"gitlab.bertha.cloud/partitio/lab/grpc/registry/noop"
)

type Client interface {
	Dial(name string, version string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
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
	return c, nil
}

type client struct {
	pool *pool
	opts *options
}

func (c client) Dial(name, version string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if c.opts.tlsConfig == nil && c.opts.Secure() {
		c.opts.tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if c.opts.tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.opts.tlsConfig)))
	}
	addr := fmt.Sprintf("%s:///%s", c.opts.registry.String(), name)
	if version != "" {
		addr = addr + ":" + strings.TrimSpace(version)
	}
	pc, err := c.pool.getConn(addr, opts...)
	if err != nil {
	  return nil, err
	}
	return pc.ClientConn, nil
}
