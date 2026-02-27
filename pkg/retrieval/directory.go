// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DirectoryEntry represents a file or directory entry.
type DirectoryEntry struct {
	Path         string      `json:"path"`
	Name         string      `json:"name"`
	IsDir        bool        `json:"is_dir"`
	Size         int64       `json:"size"`
	ModTime      int64       `json:"mod_time"`
	IsLeaf       bool        `json:"is_leaf"`
	ParentURI    string      `json:"parent_uri,omitempty"`
	URI          string      `json:"uri,omitempty"`
	Abstract     string      `json:"abstract,omitempty"`
	ContentType  string      `json:"content_type,omitempty"`
}

// DirectoryTraverser handles recursive directory traversal.
type DirectoryTraverser struct {
	// Maximum depth for recursive traversal (0 = unlimited)
	MaxDepth int

	// File patterns to include (e.g., ["*.go", "*.md"])
	IncludePatterns []string

	// File patterns to exclude (e.g., ["*_test.go", ".git/*"])
	ExcludePatterns []string

	// Follow symlinks
	FollowSymlinks bool

	// Include hidden files/directories
	IncludeHidden bool

	// Maximum file size in bytes (0 = no limit)
	MaxFileSize int64
}

// NewDirectoryTraverser creates a new DirectoryTraverser with default options.
func NewDirectoryTraverser() *DirectoryTraverser {
	return &DirectoryTraverser{
		MaxDepth:        0,       // unlimited
		IncludePatterns: nil,     // include all
		ExcludePatterns: []string{"*_test.go", ".git/*", "node_modules/*"},
		FollowSymlinks:  false,
		IncludeHidden:   false,
		MaxFileSize:     0,       // no limit
	}
}

// Traverse performs recursive directory traversal.
func (dt *DirectoryTraverser) Traverse(ctx context.Context, rootPath string) ([]DirectoryEntry, error) {
	var entries []DirectoryEntry
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []error

	entryChan := make(chan *DirectoryEntry, 100)

	// Start worker pool
	workerCount := 4
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range entryChan {
				mu.Lock()
				entries = append(entries, *entry)
				mu.Unlock()
			}
		}()
	}

	// Walk the directory
	err := dt.walk(ctx, rootPath, 0, entryChan, &errs)
	if err != nil {
		return nil, err
	}

	close(entryChan)
	wg.Wait()

	return entries, nil
}

// walk recursively walks directory.
func (dt *DirectoryTraverser) walk(ctx context.Context, path string, depth int, entryChan chan<- *DirectoryEntry, errs *[]error) error {
	if dt.MaxDepth > 0 && depth > dt.MaxDepth {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fullPath := filepath.Join(path, entry.Name())

		// Skip hidden files/directories if not included
		if !dt.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Check exclude patterns
		if dt.matchesAnyPattern(fullPath, dt.ExcludePatterns) {
			continue
		}

		// Check include patterns (if specified)
		if len(dt.IncludePatterns) > 0 && !dt.matchesAnyPattern(fullPath, dt.IncludePatterns) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			*errs = append(*errs, err)
			continue
		}

		isSymlink := entry.Type()&os.ModeSymlink != 0
		if isSymlink && !dt.FollowSymlinks {
			continue
		}

		dirEntry := &DirectoryEntry{
			Path:    fullPath,
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		}

		if !entry.IsDir() {
			// Check file size limit
			if dt.MaxFileSize > 0 && info.Size() > dt.MaxFileSize {
				continue
			}
			dirEntry.IsLeaf = true
			dirEntry.ContentType = dt.detectContentType(fullPath)
		}

		entryChan <- dirEntry

		// Recurse into subdirectories
		if entry.IsDir() {
			dt.walk(ctx, fullPath, depth+1, entryChan, errs)
		}
	}

	return nil
}

// matchesAnyPattern checks if path matches any of the patterns.
func (dt *DirectoryTraverser) matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Convert glob pattern to match
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Also check full path for patterns like .git/*
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(path, suffix) {
				return true
			}
		}
	}
	return false
}

// detectContentType detects content type from file extension.
func (dt *DirectoryTraverser) detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "text/x-go"
	case ".py":
		return "text/x-python"
	case ".js", ".jsx":
		return "text/javascript"
	case ".ts", ".tsx":
		return "text/typescript"
	case ".java":
		return "text/x-java"
	case ".c", ".h":
		return "text/x-c"
	case ".cpp", ".cc", ".cxx":
		return "text/x-c++"
	case ".rs":
		return "text/x-rust"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/x-yaml"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".sql":
		return "text/x-sql"
	case ".sh", ".bash":
		return "application/x-sh"
	case ".toml":
		return "application/toml"
	default:
		return "text/plain"
	}
}

// TraverseWithFilter performs traversal with additional filtering.
func (dt *DirectoryTraverser) TraverseWithFilter(ctx context.Context, rootPath string, filterFn func(*DirectoryEntry) bool) ([]DirectoryEntry, error) {
	allEntries, err := dt.Traverse(ctx, rootPath)
	if err != nil {
		return nil, err
	}

	if filterFn == nil {
		return allEntries, nil
	}

	var filtered []DirectoryEntry
	for _, entry := range allEntries {
		if filterFn(&entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}
