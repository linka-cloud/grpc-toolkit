package pprof

import (
	"context"
	"runtime"

	"github.com/grafana/pyroscope-go"

	"go.linka.cloud/grpc-toolkit/logger"
)

func Init(ctx context.Context, app string, opts ...Option) {
	if app == "" {
		panic("application name is required to start pyroscope profiler")
	}
	o := defaultOptions
	for _, v := range opts {
		v(&o)
	}
	if valueOrEnv(o.address, o.addressEnv) == "" {
		return
	}
	runtime.SetMutexProfileFraction(o.mutexProfileFraction)
	runtime.SetBlockProfileRate(o.blockProfileRate)
	log := logger.C(ctx).WithFields("service", "pyroscope")
	log.Info("starting pyroscope profiler")
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   app,
		ServerAddress:     valueOrEnv(o.address, o.addressEnv),
		BasicAuthUser:     valueOrEnv(o.user, o.userEnv),
		BasicAuthPassword: valueOrEnv(o.password, o.passwordEnv),
		Logger:            log,
		ProfileTypes:      o.profiles,
	})
	if err != nil {
		log.WithError(err).Error("failed to start pyroscope")
	}
}
