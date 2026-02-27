// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AnthropicProvider implements Provider for Anthropic Claude.
type AnthropicProvider struct {
	APIKey   string
	BaseURL  string
	Model    string
	HTTPClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider.
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		APIKey:   apiKey,
		BaseURL:  "https://api.anthropic.com/v1",
		Model:    model,
		HTTPClient: &http.Client{},
	}
}

const anthropicAPIVersion = "2023-06-01"

// anthropicChatRequest converts to Anthropic format.
type anthropicChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Chat creates a chat completion.
func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.Model
	}

	url := p.BaseURL + "/messages"
	body, err := json.Marshal(anthropicChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Role      string `json:"role"`
		Content   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var content string
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}

	choices := []Choice{
		{
			Index: 0,
			Message: Message{
				Role:    RoleAssistant,
				Content: content,
			},
			FinishReason: result.StopReason,
		},
	}

	return &ChatResponse{
		ID:   result.ID,
		Model: req.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}, nil
}

// ChatStream creates a streaming chat completion.
func (p *AnthropicProvider) ChatStream(ctx context.Context, req *ChatRequest) (StreamReader, error) {
	if req.Model == "" {
		req.Model = p.Model
	}
	req.Stream = true

	url := p.BaseURL + "/messages"
	body, err := json.Marshal(anthropicChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	return &anthropicStreamReader{reader: resp.Body}, nil
}

// Embed creates embeddings - not supported by Anthropic.
func (p *AnthropicProvider) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Anthropic")
}

// Close closes the provider.
func (p *AnthropicProvider) Close() error {
	return nil
}

type anthropicStreamReader struct {
	reader io.Reader
}

func (r *anthropicStreamReader) Recv() (*StreamResponse, error) {
	scanner := bufio.NewScanner(r.reader)
	if !scanner.Scan() {
		return nil, io.EOF
	}

	line := scanner.Text()
	if !strings.HasPrefix(line, "data: ") {
		return nil, nil
	}

	data := strings.TrimPrefix(line, "data: ")
	if strings.HasPrefix(data, "{") {
		var event struct {
			Type string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		if event.Type == "content_block_delta" {
			return &StreamResponse{
				Choices: []StreamChoice{
					{
						Delta: Message{
							Content: event.Delta.Text,
						},
					},
				},
			}, nil
		}
	}

	return nil, nil
}

func (r *anthropicStreamReader) Close() error {
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
