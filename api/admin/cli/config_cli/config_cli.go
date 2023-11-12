package config_cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wissance/Ferrum/api/admin/cli/config_cli/errors_config_cli"
	"github.com/wissance/Ferrum/api/admin/cli/domain"
	"github.com/wissance/Ferrum/config"
)

type (
	Config struct {
		Parameters
		DataSourceConfig config.DataSourceConfig
		LoggingConfig    config.LoggingConfig
	}

	Parameters struct {
		Operation   domain.OperationType
		Resource    domain.ResourceType
		Resource_id string
		Params      string
		Value       []byte
	}

	configs struct {
		DataSourceConfig config.DataSourceConfig `json:"data_source"`
		LoggingConfig    config.LoggingConfig    `json:"logging"`
	}
)

func NewConfig() (*Config, error) {
	configParameters := parseCmdParameters()

	configs, err := getConfigs()
	if err != nil {
		return nil, fmt.Errorf("getConfigs failed: %w", err)
	}

	cfg := &Config{
		Parameters:       *configParameters,
		DataSourceConfig: configs.DataSourceConfig,
		LoggingConfig:    configs.LoggingConfig,
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate failed: %w", err)
	}

	return cfg, nil
}

func parseCmdParameters() *Parameters {
	Operation := flag.String("operation", "", "")
	Resource := flag.String("resource", "", "")
	Resource_id := flag.String("resource_id", "", "")
	Params := flag.String("params", "", "")
	Value := flag.String("value", "", "")

	flag.Parse()

	configParameters := &Parameters{
		Operation:   domain.OperationType(*Operation),
		Resource:    domain.ResourceType(*Resource),
		Resource_id: *Resource_id,
		Params:      *Params,
		Value:       []byte(*Value),
	}

	return configParameters
}

func getConfigs() (*configs, error) {
	pathToConfig, err := getPathToConfig()
	if err != nil {
		return nil, fmt.Errorf("getPathToConfig failed: %w", err)
	}
	fileData, err := os.ReadFile(pathToConfig)
	if err != nil {
		return nil, fmt.Errorf("an error occurred during config file reading: %w", err)
	}

	var cliConfig configs
	if err = json.Unmarshal(fileData, &cliConfig); err != nil {
		return nil, fmt.Errorf("an error occurred during config file unmarshal: %w", err)
	}

	return &cliConfig, nil
}

func getPathToConfig() (string, error) {
	pathToExecutable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("an error occurred during path to executable file: %w", err)
	}
	pathToDirWithCurrentlyExecutable := filepath.Dir(pathToExecutable)
	pathToConfig := fmt.Sprintf("%s/config.json", pathToDirWithCurrentlyExecutable)
	return pathToConfig, nil
}

func validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("validate failed: %w", errors_config_cli.ErrNil)
	}

	if cfg.Resource != domain.RealmResource && cfg.Resource != domain.ClientResource && cfg.Resource != domain.UserResource {
		return fmt.Errorf("bad Resource \"%s\": %w", cfg.Resource, errors_config_cli.ErrBadResource)
	}

	if cfg.Operation != domain.GetOperation && cfg.Operation != domain.CreateOperation &&
		cfg.Operation != domain.DeleteOperation && cfg.Operation != domain.UpdateOperation {
		return fmt.Errorf("bad Operation \"%s\": %w", cfg.Operation, errors_config_cli.ErrBadOperation)
	}

	return nil
}
