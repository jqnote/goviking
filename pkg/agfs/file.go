// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"io"
	"os"
	"path/filepath"
)

// Read reads the contents of a file at the given URI.
func (a *AGFS) Read(uri string, offset int64, size int64) ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return nil, ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, ErrIsDirectory
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if offset > 0 {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
	}

	var data []byte
	if size > 0 {
		data = make([]byte, size)
		n, err := file.Read(data)
		if err != nil && err != io.EOF {
			return nil, err
		}
		data = data[:n]
	} else {
		data, err = io.ReadAll(file)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// Write writes data to a file at the given URI.
func (a *AGFS) Write(uri string, data []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	// Ensure parent directory exists
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Append appends data to a file at the given URI.
func (a *AGFS) Append(uri string, data []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	// Check if file exists
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Ensure parent directory exists
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	// Append data
	combined := append(existing, data...)
	return os.WriteFile(path, combined, 0644)
}

// Delete deletes a file at the given URI.
func (a *AGFS) Delete(uri string, recursive bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}

	if info.IsDir() {
		if recursive {
			return os.RemoveAll(path)
		}
		return os.Remove(path)
	}

	return os.Remove(path)
}

// Move moves a file or directory from one URI to another.
func (a *AGFS) Move(oldURI, newURI string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	oldURI = a.normalizeURI(oldURI)
	newURI = a.normalizeURI(newURI)

	oldPath := a.URIToPath(oldURI)
	newPath := a.URIToPath(newURI)

	if oldPath == "" || newPath == "" {
		return ErrInvalidURI
	}

	// Check source exists
	_, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}

	// Ensure destination parent directory exists
	parent := filepath.Dir(newPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	// Check if destination already exists
	if _, err := os.Stat(newPath); err == nil {
		return ErrAlreadyExists
	}

	return os.Rename(oldPath, newPath)
}

// Copy copies a file or directory from one URI to another.
func (a *AGFS) Copy(oldURI, newURI string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	oldURI = a.normalizeURI(oldURI)
	newURI = a.normalizeURI(newURI)

	oldPath := a.URIToPath(oldURI)
	newPath := a.URIToPath(newURI)

	if oldPath == "" || newPath == "" {
		return ErrInvalidURI
	}

	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}

	// Ensure destination parent directory exists
	parent := filepath.Dir(newPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	if info.IsDir() {
		return a.copyDir(oldPath, newPath)
	}

	return a.copyFile(oldPath, newPath)
}

// copyDir recursively copies a directory.
func (a *AGFS) copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := a.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := a.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file.
func (a *AGFS) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Stat returns information about a file or directory at the given URI.
func (a *AGFS) Stat(uri string) (*Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return nil, ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	isDir := info.IsDir()
	return &Entry{
		Name:     filepath.Base(path),
		Path:     path,
		URI:      a.PathToURI(path),
		Size:     info.Size(),
		Mode:     info.Mode(),
		ModTime:  info.ModTime(),
		IsDir:    isDir,
		FileType: determineFileType(uri, isDir),
	}, nil
}

// Exists checks if a file or directory exists at the given URI.
func (a *AGFS) Exists(uri string) bool {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if the URI points to a directory.
func (a *AGFS) IsDir(uri string) bool {
	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
