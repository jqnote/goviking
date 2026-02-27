// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/jqnote/goviking/pkg/llm"
)

// MemoryDeduper handles memory deduplication with LLM-based decision making.
type MemoryDeduper struct {
	client           llm.Provider
	threshold       float64
	mergePromptTmpl string
}

// DedupDecision represents the decision for handling duplicate memories.
type DedupDecision string

const (
	DedupDecisionMerge   DedupDecision = "merge"   // Merge similar memories
	DedupDecisionCreate  DedupDecision = "create"  // Create new memory
	DedupDecisionDelete  DedupDecision = "delete"  // Delete duplicate
	DedupDecisionKeepBoth DedupDecision = "keep"   // Keep both memories
)

// NewMemoryDeduper creates a new memory deduper.
func NewMemoryDeduper(client llm.Provider, threshold float64) *MemoryDeduper {
	if threshold == 0 {
		threshold = 0.8 // 80% similarity threshold
	}
	return &MemoryDeduper{
		client:           client,
		threshold:       threshold,
		mergePromptTmpl: defaultMergePrompt,
	}
}

// Dedup performs deduplication on memories.
func (d *MemoryDeduper) Dedup(ctx context.Context, memories []*ExtractedMemory) ([]*ExtractedMemory, error) {
	if len(memories) <= 1 {
		return memories, nil
	}

	// First pass: simple similarity-based deduplication
	groups := d.groupSimilar(memories)

	// Second pass: LLM-based decision for each group
	var result []*ExtractedMemory
	for _, group := range groups {
		if len(group) <= 1 {
			result = append(result, group[0])
			continue
		}

		// Use LLM to decide how to handle the group
		decisions, err := d.decideMergeOrDelete(ctx, group)
		if err != nil {
			// Fall back to simple merge
			merged := d.simpleMerge(group)
			result = append(result, merged)
			continue
		}

		for i, decision := range decisions {
			switch decision {
			case DedupDecisionMerge, DedupDecisionKeepBoth:
				if i == 0 || decision == DedupDecisionMerge {
					result = append(result, group[i])
				}
			case DedupDecisionCreate:
				result = append(result, group[i])
			case DedupDecisionDelete:
				// Skip this memory
			}
		}
	}

	return result, nil
}

// groupSimilar groups similar memories together.
func (d *MemoryDeduper) groupSimilar(memories []*ExtractedMemory) [][]*ExtractedMemory {
	var groups [][]*ExtractedMemory

	for _, m := range memories {
		added := false
		for i, group := range groups {
			// Check if this memory is similar to any in the group
			if d.calculateSimilarity(m.Content, group[0].Content) >= d.threshold {
				groups[i] = append(group, m)
				added = true
				break
			}
		}
		if !added {
			groups = append(groups, []*ExtractedMemory{m})
		}
	}

	return groups
}

// calculateSimilarity calculates cosine similarity between two strings.
func (d *MemoryDeduper) calculateSimilarity(a, b string) float64 {
	// Simple word-based similarity
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	// Count common words
	wordCountA := make(map[string]int)
	wordCountB := make(map[string]int)

	for _, w := range wordsA {
		wordCountA[w]++
	}
	for _, w := range wordsB {
		wordCountB[w]++
	}

	// Calculate Jaccard similarity
	common := 0
	for w := range wordCountA {
		if _, ok := wordCountB[w]; ok {
			common++
		}
	}

	total := len(wordCountA) + len(wordCountB) - common
	if total == 0 {
		return 1
	}

	return float64(common) / float64(total)
}

// decideMergeOrDelete uses LLM to decide how to handle duplicate memories.
func (d *MemoryDeduper) decideMergeOrDelete(ctx context.Context, memories []*ExtractedMemory) ([]DedupDecision, error) {
	if len(memories) == 0 {
		return nil, nil
	}

	if len(memories) == 1 {
		return []DedupDecision{DedupDecisionCreate}, nil
	}

	// Build prompt with all memories
	var memList strings.Builder
	for i, m := range memories {
		memList.WriteString(fmt.Sprintf("[%d] %s (importance: %.2f, category: %s)\n",
			i+1, m.Content, m.Importance, m.Category))
	}

	prompt := fmt.Sprintf(d.mergePromptTmpl, memList.String())

	resp, err := d.client.Chat(ctx, &llm.ChatRequest{
		Model:       "",
		Temperature: 0.3,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a memory deduplication assistant. Analyze the memories and decide how to handle duplicates."},
			{Role: llm.RoleUser, Content: prompt},
		},
		MaxTokens: 500,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM decision: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	// Parse the response
	return d.parseDecisions(resp.Choices[0].Message.Content, len(memories))
}

// parseDecisions parses the LLM response into decisions.
func (d *MemoryDeduper) parseDecisions(response string, count int) ([]DedupDecision, error) {
	// Simple parsing: look for keywords in response
	decisions := make([]DedupDecision, count)

	// Default to keep the first one
	for i := range decisions {
		decisions[i] = DedupDecisionKeepBoth
	}

	lower := strings.ToLower(response)
	if strings.Contains(lower, "merge") {
		decisions[0] = DedupDecisionMerge
		for i := 1; i < count; i++ {
			decisions[i] = DedupDecisionDelete
		}
	} else if strings.Contains(lower, "keep") {
		// Keep all unique ones
	} else {
		// Default behavior: keep the highest importance one
		decisions[0] = DedupDecisionMerge
	}

	return decisions, nil
}

// simpleMerge merges multiple similar memories into one.
func (d *MemoryDeduper) simpleMerge(memories []*ExtractedMemory) *ExtractedMemory {
	if len(memories) == 0 {
		return nil
	}
	if len(memories) == 1 {
		return memories[0]
	}

	// Keep the one with highest importance
	best := memories[0]
	for _, m := range memories[1:] {
		if m.Importance > best.Importance {
			best = m
		}
	}

	return best
}

// MergeMemories merges two memories into one.
func (d *MemoryDeduper) MergeMemories(a, b *ExtractedMemory) (*ExtractedMemory, error) {
	if a.Category != b.Category {
		// Different categories, can't merge
		return nil, fmt.Errorf("cannot merge memories of different categories")
	}

	// Use the one with higher importance as base
	base := a
	if b.Importance > a.Importance {
		base = b
	}

	// Calculate combined importance (slightly reduced to avoid over-weighting)
	combinedImportance := math.Min(1.0, (a.Importance+b.Importance)*0.9)

	return &ExtractedMemory{
		Content:    base.Content,
		Importance: combinedImportance,
		Category:   base.Category,
		SessionID:  base.SessionID,
		CreatedAt:  base.CreatedAt,
	}, nil
}

const defaultMergePrompt = `Analyze the following memories and decide how to handle duplicates:

%s

For each memory, decide whether to:
- merge: Combine this memory with others (keep only the merged version)
- create: Keep this as a new unique memory
- delete: Remove this duplicate memory

Return your decision for each memory in order, one per line, starting with the decision keyword.
Example:
merge
delete
keep
`

// DedupConfig holds configuration for deduplication.
type DedupConfig struct {
	Threshold      float64 // Similarity threshold (0-1)
	UseLLM        bool     // Use LLM for merge decisions
	MaxGroupSize  int      // Maximum memories to process in one group
}

// DefaultDedupConfig returns default deduplication configuration.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		Threshold:     0.8,
		UseLLM:       true,
		MaxGroupSize: 10,
	}
}
