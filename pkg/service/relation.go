// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrRelationExists is returned when relation already exists.
	ErrRelationExists = errors.New("relation already exists")
	// ErrRelationNotFound is returned when relation is not found.
	ErrRelationNotFound = errors.New("relation not found")
)

// Relation represents a relation between resources.
type Relation struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// RelationService provides relation management functionality.
type RelationService struct {
	relations map[string]map[string]*Relation // source -> target -> Relation
	mu        sync.RWMutex
}

// NewRelationService creates a new relation service.
func NewRelationService() *RelationService {
	return &RelationService{
		relations: make(map[string]map[string]*Relation),
	}
}

// CreateRelation creates a new relation.
func (s *RelationService) CreateRelation(ctx context.Context, source string, target string, relType string) (*Relation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if relation already exists
	if s.relations[source] != nil {
		if _, exists := s.relations[source][target]; exists {
			return nil, ErrRelationExists
		}
	}

	relation := &Relation{
		ID:        uuid.New().String(),
		Source:    source,
		Target:    target,
		Type:      relType,
		CreatedAt: time.Now().UTC(),
	}

	if s.relations[source] == nil {
		s.relations[source] = make(map[string]*Relation)
	}
	s.relations[source][target] = relation

	return relation, nil
}

// GetRelated gets all related resources.
func (s *RelationService) GetRelated(ctx context.Context, resource string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []string

	// Get resources that this resource relates to
	if relations, ok := s.relations[resource]; ok {
		for target := range relations {
			results = append(results, target)
		}
	}

	// Get resources that relate to this resource
	for source, relations := range s.relations {
		if source == resource {
			continue
		}
		for target := range relations {
			if target == resource {
				results = append(results, source)
			}
		}
	}

	return results, nil
}

// GetRelations gets all relations for a resource.
func (s *RelationService) GetRelations(ctx context.Context, resource string) ([]*Relation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Relation

	// Get outgoing relations
	if relations, ok := s.relations[resource]; ok {
		for _, rel := range relations {
			results = append(results, rel)
		}
	}

	// Get incoming relations
	for source, relations := range s.relations {
		if source == resource {
			continue
		}
		for _, rel := range relations {
			if rel.Target == resource {
				results = append(results, rel)
			}
		}
	}

	return results, nil
}

// DeleteRelation deletes a relation.
func (s *RelationService) DeleteRelation(ctx context.Context, source string, target string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.relations[source] == nil {
		return ErrRelationNotFound
	}

	if _, exists := s.relations[source][target]; !exists {
		return ErrRelationNotFound
	}

	delete(s.relations[source], target)
	return nil
}

// DeleteAllRelations deletes all relations for a resource.
func (s *RelationService) DeleteAllRelations(ctx context.Context, resource string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete outgoing relations
	delete(s.relations, resource)

	// Delete incoming relations
	for source := range s.relations {
		for target := range s.relations[source] {
			if target == resource {
				delete(s.relations[source], target)
			}
		}
	}

	return nil
}

// GetAllRelations gets all relations in the system.
func (s *RelationService) GetAllRelations(ctx context.Context) ([]*Relation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Relation
	for _, relations := range s.relations {
		for _, rel := range relations {
			results = append(results, rel)
		}
	}

	return results, nil
}
