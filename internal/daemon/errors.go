package daemon

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mszostok/job-runner/pkg/job"
)

var (
	// NilRequestInputError indicates that the received request was nil.
	NilRequestInputError = status.Error(codes.InvalidArgument, "request cannot be nil")
)

func TranslateError(err error) error {
	if err == nil {
		return nil
	}

	_, ok := status.FromError(err)
	if ok {
		return err // it's already gRPC error format
	}

	switch {
	case job.IsConflictError(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case job.IsNotFoundError(err):
		return status.Error(codes.NotFound, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
