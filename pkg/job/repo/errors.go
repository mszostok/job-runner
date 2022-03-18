package repo

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned if Job was not found.
	ErrNotFound = errors.New("not found")
)

type IDCConflictError struct {
	id string
}

func NewIDCConflictError(id string) *IDCConflictError {
	return &IDCConflictError{id: id}
}

func (e IDCConflictError) Error() string {
	return fmt.Sprintf("id %q is already present in database", e.id)
}
