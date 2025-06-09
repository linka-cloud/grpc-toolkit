package otel

import (
	"context"
	"fmt"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const dummySpanName = "__dummy__"

type Provider interface {
	TraceURL(span trace.Span) string
	ReportError(ctx context.Context, err error, opts ...trace.EventOption)
	ReportPanic(ctx context.Context, val any)
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
	TracerProvider() *sdktrace.TracerProvider
	MeterProvider() *sdkmetric.MeterProvider
	LoggerProvider() *sdklog.LoggerProvider
}

type client struct {
	dsn    *DSN
	tracer trace.Tracer

	tp *sdktrace.TracerProvider
	mp *sdkmetric.MeterProvider
	lp *sdklog.LoggerProvider
}

func newClient(dsn *DSN) *client {
	return &client{
		dsn:    dsn,
		tracer: otel.Tracer("otel-go"),
	}
}

func (c *client) Shutdown(ctx context.Context) (lastErr error) {
	ctx = context.WithoutCancel(ctx)
	if c.tp != nil {
		if err := c.tp.Shutdown(ctx); err != nil {
			lastErr = err
		}
		c.tp = nil
	}
	if c.mp != nil {
		if err := c.mp.Shutdown(ctx); err != nil {
			lastErr = err
		}
		c.mp = nil
	}
	if c.lp != nil {
		if err := c.lp.Shutdown(ctx); err != nil {
			lastErr = err
		}
		c.lp = nil
	}
	return lastErr
}

func (c *client) ForceFlush(ctx context.Context) (lastErr error) {
	if c.tp != nil {
		if err := c.tp.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	if c.mp != nil {
		if err := c.mp.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	if c.lp != nil {
		if err := c.lp.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// TraceURL returns the trace URL for the span.
func (c *client) TraceURL(span trace.Span) string {
	sctx := span.SpanContext()
	return fmt.Sprintf("%s/traces/%s?span_id=%s",
		c.dsn.SiteURL(), sctx.TraceID(), sctx.SpanID().String())
}

// ReportError reports an error as a span event creating a dummy span if necessary.
func (c *client) ReportError(ctx context.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		_, span = c.tracer.Start(ctx, dummySpanName)
		defer span.End()
	}

	span.RecordError(err, opts...)
}

// ReportPanic is used with defer to report panics.
func (c *client) ReportPanic(ctx context.Context, val any) {
	c.reportPanic(ctx, val)
	// Force flush since we are about to exit on panic.
	if c.tp != nil {
		_ = c.tp.ForceFlush(ctx)
	}
}

func (c *client) TracerProvider() *sdktrace.TracerProvider {
	if c.tp == nil {
		return sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithIDGenerator(newIDGenerator()),
		)
	}
	return c.tp
}

func (c *client) MeterProvider() *sdkmetric.MeterProvider {
	if c.mp == nil {
		return sdkmetric.NewMeterProvider()
	}
	return c.mp
}

func (c *client) LoggerProvider() *sdklog.LoggerProvider {
	if c.lp == nil {
		return sdklog.NewLoggerProvider()
	}
	return c.lp
}

func (c *client) reportPanic(ctx context.Context, val interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		_, span = c.tracer.Start(ctx, dummySpanName)
		defer span.End()
	}

	stackTrace := make([]byte, 2048)
	n := runtime.Stack(stackTrace, false)

	span.AddEvent(
		"log",
		trace.WithAttributes(
			attribute.String("log.severity", "panic"),
			attribute.String("log.message", fmt.Sprint(val)),
			attribute.String("exception.stackstrace", string(stackTrace[:n])),
		),
	)
}
