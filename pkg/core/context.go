// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package core provides core context database functionality with tiered loading (L0/L1/L2).
package core

import (
	"time"

	"github.com/google/uuid"
)

// ContextType represents the type of context.
type ContextType string

const (
	ContextTypeSkill   ContextType = "skill"
	ContextTypeMemory  ContextType = "memory"
	ContextTypeResource ContextType = "resource"
)

// Category represents the category of context.
type Category string

const (
	CategoryPatterns    Category = "patterns"
	CategoryCases       Category = "cases"
	CategoryProfile     Category = "profile"
	CategoryPreferences Category = "preferences"
	CategoryEntities    Category = "entities"
	CategoryEvents      Category = "events"
)

// ResourceContentType represents the content type of a resource.
type ResourceContentType string

const (
	ResourceContentTypeText   ResourceContentType = "text"
	ResourceContentTypeImage  ResourceContentType = "image"
	ResourceContentTypeVideo  ResourceContentType = "video"
	ResourceContentTypeAudio  ResourceContentType = "audio"
	ResourceContentTypeBinary ResourceContentType = "binary"
)

// ContextTier represents the tier level of context (L0, L1, L2).
type ContextTier int

const (
	TierL0 ContextTier = iota // Essential context, always loaded
	TierL1                    // Context loaded on demand
	TierL2                    // Archive context, loaded when needed
)

// Vectorize holds vectorization data for a context.
type Vectorize struct {
	Text string `json:"text"`
}

// Context represents a unified context entry for all context types.
type Context struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	ParentURI    string            `json:"parent_uri,omitempty"`
	IsLeaf       bool              `json:"is_leaf"`
	Abstract     string            `json:"abstract"`
	ContextType  ContextType       `json:"context_type"`
	Category     Category          `json:"category,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ActiveCount  int64             `json:"active_count"`
	RelatedURI   []string          `json:"related_uri,omitempty"`
	Meta         map[string]any    `json:"meta,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	Vector       []float64         `json:"vector,omitempty"`
	Vectorize    Vectorize         `json:"vectorize"`
	Tier         ContextTier       `json:"tier"`
}

// NewContext creates a new Context with default values.
func NewContext(uri string) *Context {
	now := time.Now().UTC()
	return &Context{
		ID:          uuid.New().String(),
		URI:         uri,
		IsLeaf:      false,
		ContextType: deriveContextType(uri),
		Category:    deriveCategory(uri),
		CreatedAt:   now,
		UpdatedAt:   now,
		ActiveCount: 0,
		RelatedURI:  []string{},
		Meta:        make(map[string]any),
		Vectorize:   Vectorize{},
		Tier:        TierL1, // Default to L1, can be changed
	}
}

// deriveContextType derives the context type from URI prefix.
func deriveContextType(uri string) ContextType {
	if hasPrefix(uri, "viking://agent/skills") {
		return ContextTypeSkill
	}
	if contains(uri, "memories") {
		return ContextTypeMemory
	}
	return ContextTypeResource
}

// deriveCategory derives the category from URI prefix.
func deriveCategory(uri string) Category {
	if hasPrefix(uri, "viking://agent/memories") {
		if contains(uri, "patterns") {
			return CategoryPatterns
		}
		if contains(uri, "cases") {
			return CategoryCases
		}
	}
	if hasPrefix(uri, "viking://user/memories") {
		if contains(uri, "profile") {
			return CategoryProfile
		}
		if contains(uri, "preferences") {
			return CategoryPreferences
		}
		if contains(uri, "entities") {
			return CategoryEntities
		}
		if contains(uri, "events") {
			return CategoryEvents
		}
	}
	return ""
}

// UpdateActivity updates activity statistics.
func (c *Context) UpdateActivity() {
	c.ActiveCount++
	c.UpdatedAt = time.Now().UTC()
}

// GetVectorizationText returns text for vectorization.
func (c *Context) GetVectorizationText() string {
	return c.Vectorize.Text
}

// ToMap converts context to map for storage.
func (c *Context) ToMap() map[string]any {
	result := map[string]any{
		"id":           c.ID,
		"uri":          c.URI,
		"parent_uri":   c.ParentURI,
		"is_leaf":      c.IsLeaf,
		"abstract":     c.Abstract,
		"context_type": string(c.ContextType),
		"category":     string(c.Category),
		"created_at":   c.CreatedAt.Format(time.RFC3339),
		"updated_at":   c.UpdatedAt.Format(time.RFC3339),
		"active_count": c.ActiveCount,
		"vector":       c.Vector,
		"meta":         c.Meta,
		"related_uri":  c.RelatedURI,
		"session_id":   c.SessionID,
		"tier":         int(c.Tier),
	}

	if c.UserID != "" {
		result["user_id"] = c.UserID
	}

	// Add skill-specific fields from meta
	if c.ContextType == ContextTypeSkill {
		result["name"] = c.Meta["name"]
		result["description"] = c.Meta["description"]
	}

	return result
}

// FromMap creates a Context from a map.
func FromMap(data map[string]any) *Context {
	c := &Context{
		ID:          getString(data, "id", ""),
		URI:         getString(data, "uri", ""),
		ParentURI:   getString(data, "parent_uri", ""),
		IsLeaf:      getBool(data, "is_leaf", false),
		Abstract:    getString(data, "abstract", ""),
		ContextType: ContextType(getString(data, "context_type", "")),
		Category:    Category(getString(data, "category", "")),
		ActiveCount: getInt64(data, "active_count", 0),
		RelatedURI:  getStringSlice(data, "related_uri"),
		SessionID:   getString(data, "session_id", ""),
		UserID:      getString(data, "user_id", ""),
		Tier:        ContextTier(getInt(data, "tier", int(TierL1))),
		Meta:        getMap(data, "meta"),
		Vector:      getFloat64Slice(data, "vector"),
		Vectorize:   Vectorize{Text: getString(data, "vectorize_text", "")},
	}

	// Parse timestamps
	if createdAt := getString(data, "created_at", ""); createdAt != "" {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			c.CreatedAt = t
		}
	}
	if updatedAt := getString(data, "updated_at", ""); updatedAt != "" {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			c.UpdatedAt = t
		}
	}

	// Set defaults if not set
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return c
}

// Helper functions
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getString(m map[string]any, key, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBool(m map[string]any, key string, def bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getInt(m map[string]any, key string, def int) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return def
}

func getInt64(m map[string]any, key string, def int64) int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return int64(n)
		case int64:
			return n
		case float64:
			return int64(n)
		}
	}
	return def
}

func getStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]any); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if s, ok := item.(string); ok {
					result[i] = s
				}
			}
			return result
		}
	}
	return nil
}

func getFloat64Slice(m map[string]any, key string) []float64 {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]any); ok {
			result := make([]float64, len(slice))
			for i, item := range slice {
				switch n := item.(type) {
				case float64:
					result[i] = n
				case int:
					result[i] = float64(n)
				}
			}
			return result
		}
	}
	return nil
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return make(map[string]any)
}
