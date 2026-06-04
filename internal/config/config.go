package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBPath             string `json:"dbPath"`
	UseCredentialStore bool   `json:"useCredentialStore"`
	PromptCharCount    int    `json:"promptCharCount"`
	IdleTimeoutSeconds int    `json:"idleTimeoutSeconds"`
}

func Default() *Config {
	return &Config{
		PromptCharCount:    3,
		IdleTimeoutSeconds: 300,
	}
}

func configPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "keepassview", "config.json"), nil
}

func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no config found (run keepassview init)")
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if cfg.DBPath == "" {
		return nil, fmt.Errorf("config missing dbPath (run keepassview init)")
	}
	if cfg.IdleTimeoutSeconds < 30 {
		cfg.IdleTimeoutSeconds = 30
	}
	if cfg.PromptCharCount < 3 {
		cfg.PromptCharCount = 3
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
