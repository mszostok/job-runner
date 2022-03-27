package file

import "fmt"

// ConflictError represents an error indicating that Job is not unique.
type ConflictError struct {
	jobName string
}

// NewConflictError returns a new ConflictError instance.
func NewConflictError(jobName string) *ConflictError {
	return &ConflictError{jobName: jobName}
}

// Error returns error message.
func (e ConflictError) Error() string {
	return fmt.Sprintf("Job name %q is not unique and dedicated log file cannot be created", e.jobName)
}
