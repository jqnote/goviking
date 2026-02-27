// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"container/heap"
	"fmt"
	"sort"
	"sync"
	"time"
)

// RetrieverConfig contains configuration for the retriever.
type RetrieverConfig struct {
	// Maximum convergence rounds (stop after multiple rounds with unchanged topk)
	MaxConvergenceRounds int

	// Maximum relations per resource
	MaxRelations int

	// Score propagation coefficient
	ScorePropagationAlpha float64

	// Directory dominance ratio
	DirectoryDominanceRatio float64

	// Global search topk
	GlobalSearchTopK int

	// Default score threshold
	ScoreThreshold float64
}

// DefaultRetrieverConfig returns default retriever configuration.
func DefaultRetrieverConfig() RetrieverConfig {
	return RetrieverConfig{
		MaxConvergenceRounds:    3,
		MaxRelations:           5,
		ScorePropagationAlpha:  0.5,
		DirectoryDominanceRatio: 1.2,
		GlobalSearchTopK:       3,
		ScoreThreshold:         0.0,
	}
}

// SearchResultHeap implements heap.Interface for priority queue.
type SearchResultHeap []SearchResult

func (h SearchResultHeap) Len() int           { return len(h) }
func (h SearchResultHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h SearchResultHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *SearchResultHeap) Push(x interface{}) { *h = append(*h, x.(SearchResult)) }
func (h *SearchResultHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// RetrievalResult contains result from recursive search.
type RetrievalResult struct {
	URI       string
	Score     float64
	IsLeaf    bool
	Abstract  string
	ParentURI string
}

// HierarchicalRetriever implements hierarchical retrieval with directory traversal.
type HierarchicalRetriever struct {
	config      RetrieverConfig
	embedder    Embedder
	vectorStore VectorStore
	trajectory  *TrajectoryLogger
	hybridSearch *HybridSearch

	mu sync.RWMutex
}

// NewHierarchicalRetriever creates a new HierarchicalRetriever.
func NewHierarchicalRetriever(embedder Embedder, vectorStore VectorStore, config RetrieverConfig) *HierarchicalRetriever {
	var hs *HybridSearch
	if embedder != nil && vectorStore != nil {
		ss := NewSemanticSearch(embedder, vectorStore)
		hs = NewHybridSearch(ss, 0.5)
	}

	return &HierarchicalRetriever{
		config:       config,
		embedder:     embedder,
		vectorStore:  vectorStore,
		trajectory:   NewTrajectoryLogger(),
		hybridSearch: hs,
	}
}

// Retrieve performs hierarchical retrieval.
func (hr *HierarchicalRetriever) Retrieve(ctx context.Context, query TypedQuery, opts SearchOptions) (*QueryResult, error) {
	// Create trajectory
	trajectory := hr.trajectory.CreateTrajectory(query.Query)
	thinkingTrace := &ThinkingTrace{StartTime: time.Now()}

	// Determine target directories
	targetDirs := opts.TargetDirectories
	if len(targetDirs) == 0 {
		targetDirs = hr.getRootURIsForType(query.ContextType)
	}

	thinkingTrace.AddEvent(TraceEventSearchDirectoryStart,
		fmt.Sprintf("Starting retrieval for query: %s", query.Query),
		map[string]interface{}{
			"target_directories": targetDirs,
			"context_type":       query.ContextType,
		}, query.Query)

	// Generate query vector
	var queryVector *EmbedResult
	if hr.embedder != nil {
		var err error
		queryVector, err = hr.embedder.Embed(ctx, query.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to embed query: %w", err)
		}
	}

	// Global vector search to supplement starting points
	startingPoints := hr.getGlobalSearchResults(ctx, queryVector, targetDirs)

	// Merge starting points
	mergedPoints := hr.mergeStartingPoints(query.Query, targetDirs, startingPoints)

	// Recursive search
	candidates, err := hr.recursiveSearch(ctx, query.Query, queryVector, mergedPoints, opts, trajectory, thinkingTrace)
	if err != nil {
		return nil, fmt.Errorf("recursive search failed: %w", err)
	}

	// Convert to matched contexts
	matched := hr.convertToMatchedContexts(candidates, query.ContextType)

	thinkingTrace.AddEvent(TraceEventSearchSummary,
		fmt.Sprintf("Retrieval complete, found %d results", len(matched)),
		map[string]interface{}{
			"total_results":    len(matched),
			"searched_dirs":   len(targetDirs),
			"statistics":      thinkingTrace.GetStatistics(),
		}, query.Query)

	return &QueryResult{
		Query:               query,
		MatchedContexts:    matched,
		SearchedDirectories: targetDirs,
		ThinkingTrace:       thinkingTrace,
	}, nil
}

// getGlobalSearchResults performs global vector search.
func (hr *HierarchicalRetriever) getGlobalSearchResults(ctx context.Context, queryVector *EmbedResult, targetDirs []string) []SearchResult {
	if queryVector == nil || hr.vectorStore == nil {
		return []SearchResult{}
	}

	results, err := hr.vectorStore.Search(ctx, queryVector, hr.config.GlobalSearchTopK, nil)
	if err != nil {
		return []SearchResult{}
	}

	return results
}

// mergeStartingPoints merges global search results with target directories.
func (hr *HierarchicalRetriever) mergeStartingPoints(query string, rootURIs []string, globalResults []SearchResult) []HeapItem {
	items := make([]HeapItem, 0)
	seen := make(map[string]bool)

	// Add global search results
	for _, r := range globalResults {
		items = append(items, HeapItem{URI: r.URI, Score: r.Score})
		seen[r.URI] = true
	}

	// Add root directories
	for _, uri := range rootURIs {
		if !seen[uri] {
			items = append(items, HeapItem{URI: uri, Score: 0.0})
			seen[uri] = true
		}
	}

	return items
}

// recursiveSearch performs recursive directory search with score propagation.
func (hr *HierarchicalRetriever) recursiveSearch(
	ctx context.Context,
	query string,
	queryVector *EmbedResult,
	startingPoints []HeapItem,
	opts SearchOptions,
	trajectory *Trajectory,
	thinkingTrace *ThinkingTrace,
) ([]RetrievalResult, error) {

	type QueueItem struct {
		URI   string
		Score float64
	}

	dirQueue := &SearchResultHeap{}
	heap.Init(dirQueue)

	visited := make(map[string]bool)
	var collected []RetrievalResult
	prevTopKURIs := make(map[string]bool)
	convergenceRounds := 0
	depth := 0

	alpha := hr.config.ScorePropagationAlpha

	// Initialize queue with starting points
	for _, sp := range startingPoints {
		heap.Push(dirQueue, SearchResult{URI: sp.URI, Score: sp.Score})
	}

	for dirQueue.Len() > 0 {
		select {
		case <-ctx.Done():
			return collected, ctx.Err()
		default:
		}

		item := heap.Pop(dirQueue).(SearchResult)
		currentURI := item.URI
		currentScore := item.Score

		if visited[currentURI] {
			continue
		}
		visited[currentURI] = true

		// Add to trajectory
		trajectory.AddNode(currentURI, depth, currentScore, nil)

		thinkingTrace.AddEvent(TraceEventSearchDirectoryStart,
			fmt.Sprintf("Searching directory: %s", currentURI),
			map[string]interface{}{
				"uri":   currentURI,
				"score": currentScore,
			}, query)

		// Search children
		children, err := hr.searchChildren(ctx, currentURI, queryVector, opts.Limit*2)
		if err != nil {
			continue
		}

		for _, child := range children {
			// Calculate final score with propagation
			finalScore := alpha*child.Score + (1-alpha)*currentScore

			// Check threshold
			thresholdPassed := func() bool {
				if opts.ScoreGTE {
					return finalScore >= opts.ScoreThreshold
				}
				return finalScore > opts.ScoreThreshold
			}()

			if !thresholdPassed {
				thinkingTrace.AddEvent(TraceEventCandidateExcluded,
					fmt.Sprintf("Excluded %s (score %.4f below threshold %.4f)", child.URI, finalScore, opts.ScoreThreshold),
					map[string]interface{}{
						"uri":    child.URI,
						"score":  finalScore,
						"reason": "below_threshold",
					}, query)
				continue
			}

			// Check if already collected
			alreadyCollected := false
			for _, c := range collected {
				if c.URI == child.URI {
					alreadyCollected = true
					break
				}
			}

			if !alreadyCollected {
				collected = append(collected, RetrievalResult{
					URI:       child.URI,
					Score:     finalScore,
					IsLeaf:    child.IsLeaf,
					Abstract:  child.Abstract,
				})

				thinkingTrace.AddEvent(TraceEventCandidateSelected,
					fmt.Sprintf("Added %s to candidates (score: %.4f)", child.URI, finalScore),
					map[string]interface{}{
						"uri":   child.URI,
						"score": finalScore,
					}, query)
			}

			// Add non-leaf children to queue
			if !child.IsLeaf {
				heap.Push(dirQueue, SearchResult{URI: child.URI, Score: finalScore})
				trajectory.AddEdge(currentURI, child.URI)

				thinkingTrace.AddEvent(TraceEventDirectoryQueued,
					fmt.Sprintf("Queued subdirectory: %s", child.URI),
					map[string]interface{}{
						"uri":   child.URI,
						"score": finalScore,
					}, query)
			}
		}

		// Convergence check
		currentTopK := hr.getTopK(collected, opts.Limit)
		currentTopKURIs := make(map[string]bool)
		for _, c := range currentTopK {
			currentTopKURIs[c.URI] = true
		}

		if hr.mapsEqual(currentTopKURIs, prevTopKURIs) && len(currentTopKURIs) >= opts.Limit {
			convergenceRounds++
			thinkingTrace.AddEvent(TraceEventConvergenceCheck,
				fmt.Sprintf("Convergence round %d", convergenceRounds),
				map[string]interface{}{
					"round":       convergenceRounds,
					"topk_uris":   currentTopKURIs,
					"prev_topk":   prevTopKURIs,
				}, query)

			if convergenceRounds >= hr.config.MaxConvergenceRounds {
				thinkingTrace.AddEvent(TraceEventSearchConverged,
					"Search converged",
					map[string]interface{}{
						"rounds":       convergenceRounds,
						"total_found":  len(collected),
					}, query)
				break
			}
		} else {
			convergenceRounds = 0
		}
		prevTopKURIs = currentTopKURIs
		depth++
	}

	// Sort by score
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].Score > collected[j].Score
	})

	if len(collected) > opts.Limit {
		collected = collected[:opts.Limit]
	}

	return collected, nil
}

// searchChildren searches for children of a directory.
func (hr *HierarchicalRetriever) searchChildren(ctx context.Context, parentURI string, queryVector *EmbedResult, limit int) ([]SearchResult, error) {
	if hr.vectorStore == nil {
		return []SearchResult{}, nil
	}

	filter := map[string]interface{}{
		"parent_uri": parentURI,
	}

	results, err := hr.vectorStore.Search(ctx, queryVector, limit, filter)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// getTopK returns top k results by score.
func (hr *HierarchicalRetriever) getTopK(results []RetrievalResult, k int) []RetrievalResult {
	if k >= len(results) {
		return results
	}
	return results[:k]
}

// mapsEqual compares two string maps.
func (hr *HierarchicalRetriever) mapsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// getRootURIsForType returns root URIs for context type.
func (hr *HierarchicalRetriever) getRootURIsForType(contextType ContextType) []string {
	switch contextType {
	case ContextTypeMemory:
		return []string{"viking://user/memories", "viking://agent/memories"}
	case ContextTypeResource:
		return []string{"viking://resources"}
	case ContextTypeSkill:
		return []string{"viking://agent/skills"}
	default:
		return []string{}
	}
}

// convertToMatchedContexts converts retrieval results to matched contexts.
func (hr *HierarchicalRetriever) convertToMatchedContexts(candidates []RetrievalResult, contextType ContextType) []MatchedContext {
	results := make([]MatchedContext, 0, len(candidates))

	for _, c := range candidates {
		results = append(results, MatchedContext{
			URI:         c.URI,
			ContextType: contextType,
			IsLeaf:      c.IsLeaf,
			Abstract:    c.Abstract,
			Score:       c.Score,
		})
	}

	return results
}

// GetTrajectory returns the retrieval trajectory.
func (hr *HierarchicalRetriever) GetTrajectory(rootURI string) (*Trajectory, bool) {
	return hr.trajectory.GetTrajectory(rootURI)
}

// HeapItem is used for priority queue.
type HeapItem struct {
	URI   string
	Score float64
}
