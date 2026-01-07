package pprof

import (
	"os"

	"github.com/grafana/pyroscope-go"
)

const (
	PyroscopeAddressEnv  = "PYROSCOPE_ADDRESS"
	PyroscopeUserEnv     = "PYROSCOPE_USER"
	PyroscopePasswordEnv = "PYROSCOPE_PASSWORD"
)

type Option func(*options)

func WithAddress(address string) Option {
	return func(o *options) {
		if address != "" {
			o.address = address
		}
	}
}

func WithUser(user string) Option {
	return func(o *options) {
		if user != "" {
			o.user = user
		}
	}
}

func WithPassword(password string) Option {
	return func(o *options) {
		if password != "" {
			o.password = password
		}
	}
}

func WithAddressEnv(env string) Option {
	return func(o *options) {
		if env != "" {
			o.addressEnv = env
		}
	}
}

func WithUserEnv(env string) Option {
	return func(o *options) {
		if env != "" {
			o.userEnv = env
		}
	}
}

func WithPasswordEnv(env string) Option {
	return func(o *options) {
		if env != "" {
			o.passwordEnv = env
		}
	}
}

func WithMutexProfileFraction(fraction int) Option {
	return func(o *options) {
		o.mutexProfileFraction = fraction
	}
}

func WithBlockProfileRate(rate int) Option {
	return func(o *options) {
		o.blockProfileRate = rate
	}
}

func WithProfiles(profiles ...pyroscope.ProfileType) Option {
	return func(o *options) {
		if len(profiles) != 0 {
			o.profiles = profiles
		}
	}
}

type options struct {
	address  string
	user     string
	password string

	addressEnv  string
	userEnv     string
	passwordEnv string

	mutexProfileFraction int
	blockProfileRate     int

	profiles []pyroscope.ProfileType
}

var defaultOptions = options{
	addressEnv:  PyroscopeAddressEnv,
	userEnv:     PyroscopeUserEnv,
	passwordEnv: PyroscopePasswordEnv,

	mutexProfileFraction: 5,
	blockProfileRate:     5,

	profiles: []pyroscope.ProfileType{
		pyroscope.ProfileCPU,
		pyroscope.ProfileInuseObjects,
		pyroscope.ProfileAllocObjects,
		pyroscope.ProfileInuseSpace,
		pyroscope.ProfileAllocSpace,
		pyroscope.ProfileGoroutines,
		pyroscope.ProfileMutexCount,
		pyroscope.ProfileMutexDuration,
		pyroscope.ProfileBlockCount,
		pyroscope.ProfileBlockDuration,
	},
}

func valueOrEnv(value, env string) string {
	if value != "" {
		return value
	}
	return os.Getenv(env)
}
