module go.linka.cloud/grpc-toolkit

go 1.23.0

toolchain go1.24.3

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/alta/protopatch v0.5.3
	github.com/bombsimon/logrusr/v4 v4.0.0
	github.com/caitlinelfring/go-env-default v1.1.0
	github.com/envoyproxy/protoc-gen-validate v1.2.1
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.5.4
	github.com/fullstorydev/grpchan v1.1.1
	github.com/go-logr/logr v1.4.2
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	github.com/jaredfolkins/badactor v1.2.0
	github.com/johnbellone/grpc-middleware-sentry v0.3.0
	github.com/justinas/alice v1.2.0
	github.com/miekg/dns v1.1.41
	github.com/pires/go-proxyproto v0.7.0
	github.com/planetscale/vtprotobuf v0.6.1-0.20240917153116-6f2963f01587
	github.com/prometheus/client_golang v1.22.0
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.9.3
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.10.0
	github.com/tailscale/peercred v0.0.0-20250107143737-35a0c7bd7edc
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	github.com/traefik/grpc-web v0.16.0
	github.com/uptrace/opentelemetry-go-extra/otellogrus v0.3.2
	go.linka.cloud/protoc-gen-defaults v0.4.0
	go.linka.cloud/protoc-gen-go-fields v0.4.0
	go.linka.cloud/protofilters v0.8.1
	go.opentelemetry.io/contrib/bridges/otellogrus v0.11.0
	go.opentelemetry.io/contrib/bridges/prometheus v0.61.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.61.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.12.2
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.36.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.36.0
	go.opentelemetry.io/otel/log v0.12.2
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/log v0.12.2
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.40.0
	golang.org/x/sync v0.14.0
	golang.org/x/sys v0.33.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237
	google.golang.org/grpc v1.72.1
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.5.1
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/getsentry/sentry-go v0.24.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jhump/protoreflect v1.11.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lyft/protoc-gen-star v0.6.2 // indirect
	github.com/lyft/protoc-gen-star/v2 v2.0.4-0.20230330145011-496ad1ac90a4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.3.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

replace (
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	github.com/grpc-ecosystem/go-grpc-prometheus => github.com/linka-cloud/go-grpc-prometheus v1.2.0-lk
	github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/linka-cloud/grpc-gateway/v2 v2.20.0-lk
	nhooyr.io/websocket => github.com/coder/websocket v1.8.6
)
