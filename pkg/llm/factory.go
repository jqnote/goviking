// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"fmt"
)

// ProviderType represents the type of LLM provider.
type ProviderType string

const (
	// ProviderOpenAI represents OpenAI.
	ProviderOpenAI ProviderType = "openai"
	// ProviderAnthropic represents Anthropic.
	ProviderAnthropic ProviderType = "anthropic"
)

// Config holds LLM provider configuration.
type Config struct {
	Type     ProviderType `json:"type"`
	APIKey   string      `json:"api_key"`
	BaseURL  string      `json:"base_url,omitempty"`
	Model    string      `json:"model"`
}

// NewProvider creates a new provider based on config.
func NewProvider(config Config) (Provider, error) {
	switch config.Type {
	case ProviderOpenAI:
		return NewOpenAIProvider(config.APIKey, config.BaseURL, config.Model), nil
	case ProviderAnthropic:
		return NewAnthropicProvider(config.APIKey, config.Model), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", config.Type)
	}
}
