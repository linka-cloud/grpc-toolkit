package ban

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
)

const (
	Unauthorized    = "Unauthorized"
	Unauthenticated = "Unauthenticated"
)

var (
	defaultOptions = options{
		cap:            1024,
		reaperInterval: 10 * time.Minute,
		rules:          defaultRules,
		actorFunc:      DefaultActorFunc,
	}

	defaultRules = []Rule{
		{
			Name:        Unauthorized,
			Message:     "Too many unauthorized requests",
			Code:        codes.PermissionDenied,
			StrikeLimit: 3,
			ExpireBase:  time.Second * 10,
			Sentence:    time.Second * 10,
		},
		{
			Name:        Unauthenticated,
			Message:     "Too many unauthenticated requests",
			Code:        codes.Unauthenticated,
			StrikeLimit: 3,
			ExpireBase:  time.Second * 10,
			Sentence:    time.Second * 10,
		},
	}
)

func DefaultActorFunc(ctx context.Context) (string, bool, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false, nil
	}
	if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
		return host, true, nil
	}
	return p.Addr.String(), true, nil
}

type Option func(*options)

func WithCapacity(cap int32) Option {
	return func(o *options) {
		o.cap = cap
	}
}

func WithRules(rules ...Rule) Option {
	return func(o *options) {
		o.rules = rules
	}
}

func WithReaperInterval(interval time.Duration) Option {
	return func(o *options) {
		o.reaperInterval = interval
	}
}

func WithActorFunc(f func(context.Context) (name string, found bool, err error)) Option {
	return func(o *options) {
		o.actorFunc = f
	}
}

type options struct {
	cap            int32
	rules          []Rule
	reaperInterval time.Duration
	actorFunc      func(ctx context.Context) (name string, found bool, err error)
}
