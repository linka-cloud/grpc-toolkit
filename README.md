# gRPC 

A utility module, implementing some of the [go-micro](https://github.com/micro/go-micro) patterns with pure gRPC ecosystem modules.

Features:
- [x] simple configuration with options
- [x] embeded gorm database with options (branch db)
- [x] simple TLS configuration
- [ ] TLS auth
- [ ] client connection pool
- [ ] registry / resolver resolution
    - [ ] mdns
    - [ ] kubernetes
- [ ] default interceptors implementation: 
    - [ ] validation
    - [ ] health
    - [ ] context logger
    - [ ] sentry
    - [ ] rate-limiting
    - [ ] auth claim in context
    - [ ] recovery
    - [ ] tracing (open-tracing)
    - [ ] metrics (prometheus)
    - [ ] retries
    - [ ] context DB / transaction
    - ...
- [ ] api gateway with middleware:
    - [ ] auth
    - [ ] cors
    - [ ] logging
    - [ ] tracing
    - [ ] metrics
- [ ] broker, based on nats-streaming

### Used modules:
- https://github.com/grpc-ecosystem/go-grpc-middleware
- https://github.com/grpc-ecosystem/grpc-opentracing
- https://github.com/grpc-ecosystem/go-grpc-prometheus
- https://github.com/grpc-ecosystem/grpc-gateway
