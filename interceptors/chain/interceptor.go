package chain

import (
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
)

type Option func(*chain)

func WithInterceptors(i ...interceptors.Interceptors) Option {
	return func(c *chain) {
		for _, i := range i {
			if i := i.UnaryServerInterceptor(); i != nil {
				c.usi = append(c.usi, i)
			}
			if i := i.StreamServerInterceptor(); i != nil {
				c.ssi = append(c.ssi, i)
			}
			if i := i.UnaryClientInterceptor(); i != nil {
				c.uci = append(c.uci, i)
			}
			if i := i.StreamClientInterceptor(); i != nil {
				c.sci = append(c.sci, i)
			}
		}
	}
}

func WithServerInterceptors(si ...interceptors.ServerInterceptors) Option {
	return func(c *chain) {
		for _, i := range si {
			if i := i.UnaryServerInterceptor(); i != nil {
				c.usi = append(c.usi, i)
			}
			if i := i.StreamServerInterceptor(); i != nil {
				c.ssi = append(c.ssi, i)
			}
		}
	}
}

func WithClientInterceptors(ci ...interceptors.ClientInterceptors) Option {
	return func(c *chain) {
		for _, i := range ci {
			if i := i.UnaryClientInterceptor(); i != nil {
				c.uci = append(c.uci, i)
			}
			if i := i.StreamClientInterceptor(); i != nil {
				c.sci = append(c.sci, i)
			}
		}
	}
}

func WithUnaryServerInterceptors(usi ...grpc.UnaryServerInterceptor) Option {
	return func(c *chain) {
		for _, i := range usi {
			if i != nil {
				c.usi = append(c.usi, i)
			}
		}
	}
}

func WithStreamServerInterceptors(ssi ...grpc.StreamServerInterceptor) Option {
	return func(c *chain) {
		for _, i := range ssi {
			if i != nil {
				c.ssi = append(c.ssi, i)
			}
		}
	}
}

func WithUnaryClientInterceptors(uci ...grpc.UnaryClientInterceptor) Option {
	return func(c *chain) {
		for _, i := range uci {
			if i != nil {
				c.uci = append(c.uci, i)
			}
		}
	}
}

func WithStreamClientInterceptors(sci ...grpc.StreamClientInterceptor) Option {
	return func(c *chain) {
		for _, i := range sci {
			if i != nil {
				c.sci = append(c.sci, i)
			}
		}
	}
}

func New(opts ...Option) interceptors.Interceptors {
	c := &chain{}
	for _, o := range opts {
		o(c)
	}
	return c
}

type chain struct {
	usi []grpc.UnaryServerInterceptor
	ssi []grpc.StreamServerInterceptor
	uci []grpc.UnaryClientInterceptor
	sci []grpc.StreamClientInterceptor
}

func (c *chain) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return UnaryServer(c.usi...)
}

func (c *chain) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return StreamServer(c.ssi...)
}

func (c *chain) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return UnaryClient(c.uci...)
}

func (c *chain) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return StreamClient(c.sci...)
}
