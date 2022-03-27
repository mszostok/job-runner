package job

import (
	"github.com/spf13/cobra"
)

// NewCmd returns a new cobra.Command subcommand for Job related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "job",
		Short: "This command consists of multiple subcommands managing Job",
	}

	root.AddCommand(
		NewRun(),
		NewGet(),
		NewLogs(),
		NewStop(),
	)
	return root
}
