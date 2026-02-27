// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	session := NewSession("user123")
	if session == nil {
		t.Fatal("Session is nil")
	}
	if session.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", session.UserID)
	}
	if session.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if session.State != StateActive {
		t.Errorf("Expected state 'active', got '%s'", session.State)
	}
}

func TestSessionAddMessage(t *testing.T) {
	session := NewSession("user123")
	msg := session.AddMessage(RoleUser, "Hello")

	if msg == nil {
		t.Fatal("Message is nil")
	}
	if msg.Role != RoleUser {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", msg.Content)
	}
	if session.TotalTurns != 1 {
		t.Errorf("Expected TotalTurns 1, got %d", session.TotalTurns)
	}
}

func TestSessionAddToolCall(t *testing.T) {
	session := NewSession("user123")
	msg := session.AddToolCall("search", `{"query": "test"}`)

	if msg == nil {
		t.Fatal("Message is nil")
	}
	if len(msg.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(msg.ToolCalls))
	}
	if session.SkillsUsed != 1 {
		t.Errorf("Expected SkillsUsed 1, got %d", session.SkillsUsed)
	}
}

func TestSessionPauseResume(t *testing.T) {
	session := NewSession("user123")

	err := session.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	if session.State != StatePaused {
		t.Errorf("Expected state 'paused', got '%s'", session.State)
	}

	err = session.Resume()
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if session.State != StateActive {
		t.Errorf("Expected state 'active', got '%s'", session.State)
	}
}

func TestSessionClose(t *testing.T) {
	session := NewSession("user123")

	err := session.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if session.State != StateClosed {
		t.Errorf("Expected state 'closed', got '%s'", session.State)
	}
	if session.ClosedAt == nil {
		t.Error("ClosedAt should be set")
	}
}

func TestSessionIncrementCounters(t *testing.T) {
	session := NewSession("user123")

	session.IncrementContextsUsed()
	if session.ContextsUsed != 1 {
		t.Errorf("Expected ContextsUsed 1, got %d", session.ContextsUsed)
	}

	session.IncrementMemoriesExtracted()
	if session.MemoriesExtracted != 1 {
		t.Errorf("Expected MemoriesExtracted 1, got %d", session.MemoriesExtracted)
	}

	session.IncrementCompression()
	if session.CompressionCount != 1 {
		t.Errorf("Expected CompressionCount 1, got %d", session.CompressionCount)
	}

	session.AddTokens(100)
	if session.TotalTokens != 100 {
		t.Errorf("Expected TotalTokens 100, got %d", session.TotalTokens)
	}

	session.SetSummary("Test summary")
	if session.Summary != "Test summary" {
		t.Errorf("Expected summary 'Test summary', got '%s'", session.Summary)
	}
}
