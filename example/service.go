package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	logging2 "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"go.linka.cloud/grpc-tookit/example/pb"
	"go.linka.cloud/grpc-toolkit/interceptors/auth"
	"go.linka.cloud/grpc-toolkit/interceptors/ban"
	"go.linka.cloud/grpc-toolkit/interceptors/defaulter"
	"go.linka.cloud/grpc-toolkit/interceptors/iface"
	"go.linka.cloud/grpc-toolkit/interceptors/logging"
	metrics2 "go.linka.cloud/grpc-toolkit/interceptors/metrics"
	"go.linka.cloud/grpc-toolkit/interceptors/tracing"
	validation2 "go.linka.cloud/grpc-toolkit/interceptors/validation"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/service"
)

func newService(ctx context.Context, opts ...service.Option) (service.Service, error) {
	log := logger.C(ctx)
	metrics := metrics2.NewInterceptors(metrics2.WithExemplarFromContext(metrics2.DefaultExemplarFromCtx))

	address := "0.0.0.0:9991"

	var svc service.Service
	opts = append(opts,
		service.WithContext(ctx),
		service.WithAddress(address),
		// service.WithRegistry(mdns.NewRegistry()),
		service.WithReflection(true),
		service.WithoutCmux(),
		service.WithGateway(pb.RegisterGreeterHandler),
		service.WithGatewayPrefix("/rest"),
		service.WithGRPCWeb(true),
		service.WithGRPCWebPrefix("/grpc"),
		service.WithMiddlewares(otelhttp.NewMiddleware("hello"), httpLogger),
		service.WithInterceptors(
			tracing.NewInterceptors(),
			metrics,
			logging.New(ctx, logging2.WithFieldsFromContext(func(ctx context.Context) logging2.Fields {
				if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
					return logging2.Fields{"traceid", span.TraceID().String(), "spanid", span.SpanID().String()}
				}
				return nil
			})),
		),
		service.WithServerInterceptors(
			ban.NewInterceptors(ban.WithDefaultJailDuration(time.Second), ban.WithDefaultCallback(func(action ban.Action, actor string, rule *ban.Rule) error {
				log.WithFields("action", action, "actor", actor, "rule", rule.Name).Info("ban callback")
				return nil
			})),
			auth.NewServerInterceptors(auth.WithBasicValidators(func(ctx context.Context, user, password string) (context.Context, error) {
				if !auth.Equals(user, "admin") || !auth.Equals(password, "admin") {
					return ctx, fmt.Errorf("invalid user or password")
				}
				log.Infof("request authenticated")
				return ctx, nil
			})),
		),
		service.WithInterceptors(
			defaulter.NewInterceptors(),
			validation2.NewInterceptors(true),
		),
		// enable server interface interceptor
		service.WithServerInterceptors(iface.New()),
	)
	svc, err := service.New(opts...)
	if err != nil {
		return nil, err
	}
	pb.RegisterGreeterServer(svc, &GreeterHandler{})
	metrics.Register(svc)
	return svc, nil
}

func httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		log := logger.From(request.Context()).WithFields(
			"method", request.Method,
			"host", request.Host,
			"path", request.URL.Path,
			"remoteAddress", request.RemoteAddr,
		)
		next.ServeHTTP(writer, request)
		log.WithField("duration", time.Since(start)).Info()
	})
}
