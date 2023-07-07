package auth

import (
	"context"
	"encoding/base64"
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"go.linka.cloud/grpc-toolkit/errors"
	"go.linka.cloud/grpc-toolkit/interceptors"
	"go.linka.cloud/grpc-toolkit/interceptors/metadata"
)

func BasicAuth(user, password string) string {
	return "basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
}

type BasicValidator func(ctx context.Context, user, password string) (context.Context, error)

func makeBasicAuthFunc(v BasicValidator) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		a, err := grpc_auth.AuthFromMD(ctx, "basic")
		if err != nil {
			return ctx, err
		}
		c, err := base64.StdEncoding.DecodeString(a)
		if err != nil {
			return ctx, err
		}
		cs := string(c)
		s := strings.IndexByte(cs, ':')
		if s < 0 {
			return ctx, errors.Unauthenticatedf("malformed basic auth")
		}
		return v(ctx, cs[:s], cs[s+1:])
	}
}

func NewBasicAuthClientIntereptors(user, password string) interceptors.ClientInterceptors {
	return metadata.NewInterceptors("authorization", BasicAuth(user, password))
}
