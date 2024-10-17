package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"

	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/interceptors/metadata"
)

type TokenValidator func(ctx context.Context, token string) (context.Context, error)

func makeTokenAuthFunc(v TokenValidator) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		a, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return ctx, err
		}
		return v(ctx, a)
	}
}

func NewBearerClientInterceptors(token string) interceptors.ClientInterceptors {
	return metadata.NewInterceptors("authorization", "bearer "+token)
}
