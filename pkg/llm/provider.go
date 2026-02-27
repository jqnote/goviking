// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package llm provides multi-provider LLM integration.
package llm

import (
	"context"
)

// Role represents the role of a message.
type Role string

const (
	// RoleSystem represents the system.
	RoleSystem Role = "system"
	// RoleUser represents the user.
	RoleUser Role = "user"
	// RoleAssistant represents the assistant.
	RoleAssistant Role = "assistant"
	// RoleTool represents a tool.
	RoleTool Role = "tool"
)

// Message represents a chat message.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage   `json:"usage"`
}

// Choice represents a chat completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

// StreamResponse represents a streaming chat response.
type StreamResponse struct {
	ID      string          `json:"id"`
	Model   string          `json:"model"`
	Choices []StreamChoice  `json:"choices"`
}

// StreamChoice represents a streaming choice.
type StreamChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

// EmbeddingRequest represents an embedding request.
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"` // string or []string
}

// EmbeddingResponse represents an embedding response.
type EmbeddingResponse struct {
	Data []Embedding `json:"data"`
	Usage Usage     `json:"usage"`
}

// Embedding represents a single embedding.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// Provider represents an LLM provider interface.
type Provider interface {
	// Chat creates a chat completion.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	// ChatStream creates a streaming chat completion.
	ChatStream(ctx context.Context, req *ChatRequest) (StreamReader, error)
	// Embed creates embeddings.
	Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	// Close closes the provider.
	Close() error
}

// StreamReader reads streaming responses.
type StreamReader interface {
	// Recv receives the next streaming response.
	Recv() (*StreamResponse, error)
	// Close closes the stream.
	Close() error
}
