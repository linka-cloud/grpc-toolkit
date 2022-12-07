package ban

import (
	"time"

	"github.com/jaredfolkins/badactor"
	"google.golang.org/grpc/codes"
)

type ActionCallback func(action Action, actor string, rule *Rule) error

type Action int

const (
	Jailed Action = iota
	Released
)

func (a Action) String() string {
	switch a {
	case Jailed:
		return "Jailed"
	case Released:
		return "Released"
	default:
		return "Unknown"
	}
}

type Rule struct {
	Name         string
	Message      string
	Code         codes.Code
	StrikeLimit  int
	JailDuration time.Duration
	// Callback is an optional function to call when an Actor isJailed or released because of timeServed
	Callback ActionCallback
}

type action struct {
	fn ActionCallback
}

func (a2 *action) WhenJailed(a *badactor.Actor, r *badactor.Rule) error {
	if a2.fn == nil {
		return nil
	}
	return a2.fn(Jailed, a.Name(), &Rule{
		Name:         r.Name,
		Message:      r.Message,
		StrikeLimit:  r.StrikeLimit,
		JailDuration: r.ExpireBase,
	})
}

func (a2 *action) WhenTimeServed(a *badactor.Actor, r *badactor.Rule) error {
	if a2.fn == nil {
		return nil
	}
	return a2.fn(Released, a.Name(), &Rule{
		Name:         r.Name,
		Message:      r.Message,
		StrikeLimit:  r.StrikeLimit,
		JailDuration: r.ExpireBase,
	})
}
