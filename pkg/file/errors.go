package file

import "fmt"

type ConflictError struct {
	jobName string
}

func NewConflictError(jobName string) *ConflictError {
	return &ConflictError{jobName: jobName}
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("Job name %q is not unique and dedicated log file cannot be created", e.jobName)
}
