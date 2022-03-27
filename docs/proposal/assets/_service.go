//go:build asset
// +build asset

package job

import (
	"context"
	"time"
)

// Status specifies human-readable Cmd status.
type Status string

const (
	Running   Status = "RUNNING"
	Failed    Status = "FAILED"
	Succeeded Status = "SUCCEEDED"
)

// Service provides functionality to run/stop/watch arbitrary Linux processes.
type Service struct{}

func NewService(opts ...ServiceOption) *Service {
	o := &ServiceOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return &Service{}
}

type RunInput struct {
	// Name specifies Cmd name.
	Name string
	// Command is the path of the command to run.
	Command string
	// Args holds command line arguments.
	Args []string
	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	Env []string
	// TODO: Resources specifies Cmd's system resources limits.
	// In the first version not supported. Use globals defined on Agent side.
	//Resources Resources
}

type RunOutput struct{}

func (l *Service) Run(ctx context.Context, in RunInput) (RunOutput, error) {
	return RunOutput{}, nil
}

type GetInput struct {
	// Name specifies Cmd name.
	Name string
}

type GetOutput struct {
	// CreatedBy specifies the tenant that executed a given Cmd.
	CreatedBy string
	// Status of a given Cmd.
	Status Status
	// ExitCode of the exited process.
	ExitCode int
}

func (l *Service) Get(ctx context.Context, in GetInput) (GetOutput, error) {
	return GetOutput{}, nil
}

type StreamLogsInput struct {
	// Name specifies Cmd name.
	Name string
}

type StreamLogsOutput struct {
	// Output represents the streamed Cmd logs. It is from start of Cmd execution.
	Output <-chan string
	// Error allows communicating issues encountered during logs streaming.
	Error <-chan error
}

func (l *Service) StreamLogs(ctx context.Context, in StreamLogsInput) (StreamLogsOutput, error) {
	return StreamLogsOutput{}, nil
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

func (l *Service) Stop(ctx context.Context, in StopInput) (StopOutput, error) {
	return StopOutput{}, nil
}
