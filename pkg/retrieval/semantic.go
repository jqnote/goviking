// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"math"
	"sort"
	"sync"
)

// SearchResult represents a search result with score.
type SearchResult struct {
	URI       string                 `json:"uri"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Abstract  string                 `json:"abstract,omitempty"`
	IsLeaf    bool                   `json:"is_leaf"`
	ParentURI string                 `json:"parent_uri,omitempty"`
}

// VectorStore defines interface for vector storage and search.
type VectorStore interface {
	// Search performs vector similarity search.
	Search(ctx context.Context, query *EmbedResult, limit int, filter map[string]interface{}) ([]SearchResult, error)

	// Add adds vectors to the store.
	Add(ctx context.Context, vectors []SearchResult) error

	// Delete deletes vectors from the store.
	Delete(ctx context.Context, uris []string) error

	// Close closes the store.
	Close() error
}

// SemanticSearch performs semantic search using vector embeddings.
type SemanticSearch struct {
	embedder  Embedder
	vectorStore VectorStore
}

// NewSemanticSearch creates a new SemanticSearch.
func NewSemanticSearch(embedder Embedder, vectorStore VectorStore) *SemanticSearch {
	return &SemanticSearch{
		embedder:    embedder,
		vectorStore: vectorStore,
	}
}

// Search performs semantic search.
func (ss *SemanticSearch) Search(ctx context.Context, query string, limit int, filter map[string]interface{}) ([]SearchResult, error) {
	// Embed the query
	embedResult, err := ss.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	// Search vector store
	results, err := ss.vectorStore.Search(ctx, embedResult, limit, filter)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// SearchBatch performs batch semantic search.
func (ss *SemanticSearch) SearchBatch(ctx context.Context, queries []string, limit int) ([][]SearchResult, error) {
	// Embed all queries
	embedResults, err := ss.embedder.EmbedBatch(ctx, queries)
	if err != nil {
		return nil, err
	}

	// Search for each query
	results := make([][]SearchResult, len(queries))
	for i, embedResult := range embedResults {
		result, err := ss.vectorStore.Search(ctx, embedResult, limit, nil)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}

	return results, nil
}

// CosineSimilarity calculates cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// EuclideanDistance calculates Euclidean distance between two vectors.
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// DotProduct calculates dot product between two vectors.
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

// InMemoryVectorStore is a simple in-memory vector store.
type InMemoryVectorStore struct {
	vectors map[string][]float64
	metadata map[string]map[string]interface{}
	dimension int
	mu       sync.RWMutex
}

// NewInMemoryVectorStore creates a new InMemoryVectorStore.
func NewInMemoryVectorStore(dimension int) *InMemoryVectorStore {
	return &InMemoryVectorStore{
		vectors:  make(map[string][]float64),
		metadata: make(map[string]map[string]interface{}),
		dimension: dimension,
	}
}

// Search implements VectorStore interface.
func (vs *InMemoryVectorStore) Search(ctx context.Context, query *EmbedResult, limit int, filter map[string]interface{}) ([]SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if query.DenseVector == nil || len(query.DenseVector) == 0 {
		return []SearchResult{}, nil
	}

	var results []SearchResult

	for uri, vector := range vs.vectors {
		score := CosineSimilarity(query.DenseVector, vector)
		results = append(results, SearchResult{
			URI:      uri,
			Score:    score,
			Metadata: vs.metadata[uri],
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Add implements VectorStore interface.
func (vs *InMemoryVectorStore) Add(ctx context.Context, vectors []SearchResult) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	for _, v := range vectors {
		// Store would have dense vector in metadata
		if vec, ok := v.Metadata["vector"].([]float64); ok {
			vs.vectors[v.URI] = vec
		}
		vs.metadata[v.URI] = v.Metadata
	}
	return nil
}

// Delete implements VectorStore interface.
func (vs *InMemoryVectorStore) Delete(ctx context.Context, uris []string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	for _, uri := range uris {
		delete(vs.vectors, uri)
		delete(vs.metadata, uri)
	}
	return nil
}

// Close implements VectorStore interface.
func (vs *InMemoryVectorStore) Close() error {
	return nil
}

// AddVector adds a vector to the store directly.
func (vs *InMemoryVectorStore) AddVector(uri string, vector []float64, metadata map[string]interface{}) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.vectors[uri] = vector
	vs.metadata[uri] = metadata
}

// GetVector gets a vector by URI.
func (vs *InMemoryVectorStore) GetVector(uri string) ([]float64, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	vec, ok := vs.vectors[uri]
	return vec, ok
}
