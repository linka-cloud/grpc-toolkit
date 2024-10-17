package auth

import (
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
)

type Option func(o *options)

// WithMethods change the behaviour to not protect by default, it takes a list of fully qualified method names to protect, e.g. /helloworld.Greeter/SayHello
func WithMethods(methods ...string) Option {
	return func(o *options) {
		o.methods = append(o.methods, methods...)
	}
}

// WithIgnoredMethods bypass auth for the given methods, it takes a list of fully qualified method name, e.g. /helloworld.Greeter/SayHello
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
