package config

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	agentAuthKeyPrefix = "agent"
	context            = "context"
)

// Initialize initializes config store for LPR CLI.
func Initialize() error {
	// Obtain a suitable location for application config files.
	// If the directories don't exist, they will be created relative to the base config directory.
	configFilePath, err := xdg.ConfigFile("lpr/config.yaml")
	if err != nil {
		return errors.Wrap(err, "while getting default config path")
	}

	viper.SetConfigFile(configFilePath)
	err = viper.ReadInConfig()

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return viper.WriteConfig() // write initial config.
		}
		return errors.Wrap(err, "while reading configuration")
	}

	return nil
}

// SetDefaultContext updates context.
func SetDefaultContext(alias string) error {
	viper.Set(context, getAgentStoreKey(alias))
	if err := viper.WriteConfig(); err != nil {
		return errors.Wrap(err, "while writing default context into config file")
	}
	return nil
}

// SetAgentAuthDetails persists auth information for related Agent.
func SetAgentAuthDetails(in Agent) error {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if _, found := cfg.Agent[in.Alias]; found {
		return fmt.Errorf("alias %s is already taken", in.Alias)
	}

	key := getAgentStoreKey(in.Alias)
	viper.Set(key, in)

	if cfg.Context == "" {
		viper.Set(context, key)
	}

	if err := viper.WriteConfig(); err != nil {
		return errors.Wrap(err, "while writing Agent auth into config file")
	}
	return nil
}

// DeleteAgentAuthDetails deletes auth information.
func DeleteAgentAuthDetails(alias string) error {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if _, found := cfg.Agent[alias]; !found {
		return nil
	}

	delete(cfg.Agent, alias)
	viper.Set(agentAuthKeyPrefix, cfg.Agent)

	// Update default ctx, there is no order, just get the first one
	if cfg.Context == getAgentStoreKey(alias) {
		cfg.Context = getFirstFromMap(cfg.Agent)
		viper.Set(context, cfg.Context)
	}

	if err := viper.WriteConfig(); err != nil {
		return errors.Wrap(err, "while removing Agent auth from the config file")
	}
	return nil
}

// GetAgentAuthDetails returns information about Agent's auth method.
func GetAgentAuthDetails() (Agent, error) {
	defaultContext := viper.GetString(context)

	var out Agent
	if err := viper.UnmarshalKey(defaultContext, &out); err != nil {
		return Agent{}, err
	}
	if out.IsEmpty() {
		return Agent{}, errors.New("Not logged in to any server")
	}
	return out, nil
}

// GetAgentsAlias all available agents aliases.
func GetAgentsAlias() ([]string, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(cfg.Agent))
	for _, item := range cfg.Agent {
		out = append(out, item.Alias)
	}
	return out, nil
}

func getAgentStoreKey(alias string) string {
	return fmt.Sprintf("%s.%s", agentAuthKeyPrefix, alias)
}

func getFirstFromMap(in map[string]Agent) string {
	for _, item := range in {
		return getAgentStoreKey(item.Alias)
	}
	return ""
}
