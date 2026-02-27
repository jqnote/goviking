// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package session provides session management with automatic memory extraction.
package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrSessionNotFound is returned when session is not found.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionClosed is returned when session is already closed.
	ErrSessionClosed = errors.New("session closed")
	// ErrInvalidState is returned when session state is invalid.
	ErrInvalidState = errors.New("invalid session state")
)

// State represents the state of a session.
type State string

const (
	// StateActive means the session is currently active.
	StateActive State = "active"
	// StatePaused means the session is paused.
	StatePaused State = "paused"
	// StateClosed means the session is closed.
	StateClosed State = "closed"
)

// Role represents the role of a message sender.
type Role string

const (
	// RoleUser represents the user.
	RoleUser Role = "user"
	// RoleAssistant represents the assistant.
	RoleAssistant Role = "assistant"
	// RoleSystem represents the system.
	RoleSystem Role = "system"
	// RoleTool represents a tool.
	RoleTool Role = "tool"
)

// Message represents a message in a session.
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Name      string    `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
}

// FunctionCall represents a function call.
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments string                 `json:"arguments"`
}

// Session represents a session for agent interactions.
type Session struct {
	ID                string         `json:"id"`
	SessionID         string         `json:"session_id"`
	UserID            string         `json:"user_id"`
	State             State          `json:"state"`
	TotalTurns        int64          `json:"total_turns"`
	TotalTokens       int64          `json:"total_tokens"`
	CompressionCount  int64          `json:"compression_count"`
	ContextsUsed      int64          `json:"contexts_used"`
	SkillsUsed        int64          `json:"skills_used"`
	MemoriesExtracted int64          `json:"memories_extracted"`
	Summary           string         `json:"summary"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ClosedAt          *time.Time     `json:"closed_at,omitempty"`
}

// NewSession creates a new session.
func NewSession(userID string) *Session {
	now := time.Now().UTC()
	return &Session{
		ID:                uuid.New().String(),
		SessionID:         uuid.New().String(),
		UserID:            userID,
		State:             StateActive,
		TotalTurns:        0,
		TotalTokens:       0,
		CompressionCount:  0,
		ContextsUsed:      0,
		SkillsUsed:        0,
		MemoriesExtracted: 0,
		Summary:           "",
		Metadata:          make(map[string]any),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// AddMessage adds a message to the session.
func (s *Session) AddMessage(role Role, content string) *Message {
	msg := &Message{
		ID:        uuid.New().String(),
		SessionID: s.SessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	s.TotalTurns++
	s.UpdatedAt = time.Now().UTC()
	return msg
}

// AddToolCall adds a tool call to the session.
func (s *Session) AddToolCall(name, arguments string) *Message {
	msg := &Message{
		ID:        uuid.New().String(),
		SessionID: s.SessionID,
		Role:      RoleTool,
		Content:   "",
		ToolCalls: []ToolCall{
			{
				ID:   uuid.New().String(),
				Type: "function",
				Function: FunctionCall{
					Name:      name,
					Arguments: arguments,
				},
			},
		},
		CreatedAt: time.Now().UTC(),
	}
	s.SkillsUsed++
	s.UpdatedAt = time.Now().UTC()
	return msg
}

// Pause pauses the session.
func (s *Session) Pause() error {
	if s.State != StateActive {
		return ErrInvalidState
	}
	s.State = StatePaused
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Resume resumes the session.
func (s *Session) Resume() error {
	if s.State != StatePaused {
		return ErrInvalidState
	}
	s.State = StateActive
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Close closes the session.
func (s *Session) Close() error {
	if s.State == StateClosed {
		return ErrSessionClosed
	}
	s.State = StateClosed
	now := time.Now().UTC()
	s.ClosedAt = &now
	s.UpdatedAt = now
	return nil
}

// IncrementContextsUsed increments the contexts used counter.
func (s *Session) IncrementContextsUsed() {
	s.ContextsUsed++
	s.UpdatedAt = time.Now().UTC()
}

// IncrementMemoriesExtracted increments the memories extracted counter.
func (s *Session) IncrementMemoriesExtracted() {
	s.MemoriesExtracted++
	s.UpdatedAt = time.Now().UTC()
}

// IncrementCompression increments the compression counter.
func (s *Session) IncrementCompression() {
	s.CompressionCount++
	s.UpdatedAt = time.Now().UTC()
}

// AddTokens adds tokens to the session.
func (s *Session) AddTokens(tokens int64) {
	s.TotalTokens += tokens
	s.UpdatedAt = time.Now().UTC()
}

// SetSummary sets the session summary.
func (s *Session) SetSummary(summary string) {
	s.Summary = summary
	s.UpdatedAt = time.Now().UTC()
}

// Manager handles session lifecycle.
type Manager interface {
	// Create creates a new session.
	Create(ctx context.Context, userID string) (*Session, error)
	// Get retrieves a session by ID.
	Get(ctx context.Context, sessionID string) (*Session, error)
	// Update updates a session.
	Update(ctx context.Context, session *Session) error
	// Delete deletes a session.
	Delete(ctx context.Context, sessionID string) error
	// List lists sessions for a user.
	List(ctx context.Context, userID string) ([]*Session, error)
	// AddMessage adds a message to a session.
	AddMessage(ctx context.Context, sessionID string, msg *Message) error
	// GetMessages retrieves messages for a session.
	GetMessages(ctx context.Context, sessionID string) ([]*Message, error)
}
