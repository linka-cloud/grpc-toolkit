package otellog

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	otellogrus2 "go.opentelemetry.io/contrib/bridges/otellogrus"

	"go.linka.cloud/grpc-toolkit/logger"
)

func Setup(ctx context.Context, name string, levels ...logger.Level) logger.Logger {
	log := logger.C(ctx).WithFields("name", name)
	log.Logger().AddHook(otellogrus2.NewHook(name, otellogrus2.WithLevels(levels)))
	log.Logger().AddHook(otellogrus.NewHook(otellogrus.WithLevels(levels...)))
	return log
}
