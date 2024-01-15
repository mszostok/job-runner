package main

import (
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/cmd/cli/auth"
	"github.com/mszostok/job-runner/cmd/cli/job"
	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
)

// NewRoot returns a root cobra.Command for the whole Agent utility.
func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   cli.Name,
		Short: "Linux Process Runner Client",
		Long: heredoc.WithCLIName(`
        <cli> - Linux Process Runner Client

        A utility that simplifies interaction with LPR Agent.
        `, cli.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		job.NewCmd(),
		auth.NewCmd(),
	)

	return rootCmd
}
