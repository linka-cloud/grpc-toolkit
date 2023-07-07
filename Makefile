MODULE = go.linka.cloud/grpc-toolkit


PROTO_BASE_PATH = $(PWD)

INCLUDE_PROTO_PATH = -I$(PROTO_BASE_PATH) \
	-I $(shell go list -m -f {{.Dir}} google.golang.org/protobuf) \
	-I $(shell go list -m -f {{.Dir}} go.linka.cloud/protoc-gen-defaults) \
	-I $(shell go list -m -f {{.Dir}} go.linka.cloud/protofilters) \
	-I $(shell go list -m -f {{.Dir}} github.com/envoyproxy/protoc-gen-validate) \
	-I $(shell go list -m -f {{.Dir}} github.com/alta/protopatch) \
	-I $(shell go list -m -f {{.Dir}} github.com/grpc-ecosystem/grpc-gateway/v2)

PROTO_OPTS = paths=source_relative


$(shell mkdir -p .bin)

export GOBIN=$(PWD)/.bin

export PATH := $(GOBIN):$(PATH)

bin:
	@go install github.com/golang/protobuf/protoc-gen-go
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	@go install go.linka.cloud/protoc-gen-defaults
	@go install go.linka.cloud/protoc-gen-go-fields
	@go install github.com/envoyproxy/protoc-gen-validate
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	@go install github.com/alta/protopatch/cmd/protoc-gen-go-patch
	@go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto

clean: clean-bin clean-proto

clean-bin:
	@rm -rf .bin

clean-proto:
	@find $(PROTO_BASE_PATH) -name '*.pb*.go' -type f -exec rm {} \;

.PHONY: proto
proto: tools.go gen-proto lint


.PHONY: gen-proto
gen-proto: bin
	@find $(PROTO_BASE_PATH) -name '*.proto' -type f -not -path "$(PWD)/google/*" -exec \
    	protoc $(INCLUDE_PROTO_PATH) \
    		--go-patch_out=plugin=go,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=go-grpc,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=defaults,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=go-fields,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=grpc-gateway,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=openapiv2:. \
    		--go-patch_out=plugin=go-vtproto,features=marshal+unmarshal+size,$(PROTO_OPTS):. \
    		--go-patch_out=plugin=validate,lang=go,$(PROTO_OPTS):. {} \;

.PHONY: lint
lint:
	@goimports -w -local $(MODULE) $(PWD)
	@gofmt -w $(PWD)

.PHONY: tests
tests: proto
	@go test -v ./...

