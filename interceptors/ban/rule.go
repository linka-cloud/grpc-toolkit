package ban

import (
	"time"

	"github.com/jaredfolkins/badactor"
	"google.golang.org/grpc/codes"
)

type Rule struct {
	Name        string
	Message     string
	Code        codes.Code
	StrikeLimit int
	ExpireBase  time.Duration
	Sentence    time.Duration
	// WhenJailed is an optional function to call when an Actor isJailed
	WhenJailed func(actor string, r *Rule) error
	// WhenTimeServed is an optional function to call when an Actor is released because of timeServed
	WhenTimeServed func(actor string, r *Rule) error
}

type action struct {
	whenJailed     func(actor string, r *Rule) error
	whenTimeServed func(actor string, r *Rule) error
}

func (a2 *action) WhenJailed(a *badactor.Actor, r *badactor.Rule) error {
	if a2.whenJailed != nil {
		return a2.whenJailed(a.Name(), &Rule{
			Name:        r.Name,
			Message:     r.Message,
			StrikeLimit: r.StrikeLimit,
			ExpireBase:  r.ExpireBase,
			Sentence:    r.Sentence,
		})
	}
	return nil
}

func (a2 *action) WhenTimeServed(a *badactor.Actor, r *badactor.Rule) error {
	if a2.whenTimeServed != nil {
		return a2.whenTimeServed(a.Name(), &Rule{
			Name:        r.Name,
			Message:     r.Message,
			StrikeLimit: r.StrikeLimit,
			ExpireBase:  r.ExpireBase,
			Sentence:    r.Sentence,
		})
	}
	return nil
}
