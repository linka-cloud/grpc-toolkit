package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.bertha.cloud/partitio/lab/grpc/client"
	"gitlab.bertha.cloud/partitio/lab/grpc/registry/mdns"
	"gitlab.bertha.cloud/partitio/lab/grpc/service"
)

type GreeterHandler struct{
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
		time.Sleep(time.Second)
	}
	return nil
}

func main() {
	name := "greeter"
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	ready := make(chan struct{})
	defer cancel()
	var svc service.Service
	var err error
	svc, err = service.New(
		service.WithContext(ctx),
		service.WithName(name),
		service.WithVersion("v0.0.1"),
		service.WithAddress("0.0.0.0:9991"),
		service.WithRegistry(mdns.NewRegistry()),
		service.WithReflection(true),
		service.WithSecure(true),
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
		client.WithSecure(true),
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
	cancel()
	<-done
}
