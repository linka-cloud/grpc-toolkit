package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"

	"go.linka.cloud/grpc-toolkit/interceptors/chain"
	"go.linka.cloud/grpc-toolkit/registry/noop"
)

type Client interface {
	grpc.ClientConnInterface
}

func New(opts ...Option) (Client, error) {
	c := &client{opts: &options{dialOptions: []grpc.DialOption{grpc.WithContextDialer(dial)}}}
	for _, o := range opts {
		o(c.opts)
	}
	if c.opts.registry == nil {
		c.opts.registry = noop.New()
	}
	resolver.Register(c.opts.registry.ResolverBuilder())
	if err := c.opts.parseTLSConfig(); err != nil {
		return nil, err
	}
	if c.opts.tlsConfig != nil {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(c.opts.tlsConfig)))
	}
	if !c.opts.secure && c.opts.tlsConfig == nil {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if len(c.opts.unaryInterceptors) > 0 {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithUnaryInterceptor(chain.UnaryClient(c.opts.unaryInterceptors...)))
	}
	if len(c.opts.streamInterceptors) > 0 {
		c.opts.dialOptions = append(c.opts.dialOptions, grpc.WithStreamInterceptor(chain.StreamClient(c.opts.streamInterceptors...)))
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
	cc, err := grpc.Dial(c.addr, c.opts.dialOptions...)
	if err != nil {
		return nil, err
	}
	c.cc = cc
	return c, nil
}

type client struct {
	addr string
	opts *options
	cc   *grpc.ClientConn
}

func (c *client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return c.cc.Invoke(ctx, method, args, reply, opts...)
}

func (c *client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.cc.NewStream(ctx, desc, method, opts...)
}

func parseDialTarget(target string) (string, string) {
	net := "tcp"
	m1 := strings.Index(target, ":")
	m2 := strings.Index(target, ":/")
	// handle unix:addr which will fail with url.Parse
	if m1 >= 0 && m2 < 0 {
		if n := target[0:m1]; n == "unix" {
			return n, target[m1+1:]
		}
	}
	if strings.HasPrefix(target, `\\.\pipe\`) {
		net = "pipe"
		return net, target
	}
	if m2 >= 0 {
		t, err := url.Parse(target)
		if err != nil {
			return net, target
		}
		scheme := t.Scheme
		addr := t.Host
		if scheme == "unix" {
			addr += t.Path
		}
		return scheme, addr
	}
	return net, target
}
