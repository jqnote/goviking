// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

//go:build sqlite3
// +build sqlite3

package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSQLiteStorage_ContextCRUD(t *testing.T) {
	// Create temp file for test database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create storage
	storage, err := NewSQLiteStorage(Config{
		DBPath:         tmpFile.Name(),
		MaxOpenConns:   5,
		MaxIdleConns:   2,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test CreateContext
	testContext := &Context{
		ID:          uuid.New().String(),
		URI:         "viking://test/context1",
		Type:        ContextTypeFile,
		ContextType: "document",
		ParentURI:   "viking://test",
		IsLeaf:      true,
		Name:        "test context",
		Description: "A test context",
		Tags:        "tag1,tag2",
		Abstract:    "Test abstract",
		ActiveCount: 0,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := storage.CreateContext(ctx, testContext); err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Test GetContext
	retrieved, err := storage.GetContext(ctx, testContext.ID)
	if err != nil {
		t.Fatalf("failed to get context: %v", err)
	}
	if retrieved == nil {
		t.Fatal("context is nil")
	}
	if retrieved.URI != testContext.URI {
		t.Errorf("expected URI %s, got %s", testContext.URI, retrieved.URI)
	}

	// Test UpdateContext
	testContext.Description = "Updated description"
	if err := storage.UpdateContext(ctx, testContext); err != nil {
		t.Fatalf("failed to update context: %v", err)
	}

	updated, err := storage.GetContext(ctx, testContext.ID)
	if err != nil {
		t.Fatalf("failed to get updated context: %v", err)
	}
	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got '%s'", updated.Description)
	}

	// Test DeleteContext
	if err := storage.DeleteContext(ctx, testContext.ID); err != nil {
		t.Fatalf("failed to delete context: %v", err)
	}

	deleted, err := storage.GetContext(ctx, testContext.ID)
	if err != nil {
		t.Fatalf("failed to get deleted context: %v", err)
	}
	if deleted != nil {
		t.Error("expected nil context after delete")
	}
}

func TestSQLiteStorage_SessionCRUD(t *testing.T) {
	// Create temp file for test database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create storage
	storage, err := NewSQLiteStorage(Config{
		DBPath:         tmpFile.Name(),
		MaxOpenConns:   5,
		MaxIdleConns:   2,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test CreateSession
	testSession := &Session{
		ID:               uuid.New().String(),
		SessionID:        uuid.New().String(),
		UserID:           "user123",
		TotalTurns:       10,
		TotalTokens:      1000,
		CompressionCount: 2,
		ContextsUsed:     5,
		SkillsUsed:       3,
		MemoriesExtracted: 4,
		Summary:          "Test session summary",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := storage.CreateSession(ctx, testSession); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test GetSession
	retrieved, err := storage.GetSession(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if retrieved == nil {
		t.Fatal("session is nil")
	}
	if retrieved.SessionID != testSession.SessionID {
		t.Errorf("expected SessionID %s, got %s", testSession.SessionID, retrieved.SessionID)
	}

	// Test QuerySessions
	sessions, err := storage.QuerySessions(ctx, QueryOptions{
		Filter: &Filter{
			Op: "and",
			Conds: []FilterCondition{
				{Op: "must", Field: "user_id", Value: "user123"},
			},
		},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
}

func TestSQLiteStorage_QueryContexts(t *testing.T) {
	// Create temp file for test database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create storage
	storage, err := NewSQLiteStorage(Config{
		DBPath:         tmpFile.Name(),
		MaxOpenConns:   5,
		MaxIdleConns:   2,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Create multiple contexts
	for i := 0; i < 5; i++ {
		c := &Context{
			ID:          uuid.New().String(),
			URI:         "viking://test/context" + string(rune('a'+i)),
			Type:        ContextTypeFile,
			ContextType: "document",
			ParentURI:   "viking://test",
			IsLeaf:      true,
			Name:        "test context " + string(rune('a' + i)),
			Tags:        "tag1",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		if err := storage.CreateContext(ctx, c); err != nil {
			t.Fatalf("failed to create context %d: %v", i, err)
		}
	}

	// Test QueryContexts with filter
	contexts, err := storage.QueryContexts(ctx, QueryOptions{
		Filter: &Filter{
			Op: "and",
			Conds: []FilterCondition{
				{Op: "must", Field: "parent_uri", Value: "viking://test"},
				{Op: "must", Field: "type", Value: "file"},
			},
		},
		OrderBy:  "name",
		OrderDesc: false,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("failed to query contexts: %v", err)
	}
	if len(contexts) != 5 {
		t.Errorf("expected 5 contexts, got %d", len(contexts))
	}
}
