// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("http://localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if client == nil {
		t.Fatal("Client is nil")
	}
}

func TestNewClientDefaultURL(t *testing.T) {
	client, err := NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("Expected default URL, got %s", client.baseURL)
	}
}

func TestClientWithHTTPClient(t *testing.T) {
	customClient := &http.Client{}
	client, err := NewClient("http://localhost:8080", WithHTTPClient(customClient))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if client.httpClient != customClient {
		t.Error("Custom HTTP client not set")
	}
}
