// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"fmt"
	"time"
)

// SessionCompressor handles session compression with extraction and deduplication.
type SessionCompressor struct {
	extractor MemoryExtractor
	deduper   *MemoryDeduper
	summarizer Summarizer
	config    CompressionConfig
}

// CompressionConfig holds configuration for session compression.
type CompressionConfig struct {
	Threshold     int           // Messages before triggering compression
	KeepRecent    int           // Number of recent messages to keep
	MaxTokens     int           // Maximum tokens after compression
	MinImportance float64       // Minimum importance to keep memories
	AutoExtract   bool          // Auto extract memories during compression
	AutoDedup     bool          // Auto deduplicate memories
	Interval      time.Duration // Compression check interval
}

// DefaultCompressionConfig returns default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Threshold:     50,
		KeepRecent:    5,
		MaxTokens:     4000,
		MinImportance: 0.3,
		AutoExtract:   true,
		AutoDedup:     true,
		Interval:      5 * time.Minute,
	}
}

// NewSessionCompressor creates a new session compressor.
func NewSessionCompressor(extractor MemoryExtractor, deduper *MemoryDeduper, summarizer Summarizer, config CompressionConfig) *SessionCompressor {
	if config.Threshold == 0 {
		config.Threshold = 50
	}
	if config.KeepRecent == 0 {
		config.KeepRecent = 5
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4000
	}

	return &SessionCompressor{
		extractor: extractor,
		deduper:   deduper,
		summarizer: summarizer,
		config:    config,
	}
}

// CompressionResult holds the result of session compression.
type SessionCompressionResult struct {
	MessagesCompressed int                   // Number of messages compressed
	MemoriesExtracted int                    // Number of memories extracted
	MemoriesRemoved   int                    // Number of duplicate memories removed
	TokensSaved       int64                  // Estimated tokens saved
	Summary           string                  // Summary if summarization was used
	ExtractedMemories []*ExtractedMemory     // Extracted memories
}

// Compress compresses session messages.
func (c *SessionCompressor) Compress(ctx context.Context, messages []*Message) (*SessionCompressionResult, error) {
	if len(messages) <= c.config.KeepRecent {
		return &SessionCompressionResult{}, nil
	}

	result := &SessionCompressionResult{}

	// Separate recent messages
	recentCount := c.config.KeepRecent
	if recentCount > len(messages) {
		recentCount = len(messages)
	}
	_ = messages[len(messages)-recentCount:]
	olderMsgs := messages[:len(messages)-recentCount]

	result.MessagesCompressed = len(olderMsgs)

	// Option 1: Extract important memories
	if c.config.AutoExtract && c.extractor != nil {
		memories, err := c.extractor.Extract(ctx, olderMsgs)
		if err != nil {
			return nil, fmt.Errorf("failed to extract memories: %w", err)
		}

		result.MemoriesExtracted = len(memories)

		// Option 2: Deduplicate memories
		if c.config.AutoDedup && c.deduper != nil && len(memories) > 1 {
			deduped, err := c.deduper.Dedup(ctx, memories)
			if err != nil {
				return nil, fmt.Errorf("failed to deduplicate: %w", err)
			}
			result.MemoriesRemoved = len(memories) - len(deduped)
			result.ExtractedMemories = deduped
		} else {
			result.ExtractedMemories = memories
		}
	}

	// Option 3: Summarize if still over token budget
	if c.summarizer != nil {
		estimatedTokens := estimateTokens(olderMsgs)
		if int64(estimatedTokens) > int64(c.config.MaxTokens) {
			summary, tokensSaved, err := c.summarizer.Compress(ctx, olderMsgs, c.config.MaxTokens)
			if err != nil {
				return nil, fmt.Errorf("failed to summarize: %w", err)
			}
			result.Summary = summary
			result.TokensSaved = tokensSaved
		}
	}

	return result, nil
}

// ShouldCompress checks if compression should be triggered.
func (c *SessionCompressor) ShouldCompress(messageCount int) bool {
	return messageCount >= c.config.Threshold
}

// CompressWithTrigger compresses messages if threshold is reached.
func (c *SessionCompressor) CompressWithTrigger(ctx context.Context, messages []*Message) (*SessionCompressionResult, bool, error) {
	if !c.ShouldCompress(len(messages)) {
		return nil, false, nil
	}

	result, err := c.Compress(ctx, messages)
	if err != nil {
		return nil, false, err
	}

	return result, true, nil
}

// estimateTokens estimates the number of tokens in messages.
func estimateTokens(messages []*Message) int {
	total := 0
	for _, msg := range messages {
		// Rough estimate: 1 token ≈ 4 characters
		total += len(msg.Content) / 4
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				total += len(tc.Function.Name) / 4
				total += len(tc.Function.Arguments) / 4
			}
		}
	}
	return total
}

// FilterMemoriesByImportance filters memories by minimum importance.
func (c *SessionCompressor) FilterMemoriesByImportance(memories []*ExtractedMemory) []*ExtractedMemory {
	var filtered []*ExtractedMemory
	for _, m := range memories {
		if m.Importance >= c.config.MinImportance {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// GetConfig returns the current compression configuration.
func (c *SessionCompressor) GetConfig() CompressionConfig {
	return c.config
}

// SetThreshold sets the compression threshold.
func (c *SessionCompressor) SetThreshold(threshold int) {
	c.config.Threshold = threshold
}

// SetMaxTokens sets the maximum tokens after compression.
func (c *SessionCompressor) SetMaxTokens(maxTokens int) {
	c.config.MaxTokens = maxTokens
}

// SetKeepRecent sets the number of recent messages to keep.
func (c *SessionCompressor) SetKeepRecent(keepRecent int) {
	c.config.KeepRecent = keepRecent
}
