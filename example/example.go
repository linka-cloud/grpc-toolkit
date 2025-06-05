package main

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	"go.linka.cloud/grpc-tookit/example/pb"
	"go.linka.cloud/grpc-toolkit/client"
	"go.linka.cloud/grpc-toolkit/interceptors/auth"
	"go.linka.cloud/grpc-toolkit/interceptors/tracing"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/logger/otellog"
	"go.linka.cloud/grpc-toolkit/otel"
	"go.linka.cloud/grpc-toolkit/service"
)

func run(ctx context.Context, opts ...service.Option) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	name := "greeter"
	version := "v0.0.1"
	secure := true

	log := otellog.Setup(ctx, name, logger.PanicLevel, logger.FatalLevel, logger.ErrorLevel, logger.WarnLevel, logger.InfoLevel).WithReportCaller(true)
	ctx = logger.Set(ctx, log)
	done := make(chan struct{})
	ready := make(chan struct{})

	otel.Configure(
		// otel.WithDSN("http://127.0.0.1:4318"),
		otel.WithServiceName(name),
		otel.WithServiceVersion(version),
		otel.WithDeploymentEnvironment("tests"),
		otel.WithTraceSampler(sdktrace.AlwaysSample()),
		otel.WithMetricPrometheusBridge(),
	)
	defer otel.Shutdown(context.WithoutCancel(ctx))

	address := "0.0.0.0:9991"

	var svc service.Service
	opts = append(opts,
		service.WithContext(ctx),
		service.WithName(name),
		service.WithVersion(version),
		service.WithAddress(address),
		service.WithSecure(secure),
		service.WithAfterStart(func() error {
			log.Info("Server listening on", svc.Options().Address())
			close(ready)
			return nil
		}),
		service.WithAfterStop(func() error {
			log.Info("Stopping server")
			close(done)
			return nil
		}),
	)
	svc, err := newService(ctx, opts...)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := svc.Start(); err != nil {
			panic(err)
		}
	}()
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				// Opt into OpenMetrics e.g. to support exemplars.
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9992", nil); err != nil {
			panic(err)
		}
	}()
	<-ready
	copts := []client.Option{
		// client.WithName(name),
		// client.WithVersion(version),
		client.WithAddress("localhost:9991"),
		// client.WithRegistry(mdns.NewRegistry()),
		client.WithSecure(secure),
		client.WithInterceptors(tracing.NewClientInterceptors()),
		// client.WithUnaryInterceptors(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 	logger.From(ctx).WithFields("party", "client", "method", method).Info(req)
		// 	return invoker(ctx, method, req, reply, cc, opts...)
		// }),
	}
	s, err := client.New(copts...)
	if err != nil {
		log.Fatal(err)
	}
	g := pb.NewGreeterClient(s)
	h := grpc_health_v1.NewHealthClient(s)
	for i := 0; i < 5; i++ {
		_, err := h.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			log.Error(err)
		} else {
			log.Fatalf("expected error")
		}
	}
	log.Infof("waiting for unban")
	time.Sleep(time.Second)
	s, err = client.New(append(copts, client.WithInterceptors(auth.NewBasicAuthClientIntereptors("admin", "admin")))...)
	if err != nil {
		log.Fatal(err)
	}
	g = pb.NewGreeterClient(s)
	h = grpc_health_v1.NewHealthClient(s)
	hres, err := h.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "helloworld.Greeter"})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("status: %v", hres.Status)

	md := metadata.MD{}
	res, err := g.SayHello(ctx, &pb.HelloRequest{Name: "test"}, grpc.Header(&md))
	if err != nil {
		log.Fatal(err)
	}
	logMetadata(ctx, md)
	log.Infof("received message: %s", res.Message)
	md = metadata.MD{}
	res, err = g.SayHello(ctx, &pb.HelloRequest{}, grpc.Header(&md))
	if err == nil {
		log.Fatal("expected validation error")
	}
	logMetadata(ctx, md)
	stream, err := g.SayHelloStream(ctx, &pb.HelloStreamRequest{Name: "test"}, grpc.Header(&md))
	if err != nil {
		log.Fatal(err)
	}
	if md, err := stream.Header(); err == nil {
		logMetadata(ctx, md)
	}
	for {
		m, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("received stream message: %s", m.Message)
	}
	scheme := "http://"
	var (
		tlsConfig *tls.Config
		dial      func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error)
	)
	if secure {
		scheme = "https://"
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		dial = func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		}
	}
	httpc := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP:       true,
			TLSClientConfig: tlsConfig,
			DialTLSContext:  dial,
		},
	}
	req := `{"name":"test"}`

	do := func(url, contentType string) {
		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(req))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("content-type", contentType)
		req.Header.Set("authorization", auth.BasicAuth("admin", "admin"))
		resp, err := httpc.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(string(b))
	}
	do(scheme+address+"/rest/api/v1/greeter/hello", "application/json")
	do(scheme+address+"/grpc/helloworld.Greeter/SayHello", "application/grpc-web+json")

	if err := readSvcs(ctx, s); err != nil {
		log.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	cancel()
	<-done
}

func logMetadata(ctx context.Context, md metadata.MD) {
	log := logger.From(ctx)
	for k, v := range md {
		log.Infof("%s: %v", k, v)
	}
}
