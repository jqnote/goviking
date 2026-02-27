// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"testing"
	"time"
)

func TestContextCreation(t *testing.T) {
	ctx := NewContext("viking://agent/skills/test")

	if ctx.URI != "viking://agent/skills/test" {
		t.Errorf("expected URI viking://agent/skills/test, got %s", ctx.URI)
	}

	if ctx.ContextType != ContextTypeSkill {
		t.Errorf("expected context type skill, got %s", ctx.ContextType)
	}

	if ctx.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestTieredContext(t *testing.T) {
	tc := NewTieredContext()

	// Add contexts to different tiers
	l0Ctx := NewContext("viking://test/l0")
	l0Ctx.Tier = TierL0
	l1Ctx := NewContext("viking://test/l1")
	l1Ctx.Tier = TierL1
	l2Ctx := NewContext("viking://test/l2")
	l2Ctx.Tier = TierL2

	tc.Add(l0Ctx)
	tc.Add(l1Ctx)
	tc.Add(l2Ctx)

	if len(tc.GetL0()) != 1 {
		t.Errorf("expected 1 L0 context, got %d", len(tc.GetL0()))
	}

	if len(tc.GetL1()) != 1 {
		t.Errorf("expected 1 L1 context, got %d", len(tc.GetL1()))
	}

	if len(tc.GetL2()) != 1 {
		t.Errorf("expected 1 L2 context, got %d", len(tc.GetL2()))
	}

	if tc.Count() != 3 {
		t.Errorf("expected 3 total contexts, got %d", tc.Count())
	}
}

func TestContextBuilder(t *testing.T) {
	// Create test contexts
	memories := []*Context{
		{URI: "viking://user/memories/profile", Abstract: "User profile", ContextType: ContextTypeMemory},
	}
	resources := []*Context{
		{URI: "viking://resource/doc1", Abstract: "Document 1", ContextType: ContextTypeResource},
	}
	skills := []*Context{
		{URI: "viking://agent/skills/bash", Abstract: "Bash skill", ContextType: ContextTypeSkill},
	}

	builder := NewContextBuilder().
		AddMemorySource(memories).
		AddResourceSource(resources).
		AddSkillSource(skills)

	result := builder.Build()

	if len(result) != 3 {
		t.Errorf("expected 3 contexts, got %d", len(result))
	}

	// Test deduplication
	builder2 := NewContextBuilder().
		AddMemorySource(memories).
		AddMemorySource(memories)

	result2 := builder2.Build()
	if len(result2) != 1 {
		t.Errorf("expected 1 context after dedup, got %d", len(result2))
	}
}

func TestCompression(t *testing.T) {
	original := "This is a long test string that should be compressed. " +
		"It has many repeated words and phrases to demonstrate compression. " +
		"Compression should reduce the size of the text."

	compressed := CompressText(original)

	if compressed == original {
		t.Error("compressed text should be different from original")
	}

	decompressed, err := DecompressText(compressed)
	if err != nil {
		t.Errorf("decompression failed: %v", err)
	}

	if decompressed != original {
		t.Error("decompressed text does not match original")
	}
}

func TestPersistence(t *testing.T) {
	tc := NewTieredContext()

	ctx := NewContext("viking://test/persist")
	ctx.Abstract = "Test context for persistence"
	ctx.Tier = TierL0
	tc.Add(ctx)

	handler := NewPersistenceHandler(&PersistenceConfig{
		StoragePath: "/tmp/goviking_test",
	}, tc, "test-session")

	// Save
	if err := handler.Save(); err != nil {
		t.Errorf("save failed: %v", err)
	}

	// Check exists
	if !handler.Exists() {
		t.Error("persisted file should exist")
	}

	// Load into new tiered context
	tc2 := NewTieredContext()
	handler2 := NewPersistenceHandler(&PersistenceConfig{
		StoragePath: "/tmp/goviking_test",
	}, tc2, "test-session")

	if err := handler2.Load(); err != nil {
		t.Errorf("load failed: %v", err)
	}

	if tc2.Count() != 1 {
		t.Errorf("expected 1 context after load, got %d", tc2.Count())
	}

	// Clean up
	handler.Delete()
}

func TestWindowManagement(t *testing.T) {
	tc := NewTieredContext()

	// Add multiple contexts
	for i := 0; i < 5; i++ {
		ctx := NewContext(fmt.Sprintf("viking://test/ctx%d", i))
		ctx.Abstract = "This is test context number " + fmt.Sprintf("%d with some text content.", i)
		ctx.Tier = ContextTier(i % 3)
		tc.Add(ctx)
	}

	config := DefaultContextWindowConfig()
	config.MaxTokens = 100
	window := NewContextWindow(config, tc, NewSimpleTokenCounter())

	info := window.GetWindowInfo()
	fmt.Printf("Window info: %s\n", info)

	if info.ApproachingLimit {
		fmt.Println("Context approaching limit - optimizing")
		optimized, err := window.OptimizeWindow()
		if err != nil {
			t.Errorf("optimization failed: %v", err)
		}
		fmt.Printf("Optimized to %d contexts\n", len(optimized))
	}
}

func TestBuildingTree(t *testing.T) {
	tree := NewBuildingTree()

	// Add root
	root := NewContext("viking://agent/skills")
	root.IsLeaf = false
	tree.AddContext(root)

	// Add child
	child := NewContext("viking://agent/skills/bash")
	child.ParentURI = "viking://agent/skills"
	child.IsLeaf = true
	tree.AddContext(child)

	if tree.Len() != 2 {
		t.Errorf("expected 2 contexts, got %d", tree.Len())
	}

	parent := tree.Parent("viking://agent/skills/bash")
	if parent == nil {
		t.Error("expected parent to be found")
	}

	children := tree.GetChildren("viking://agent/skills")
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}

	path := tree.GetPathToRoot("viking://agent/skills/bash")
	if len(path) != 2 {
		t.Errorf("expected path length 2, got %d", len(path))
	}
}

func TestContextToMap(t *testing.T) {
	now := time.Now().UTC()
	ctx := &Context{
		ID:          "test-id",
		URI:         "viking://test/uri",
		Abstract:    "Test abstract",
		ContextType: ContextTypeSkill,
		CreatedAt:   now,
		UpdatedAt:   now,
		ActiveCount: 5,
		Meta:        map[string]any{"key": "value"},
	}

	m := ctx.ToMap()

	if m["id"] != "test-id" {
		t.Errorf("expected id test-id, got %v", m["id"])
	}

	if m["uri"] != "viking://test/uri" {
		t.Errorf("expected uri viking://test/uri, got %v", m["uri"])
	}
}

func TestFromMap(t *testing.T) {
	now := time.Now().UTC()
	data := map[string]any{
		"id":           "test-id",
		"uri":          "viking://test/uri",
		"abstract":     "Test abstract",
		"context_type": "skill",
		"created_at":   now.Format(time.RFC3339),
		"updated_at":   now.Format(time.RFC3339),
		"active_count": int64(5),
	}

	ctx := FromMap(data)

	if ctx.ID != "test-id" {
		t.Errorf("expected id test-id, got %s", ctx.ID)
	}

	if ctx.ContextType != ContextTypeSkill {
		t.Errorf("expected context type skill, got %s", ctx.ContextType)
	}
}

func ExampleContext() {
	ctx := NewContext("viking://agent/skills/bash")
	ctx.Abstract = "Execute shell commands"
	ctx.Meta = map[string]any{
		"name":        "bash",
		"description": "Execute shell commands in a terminal",
	}

	fmt.Printf("Context: %s\n", ctx.URI)
	fmt.Printf("Type: %s\n", ctx.ContextType)
	fmt.Printf("Abstract: %s\n", ctx.Abstract)

	// Output:
	// Context: viking://agent/skills/bash
	// Type: skill
	// Abstract: Execute shell commands
}

func ExampleTieredContext() {
	tc := NewTieredContext()

	ctx1 := NewContext("viking://test/l0")
	ctx1.Tier = TierL0
	ctx1.Abstract = "Essential context"

	ctx2 := NewContext("viking://test/l1")
	ctx2.Tier = TierL1
	ctx2.Abstract = "On-demand context"

	tc.Add(ctx1)
	tc.Add(ctx2)

	fmt.Printf("Total contexts: %d\n", tc.Count())
	fmt.Printf("L0 count: %d\n", tc.CountByTier(TierL0))
	fmt.Printf("L1 count: %d\n", tc.CountByTier(TierL1))

	// Output:
	// Total contexts: 2
	// L0 count: 1
	// L1 count: 1
}
