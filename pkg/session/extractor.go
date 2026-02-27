// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"time"
)

// MemoryExtractor extracts important information from sessions.
type MemoryExtractor interface {
	// Extract extracts memories from session messages.
	Extract(ctx context.Context, messages []*Message) ([]*ExtractedMemory, error)
	// ExtractByCategory extracts memories for a specific category.
	ExtractByCategory(ctx context.Context, messages []*Message, category Category) ([]*ExtractedMemory, error)
	// ExtractAllCategories extracts memories from all categories.
	ExtractAllCategories(ctx context.Context, messages []*Message) (map[Category][]*ExtractedMemory, error)
}

// ExtractedMemory represents an extracted memory.
type ExtractedMemory struct {
	Content    string    `json:"content"`
	Importance float64   `json:"importance"`
	Category   string    `json:"category"`
	SessionID  string    `json:"session_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ExtractorConfig holds configuration for memory extraction.
type ExtractorConfig struct {
	MinImportance  float64   // Minimum importance threshold (0-1)
	MaxMemories    int       // Maximum memories to extract per batch
	SessionID      string    // Session ID for extracted memories
	UseNewCategories bool    // Use new 6-category system (profile, preference, entity, event, case, pattern)
}

// DefaultExtractorConfig returns default extractor configuration.
func DefaultExtractorConfig(sessionID string) ExtractorConfig {
	return ExtractorConfig{
		MinImportance: 0.5,
		MaxMemories:   10,
		SessionID:     sessionID,
	}
}

// Summarizer creates summaries of session content.
type Summarizer interface {
	// Summarize creates a summary of messages.
	Summarize(ctx context.Context, messages []*Message) (string, error)
	// Compress compresses messages into a summary.
	Compress(ctx context.Context, messages []*Message, maxTokens int) (string, int64, error)
}

// SummarizerConfig holds configuration for summarization.
type SummarizerConfig struct {
	MaxTokens      int   // Maximum tokens in summary
	KeepRecentMsgs int   // Number of recent messages to keep unchanged
}

// DefaultSummarizerConfig returns default summarizer configuration.
func DefaultSummarizerConfig() SummarizerConfig {
	return SummarizerConfig{
		MaxTokens:      1000,
		KeepRecentMsgs: 5,
	}
}

// Compressor compresses session content.
type Compressor interface {
	// Compress compresses content to fit within token limit.
	Compress(ctx context.Context, content string, maxTokens int) (string, int64, error)
}

// CompressionResult holds the result of compression.
type CompressionResult struct {
	Content      string `json:"content"`
	OriginalLen  int    `json:"original_len"`
	CompressedLen int    `json:"compressed_len"`
	TokensSaved  int64  `json:"tokens_saved"`
}

// Config holds configuration for session management.
type Config struct {
	// Session config
	SessionTimeout time.Duration // Session timeout duration
	MaxMessages    int           // Maximum messages before compression

	// Memory extraction config
	Extractor ExtractorConfig

	// Summarization config
	Summarizer SummarizerConfig

	// Compression config
	CompressionThreshold int // Messages before triggering compression
	CompressionRatio    float64
}

// DefaultConfig returns default session configuration.
func DefaultConfig() Config {
	return Config{
		SessionTimeout:      24 * time.Hour,
		MaxMessages:         100,
		Extractor:           ExtractorConfig{},
		Summarizer:         DefaultSummarizerConfig(),
		CompressionThreshold: 50,
		CompressionRatio:    0.5,
	}
}
