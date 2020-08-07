package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/jinzhu/gorm"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	
	"gitlab.bertha.cloud/partitio/lab/grpc/certs"
)

/*
GLOBAL OPTIONS:
   	--client value                  Client for go-micro; rpc [$MICRO_CLIENT]
   	--client_request_timeout value  Sets the client request timeout. e.g 500ms, 5s, 1m. Default: 5s [$MICRO_CLIENT_REQUEST_TIMEOUT]
   	--client_retries value          Sets the client retries. Default: 1 (default: 1) [$MICRO_CLIENT_RETRIES]
   	--client_pool_size value        Sets the client connection pool size. Default: 1 (default: 0) [$MICRO_CLIENT_POOL_SIZE]
   	--client_pool_ttl value         Sets the client connection pool ttl. e.g 500ms, 5s, 1m. Default: 1m [$MICRO_CLIENT_POOL_TTL]
   	--help, -h                      show help

	--secure				  SECURE
	--ca_cert				  CA_CERT
	--server_cert			  SERVER_CERT
	--server_key			  SERVER_KEY
	
   	--register_ttl            REGISTER_TTL
	--register_interval       REGISTER_INTERVAL
	
	--server_address          SERVER_ADDRESS	   
   	--server_name             SERVER_NAME
	   
	--broker                  BROKER
   	--broker_address          BROKER_ADDRESS
	   
	--registry                REGISTRY
   	--registry_address        REGISTRY_ADDRESS
	   
	--db_path                 DB_PATH
*/

type Options interface {
	Context() context.Context
	Name() string
	Address() string
	Reflection() bool
	Secure() bool
	CACert() string
	Cert() string
	Key() string
	TLSConfig() *tls.Config
	DB() *gorm.DB
	BeforeStart() []func() error
	AfterStart() []func() error
	BeforeStop() []func() error
	AfterStop() []func() error
	ServerOpts() []grpc.ServerOption
	ServerInterceptors() []grpc.UnaryServerInterceptor
	StreamServerInterceptors() []grpc.StreamServerInterceptor
	ClientInterceptors() []grpc.UnaryClientInterceptor
	StreamClientInterceptors() []grpc.StreamClientInterceptor
	Defaults()
}

func NewOptions() *options {
	return &options{
		ctx:     context.Background(),
		address: ":0",
	}
}

func (o *options) Defaults() {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
	if o.address == "" {
		o.address = "0.0.0.0:0"
	}
}

type Option func(*options)

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
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

func WithReflection(r bool) Option {
	return func(o *options) {
		o.reflection = r
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

func WithDB(dialect string, args ...interface{}) Option {
	db, err := gorm.Open(dialect, args...)
	return func(o *options) {
		o.db = db
		o.error = multierr.Append(o.error, err)
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

func WithUnaryClientInterceptor(i ...grpc.UnaryClientInterceptor) Option {
	return func(o *options) {
		o.clientInterceptors = append(o.clientInterceptors, i...)
	}
}

// WithUnaryServerInterceptor adds unary Wrapper interceptors to the options passed into the server
func WithUnaryServerInterceptor(i ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.serverInterceptors = append(o.serverInterceptors, i...)
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

type options struct {
	ctx        context.Context
	name       string
	address    string
	secure     bool
	reflection bool
	caCert     string
	cert       string
	key        string
	tlsConfig  *tls.Config
	db         *gorm.DB

	beforeStart []func() error
	afterStart  []func() error
	beforeStop  []func() error
	afterStop   []func() error

	serverOpts []grpc.ServerOption

	serverInterceptors       []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor

	clientInterceptors       []grpc.UnaryClientInterceptor
	streamClientInterceptors []grpc.StreamClientInterceptor

	error error
}

func (o *options) Name() string {
	return o.name
}

func (o *options) Context() context.Context {
	return o.ctx
}

func (o *options) Address() string {
	return o.address
}

func (o *options) Reflection() bool {
	return o.reflection
}

func (o *options) Secure() bool {
	return o.secure
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

func (o *options) DB() *gorm.DB {
	return o.db
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
	return o.serverInterceptors
}

func (o *options) StreamServerInterceptors() []grpc.StreamServerInterceptor {
	return o.streamServerInterceptors
}

func (o *options) ClientInterceptors() []grpc.UnaryClientInterceptor {
	return o.clientInterceptors
}

func (o *options) StreamClientInterceptors() []grpc.StreamClientInterceptor {
	return o.streamClientInterceptors
}

func (o *options) parseTLSConfig() error {
	if (o.tlsConfig != nil) {
		return nil
	}
	if !o.hasTLSConfig() {
		if !o.secure {
			return nil
		}
		cert, err := certs.New(o.address, "localhost", "127.0.0.1", o.name)
		if err != nil {
			return err
		}
		o.tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		return nil
	}
	caCert, err := ioutil.ReadFile(o.caCert)
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
