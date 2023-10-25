package cli_config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wissance/Ferrum/config"
	"github.com/wissance/stringFormatter"
)

type (
	// Config -.
	Config struct {
		Parameters
		DataSourceConfig config.DataSourceConfig
		LoggingConfig    config.LoggingConfig
	}

	// Parameters -.
	Parameters struct {
		Operation   string
		Resource    string
		Namespace   string
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
		return nil, err
	}

	configs.DataSourceConfig.Options[config.Namespace] = configParameters.Namespace
	cfg := &Config{
		Parameters:       *configParameters,
		DataSourceConfig: configs.DataSourceConfig,
		LoggingConfig:    configs.LoggingConfig,
	}

	return cfg, nil
}

func parseCmdParameters() *Parameters {
	Operation := flag.String("operation", "", "")
	Resource := flag.String("resource", "", "")
	Namespace := flag.String("namespace", "", "")
	Resource_id := flag.String("resource_id", "", "")
	Params := flag.String("params", "", "")
	Value := flag.String("value", "", "")

	flag.Parse()

	configParameters := &Parameters{
		Operation:   *Operation,
		Resource:    *Resource,
		Namespace:   *Namespace,
		Resource_id: *Resource_id,
		Params:      *Params,
		Value:       []byte(*Value),
	}

	return configParameters
}

func getConfigs() (*configs, error) {
	absPath, err := filepath.Abs("./api/admin/cli/config.json")
	if err != nil {
		return nil, fmt.Errorf(stringFormatter.Format("An error occurred during getting config file abs path: {0}", err.Error()))
	}
	fileData, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf(stringFormatter.Format("An error occurred during config file reading: {0}", err.Error()))
	}

	cliConfig := &configs{}
	if err = json.Unmarshal(fileData, cliConfig); err != nil {
		return nil, fmt.Errorf(stringFormatter.Format("An error occurred during config file unmarshal: {0}", err.Error()))
	}

	return cliConfig, nil
}
