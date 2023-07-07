package client

import (
	"crypto/tls"

	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/registry"
)

type Options interface {
	Name() string
	Version() string
	Address() string
	Secure() bool
	Registry() registry.Registry
	TLSConfig() *tls.Config
	DialOptions() []grpc.DialOption
	UnaryInterceptors() []grpc.UnaryClientInterceptor
	StreamInterceptors() []grpc.StreamClientInterceptor
}

type Option func(*options)

func WithRegistry(registry registry.Registry) Option {
	return func(o *options) {
		o.registry = registry
	}
}

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

func WithAddress(address string) Option {
	return func(o *options) {
		o.addr = address
	}
}

func WithTLSConfig(conf *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = conf
	}
}

func WithSecure(s bool) Option {
	return func(o *options) {
		o.secure = s
	}
}

func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(o *options) {
		o.dialOptions = opts
	}
}

func WithInterceptors(i ...interceptors.ClientInterceptors) Option {
	return func(o *options) {
		for _, v := range i {
			o.unaryInterceptors = append(o.unaryInterceptors, v.UnaryClientInterceptor())
			o.streamInterceptors = append(o.streamInterceptors, v.StreamClientInterceptor())
		}
	}
}

func WithUnaryInterceptors(i ...grpc.UnaryClientInterceptor) Option {
	return func(o *options) {
		o.unaryInterceptors = append(o.unaryInterceptors, i...)
	}
}

func WithStreamInterceptors(i ...grpc.StreamClientInterceptor) Option {
	return func(o *options) {
		o.streamInterceptors = append(o.streamInterceptors, i...)
	}
}

type options struct {
	registry    registry.Registry
	name        string
	version     string
	addr        string
	tlsConfig   *tls.Config
	secure      bool
	dialOptions []grpc.DialOption

	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
}

func (o *options) Name() string {
	return o.name
}

func (o *options) Version() string {
	return o.version
}

func (o *options) Address() string {
	return o.addr
}

func (o *options) Registry() registry.Registry {
	return o.registry
}

func (o *options) TLSConfig() *tls.Config {
	return o.tlsConfig
}

func (o *options) Secure() bool {
	return o.secure
}

func (o *options) DialOptions() []grpc.DialOption {
	return o.dialOptions
}

func (o *options) UnaryInterceptors() []grpc.UnaryClientInterceptor {
	return o.unaryInterceptors
}

func (o *options) StreamInterceptors() []grpc.StreamClientInterceptor {
	return o.streamInterceptors
}
