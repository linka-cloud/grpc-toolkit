package validation

import (
	"context"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"go.linka.cloud/grpc/errors"
	"go.linka.cloud/grpc/interceptors"
)

type validatorAll interface {
	Validate() error
	ValidateAll() error
}

// The validate interface starting with protoc-gen-validate v0.6.0.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/455.
type validator interface {
	Validate(all bool) error
}

// The validate interface prior to protoc-gen-validate v0.6.0.
type validatorLegacy interface {
	Validate() error
}

type validatorMultiError interface {
	AllErrors() []error
}

type validatorError interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
}

func validatorErrorToGrpc(e validatorError) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       e.Field(),
		Description: e.Reason(),
	}
}

func errToStatus(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case validatorError:
		return errors.InvalidArgumentD(err, validatorErrorToGrpc(v))
	case validatorMultiError:
		var details []proto.Message
		for _, v := range v.AllErrors() {
			if d, ok := v.(validatorError); ok {
				details = append(details, validatorErrorToGrpc(d))
			}
		}
		return errors.InvalidArgumentd(err, details...)
	default:
		return errors.InvalidArgument(err)
	}
}

func (i interceptor) validate(req interface{}) error {
	switch v := req.(type) {
	case validatorAll:
		if i.all {
			return errToStatus(v.ValidateAll())
		}
		return errToStatus(v.Validate())
	case validatorLegacy:
		return errToStatus(v.Validate())
	case validator:
		return errToStatus(v.Validate(i.all))
	}
	return nil
}

type interceptor struct {
	all bool
}

func NewInterceptors(validateAll bool) interceptors.Interceptors {
	return &interceptor{all: validateAll}
}

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming messages.
//
// Invalid messages will be rejected with `InvalidArgument` before reaching any userspace handlers.
func (i interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := i.validate(req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that validates outgoing messages.
//
// Invalid messages will be rejected with `InvalidArgument` before sending the request to server.
func (i interceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := i.validate(req); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates incoming messages.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies based on the
// type of the RPC. For `ServerStream` (1:m) requests, it will happen before reaching any userspace
// handlers. For `ClientStream` (n:1) or `BidiStream` (n:m) RPCs, the messages will be rejected on
// calls to `stream.Recv()`.
func (i interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{ServerStream: stream, i: i}
		return handler(srv, wrapper)
	}
}

func (i interceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		desc.Handler = (&sendWrapper{handler: desc.Handler, i: i}).Handler()
		return streamer(ctx, desc, cc, method)
	}
}

type recvWrapper struct {
	i interceptor
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if err := s.i.validate(m); err != nil {
		return err
	}
	return nil
}

type sendWrapper struct {
	i interceptor
	grpc.ServerStream
	handler grpc.StreamHandler
}

func (s *sendWrapper) Handler() grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		return s.handler(srv, s)
	}
}

func (s *sendWrapper) SendMsg(m interface{}) error {
	if err := s.i.validate(m); err != nil {
		return err
	}
	return s.ServerStream.SendMsg(m)
}
