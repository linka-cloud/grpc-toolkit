package resolver

import (
	"strings"

	"google.golang.org/grpc/resolver"

	"go.linka.cloud/grpc/registry"
)

func New(reg registry.Registry) resolver.Builder {
	return &builder{reg: reg}
}

type builder struct {
	reg registry.Registry
}

func (r builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	rslvr := &resolvr{reg: r.reg, target: target, cc: cc}
	go rslvr.run()
	return rslvr, nil
}

func (r builder) Scheme() string {
	return r.reg.String()
}

type resolvr struct {
	reg    registry.Registry
	target resolver.Target
	cc     resolver.ClientConn
	addrs  []resolver.Address
}

func (r *resolvr) run() {
	if r.reg.String() == "noop" {
		return
	}
	var name, version string
	parts := strings.Split(r.target.Endpoint, ":")
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	svc, err := r.reg.GetService(name)
	if err != nil {
		return
	}
	for _, v := range svc {
		if v.Name != name || v.Version != version {
			continue
		}
		for _, vv := range v.Nodes {
			r.addrs = append(r.addrs, resolver.Address{Addr: vv.Address})
		}
	}
	r.cc.UpdateState(resolver.State{Addresses: r.addrs})
	w, err := r.reg.Watch(registry.WatchService(r.target.Endpoint))
	if err != nil {
		return
	}
	defer w.Stop()
	for {
		res, err := w.Next()
		if err != nil {
			return
		}
		// TODO(adphi): implement
		switch res.Action {
		case "create":

		case "delete":

		}
		r.cc.UpdateState(resolver.State{Addresses: r.addrs})
	}
}

func (r *resolvr) ResolveNow(options resolver.ResolveNowOptions) {
	if r.reg.String() == "noop" {
		r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: r.target.Endpoint}}})
	}
}

func (r *resolvr) Close() {}
