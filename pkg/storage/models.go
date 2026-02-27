// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package storage provides pluggable storage backends for OpenViking.
package storage

import (
	"time"
)

// ContextType represents the type of context.
type ContextType string

const (
	ContextTypeFile     ContextType = "file"
	ContextTypeDirectory ContextType = "directory"
	ContextTypeSession   ContextType = "session"
	ContextTypeMemory    ContextType = "memory"
	ContextTypeSkill     ContextType = "skill"
)

// Context represents a context entry in the database.
type Context struct {
	ID          string      `json:"id" db:"id"`
	URI         string      `json:"uri" db:"uri"`
	Type        ContextType `json:"type" db:"type"`
	ContextType string      `json:"context_type" db:"context_type"`
	ParentURI   string      `json:"parent_uri" db:"parent_uri"`
	IsLeaf      bool        `json:"is_leaf" db:"is_leaf"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Tags        string      `json:"tags" db:"tags"`
	Abstract    string      `json:"abstract" db:"abstract"`
	ActiveCount int64       `json:"active_count" db:"active_count"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// Session represents a session in the database.
type Session struct {
	ID             string    `json:"id" db:"id"`
	SessionID      string    `json:"session_id" db:"session_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	TotalTurns     int64     `json:"total_turns" db:"total_turns"`
	TotalTokens    int64     `json:"total_tokens" db:"total_tokens"`
	CompressionCount int64   `json:"compression_count" db:"compression_count"`
	ContextsUsed   int64     `json:"contexts_used" db:"contexts_used"`
	SkillsUsed     int64     `json:"skills_used" db:"skills_used"`
	MemoriesExtracted int64  `json:"memories_extracted" db:"memories_extracted"`
	Summary        string    `json:"summary" db:"summary"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// SessionMessage represents a message in a session.
type SessionMessage struct {
	ID         string    `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	Role       string    `json:"role" db:"role"`
	Content    string    `json:"content" db:"content"`
	OrderIndex int64     `json:"order_index" db:"order_index"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Memory represents an extracted memory from sessions.
type Memory struct {
	ID          string    `json:"id" db:"id"`
	SessionID   string    `json:"session_id" db:"session_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Content     string    `json:"content" db:"content"`
	Importance  float64   `json:"importance" db:"importance"`
	Tags        string    `json:"tags" db:"tags"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// File represents a file metadata entry.
type File struct {
	ID          string    `json:"id" db:"id"`
	URI         string    `json:"uri" db:"uri"`
	Name        string    `json:"name" db:"name"`
	Size        int64     `json:"size" db:"size"`
	ContentType string    `json:"content_type" db:"content_type"`
	Checksum    string    `json:"checksum" db:"checksum"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Usage represents a usage record for contexts/skills.
type Usage struct {
	ID          string    `json:"id" db:"id"`
	SessionID   string    `json:"session_id" db:"session_id"`
	URI         string    `json:"uri" db:"uri"`
	Type        string    `json:"type" db:"type"` // "context" or "skill"
	Contribution float64 `json:"contribution" db:"contribution"`
	Input       string    `json:"input" db:"input"`
	Output      string    `json:"output" db:"output"`
	Success     bool      `json:"success" db:"success"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
}

// RelationEntry represents a relation between URIs.
type RelationEntry struct {
	ID        string    `json:"id" db:"id"`
	URIs      string    `json:"uris" db:"uris"` // JSON array of URIs
	Reason    string    `json:"reason" db:"reason"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
