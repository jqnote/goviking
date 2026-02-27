// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"encoding/json"
	"os"
)

// Client provides a high-level API for the AGFS.
type Client struct {
	agfs      *AGFS
	relations *RelationManager
}

// NewClient creates a new AGFS client with the given configuration.
func NewClient(config Config) (*Client, error) {
	agfs, err := New(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		agfs:      agfs,
		relations: NewRelationManager(agfs),
	}, nil
}

// AGFS returns the underlying AGFS instance.
func (c *Client) AGFS() *AGFS {
	return c.agfs
}

// Relations returns the relation manager.
func (c *Client) Relations() *RelationManager {
	return c.relations
}

// Close closes the client (no-op for local filesystem).
func (c *Client) Close() error {
	return nil
}

// Ping checks if the filesystem is accessible.
func (c *Client) Ping() error {
	_, err := os.Stat(c.agfs.rootPath)
	return err
}

// ========== Convenience Methods ==========

// ReadText reads a text file and returns the content as a string.
func (c *Client) ReadText(uri string) (string, error) {
	data, err := c.agfs.Read(uri, 0, -1)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteText writes a text file from a string.
func (c *Client) WriteText(uri, content string) error {
	return c.agfs.Write(uri, []byte(content))
}

// AppendText appends text to a file.
func (c *Client) AppendText(uri, content string) error {
	return c.agfs.Append(uri, []byte(content))
}

// ReadJSON reads a JSON file and unmarshals it into the given interface.
func (c *Client) ReadJSON(uri string, v interface{}) error {
	data, err := c.agfs.Read(uri, 0, -1)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals the given interface as JSON and writes it to the file.
func (c *Client) WriteJSON(uri string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.agfs.Write(uri, data)
}

// ListDir is an alias for List that lists directory contents.
func (c *Client) ListDir(uri string) ([]Entry, error) {
	return c.agfs.List(uri, false)
}

// ListDirAll is an alias for List that lists directory contents including hidden files.
func (c *Client) ListDirAll(uri string) ([]Entry, error) {
	return c.agfs.List(uri, true)
}

// GetTree returns the directory tree.
func (c *Client) GetTree(uri string, maxDepth int) ([]TreeEntry, error) {
	return c.agfs.Tree(uri, maxDepth)
}

// CreateDir creates a new directory.
func (c *Client) CreateDir(uri string) error {
	return c.agfs.Mkdir(uri, 0755, false)
}

// CreateDirAll creates a new directory and all parent directories.
func (c *Client) CreateDirAll(uri string) error {
	return c.agfs.Mkdir(uri, 0755, true)
}

// RemoveDir removes a directory.
func (c *Client) RemoveDir(uri string) error {
	return c.agfs.Rmdir(uri, false)
}

// RemoveDirAll removes a directory and all its contents.
func (c *Client) RemoveDirAll(uri string) error {
	return c.agfs.Rmdir(uri, true)
}

// RemoveFile removes a file.
func (c *Client) RemoveFile(uri string) error {
	return c.agfs.Delete(uri, false)
}

// Rename renames (moves) a file or directory.
func (c *Client) Rename(oldURI, newURI string) error {
	return c.agfs.Move(oldURI, newURI)
}

// CopyFile copies a file or directory.
func (c *Client) CopyFile(oldURI, newURI string) error {
	return c.agfs.Copy(oldURI, newURI)
}

// FileInfo returns information about a file or directory.
func (c *Client) FileInfo(uri string) (*Entry, error) {
	return c.agfs.Stat(uri)
}

// FileExists checks if a file or directory exists.
func (c *Client) FileExists(uri string) bool {
	return c.agfs.Exists(uri)
}

// IsDirectory checks if the URI points to a directory.
func (c *Client) IsDirectory(uri string) bool {
	return c.agfs.IsDir(uri)
}

// GetAbstract reads the abstract (L0) of a directory.
func (c *Client) GetAbstract(uri string) (string, error) {
	return c.agfs.ReadAbstract(uri)
}

// SetAbstract writes the abstract (L0) of a directory.
func (c *Client) SetAbstract(uri, abstract string) error {
	return c.agfs.WriteAbstract(uri, abstract)
}

// GetOverview reads the overview (L1) of a directory.
func (c *Client) GetOverview(uri string) (string, error) {
	return c.agfs.ReadOverview(uri)
}

// SetOverview writes the overview (L1) of a directory.
func (c *Client) SetOverview(uri, overview string) error {
	return c.agfs.WriteOverview(uri, overview)
}

// GetContext reads all context levels (L0, L1, L2) of a URI.
func (c *Client) GetContext(uri string) (*ContextFile, error) {
	return c.agfs.ReadContext(uri)
}

// SetContext writes all context levels (L0, L1, L2) of a URI.
func (c *Client) SetContext(uri, abstract, overview, content string, isLeaf bool) error {
	return c.agfs.WriteContext(uri, abstract, overview, content, isLeaf)
}

// Search performs a grep search in a directory.
func (c *Client) Search(uri, pattern string) ([]GrepMatch, error) {
	return c.agfs.Grep(uri, pattern, false)
}

// SearchCI performs a case-insensitive grep search.
func (c *Client) SearchCI(uri, pattern string) ([]GrepMatch, error) {
	return c.agfs.Grep(uri, pattern, true)
}

// Glob performs pattern matching on files.
func (c *Client) Glob(uri, pattern string) ([]string, error) {
	return c.agfs.Glob(uri, pattern)
}

// Link creates a relation between directories.
func (c *Client) Link(fromURI string, uris []string, reason string) error {
	return c.relations.Link(fromURI, uris, reason)
}

// Unlink removes a relation between directories.
func (c *Client) Unlink(fromURI, targetURI string) error {
	return c.relations.Unlink(fromURI, targetURI)
}

// GetLinks returns all relations from a directory.
func (c *Client) GetLinks(uri string) ([]RelationEntry, error) {
	return c.relations.GetRelations(uri)
}

// GetLinkedURIs returns all URIs linked from a directory.
func (c *Client) GetLinkedURIs(uri string) ([]string, error) {
	return c.relations.GetRelatedURIs(uri)
}
