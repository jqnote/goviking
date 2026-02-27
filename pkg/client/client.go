// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package client provides GoViking client SDK.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a synchronous client for GoViking.
type Client struct {
	baseURL  string
	httpClient *http.Client
}

// Option is a client option.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		client.httpClient = c
	}
}

// NewClient creates a new GoViking client.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// ContextService provides context operations.
type ContextService service

type service struct {
	client *Client
}

// Context represents a context entry.
type Context struct {
	ID          string                 `json:"id"`
	URI         string                 `json:"uri"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// Session represents a session.
type Session struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	State       string                 `json:"state"`
	Summary     string                 `json:"summary,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// CreateContext creates a new context.
func (c *Client) CreateContext(ctx context.Context, req *Context) (*Context, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/contexts", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create context failed: %d", resp.StatusCode)
	}

	var result Context
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContext retrieves a context by ID.
func (c *Client) GetContext(ctx context.Context, id string) (*Context, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/contexts/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get context failed: %d", resp.StatusCode)
	}

	var result Context
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListContexts lists all contexts.
func (c *Client) ListContexts(ctx context.Context) ([]Context, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/contexts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list contexts failed: %d", resp.StatusCode)
	}

	var result []Context
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteContext deletes a context.
func (c *Client) DeleteContext(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/contexts/%s", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete context failed: %d", resp.StatusCode)
	}

	return nil
}

// CreateSession creates a new session.
func (c *Client) CreateSession(ctx context.Context, req *Session) (*Session, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/sessions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create session failed: %d", resp.StatusCode)
	}

	var result Session
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSession retrieves a session by ID.
func (c *Client) GetSession(ctx context.Context, id string, mustExist bool) (*Session, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/sessions/%s?must_exist=%v", id, mustExist), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound && !mustExist {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get session failed: %d", resp.StatusCode)
	}

	var result Session
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SessionExists checks if a session exists.
func (c *Client) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	resp, err := c.doRequest(ctx, "HEAD", fmt.Sprintf("/api/v1/sessions/%s", sessionID), nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("check session exists failed: %d", resp.StatusCode)
}

// ListSessions lists all sessions.
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/sessions", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list sessions failed: %d", resp.StatusCode)
	}

	var result []Session
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// doRequest performs an HTTP request.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	if reqBody != nil {
		req.Body = nil
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(reqBody)), nil
		}
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
