// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

// PatternMatcher matches paths against patterns.
type PatternMatcher struct {
	includePatterns []*regexp.Regexp
	excludePatterns []*regexp.Regexp
}

// NewPatternMatcher creates a new PatternMatcher.
func NewPatternMatcher(includePatterns, excludePatterns []string) (*PatternMatcher, error) {
	pm := &PatternMatcher{
		includePatterns: make([]*regexp.Regexp, 0, len(includePatterns)),
		excludePatterns: make([]*regexp.Regexp, 0, len(excludePatterns)),
	}

	// Compile include patterns
	for _, p := range includePatterns {
		re, err := globToRegex(p)
		if err != nil {
			return nil, err
		}
		pm.includePatterns = append(pm.includePatterns, re)
	}

	// Compile exclude patterns
	for _, p := range excludePatterns {
		re, err := globToRegex(p)
		if err != nil {
			return nil, err
		}
		pm.excludePatterns = append(pm.excludePatterns, re)
	}

	return pm, nil
}

// globToRegex converts glob pattern to regex.
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// Convert glob to regex
	var sb strings.Builder
	sb.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			sb.WriteString(".*")
		case '?':
			sb.WriteString(".")
		case '.':
			sb.WriteString(`\.`)
		case '\\':
			sb.WriteString(`\\`)
		default:
			sb.WriteByte(pattern[i])
		}
	}

	sb.WriteString("$")

	return regexp.Compile(sb.String())
}

// Match checks if path matches the patterns.
func (pm *PatternMatcher) Match(path string) bool {
	// Check exclude patterns first
	for _, re := range pm.excludePatterns {
		if re.MatchString(path) {
			return false
		}
	}

	// If no include patterns, accept all
	if len(pm.includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, re := range pm.includePatterns {
		if re.MatchString(path) {
			return true
		}
	}

	return false
}

// PathRetriever retrieves files based on path patterns.
type PathRetriever struct {
	directoryTraverser *DirectoryTraverser
	patternMatcher     *PatternMatcher
}

// NewPathRetriever creates a new PathRetriever.
func NewPathRetriever(includePatterns, excludePatterns []string) (*PathRetriever, error) {
	pm, err := NewPatternMatcher(includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}

	return &PathRetriever{
		directoryTraverser: NewDirectoryTraverser(),
		patternMatcher:     pm,
	}, nil
}

// RetrieveByPath retrieves files/directories matching patterns.
func (pr *PathRetriever) RetrieveByPath(ctx context.Context, rootPath string) ([]DirectoryEntry, error) {
	entries, err := pr.directoryTraverser.Traverse(ctx, rootPath)
	if err != nil {
		return nil, err
	}

	var matched []DirectoryEntry
	for _, entry := range entries {
		relPath, err := filepath.Rel(rootPath, entry.Path)
		if err != nil {
			continue
		}

		if pr.patternMatcher.Match(relPath) || pr.patternMatcher.Match(entry.Name) {
			matched = append(matched, entry)
		}
	}

	return matched, nil
}

// RetrieveByPattern retrieves files matching specific pattern.
func (pr *PathRetriever) RetrieveByPattern(ctx context.Context, rootPath, pattern string) ([]DirectoryEntry, error) {
	// Create temporary pattern matcher for this pattern
	pm, err := NewPatternMatcher([]string{pattern}, nil)
	if err != nil {
		return nil, err
	}

	entries, err := pr.directoryTraverser.Traverse(ctx, rootPath)
	if err != nil {
		return nil, err
	}

	var matched []DirectoryEntry
	for _, entry := range entries {
		if pm.Match(entry.Name) {
			matched = append(matched, entry)
		}
	}

	return matched, nil
}

// FilterByExtension filters entries by file extension.
func FilterByExtension(entries []DirectoryEntry, extensions []string) []DirectoryEntry {
	if len(extensions) == 0 {
		return entries
	}

	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}

	var filtered []DirectoryEntry
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name))
		if extMap[ext] {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// FilterBySize filters entries by file size.
func FilterBySize(entries []DirectoryEntry, minSize, maxSize int64) []DirectoryEntry {
	var filtered []DirectoryEntry
	for _, entry := range entries {
		if entry.IsDir {
			continue
		}
		if minSize > 0 && entry.Size < minSize {
			continue
		}
		if maxSize > 0 && entry.Size > maxSize {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

// FilterByModTime filters entries by modification time.
func FilterByModTime(entries []DirectoryEntry, after, before int64) []DirectoryEntry {
	var filtered []DirectoryEntry
	for _, entry := range entries {
		if after > 0 && entry.ModTime < after {
			continue
		}
		if before > 0 && entry.ModTime > before {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}
