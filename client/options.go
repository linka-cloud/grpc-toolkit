package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

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
	CA() string
	Cert() string
	Key() string
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

func WithCA(ca string) Option {
	return func(o *options) {
		o.caCert = ca
	}
}

func WithCert(cert string) Option {
	return func(o *options) {
		o.cert = cert
	}
}

func WithKey(key string) Option {
	return func(o *options) {
		o.key = key
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
	registry registry.Registry
	name     string
	version  string
	addr     string

	caCert      string
	cert        string
	key         string
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

func (o *options) CA() string {
	return o.caCert
}

func (o *options) Cert() string {
	return o.cert
}

func (o *options) Key() string {
	return o.key
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

func (o *options) hasTLSConfig() bool {
	return o.caCert != "" && o.cert != "" && o.key != "" && o.tlsConfig == nil
}

func (o *options) parseTLSConfig() error {
	if o.tlsConfig != nil {
		return nil
	}
	if !o.hasTLSConfig() {
		if !o.secure {
			return nil
		}
		o.tlsConfig = &tls.Config{InsecureSkipVerify: true}
		return nil
	}
	caCert, err := os.ReadFile(o.caCert)
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return fmt.Errorf("failed to load CA Cert from %s", o.caCert)
	}
	cert, err := tls.LoadX509KeyPair(o.cert, o.key)
	if err != nil {
		return err
	}
	o.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	return nil
}
