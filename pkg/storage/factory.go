// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"
)

// BackendType represents the type of storage backend.
type BackendType string

const (
	BackendSQLite BackendType = "sqlite"
	// BackendPostgres BackendType = "postgres"
	// BackendMemory   BackendType = "memory"
)

// NewStorage creates a new storage instance based on the backend type.
func NewStorage(backend BackendType, cfg Config) (StorageInterface, error) {
	switch backend {
	case BackendSQLite:
		return NewSQLiteStorage(cfg)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", backend)
	}
}

// InitStorage initializes the storage with default configuration.
func InitStorage(dbPath string) (StorageInterface, error) {
	cfg := DefaultConfig()
	cfg.DBPath = dbPath
	return NewStorage(BackendSQLite, cfg)
}

// StorageFromContext retrieves storage from context or creates a new one.
func StorageFromContext(ctx context.Context, key string) (StorageInterface, bool) {
	val := ctx.Value(key)
	if val == nil {
		return nil, false
	}
	storage, ok := val.(StorageInterface)
	return storage, ok
}

// WithStorage adds storage to context.
func WithStorage(ctx context.Context, key string, storage StorageInterface) context.Context {
	return context.WithValue(ctx, key, storage)
}
