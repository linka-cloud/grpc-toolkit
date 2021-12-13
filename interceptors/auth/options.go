package auth

import (
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type Option func(o *options)

func WithMethods(methods ...string) Option {
	return func(o *options) {
		o.methods = append(o.methods, methods...)
	}
}

func WithIgnoredMethods(methods ...string) Option {
	return func(o *options) {
		o.ignoredMethods = append(o.ignoredMethods, methods...)
	}
}

func WithBasicValidators(validators ...BasicValidator) Option {
	var authFns []grpc_auth.AuthFunc
	for _, v := range validators {
		authFns = append(authFns, makeBasicAuthFunc(v))
	}
	return func(o *options) {
		o.authFns = append(o.authFns, authFns...)
	}
}

func WithTokenValidators(validators ...TokenValidator) Option {
	var authFns []grpc_auth.AuthFunc
	for _, v := range validators {
		authFns = append(authFns, makeTokenAuthFunc(v))
	}
	return func(o *options) {
		o.authFns = append(o.authFns, authFns...)
	}
}

func WithX509Validators(validators ...X509Validator) Option {
	var authFns []grpc_auth.AuthFunc
	for _, v := range validators {
		authFns = append(authFns, makeX509AuthFunc(v))
	}
	return func(o *options) {
		o.authFns = append(o.authFns, authFns...)
	}
}

type options struct {
	methods        []string
	ignoredMethods []string

	authFns []grpc_auth.AuthFunc
}
