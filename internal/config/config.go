package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server" json:"server"`
	Worker  WorkerConfig  `yaml:"worker" json:"worker"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

type ServerConfig struct {
	Port int `yaml:"port" json:"port"`
}

type WorkerConfig struct {
	Count int `yaml:"count" json:"count"`
}

type LoggingConfig struct {
	Level     string `yaml:"level" json:"level"`
	Format    string `yaml:"format" json:"format"`
	DebugMode bool   `yaml:"debug_mode" json:"debug_mode"`
}

// DefaultConfig returns default configuration values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Worker: WorkerConfig{
			Count: 3,
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "json",
			DebugMode: false,
		},
	}
}

// LoadConfig loads configuration from YAML file and environment variables
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	if err := loadFromYAML(config, "config.yaml"); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config.yaml: %w", err)
		}
	}

	loadFromEnv(config)

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromYAML loads configuration from a YAML file
func loadFromYAML(config *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if count := os.Getenv("WORKER_COUNT"); count != "" {
		if c, err := strconv.Atoi(count); err == nil && c > 0 {
			config.Worker.Count = c
		}
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = strings.ToLower(level)
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logging.Format = strings.ToLower(format)
	}
	if debug := os.Getenv("DEBUG"); debug != "" {
		config.Logging.DebugMode = debug == "true" || debug == "1"
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Worker.Count <= 0 {
		return fmt.Errorf("worker count must be positive: %d", config.Worker.Count)
	}

	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[config.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", config.Logging.Format)
	}

	return nil
}

// GetServerAddr returns the server address string
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// IsDebugMode returns true if debug mode is enabled
func (c *Config) IsDebugMode() bool {
	return c.Logging.DebugMode || c.Logging.Level == "debug"
}
