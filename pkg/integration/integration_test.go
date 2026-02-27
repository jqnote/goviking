// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jqnote/goviking/pkg/llm"
	"github.com/jqnote/goviking/pkg/retrieval"
	"github.com/jqnote/goviking/pkg/service"
	"github.com/jqnote/goviking/pkg/session"
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
			{Embedding: []float64{0.1, 0.2, 0.3}, Index: 0},
		},
	}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

// TestSessionHotnessIntegration tests Session Hotness functionality.
func TestSessionHotnessIntegration(t *testing.T) {
	// Create hotness scorer
	config := retrieval.HotnessConfig{
		Alpha:        0.2,
		HalfLifeDays: 7,
	}
	scorer := retrieval.NewHotnessScorer(config)

	// Test CalculateHotness
	tests := []struct {
		name         string
		accessCount  int
		lastAccess   time.Time
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:        "high access recent",
			accessCount: 100,
			lastAccess:  time.Now(),
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name:        "low access old",
			accessCount: 1,
			lastAccess:  time.Now().Add(-30 * 24 * time.Hour),
			expectedMin: 0.0,
			expectedMax: 0.3,
		},
		{
			name:        "medium access",
			accessCount: 10,
			lastAccess:  time.Now().Add(-7 * 24 * time.Hour),
			expectedMin: 0.3,
			expectedMax: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateHotness(tt.accessCount, tt.lastAccess)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("Expected score between %v and %v, got %v", tt.expectedMin, tt.expectedMax, score)
			}
		})
	}

	// Test HybridScore
	t.Run("HybridScore", func(t *testing.T) {
		hotnessScore := scorer.CalculateHotness(50, time.Now())
		hybridScore := scorer.HybridScore(0.8, hotnessScore)
		if hybridScore <= 0 || hybridScore > 1 {
			t.Errorf("Expected hybrid score in (0, 1], got %v", hybridScore)
		}
	})
}

// TestMemoryExtractionIntegration tests Memory Extraction functionality.
func TestMemoryExtractionIntegration(t *testing.T) {
	mockProvider := NewMockLLMProvider()
	config := session.DefaultExtractorConfig("test-session")
	extractor := session.NewLLMExtractor(mockProvider, config)

	messages := []*session.Message{
		{
			Role:      session.RoleUser,
			Content:   "I prefer concise responses",
			CreatedAt: time.Now(),
		},
		{
			Role:      session.RoleAssistant,
			Content:   "I'll keep responses brief",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()

	// Test Extract
	t.Run("Extract", func(t *testing.T) {
		memories, err := extractor.Extract(ctx, messages)
		if err != nil {
			t.Fatalf("Extract failed: %v", err)
		}
		if memories == nil {
			t.Error("Expected memories, got nil")
		}
	})

	// Test ExtractByCategory
	t.Run("ExtractByCategory", func(t *testing.T) {
		memories, err := extractor.ExtractByCategory(ctx, messages, session.CategoryPreference)
		if err != nil {
			t.Fatalf("ExtractByCategory failed: %v", err)
		}
		if memories == nil {
			t.Error("Expected memories, got nil")
		}
	})

	// Test ExtractAllCategories
	t.Run("ExtractAllCategories", func(t *testing.T) {
		results, err := extractor.ExtractAllCategories(ctx, messages)
		if err != nil {
			t.Fatalf("ExtractAllCategories failed: %v", err)
		}
		if results == nil {
			t.Error("Expected results, got nil")
		}
	})
}

// TestMemoryDeduplicationIntegration tests Memory Deduplication.
func TestMemoryDeduplicationIntegration(t *testing.T) {
	mockProvider := NewMockLLMProvider()
	deduper := session.NewMemoryDeduper(mockProvider, 0.8)

	memories := []*session.ExtractedMemory{
		{Content: "User likes Python", Importance: 0.8, Category: "preference"},
		{Content: "User likes Python", Importance: 0.9, Category: "preference"},
		{Content: "User prefers Go", Importance: 0.7, Category: "preference"},
	}

	ctx := context.Background()
	result, err := deduper.Dedup(ctx, memories)
	if err != nil {
		t.Fatalf("Dedup failed: %v", err)
	}

	if len(result) >= len(memories) {
		t.Errorf("Expected fewer memories after dedup, got %d -> %d", len(memories), len(result))
	}
}

// TestSessionCompressionIntegration tests Session Compression.
func TestSessionCompressionIntegration(t *testing.T) {
	mockProvider := NewMockLLMProvider()
	extractor := session.NewLLMExtractor(mockProvider, session.DefaultExtractorConfig("test"))
	deduper := session.NewMemoryDeduper(mockProvider, 0.8)

	config := session.DefaultCompressionConfig()
	config.Threshold = 3
	config.KeepRecent = 2

	compressor := session.NewSessionCompressor(extractor, deduper, nil, config)

	messages := []*session.Message{
		{Role: session.RoleUser, Content: "Hello", CreatedAt: time.Now()},
		{Role: session.RoleUser, Content: "World", CreatedAt: time.Now()},
		{Role: session.RoleUser, Content: "Test", CreatedAt: time.Now()},
		{Role: session.RoleUser, Content: "Data", CreatedAt: time.Now()},
	}

	ctx := context.Background()

	t.Run("ShouldCompress", func(t *testing.T) {
		if !compressor.ShouldCompress(5) {
			t.Error("Expected ShouldCompress to return true for 5 messages")
		}
		if compressor.ShouldCompress(2) {
			t.Error("Expected ShouldCompress to return false for 2 messages")
		}
	})

	t.Run("Compress", func(t *testing.T) {
		result, err := compressor.Compress(ctx, messages)
		if err != nil {
			t.Fatalf("Compress failed: %v", err)
		}
		if result == nil {
			t.Error("Expected compression result, got nil")
		}
	})
}

// TestDebugServiceIntegration tests DebugService.
func TestDebugServiceIntegration(t *testing.T) {
	debugSvc := service.NewDebugService()
	ctx := context.Background()

	t.Run("ComponentHealthCheck", func(t *testing.T) {
		status, err := debugSvc.ComponentHealthCheck(ctx, "queue")
		if err != nil {
			t.Fatalf("ComponentHealthCheck failed: %v", err)
		}
		if status.Name != "queue" {
			t.Errorf("Expected name 'queue', got '%s'", status.Name)
		}
	})

	t.Run("OverallStatus", func(t *testing.T) {
		statuses, err := debugSvc.OverallStatus(ctx)
		if err != nil {
			t.Fatalf("OverallStatus failed: %v", err)
		}
		if len(statuses) == 0 {
			t.Error("Expected some component statuses")
		}
	})

	t.Run("GetDetailedStatus", func(t *testing.T) {
		details, err := debugSvc.GetDetailedStatus(ctx)
		if err != nil {
			t.Fatalf("GetDetailedStatus failed: %v", err)
		}
		if details == nil {
			t.Error("Expected details, got nil")
		}
	})
}

// TestRelationServiceIntegration tests RelationService.
func TestRelationServiceIntegration(t *testing.T) {
	relationSvc := service.NewRelationService()
	ctx := context.Background()

	t.Run("CreateRelation", func(t *testing.T) {
		rel, err := relationSvc.CreateRelation(ctx, "user:1", "doc:1", "owns")
		if err != nil {
			t.Fatalf("CreateRelation failed: %v", err)
		}
		if rel.Source != "user:1" || rel.Target != "doc:1" {
			t.Errorf("Expected source='user:1', target='doc:1'")
		}
	})

	t.Run("GetRelated", func(t *testing.T) {
		_, _ = relationSvc.CreateRelation(ctx, "user:2", "doc:2", "owns")
		related, err := relationSvc.GetRelated(ctx, "user:2")
		if err != nil {
			t.Fatalf("GetRelated failed: %v", err)
		}
		if len(related) == 0 {
			t.Error("Expected some related resources")
		}
	})

	t.Run("DeleteRelation", func(t *testing.T) {
		_, _ = relationSvc.CreateRelation(ctx, "user:3", "doc:3", "owns")
		err := relationSvc.DeleteRelation(ctx, "user:3", "doc:3")
		if err != nil {
			t.Fatalf("DeleteRelation failed: %v", err)
		}
	})
}

// TestSearchServiceIntegration tests SearchService.
func TestSearchServiceIntegration(t *testing.T) {
	searchSvc := service.NewSearchService()
	ctx := context.Background()

	t.Run("Search", func(t *testing.T) {
		req := &service.SearchRequest{
			Query:      "test query",
			Limit:      10,
			Personalize: false,
		}
		results, err := searchSvc.Search(ctx, req)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if results == nil {
			t.Error("Expected results, got nil")
		}
	})

	t.Run("SearchWithFilters", func(t *testing.T) {
		req := &service.SearchRequest{
			Query:   "test",
			Filters: map[string]string{"type": "document"},
			Limit:   10,
		}
		results, err := searchSvc.Search(ctx, req)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if results == nil {
			t.Error("Expected results, got nil")
		}
	})

	t.Run("SearchWithPersonalization", func(t *testing.T) {
		req := &service.SearchRequest{
			Query:       "python",
			SessionID:   "session-1",
			Personalize: true,
			Limit:       10,
		}
		results, err := searchSvc.Search(ctx, req)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if results == nil {
			t.Error("Expected results, got nil")
		}
	})
}

// TestFSServiceIntegration tests FSService.
func TestFSServiceIntegration(t *testing.T) {
	testDir := "/tmp/goviking-integration-test"
	os.MkdirAll(testDir, 0755)
	fsSvc := service.NewFSService(testDir)
	ctx := context.Background()

	t.Run("Mkdir", func(t *testing.T) {
		err := fsSvc.Mkdir(ctx, testDir+"/subdir")
		if err != nil {
			t.Fatalf("Mkdir failed: %v", err)
		}
	})

	t.Run("WriteAndRead", func(t *testing.T) {
		testPath := testDir + "/test.txt"
		testContent := "Hello, World!"

		err := fsSvc.Write(ctx, testPath, testContent)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		content, err := fsSvc.Read(ctx, testPath)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if content != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, content)
		}
	})

	t.Run("List", func(t *testing.T) {
		files, err := fsSvc.List(ctx, testDir)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(files) == 0 {
			t.Error("Expected some files")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := fsSvc.Delete(ctx, testDir+"/test.txt")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	// Cleanup
	os.RemoveAll(testDir)
}

// TestPackServiceIntegration tests PackService.
func TestPackServiceIntegration(t *testing.T) {
	testDir := "/tmp/goviking-pack-test"
	os.MkdirAll(testDir, 0755)
	fsSvc := service.NewFSService(testDir)
	packSvc := service.NewPackService(fsSvc)
	ctx := context.Background()

	// Create test file
	testPath := testDir + "/test.txt"
	testContent := "Test content"
	_ = fsSvc.Write(ctx, testPath, testContent)

	t.Run("Export", func(t *testing.T) {
		data, err := packSvc.Export(ctx, []string{testPath})
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("Expected data, got empty")
		}
	})

	t.Run("Validate", func(t *testing.T) {
		data, _ := packSvc.Export(ctx, []string{testPath})
		valid, msg, err := packSvc.Validate(ctx, data)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if !valid {
			t.Errorf("Expected valid, got false: %s", msg)
		}
	})

	// Cleanup
	os.RemoveAll(testDir)
}
