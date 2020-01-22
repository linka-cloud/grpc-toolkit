package main

import (
	"context"
	"fmt"
	"time"

	service2 "gitlab.bertha.cloud/partitio/grpc-service/service"
)

type GreeterHandler struct{}

func hello(name string) string {
	return fmt.Sprintf("Hello %s !", name)
}

func (g *GreeterHandler) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: hello(req.Name)}, nil
}

func (g *GreeterHandler) SayHelloStream(req *HelloStreamRequest, s Greeter_SayHelloStreamServer) error {
	for i := int64(0); i < req.Count; i++ {
		if err := s.Send(&HelloReply{Message: hello(req.Name)}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var svc service2.Service
	var err error
	svc, err = service2.New(
		service2.WithContext(ctx),
		service2.WithName("Greeting"),
		service2.WithAfterStart(func() error {
			fmt.Println("Server listening on", svc.Options().Address())
			return nil
		}),
		service2.WithAfterStop(func() error {
			fmt.Println("Stopping server")
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}
	RegisterGreeterServer(svc.Server(), &GreeterHandler{})
	if err := svc.Start(); err != nil {
		panic(err)
	}
}
