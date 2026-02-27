// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"testing"
)

func TestContextServiceCreate(t *testing.T) {
	svc := NewContextService()
	req := &CreateContextRequest{
		URI:      "viking://memory/test",
		Type:     "memory",
		Name:     "Test Context",
		Content:  "Test content",
		Metadata: map[string]any{"key": "value"},
	}

	ctx := context.Background()
	result, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}
	if result.URI != req.URI {
		t.Errorf("Expected URI '%s', got '%s'", req.URI, result.URI)
	}
	if result.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestContextServiceValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateContextRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &CreateContextRequest{URI: "test", Type: "memory"},
			wantErr: false,
		},
		{
			name:    "missing URI",
			req:     &CreateContextRequest{Type: "memory"},
			wantErr: true,
		},
		{
			name:    "missing Type",
			req:     &CreateContextRequest{URI: "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionServiceCreate(t *testing.T) {
	svc := NewSessionService()
	req := &CreateSessionRequest{
		UserID:   "user123",
		Metadata: map[string]any{"source": "web"},
	}

	ctx := context.Background()
	result, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}
	if result.UserID != req.UserID {
		t.Errorf("Expected UserID '%s', got '%s'", req.UserID, result.UserID)
	}
	if result.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if result.State != "active" {
		t.Errorf("Expected state 'active', got '%s'", result.State)
	}
}

func TestSessionServiceValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateSessionRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &CreateSessionRequest{UserID: "user123"},
			wantErr: false,
		},
		{
			name:    "missing UserID",
			req:     &CreateSessionRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionServiceResume(t *testing.T) {
	svc := NewSessionService()
	ctx := context.Background()

	result, err := svc.Resume(ctx, "session123")
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}
	if result.ID != "session123" {
		t.Errorf("Expected ID 'session123', got '%s'", result.ID)
	}
}
