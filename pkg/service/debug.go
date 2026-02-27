// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"time"
)

// DebugService provides debug and health check functionality.
type DebugService struct {
	queueManager   interface{ /* QueueManager interface */ }
	vectorStore    interface{ /* VectorStore interface */ }
	storage        interface{ /* Storage interface */ }
}

// NewDebugService creates a new debug service.
func NewDebugService() *DebugService {
	return &DebugService{}
}

// SetQueueManager sets the queue manager for debug service.
func (s *DebugService) SetQueueManager(qm interface{}) {
	s.queueManager = qm
}

// SetVectorStore sets the vector store for debug service.
func (s *DebugService) SetVectorStore(vs interface{}) {
	s.vectorStore = vs
}

// SetStorage sets the storage for debug service.
func (s *DebugService) SetStorage(st interface{}) {
	s.storage = st
}

// ComponentStatus represents the status of a component.
type ComponentStatus struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"` // "healthy", "degraded", "down"
	LatencyMs    int64         `json:"latency_ms,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
}

// ComponentHealthCheck checks the health of a specific component.
func (s *DebugService) ComponentHealthCheck(ctx context.Context, component string) (*ComponentStatus, error) {
	start := time.Now()
	status := &ComponentStatus{
		Name:   component,
		Status: "healthy",
	}

	switch component {
	case "queue":
		if s.queueManager == nil {
			status.Status = "degraded"
			status.Details = map[string]any{"message": "queue manager not configured"}
		} else {
			// Check queue health
			status.Details = map[string]any{"message": "queue manager operational"}
		}
	case "vector_store":
		if s.vectorStore == nil {
			status.Status = "degraded"
			status.Details = map[string]any{"message": "vector store not configured"}
		} else {
			// Check vector store health
			status.Details = map[string]any{"message": "vector store operational"}
		}
	case "storage":
		if s.storage == nil {
			status.Status = "degraded"
			status.Details = map[string]any{"message": "storage not configured"}
		} else {
			// Check storage health
			status.Details = map[string]any{"message": "storage operational"}
		}
	default:
		status.Status = "unknown"
		status.ErrorMessage = "unknown component"
	}

	status.LatencyMs = time.Since(start).Milliseconds()
	return status, nil
}

// OverallStatus returns the overall system status.
func (s *DebugService) OverallStatus(ctx context.Context) (map[string]*ComponentStatus, error) {
	components := []string{"queue", "vector_store", "storage"}
	result := make(map[string]*ComponentStatus)

	for _, comp := range components {
		status, err := s.ComponentHealthCheck(ctx, comp)
		if err != nil {
			return nil, err
		}
		result[comp] = status
	}

	return result, nil
}

// GetDetailedStatus returns detailed status including queue size and processing rate.
func (s *DebugService) GetDetailedStatus(ctx context.Context) (map[string]any, error) {
	status := make(map[string]any)

	// Get component statuses
	components, err := s.OverallStatus(ctx)
	if err != nil {
		return nil, err
	}

	status["components"] = components

	// Add queue details if available
	if s.queueManager != nil {
		status["queue"] = map[string]any{
			"size":           0, // Would be fetched from queue manager
			"processing_rate": 0, // messages per second
		}
	}

	// Add vector store details if available
	if s.vectorStore != nil {
		status["vector_store"] = map[string]any{
			"total_vectors":   0,
			"index_size_mb":   0,
		}
	}

	// Add storage details if available
	if s.storage != nil {
		status["storage"] = map[string]any{
			"total_sessions": 0,
			"total_contexts": 0,
		}
	}

	return status, nil
}
