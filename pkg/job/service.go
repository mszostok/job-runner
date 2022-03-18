package job

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/mszostok/job-runner/internal/shutdown"
	"github.com/mszostok/job-runner/pkg/cgroup"
	"github.com/mszostok/job-runner/pkg/file"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

var _ shutdown.ShutdownableService = &Service{}

const cgroupDefaultParentName = "LPR"

type Storage interface {
	Insert(in repo.InsertInput) error
	Get(in repo.GetInput) (repo.GetOutput, error)
	Update(in repo.UpdateInput) error
}

type FileLogger interface {
	ReadAndFollow() (io.ReadCloser, error)
	NewSink(name string) (io.Writer, error)
}

// Service provides functionality to run/stop/watch arbitrary Linux processes.
type Service struct {
	jobStorage Storage
	fileLogger *file.Logger

	stopMux       sync.Mutex
	createProcCmd func(in RunInput, sink io.Writer) (*exec.Cmd, error)
}

func NewService(jobStorage Storage, logger *file.Logger, opts ...ServiceOption) (*Service, error) {
	svc := &Service{
		jobStorage:    jobStorage,
		fileLogger:    logger,
		createProcCmd: wrapProcForChildExecution,
	}

	for _, option := range opts {
		option(svc)
	}

	return svc, nil
}

func (l *Service) Run(_ context.Context, in RunInput) (*RunOutput, error) {
	sink, releaseSink, err := l.fileLogger.NewSink(in.Name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create log sink")
	}

	cmd, err := l.createProcCmd(in, sink)
	if err != nil {
		return nil, errors.Wrap(err, "while wrapping for child proc execution")
	}

	job := repo.JobDefinition{
		Name:        in.Name,
		Tenant:      in.Tenant,
		Cmd:         cmd,
		RunFinished: make(chan struct{}),
		Status:      string(Running),
	}
	// TODO(simplification): we should think when we want to delete it. We could:
	// - add `defer` to remove stored Job in case of any error in the below steps,
	// - or preserve it so it could be later fetch via `Get`,
	//   - and may add some sanity "cron job" that will care of cleanup such "orphans"
	if err := l.jobStorage.Insert(repo.InsertInput{Job: job}); err != nil {
		return nil, errors.Wrap(err, "while storing Job")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "while starting Job")
	}

	go l.watchRunningProcess(job, func() error {
		// NOTE: We cannot use `cmd.Wait` multiple times, so we need to use dedicated channel
		// to inform others about finished cmd.
		close(job.RunFinished)
		return releaseSink()
	})

	return &RunOutput{}, nil
}

func (l *Service) Get(_ context.Context, in GetInput) (*GetOutput, error) {
	out, err := l.jobStorage.Get(repo.GetInput(in))
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Job from storage")
	}

	return &GetOutput{
		CreatedBy: out.Job.Tenant,
		Status:    Status(out.Job.Status),
		ExitCode:  out.Job.ExitCode,
	}, nil
}

func (l *Service) StreamLogs(ctx context.Context, in StreamLogsInput) (*StreamLogsOutput, error) {
	out, err := l.jobStorage.Get(repo.GetInput(in))
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Job from storage")
	}

	outChan, errChan, err := l.fileLogger.ReadAndFollow(ctx, out.Job.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while reading Job's logs")
	}
	return &StreamLogsOutput{
		Output: outChan,
		Error:  errChan,
	}, nil
}

// Stop stops a given Job.
// TODO(simplification): handle input context cancellation.
func (l *Service) Stop(_ context.Context, in StopInput) (*StopOutput, error) {
	// TODO(simplification):
	// Currently `Stop` is locked for all Jobs. It can be changed to named mutex
	// or we can get rid of mutex usage and think about sync.Once{}
	l.stopMux.Lock()
	defer l.stopMux.Unlock()

	out, err := l.jobStorage.Get(repo.GetInput{Name: in.Name})
	if err != nil {
		return nil, errors.Wrap(err, "while getting out definition")
	}
	status := Status(out.Job.Status)

	if status.IsFinished() {
		return &StopOutput{
			Status:   status,
			ExitCode: out.Job.ExitCode,
		}, nil
	}

	_ = out.Job.Cmd.Process.Signal(syscall.SIGTERM)
	if in.GracePeriod != 0 {
		scheduleHardKill := time.AfterFunc(in.GracePeriod, func() {
			_ = out.Job.Cmd.Process.Kill() // err is handled by statusForCmd
		})
		defer scheduleHardKill.Stop() // cancel hard kill if out.Job.Cmd.Wait() finished before grace period
	}

	<-out.Job.RunFinished

	status, exitCode := l.statusForCmd(out.Job.Cmd)
	return &StopOutput{
		Status:   status,
		ExitCode: exitCode,
	}, nil
}

func (l *Service) Shutdown() error {
	// TODO: Here we should list all running Jobs and trigger `Stop` for them.
	return nil
}

func (l *Service) watchRunningProcess(job repo.JobDefinition, release func() error) {
	defer func() { // release file used for logs (stdout, stderr)
		// TODO(simplification): handle error gracefully
		_ = release()
	}()

	_ = job.Cmd.Wait()
	status, exitCode := l.statusForCmd(job.Cmd)

	// TODO(simplification): handle error:
	//  - log it (zap/logrus)
	//  - execute retry. If after X retries we still get an error, push it to a dead letter queue.
	_ = l.jobStorage.Update(repo.UpdateInput{
		Name:     job.Name,
		Status:   string(status),
		ExitCode: exitCode,
	})
}

// statusForCmd can be called only if `Wait` was already executed for a given cmd.
func (l *Service) statusForCmd(cmd *exec.Cmd) (Status, int) {
	if cmd.ProcessState.Success() {
		return Succeeded, 0
	}

	sysStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if sysStatus.Signaled() {
		return Terminated, cmd.ProcessState.ExitCode()
	}

	return Failed, cmd.ProcessState.ExitCode()
}

func wrapProcForChildExecution(in RunInput, sink io.Writer) (*exec.Cmd, error) {
	cgroupPath := getJobCgroupPath(in.Name)

	selfBin, err := os.Executable()
	if err != nil {
		return nil, err
	}

	childArgs := []string{"start", "child"}
	for _, env := range in.Env {
		childArgs = append(childArgs, "--env", env)
	}
	childArgs = append(childArgs, "--cgroup-procs-path", cgroupPath)
	childArgs = append(childArgs, "--", in.Command)
	childArgs = append(childArgs, in.Args...)

	cmd := exec.Command(selfBin, childArgs...)
	cmd.Stderr = sink
	cmd.Stdout = sink

	err = cgroup.BootstrapChild(cgroupPath, DefaultProcResources)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func getJobCgroupPath(name string) string {
	return filepath.Join(cgroup.PseudoFsPrefix, cgroupDefaultParentName, name)
}

func directProcExecution(in RunInput, sink io.Writer) (*exec.Cmd, error) {
	// This needs to be allowed, but we need to be aware of potential risk:
	//   https://github.com/securego/gosec/issues/204#issuecomment-384474356
	// #nosec G204
	cmd := exec.Command(in.Command, in.Args...)
	cmd.Env = in.Env
	cmd.Stderr = sink
	cmd.Stdout = sink

	return cmd, nil
}
