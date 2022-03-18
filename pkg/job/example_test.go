package job_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mszostok/job-runner/pkg/file"
	"github.com/mszostok/job-runner/pkg/job"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

const (
	tenant  = "testing"
	jobName = "example-run"
)

const defaultTempDirPattern = "agent-logs-"

// Example demonstrates how to:
// - execute a Job (Linux process),
// - get information about started Job,
// - and stream Job's logs.
func Example() {
	dir, err := os.MkdirTemp("", defaultTempDirPattern)
	fatalOnErr(err)
	defer func() {
		err := os.RemoveAll(dir)
		fatalOnErr(err)
	}()

	flog, err := file.NewLogger(file.WithLogsDir(dir))
	fatalOnErr(err)

	svc, err := job.NewService(repo.NewInMemory(), flog, job.WithoutCgroup())
	fatalOnErr(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = svc.Run(ctx, job.RunInput{
		Tenant:  tenant,
		Name:    jobName,
		Command: "sh",
		Args:    []string{"-c", "sleep 1 && echo $MOTTO"},
		Env:     []string{"MOTTO=hakuna_matata"},
	})
	fatalOnErr(err)

	getOut, err := svc.Get(ctx, job.GetInput{Name: jobName})
	fatalOnErr(err)

	fmt.Printf("'Get' of just started Job: %s\n\n", getOut)
	time.Sleep(time.Second)

	stream, err := svc.StreamLogs(ctx, job.StreamLogsInput{Name: jobName})
	fatalOnErr(err)

	fmt.Println("Stream logs:")
	err = job.ForwardStreamLogs(ctx, os.Stdout, stream)
	fatalOnErr(err)

	getOut, err = svc.Get(ctx, job.GetInput{Name: jobName})
	fatalOnErr(err)

	fmt.Println()
	fmt.Printf("'Get' of finished Job: %s", getOut)

	// Output:
	// 'Get' of just started Job: Job created by "testing" is still running
	//
	// Stream logs:
	// hakuna_matata
	//
	// 'Get' of finished Job: Job created by "testing" is in "SUCCEEDED" state with exit code 0
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
