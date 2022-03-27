package job

import (
	"github.com/cockroachdb/errors/errbase"
)

// TODO: it will be good to establish error handling between layers (DSO, domain, DTO), so we can easily
// map repository (in-memory, redis, etc.) NotFound error to domain NotFound error (library) and later to any DTO NotFound error (gRPC, REST, etc.)

// IsNotFoundError checks if any underlying error implements NotFound error interface a.k.a behaviour NotFound error.
func IsNotFoundError(err error) bool {
	type notFound interface {
		NotFound()
	}
	return AppliesToAny(err, func(err error) bool {
		_, ok := err.(notFound)
		return ok
	})
}

// IsConflictError checks if any underlying error implements Conflict error interface a.k.a behaviour Conflict error.
func IsConflictError(err error) bool {
	type conflict interface {
		Conflict()
	}
	return AppliesToAny(err, func(err error) bool {
		_, ok := err.(conflict)
		return ok
	})
}

// AppliesToAny checks if given condition applies to any error in the 'cause' chain.
// It supports both errors implementing:
// - causer, via `Cause()` method, from community libraries,
// - and `Wrapper` via `Unwrap()` method, from the Go 2 error proposal.
func AppliesToAny(err error, applies func(error) bool) bool {
	return GetFirstThatApplies(err, applies) != nil
}

// GetFirstThatApplies returns 1st errors that fulfil the requirements.
func GetFirstThatApplies(err error, applies func(error) bool) error {
	for err != nil {
		if applies(err) {
			return err
		}
		err = errbase.UnwrapOnce(err)
	}

	return nil
}
