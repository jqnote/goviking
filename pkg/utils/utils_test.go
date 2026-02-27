// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

func TestGenerateIDWithPrefix(t *testing.T) {
	id := GenerateIDWithPrefix("test")
	if len(id) <= 5 {
		t.Error("ID with prefix should be longer than prefix")
	}
}

func TestNow(t *testing.T) {
	now := Now()
	if now.IsZero() {
		t.Error("Now should not be zero")
	}
}

func TestNowMillis(t *testing.T) {
	millis := NowMillis()
	if millis == 0 {
		t.Error("NowMillis should not be zero")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"short", 10, "short"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("TruncateString(%q, %d) = %q; want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world"},
		{"hello  world", "hello world"},
		{"hello\nworld", "hello world"},
		{"  hello  ", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := CleanWhitespace(tt.input)
		if result != tt.expected {
			t.Errorf("CleanWhitespace(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !Contains(slice, "a") {
		t.Error("Should contain 'a'")
	}
	if !Contains(slice, "c") {
		t.Error("Should contain 'c'")
	}
	if Contains(slice, "d") {
		t.Error("Should not contain 'd'")
	}
}

func TestUnique(t *testing.T) {
	slice := []string{"a", "b", "a", "c", "b"}
	result := Unique(slice)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique items, got %d", len(result))
	}
}
