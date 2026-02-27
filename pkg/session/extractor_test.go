// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"testing"
	"time"

	"github.com/jqnote/goviking/pkg/llm"
)

// MockLLMProvider is a mock LLM provider for testing.
type MockLLMProvider struct {
	responses map[string]*llm.ChatResponse
}

func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{
		responses: make(map[string]*llm.ChatResponse),
	}
}

func (m *MockLLMProvider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// Return mock response based on request content
	content := ""
	for _, msg := range req.Messages {
		content += msg.Content
	}

	if resp, ok := m.responses[content]; ok {
		return resp, nil
	}

	// Default mock response
	return &llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Content: `[{"content": "Test memory", "importance": 0.8, "category": "preference"}]`,
				},
			},
		},
	}, nil
}

func (m *MockLLMProvider) ChatStream(ctx context.Context, req *llm.ChatRequest) (llm.StreamReader, error) {
	return nil, nil
}

func (m *MockLLMProvider) Embed(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return &llm.EmbeddingResponse{
		Data: []llm.Embedding{
			{
				Embedding: []float64{0.1, 0.2, 0.3},
				Index:     0,
			},
		},
	}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

func (m *MockLLMProvider) AddResponse(query string, response *llm.ChatResponse) {
	m.responses[query] = response
}

func TestLLMExtractorExtract(t *testing.T) {
	mock := NewMockLLMProvider()
	config := DefaultExtractorConfig("test-session")
	extractor := NewLLMExtractor(mock, config)

	messages := []*Message{
		{
			Role:      "user",
			Content:   "I prefer concise responses",
			CreatedAt: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   "I'll give you concise responses",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()
	memories, err := extractor.Extract(ctx, messages)

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(memories) == 0 {
		t.Fatal("Expected memories, got none")
	}

	if memories[0].SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", memories[0].SessionID)
	}
}

func TestLLMExtractorExtractByCategory(t *testing.T) {
	mock := NewMockLLMProvider()
	config := DefaultExtractorConfig("test-session")
	extractor := NewLLMExtractor(mock, config)

	messages := []*Message{
		{
			Role:      "user",
			Content:   "My name is John",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()
	memories, err := extractor.ExtractByCategory(ctx, messages, CategoryProfile)

	if err != nil {
		t.Fatalf("ExtractByCategory failed: %v", err)
	}

	// Should have category set to profile
	for _, m := range memories {
		if m.Category != string(CategoryProfile) {
			t.Errorf("Expected category 'profile', got '%s'", m.Category)
		}
	}
}

func TestLLMExtractorExtractAllCategories(t *testing.T) {
	mock := NewMockLLMProvider()
	config := DefaultExtractorConfig("test-session")
	extractor := NewLLMExtractor(mock, config)

	messages := []*Message{
		{
			Role:      "user",
			Content:   "I like Python and Go",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()
	results, err := extractor.ExtractAllCategories(ctx, messages)

	if err != nil {
		t.Fatalf("ExtractAllCategories failed: %v", err)
	}

	// Should return a map (may be empty if LLM returns empty)
	// Just verify the function works without error
	if results == nil {
		t.Error("Expected results map, got nil")
	}
}

func TestGetCategoryImportance(t *testing.T) {
	tests := []struct {
		category Category
		expected float64
	}{
		{CategoryProfile, 0.9},
		{CategoryPreference, 0.8},
		{CategoryEntity, 0.7},
		{CategoryEvent, 0.6},
		{CategoryCase, 0.7},
		{CategoryPattern, 0.5},
		{"unknown", 0.5}, // Default
	}

	for _, tt := range tests {
		result := GetCategoryImportance(tt.category)
		if result != tt.expected {
			t.Errorf("Category %s: expected %v, got %v", tt.category, tt.expected, result)
		}
	}
}

func TestCategoryPromptsExist(t *testing.T) {
	expectedCategories := []Category{
		CategoryProfile, CategoryPreference, CategoryEntity,
		CategoryEvent, CategoryCase, CategoryPattern,
	}

	for _, cat := range expectedCategories {
		if _, ok := CategoryPrompts[cat]; !ok {
			t.Errorf("Missing prompt for category: %s", cat)
		}
	}
}

func TestExtractorConfigDefaults(t *testing.T) {
	config := DefaultExtractorConfig("test-id")

	if config.MinImportance != 0.5 {
		t.Errorf("Expected MinImportance 0.5, got %v", config.MinImportance)
	}
	if config.MaxMemories != 10 {
		t.Errorf("Expected MaxMemories 10, got %v", config.MaxMemories)
	}
	if config.SessionID != "test-id" {
		t.Errorf("Expected SessionID 'test-id', got '%s'", config.SessionID)
	}
}

func TestDeduper(t *testing.T) {
	d := NewDeduper(0.8)

	// Use exact same content for duplicate detection
	memories := []*ExtractedMemory{
		{
			Content:    "User prefers concise responses",
			Importance: 0.8,
			Category:   "preference",
		},
		{
			Content:    "User prefers concise responses",
			Importance: 0.9,
			Category:   "preference",
		},
		{
			Content:    "User likes Python",
			Importance: 0.7,
			Category:   "preference",
		},
	}

	result := d.Dedup(memories)

	// First two should be merged (exact match), third should be kept
	if len(result) != 2 {
		t.Errorf("Expected 2 memories after dedup, got %d", len(result))
	}

	// Verify the higher importance one is kept
	found := false
	for _, m := range result {
		if m.Importance == 0.9 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to keep memory with highest importance")
	}
}

func TestAutoExtractor(t *testing.T) {
	mock := NewMockLLMProvider()
	config := Config{
		MaxMessages: 2,
		Extractor:   DefaultExtractorConfig("test"),
		Summarizer:  DefaultSummarizerConfig(),
	}

	ae := NewAutoExtractor(mock, config)

	msg := &Message{
		Role:      "user",
		Content:   "Hello",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	_, err := ae.AddMessage(ctx, msg)

	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	// Add another message to trigger extraction
	msg2 := &Message{
		Role:      "user",
		Content:   "World",
		CreatedAt: time.Now(),
	}

	_, err = ae.AddMessage(ctx, msg2)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	// Should have extracted
	if len(ae.GetMessages()) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(ae.GetMessages()))
	}
}

func TestAutoExtractorClear(t *testing.T) {
	mock := NewMockLLMProvider()
	config := Config{
		MaxMessages: 10,
		Extractor:   DefaultExtractorConfig("test"),
		Summarizer:  DefaultSummarizerConfig(),
	}

	ae := NewAutoExtractor(mock, config)

	msg := &Message{
		Role:      "user",
		Content:   "Test",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	ae.AddMessage(ctx, msg)

	ae.Clear()

	if len(ae.GetMessages()) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(ae.GetMessages()))
	}
}
