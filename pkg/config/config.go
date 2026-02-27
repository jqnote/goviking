// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for GoViking.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for GoViking.
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server"`

	// Storage configuration
	Storage StorageConfig `mapstructure:"storage"`

	// LLM configuration
	LLM LLMConfig `mapstructure:"llm"`

	// Retrieval configuration
	Retrieval RetrievalConfig `mapstructure:"retrieval"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// StorageConfig holds storage configuration.
type StorageConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	InMemory bool   `mapstructure:"in_memory"`
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

// RetrievalConfig holds retrieval configuration.
type RetrievalConfig struct {
	EmbeddingModel string  `mapstructure:"embedding_model"`
	Similarity    float64 `mapstructure:"similarity_threshold"`
	MaxResults    int     `mapstructure:"max_results"`
}

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("storage.type", "sqlite")
	v.SetDefault("storage.path", "openviking.db")
	v.SetDefault("storage.in_memory", false)
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("retrieval.embedding_model", "text-embedding-3-small")
	v.SetDefault("retrieval.similarity_threshold", 0.7)
	v.SetDefault("retrieval.max_results", 10)

	// If config path provided, use it
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in default locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.goviking")
		v.AddConfigPath("/etc/goviking")
	}

	// Environment variables override
	v.SetEnvPrefix("GOVIKING")
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, use defaults
	}

	// Unmarshal to struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadDefault loads configuration with defaults.
func LoadDefault() (*Config, error) {
	return Load("")
}

// Save saves configuration to file.
func Save(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SaveYAML saves configuration as YAML string.
func SaveYAML(cfg *Config) (string, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetConfigPath returns the default config path.
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goviking", "config.yaml")
}
