// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RelationManager handles relation management between directories.
type RelationManager struct {
	agfs *AGFS
}

// NewRelationManager creates a new RelationManager.
func NewRelationManager(agfs *AGFS) *RelationManager {
	return &RelationManager{agfs: agfs}
}

// Link creates a relation from one directory to others.
func (r *RelationManager) Link(fromURI string, uris []string, reason string) error {
	r.agfs.mu.Lock()
	defer r.agfs.mu.Unlock()

	fromURI = r.agfs.normalizeURI(fromURI)
	path := r.agfs.URIToPath(fromURI)
	if path == "" {
		return ErrInvalidURI
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ErrNotADirectory
	}

	// Read existing relations
	relations, err := r.readRelationTable(path)
	if err != nil {
		relations = make([]RelationEntry, 0)
	}

	// Generate new ID
	id := generateLinkID(relations)

	// Add new relation
	entry := RelationEntry{
		ID:        id,
		URIs:      uris,
		Reason:    reason,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	relations = append(relations, entry)

	// Write relations
	return r.writeRelationTable(path, relations)
}

// Unlink removes a relation from a directory.
func (r *RelationManager) Unlink(fromURI, targetURI string) error {
	r.agfs.mu.Lock()
	defer r.agfs.mu.Unlock()

	fromURI = r.agfs.normalizeURI(fromURI)
	path := r.agfs.URIToPath(fromURI)
	if path == "" {
		return ErrInvalidURI
	}

	// Read existing relations
	relations, err := r.readRelationTable(path)
	if err != nil {
		return err
	}

	// Find and remove target URI
	found := false
	for i, entry := range relations {
		for j, uri := range entry.URIs {
			if uri == targetURI {
				// Remove URI from entry
				entry.URIs = append(entry.URIs[:j], entry.URIs[j+1:]...)
				relations[i] = entry
				found = true
				break
			}
		}
	}

	if !found {
		return ErrNotFound
	}

	// Remove entries with no URIs
	newRelations := make([]RelationEntry, 0)
	for _, entry := range relations {
		if len(entry.URIs) > 0 {
			newRelations = append(newRelations, entry)
		}
	}

	// Write relations
	return r.writeRelationTable(path, newRelations)
}

// GetRelations returns all relations from a directory.
func (r *RelationManager) GetRelations(uri string) ([]RelationEntry, error) {
	r.agfs.mu.RLock()
	defer r.agfs.mu.RUnlock()

	uri = r.agfs.normalizeURI(uri)
	path := r.agfs.URIToPath(uri)
	if path == "" {
		return nil, ErrInvalidURI
	}

	return r.readRelationTable(path)
}

// GetRelatedURIs returns all URIs related to a directory.
func (r *RelationManager) GetRelatedURIs(uri string) ([]string, error) {
	relations, err := r.GetRelations(uri)
	if err != nil {
		return nil, err
	}

	var uris []string
	for _, entry := range relations {
		uris = append(uris, entry.URIs...)
	}

	return uris, nil
}

// readRelationTable reads the relation table from a directory.
func (r *RelationManager) readRelationTable(dirPath string) ([]RelationEntry, error) {
	relPath := filepath.Join(dirPath, ".relations.json")
	data, err := os.ReadFile(relPath)
	if os.IsNotExist(err) {
		return make([]RelationEntry, 0), nil
	}
	if err != nil {
		return nil, err
	}

	var relations []RelationEntry
	if err := json.Unmarshal(data, &relations); err != nil {
		return nil, err
	}

	return relations, nil
}

// writeRelationTable writes the relation table to a directory.
func (r *RelationManager) writeRelationTable(dirPath string, relations []RelationEntry) error {
	relPath := filepath.Join(dirPath, ".relations.json")
	data, err := json.MarshalIndent(relations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(relPath, data, 0644)
}

// generateLinkID generates a unique ID for a new link.
func generateLinkID(relations []RelationEntry) string {
	// Find the highest existing ID
	maxID := 0
	for _, rel := range relations {
		if len(rel.ID) > 5 && rel.ID[:5] == "link_" {
			var num int
			if _, err := fmt.Sscanf(rel.ID, "link_%d", &num); err == nil {
				if num > maxID {
					maxID = num
				}
			}
		}
	}
	return fmt.Sprintf("link_%d", maxID+1)
}
