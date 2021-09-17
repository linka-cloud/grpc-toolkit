module go.linka.cloud/grpc

go 1.13

require (
	github.com/alta/protopatch v0.3.4
	github.com/envoyproxy/protoc-gen-validate v0.6.0
	github.com/fullstorydev/grpchan v1.0.2-0.20201120232431-d0ab778aeebd
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/improbable-eng/grpc-web v0.14.1
	github.com/jinzhu/gorm v1.9.12
	github.com/lyft/protoc-gen-star v0.6.0 // indirect
	github.com/miekg/dns v1.1.35
	github.com/planetscale/vtprotobuf v0.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.7.0
	go.linka.cloud/protoc-gen-defaults v0.1.0
	go.linka.cloud/protoc-gen-go-fields v0.1.1
	go.linka.cloud/protofilters v0.2.2
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	google.golang.org/genproto v0.0.0-20210916144049-3192f974c780
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/linka-cloud/grpc-gateway/v2 v2.5.1-0.20210917084803-33b6d54c9e11
