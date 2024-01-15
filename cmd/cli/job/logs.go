package job

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/pkg/api/grpc"
)

// NewLogs returns a new cobra.Command for fetching Job's related logs.
func NewLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs NAME",
		Short: "Prints the logs for a Job",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			client, cleanup, err := cli.NewDefaultGRPCAgentClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Printf("while cleaning up connection: %v", err)
				}
			}()

			out, err := client.StreamLogs(c.Context(), &grpc.StreamLogsRequest{
				Name: args[0],
			})
			if err != nil { // TODO(simplification): to improve UX, gRPC errors can be translated to a user friendly messages
				return err
			}

			return grpc.ForwardStreamLogs(c.OutOrStdout(), out)
		},
	}

	return cmd
}
