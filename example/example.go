package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	greflectsvc "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"go.linka.cloud/grpc-toolkit/client"
	"go.linka.cloud/grpc-toolkit/interceptors/auth"
	"go.linka.cloud/grpc-toolkit/interceptors/ban"
	"go.linka.cloud/grpc-toolkit/interceptors/defaulter"
	"go.linka.cloud/grpc-toolkit/interceptors/iface"
	metrics2 "go.linka.cloud/grpc-toolkit/interceptors/metrics"
	validation2 "go.linka.cloud/grpc-toolkit/interceptors/validation"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/service"
)

var (
	_ iface.UnaryInterceptor  = (*GreeterHandler)(nil)
	_ iface.StreamInterceptor = (*GreeterHandler)(nil)
)

type GreeterHandler struct {
	UnimplementedGreeterServer
}

func hello(name string) string {
	return fmt.Sprintf("Hello %s !", name)
}

func (g *GreeterHandler) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: hello(req.Name)}, nil
}

func (g *GreeterHandler) SayHelloStream(req *HelloStreamRequest, s Greeter_SayHelloStreamServer) error {
	for i := int64(0); i < req.Count; i++ {
		if err := s.Send(&HelloReply{Message: fmt.Sprintf("Hello %s (%d)!", req.Name, i+1)}); err != nil {
			return err
		}
		// time.Sleep(time.Second)
	}
	return nil
}

func (g *GreeterHandler) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.C(ctx).Infof("called service interface unary interceptor")
		return handler(ctx, req)
	}
}

func (g *GreeterHandler) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.C(ss.Context()).Infof("called service interface stream interceptor")
		return handler(srv, ss)
	}
}

func httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		log := logger.From(request.Context()).WithFields(
			"method", request.Method,
			"host", request.Host,
			"path", request.URL.Path,
			"remoteAddress", request.RemoteAddr,
		)
		next.ServeHTTP(writer, request)
		log.WithField("duration", time.Since(start)).Info()
	})
}

func main() {
	f, opts := service.NewFlagSet()
	cmd := &cobra.Command{
		Use: "example",
		Run: func(cmd *cobra.Command, args []string) {
			run(opts)
		},
	}
	cmd.Flags().AddFlagSet(f)
	cmd.Execute()
}

func run(opts ...service.Option) {
	name := "greeter"
	version := "v0.0.1"
	secure := true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logger.New().WithFields("service", name)
	ctx = logger.Set(ctx, log)
	done := make(chan struct{})
	ready := make(chan struct{})
	var svc service.Service
	var err error
	metrics := metrics2.DefaultInterceptors()
	validation := validation2.NewInterceptors(true)
	defaulter := defaulter.NewInterceptors()
	address := "0.0.0.0:9991"
	opts = append(opts, service.WithContext(ctx),
		service.WithName(name),
		service.WithVersion(version),
		service.WithAddress(address),
		// service.WithRegistry(mdns.NewRegistry()),
		service.WithReflection(true),
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
		service.WithoutCmux(),
		service.WithGateway(RegisterGreeterHandler),
		service.WithGatewayPrefix("/rest"),
		service.WithGRPCWeb(true),
		service.WithGRPCWebPrefix("/grpc"),
		service.WithMiddlewares(httpLogger),
		service.WithInterceptors(metrics),
		service.WithServerInterceptors(
			ban.NewInterceptors(ban.WithDefaultJailDuration(time.Second), ban.WithDefaultCallback(func(action ban.Action, actor string, rule *ban.Rule) error {
				log.WithFields("action", action, "actor", actor, "rule", rule.Name).Info("ban callback")
				return nil
			})),
			auth.NewServerInterceptors(auth.WithBasicValidators(func(ctx context.Context, user, password string) (context.Context, error) {
				if !auth.Equals(user, "admin") || !auth.Equals(password, "admin") {
					return ctx, fmt.Errorf("invalid user or password")
				}
				log.Infof("request authenticated")
				return ctx, nil
			})),
		),
		service.WithInterceptors(defaulter, validation),
		// enable server interface interceptor
		service.WithServerInterceptors(iface.New()),
	)
	svc, err = service.New(opts...)
	if err != nil {
		panic(err)
	}
	RegisterGreeterServer(svc, &GreeterHandler{})
	metrics.EnableHandlingTimeHistogram()
	metrics.Register(svc)
	go func() {
		if err := svc.Start(); err != nil {
			panic(err)
		}
	}()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
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
		client.WithUnaryInterceptors(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			logger.From(ctx).WithFields("party", "client", "method", method).Info(req)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	}
	s, err := client.New(copts...)
	if err != nil {
		log.Fatal(err)
	}
	g := NewGreeterClient(s)
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
	g = NewGreeterClient(s)
	h = grpc_health_v1.NewHealthClient(s)
	hres, err := h.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "helloworld.Greeter"})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("status: %v", hres.Status)

	md := metadata.MD{}
	res, err := g.SayHello(ctx, &HelloRequest{Name: "test"}, grpc.Header(&md))
	if err != nil {
		log.Fatal(err)
	}
	logMetadata(ctx, md)
	log.Infof("received message: %s", res.Message)
	md = metadata.MD{}
	res, err = g.SayHello(ctx, &HelloRequest{}, grpc.Header(&md))
	if err == nil {
		log.Fatal("expected validation error")
	}
	logMetadata(ctx, md)
	stream, err := g.SayHelloStream(ctx, &HelloStreamRequest{Name: "test"}, grpc.Header(&md))
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
	cancel()
	<-done
}

func logMetadata(ctx context.Context, md metadata.MD) {
	log := logger.From(ctx)
	for k, v := range md {
		log.Infof("%s: %v", k, v)
	}
}

func readSvcs(ctx context.Context, c client.Client) (err error) {
	log := logger.From(ctx)
	rc := greflectsvc.NewServerReflectionClient(c)
	rstream, err := rc.ServerReflectionInfo(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := rstream.CloseSend(); err2 != nil && err == nil {
			err = err2
		}
	}()
	if err = rstream.Send(&greflectsvc.ServerReflectionRequest{MessageRequest: &greflectsvc.ServerReflectionRequest_ListServices{}}); err != nil {
		return err
	}
	var rres *greflectsvc.ServerReflectionResponse
	rres, err = rstream.Recv()
	if err != nil {
		return err
	}
	rlist, ok := rres.MessageResponse.(*greflectsvc.ServerReflectionResponse_ListServicesResponse)
	if !ok {
		return fmt.Errorf("unexpected reflection response type: %T", rres.MessageResponse)
	}
	for _, v := range rlist.ListServicesResponse.Service {
		if v.Name == "grpc.reflection.v1alpha.ServerReflection" {
			continue
		}
		parts := strings.Split(v.Name, ".")
		if len(parts) < 2 {
			return fmt.Errorf("malformed service name: %s", v.Name)
		}
		pkg := strings.Join(parts[:len(parts)-1], ".")
		svc := parts[len(parts)-1]
		if err = rstream.Send(&greflectsvc.ServerReflectionRequest{MessageRequest: &greflectsvc.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: v.Name,
		}}); err != nil {
			return err
		}
		rres, err = rstream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		rfile, ok := rres.MessageResponse.(*greflectsvc.ServerReflectionResponse_FileDescriptorResponse)
		if !ok {
			return fmt.Errorf("unexpected reflection response type: %T", rres.MessageResponse)
		}
		fdps := make(map[string]*descriptorpb.DescriptorProto)
		var sdp *descriptorpb.ServiceDescriptorProto
		for _, v := range rfile.FileDescriptorResponse.FileDescriptorProto {
			fdp := &descriptorpb.FileDescriptorProto{}
			if err = proto.Unmarshal(v, fdp); err != nil {
				return err
			}
			for _, s := range fdp.GetService() {
				if fdp.GetPackage() == pkg && s.GetName() == svc {
					if sdp != nil {
						log.Warnf("service already found: %s.%s", fdp.GetPackage(), s.GetName())
						continue
					}
					sdp = s
				}
			}
			for _, m := range fdp.GetMessageType() {
				fdps[fdp.GetPackage()+"."+m.GetName()] = m
			}
		}
		if sdp == nil {
			return fmt.Errorf("%s: service not found", v.Name)
		}
		for _, m := range sdp.GetMethod() {
			log.Infof("%s: %s", v.Name, m.GetName())
		}
	}
	return nil
}
