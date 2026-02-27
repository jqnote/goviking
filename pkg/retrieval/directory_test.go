// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryTraverser(t *testing.T) {
	// Create a temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "retrieval-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"file1.txt",
		"subdir/file2.go",
		"subdir/file3.md",
		"subdir/deep/file4.py",
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test traversal
	traverser := NewDirectoryTraverser()
	traverser.IncludeHidden = false
	traverser.MaxDepth = 0

	ctx := context.Background()
	entries, err := traverser.Traverse(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Traverse failed: %v", err)
	}

	// Count files (not directories)
	fileCount := 0
	for _, e := range entries {
		if e.IsLeaf {
			fileCount++
		}
	}

	if fileCount != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), fileCount)
	}
}

func TestDirectoryTraverserWithMaxDepth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "retrieval-depth-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	paths := []string{
		"level1/level2/level3/file.txt",
	}
	for _, p := range paths {
		path := filepath.Join(tmpDir, p)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test with depth limit
	traverser := NewDirectoryTraverser()
	traverser.MaxDepth = 1

	ctx := context.Background()
	entries, err := traverser.Traverse(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Traverse failed: %v", err)
	}

	// Should find level1 directory but not level2/level3
	foundLevel3 := false
	for _, e := range entries {
		if e.Path == filepath.Join(tmpDir, "level1/level2/level3") {
			foundLevel3 = true
		}
	}

	if foundLevel3 {
		t.Error("Should not have traversed beyond max depth")
	}
}

func TestPatternMatcher(t *testing.T) {
	pm, err := NewPatternMatcher([]string{"*.go", "*.md"}, []string{"*_test.go"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"readme.md", true},
		{"util_test.go", false}, // excluded
		{"image.png", false},    // not included
	}

	for _, tt := range tests {
		result := pm.Match(tt.path)
		if result != tt.expected {
			t.Errorf("PatternMatcher.Match(%s) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}
