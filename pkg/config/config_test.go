// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config is nil")
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected server.host 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server.port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Storage.Type != "sqlite" {
		t.Errorf("Expected storage.type 'sqlite', got '%s'", cfg.Storage.Type)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("Expected llm.provider 'openai', got '%s'", cfg.LLM.Provider)
	}
}

func TestSaveAndLoad(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 9000,
		},
		Storage: StorageConfig{
			Type:     "memory",
			InMemory: true,
		},
		LLM: LLMConfig{
			Provider: "anthropic",
			Model:    "claude-3",
		},
	}

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = Save(cfg, tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load it back
	loaded, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", loaded.Server.Host)
	}
	if loaded.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", loaded.Server.Port)
	}
	if loaded.Storage.Type != "memory" {
		t.Errorf("Expected storage.type 'memory', got '%s'", loaded.Storage.Type)
	}
	if loaded.LLM.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", loaded.LLM.Provider)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("Config path should not be empty")
	}
}
