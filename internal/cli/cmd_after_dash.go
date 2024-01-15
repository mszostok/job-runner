package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func ExtractExecCommandAfterDash(c *cobra.Command, args []string) (name string, arg []string, err error) {
	argsLenAtDash := c.ArgsLenAtDash()
	// Check if there are args after dash (--)
	if argsLenAtDash == -1 || len(args) < 1 {
		return "", nil, errors.New("wrong input format, please specify cmd and args after dash (--)")
	}
	toExec := args[argsLenAtDash:]
	return toExec[0], toExec[1:], nil
}
