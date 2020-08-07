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
   	--register_ttl value            Register TTL in seconds (default: 0) [$MICRO_REGISTER_TTL]
   	--register_interval value       Register interval in seconds (default: 0) [$MICRO_REGISTER_INTERVAL]
   	--server value                  Server for go-micro; rpc [$MICRO_SERVER]
   	--server_name value             Name of the server. go.micro.srv.example [$MICRO_SERVER_NAME]
   	--server_version value          Version of the server. 1.1.0 [$MICRO_SERVER_VERSION]
   	--server_id value               Id of the server. Auto-generated if not specified [$MICRO_SERVER_ID]
   	--server_address value          Bind address for the server. 127.0.0.1:8080 [$MICRO_SERVER_ADDRESS]
   	--server_advertise value        Used instead of the server_address when registering with discovery. 127.0.0.1:8080 [$MICRO_SERVER_ADVERTISE]
   	--server_metadata value         A list of key-value pairs defining metadata. version=1.0.0 [$MICRO_SERVER_METADATA]
   	--broker value                  Broker for pub/sub. http, nats, rabbitmq [$MICRO_BROKER]
   	--broker_address value          Comma-separated list of broker addresses [$MICRO_BROKER_ADDRESS]
   	--registry value                Registry for discovery. consul, mdns [$MICRO_REGISTRY]
   	--registry_address value        Comma-separated list of registry addresses [$MICRO_REGISTRY_ADDRESS]
   	--selector value                Selector used to pick nodes for querying [$MICRO_SELECTOR]
   	--transport value               Transport mechanism used; http [$MICRO_TRANSPORT]
   	--transport_address value       Comma-separated list of transport addresses [$MICRO_TRANSPORT_ADDRESS]
   	--db_path value                 Path to sqlite db (e.g. /data/agents.db) (default: "agents.db") [$DB_PATH]
   	--help, -h                      show help

   	--register_ttl            REGISTER_TTL
   	--register_interval       REGISTER_INTERVAL
   	--server_name             SERVER_NAME
   	--server_version          SERVER_VERSION
   	--server_id               SERVER_ID
   	--server_advertise        SERVER_ADVERTISE
   	--broker                  BROKER
   	--broker_address          BROKER_ADDRESS
   	--registry                REGISTRY
   	--registry_address        REGISTRY_ADDRESS
   	--selector                SELECTOR
   	--transport               TRANSPORT
   	--transport_address       TRANSPORT_ADDRESS
   	--db_path                 DB_PATH

	--server_address          SERVER_ADDRESS
   	--ca_cert				  CA_CERT
	--server_cert			  SERVER_CERT
    --server_key			  SERVER_KEY
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

// Context specifies a context for the service.
// Can be used to signal shutdown of the service.
// Can be used for extra option values.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// Address sets the address of the server
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

// WrapHandler adds a handler Wrapper to a list of options passed into the server
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

// WrapSubscriber adds a subscriber Wrapper to a list of options passed into the server
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
