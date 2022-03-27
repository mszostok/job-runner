package job

import (
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/printer"
	"github.com/mszostok/job-runner/pkg/api/grpc"
)

const infiniteGracePeriod = time.Duration(0)

type StopOptions struct {
	Name        string
	GracePeriod time.Duration
}

// NewStop returns a new cobra.Command for stopping Job.
func NewStop() *cobra.Command {
	var (
		opts       StopOptions
		jobPrinter = printer.NewForJob(os.Stdout)
	)

	cmd := &cobra.Command{
		Use:   "stop NAME",
		Short: "Stops a given Job",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.Name = args[0]

			client, cleanup, err := cli.NewDefaultGRPCAgentClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Printf("while cleaning up connection: %v", err)
				}
			}()

			status := printer.NewStatus(c.OutOrStdout())

			status.Step("Stopping %q with %s grace period", opts.Name, gracePeriodString(opts.GracePeriod))
			_, err = client.Stop(c.Context(), &grpc.StopRequest{
				Name:        opts.Name,
				GracePeriod: ptrDuration(opts.GracePeriod),
			})
			status.End(err == nil)
			// TODO(simplification): to improve UX, gRPC errors can be translated to more user friendly messages
			return err
		},
	}

	flags := cmd.Flags()
	flags.DurationVar(&opts.GracePeriod, "grace-period", infiniteGracePeriod, "Represents a period of time given to the Job to terminate gracefully. Zero means infinite.")
	jobPrinter.RegisterFlags(flags)

	return cmd
}

func ptrDuration(in time.Duration) *time.Duration {
	return &in
}

func gracePeriodString(in time.Duration) string {
	if in == infiniteGracePeriod {
		return "infinite"
	}
	return in.String()
}
