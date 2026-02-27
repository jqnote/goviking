// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ContextFile represents a context file with L0, L1, and L2 content.
type ContextFile struct {
	URI      string
	Abstract string  // L0: Summary
	Overview string  // L1: Description
	Content  string  // L2: Full content
	IsLeaf   bool    // Whether this is a leaf node
	FileType FileType
}

// WriteContext writes a context file with abstract, overview, and content.
func (a *AGFS) WriteContext(uri, abstract, overview, content string, isLeaf bool) error {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	// Ensure directory exists
	if err := a.Mkdir(uri, 0755, true); err != nil {
		return err
	}

	// Write abstract file
	if abstract != "" {
		abstractPath := filepath.Join(path, ".abstract.md")
		if err := a.WriteFile(abstractPath, []byte(abstract)); err != nil {
			return err
		}
	}

	// Write overview file
	if overview != "" {
		overviewPath := filepath.Join(path, ".overview.md")
		if err := a.WriteFile(overviewPath, []byte(overview)); err != nil {
			return err
		}
	}

	// Write content file
	if content != "" {
		contentPath := filepath.Join(path, "content.md")
		if err := a.WriteFile(contentPath, []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// WriteFile is an alias for Write that takes a filesystem path.
func (a *AGFS) WriteFile(path string, data []byte) error {
	// Ensure parent directory exists
	parent := filepath.Dir(path)
	if err := a.Mkdir(a.PathToURI(parent), 0755, true); err != nil {
		// Ignore error if parent is not a URI
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	return os.WriteFile(path, data, 0644)
}

// ReadAbstract reads the abstract (L0) content of a directory.
func (a *AGFS) ReadAbstract(uri string) (string, error) {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return "", ErrInvalidURI
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrNotADirectory
	}

	abstractPath := filepath.Join(path, ".abstract.md")
	data, err := os.ReadFile(abstractPath)
	if err != nil {
		return "", ErrNotFound
	}

	return string(data), nil
}

// ReadOverview reads the overview (L1) content of a directory.
func (a *AGFS) ReadOverview(uri string) (string, error) {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return "", ErrInvalidURI
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrNotADirectory
	}

	overviewPath := filepath.Join(path, ".overview.md")
	data, err := os.ReadFile(overviewPath)
	if err != nil {
		return "", ErrNotFound
	}

	return string(data), nil
}

// ReadContent reads the content (L2) of a directory or file.
func (a *AGFS) ReadContent(uri string) (string, error) {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return "", ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	var contentPath string
	if info.IsDir() {
		contentPath = filepath.Join(path, "content.md")
	} else {
		contentPath = path
	}

	data, err := os.ReadFile(contentPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ReadContext reads all context levels (L0, L1, L2) for a URI.
func (a *AGFS) ReadContext(uri string) (*ContextFile, error) {
	uri = a.normalizeURI(uri)

	entry, err := a.Stat(uri)
	if err != nil {
		return nil, err
	}

	ctx := &ContextFile{
		URI:      uri,
		IsLeaf:   !entry.IsDir,
		FileType: entry.FileType,
	}

	// Try to read abstract
	if abstract, err := a.ReadAbstract(uri); err == nil {
		ctx.Abstract = abstract
	}

	// Try to read overview
	if overview, err := a.ReadOverview(uri); err == nil {
		ctx.Overview = overview
	}

	// Try to read content
	if content, err := a.ReadContent(uri); err == nil {
		ctx.Content = content
	}

	return ctx, nil
}

// WriteAbstract writes the abstract (L0) content for a directory.
func (a *AGFS) WriteAbstract(uri, abstract string) error {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	// Ensure directory exists
	if err := a.Mkdir(uri, 0755, true); err != nil {
		return err
	}

	abstractPath := filepath.Join(path, ".abstract.md")
	return os.WriteFile(abstractPath, []byte(abstract), 0644)
}

// WriteOverview writes the overview (L1) content for a directory.
func (a *AGFS) WriteOverview(uri, overview string) error {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	// Ensure directory exists
	if err := a.Mkdir(uri, 0755, true); err != nil {
		return err
	}

	overviewPath := filepath.Join(path, ".overview.md")
	return os.WriteFile(overviewPath, []byte(overview), 0644)
}

// WriteContent writes the content (L2) for a directory or file.
func (a *AGFS) WriteContent(uri, content string) error {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	var contentPath string
	if info.IsDir() {
		// Ensure directory exists
		if err := a.Mkdir(uri, 0755, true); err != nil {
			return err
		}
		contentPath = filepath.Join(path, "content.md")
	} else {
		contentPath = path
	}

	return os.WriteFile(contentPath, []byte(content), 0644)
}

// Grep searches for a pattern in files within a directory.
func (a *AGFS) Grep(uri, pattern string, caseInsensitive bool) ([]GrepMatch, error) {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return nil, ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, ErrNotADirectory
	}

	var matches []GrepMatch
	err = a.grepRecursive(path, a.PathToURI(path), pattern, caseInsensitive, &matches)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// GrepMatch represents a single grep match.
type GrepMatch struct {
	URI     string `json:"uri"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// grepRecursive recursively searches for a pattern.
func (a *AGFS) grepRecursive(dirPath, dirURI, pattern string, caseInsensitive bool, matches *[]GrepMatch) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}

		entryPath := filepath.Join(dirPath, name)
		entryURI := a.PathToURI(entryPath)

		if entry.IsDir() {
			if err := a.grepRecursive(entryPath, entryURI, pattern, caseInsensitive, matches); err != nil {
				return err
			}
		} else {
			// Skip hidden files
			if len(name) > 0 && name[0] == '.' {
				continue
			}

			if err := a.grepFile(entryPath, entryURI, pattern, caseInsensitive, matches); err != nil {
				return err
			}
		}
	}

	return nil
}

// grepFile searches for a pattern in a single file.
func (a *AGFS) grepFile(filePath, fileURI, pattern string, caseInsensitive bool, matches *[]GrepMatch) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	if caseInsensitive {
		content = strings.ToLower(content)
		pattern = strings.ToLower(pattern)
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			*matches = append(*matches, GrepMatch{
				URI:     fileURI,
				Line:    i + 1,
				Content: strings.TrimRight(lines[i], "\r"),
			})
		}
	}

	return nil
}

// Glob performs pattern matching on files in a directory.
func (a *AGFS) Glob(uri, pattern string) ([]string, error) {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return nil, ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, ErrNotADirectory
	}

	var results []string
	err = a.globRecursive(path, uri, pattern, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// globRecursive recursively performs pattern matching.
func (a *AGFS) globRecursive(dirPath, dirURI, pattern string, results *[]string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}

		entryPath := filepath.Join(dirPath, name)
		entryURI := a.PathToURI(entryPath)

		// Simple pattern matching (supports * and ?)
		if matchPattern(name, pattern) {
			*results = append(*results, entryURI)
		}

		if entry.IsDir() {
			if err := a.globRecursive(entryPath, entryURI, pattern, results); err != nil {
				return err
			}
		}
	}

	return nil
}

// matchPattern performs simple glob pattern matching.
func matchPattern(name, pattern string) bool {
	// Simple implementation - just support * wildcard
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return name == pattern
	}

	idx := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		pos := strings.Index(name[idx:], part)
		if pos == -1 {
			return false
		}
		idx += pos + len(part)
	}
	return true
}

// Touch updates the modification time of a file or creates it if it doesn't exist.
func (a *AGFS) Touch(uri string) error {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Create empty file
		return os.WriteFile(path, []byte{}, 0644)
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return ErrIsDirectory
	}

	return os.Chtimes(path, time.Now(), time.Now())
}
