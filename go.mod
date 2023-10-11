module go.linka.cloud/grpc-toolkit

go 1.20

require (
	github.com/alta/protopatch v0.5.3
	github.com/bombsimon/logrusr/v4 v4.0.0
	github.com/caitlinelfring/go-env-default v1.1.0
	github.com/envoyproxy/protoc-gen-validate v1.0.2
	github.com/fsnotify/fsnotify v1.5.1
	github.com/fullstorydev/grpchan v1.1.1
	github.com/go-logr/logr v1.2.3
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/improbable-eng/grpc-web v0.14.1
	github.com/jaredfolkins/badactor v1.2.0
	github.com/johnbellone/grpc-middleware-sentry v0.2.0
	github.com/justinas/alice v1.2.0
	github.com/miekg/dns v1.1.41
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pires/go-proxyproto v0.7.0
	github.com/planetscale/vtprotobuf v0.5.0
	github.com/prometheus/client_golang v1.15.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.9.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	go.linka.cloud/protoc-gen-defaults v0.4.0
	go.linka.cloud/protoc-gen-go-fields v0.4.0
	go.linka.cloud/protofilters v0.8.1
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.17.0
	golang.org/x/sync v0.3.0
	google.golang.org/genproto/googleapis/api v0.0.0-20230913181813-007df8e322eb
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230913181813-007df8e322eb
	google.golang.org/grpc v1.57.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/getsentry/sentry-go v0.11.0 // indirect
	github.com/gin-gonic/gin v1.7.7 // indirect
	github.com/gobwas/httphead v0.0.0-20200921212729-da3d93bc3c58 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.0.4 // indirect
	github.com/golang/glog v1.1.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jhump/protoreflect v1.11.0 // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lyft/protoc-gen-star v0.6.2 // indirect
	github.com/lyft/protoc-gen-star/v2 v2.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/genproto v0.0.0-20230913181813-007df8e322eb // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)

replace (
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	github.com/grpc-ecosystem/go-grpc-prometheus => github.com/linka-cloud/go-grpc-prometheus v1.2.0-lk
	github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/linka-cloud/grpc-gateway/v2 v2.5.1-0.20230307172009-9d6e1ebe3907
)
