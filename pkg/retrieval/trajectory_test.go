// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"testing"
)

func TestTrajectory(t *testing.T) {
	traj := NewTrajectory("/root")

	// Add nodes
	traj.AddNode("/root", 0, 0.0, nil)
	traj.AddNode("/root/dir1", 1, 0.5, nil)
	traj.AddNode("/root/dir1/subdir", 2, 0.8, nil)

	// Add edges
	traj.AddEdge("/root", "/root/dir1")
	traj.AddEdge("/root/dir1", "/root/dir1/subdir")

	// Test GetPath
	path := traj.GetPath()
	if len(path) != 3 {
		t.Errorf("Expected path length 3, got %d", len(path))
	}

	// Test GetAncestors
	ancestors := traj.GetAncestors("/root/dir1/subdir")
	if len(ancestors) != 2 {
		t.Errorf("Expected 2 ancestors, got %d", len(ancestors))
	}
}

func TestTrajectoryLogger(t *testing.T) {
	logger := NewTrajectoryLogger()

	// Create trajectories
	traj1 := logger.CreateTrajectory("/root1")
	traj1.AddNode("/root1", 0, 0.0, nil)

	traj2 := logger.CreateTrajectory("/root2")
	traj2.AddNode("/root2", 0, 0.0, nil)

	// Get trajectories
	retrieved1, ok := logger.GetTrajectory("/root1")
	if !ok {
		t.Error("Failed to retrieve trajectory /root1")
	}
	if len(retrieved1.Path) != 1 {
		t.Errorf("Expected path length 1, got %d", len(retrieved1.Path))
	}

	// Get all trajectories
	all := logger.GetAllTrajectories()
	if len(all) != 2 {
		t.Errorf("Expected 2 trajectories, got %d", len(all))
	}
}
