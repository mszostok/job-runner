package auth

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli/config"
)

// NewCmd returns a new cobra.Command subcommand for auth related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "auth",
		Short: "This command consists of commands related to authN/authZ",
	}

	root.AddCommand(
		NewLogin(),
		NewLogout(),
		NewUse(),
		// TODO: add list cmd
	)
	return root
}

func resolveInputAlias(args []string, promptMsg string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	return askForAlias(promptMsg)
}

func askForAlias(msg string) (string, error) {
	candidates, err := config.GetAgentsAlias()
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("Not logged in to any server")
	}

	var serverAddress string
	err = survey.AskOne(&survey.Select{
		Message: msg,
		Options: candidates,
	}, &serverAddress)
	if err != nil {
		return "", err
	}

	return serverAddress, nil
}
