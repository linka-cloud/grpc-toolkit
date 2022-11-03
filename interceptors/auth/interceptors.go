package auth

import (
	"context"
	"crypto/subtle"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"go.linka.cloud/grpc/interceptors"
)

func ChainedAuthFuncs(fn ...grpc_auth.AuthFunc) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		spb := status.New(codes.Unauthenticated, codes.Unauthenticated.String()).Proto()
		for _, v := range fn {
			ctx2, err := v(ctx)
			if err == nil {
				return ctx2, nil
			}
			s, ok := status.FromError(err)
			if !ok {
				return ctx2, err
			}
			if spb.Code != s.Proto().Code {
				spb.Code = s.Proto().Code
			}
			d, _ := anypb.New(s.Proto())
			spb.Details = append(spb.Details, d)
			spb.Message += ", " + s.Proto().Message
		}
		return ctx, status.FromProto(spb).Err()
	}
}

func NewServerInterceptors(opts ...Option) interceptors.ServerInterceptors {
	o := options{}
	for _, v := range opts {
		v(&o)
	}
	return &interceptor{o: o, authFn: ChainedAuthFuncs(o.authFns...)}
}

type interceptor struct {
	o      options
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
	for _, v := range i.o.ignoredMethods {
		if v == endpoint {
			return true
		}
	}
	if len(i.o.methods) == 0 {
		return false
	}
	for _, v := range i.o.methods {
		if v == endpoint {
			return false
		}
	}
	return true
}

func Equals(s1, s2 string) bool {
	return subtle.ConstantTimeCompare([]byte(s1), []byte(s2)) == 1
}
