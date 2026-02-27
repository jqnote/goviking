// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// ErrBackendNotSupported is returned when backend is not supported.
	ErrBackendNotSupported = errors.New("backend not supported")
	// ErrNotInitialized is returned when client is not initialized.
	ErrNotInitialized = errors.New("client not initialized")
)

// BackendType represents the type of AGFS backend.
type BackendType string

const (
	BackendLocal  BackendType = "local"
	BackendMemory BackendType = "memory"
)

// AGFSConfig represents AGFS configuration.
type AGFSConfig struct {
	Backend BackendType
	RootPath string // For local backend
}

// BindingClient provides AGFS binding functionality.
type BindingClient struct {
	config    AGFSConfig
	backend   Backend
	initialized bool
	mu        sync.RWMutex
}

// Backend defines the interface for AGFS backends.
type Backend interface {
	// Read reads a file.
	Read(path string) (string, error)
	// Write writes a file.
	Write(path string, content string) error
	// Delete deletes a file.
	Delete(path string) error
	// List lists files in a directory.
	List(path string) ([]FileInfo, error)
	// Mkdir creates a directory.
	Mkdir(path string) error
	// Exists checks if path exists.
	Exists(path string) (bool, error)
}

// FileInfo represents file information.
type FileInfo struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// MemoryBackend provides in-memory filesystem.
type MemoryBackend struct {
	files map[string]*memoryFile
	mu    sync.RWMutex
}

type memoryFile struct {
	content string
	isDir   bool
	modTime time.Time
}

// NewMemoryBackend creates a new memory backend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		files: make(map[string]*memoryFile),
	}
}

func (m *MemoryBackend) Read(path string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, ok := m.files[path]
	if !ok {
		return "", os.ErrNotExist
	}
	if file.isDir {
		return "", fmt.Errorf("%s is a directory", path)
	}
	return file.content, nil
}

func (m *MemoryBackend) Write(path string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[path] = &memoryFile{
		content: content,
		isDir:   false,
		modTime: time.Now(),
	}
	return nil
}

func (m *MemoryBackend) Delete(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.files[path]; !ok {
		return os.ErrNotExist
	}
	delete(m.files, path)
	return nil
}

func (m *MemoryBackend) List(path string) ([]FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []FileInfo
	prefix := path
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	for name, file := range m.files {
		if strings.HasPrefix(name, prefix) {
			relPath := strings.TrimPrefix(name, prefix)
			if relPath == "" {
				continue
			}
			// Only list direct children
			if strings.Contains(relPath, "/") {
				continue
			}
			results = append(results, FileInfo{
				Name:    relPath,
				Path:    name,
				IsDir:   file.isDir,
				Size:    int64(len(file.content)),
				ModTime: file.modTime,
			})
		}
	}

	return results, nil
}

func (m *MemoryBackend) Mkdir(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[path] = &memoryFile{
		content: "",
		isDir:   true,
		modTime: time.Now(),
	}
	return nil
}

func (m *MemoryBackend) Exists(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.files[path]
	return ok, nil
}

// LocalBackend provides local filesystem access.
type LocalBackend struct {
	rootPath string
}

// NewLocalBackend creates a new local backend.
func NewLocalBackend(rootPath string) *LocalBackend {
	return &LocalBackend{
		rootPath: rootPath,
	}
}

func (l *LocalBackend) resolvePath(path string) string {
	if l.rootPath == "" {
		return path
	}
	return filepath.Join(l.rootPath, path)
}

func (l *LocalBackend) Read(path string) (string, error) {
	data, err := os.ReadFile(l.resolvePath(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (l *LocalBackend) Write(path string, content string) error {
	fullPath := l.resolvePath(parentDir(path))
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}
	return os.WriteFile(l.resolvePath(path), []byte(content), 0644)
}

func (l *LocalBackend) Delete(path string) error {
	return os.Remove(l.resolvePath(path))
}

func (l *LocalBackend) List(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(l.resolvePath(path))
	if err != nil {
		return nil, err
	}

	var results []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		results = append(results, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return results, nil
}

func (l *LocalBackend) Mkdir(path string) error {
	return os.MkdirAll(l.resolvePath(path), 0755)
}

func (l *LocalBackend) Exists(path string) (bool, error) {
	_, err := os.Stat(l.resolvePath(path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func parentDir(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	return dir
}

// NewBindingClient creates a new AGFS binding client.
func NewBindingClient() *BindingClient {
	return &BindingClient{}
}

// Init initializes the AGFS client.
func (c *BindingClient) Init(config AGFSConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = config

	switch config.Backend {
	case BackendLocal:
		c.backend = NewLocalBackend(config.RootPath)
	case BackendMemory:
		c.backend = NewMemoryBackend()
	default:
		return ErrBackendNotSupported
	}

	c.initialized = true
	return nil
}

// Read reads a file through AGFS binding.
func (c *BindingClient) Read(ctx context.Context, path string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return "", ErrNotInitialized
	}

	return c.backend.Read(path)
}

// Write writes a file through AGFS binding.
func (c *BindingClient) Write(ctx context.Context, path string, content string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	return c.backend.Write(path, content)
}

// Delete deletes a file through AGFS binding.
func (c *BindingClient) Delete(ctx context.Context, path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	return c.backend.Delete(path)
}

// List lists files through AGFS binding.
func (c *BindingClient) List(ctx context.Context, path string) ([]FileInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil, ErrNotInitialized
	}

	return c.backend.List(path)
}

// Mkdir creates a directory through AGFS binding.
func (c *BindingClient) Mkdir(ctx context.Context, path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	return c.backend.Mkdir(path)
}

// Exists checks if path exists through AGFS binding.
func (c *BindingClient) Exists(ctx context.Context, path string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return false, ErrNotInitialized
	}

	return c.backend.Exists(path)
}
