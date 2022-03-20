package start

import (
	"fmt"

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

func extractCommandToExecute(c *cobra.Command, args []string) (name string, arg []string, err error) {
	argsLenAtDash := c.ArgsLenAtDash()
	// Check if there are args after dash (--)
	if argsLenAtDash == -1 || len(args) < 1 {
		return "", nil, fmt.Errorf("wrong input format, please specify cmd and args after dash (--)")
	}
	toExec := args[argsLenAtDash:]
	return toExec[0], toExec[1:], nil
}
