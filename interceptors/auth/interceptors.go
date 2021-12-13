package auth

import (
	"context"
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.linka.cloud/grpc/interceptors"
)

func ChainedAuthFuncs(fn ...grpc_auth.AuthFunc) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		code := codes.Unauthenticated
		for _, v := range fn {
			ctx2, err := v(ctx)
			if err == nil {
				return ctx2, nil
			}
			s, ok := status.FromError(err)
			if !ok {
				return ctx2, err
			}
			if s.Code() == codes.PermissionDenied {
				code = codes.PermissionDenied
			}
		}
		return ctx, status.Error(code, code.String())
	}
}

func NewServerInterceptors(opts ...Option) interceptors.ServerInterceptors {
	o := options{}
	for _, v := range opts {
		v(&o)
	}
	return &interceptor{o: o, authFn: ChainedAuthFuncs(o.authFns...)}
}

type interceptor struct{
	o options
	authFn grpc_auth.AuthFunc
}

func (i *interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	a := grpc_auth.UnaryServerInterceptor(i.authFn)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if i.isNotProtected(info.FullMethod) {
			return handler(ctx, req)
		}
		return a(ctx, req, info, handler)
	}
}

func (i *interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	a := grpc_auth.StreamServerInterceptor(i.authFn)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if i.isNotProtected(info.FullMethod) {
			return handler(srv, ss)
		}
		return a(srv, ss, info, handler)
	}
}

func (i *interceptor) isNotProtected(endpoint string) bool {
	// default to not ignored
	if len(i.o.ignoredMethods) == 0 && len(i.o.methods) == 0 {
		return false
	}
	// endpoint is like /helloworld.Greeter/SayHello
	parts := strings.Split(strings.TrimPrefix(endpoint, "/"), "/")
	// invalid endpoint format
	if len(parts) != 2 {
		return false
	}
	method := parts[1]
	for _, v := range i.o.ignoredMethods {
		if v == method {
			return true
		}
	}
	if len(i.o.methods) == 0 {
		return false
	}
	for _, v := range i.o.methods {
		if v == method {
			return false
		}
	}
	return true
}
