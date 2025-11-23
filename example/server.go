package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"go.linka.cloud/grpc-tookit/example/pb"
	"go.linka.cloud/grpc-toolkit/interceptors/iface"
	"go.linka.cloud/grpc-toolkit/logger"
)

var (
	_ iface.UnaryInterceptor  = (*GreeterHandler)(nil)
	_ iface.StreamInterceptor = (*GreeterHandler)(nil)
)

type GreeterHandler struct {
	pb.UnimplementedGreeterServer
}

func hello(name string) string {
	return fmt.Sprintf("Hello %s !", name)
}

func (g *GreeterHandler) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("name", req.Name))
	logger.C(ctx).Infof("replying to %s", req.Name)
	if p, ok := peer.FromContext(ctx); ok {
		logger.C(ctx).Infof("peer auth info: %+v", p.AuthInfo)
	} else {
		logger.C(ctx).Infof("no peer info")
	}
	return &pb.HelloReply{Message: hello(req.Name)}, nil
}

func (g *GreeterHandler) SayHelloStream(req *pb.HelloStreamRequest, s pb.Greeter_SayHelloStreamServer) error {
	log := logger.C(s.Context())
	for i := int64(0); i < req.Count; i++ {
		log.Infof("sending message %d", i+1)
		if err := s.Send(&pb.HelloReply{Message: fmt.Sprintf("Hello %s (%d)!", req.Name, i+1)}); err != nil {
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
