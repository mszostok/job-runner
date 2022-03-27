package job

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
	"github.com/mszostok/job-runner/internal/cli/printer"
	"github.com/mszostok/job-runner/pkg/api/grpc"
)

type RunOptions struct {
	Env []string
}

// NewRun returns a new cobra.Command for running Job.
func NewRun() *cobra.Command {
	var opts RunOptions

	cmd := &cobra.Command{
		Use:   `run NAME [--env="key=value"] -- [COMMAND] [args...]`,
		Short: "Runs a given Job",
		Args:  cobra.MinimumNArgs(2),
		Example: heredoc.WithCLIName(`
			# Start the "episode-42" Job which prints "test"
			<cli> job run episode-42 --  sh -c 'echo test'

			# Start the "episode-42" Job using command and custom arguments
			<cli> job run episode-42 -- <cmd> <arg1> ... <argN>
		`, cli.Name),
		RunE: func(c *cobra.Command, args []string) error {
			runCmd, runArgs, err := cli.ExtractExecCommandAfterDash(c, args)
			if err != nil {
				return err
			}

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

			status.Step("Scheduling Job")
			_, err = client.Run(c.Context(), &grpc.RunRequest{
				Name:    args[0],
				Command: runCmd,
				Args:    runArgs,
				Env:     opts.Env,
			})
			status.End(err == nil)
			// TODO(simplification): to improve UX, gRPC errors can be translated to more user friendly messages
			return err
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Env, "env", "e", []string{}, `Specifies the environment of the process. Each entry is of the form "key=value".`)

	return cmd
}
