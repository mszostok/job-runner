package start

import (
	"log"

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

// NewDaemon returns a new cobra.Command for starting main process.
func NewDaemon() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Starts a long living Agent process.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			jobName := args[0]

			err := cgroup.BootstrapParent(daemonCGroupPath, cgroup.MemoryController, cgroup.CPUController, cgroup.IOController, cgroup.CPUSetController)
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

			// TODO(server): Implement gRPC server that is long-living and can handle `command` execution.
			// For now just showcase that run + logs work.

			log.Printf("Running %q Job\n", jobName)
			_, err = svc.Run(c.Context(), job.RunInput{
				Tenant:  tenant,
				Name:    jobName,
				Command: "sh",
				Args:    []string{"-c", "sleep 60 && echo $MOTTO"},
				Env:     []string{"MOTTO=hakuna_matata"},
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

			log.Println("Waiting for ctrl+c to execute Agent graceful shutdown")
			<-c.Context().Done()

			if err := shutdownManager.Shutdown(); err != nil {
				log.Printf("Graceful shutdown unsuccessful: %v", err)
				return err
			}

			log.Println("Successful graceful shutdown")
			return nil
		},
	}

	return cmd
}
