// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package retrieval provides context retrieval with semantic search, directory traversal, and trajectory tracking.
package retrieval

// Core types
// - ContextType: Type of context (memory, resource, skill)
// - TypedQuery: Query targeting a specific context type
// - QueryPlan: Contains multiple TypedQueries
// - MatchedContext: Matched context from retrieval
// - QueryResult: Result for a single TypedQuery
// - FindResult: Final result from search
// - SearchOptions: Options for retrieval

// Trajectory
// - Trajectory: Tracks retrieval path
// - TrajectoryLogger: Logs multiple retrieval trajectories

// Search
// - SemanticSearch: Vector-based semantic search
// - HybridSearch: Combines keyword and semantic search
// - KeywordSearch: BM25-based keyword search
// - VectorStore: Interface for vector storage

// Directory
// - DirectoryTraverser: Recursive directory traversal
// - DirectoryEntry: File or directory entry

// Pattern
// - PatternMatcher: Matches paths against patterns
// - PathRetriever: Retrieves files based on path patterns

// Embedder
// - Embedder: Interface for text embedding
// - EmbedResult: Embedding result with dense/sparse vectors
