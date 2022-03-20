package start

import (
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/shutdown"
	"github.com/mszostok/job-runner/pkg/cgroup"
	"github.com/mszostok/job-runner/pkg/file"
	"github.com/mszostok/job-runner/pkg/job"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

const (
	tenant           = "testing"
	daemonCGroupPath = "LPR"
)

// NewDaemon returns a new cobra.Command for starting daemon process.
func NewDaemon() *cobra.Command {
	var (
		jobName string
		envs    []string
	)
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Starts a long living Agent process.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) (err error) {
			name, arg, err := extractCommandToExecute(c, args)
			if err != nil {
				return err
			}

			err = cgroup.BootstrapParent(daemonCGroupPath, cgroup.MemoryController, cgroup.CPUController, cgroup.IOController, cgroup.CPUSetController)
			if err != nil {
				return err
			}

			flog, err := file.NewLogger()
			if err != nil {
				return err
			}

			svc, err := job.NewService(repo.NewInMemory(), flog)
			if err != nil {
				return err
			}

			shutdownManager := &shutdown.ParentService{}
			shutdownManager.Register(flog)
			shutdownManager.Register(svc)

			defer func() {
				shErr := shutdownManager.Shutdown()
				if shErr != nil {
					log.Printf("Graceful shutdown unsuccessful: %v", shErr)
					err = multierror.Append(err, shErr).ErrorOrNil()
					return
				}
				log.Println("Successful graceful shutdown")
			}()

			// TODO(server): Implement gRPC server that is long-living and can handle `command` execution.
			// For now just showcase that run + logs work.

			log.Printf("Running %q Job\n", jobName)
			_, err = svc.Run(c.Context(), job.RunInput{
				Tenant:  tenant,
				Name:    jobName,
				Command: name,
				Args:    arg,
				Env:     envs,
			})
			if err != nil {
				return err
			}

			log.Println("Starts streaming")
			stream, err := svc.StreamLogs(c.Context(), job.StreamLogsInput{Name: jobName})
			if err != nil {
				return err
			}
			if err := job.ForwardStreamLogs(c.Context(), c.OutOrStdout(), stream); err != nil {
				return err
			}
			log.Println("Streaming finished")

			log.Println("Waiting for TERM signal to execute Agent graceful shutdown")
			<-c.Context().Done()

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&envs, "env", "e", []string{}, `Specifies the environment of the process. Each entry is of the form "key=value".`)
	cmd.Flags().StringVarP(&jobName, "name", "n", "test-v1", `Specifies the environment of the process. Each entry is of the form "key=value".`)
	return cmd
}
