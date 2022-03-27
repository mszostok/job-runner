package auth

import (
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/config"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
)

// NewUse returns a new cobra.Command for changing currently used credentials' config.
func NewUse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use [SERVER]",
		Short: "Change used credentials' config",
		Example: heredoc.WithCLIName(`
			# Selects which Agent server to use of via a prompt
			<cli> use

			# Sets the specified Agent server
			<cli> use localhost:50051
		`, cli.Name),
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			alias, err := resolveInputAlias(args, "Which Agent server do you want to set as the default?")
			if err != nil {
				return err
			}

			return config.SetDefaultContext(alias)
		},
	}

	return cmd
}
