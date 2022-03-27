package auth

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/mszostok/job-runner/internal/cli"
	"github.com/mszostok/job-runner/internal/cli/config"
	"github.com/mszostok/job-runner/internal/cli/heredoc"
	"github.com/mszostok/job-runner/internal/cli/printer"
	"github.com/mszostok/job-runner/pkg/api/grpc"
)

// LoginOptions holds login options which are also compatible with go-survey used for interactive mode.
type LoginOptions struct {
	Alias string `survey:"alias"`

	AgentURL        string `survey:"agent-url"`
	AgentCAFilePath string `survey:"agent-ca"`

	ClientCertFilePath string `survey:"client-cert"`
	ClientKeyFilePath  string `survey:"client-key"`
}

// NewLogin returns a new cobra.Command for logging into Agent.
func NewLogin() *cobra.Command {
	var opts LoginOptions

	cmd := &cobra.Command{
		Use:   "login [SERVER]",
		Short: "Login to a given Agent server",
		Example: heredoc.WithCLIName(`
			# Select what server to log in via a prompt
			<cli> login

			# Specify server name and specify the user
			<cli> login localhost:50051 --agent-ca-cert ./ca_cert.pem --client-cert ./client_cert.pem --client-key ./client_key.pem
		`, cli.Name),
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) (err error) {
			input, err := resolveAgentConfig(opts, args)
			if err != nil {
				return err
			}

			if err := normalize(&input); err != nil {
				return err
			}

			status := printer.NewStatus(c.OutOrStdout())
			defer func() {
				status.End(err == nil)
			}()

			status.Step("Verifying connection...")
			client, cleanup, err := cli.NewGRPCAgentClient(input)
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Printf("while cleaning up connection: %v", err)
				}
			}()

			_, err = client.Ping(c.Context(), &grpc.PingRequest{Message: "Ping"})
			if err != nil {
				return err
			}

			status.Step("Storing configuration with alias %s...", input.Alias)
			return config.SetAgentAuthDetails(input)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&opts.Alias, "alias", "", "Alias for a given Agent configuration. If not provided, default to normalized Agent URL.")
	flags.StringVar(&opts.AgentCAFilePath, "agent-ca-cert", "", "Path on the local disk to CA certificate to verify the Agent server's certificate.")
	flags.StringVar(&opts.ClientCertFilePath, "client-cert", "", "Path on the local disk to client certificate to use for auth to the Agent's server.")
	flags.StringVar(&opts.ClientKeyFilePath, "client-key", "", "Path on the local disk to client private key to use for auth to the Agent's server.")

	return cmd
}

func resolveAgentConfig(answers LoginOptions, args []string) (config.Agent, error) {
	if len(args) > 0 {
		answers.AgentURL = args[0]
	}

	var qs []*survey.Question
	if answers.AgentURL == "" {
		qs = append(qs, &survey.Question{
			Name: "agent-url",
			Prompt: &survey.Input{
				Message: "Agent's server address: ",
			},
			Validate: survey.Required,
		})
	}
	if answers.Alias == "" {
		qs = append(qs, &survey.Question{
			Name: "alias",
			Prompt: &survey.Input{
				Message: "Alias: ",
				Default: strings.ReplaceAll(answers.AgentURL, ".", "-"),
			},
		})
	}

	if answers.AgentCAFilePath == "" {
		qs = append(qs, &survey.Question{
			Name: "agent-ca",
			Prompt: &survey.Input{
				Message: "CA filepath to verity Agent's cert: ",
				Suggest: filePathComplete,
			},
			Validate: survey.Required,
		})
	}

	if answers.ClientCertFilePath == "" {
		qs = append(qs, &survey.Question{
			Name: "client-cert",
			Prompt: &survey.Input{
				Message: "Client certificate filepath: ",
				Suggest: filePathComplete,
			},
			Validate: survey.Required,
		})
	}

	if answers.ClientKeyFilePath == "" {
		qs = append(qs, &survey.Question{
			Name: "client-key",
			Prompt: &survey.Input{
				Message: "Client private key filepath: ",
				Suggest: filePathComplete,
			},
			Validate: survey.Required,
		})
	}

	// perform the questions if needed
	err := survey.Ask(qs, &answers)
	if err != nil {
		return config.Agent{}, errors.Wrap(err, "while asking for server")
	}

	return config.Agent{
		Alias:           answers.Alias,
		ServerURL:       answers.AgentURL,
		AgentCAFilePath: answers.AgentCAFilePath,
		ClientAuth: config.ClientAuth{
			ClientCertAuth: config.ClientCertAuth{
				CertFilePath: answers.ClientCertFilePath,
				KeyFilePath:  answers.ClientKeyFilePath,
			},
		},
	}, nil
}

func filePathComplete(toComplete string) []string {
	files, _ := filepath.Glob(toComplete + "*")
	return files
}

func normalize(input *config.Agent) error {
	if input.Alias == "" {
		input.Alias = strings.ReplaceAll(input.ServerURL, ".", "-")
	}

	var err error
	input.AgentCAFilePath, err = filepath.Abs(input.AgentCAFilePath)
	if err != nil {
		return err
	}
	input.ClientAuth.CertFilePath, err = filepath.Abs(input.ClientAuth.CertFilePath)
	if err != nil {
		return err
	}

	input.ClientAuth.KeyFilePath, err = filepath.Abs(input.ClientAuth.KeyFilePath)
	if err != nil {
		return err
	}

	return nil
}
