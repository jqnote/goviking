// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package utils provides common utility functions for GoViking.
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// GenerateID generates a unique ID.
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateIDWithPrefix generates a unique ID with prefix.
func GenerateIDWithPrefix(prefix string) string {
	id := GenerateID()
	if prefix != "" {
		return prefix + "_" + id
	}
	return id
}

// Now returns current UTC time.
func Now() time.Time {
	return time.Now().UTC()
}

// NowMillis returns current time in milliseconds.
func NowMillis() int64 {
	return time.Now().UnixMilli()
}

// FormatTime formats time to RFC3339.
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses time from RFC3339.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// TruncateString truncates a string to max length.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CleanWhitespace cleans extra whitespace from string.
func CleanWhitespace(s string) string {
	// Simple implementation - trim first, then collapse internal spaces
	s = strings.TrimSpace(s)
	result := ""
	inSpace := false
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if !inSpace {
				result += " "
				inSpace = true
			}
		} else {
			result += string(c)
			inSpace = false
		}
	}
	return result
}

// Contains checks if slice contains string.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Unique returns unique strings from slice.
func Unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
