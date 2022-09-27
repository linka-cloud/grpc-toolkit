# gRPC 

A utility module, largely taken from the [go-micro](https://github.com/micro/go-micro) patterns (and a good amount of code too...) 
with pure gRPC ecosystem modules.

Principles:
- Pluggable
- No singleton

Features:
- [x] simple configuration with options
- [x] simple TLS configuration
- [ ] TLS auth
- [ ] client connection pool
- [ ] registry / resolver resolution
    - [ ] mdns
    - [ ] kubernetes
- [ ] default interceptors implementation:
    - [ ] context request id
    - [x] defaulter
    - [x] validation
    - [ ] health
    - [ ] context logger
    - [x] sentry
    - [ ] rate-limiting
    - [x] ban
    - [ ] auth claim in context
    - [x] recovery (server side only)
    - [x] tracing (open-tracing)
    - [x] metrics (prometheus)
    - [ ] retries
    - [ ] context DB / transaction
    - ...
- [ ] grpc web / api gateway with middleware:
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
- https://github.com/jaredfolkins/badactor
- https://github.com/johnbellone/grpc-middleware-sentry
- https://github.com/improbable-eng/grpc-web
