package auth

import (
	"context"
)

type authKey struct{}

func Context[T any](ctx context.Context, auth T) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

func FromContext[T any](ctx context.Context) (T, bool) {
	auth, ok := ctx.Value(authKey{}).(T)
	return auth, ok
}

func MustTokenFromContext[T any](ctx context.Context) T {
	auth, ok := FromContext[T](ctx)
	if !ok {
		panic("no auth in context")
	}
	return auth
}
