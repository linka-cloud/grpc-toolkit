package client

import (
	"crypto/tls"

	"google.golang.org/grpc"

	"go.linka.cloud/grpc/registry"
)

type Options interface {
	Name() string
	Version() string
	Registry() registry.Registry
	TLSConfig() *tls.Config
	DialOptions() []grpc.DialOption
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

type options struct {
	registry    registry.Registry
	name        string
	version     string
	tlsConfig   *tls.Config
	secure      bool
	dialOptions []grpc.DialOption
}

func (o *options) Name() string {
	return o.name
}

func (o *options) Version() string {
	return o.version
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
