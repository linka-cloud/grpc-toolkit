package ban

import (
	"context"
)

type key struct{}

type value struct {
	ban   *ban
	actor string
	done  bool
}

func set(ctx context.Context, b *ban, actor string) context.Context {
	return context.WithValue(ctx, key{}, &value{ban: b, actor: actor})
}

func Infraction(ctx context.Context, rule string) error {
	v, ok := ctx.Value(key{}).(*value)
	if !ok {
		return nil
	}
	v.done = true
	return v.ban.s.Infraction(v.actor, rule)
}

func Actor(ctx context.Context) string {
	v, ok := ctx.Value(key{}).(*value)
	if !ok {
		return ""
	}
	return v.actor
}
