// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidPackData is returned when pack data is invalid.
	ErrInvalidPackData = errors.New("invalid pack data")
)

// PackService provides pack/export/import functionality.
type PackService struct {
	fsService *FSService
}

// OVPackHeader represents the header of an OVPack file.
type OVPackHeader struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"` // "session", "context", "full"
	Checksum  string    `json:"checksum"`
}

// OVPack represents an OVPack file.
type OVPack struct {
	Header  OVPackHeader `json:"header"`
	Files   []PackFile   `json:"files"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// PackFile represents a file in an OVPack.
type PackFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Type    string `json:"type"` // "file", "dir"
}

// NewPackService creates a new pack service.
func NewPackService(fsService *FSService) *PackService {
	return &PackService{
		fsService: fsService,
	}
}

// Export exports files to OVPack format.
func (s *PackService) Export(ctx context.Context, paths []string) ([]byte, error) {
	if len(paths) == 0 {
		return nil, errors.New("no paths specified")
	}

	pack := OVPack{
		Header: OVPackHeader{
			Version:   "1.0",
			CreatedAt: time.Now().UTC(),
			Type:      "full",
		},
		Files: make([]PackFile, 0),
		Meta:  make(map[string]any),
	}

	// Export each path
	for _, path := range paths {
		if s.fsService == nil {
			continue
		}

		// Try to read as file first
		content, err := s.fsService.Read(ctx, path)
		if err == nil {
			pack.Files = append(pack.Files, PackFile{
				Path:    path,
				Content: content,
				Type:    "file",
			})
			continue
		}

		// Try to list as directory
		files, err := s.fsService.List(ctx, path)
		if err == nil {
			for _, f := range files {
				if f.IsDir {
					continue
				}
				content, err := s.fsService.Read(ctx, f.Path)
				if err == nil {
					pack.Files = append(pack.Files, PackFile{
						Path:    f.Path,
						Content: content,
						Type:    "file",
					})
				}
			}
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(pack)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pack: %w", err)
	}

	// Add simple checksum (first 8 chars of UUID as placeholder)
	pack.Header.Checksum = uuid.New().String()[:8]

	return data, nil
}

// Import imports OVPack data.
func (s *PackService) Import(ctx context.Context, data []byte) error {
	// Validate data
	if len(data) == 0 {
		return ErrInvalidPackData
	}

	// Parse JSON
	var pack OVPack
	if err := json.Unmarshal(data, &pack); err != nil {
		return fmt.Errorf("failed to unmarshal pack: %w", err)
	}

	// Validate header
	if pack.Header.Version == "" {
		return fmt.Errorf("invalid pack: missing version")
	}

	// Import each file
	for _, file := range pack.Files {
		if file.Type != "file" {
			continue
		}

		if s.fsService != nil {
			if err := s.fsService.Write(ctx, file.Path, file.Content); err != nil {
				return fmt.Errorf("failed to write %s: %w", file.Path, err)
			}
		}
	}

	return nil
}

// Validate validates OVPack data before import.
func (s *PackService) Validate(ctx context.Context, data []byte) (bool, string, error) {
	if len(data) == 0 {
		return false, "empty data", nil
	}

	// Try to parse
	var pack OVPack
	if err := json.Unmarshal(data, &pack); err != nil {
		return false, fmt.Sprintf("invalid JSON: %v", err), nil
	}

	// Check version
	if pack.Header.Version == "" {
		return false, "missing version", nil
	}

	// Check files
	if len(pack.Files) == 0 {
		return false, "no files in pack", nil
	}

	return true, "valid", nil
}

// PackExporter exports data in custom formats.
type PackExporter struct{}

// NewPackExporter creates a new pack exporter.
func NewPackExporter() *PackExporter {
	return &PackExporter{}
}

// ExportJSON exports data as JSON.
func (e *PackExporter) ExportJSON(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// ExportJSONStream exports data as JSON stream.
func (e *PackExporter) ExportJSONStream(data any) (io.Reader, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(dataBytes), nil
}
