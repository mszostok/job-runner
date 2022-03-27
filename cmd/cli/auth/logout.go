package auth

import (
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/config"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
	"github.com/mszostok/job-runner/internal/cli/printer"
)

// NewLogout returns a new cobra.Command for logging into Agent.
func NewLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [SERVER]",
		Short: "Logout from the Agent server",
		Example: heredoc.WithCLIName(`
			# Select what server to log out of via a prompt
			<cli> logout

			# Logout of a specified Agent server
			<cli> logout localhost:50051
		`, cli.Name),
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			alias, err := resolveInputAlias(args, "What server do you want to log out of? ")
			if err != nil {
				return err
			}

			status := printer.NewStatus(c.OutOrStdout())
			status.Step("Removing login credentials for %s", alias)
			err = config.DeleteAgentAuthDetails(alias)
			status.End(err == nil)
			return err
		},
	}

	return cmd
}
