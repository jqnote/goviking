// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrFileNotFound is returned when file is not found.
	ErrFileNotFound = errors.New("file not found")
	// ErrFileExists is returned when file already exists.
	ErrFileExists = errors.New("file already exists")
	// ErrInvalidPath is returned when path is invalid.
	ErrInvalidPath = errors.New("invalid path")
)

// FileInfo represents file information.
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"is_dir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// FSService provides filesystem-like operations.
type FSService struct {
	basePath string
}

// NewFSService creates a new filesystem service.
func NewFSService(basePath string) *FSService {
	return &FSService{
		basePath: basePath,
	}
}

// List lists files in a directory.
func (s *FSService) List(ctx context.Context, path string) ([]FileInfo, error) {
	fullPath := s.resolvePath(path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

// Mkdir creates a directory.
func (s *FSService) Mkdir(ctx context.Context, path string) error {
	fullPath := s.resolvePath(path)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}
	return nil
}

// Read reads a file.
func (s *FSService) Read(ctx context.Context, path string) (string, error) {
	fullPath := s.resolvePath(path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrFileNotFound
		}
		return "", err
	}

	return string(data), nil
}

// Write writes content to a file.
func (s *FSService) Write(ctx context.Context, path string, content string) error {
	fullPath := s.resolvePath(path)

	// Ensure parent directory exists
	parent := filepath.Dir(fullPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}
	return nil
}

// Delete deletes a file or directory.
func (s *FSService) Delete(ctx context.Context, path string) error {
	fullPath := s.resolvePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(fullPath)
	}
	return os.Remove(fullPath)
}

// Move moves a file or directory.
func (s *FSService) Move(ctx context.Context, from string, to string) error {
	fromPath := s.resolvePath(from)
	toPath := s.resolvePath(to)

	// Check source exists
	if _, err := os.Stat(fromPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}

	// Ensure target parent exists
	parent := filepath.Dir(toPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	return os.Rename(fromPath, toPath)
}

// Tree returns a tree representation of the directory.
func (s *FSService) Tree(ctx context.Context, path string) (string, error) {
	fullPath := s.resolvePath(path)

	var sb strings.Builder
	err := filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(fullPath, p)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		depth := len(filepath.SplitList(relPath))
		indent := ""
		for i := 0; i < depth-1; i++ {
			indent += "  "
		}

		if info.IsDir() {
			sb.WriteString(indent + "📁 " + info.Name() + "\n")
		} else {
			sb.WriteString(indent + "📄 " + info.Name() + "\n")
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// resolvePath resolves a relative path to an absolute path within basePath.
func (s *FSService) resolvePath(path string) string {
	if s.basePath == "" {
		return path
	}
	return filepath.Join(s.basePath, path)
}
