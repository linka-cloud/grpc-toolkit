module go.linka.cloud/grpc

go 1.13

require (
	github.com/alta/protopatch v0.3.4
	github.com/envoyproxy/protoc-gen-validate v0.6.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fullstorydev/grpchan v1.0.2-0.20201120232431-d0ab778aeebd
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/improbable-eng/grpc-web v0.14.1
	github.com/jinzhu/gorm v1.9.12
	github.com/johnbellone/grpc-middleware-sentry v0.2.0
	github.com/justinas/alice v1.2.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lyft/protoc-gen-star v0.6.0 // indirect
	github.com/miekg/dns v1.1.35
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/planetscale/vtprotobuf v0.2.0
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	go.linka.cloud/protoc-gen-defaults v0.1.0
	go.linka.cloud/protoc-gen-go-fields v0.1.1
	go.linka.cloud/protofilters v0.2.2
	go.uber.org/multierr v1.7.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210916144049-3192f974c780
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/linka-cloud/grpc-gateway/v2 v2.5.1-0.20210917084803-33b6d54c9e11
