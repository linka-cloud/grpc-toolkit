package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/bridges/prometheus"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func configureMetrics(ctx context.Context, conf *config) *sdkmetric.MeterProvider {
	opts := conf.metricOptions
	if res := conf.newResource(); res != nil {
		opts = append(opts, sdkmetric.WithResource(res))
	}

	for _, dsn := range conf.dsn {
		dsn, err := ParseDSN(dsn)
		if err != nil {
			log.WithError(err).Error("ParseDSN failed")
			continue
		}

		exp, err := otlpmetricClient(ctx, conf, dsn)
		if err != nil {
			log.WithError(err).Warn("otlpmetricClient")
			continue
		}

		ropts := []sdkmetric.PeriodicReaderOption{sdkmetric.WithInterval(5 * time.Second)}
		if conf.metricPrometheusBridge {
			ropts = append(ropts, sdkmetric.WithProducer(prometheus.NewMetricProducer()))
		}
		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, ropts...)))
	}

	provider := sdkmetric.NewMeterProvider(opts...)
	otel.SetMeterProvider(provider)

	if err := runtimemetrics.Start(); err != nil {
		log.WithError(err).Error("runtimemetrics.Start failed")
	}

	return provider
}

func otlpmetricClient(ctx context.Context, conf *config, dsn *DSN) (sdkmetric.Exporter, error) {
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(dsn.OTLPHttpEndpoint()),
		otlpmetrichttp.WithHeaders(dsn.Headers()),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithTemporalitySelector(preferDeltaTemporalitySelector),
	}

	if conf.tlsConf != nil {
		options = append(options, otlpmetrichttp.WithTLSClientConfig(conf.tlsConf))
	} else if dsn.Scheme == "http" {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	return otlpmetrichttp.New(ctx, options...)
}

func preferDeltaTemporalitySelector(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	switch kind {
	case sdkmetric.InstrumentKindCounter,
		sdkmetric.InstrumentKindObservableCounter,
		sdkmetric.InstrumentKindHistogram:
		return metricdata.DeltaTemporality
	default:
		return metricdata.CumulativeTemporality
	}
}
