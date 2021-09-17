package service

import (
	"net"
	"strings"
	"time"

	"go.linka.cloud/grpc/registry"
	"go.linka.cloud/grpc/utils/addr"
	"go.linka.cloud/grpc/utils/backoff"
	net2 "go.linka.cloud/grpc/utils/net"
)

func (s *service) register() error {
	const (
		defaultRegisterInterval = time.Second * 30
		defaultRegisterTTL      = time.Second * 90
	)
	regFunc := func(service *registry.Service) error {
		var regErr error

		for i := 0; i < 3; i++ {
			// set the ttl
			rOpts := []registry.RegisterOption{registry.RegisterTTL(defaultRegisterTTL)}
			// attempt to register
			if err := s.opts.Registry().Register(service, rOpts...); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(backoff.Do(i + 1))
				continue
			}
			// success so nil error
			regErr = nil
			break
		}

		return regErr
	}

	var err error
	var advt, host, port string

	// // check the advertise address first
	// // if it exists then use it, otherwise
	// // use the address
	// if len(config.Advertise) > 0 {
	//	advt = config.Advertise
	// } else {
	advt = s.opts.address
	// }

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = s.opts.address
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// register service
	node := &registry.Node{
		Id:      s.opts.name + "-" + s.id,
		Address: net2.HostPort(addr, port),
	}

	s.regSvc = &registry.Service{
		Name:    s.opts.name,
		Version: s.opts.version,
		Nodes:   []*registry.Node{node},
	}

	// register the service
	if err := regFunc(s.regSvc); err != nil {
		return err
	}

	return nil
}
