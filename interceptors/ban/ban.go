package ban

import (
	"context"

	"github.com/jaredfolkins/badactor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.linka.cloud/grpc/interceptors"
	"go.linka.cloud/grpc/logger"
)

type ban struct {
	s     *badactor.Studio
	rules map[codes.Code]Rule
	actor func(ctx context.Context) (string, bool, error)
}

func NewInterceptors(opts ...Option) interceptors.ServerInterceptors {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	s := badactor.NewStudio(o.cap)
	rules := make(map[codes.Code]Rule)
	for _, r := range o.rules {
		rules[r.Code] = r
		s.AddRule(&badactor.Rule{
			Name:        r.Name,
			Message:     r.Message,
			StrikeLimit: r.StrikeLimit,
			ExpireBase:  r.ExpireBase,
			Sentence:    r.Sentence,
			Action: &action{
				whenJailed:     r.WhenJailed,
				whenTimeServed: r.WhenTimeServed,
			},
		})
	}
	// we ignore the error because CreateDirectors never returns an error
	_ = s.CreateDirectors(o.cap)
	s.StartReaper(o.reaperInterval)
	return &ban{s: s, rules: rules, actor: o.actorFunc}
}

func (b *ban) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		actor, ok, err := b.check(ctx)
		if err != nil {
			return nil, err
		}
		ctx = set(ctx, b, actor)
		if !ok {
			return handler(ctx, req)
		}
		for _, v := range b.rules {
			if b.s.IsJailedFor(actor, v.Name) {
				return nil, status.Error(v.Code, v.Message)
			}
		}
		res, err := handler(ctx, req)
		if err != nil {
			return nil, b.handleErr(ctx, actor, err)
		}
		return res, nil
	}
}

func (b *ban) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		actor, ok, err := b.check(ss.Context())
		if err != nil {
			return err
		}
		ss = interceptors.NewContextServerStream(set(ss.Context(), b, actor), ss)
		if !ok {
			return handler(srv, ss)
		}
		if err := handler(srv, ss); err != nil {
			return b.handleErr(ss.Context(), actor, err)
		}
		return nil
	}
}

func (b *ban) check(ctx context.Context) (actor string, ok bool, err error) {
	actor, ok, err = b.actor(ctx)
	if err != nil {
		return "", false, err
	}
	if !ok {
		return "", false, nil
	}
	for _, v := range b.rules {
		if b.s.IsJailedFor(actor, v.Name) {
			return actor, false, status.Error(v.Code, v.Message)
		}
	}
	return actor, true, nil
}

func (b *ban) handleErr(ctx context.Context, actor string, err error) error {
	v, ok := ctx.Value(key{}).(*value)
	if !ok || v.done {
		return err
	}
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	r, ok := b.rules[s.Code()]
	if !ok {
		return err
	}
	if err := b.s.Infraction(actor, r.Name); err != nil {
		logger.C(ctx).Warnf("%s: failed to add infraction: %v", r.Name, err)
	}
	return err
}
