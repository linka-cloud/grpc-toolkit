package otel

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"go.linka.cloud/grpc-toolkit/logger"
)

var dummy = newClient(&DSN{})

// Configure configures OpenTelemetry.
// By default, it:
//   - creates tracer provider;
//   - registers span exporter;
//   - sets tracecontext + baggage composite context propagator.
//
// You can use OTEL_DISABLED env var to completely skip otel configuration.
func Configure(ctx context.Context, opts ...Option) Provider {
	if _, ok := os.LookupEnv("OTEL_DISABLED"); ok {
		return dummy
	}

	log := logger.C(ctx)

	conf := newConfig(opts)

	if !conf.tracingEnabled && !conf.metricsEnabled && !conf.loggingEnabled {
		return dummy
	}

	if len(conf.dsn) == 0 {
		log.Warn("no DSN provided (otel-go is disabled)")
		return dummy
	}

	dsn, err := ParseDSN(conf.dsn[0])
	if err != nil {
		log.Warnf("invalid DSN: %s (otel is disabled)", err)
		return dummy
	}

	if strings.HasSuffix(dsn.Host, ":4318") {
		log.Warnf("otel-go uses OTLP/gRPC exporter, but got host %q", dsn.Host)
	}

	client := newClient(dsn)

	configurePropagator(conf)
	if conf.tracingEnabled {
		client.tp = configureTracing(ctx, conf)
	}
	if conf.metricsEnabled {
		client.mp = configureMetrics(ctx, conf)
	}
	if conf.loggingEnabled {
		client.lp = configureLogging(ctx, conf)
	}

	return client
}

func configurePropagator(conf *config) {
	textMapPropagator := conf.textMapPropagator
	if textMapPropagator == nil {
		textMapPropagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}
	otel.SetTextMapPropagator(textMapPropagator)
}
