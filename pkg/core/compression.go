// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
)

// CompressionLevel represents the compression level.
type CompressionLevel int

const (
	CompressionLevelFast CompressionLevel = iota
	CompressionLevelDefault
	CompressionLevelBest
)

// CompressText compresses text using gzip.
func CompressText(text string) string {
	if text == "" {
		return ""
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return text
	}

	_, err = writer.Write([]byte(text))
	if err != nil {
		return text
	}

	err = writer.Close()
	if err != nil {
		return text
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// DecompressText decompresses gzip-compressed text.
func DecompressText(compressed string) (string, error) {
	if compressed == "" {
		return "", nil
	}

	// Try to decode as base64
	data, err := base64.StdEncoding.DecodeString(compressed)
	if err != nil {
		return compressed, nil // Not compressed, return as-is
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return compressed, nil
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		return compressed, err
	}

	return string(result), nil
}

// CompressWithLevel compresses text with specified compression level.
func CompressWithLevel(text string, level CompressionLevel) string {
	if text == "" {
		return ""
	}

	var buf bytes.Buffer
	var writer *gzip.Writer
	var err error

	switch level {
	case CompressionLevelFast:
		writer, err = gzip.NewWriterLevel(&buf, gzip.HuffmanOnly)
	case CompressionLevelBest:
		writer, err = gzip.NewWriterLevel(&buf, gzip.BestCompression)
	default:
		writer, err = gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	}

	if err != nil {
		return text
	}

	_, err = writer.Write([]byte(text))
	if err != nil {
		return text
	}

	err = writer.Close()
	if err != nil {
		return text
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// SummarizeText summarizes text to fit within a token limit.
func SummarizeText(text string, maxTokens int, tokenCounter TokenCounter) string {
	if text == "" {
		return ""
	}

	currentTokens := tokenCounter.CountTokens(text)
	if currentTokens <= maxTokens {
		return text
	}

	// Simple summarization: truncate to approximate token count
	// A more sophisticated implementation would use semantic extraction
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	// Estimate words to keep based on token ratio
	targetWords := len(words) * maxTokens / currentTokens
	if targetWords > len(words) {
		targetWords = len(words)
	}

	summary := strings.Join(words[:targetWords], " ")
	if targetWords < len(words) {
		summary += "..."
	}

	return summary
}

// TruncateText truncates text to fit within token limit.
func TruncateText(text string, maxTokens int, tokenCounter TokenCounter) string {
	if text == "" {
		return ""
	}

	currentTokens := tokenCounter.CountTokens(text)
	if currentTokens <= maxTokens {
		return text
	}

	// Binary search for the right length
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	low, high := 0, len(words)
	for low < high {
		mid := (low + high + 1) / 2
		testText := strings.Join(words[:mid], " ")
		if tokenCounter.CountTokens(testText) <= maxTokens {
			low = mid
		} else {
			high = mid - 1
		}
	}

	result := strings.Join(words[:low], " ")
	if low < len(words) {
		result += "..."
	}

	return result
}

// CompressibleContent represents content that can be compressed.
type CompressibleContent struct {
	Original    string
	Compressed string
	IsCompressed bool
}

// NewCompressibleContent creates a new CompressibleContent.
func NewCompressibleContent(original string) *CompressibleContent {
	return &CompressibleContent{
		Original:    original,
		Compressed: "",
		IsCompressed: false,
	}
}

// Compress compresses the content.
func (c *CompressibleContent) Compress() {
	if c.Original == "" || c.IsCompressed {
		return
	}
	c.Compressed = CompressText(c.Original)
	c.IsCompressed = true
}

// Decompress decompresses the content.
func (c *CompressibleContent) Decompress() (string, error) {
	if !c.IsCompressed {
		return c.Original, nil
	}
	return DecompressText(c.Compressed)
}

// Get returns the original or decompressed content.
func (c *CompressibleContent) Get() (string, error) {
	return c.Decompress()
}

// CompressionStats holds compression statistics.
type CompressionStats struct {
	OriginalSize  int `json:"original_size"`
	CompressedSize int `json:"compressed_size"`
	Ratio        float64 `json:"ratio"`
	TokensSaved  int `json:"tokens_saved"`
}

// CalculateStats calculates compression statistics.
func CalculateStats(original, compressed string, tokenCounter TokenCounter) *CompressionStats {
	stats := &CompressionStats{
		OriginalSize: len(original),
		CompressedSize: len(compressed),
	}

	if stats.OriginalSize > 0 {
		stats.Ratio = float64(stats.CompressedSize) / float64(stats.OriginalSize)
	}

	// Estimate tokens
	originalTokens := tokenCounter.CountTokens(original)
	compressedTokens := tokenCounter.CountTokens(compressed)
	stats.TokensSaved = originalTokens - compressedTokens

	return stats
}
