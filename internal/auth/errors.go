package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO(simplification): such errors are not comparable, e.g. with `errors.Is`, we can make them:
// - Sentinel error - e.g. ErrMissingCert var,
// - Typed error - e.g. MissingCertError struct,
// - Behaviour error - allow interface assertion, e.g. with IsMissingCertError() method.

// NewGRPCMissingCertError returns error indicating that client certificate is missing on gRPC call.
func NewGRPCMissingCertError() error {
	return status.Error(codes.Unauthenticated, "missing client certificate")
}

// NewGRPCInvalidCertError returns error indicating that client certificate was present on gRPC call but it is incorrect.
func NewGRPCInvalidCertError(err error) error {
	return status.Errorf(codes.Unauthenticated, "invalid client certificate: %v", err)
}

// NewGRPCPermissionDeniedError returns error indicating that client certificate was present on gRPC call, it was correct,
// but given user doesn't have enough permission to perform a given action.
func NewGRPCPermissionDeniedError() error {
	return status.Errorf(codes.PermissionDenied, "client certificate doesn't enough permissions to perform this action")
}
