package start

import (
	"github.com/spf13/cobra"
)

// NewCmd returns a new cobra.Command subcommand for Start related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "start",
		Short: "This command consists of multiple subcommands to execute Agent",
	}

	root.AddCommand(
		NewChild(),
		NewDaemon(),
	)
	return root
}
