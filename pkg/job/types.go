package job

import (
	"fmt"
	"time"
)

// Status specifies human-readable Cmd status.
type Status string

const (
	Running    Status = "RUNNING"
	Failed     Status = "FAILED"
	Succeeded  Status = "SUCCEEDED"
	Terminated Status = "TERMINATED"
)

func (s Status) IsFinished() bool {
	return s != Running
}

type RunInput struct {
	// Tenant specifies the tenant of a given Job.
	Tenant string
	// Name specifies Cmd name.
	Name string
	// Command is the path of the command to run.
	Command string
	// Args holds command line arguments.
	Args []string
	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	Env []string
	// TODO(simplification): Resources specifies Cmd's system resources limits.
	// In the first version not supported. Use globals defined on Agent side.
	//Resources Resources
}

type RunOutput struct{}

type GetInput struct {
	// Name specifies Cmd name.
	Name string
}

type GetOutput struct {
	// CreatedBy specifies the tenant that executed a given Cmd.
	CreatedBy string
	// Status of a given Cmd.
	Status Status
	// ExitCode of the exited process. While Status in Running, exit code should be ignored.
	ExitCode int
}

func (g GetOutput) String() string {
	switch g.Status {
	case Running:
		return fmt.Sprintf("Job created by %q is still running", g.CreatedBy)
	default:
		return fmt.Sprintf("Job created by %q is in %q state with exit code %d", g.CreatedBy, g.Status, g.ExitCode)
	}
}

type StreamLogsInput struct {
	// Name specifies Cmd name.
	Name string
}

type StreamLogsOutput struct {
	// Output represents the streamed Cmd logs. It is from start of Cmd execution.
	Output <-chan []byte
	// Error allows communicating issues encountered during logs streaming.
	Error <-chan error
}

type StopInput struct {
	// Name specifies Cmd name.
	Name string
	// GracePeriod represents a period of time given to the Cmd to terminate gracefully.
	GracePeriod time.Duration
}

type StopOutput struct {
	// Status of a given Cmd.
	Status Status
	// ExitCode of the exited process.
	ExitCode int
}
