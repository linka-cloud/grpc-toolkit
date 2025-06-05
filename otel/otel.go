package otel

import (
	"context"
	"os"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"go.linka.cloud/grpc-toolkit/logger"
)

var log = logger.StandardLogger().WithField("name", "otel")

// Configure configures OpenTelemetry.
// By default, it:
//   - creates tracer provider;
//   - registers span exporter;
//   - sets tracecontext + baggage composite context propagator.
//
// You can use OTEL_DISABLED env var to completely skip otel configuration.
func Configure(opts ...Option) {
	if _, ok := os.LookupEnv("OTEL_DISABLED"); ok {
		return
	}

	ctx := context.TODO()
	conf := newConfig(opts)

	if !conf.tracingEnabled && !conf.metricsEnabled && !conf.loggingEnabled {
		return
	}

	if len(conf.dsn) == 0 {
		log.Warn("no DSN provided (otel-go is disabled)")
		return
	}

	dsn, err := ParseDSN(conf.dsn[0])
	if err != nil {
		log.Warnf("invalid DSN: %s (otel is disabled)", err)
		return
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

	atomicClient.Store(client)
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

// ------------------------------------------------------------------------------

var (
	fallbackClient = newClient(&DSN{})
	atomicClient   atomic.Value
)

func activeClient() *client {
	v := atomicClient.Load()
	if v == nil {
		return fallbackClient
	}
	return v.(*client)
}

func TraceURL(span trace.Span) string {
	return activeClient().TraceURL(span)
}

func ReportError(ctx context.Context, err error, opts ...trace.EventOption) {
	activeClient().ReportError(ctx, err, opts...)
}

func ReportPanic(ctx context.Context, val any) {
	activeClient().ReportPanic(ctx, val)
}

func Shutdown(ctx context.Context) error {
	return activeClient().Shutdown(ctx)
}

func ForceFlush(ctx context.Context) error {
	return activeClient().ForceFlush(ctx)
}

func TracerProvider() *sdktrace.TracerProvider {
	return activeClient().tp
}

// SetLogger sets the logger to the given one.
func SetLogger(logger logger.Logger) {
	log = logger
}
