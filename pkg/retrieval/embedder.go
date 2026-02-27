// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import "context"

// EmbedResult contains embedding result with dense and/or sparse vectors.
type EmbedResult struct {
	DenseVector  []float64          `json:"dense_vector,omitempty"`
	SparseVector map[string]float64 `json:"sparse_vector,omitempty"`
}

// IsDense checks if result contains dense vector.
func (e *EmbedResult) IsDense() bool {
	return e.DenseVector != nil
}

// IsSparse checks if result contains sparse vector.
func (e *EmbedResult) IsSparse() bool {
	return e.SparseVector != nil
}

// IsHybrid checks if result is hybrid (both dense and sparse).
func (e *EmbedResult) IsHybrid() bool {
	return e.DenseVector != nil && e.SparseVector != nil
}

// Embedder defines interface for text embedding.
type Embedder interface {
	// Embed performs embedding on a single text.
	Embed(ctx context.Context, text string) (*EmbedResult, error)

	// EmbedBatch performs batch embedding on multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([]*EmbedResult, error)

	// GetDimension returns the embedding dimension.
	GetDimension() int

	// Close releases resources.
	Close() error
}

// DenseEmbedder defines interface for dense vector embedding.
type DenseEmbedder interface {
	Embedder

	// GetDenseDimension returns the dense vector dimension.
	GetDenseDimension() int
}

// SparseEmbedder defines interface for sparse vector embedding (BM25-like).
type SparseEmbedder interface {
	Embedder

	// GetIndexName returns the index name for sparse vectors.
	GetIndexName() string
}

// HybridEmbedder defines interface for hybrid embedding (dense + sparse).
type HybridEmbedder interface {
	Embedder

	// GetDenseDimension returns the dense vector dimension.
	GetDenseDimension() int

	// GetIndexName returns the index name for sparse vectors.
	GetIndexName() string
}
