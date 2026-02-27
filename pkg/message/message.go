// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package message provides message handling for GoViking.
package message

import (
	"encoding/json"
	"fmt"
)

// FormatType represents the message format type.
type FormatType string

const (
	// FormatOpenAI represents OpenAI format.
	FormatOpenAI FormatType = "openai"
	// FormatAnthropic represents Anthropic format.
	FormatAnthropic FormatType = "anthropic"
)

// Message represents a generic message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// FormatForProvider formats a message for a specific provider.
func FormatForProvider(msgs []Message, provider FormatType) (interface{}, error) {
	switch provider {
	case FormatOpenAI:
		return formatOpenAI(msgs)
	case FormatAnthropic:
		return formatAnthropic(msgs)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// formatOpenAI formats messages for OpenAI.
func formatOpenAI(msgs []Message) ([]Message, error) {
	return msgs, nil
}

// formatAnthropic formats messages for Anthropic.
func formatAnthropic(msgs []Message) ([]map[string]string, error) {
	result := make([]map[string]string, len(msgs))
	for i, msg := range msgs {
		result[i] = map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		}
	}
	return result, nil
}

// ParseResponse parses a response from a provider.
func ParseResponse(data []byte, provider FormatType) (*Response, error) {
	switch provider {
	case FormatOpenAI:
		return parseOpenAIResponse(data)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// Response represents a parsed response.
type Response struct {
	Content string
	Usage   Usage
}

// Usage represents token usage.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// parseOpenAIResponse parses an OpenAI response.
func parseOpenAIResponse(data []byte) (*Response, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content: resp.Choices[0].Message.Content,
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}
