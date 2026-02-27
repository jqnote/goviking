// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// SearchResult represents a search result.
type SearchResult struct {
	ID        string         `json:"id"`
	URI       string         `json:"uri"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Score     float64        `json:"score"`
	Type      string         `json:"type"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
}

// SearchService provides search functionality.
type SearchService struct {
	// Embed retrieval components (would be injected)
	hybridSearch interface{ /* HybridRetriever interface */ }

	// Personalization data
	personalization map[string]map[string]float64 // sessionID -> term -> boost
	mu              sync.RWMutex

	// Filter support
	typeIndex map[string][]string // type -> result IDs
}

// NewSearchService creates a new search service.
func NewSearchService() *SearchService {
	return &SearchService{
		personalization: make(map[string]map[string]float64),
		typeIndex:       make(map[string][]string),
	}
}

// SetHybridSearch sets the hybrid search implementation.
func (s *SearchService) SetHybridSearch(hs interface{}) {
	s.hybridSearch = hs
}

// SearchRequest represents a search request.
type SearchRequest struct {
	Query      string
	SessionID  string
	Filters    map[string]string
	Limit      int
	Offset     int
	Personalize bool
}

// Search performs a search.
func (s *SearchService) Search(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	results := s.basicSearch(ctx, req.Query, req.Limit)

	// Apply personalization if enabled
	if req.Personalize && req.SessionID != "" {
		results = s.applyPersonalization(ctx, results, req.SessionID)
	}

	// Apply filters
	if len(req.Filters) > 0 {
		results = s.applyFilters(ctx, results, req.Filters)
	}

	// Apply pagination
	if req.Offset > len(results) {
		return []SearchResult{}, nil
	}

	end := req.Offset + req.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[req.Offset:end], nil
}

// basicSearch performs a basic search (placeholder).
func (s *SearchService) basicSearch(ctx context.Context, query string, limit int) []SearchResult {
	// This would integrate with the actual retrieval system
	// For now, return placeholder results
	return []SearchResult{
		{
			ID:      uuid.New().String(),
			URI:     "/docs/example",
			Title:   "Example Document",
			Content: "Content matching: " + query,
			Score:   0.9,
			Type:    "document",
		},
	}
}

// applyPersonalization applies session-based personalization.
func (s *SearchService) applyPersonalization(ctx context.Context, results []SearchResult, sessionID string) []SearchResult {
	s.mu.RLock()
	boosts := s.personalization[sessionID]
	s.mu.RUnlock()

	if len(boosts) == 0 {
		return results
	}

	// Apply boosts based on session history
	for i := range results {
		for term, boost := range boosts {
			if contains(results[i].Content, term) || contains(results[i].Title, term) {
				results[i].Score *= (1 + boost)
			}
		}
	}

	return results
}

// applyFilters applies filters to search results.
func (s *SearchService) applyFilters(ctx context.Context, results []SearchResult, filters map[string]string) []SearchResult {
	var filtered []SearchResult

	for _, r := range results {
		match := true
		for key, value := range filters {
			switch key {
			case "type":
				if r.Type != value {
					match = false
				}
			case "session_id":
				if r.SessionID != value {
					match = false
				}
			default:
				// Check metadata
				if r.Metadata != nil {
					if v, ok := r.Metadata[key]; !ok || v != value {
						match = false
					}
				}
			}
		}
		if match {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// RecordSearch records a search for personalization.
func (s *SearchService) RecordSearch(ctx context.Context, sessionID string, query string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.personalization[sessionID] == nil {
		s.personalization[sessionID] = make(map[string]float64)
	}

	// Boost terms from this search for future personalization
	terms := tokenizeQuery(query)
	for _, term := range terms {
		s.personalization[sessionID][term] += 0.1
	}
}

// tokenizeQuery tokenizes a search query.
func tokenizeQuery(query string) []string {
	// Simple tokenization
	var tokens []string
	var current []rune

	for _, r := range query {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				tokens = append(tokens, string(current))
				current = nil
			}
		}
	}

	if len(current) > 0 {
		tokens = append(tokens, string(current))
	}

	return tokens
}

// contains is a simple string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IndexResult indexes a result for faster filtering.
func (s *SearchService) IndexResult(result SearchResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.typeIndex[result.Type] = append(s.typeIndex[result.Type], result.ID)
}

// ClearIndex clears the search index.
func (s *SearchService) ClearIndex() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.typeIndex = make(map[string][]string)
}
