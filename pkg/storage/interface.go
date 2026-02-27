// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package storage provides pluggable storage backends for OpenViking.
package storage

import (
	"context"
	"time"
)

// FilterCondition represents a filter condition for queries.
type FilterCondition struct {
	Op       string      `json:"op"` // "and", "or", "must", "range", "prefix", "contains"
	Field    string      `json:"field,omitempty"`
	Conds    interface{} `json:"conds,omitempty"`
	Prefix   string      `json:"prefix,omitempty"`
	Substr   string      `json:"substring,omitempty"`
	GTE      interface{} `json:"gte,omitempty"`
	GT       interface{} `json:"gt,omitempty"`
	LTE      interface{} `json:"lte,omitempty"`
	LT       interface{} `json:"lt,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// Filter represents filter conditions for queries.
type Filter struct {
	Op    string            `json:"op"` // "and", "or"
	Conds []FilterCondition `json:"conds"`
}

// QueryOptions contains options for query operations.
type QueryOptions struct {
	Filter       *Filter
	Limit        int
	Offset       int
	OutputFields []string
	OrderBy      string
	OrderDesc    bool
	WithVector   bool
}

// StorageInterface defines the interface for storage backends.
type StorageInterface interface {
	// Context operations
	CreateContext(ctx context.Context, context *Context) error
	GetContext(ctx context.Context, id string) (*Context, error)
	UpdateContext(ctx context.Context, context *Context) error
	DeleteContext(ctx context.Context, id string) error
	QueryContexts(ctx context.Context, opts QueryOptions) ([]Context, error)

	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id string) error
	QuerySessions(ctx context.Context, opts QueryOptions) ([]Session, error)

	// SessionMessage operations
	CreateSessionMessage(ctx context.Context, msg *SessionMessage) error
	GetSessionMessages(ctx context.Context, sessionID string) ([]SessionMessage, error)
	DeleteSessionMessages(ctx context.Context, sessionID string) error

	// Memory operations
	CreateMemory(ctx context.Context, memory *Memory) error
	GetMemory(ctx context.Context, id string) (*Memory, error)
	UpdateMemory(ctx context.Context, memory *Memory) error
	DeleteMemory(ctx context.Context, id string) error
	QueryMemories(ctx context.Context, opts QueryOptions) ([]Memory, error)

	// File operations
	CreateFile(ctx context.Context, file *File) error
	GetFile(ctx context.Context, id string) (*File, error)
	UpdateFile(ctx context.Context, file *File) error
	DeleteFile(ctx context.Context, id string) error
	QueryFiles(ctx context.Context, opts QueryOptions) ([]File, error)

	// Usage operations
	CreateUsage(ctx context.Context, usage *Usage) error
	QueryUsage(ctx context.Context, opts QueryOptions) ([]Usage, error)

	// Relation operations
	CreateRelation(ctx context.Context, relation *RelationEntry) error
	QueryRelations(ctx context.Context, uri string) ([]RelationEntry, error)
	DeleteRelation(ctx context.Context, id string) error

	// Collection management
	CreateCollection(name string, schema map[string]interface{}) error
	DropCollection(name string) error
	CollectionExists(name string) bool
	ListCollections() ([]string, error)

	// Health and lifecycle
	Ping(ctx context.Context) error
	Close() error

	// Transaction support
	Transaction(ctx context.Context, fn func(tx interface{}) error) error
}

// CollectionSchema represents a collection schema definition.
type CollectionSchema struct {
	Name        string
	Description string
	Fields      []FieldDefinition
	Indexes     []string
}

// FieldDefinition represents a field in a collection schema.
type FieldDefinition struct {
	Name     string
	Type     string
	IsPrimaryKey bool
	Dim      int // for vector fields
}

// Config holds storage configuration.
type Config struct {
	DBPath        string
	MaxOpenConns  int
	MaxIdleConns  int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns default storage configuration.
func DefaultConfig() Config {
	return Config{
		DBPath:         "openviking.db",
		MaxOpenConns:  25,
		MaxIdleConns:  5,
		ConnMaxLifetime: time.Hour,
	}
}
