// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package retrieval provides context retrieval with semantic search and directory traversal.
package retrieval

import (
	"time"
)

// ContextType represents the type of context.
type ContextType string

const (
	ContextTypeMemory   ContextType = "memory"
	ContextTypeResource ContextType = "resource"
	ContextTypeSkill    ContextType = "skill"
)

// RetrieverMode defines the retrieval mode.
type RetrieverMode string

const (
	RetrieverModeThinking RetrieverMode = "thinking"
	RetrieverModeQuick    RetrieverMode = "quick"
)

// TypedQuery represents a query targeting a specific context type.
type TypedQuery struct {
	Query              string       `json:"query"`
	ContextType        ContextType  `json:"context_type"`
	Intent             string       `json:"intent"`
	Priority           int          `json:"priority"`
	TargetDirectories []string     `json:"target_directories,omitempty"`
}

// QueryPlan contains multiple TypedQueries.
type QueryPlan struct {
	Queries        []TypedQuery `json:"queries"`
	SessionContext string       `json:"session_context"`
	Reasoning      string       `json:"reasoning"`
}

// RelatedContext represents related context with summary.
type RelatedContext struct {
	URI      string `json:"uri"`
	Abstract string `json:"abstract"`
}

// MatchedContext represents matched context from retrieval.
type MatchedContext struct {
	URI         string           `json:"uri"`
	ContextType ContextType      `json:"context_type"`
	IsLeaf      bool             `json:"is_leaf"`
	Abstract    string           `json:"abstract"`
	Overview    string           `json:"overview,omitempty"`
	Category    string           `json:"category"`
	Score       float64          `json:"score"`
	MatchReason string           `json:"match_reason,omitempty"`
	Relations   []RelatedContext `json:"relations,omitempty"`
}

// QueryResult represents result for a single TypedQuery.
type QueryResult struct {
	Query              TypedQuery        `json:"query"`
	MatchedContexts    []MatchedContext  `json:"matched_contexts"`
	SearchedDirectories []string         `json:"searched_directories"`
	ThinkingTrace     *ThinkingTrace    `json:"thinking_trace,omitempty"`
}

// FindResult represents final result from search.
type FindResult struct {
	Memories    []MatchedContext `json:"memories"`
	Resources   []MatchedContext `json:"resources"`
	Skills      []MatchedContext `json:"skills"`
	QueryPlan   *QueryPlan       `json:"query_plan,omitempty"`
	QueryResults []QueryResult   `json:"query_results,omitempty"`
	Total       int              `json:"total"`
}

// TraceEventType represents types of trace events.
type TraceEventType string

const (
	TraceEventSearchDirectoryStart   TraceEventType = "search_directory_start"
	TraceEventSearchDirectoryResult  TraceEventType = "search_directory_result"
	TraceEventEmbeddingScores       TraceEventType = "embedding_scores"
	TraceEventRerankScores          TraceEventType = "rerank_scores"
	TraceEventCandidateSelected     TraceEventType = "candidate_selected"
	TraceEventCandidateExcluded     TraceEventType = "candidate_excluded"
	TraceEventDirectoryQueued       TraceEventType = "directory_queued"
	TraceEventConvergenceCheck      TraceEventType = "convergence_check"
	TraceEventSearchConverged       TraceEventType = "search_converged"
	TraceEventSearchSummary         TraceEventType = "search_summary"
)

// TraceEvent represents a single trace event.
type TraceEvent struct {
	EventType TraceEventType         `json:"event_type"`
	Timestamp float64                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	QueryID   string                 `json:"query_id,omitempty"`
}

// ThinkingTrace captures the retrieval decision process.
type ThinkingTrace struct {
	StartTime time.Time   `json:"start_time"`
	Events    []TraceEvent `json:"events"`
}

// AddEvent adds a trace event.
func (t *ThinkingTrace) AddEvent(eventType TraceEventType, message string, data map[string]interface{}, queryID string) {
	if t.StartTime.IsZero() {
		t.StartTime = time.Now()
	}
	event := TraceEvent{
		EventType: eventType,
		Timestamp: time.Since(t.StartTime).Seconds(),
		Message:   message,
		Data:      data,
		QueryID:   queryID,
	}
	t.Events = append(t.Events, event)
}

// GetStatistics returns summary statistics from events.
func (t *ThinkingTrace) GetStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"total_events":            len(t.Events),
		"duration_seconds":        0.0,
		"directories_searched":    0,
		"candidates_collected":    0,
		"candidates_excluded":     0,
		"convergence_rounds":      0,
	}
	if len(t.Events) > 0 {
		stats["duration_seconds"] = t.Events[len(t.Events)-1].Timestamp
	}
	for _, event := range t.Events {
		switch event.EventType {
		case TraceEventSearchDirectoryResult:
			stats["directories_searched"] = stats["directories_searched"].(int) + 1
		case TraceEventCandidateSelected:
			count, _ := event.Data["count"].(int)
			stats["candidates_collected"] = stats["candidates_collected"].(int) + count
		case TraceEventCandidateExcluded:
			count, _ := event.Data["count"].(int)
			stats["candidates_excluded"] = stats["candidates_excluded"].(int) + count
		case TraceEventConvergenceCheck:
			round, _ := event.Data["round"].(int)
			stats["convergence_rounds"] = round
		}
	}
	return stats
}

// SearchOptions contains options for retrieval operations.
type SearchOptions struct {
	Limit             int
	Mode              RetrieverMode
	ScoreThreshold    float64
	ScoreGTE          bool
	TargetDirectories []string
	MetadataFilter    map[string]interface{}
}

// DefaultSearchOptions returns default search options.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Limit:          5,
		Mode:           RetrieverModeThinking,
		ScoreThreshold: 0.0,
		ScoreGTE:       false,
	}
}
