// Package config provides configuration management for the VJVector database.
// It handles loading and parsing of YAML configuration files with sensible defaults.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Embedding EmbeddingConfig `yaml:"embedding"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	DataDir    string `yaml:"data_dir"`
	MaxVectors int    `yaml:"max_vectors"`
	IndexType  string `yaml:"index_type"`
}

// EmbeddingConfig holds embedding model configuration
type EmbeddingConfig struct {
	ModelName string `yaml:"model_name"`
	Dimension int    `yaml:"dimension"`
	BatchSize int    `yaml:"batch_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads configuration from a file
func Load(configPath string) (*Config, error) {
	// Validate configPath to prevent directory traversal attacks
	if configPath == "" || strings.Contains(configPath, "..") || strings.Contains(configPath, "/") || strings.Contains(configPath, "\\") {
		return nil, fmt.Errorf("invalid config path: %s", configPath)
	}
	// Only allow simple filenames in current directory (no path separators)
	if strings.Contains(configPath, "/") || strings.Contains(configPath, "\\") {
		return nil, fmt.Errorf("config path must be a simple filename: %s", configPath)
	}
	// nolint:gosec // Path is validated above to prevent directory traversal
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	setDefaults(&config)

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Database.DataDir == "" {
		config.Database.DataDir = "./data"
	}
	if config.Database.MaxVectors == 0 {
		config.Database.MaxVectors = 1000000
	}
	if config.Database.IndexType == "" {
		config.Database.IndexType = "hnsw"
	}
	if config.Embedding.ModelName == "" {
		config.Embedding.ModelName = "text-embedding-ada-002"
	}
	if config.Embedding.Dimension == 0 {
		config.Embedding.Dimension = 1536
	}
	if config.Embedding.BatchSize == 0 {
		config.Embedding.BatchSize = 100
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
}
