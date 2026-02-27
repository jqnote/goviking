// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package agfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:     tmpDir,
		URIPrefix:    "viking://",
		EnableMemories: true,
		EnableResources: true,
		EnableSkills:   true,
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	// Check that root directories were created
	expectedDirs := []string{
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

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	_ = agfs // Use the variable
}

func TestURIToPath(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	tests := []struct {
		uri    string
		expect string
	}{
		{"viking://", tmpDir},
		{"viking://user", filepath.Join(tmpDir, "user")},
		{"viking://user/memories", filepath.Join(tmpDir, "user", "memories")},
		{"viking://agent/skills", filepath.Join(tmpDir, "agent", "skills")},
	}

	for _, tt := range tests {
		path := agfs.URIToPath(tt.uri)
		if path != tt.expect {
			t.Errorf("URIToPath(%s) = %s; want %s", tt.uri, path, tt.expect)
		}
	}
}

func TestPathToURI(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	tests := []struct {
		path string
		want string
	}{
		{tmpDir, "viking://"},
		{filepath.Join(tmpDir, "user"), "viking://user"},
		{filepath.Join(tmpDir, "user", "memories"), "viking://user/memories"},
	}

	for _, tt := range tests {
		uri := agfs.PathToURI(tt.path)
		if uri != tt.want {
			t.Errorf("PathToURI(%s) = %s; want %s", tt.path, uri, tt.want)
		}
	}
}

func TestMkdirAndList(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	// Create a directory
	testURI := "viking://user/memories/testdir"
	if err := agfs.Mkdir(testURI, 0755, false); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// List the parent directory
	parentURI := "viking://user/memories"
	entries, err := agfs.List(parentURI, false)
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	// Check that our directory is in the list
	found := false
	for _, entry := range entries {
		if entry.Name == "testdir" && entry.IsDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find testdir in listing")
	}
}

func TestWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	testURI := "viking://resources/test.txt"
	content := []byte("Hello, World!")

	// Write file
	if err := agfs.Write(testURI, content); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read file
	data, err := agfs.Read(testURI, 0, -1)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Read content = %s; want %s", string(data), string(content))
	}
}

func TestStat(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	// Write a test file
	testURI := "viking://resources/test.txt"
	content := []byte("Test content")
	if err := agfs.Write(testURI, content); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Stat the file
	entry, err := agfs.Stat(testURI)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if entry.Name != "test.txt" {
		t.Errorf("Name = %s; want test.txt", entry.Name)
	}
	if entry.IsDir {
		t.Error("Expected IsDir = false")
	}
	if entry.Size != int64(len(content)) {
		t.Errorf("Size = %d; want %d", entry.Size, len(content))
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	testURI := "viking://resources/test.txt"

	// Write a test file
	if err := agfs.Write(testURI, []byte("Test")); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Delete the file
	if err := agfs.Delete(testURI, false); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Check that it's gone
	if agfs.Exists(testURI) {
		t.Error("Expected file to be deleted")
	}
}

func TestMove(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	oldURI := "viking://resources/old.txt"
	newURI := "viking://resources/new.txt"

	// Write a test file
	if err := agfs.Write(oldURI, []byte("Test content")); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Move the file
	if err := agfs.Move(oldURI, newURI); err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// Check old file is gone
	if agfs.Exists(oldURI) {
		t.Error("Expected old file to be gone")
	}

	// Check new file exists
	if !agfs.Exists(newURI) {
		t.Error("Expected new file to exist")
	}

	// Check content is preserved
	data, _ := agfs.Read(newURI, 0, -1)
	if string(data) != "Test content" {
		t.Errorf("Content = %s; want Test content", string(data))
	}
}

func TestContextFiles(t *testing.T) {
	tmpDir := t.TempDir()
	config := Config{
		RootPath:  tmpDir,
		URIPrefix: "viking://",
	}

	agfs, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create AGFS: %v", err)
	}

	testURI := "viking://resources/testdir"

	// Write context
	abstract := "This is an abstract"
	overview := "This is an overview"
	content := "This is the content"

	if err := agfs.WriteContext(testURI, abstract, overview, content, true); err != nil {
		t.Fatalf("Failed to write context: %v", err)
	}

	// Read abstract
	abs, err := agfs.ReadAbstract(testURI)
	if err != nil {
		t.Fatalf("Failed to read abstract: %v", err)
	}
	if abs != abstract {
		t.Errorf("Abstract = %s; want %s", abs, abstract)
	}

	// Read overview
	ovw, err := agfs.ReadOverview(testURI)
	if err != nil {
		t.Fatalf("Failed to read overview: %v", err)
	}
	if ovw != overview {
		t.Errorf("Overview = %s; want %s", ovw, overview)
	}

	// Read content
	ctx, err := agfs.ReadContent(testURI)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}
	if ctx != content {
		t.Errorf("Content = %s; want %s", ctx, content)
	}
}

func TestFileTypeFromURI(t *testing.T) {
	tests := []struct {
		uri      string
		expected FileType
	}{
		{"viking://user/memories", FileTypeMemory},
		{"viking://user/memories/preferences", FileTypeMemory},
		{"viking://agent/skills", FileTypeSkill},
		{"viking://agent/skills/coding", FileTypeSkill},
		{"viking://resources", FileTypeResource},
		{"viking://resources/docs", FileTypeResource},
	}

	for _, tt := range tests {
		ft := FileTypeFromURI(tt.uri)
		if ft != tt.expected {
			t.Errorf("FileTypeFromURI(%s) = %s; want %s", tt.uri, ft, tt.expected)
		}
	}
}
