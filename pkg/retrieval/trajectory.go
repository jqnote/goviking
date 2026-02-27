// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"sync"
	"time"
)

// TrajectoryNode represents a node in the retrieval trajectory.
type TrajectoryNode struct {
	URI         string                 `json:"uri"`
	Depth       int                    `json:"depth"`
	Score       float64                `json:"score"`
	Timestamp   time.Duration          `json:"timestamp"`
	Children    []string               `json:"children,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Trajectory tracks the retrieval path for debugging.
type Trajectory struct {
	mu       sync.RWMutex
	RootURI  string           `json:"root_uri"`
	StartAt  time.Time        `json:"start_at"`
	Nodes    map[string]*TrajectoryNode `json:"nodes"`
	Path     []string         `json:"path"` // ordered list of visited URIs
	Parents  map[string]string `json:"parents"` // child -> parent mapping
}

// NewTrajectory creates a new Trajectory.
func NewTrajectory(rootURI string) *Trajectory {
	return &Trajectory{
		RootURI: rootURI,
		StartAt: time.Now(),
		Nodes:   make(map[string]*TrajectoryNode),
		Path:    []string{},
		Parents: make(map[string]string),
	}
}

// AddNode adds a node to the trajectory.
func (t *Trajectory) AddNode(uri string, depth int, score float64, metadata map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := &TrajectoryNode{
		URI:       uri,
		Depth:     depth,
		Score:     score,
		Timestamp: time.Since(t.StartAt),
		Metadata:  metadata,
	}
	t.Nodes[uri] = node
	t.Path = append(t.Path, uri)
}

// AddEdge adds an edge between parent and child.
func (t *Trajectory) AddEdge(parentURI, childURI string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Parents[childURI] = parentURI

	if parentNode, ok := t.Nodes[parentURI]; ok {
		parentNode.Children = append(parentNode.Children, childURI)
	}
}

// GetPath returns the ordered list of visited URIs.
func (t *Trajectory) GetPath() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]string, len(t.Path))
	copy(result, t.Path)
	return result
}

// GetPathWithDepth returns path with depth information.
func (t *Trajectory) GetPathWithDepth() []map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(t.Path))
	for _, uri := range t.Path {
		if node, ok := t.Nodes[uri]; ok {
			result = append(result, map[string]interface{}{
				"uri":      node.URI,
				"depth":    node.Depth,
				"score":    node.Score,
				"duration": node.Timestamp.Seconds(),
			})
		}
	}
	return result
}

// GetNode returns a node by URI.
func (t *Trajectory) GetNode(uri string) (*TrajectoryNode, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node, ok := t.Nodes[uri]
	return node, ok
}

// GetAncestors returns all ancestors of a URI.
func (t *Trajectory) GetAncestors(uri string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var ancestors []string
	current := uri
	for {
		parent, ok := t.Parents[current]
		if !ok {
			break
		}
		ancestors = append(ancestors, parent)
		current = parent
	}
	// Reverse to get root -> leaf order
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}
	return ancestors
}

// ToMap converts trajectory to a map for serialization.
func (t *Trajectory) ToMap() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"root_uri":  t.RootURI,
		"start_at":  t.StartAt,
		"duration":   time.Since(t.StartAt).Seconds(),
		"node_count": len(t.Nodes),
		"path":      t.Path,
		"nodes":     t.Nodes,
	}
}

// TrajectoryLogger logs retrieval trajectory.
type TrajectoryLogger struct {
	mu           sync.RWMutex
	Trajectories map[string]*Trajectory
}

// NewTrajectoryLogger creates a new TrajectoryLogger.
func NewTrajectoryLogger() *TrajectoryLogger {
	return &TrajectoryLogger{
		Trajectories: make(map[string]*Trajectory),
	}
}

// CreateTrajectory creates and registers a new trajectory.
func (l *TrajectoryLogger) CreateTrajectory(rootURI string) *Trajectory {
	l.mu.Lock()
	defer l.mu.Unlock()

	trajectory := NewTrajectory(rootURI)
	l.Trajectories[rootURI] = trajectory
	return trajectory
}

// GetTrajectory returns a trajectory by root URI.
func (l *TrajectoryLogger) GetTrajectory(rootURI string) (*Trajectory, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	trajectory, ok := l.Trajectories[rootURI]
	return trajectory, ok
}

// GetAllTrajectories returns all trajectories.
func (l *TrajectoryLogger) GetAllTrajectories() map[string]*Trajectory {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]*Trajectory, len(l.Trajectories))
	for k, v := range l.Trajectories {
		result[k] = v
	}
	return result
}
