package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/tdebuilt/nidus/internal/models"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Desktop  bool           `yaml:"-"`

	// Optional dashboard configuration (auto-imported on first startup if present)
	Settings   *models.YAMLSettings  `yaml:"settings,omitempty"`
	Categories []models.YAMLCategory `yaml:"categories,omitempty"`
	Services   []models.YAMLService  `yaml:"services,omitempty"`
}

// HasDashboardConfig returns true if the config file contains dashboard sections.
func (c Config) HasDashboardConfig() bool {
	return c.Settings != nil || len(c.Categories) > 0 || len(c.Services) > 0
}

// DashboardConfig returns the dashboard parts as a YAMLConfig.
func (c Config) DashboardConfig() models.YAMLConfig {
	cfg := models.YAMLConfig{
		Version:    2,
		Categories: c.Categories,
		Services:   c.Services,
	}
	if c.Settings != nil {
		cfg.Settings = *c.Settings
	}
	return cfg
}

type ServerConfig struct {
	Port        int    `yaml:"port"`
	BaseURL     string `yaml:"base_url"`
	BindAddress string `yaml:"bind_address"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:    3777,
			BaseURL: "http://localhost:3777",
		},
		Database: DatabaseConfig{
			Path: "./data/nidus.db",
		},
	}
}

// Load reads config from a YAML file, then applies environment variable overrides.
// If the file does not exist, defaults are used.
func Load(path string) (Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("reading config file: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parsing config file: %w", err)
		}
	}

	applyEnvOverrides(&cfg)

	if err := validate(cfg); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("NIDUS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("NIDUS_BASE_URL"); v != "" {
		cfg.Server.BaseURL = v
	}
	if v := os.Getenv("NIDUS_DB_PATH"); v != "" {
		cfg.Database.Path = v
	}
}

func validate(cfg Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Server.Port)
	}
	if cfg.Database.Path == "" {
		return fmt.Errorf("database path must not be empty")
	}
	return nil
}
