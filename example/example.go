package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	greflectsvc "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"go.linka.cloud/grpc/client"
	"go.linka.cloud/grpc/interceptors/defaulter"
	metrics2 "go.linka.cloud/grpc/interceptors/metrics"
	validation2 "go.linka.cloud/grpc/interceptors/validation"
	"go.linka.cloud/grpc/logger"
	"go.linka.cloud/grpc/service"
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
	metrics := metrics2.NewInterceptors()
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
		service.WithGateway(RegisterGreeterHandler),
		service.WithGatewayPrefix("/rest"),
		service.WithGRPCWeb(true),
		service.WithGRPCWebPrefix("/grpc"),
		service.WithMiddlewares(httpLogger),
		service.WithInterceptors(metrics, defaulter, validation),
	)
	svc, err = service.New(opts...)
	if err != nil {
		panic(err)
	}
	RegisterGreeterServer(svc, &GreeterHandler{})
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
	s, err := client.New(
		// client.WithName(name),
		// client.WithVersion(version),
		client.WithAddress("localhost:9991"),
		// client.WithRegistry(mdns.NewRegistry()),
		client.WithSecure(secure),
		client.WithUnaryInterceptors(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			logger.From(ctx).WithFields("party", "client", "method", method).Info(req)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	g := NewGreeterClient(s)
	defer cancel()
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
	if secure {
		scheme = "https://"
	}
	httpc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req := `{"name":"test"}`

	do := func(url, contentType string) {
		resp, err := httpc.Post(url, contentType, strings.NewReader(req))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
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
