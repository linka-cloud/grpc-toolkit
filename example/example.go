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

	"github.com/sirupsen/logrus"

	"go.linka.cloud/grpc/client"
	"go.linka.cloud/grpc/registry/mdns"
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

func main() {
	name := "greeter"
	secure := true
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	ready := make(chan struct{})
	defer cancel()
	var svc service.Service
	var err error
	address := "0.0.0.0:9991"
	svc, err = service.New(
		service.WithContext(ctx),
		service.WithName(name),
		service.WithVersion("v0.0.1"),
		service.WithAddress(address),
		service.WithRegistry(mdns.NewRegistry()),
		service.WithReflection(true),
		service.WithSecure(secure),
		service.WithAfterStart(func() error {
			fmt.Println("Server listening on", svc.Options().Address())
			close(ready)
			return nil
		}),
		service.WithAfterStop(func() error {
			fmt.Println("Stopping server")
			close(done)
			return nil
		}),
		service.WithGateway(RegisterGreeterHandler),
		service.WithGatewayPrefix("/rest"),
		service.WithGRPCWeb(true),
		service.WithGRPCWebPrefix("/grpc"),
	)
	if err != nil {
		panic(err)
	}
	RegisterGreeterServer(svc, &GreeterHandler{})
	go func() {
		if err := svc.Start(); err != nil {
			panic(err)
		}
	}()
	<-ready
	s, err := client.New(
		client.WithName("greeter"),
		client.WithVersion("v0.0.1"),
		client.WithRegistry(mdns.NewRegistry()),
		client.WithSecure(secure),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	g := NewGreeterClient(s)
	defer cancel()
	res, err := g.SayHello(context.Background(), &HelloRequest{Name: "test"})
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("received message: %s", res.Message)
	stream, err := g.SayHelloStream(context.Background(), &HelloStreamRequest{Name: "test", Count: 10})
	if err != nil {
		logrus.Fatal(err)
	}
	for {
		m, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("received stream message: %s", m.Message)
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
			logrus.Fatal(err)
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info(string(b))
	}
	do(scheme+address+"/rest/api/v1/greeter/hello",  "application/json")
	do(scheme+address+"/grpc/helloworld.Greeter/SayHello", "application/grpc-web+json")
	cancel()
	<-done
}

