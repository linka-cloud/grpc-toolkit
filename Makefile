

.PHONY: clean-example-proto
clean-example-proto:
	@rm example/*.pb.go

.PHONY: example-proto
example-proto:
	@protoc -I. -Iexample --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. example/example.proto
