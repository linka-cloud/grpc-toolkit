package auth

import (
	"context"
	"testing"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	assert2 "github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.linka.cloud/grpc/errors"
)

func TestNotProtectededOnly(t *testing.T) {
	assert := assert2.New(t)
	i := &interceptor{o: options{ignoredMethods: []string{"/test.Service/ignored"}}}
	assert.False(i.isNotProtected("/test.Service/protected"))
	assert.True(i.isNotProtected("/test.Service/ignored"))
}

func TestProtectedOnly(t *testing.T) {
	assert := assert2.New(t)
	i := &interceptor{o: options{methods: []string{"/test.Service/protected"}}}
	assert.False(i.isNotProtected("/test.Service/protected"))
	assert.True(i.isNotProtected("/test.Service/ignored"))
}

func TestProtectedAndIgnored(t *testing.T) {
	assert := assert2.New(t)
	i := &interceptor{o: options{methods: []string{"/test.Service/protected"}, ignoredMethods: []string{"/test.Service/ignored"}}}
	assert.True(i.isNotProtected("/test.Service/ignored"))
	assert.False(i.isNotProtected("/test.Service/protected"))
	assert.True(i.isNotProtected("/test.Service/other"))
}

func TestProtectedByDefault(t *testing.T) {
	i := &interceptor{}
	assert2.False(t, i.isNotProtected("/test.Service/noop"))
	assert2.False(t, i.isNotProtected("/test.Service/method/cannotExists"))
	assert2.False(t, i.isNotProtected("/test.Service/validMethod"))
}

var (
	adminAuth = func(ctx context.Context, user, password string) (context.Context, error) {
		if user == "admin" && password == "admin" {
			return ctx, nil
		}
		return ctx, errors.PermissionDeniedf("")
	}
	testAuth = func(ctx context.Context, user, password string) (context.Context, error) {
		if user == "test" && password == "test" {
			return ctx, nil
		}
		return ctx, errors.PermissionDeniedf("")
	}
	tokenAuth = func(ctx context.Context, token string) (context.Context, error) {
		if token == "token" {
			return ctx, nil
		}
		return ctx, errors.PermissionDeniedf("")
	}
)

func TestChainedAuthFuncs(t *testing.T) {
	wantInternalError := false
	ctx := context.Background()
	auth := ChainedAuthFuncs([]grpc_auth.AuthFunc{
		makeBasicAuthFunc(adminAuth),
		makeBasicAuthFunc(testAuth),
		makeTokenAuthFunc(tokenAuth),
		makeTokenAuthFunc(func(ctx context.Context, token string) (context.Context, error) {
			if wantInternalError {
				return ctx, errors.Internalf("ooops")
			}
			return ctx, errors.Unauthenticatedf("")
		}),
	}...)

	tests := []struct {
		name          string
		auth          string
		internalError bool
		err           bool
		code          codes.Code
	}{
		{
			name: "no auth",
			auth: "",
			err:  true,
			code: codes.Unauthenticated,
		},
		{
			name: "valid token",
			auth: "bearer token",
		},
		{
			name: "empty bearer",
			auth: "bearer  ",
			err:  true,
			code: codes.PermissionDenied,
		},
		{
			name:          "internal error",
			auth:          "bearer internal",
			internalError: true,
			err:           true,
			code:          codes.PermissionDenied,
		},
		{
			name: "multiple auth: first basic valid",
			auth: BasicAuth("admin", "admin"),
		},
		{
			name: "multiple auth: second baisc valid",
			auth: BasicAuth("test", "test"),
		},
		{
			name: "invalid auth: bearer",
			auth: "bearer noop",
			err:  true,
			code: codes.PermissionDenied,
		},
		{
			name: "invalid auth: basic",
			auth: BasicAuth("other", "other"),
			err:  true,
			code: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantInternalError = tt.internalError
			rctx, err := auth(metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", tt.auth)))
			if tt.err {
				assert2.Error(t, err)
				s, ok := status.FromError(err)
				assert2.True(t, ok)
				assert2.Equal(t, tt.code, s.Code())
			}
			assert2.NotNil(t, rctx)
		})
	}
}
