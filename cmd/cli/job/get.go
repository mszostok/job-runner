package job

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
	"github.com/mszostok/job-runner/internal/cli/printer"
	"github.com/mszostok/job-runner/pkg/api/grpc"
)

// NewGet returns a new cobra.Command for fetching a given Job.
// TODO(UX): If NAME not provided, list all Jobs, same as `kubectl get foo` does.
func NewGet() *cobra.Command {
	jobPrinter := printer.NewForJob(os.Stdout)

	cmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Returns a given Job definition",
		Args:  cobra.ExactArgs(1),
		Example: heredoc.WithCLIName(`
			# Show the Job "episode-42" in table format
			<cli> job get episode-42

			# Show the Job "episode-42" in YAML format
			<cli> job get episode-42 -oyaml

			# Show the Job "episode-42" in JSON format
			<cli> job get episode-42 -ojson
		`, cli.Name),
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

			input := grpc.GetRequest{
				Name: args[0],
			}
			out, err := client.Get(c.Context(), &input)
			if err != nil { // TODO(simplification): to improve UX, gRPC errors can be translated to a user friendly messages
				return err
			}

			return jobPrinter.Print(printer.JobDefinition{
				Name:      input.Name,
				CreatedBy: out.CreatedBy,
				Status:    out.Status.String(),
				ExitCode:  int(out.ExitCode),
			})
		},
	}

	jobPrinter.RegisterFlags(cmd.Flags())

	return cmd
}
