// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package service provides business logic layer for GoViking.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidContext is returned when context is invalid.
	ErrInvalidContext = errors.New("invalid context")
	// ErrInvalidSession is returned when session is invalid.
	ErrInvalidSession = errors.New("invalid session")
	// ErrNotFound is returned when resource is not found.
	ErrNotFound = errors.New("not found")
)

// ContextService provides context business logic.
type ContextService struct {
	// Storage would be injected here
}

// NewContextService creates a new context service.
func NewContextService() *ContextService {
	return &ContextService{}
}

// CreateContextRequest represents a create context request.
type CreateContextRequest struct {
	URI      string
	Type     string
	Name     string
	Content  string
	Metadata map[string]any
}

// Validate validates the create context request.
func (r *CreateContextRequest) Validate() error {
	if r.URI == "" {
		return errors.New("uri is required")
	}
	if r.Type == "" {
		return errors.New("type is required")
	}
	return nil
}

// Context represents a context.
type Context struct {
	ID        string         `json:"id"`
	URI       string         `json:"uri"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// Create creates a new context.
func (s *ContextService) Create(ctx context.Context, req *CreateContextRequest) (*Context, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Context{
		ID:        uuid.New().String(),
		URI:       req.URI,
		Type:      req.Type,
		Name:      req.Name,
		Content:   req.Content,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SessionService provides session business logic.
type SessionService struct {
	// Storage would be injected here
}

// NewSessionService creates a new session service.
func NewSessionService() *SessionService {
	return &SessionService{}
}

// CreateSessionRequest represents a create session request.
type CreateSessionRequest struct {
	UserID   string
	Metadata map[string]any
}

// Validate validates the create session request.
func (r *CreateSessionRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// Session represents a session.
type Session struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	UserID    string         `json:"user_id"`
	State     string         `json:"state"`
	Summary   string         `json:"summary,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Create creates a new session.
func (s *SessionService) Create(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Session{
		ID:        uuid.New().String(),
		SessionID: uuid.New().String(),
		UserID:    req.UserID,
		State:     "active",
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Resume resumes a session.
func (s *SessionService) Resume(ctx context.Context, id string) (*Session, error) {
	now := time.Now().UTC()
	return &Session{
		ID:        id,
		SessionID: id,
		UserID:    "user",
		State:     "active",
		UpdatedAt: now,
	}, nil
}

// Close closes a session.
func (s *SessionService) Close(ctx context.Context, id string) error {
	return nil
}
