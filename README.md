# gRPC 

A utility module, largely taken from the [go-micro](https://github.com/micro/go-micro) patterns (and a good amount of code too...) 
with pure gRPC ecosystem modules.

Principles:
- Pluggable
- No singleton

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
    - [ ] tracing
    - [ ] metrics
- [ ] broker, based on nats-streaming

### Used modules:
- https://github.com/grpc-ecosystem/go-grpc-middleware
- https://github.com/grpc-ecosystem/grpc-opentracing
- https://github.com/grpc-ecosystem/go-grpc-prometheus
- https://github.com/grpc-ecosystem/grpc-gateway
