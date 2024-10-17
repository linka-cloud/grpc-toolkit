package service

import (
	"net/http"
	"time"

	"github.com/traefik/grpc-web/go/grpcweb"

	"go.linka.cloud/grpc-toolkit/react"
)

var defaultWebOptions = []grpcweb.Option{
	grpcweb.WithWebsockets(true),
	grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
		return true
	}),
	grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	grpcweb.WithOriginFunc(func(origin string) bool {
		return true
	}),
	grpcweb.WithWebsocketPingInterval(time.Second),
}

func (s *service) grpcWeb(opts ...grpcweb.Option) error {
	if !s.opts.grpcWeb {
		return nil
	}
	h := grpcweb.WrapServer(s.server, append(defaultWebOptions, opts...)...)
	for _, v := range grpcweb.ListGRPCResources(s.server) {
		if s.opts.grpcWebPrefix != "" {
			s.lazyMux().Handle(s.opts.grpcWebPrefix+v, http.StripPrefix(s.opts.grpcWebPrefix, h))
		} else {
			s.lazyMux().Handle(v, h)
		}
	}
	return nil
}

func (s *service) reactApp() error {
	if !s.opts.hasReactUI {
		return nil
	}
	h, err := react.NewHandler(s.opts.reactUI, s.opts.reactUISubPath)
	if err != nil {
		return err
	}
	s.lazyMux().Handle("/", h)
	return nil
}
