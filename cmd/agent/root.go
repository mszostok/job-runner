package main

import (
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/cmd/agent/start"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
)

const Name = "agent"

// NewRoot returns a root cobra.Command for the whole Agent utility.
func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   Name,
		Short: "Linux Process Runner Agent",
		Long: heredoc.WithCLIName(`
        <cli> - Linux Process Runner Agent

        A utility that runs on Linux hosts and manages executed Linux Processes (Jobs).

        Quick Start:

            $ <cli> start daemon             # Starts Agent long living process on host.
            `, Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		start.NewCmd(),
	)

	return rootCmd
}
