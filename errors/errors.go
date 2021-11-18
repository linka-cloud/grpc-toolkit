package errors

import (
	status2 "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func InvalidArgument(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}
func InvalidArgumentF(msg string, args ...interface{}) error {
	return status.Errorf(codes.InvalidArgument, msg, args...)
}
func InvalidArgumentD(err error, details ...proto.Message) error {
	return statusErr(codes.InvalidArgument, err, details...)
}
func DeadlineExceeded(err error) error {
	return status.Error(codes.DeadlineExceeded, err.Error())
}
func DeadlineExceededF(msg string, args ...interface{}) error {
	return status.Errorf(codes.DeadlineExceeded, msg, args...)
}
func DeadlineExceededD(err error, details ...proto.Message) error {
	return statusErr(codes.DeadlineExceeded, err, details...)
}
func NotFound(err error) error {
	return status.Error(codes.NotFound, err.Error())
}
func NotFoundF(msg string, args ...interface{}) error {
	return status.Errorf(codes.NotFound, msg, args...)
}
func NotFoundD(err error, details ...proto.Message) error {
	return statusErr(codes.NotFound, err, details...)
}
func AlreadyExists(err error) error {
	return status.Error(codes.AlreadyExists, err.Error())
}
func AlreadyExistsF(msg string, args ...interface{}) error {
	return status.Errorf(codes.AlreadyExists, msg, args...)
}
func AlreadyExistsD(err error, details ...proto.Message) error {
	return statusErr(codes.AlreadyExists, err, details...)
}
func PermissionDenied(err error) error {
	return status.Error(codes.PermissionDenied, err.Error())
}
func PermissionDeniedF(msg string, args ...interface{}) error {
	return status.Errorf(codes.PermissionDenied, msg, args...)
}
func PermissionDeniedD(err error, details ...proto.Message) error {
	return statusErr(codes.PermissionDenied, err, details...)
}
func ResourceExhausted(err error) error {
	return status.Error(codes.ResourceExhausted, err.Error())
}
func ResourceExhaustedF(msg string, args ...interface{}) error {
	return status.Errorf(codes.ResourceExhausted, msg, args...)
}
func ResourceExhaustedD(err error, details ...proto.Message) error {
	return statusErr(codes.ResourceExhausted, err, details...)
}
func FailedPrecondition(err error) error {
	return status.Error(codes.FailedPrecondition, err.Error())
}
func FailedPreconditionF(msg string, args ...interface{}) error {
	return status.Errorf(codes.FailedPrecondition, msg, args...)
}
func FailedPreconditionD(err error, details ...proto.Message) error {
	return statusErr(codes.FailedPrecondition, err, details...)
}
func Aborted(err error) error {
	return status.Error(codes.Aborted, err.Error())
}
func AbortedF(msg string, args ...interface{}) error {
	return status.Errorf(codes.Aborted, msg, args...)
}
func AbortedD(err error, details ...proto.Message) error {
	return statusErr(codes.Aborted, err, details...)
}
func OutOfRange(err error) error {
	return status.Error(codes.OutOfRange, err.Error())
}
func OutOfRangeF(msg string, args ...interface{}) error {
	return status.Errorf(codes.OutOfRange, msg, args...)
}
func OutOfRangeD(err error, details ...proto.Message) error {
	return statusErr(codes.OutOfRange, err, details...)
}
func Unimplemented(err error) error {
	return status.Error(codes.Unimplemented, err.Error())
}
func UnimplementedF(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unimplemented, msg, args...)
}
func UnimplementedD(err error, details ...proto.Message) error {
	return statusErr(codes.Unimplemented, err, details...)
}
func Internal(err error) error {
	return status.Error(codes.Internal, err.Error())
}
func InternalF(msg string, args ...interface{}) error {
	return status.Errorf(codes.Internal, msg, args...)
}
func InternalD(err error, details ...proto.Message) error {
	return statusErr(codes.Internal, err, details...)
}
func Unavailable(err error) error {
	return status.Error(codes.Unavailable, err.Error())
}
func UnavailableF(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unavailable, msg, args...)
}
func UnavailableD(err error, details ...proto.Message) error {
	return statusErr(codes.Unavailable, err, details...)
}
func DataLoss(err error) error {
	return status.Error(codes.DataLoss, err.Error())
}
func DataLossF(msg string, args ...interface{}) error {
	return status.Errorf(codes.DataLoss, msg, args...)
}
func DataLossD(err error, details ...proto.Message) error {
	return statusErr(codes.DataLoss, err, details...)
}
func Unauthenticated(err error) error {
	return status.Error(codes.Unauthenticated, err.Error())
}
func UnauthenticatedF(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unauthenticated, msg, args...)
}
func UnauthenticatedD(err error, details ...proto.Message) error {
	return statusErr(codes.Unauthenticated, err, details...)
}
func statusErr(code codes.Code, err error, details ...proto.Message) error {
	return status.FromProto(&status2.Status{Code: int32(code), Message: err.Error(), Details: makeDetails(details...)}).Err()
}
func makeDetails(m ...proto.Message) []*anypb.Any {
	var out []*anypb.Any
	for _, v := range m {
		if a, err := anypb.New(v); err == nil {
			out = append(out, a)
		}
	}
	return out
}

func InvalidArgumentf(msg string, args ...interface{}) error {
	return status.Errorf(codes.InvalidArgument, msg, args...)
}
func InvalidArgumentd(err error, details ...proto.Message) error {
	return statusErr(codes.InvalidArgument, err, details...)
}
func DeadlineExceededf(msg string, args ...interface{}) error {
	return status.Errorf(codes.DeadlineExceeded, msg, args...)
}
func DeadlineExceededd(err error, details ...proto.Message) error {
	return statusErr(codes.DeadlineExceeded, err, details...)
}
func NotFoundf(msg string, args ...interface{}) error {
	return status.Errorf(codes.NotFound, msg, args...)
}
func NotFoundd(err error, details ...proto.Message) error {
	return statusErr(codes.NotFound, err, details...)
}
func AlreadyExistsf(msg string, args ...interface{}) error {
	return status.Errorf(codes.AlreadyExists, msg, args...)
}
func AlreadyExistsd(err error, details ...proto.Message) error {
	return statusErr(codes.AlreadyExists, err, details...)
}
func PermissionDeniedf(msg string, args ...interface{}) error {
	return status.Errorf(codes.PermissionDenied, msg, args...)
}
func PermissionDeniedd(err error, details ...proto.Message) error {
	return statusErr(codes.PermissionDenied, err, details...)
}
func ResourceExhaustedf(msg string, args ...interface{}) error {
	return status.Errorf(codes.ResourceExhausted, msg, args...)
}
func ResourceExhaustedd(err error, details ...proto.Message) error {
	return statusErr(codes.ResourceExhausted, err, details...)
}
func FailedPreconditionf(msg string, args ...interface{}) error {
	return status.Errorf(codes.FailedPrecondition, msg, args...)
}
func FailedPreconditiond(err error, details ...proto.Message) error {
	return statusErr(codes.FailedPrecondition, err, details...)
}
func Abortedf(msg string, args ...interface{}) error {
	return status.Errorf(codes.Aborted, msg, args...)
}
func Abortedd(err error, details ...proto.Message) error {
	return statusErr(codes.Aborted, err, details...)
}
func OutOfRangef(msg string, args ...interface{}) error {
	return status.Errorf(codes.OutOfRange, msg, args...)
}
func OutOfRanged(err error, details ...proto.Message) error {
	return statusErr(codes.OutOfRange, err, details...)
}
func Unimplementedf(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unimplemented, msg, args...)
}
func Unimplementedd(err error, details ...proto.Message) error {
	return statusErr(codes.Unimplemented, err, details...)
}
func Internalf(msg string, args ...interface{}) error {
	return status.Errorf(codes.Internal, msg, args...)
}
func Internald(err error, details ...proto.Message) error {
	return statusErr(codes.Internal, err, details...)
}
func Unavailablef(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unavailable, msg, args...)
}
func Unavailabled(err error, details ...proto.Message) error {
	return statusErr(codes.Unavailable, err, details...)
}
func DataLossf(msg string, args ...interface{}) error {
	return status.Errorf(codes.DataLoss, msg, args...)
}
func DataLossd(err error, details ...proto.Message) error {
	return statusErr(codes.DataLoss, err, details...)
}
func Unauthenticatedf(msg string, args ...interface{}) error {
	return status.Errorf(codes.Unauthenticated, msg, args...)
}
func Unauthenticatedd(err error, details ...proto.Message) error {
	return statusErr(codes.Unauthenticated, err, details...)
}
