package config

import (
	"fmt"
	"healthy-api/model"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (%s): %w", path, err)
	}

	var cfg model.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &cfg, nil
}
