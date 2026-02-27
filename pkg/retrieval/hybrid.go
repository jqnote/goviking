// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

// KeywordSearch performs keyword-based search using BM25 or simple term matching.
type KeywordSearch struct {
	// BM25 parameters
	k1 float64 // term frequency saturation parameter
	b  float64 // document length normalization parameter
}

// NewKeywordSearch creates a new KeywordSearch with default BM25 parameters.
func NewKeywordSearch() *KeywordSearch {
	return &KeywordSearch{
		k1: 1.5, // BM25 standard
		b:  0.75, // BM25 standard
	}
}

// BM25Result contains BM25 scoring information.
type BM25Result struct {
	URI       string
	Score     float64
	Matches   []string
	Frequency int
}

// Index contains term frequencies for documents.
type Index struct {
	Documents    map[string]string // URI -> content
	TermFreq     map[string]map[string]int // URI -> term -> frequency
	DocLengths   map[string]int // URI -> length
	AvgDocLength float64
	IDF         map[string]float64 // term -> IDF score
	TotalDocs   int
}

// NewIndex creates a new Index.
func NewIndex() *Index {
	return &Index{
		Documents:  make(map[string]string),
		TermFreq:   make(map[string]map[string]int),
		DocLengths: make(map[string]int),
		IDF:        make(map[string]float64),
	}
}

// AddDocument adds a document to the index.
func (idx *Index) AddDocument(uri, content string) {
	// Store document
	idx.Documents[uri] = content

	// Tokenize
	terms := tokenize(content)
	idx.DocLengths[uri] = len(terms)

	// Calculate term frequencies
	freq := make(map[string]int)
	for _, term := range terms {
		freq[term]++
	}
	idx.TermFreq[uri] = freq

	idx.TotalDocs++
}

// tokenize splits text into terms.
func tokenize(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove special characters and split
	reg := regexp.MustCompile(`[a-z0-9]+`)
	matches := reg.FindAllString(text, -1)

	return matches
}

// BuildIDF builds IDF scores for all terms.
func (idx *Index) BuildIDF() {
	// Count document frequency for each term
	docFreq := make(map[string]int)
	for _, freq := range idx.TermFreq {
		for term := range freq {
			docFreq[term]++
		}
	}

	// Calculate IDF for each term
	N := float64(idx.TotalDocs)
	for term, df := range docFreq {
		// IDF with smoothing
		idx.IDF[term] = math.Log((N - float64(df) + 0.5) / (float64(df) + 0.5) + 1)
	}

	// Calculate average document length
	var totalLen int
	for _, l := range idx.DocLengths {
		totalLen += l
	}
	if idx.TotalDocs > 0 {
		idx.AvgDocLength = float64(totalLen) / float64(idx.TotalDocs)
	}
}

// Score calculates BM25 score for a query against a document.
func (ks *KeywordSearch) Score(query string, idx *Index, uri string) float64 {
	terms := tokenize(query)
	docFreq := idx.TermFreq[uri]
	docLen := idx.DocLengths[uri]

	var score float64
	for _, term := range terms {
		tf := float64(docFreq[term])
		if tf == 0 {
			continue
		}

		idf := idx.IDF[term]

		// BM25 scoring formula
		numerator := tf * (ks.k1 + 1)
		denominator := tf + ks.k1*(1 - ks.b + ks.b*float64(docLen)/idx.AvgDocLength)
		score += idf * numerator / denominator
	}

	return score
}

// Search performs keyword search.
func (ks *KeywordSearch) Search(ctx context.Context, query string, idx *Index, limit int) []SearchResult {
	if idx.TotalDocs == 0 {
		return []SearchResult{}
	}

	var results []SearchResult

	for uri := range idx.Documents {
		score := ks.Score(query, idx, uri)
		if score > 0 {
			results = append(results, SearchResult{
				URI:    uri,
				Score:  score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// HybridSearch combines keyword and semantic search.
type HybridSearch struct {
	semanticSearch *SemanticSearch
	keywordSearch *KeywordSearch
	index         *Index
	alpha         float64 // weight for semantic search (1-alpha for keyword)
}

// NewHybridSearch creates a new HybridSearch.
func NewHybridSearch(semanticSearch *SemanticSearch, alpha float64) *HybridSearch {
	return &HybridSearch{
		semanticSearch: semanticSearch,
		keywordSearch:  NewKeywordSearch(),
		index:          NewIndex(),
		alpha:          alpha,
	}
}

// IndexDocuments indexes documents for keyword search.
func (hs *HybridSearch) IndexDocuments(ctx context.Context, documents []SearchResult) {
	for _, doc := range documents {
		hs.index.AddDocument(doc.URI, doc.Abstract)
	}
	hs.index.BuildIDF()
}

// Search performs hybrid search combining semantic and keyword search.
func (hs *HybridSearch) Search(ctx context.Context, query string, limit int, filter map[string]interface{}) ([]SearchResult, error) {
	var semanticResults []SearchResult
	var keywordResults []SearchResult
	var err error

	// Run semantic search
	if hs.semanticSearch != nil {
		semanticResults, err = hs.semanticSearch.Search(ctx, query, limit*2, filter)
		if err != nil {
			return nil, err
		}
	}

	// Run keyword search
	if hs.index.TotalDocs > 0 {
		keywordResults = hs.keywordSearch.Search(ctx, query, hs.index, limit*2)
	}

	// Merge results using RRF (Reciprocal Rank Fusion)
	combined := hs.rrfMerge(semanticResults, keywordResults, limit)

	// Normalize scores
	hs.normalizeScores(combined)

	return combined, nil
}

// rrfMerge merges results using Reciprocal Rank Fusion.
func (hs *HybridSearch) rrfMerge(semanticResults, keywordResults []SearchResult, limit int) []SearchResult {
	scores := make(map[string]float64)
	k := 60 // RRF parameter

	// Add semantic scores
	kFloat := float64(k)
	for rank, result := range semanticResults {
		scores[result.URI] += 1.0 / (float64(rank) + kFloat)
	}

	// Add keyword scores
	for rank, result := range keywordResults {
		scores[result.URI] += 1.0 / (float64(rank) + kFloat)
	}

	// Convert to results
	var results []SearchResult
	for uri, score := range scores {
		results = append(results, SearchResult{
			URI:   uri,
			Score: score,
		})
	}

	// Sort by combined score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// normalizeScores normalizes scores to 0-1 range.
func (hs *HybridSearch) normalizeScores(results []SearchResult) {
	if len(results) == 0 {
		return
	}

	maxScore := results[0].Score
	if maxScore == 0 {
		return
	}

	for i := range results {
		results[i].Score = results[i].Score / maxScore
	}
}

// SearchWithAlpha performs hybrid search with custom alpha.
func (hs *HybridSearch) SearchWithAlpha(ctx context.Context, query string, limit int, alpha float64, filter map[string]interface{}) ([]SearchResult, error) {
	oldAlpha := hs.alpha
	hs.alpha = alpha
	defer func() { hs.alpha = oldAlpha }()

	return hs.Search(ctx, query, limit, filter)
}

// HybridRetriever combines semantic and hotness scoring for retrieval.
type HybridRetriever struct {
	semanticSearch *SemanticSearch
	hotnessScorer *HotnessScorer
	alpha         float64 // weight for semantic score (1-alpha for hotness)
}

// NewHybridRetriever creates a new HybridRetriever.
func NewHybridRetriever(semanticSearch *SemanticSearch, hotnessScorer *HotnessScorer, alpha float64) *HybridRetriever {
	if alpha == 0 {
		alpha = 0.8 // Default: 80% semantic, 20% hotness
	}
	return &HybridRetriever{
		semanticSearch: semanticSearch,
		hotnessScorer: hotnessScorer,
		alpha:         alpha,
	}
}

// CombineScores combines semantic and hotness scores.
func (hr *HybridRetriever) CombineScores(semanticScore, hotnessScore float64) float64 {
	// Weighted combination: alpha * semantic + (1-alpha) * hotness
	return hr.alpha*semanticScore + (1-hr.alpha)*hotnessScore
}

// Retrieve performs hybrid retrieval with hotness scoring.
func (hr *HybridRetriever) Retrieve(ctx context.Context, query string, sessionID string, accessCount int, lastAccessTime time.Time, limit int) ([]SearchResult, error) {
	// Get semantic results
	var results []SearchResult
	var err error

	if hr.semanticSearch != nil {
		results, err = hr.semanticSearch.Search(ctx, query, limit*2, nil)
		if err != nil {
			return nil, err
		}
	}

	// Calculate hotness score
	hotnessScore := 0.0
	if hr.hotnessScorer != nil {
		hotnessScore = hr.hotnessScorer.CalculateHotness(accessCount, lastAccessTime)
	}

	// Combine scores
	for i := range results {
		results[i].Score = hr.CombineScores(results[i].Score, hotnessScore)
	}

	// Sort by combined score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Reranker re-ranks search results using cross-encoder.
type Reranker struct {
	client  interface{} // LLM provider for cross-encoder scoring
	enabled bool
}

// NewReranker creates a new Reranker.
func NewReranker(enabled bool) *Reranker {
	return &Reranker{
		enabled: enabled,
	}
}

// Rerank re-ranks search results.
func (r *Reranker) Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error) {
	if !r.enabled || len(results) == 0 {
		return results, nil
	}

	// Simple cross-encoder reranking: re-score each result with the query
	// In a real implementation, this would use a cross-encoder model
	var reranked []SearchResult
	for _, result := range results {
		// Calculate relevance score (simplified)
		relevance := r.calculateRelevance(query, result)
		result.Score = result.Score * 0.5 + relevance * 0.5
		reranked = append(reranked, result)
	}

	// Sort by new scores
	sort.Slice(reranked, func(i, j int) bool {
		return reranked[i].Score > reranked[j].Score
	})

	return reranked, nil
}

// calculateRelevance calculates relevance between query and result.
func (r *Reranker) calculateRelevance(query string, result SearchResult) float64 {
	// Simple relevance: count query terms in result content
	queryTerms := tokenize(query)
	contentTerms := tokenize(result.Abstract)

	matchCount := 0
	for _, qt := range queryTerms {
		for _, ct := range contentTerms {
			if qt == ct {
				matchCount++
				break
			}
		}
	}

	if len(queryTerms) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(queryTerms))
}

// RetrieveWithRerank performs hybrid retrieval with reranking.
func (hr *HybridRetriever) RetrieveWithRerank(ctx context.Context, query string, sessionID string, accessCount int, lastAccessTime time.Time, limit int, reranker *Reranker) ([]SearchResult, error) {
	// First pass: hybrid retrieval
	results, err := hr.Retrieve(ctx, query, sessionID, accessCount, lastAccessTime, limit*3)
	if err != nil {
		return nil, err
	}

	// Second pass: reranking
	if reranker != nil && reranker.enabled {
		results, err = reranker.Rerank(ctx, query, results)
		if err != nil {
			return nil, err
		}
	}

	// Limit final results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
