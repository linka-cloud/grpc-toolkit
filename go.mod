module go.linka.cloud/grpc

go 1.13

require (
	github.com/alta/protopatch v0.3.4
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/caitlinelfring/go-env-default v1.1.0
	github.com/envoyproxy/protoc-gen-validate v0.9.1
	github.com/fsnotify/fsnotify v1.5.1
	github.com/fullstorydev/grpchan v1.1.1
	github.com/go-logr/logr v1.2.3
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/improbable-eng/grpc-web v0.14.1
	github.com/jaredfolkins/badactor v1.2.0
	github.com/johnbellone/grpc-middleware-sentry v0.2.0
	github.com/justinas/alice v1.2.0
	github.com/miekg/dns v1.1.41
	github.com/opentracing/opentracing-go v1.1.0
	github.com/planetscale/vtprotobuf v0.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.9.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	go.linka.cloud/protoc-gen-defaults v0.1.0
	go.linka.cloud/protoc-gen-go-fields v0.1.1
	go.linka.cloud/protofilters v0.2.2
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.5.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.53.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.28.1
)

replace (
	github.com/grpc-ecosystem/go-grpc-prometheus => github.com/linka-cloud/go-grpc-prometheus v1.2.0-lk
	github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/linka-cloud/grpc-gateway/v2 v2.5.1-0.20210917084803-33b6d54c9e11
)
