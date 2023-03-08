package errors

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// BadRequestDetails returns an error details for an invalid argument.
// fd is a list of field / description pairs.
func BadRequestDetails(fd ...string) *errdetails.BadRequest {
	var fieldViolations []*errdetails.BadRequest_FieldViolation
	for i := 0; i < len(fd); i += 2 {
		fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       fd[i],
			Description: fd[i+1],
		})
	}
	return &errdetails.BadRequest{
		FieldViolations: fieldViolations,
	}
}
