package service

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var defaultGatewayOptions = []runtime.ServeMuxOption{
	runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
		return s, true
	}),
}

func (s *service) gateway(opts ...runtime.ServeMuxOption) error {
	if !s.opts.Gateway() {
		return nil
	}
	mux := runtime.NewServeMux(append(defaultGatewayOptions, opts...)...)
	if err := s.opts.gateway(s.opts.ctx, mux, s.inproc); err != nil {
		return err
	}
	if s.opts.gatewayPrefix != "" {
		s.mux.Handle(s.opts.gatewayPrefix+"/", http.StripPrefix(s.opts.gatewayPrefix, mux))
	} else {
		s.mux.Handle("/", mux)
	}
	return nil
}
