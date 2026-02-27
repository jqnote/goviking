// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"os"
	"path/filepath"
)

// Mkdir creates a new directory at the given URI.
func (a *AGFS) Mkdir(uri string, mode os.FileMode, existOk bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	uri = a.normalizeURI(uri)
	path := a.URIToPath(uri)
	if path == "" {
		return ErrInvalidURI
	}

	info, err := os.Stat(path)
	if err == nil {
		if existOk && info.IsDir() {
			return nil
		}
		return ErrAlreadyExists
	}

	if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path, mode)
}

// Rmdir removes a directory at the given URI.
func (a *AGFS) Rmdir(uri string, recursive bool) error {
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

	if !info.IsDir() {
		return ErrNotADirectory
	}

	if recursive {
		return os.RemoveAll(path)
	}

	return os.Remove(path)
}

// List lists the contents of a directory at the given URI.
func (a *AGFS) List(uri string, showHidden bool) ([]Entry, error) {
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

	if !info.IsDir() {
		return nil, ErrNotADirectory
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless requested
		if !showHidden && name != "." && name != ".." && len(name) > 0 && name[0] == '.' {
			continue
		}

		entryPath := filepath.Join(path, name)
		entryInfo, err := os.Stat(entryPath)
		if err != nil {
			continue
		}

		isDir := entry.IsDir()
		uri := a.PathToURI(entryPath)

		result = append(result, Entry{
			Name:     name,
			Path:     entryPath,
			URI:      uri,
			Size:     entryInfo.Size(),
			Mode:     entryInfo.Mode(),
			ModTime:  entryInfo.ModTime(),
			IsDir:    isDir,
			FileType: determineFileType(uri, isDir),
		})
	}

	return result, nil
}

// Tree returns the directory tree starting from the given URI.
func (a *AGFS) Tree(uri string, maxDepth int) ([]TreeEntry, error) {
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

	if !info.IsDir() {
		return nil, ErrNotADirectory
	}

	var entries []TreeEntry
	if err := a.walkTree(path, uri, 0, maxDepth, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// walkTree recursively walks the directory tree.
func (a *AGFS) walkTree(path, uri string, depth, maxDepth int, entries *[]TreeEntry) error {
	if maxDepth > 0 && depth >= maxDepth {
		return nil
	}

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		// Skip hidden files and special directories
		if name == "." || name == ".." {
			continue
		}
		if len(name) > 0 && name[0] == '.' && name != ".abstract.md" && name != ".overview.md" && name != ".relations.json" {
			continue
		}

		entryPath := filepath.Join(path, name)
		entryURI := a.PathToURI(entryPath)

		isDir := dirEntry.IsDir()
		treeEntry := TreeEntry{
			Name:  name,
			URI:   entryURI,
			IsDir: isDir,
		}

		if isDir {
			// Try to read abstract and overview
			if abstract, err := a.readAbstractFile(entryPath); err == nil {
				treeEntry.Abstract = abstract
			}
			if overview, err := a.readOverviewFile(entryPath); err == nil {
				treeEntry.Overview = overview
			}

			var children []TreeEntry
			a.walkTree(entryPath, entryURI, depth+1, maxDepth, &children)
			for i := range children {
				treeEntry.Children = append(treeEntry.Children, &children[i])
			}
		}

		*entries = append(*entries, treeEntry)
	}

	return nil
}

// readAbstractFile reads the .abstract.md file from a directory.
func (a *AGFS) readAbstractFile(dirPath string) (string, error) {
	abstractPath := filepath.Join(dirPath, ".abstract.md")
	data, err := os.ReadFile(abstractPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// readOverviewFile reads the .overview.md file from a directory.
func (a *AGFS) readOverviewFile(dirPath string) (string, error) {
	overviewPath := filepath.Join(dirPath, ".overview.md")
	data, err := os.ReadFile(overviewPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// determineFileType determines the file type from URI.
func determineFileType(uri string, isDir bool) FileType {
	if isDir {
		return FileTypeDirectory
	}
	return FileTypeFromURI(uri)
}
