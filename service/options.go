package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/certs"
	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/registry"
	"go.linka.cloud/grpc-toolkit/transport"
	"go.linka.cloud/grpc-toolkit/utils/addr"
)

var _ Options = (*options)(nil)

type RegisterGatewayFunc func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error

type Options interface {
	Context() context.Context
	Name() string
	Version() string
	Address() string

	Reflection() bool
	Health() bool

	CACert() string
	Cert() string
	Key() string
	TLSConfig() *tls.Config
	Secure() bool

	Registry() registry.Registry

	BeforeStart() []func() error
	AfterStart() []func() error
	BeforeStop() []func() error
	AfterStop() []func() error

	ServerOpts() []grpc.ServerOption
	ServerInterceptors() []grpc.UnaryServerInterceptor
	StreamServerInterceptors() []grpc.StreamServerInterceptor

	ClientInterceptors() []grpc.UnaryClientInterceptor
	StreamClientInterceptors() []grpc.StreamClientInterceptor

	Cors() cors.Options
	Mux() ServeMux
	GRPCWeb() bool
	GRPCWebPrefix() string
	GRPCWebOpts() []grpcweb.Option

	Gateway() bool
	GatewayPrefix() string
	GatewayOpts() []runtime.ServeMuxOption

	// TODO(adphi): metrics + tracing

	Default()
}

func NewOptions() *options {
	return &options{
		ctx:     context.Background(),
		address: ":0",
		health:  true,
	}
}

func (o *options) Default() {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
	if o.address == "" {
		o.address = "0.0.0.0:0"
	}
	if o.transport == nil {
		o.transport = &grpc.Server{}
	}

}

type Option func(*options)

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

func WithRegistry(registry registry.Registry) Option {
	return func(o *options) {
		o.registry = registry
	}
}

// WithContext specifies a context for the service.
// Can be used to signal shutdown of the service.
// Can be used for extra option values.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithAddress sets the address of the server
func WithAddress(addr string) Option {
	return func(o *options) {
		o.address = addr
	}
}

// WithListener specifies a listener for the service.
// It can be used to specify a custom listener.
// This will override the WithAddress and WithTLSConfig options
func WithListener(lis net.Listener) Option {
	return func(o *options) {
		o.lis = lis
	}
}

func WithReflection(r bool) Option {
	return func(o *options) {
		o.reflection = r
	}
}

func WithHealth(h bool) Option {
	return func(o *options) {
		o.health = h
	}
}

func WithSecure(s bool) Option {
	return func(o *options) {
		o.secure = s
	}
}

func WithGRPCServerOpts(opts ...grpc.ServerOption) Option {
	return func(o *options) {
		o.serverOpts = append(o.serverOpts, opts...)
	}
}

func WithCACert(path string) Option {
	return func(o *options) {
		o.caCert = path
	}
}

func WithCert(path string) Option {
	return func(o *options) {
		o.cert = path
	}
}

func WithKey(path string) Option {
	return func(o *options) {
		o.key = path
	}
}

func WithTLSConfig(conf *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = conf
	}
}

func WithBeforeStart(fn ...func() error) Option {
	return func(o *options) {
		o.beforeStart = append(o.beforeStart, fn...)
	}
}

func WithBeforeStop(fn ...func() error) Option {
	return func(o *options) {
		o.beforeStop = append(o.beforeStop, fn...)
	}
}

func WithAfterStart(fn ...func() error) Option {
	return func(o *options) {
		o.afterStart = append(o.afterStart, fn...)
	}
}

func WithAfterStop(fn ...func() error) Option {
	return func(o *options) {
		o.afterStop = append(o.afterStop, fn...)
	}
}

func WithInterceptors(i ...interceptors.Interceptors) Option {
	return func(o *options) {
		for _, v := range i {
			o.unaryServerInterceptors = append(o.unaryServerInterceptors, v.UnaryServerInterceptor())
			o.streamServerInterceptors = append(o.streamServerInterceptors, v.StreamServerInterceptor())
			o.unaryClientInterceptors = append(o.unaryClientInterceptors, v.UnaryClientInterceptor())
			o.streamClientInterceptors = append(o.streamClientInterceptors, v.StreamClientInterceptor())
		}
	}
}

func WithServerInterceptors(i ...interceptors.ServerInterceptors) Option {
	return func(o *options) {
		for _, v := range i {
			o.unaryServerInterceptors = append(o.unaryServerInterceptors, v.UnaryServerInterceptor())
			o.streamServerInterceptors = append(o.streamServerInterceptors, v.StreamServerInterceptor())
		}
	}
}

func WithClientInterceptors(i ...interceptors.ClientInterceptors) Option {
	return func(o *options) {
		for _, v := range i {
			o.unaryClientInterceptors = append(o.unaryClientInterceptors, v.UnaryClientInterceptor())
			o.streamClientInterceptors = append(o.streamClientInterceptors, v.StreamClientInterceptor())
		}
	}
}

func WithUnaryClientInterceptor(i ...grpc.UnaryClientInterceptor) Option {
	return func(o *options) {
		o.unaryClientInterceptors = append(o.unaryClientInterceptors, i...)
	}
}

// WithUnaryServerInterceptor adds unary Wrapper interceptors to the options passed into the server
func WithUnaryServerInterceptor(i ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryServerInterceptors = append(o.unaryServerInterceptors, i...)
	}
}

func WithStreamServerInterceptor(i ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamServerInterceptors = append(o.streamServerInterceptors, i...)
	}
}

func WithStreamClientInterceptor(i ...grpc.StreamClientInterceptor) Option {
	return func(o *options) {
		o.streamClientInterceptors = append(o.streamClientInterceptors, i...)
	}
}

// WithSubscriberInterceptor adds subscriber interceptors to the options passed into the server
func WithSubscriberInterceptor(w ...interface{}) Option {
	return func(o *options) {

	}
}

func WithCors(opts cors.Options) Option {
	return func(o *options) {
		o.cors = opts
	}
}

func WithMux(mux ServeMux) Option {
	return func(o *options) {
		o.mux = mux
	}
}

func WithMiddlewares(m ...Middleware) Option {
	return func(o *options) {
		o.middlewares = m
	}
}

func WithGRPCWeb(b bool) Option {
	return func(o *options) {
		o.grpcWeb = b
	}
}

func WithGRPCWebPrefix(prefix string) Option {
	return func(o *options) {
		o.grpcWebPrefix = strings.TrimSuffix(prefix, "/")
	}
}

func WithGRPCWebOpts(opts ...grpcweb.Option) Option {
	return func(o *options) {
		o.grpcWebOpts = opts
	}
}

func WithGateway(fn RegisterGatewayFunc) Option {
	return func(o *options) {
		o.gateway = fn
	}
}

func WithGatewayPrefix(prefix string) Option {
	return func(o *options) {
		o.gatewayPrefix = strings.TrimSuffix(prefix, "/")
	}
}

func WithGatewayOpts(opts ...runtime.ServeMuxOption) Option {
	return func(o *options) {
		o.gatewayOpts = opts
	}
}

// WithReactUI add static single page app serving to the http server
// subpath is the path in the read-only file embed.FS to use as root to serve
// static content
func WithReactUI(fs fs.FS, subpath string) Option {
	return func(o *options) {
		o.reactUI = fs
		o.reactUISubPath = subpath
		o.hasReactUI = true
	}
}

// WithoutCmux disables the use of cmux for http support to instead use grpc.Server.ServeHTTP method when http support is enabled
func WithoutCmux() Option {
	return func(o *options) {
		o.withoutCmux = true
	}
}

type options struct {
	ctx     context.Context
	name    string
	version string
	address string
	lis     net.Listener

	reflection bool
	health     bool

	secure    bool
	caCert    string
	cert      string
	key       string
	tlsConfig *tls.Config

	transport transport.Transport
	registry  registry.Registry

	beforeStart []func() error
	afterStart  []func() error
	beforeStop  []func() error
	afterStop   []func() error

	serverOpts []grpc.ServerOption

	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor

	unaryClientInterceptors  []grpc.UnaryClientInterceptor
	streamClientInterceptors []grpc.StreamClientInterceptor

	mux           ServeMux
	middlewares   []Middleware
	grpcWeb       bool
	grpcWebOpts   []grpcweb.Option
	grpcWebPrefix string
	gateway       RegisterGatewayFunc
	gatewayOpts   []runtime.ServeMuxOption
	cors          cors.Options

	reactUI        fs.FS
	reactUISubPath string
	hasReactUI     bool

	error         error
	gatewayPrefix string
	withoutCmux   bool
}

func (o *options) Name() string {
	return o.name
}

func (o *options) Version() string {
	return o.version
}

func (o *options) Context() context.Context {
	return o.ctx
}

func (o *options) Address() string {
	return o.address
}

func (o *options) Registry() registry.Registry {
	return o.registry
}

func (o *options) Reflection() bool {
	return o.reflection
}

func (o *options) Health() bool {
	return o.health
}

func (o *options) CACert() string {
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

func (o *options) BeforeStart() []func() error {
	return o.beforeStart
}

func (o *options) AfterStart() []func() error {
	return o.afterStart
}

func (o *options) BeforeStop() []func() error {
	return o.beforeStop
}

func (o *options) AfterStop() []func() error {
	return o.afterStop
}

func (o *options) ServerOpts() []grpc.ServerOption {
	return o.serverOpts
}

func (o *options) ServerInterceptors() []grpc.UnaryServerInterceptor {
	return o.unaryServerInterceptors
}

func (o *options) StreamServerInterceptors() []grpc.StreamServerInterceptor {
	return o.streamServerInterceptors
}

func (o *options) ClientInterceptors() []grpc.UnaryClientInterceptor {
	return o.unaryClientInterceptors
}

func (o *options) StreamClientInterceptors() []grpc.StreamClientInterceptor {
	return o.streamClientInterceptors
}

func (o *options) Cors() cors.Options {
	return o.cors
}

func (o *options) Mux() ServeMux {
	return o.mux
}

func (o *options) GRPCWeb() bool {
	return o.grpcWeb
}

func (o *options) GRPCWebPrefix() string {
	return o.grpcWebPrefix
}

func (o *options) GRPCWebOpts() []grpcweb.Option {
	return o.grpcWebOpts
}

func (o *options) Gateway() bool {
	return o.gateway != nil
}

func (o *options) GatewayPrefix() string {
	return o.gatewayPrefix
}

func (o *options) GatewayOpts() []runtime.ServeMuxOption {
	return o.gatewayOpts
}

func (o *options) WithoutCmux() bool {
	return o.withoutCmux
}

func (o *options) parseTLSConfig() error {
	if o.tlsConfig != nil {
		return nil
	}
	if !o.hasTLSConfig() {
		if !o.secure {
			return nil
		}
		var hosts []string
		if host, _, err := net.SplitHostPort(o.address); err == nil {
			if len(host) == 0 {
				hosts = addr.IPs()
			} else {
				hosts = []string{host}
			}
		}
		for i, h := range hosts {
			a, err := addr.Extract(h)
			if err != nil {
				return err
			}
			hosts[i] = a
		}
		// generate a certificate
		cert, err := certs.New(hosts...)
		if err != nil {
			return err
		}
		o.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
		}
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

func (o *options) hasTLSConfig() bool {
	return o.caCert != "" && o.cert != "" && o.key != "" && o.tlsConfig == nil
}
