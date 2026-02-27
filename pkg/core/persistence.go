// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PersistenceConfig holds configuration for persistence.
type PersistenceConfig struct {
	StoragePath    string
	AutoSave       bool
	AutoSaveInterval time.Duration
}

// DefaultPersistenceConfig returns a default configuration.
func DefaultPersistenceConfig() *PersistenceConfig {
	return &PersistenceConfig{
		StoragePath:    "./data",
		AutoSave:      false,
		AutoSaveInterval: time.Minute * 5,
	}
}

// PersistenceHandler handles context persistence and restoration.
type PersistenceHandler struct {
	config   *PersistenceConfig
	tc       *TieredContext
	sessionID string
}

// NewPersistenceHandler creates a new PersistenceHandler.
func NewPersistenceHandler(config *PersistenceConfig, tc *TieredContext, sessionID string) *PersistenceHandler {
	if config == nil {
		config = DefaultPersistenceConfig()
	}
	return &PersistenceHandler{
		config:   config,
		tc:       tc,
		sessionID: sessionID,
	}
}

// Save persists context to storage.
func (p *PersistenceHandler) Save() error {
	if p.config.StoragePath == "" {
		return fmt.Errorf("storage path not configured")
	}

	// Ensure directory exists
	if err := os.MkdirAll(p.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	filename := p.getFilename()
	data := p.marshalContext()

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// Load restores context from storage.
func (p *PersistenceHandler) Load() error {
	if p.config.StoragePath == "" {
		return fmt.Errorf("storage path not configured")
	}

	filename := p.getFilename()
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file exists yet, not an error
		}
		return fmt.Errorf("failed to read context file: %w", err)
	}

	contexts, err := p.unmarshalContext(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal context: %w", err)
	}

	// Add contexts to tiered context
	for _, ctx := range contexts {
		p.tc.Add(ctx)
	}

	return nil
}

// Delete removes persisted context from storage.
func (p *PersistenceHandler) Delete() error {
	if p.config.StoragePath == "" {
		return fmt.Errorf("storage path not configured")
	}

	filename := p.getFilename()
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to delete
		}
		return fmt.Errorf("failed to delete context file: %w", err)
	}

	return nil
}

// Exists checks if persisted context exists.
func (p *PersistenceHandler) Exists() bool {
	if p.config.StoragePath == "" {
		return false
	}
	filename := p.getFilename()
	_, err := os.Stat(filename)
	return err == nil
}

// GetLastModified returns the last modification time of persisted context.
func (p *PersistenceHandler) GetLastModified() (time.Time, error) {
	if p.config.StoragePath == "" {
		return time.Time{}, fmt.Errorf("storage path not configured")
	}

	filename := p.getFilename()
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

func (p *PersistenceHandler) getFilename() string {
	return filepath.Join(p.config.StoragePath, fmt.Sprintf("context_%s.json", p.sessionID))
}

func (p *PersistenceHandler) marshalContext() []byte {
	contexts := p.tc.GetAll()

	type serializedContext struct {
		ID           string            `json:"id"`
		URI          string            `json:"uri"`
		ParentURI    string            `json:"parent_uri,omitempty"`
		IsLeaf       bool              `json:"is_leaf"`
		Abstract     string            `json:"abstract"`
		ContextType  string            `json:"context_type"`
		Category     string            `json:"category,omitempty"`
		CreatedAt    string            `json:"created_at"`
		UpdatedAt    string            `json:"updated_at"`
		ActiveCount  int64             `json:"active_count"`
		RelatedURI   []string          `json:"related_uri,omitempty"`
		Meta         map[string]any   `json:"meta,omitempty"`
		SessionID    string            `json:"session_id,omitempty"`
		UserID       string            `json:"user_id,omitempty"`
		Vector       []float64         `json:"vector,omitempty"`
		Vectorize    Vectorize         `json:"vectorize"`
		Tier         int               `json:"tier"`
	}

	serialized := make([]serializedContext, len(contexts))
	for i, ctx := range contexts {
		serialized[i] = serializedContext{
			ID:          ctx.ID,
			URI:         ctx.URI,
			ParentURI:   ctx.ParentURI,
			IsLeaf:      ctx.IsLeaf,
			Abstract:    ctx.Abstract,
			ContextType: string(ctx.ContextType),
			Category:    string(ctx.Category),
			CreatedAt:   ctx.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   ctx.UpdatedAt.Format(time.RFC3339),
			ActiveCount: ctx.ActiveCount,
			RelatedURI:  ctx.RelatedURI,
			Meta:        ctx.Meta,
			SessionID:   ctx.SessionID,
			UserID:      ctx.UserID,
			Vector:      ctx.Vector,
			Vectorize:   ctx.Vectorize,
			Tier:        int(ctx.Tier),
		}
	}

	data, _ := json.MarshalIndent(serialized, "", "  ")
	return data
}

func (p *PersistenceHandler) unmarshalContext(data []byte) ([]*Context, error) {
	type serializedContext struct {
		ID           string            `json:"id"`
		URI          string            `json:"uri"`
		ParentURI    string            `json:"parent_uri,omitempty"`
		IsLeaf       bool              `json:"is_leaf"`
		Abstract     string            `json:"abstract"`
		ContextType  string            `json:"context_type"`
		Category     string            `json:"category,omitempty"`
		CreatedAt    string            `json:"created_at"`
		UpdatedAt    string            `json:"updated_at"`
		ActiveCount  int64             `json:"active_count"`
		RelatedURI   []string          `json:"related_uri,omitempty"`
		Meta         map[string]any   `json:"meta,omitempty"`
		SessionID    string            `json:"session_id,omitempty"`
		UserID       string            `json:"user_id,omitempty"`
		Vector       []float64         `json:"vector,omitempty"`
		Vectorize    Vectorize         `json:"vectorize"`
		Tier         int               `json:"tier"`
	}

	var serialized []serializedContext
	if err := json.Unmarshal(data, &serialized); err != nil {
		return nil, err
	}

	contexts := make([]*Context, len(serialized))
	for i, s := range serialized {
		createdAt, _ := time.Parse(time.RFC3339, s.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, s.UpdatedAt)

		contexts[i] = &Context{
			ID:          s.ID,
			URI:         s.URI,
			ParentURI:   s.ParentURI,
			IsLeaf:      s.IsLeaf,
			Abstract:    s.Abstract,
			ContextType: ContextType(s.ContextType),
			Category:    Category(s.Category),
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			ActiveCount: s.ActiveCount,
			RelatedURI:  s.RelatedURI,
			Meta:        s.Meta,
			SessionID:   s.SessionID,
			UserID:      s.UserID,
			Vector:      s.Vector,
			Vectorize:   s.Vectorize,
			Tier:        ContextTier(s.Tier),
		}
	}

	return contexts, nil
}

// Persistable is an interface for types that can be persisted.
type Persistable interface {
	Save() error
	Load() error
}

// AutoSaver handles automatic saving of context.
type AutoSaver struct {
	interval time.Duration
	handler  *PersistenceHandler
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewAutoSaver creates a new AutoSaver.
func NewAutoSaver(interval time.Duration, handler *PersistenceHandler) *AutoSaver {
	return &AutoSaver{
		interval: interval,
		handler:  handler,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start starts the auto-saver.
func (as *AutoSaver) Start() {
	go func() {
		ticker := time.NewTicker(as.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := as.handler.Save(); err != nil {
					fmt.Printf("AutoSave error: %v\n", err)
				}
			case <-as.stopCh:
				// Do a final save before stopping
				if err := as.handler.Save(); err != nil {
					fmt.Printf("Final AutoSave error: %v\n", err)
				}
				close(as.doneCh)
				return
			}
		}
	}()
}

// Stop stops the auto-saver.
func (as *AutoSaver) Stop() error {
	close(as.stopCh)
	<-as.doneCh
	return nil
}
