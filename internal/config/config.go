package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TorchDuration int `yaml:"torch_duration_turns"`
}

func DefaultConfig() Config {
	return Config{
		TorchDuration: 10,
	}
}

func Load(path string) (Config, error) {
	config := DefaultConfig()

	configPath := path
	if configPath == "" {
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return config, fmt.Errorf("failed to get user home directory: %w", err)
			}
			xdgConfigHome = filepath.Join(homeDir, ".config")
		}
		configPath = filepath.Join(xdgConfigHome, "spells", "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return config, fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte{}, 0644); err != nil {
			return config, fmt.Errorf("failed to create config file: %w", err)
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		return config, nil
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}