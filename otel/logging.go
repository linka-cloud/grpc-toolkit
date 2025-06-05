package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func configureLogging(ctx context.Context, conf *config) *sdklog.LoggerProvider {
	var opts []sdklog.LoggerProviderOption
	if res := conf.newResource(); res != nil {
		opts = append(opts, sdklog.WithResource(res))
	}

	for _, dsn := range conf.dsn {
		dsn, err := ParseDSN(dsn)
		if err != nil {
			log.WithError(err).Error("ParseDSN failed")
			continue
		}

		exp, err := newOtlpLogExporter(ctx, conf, dsn)
		if err != nil {
			log.WithError(err).Error("otlploghttp.New failed")
			continue
		}

		queueSize := queueSize()
		bspOptions := []sdklog.BatchProcessorOption{
			sdklog.WithMaxQueueSize(queueSize),
			sdklog.WithExportMaxBatchSize(queueSize),
			sdklog.WithExportInterval(10 * time.Second),
			sdklog.WithExportTimeout(10 * time.Second),
		}
		bsp := sdklog.NewBatchProcessor(exp, bspOptions...)
		opts = append(opts, sdklog.WithProcessor(bsp))
	}

	provider := sdklog.NewLoggerProvider(opts...)
	global.SetLoggerProvider(provider)

	return provider
}

func newOtlpLogExporter(
	ctx context.Context, conf *config, dsn *DSN,
) (*otlploghttp.Exporter, error) {
	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(dsn.OTLPHttpEndpoint()),
		otlploghttp.WithHeaders(dsn.Headers()),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	}

	if conf.tlsConf != nil {
		options = append(options, otlploghttp.WithTLSClientConfig(conf.tlsConf))
	} else if dsn.Scheme == "http" {
		options = append(options, otlploghttp.WithInsecure())
	}

	return otlploghttp.New(ctx, options...)
}
