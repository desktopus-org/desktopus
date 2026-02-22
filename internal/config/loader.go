package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadDesktop reads and parses a desktopus.yaml file
func LoadDesktop(path string) (*DesktopConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading desktop config: %w", err)
	}

	var cfg DesktopConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing desktop config: %w", err)
	}

	if err := ValidateDesktop(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FindDesktopConfig looks for desktopus.yaml in the given directory or path.
// If path is a directory, it looks for desktopus.yaml inside it.
// If path is a file, it uses it directly.
func FindDesktopConfig(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("accessing path %q: %w", path, err)
	}

	if info.IsDir() {
		configPath := filepath.Join(path, "desktopus.yaml")
		if _, err := os.Stat(configPath); err != nil {
			return "", fmt.Errorf("no desktopus.yaml found in %q", path)
		}
		return configPath, nil
	}

	return path, nil
}

// LoadApp reads and parses the app config file.
// Returns defaults if the file doesn't exist.
func LoadApp(path string) (*AppConfig, error) {
	cfg := DefaultAppConfig()

	if path == "" {
		return cfg, nil
	}

	// Expand ~ in path
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil
		}
		path = filepath.Join(home, path[1:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading app config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing app config: %w", err)
	}

	return cfg, nil
}
