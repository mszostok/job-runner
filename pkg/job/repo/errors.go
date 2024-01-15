package repo

import (
	"fmt"
)

// NotFoundError is returned if Job was not found.
type NotFoundError struct {
	id string
}

func NewNotFoundError(id string) *NotFoundError {
	return &NotFoundError{id: id}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Job %q not found", e.id)
}

func (e NotFoundError) NotFound() {}

// IDCConflictError is returned if Job with a given id is already present in database.
type IDCConflictError struct {
	id string
}

func NewIDCConflictError(id string) *IDCConflictError {
	return &IDCConflictError{id: id}
}

func (e IDCConflictError) Error() string {
	return fmt.Sprintf("Job %q is already present in database", e.id)
}

func (e IDCConflictError) Conflict() {}
