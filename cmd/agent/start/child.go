package start

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/pkg/cgroup"
)

// ChildOptions hold additional options for starting child process.
// The command with args should be submitted after dash (--)
type ChildOptions struct {
	Env             []string
	CGroupProcsPath string
}

// NewChild returns a new cobra.Command for starting child process.
func NewChild() *cobra.Command {
	var opts ChildOptions

	cmd := &cobra.Command{
		Use:    `child --cgroup-procs-path=path [--env="key=value"] -- [COMMAND] [args...]`,
		Short:  "Starts a child process of running Agent daemon. This is used internally by Agent",
		Hidden: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cgroup.AttachCurrentProc(opts.CGroupProcsPath); err != nil {
				return err
			}

			argsLenAtDash := c.ArgsLenAtDash()
			// Check if there are args after dash (--)
			if argsLenAtDash == -1 || len(args) < 1 {
				return fmt.Errorf("wrong input format, please specify cmd and args after dash (--)")
			}
			toRun := args[argsLenAtDash:]

			// This needs to be allowed, but we need to be aware of potential risk:
			//   https://github.com/securego/gosec/issues/204#issuecomment-384474356
			// #nosec G204
			cmd := exec.Command(toRun[0], toRun[1:]...)
			cmd.Env = opts.Env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout

			if err := cmd.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Env, "env", "e", []string{}, `Specifies the environment of the process. Each entry is of the form "key=value".`)
	cmd.Flags().StringVar(&opts.CGroupProcsPath, "cgroup-procs-path", "", "Specifies the path to procs file.")
	// error cannot happen as flag is already declared
	_ = cmd.MarkFlagRequired("cgroup-procs-path")

	return cmd
}
