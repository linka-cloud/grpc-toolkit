package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"go.linka.cloud/grpc-toolkit/errors"
)

type X509Validator func(ctx context.Context, sans []string) (context.Context, error)

// func _(ctx context.Context) {
// 	p, ok := peer.FromContext(ctx)
// 	if !ok {
// 		return
// 	}
// 	i, ok := p.AuthInfo.(credentials.TLSInfo)
// 	if !ok {
// 		return
// 	}
// 	i.State.VerifiedChains
// }

func makeX509AuthFunc(v X509Validator) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return ctx, errors.Internalf("peer not found")
		}
		i, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok {
			return ctx, errors.Unauthenticatedf("no TLS credentials")
		}
		if !i.State.HandshakeComplete {
			return ctx, errors.Unauthenticatedf("handshake not complete")
		}
		var sans []string
		for _, v := range i.State.VerifiedChains {
			if len(v) == 0 {
				continue
			}
			sans = append(sans, v[0].PermittedDNSDomains...)
		}
		return v(ctx, sans)
	}
}
