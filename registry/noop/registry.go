package noop

import (
	"errors"

	"google.golang.org/grpc/resolver"

	"go.linka.cloud/grpc-toolkit/registry"
	resolver2 "go.linka.cloud/grpc-toolkit/resolver"
)

func New() registry.Registry {
	return &noop{}
}

type noop struct{}

func (n noop) ResolverBuilder() resolver.Builder {
	return resolver2.New(n)
}

func (n noop) Init(option ...registry.Option) error {
	return nil
}

func (n noop) Options() registry.Options {
	return registry.Options{}
}

func (n noop) Register(service *registry.Service, option ...registry.RegisterOption) error {
	return nil
}

func (n noop) Deregister(service *registry.Service, option ...registry.DeregisterOption) error {
	return nil
}

func (n noop) GetService(s string, option ...registry.GetOption) ([]*registry.Service, error) {
	return nil, nil
}

func (n noop) ListServices(option ...registry.ListOption) ([]*registry.Service, error) {
	return nil, nil
}

func (n noop) Watch(option ...registry.WatchOption) (registry.Watcher, error) {
	return nil, errors.New("watch not supported")
}

func (n noop) String() string {
	return "noop"
}
