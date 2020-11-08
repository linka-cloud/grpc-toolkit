package client

import (
	"crypto/tls"

	"gitlab.bertha.cloud/partitio/lab/grpc/registry"
)

type Options interface {
	Version() string
	Registry() registry.Registry
	TLSConfig() *tls.Config
}

type Option func(*options)

func WithRegistry(registry registry.Registry) Option {
	return func(o *options) {
		o.registry = registry
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

type options struct {
	registry  registry.Registry
	version   string
	tlsConfig *tls.Config
	secure    bool
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

