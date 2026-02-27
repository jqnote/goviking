// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package agfs provides the Agent Graph File System (AGFS) implementation.
// AGFS is a filesystem-like context management system that organizes memories,
// resources, and skills in a hierarchical directory structure.
package agfs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a file or directory is not found.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists is returned when attempting to create something that already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotADirectory is returned when a directory operation is performed on a non-directory.
	ErrNotADirectory = errors.New("not a directory")
	// ErrIsDirectory is returned when a file operation is performed on a directory.
	ErrIsDirectory = errors.New("is a directory")
	// ErrInvalidURI is returned when an invalid URI is provided.
	ErrInvalidURI = errors.New("invalid URI")
	// ErrNotImplemented is returned when a feature is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")
)

// FileType represents the type of context file.
type FileType string

const (
	// FileTypeMemory represents a memory file.
	FileTypeMemory FileType = "memory"
	// FileTypeResource represents a resource file.
	FileTypeResource FileType = "resource"
	// FileTypeSkill represents a skill file.
	FileTypeSkill FileType = "skill"
	// FileTypeDirectory represents a directory.
	FileTypeDirectory FileType = "directory"
)

// Entry represents a file or directory entry.
type Entry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	URI      string    `json:"uri"`
	Size     int64     `json:"size"`
	Mode     os.FileMode `json:"mode"`
	ModTime  time.Time `json:"modTime"`
	IsDir    bool      `json:"isDir"`
	FileType FileType  `json:"fileType,omitempty"`
}

// TreeEntry represents an entry in the tree structure.
type TreeEntry struct {
	Name     string      `json:"name"`
	URI      string      `json:"uri"`
	IsDir    bool        `json:"isDir"`
	Children []*TreeEntry `json:"children,omitempty"`
	Abstract string      `json:"abstract,omitempty"`
	Overview string      `json:"overview,omitempty"`
}

// RelationEntry represents a relation between directories.
type RelationEntry struct {
	ID        string   `json:"id"`
	URIs      []string `json:"uris"`
	Reason    string   `json:"reason,omitempty"`
	CreatedAt string   `json:"created_at"`
}

// Config holds AGFS configuration.
type Config struct {
	// RootPath is the root directory for local filesystem storage.
	RootPath string
	// URIPrefix is the URI prefix for the virtual filesystem (e.g., "viking://").
	URIPrefix string
	// EnableMemories enables memory file support.
	EnableMemories bool
	// EnableResources enables resource file support.
	EnableResources bool
	// EnableSkills enables skill file support.
	EnableSkills bool
}

// DefaultConfig returns a default AGFS configuration.
func DefaultConfig() Config {
	return Config{
		RootPath:     "./data/viking",
		URIPrefix:    "viking://",
		EnableMemories: true,
		EnableResources: true,
		EnableSkills:   true,
	}
}

// AGFS represents the Agent Graph File System.
type AGFS struct {
	config    Config
	rootPath  string
	uriPrefix string
	mu        sync.RWMutex
}

// New creates a new AGFS instance with the given configuration.
func New(config Config) (*AGFS, error) {
	if config.RootPath == "" {
		config.RootPath = DefaultConfig().RootPath
	}
	if config.URIPrefix == "" {
		config.URIPrefix = DefaultConfig().URIPrefix
	}

	agfs := &AGFS{
		config:    config,
		rootPath:  config.RootPath,
		uriPrefix: config.URIPrefix,
	}

	// Ensure root directories exist
	if err := agfs.ensureRootDirs(); err != nil {
		return nil, err
	}

	return agfs, nil
}

// ensureRootDirs creates the initial directory structure.
func (a *AGFS) ensureRootDirs() error {
	dirs := []string{
		"session",
		"user/memories/preferences",
		"user/memories/entities",
		"user/memories/events",
		"agent/memories/cases",
		"agent/memories/patterns",
		"agent/instructions",
		"agent/skills",
		"resources",
	}

	for _, dir := range dirs {
		path := filepath.Join(a.rootPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// URIToPath converts a viking URI to a filesystem path.
func (a *AGFS) URIToPath(uri string) string {
	// viking://user/memories -> /local/user/memories -> {rootPath}/user/memories
	if !strings.HasPrefix(uri, a.uriPrefix) {
		return ""
	}
	remainder := strings.TrimPrefix(uri, a.uriPrefix)
	remainder = strings.TrimPrefix(remainder, "/")
	if remainder == "" {
		return a.rootPath
	}
	return filepath.Join(a.rootPath, remainder)
}

// PathToURI converts a filesystem path to a viking URI.
func (a *AGFS) PathToURI(path string) string {
	// {rootPath}/user/memories -> viking://user/memories
	if !strings.HasPrefix(path, a.rootPath) {
		// Assume it's already a viking URI
		if strings.HasPrefix(path, "/local/") {
			return "viking://" + strings.TrimPrefix(path, "/local/")
		}
		return "viking://" + strings.TrimPrefix(path, "/")
	}
	remainder := strings.TrimPrefix(path, a.rootPath)
	remainder = strings.TrimPrefix(remainder, string(filepath.Separator))
	if remainder == "" {
		return a.uriPrefix
	}
	return a.uriPrefix + remainder
}

// normalizeURI ensures the URI has the correct prefix.
func (a *AGFS) normalizeURI(uri string) string {
	if !strings.HasPrefix(uri, a.uriPrefix) {
		if strings.HasPrefix(uri, "/local/") {
			return "viking://" + strings.TrimPrefix(uri, "/local/")
		}
		if strings.HasPrefix(uri, "/") {
			return "viking://" + strings.TrimPrefix(uri, "/")
		}
	}
	return uri
}

// FileTypeFromURI determines the file type from the URI.
func FileTypeFromURI(uri string) FileType {
	uri = strings.ToLower(uri)
	if strings.Contains(uri, "/memories") {
		return FileTypeMemory
	}
	if strings.Contains(uri, "/skills") {
		return FileTypeSkill
	}
	if strings.Contains(uri, "/resources") {
		return FileTypeResource
	}
	return FileTypeResource
}
