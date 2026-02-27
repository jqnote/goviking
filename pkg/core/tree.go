// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"sync"
)

// TreeNode represents a node in the context tree.
type TreeNode struct {
	Context *Context
	Children []*TreeNode
}

// BuildingTree is a container for built context trees.
type BuildingTree struct {
	sourcePath string
	sourceFormat string

	contexts []*Context
	uriMap   map[string]*Context
	rootURI  string
	mu       sync.RWMutex
}

// NewBuildingTree creates a new BuildingTree.
func NewBuildingTree() *BuildingTree {
	return &BuildingTree{
		contexts: []*Context{},
		uriMap:   make(map[string]*Context),
	}
}

// SetSourcePath sets the source path.
func (bt *BuildingTree) SetSourcePath(path string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.sourcePath = path
}

// SetSourceFormat sets the source format.
func (bt *BuildingTree) SetSourceFormat(format string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.sourceFormat = format
}

// AddContext adds a context to the tree.
func (bt *BuildingTree) AddContext(ctx *Context) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.contexts = append(bt.contexts, ctx)
	bt.uriMap[ctx.URI] = ctx

	// Set root if this is the first context or it's a root-level context
	if ctx.ParentURI == "" {
		if bt.rootURI == "" {
			bt.rootURI = ctx.URI
		}
	}
}

// Root returns the root context.
func (bt *BuildingTree) Root() *Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.rootURI == "" {
		return nil
	}
	return bt.uriMap[bt.rootURI]
}

// Contexts returns all contexts.
func (bt *BuildingTree) Contexts() []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	result := make([]*Context, len(bt.contexts))
	copy(result, bt.contexts)
	return result
}

// Get returns a context by URI.
func (bt *BuildingTree) Get(uri string) *Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return bt.uriMap[uri]
}

// Parent returns the parent context of a URI.
func (bt *BuildingTree) Parent(uri string) *Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	ctx := bt.uriMap[uri]
	if ctx == nil || ctx.ParentURI == "" {
		return nil
	}
	return bt.uriMap[ctx.ParentURI]
}

// GetChildren returns children of a URI.
func (bt *BuildingTree) GetChildren(uri string) []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var result []*Context
	for _, ctx := range bt.contexts {
		if ctx.ParentURI == uri {
			result = append(result, ctx)
		}
	}
	return result
}

// GetPathToRoot returns the path from context to root.
func (bt *BuildingTree) GetPathToRoot(uri string) []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var path []*Context
	currentURI := uri

	for currentURI != "" {
		ctx := bt.uriMap[currentURI]
		if ctx == nil {
			break
		}
		path = append(path, ctx)
		currentURI = ctx.ParentURI
	}

	return path
}

// ToDirectoryStructure converts the tree to a directory-like structure.
func (bt *BuildingTree) ToDirectoryStructure() map[string]any {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.rootURI == "" {
		return nil
	}

	return bt.buildDir(bt.rootURI)
}

func (bt *BuildingTree) buildDir(uri string) map[string]any {
	ctx := bt.uriMap[uri]
	if ctx == nil {
		return nil
	}

	children := bt.GetChildren(uri)

	// Use semantic_title or source_title from meta, or fallback to name/uri
	title := ""
	if semanticTitle, ok := ctx.Meta["semantic_title"].(string); ok && semanticTitle != "" {
		title = semanticTitle
	} else if sourceTitle, ok := ctx.Meta["source_title"].(string); ok && sourceTitle != "" {
		title = sourceTitle
	} else if ctx.Meta != nil && ctx.Meta["name"] != nil {
		title = fmt.Sprintf("%v", ctx.Meta["name"])
	} else {
		title = "Untitled"
	}

	childrenList := make([]map[string]any, 0, len(children))
	for _, child := range children {
		if childDir := bt.buildDir(child.URI); childDir != nil {
			childrenList = append(childrenList, childDir)
		}
	}

	return map[string]any{
		"uri":      uri,
		"title":    title,
		"type":     string(ctx.ContextType),
		"children": childrenList,
	}
}

// Len returns the number of contexts.
func (bt *BuildingTree) Len() int {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return len(bt.contexts)
}

// GetContextsByType returns contexts filtered by type.
func (bt *BuildingTree) GetContextsByType(ctxType ContextType) []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var result []*Context
	for _, ctx := range bt.contexts {
		if ctx.ContextType == ctxType {
			result = append(result, ctx)
		}
	}
	return result
}

// GetContextsByCategory returns contexts filtered by category.
func (bt *BuildingTree) GetContextsByCategory(category Category) []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var result []*Context
	for _, ctx := range bt.contexts {
		if ctx.Category == category {
			result = append(result, ctx)
		}
	}
	return result
}

// GetLeafContexts returns only leaf contexts.
func (bt *BuildingTree) GetLeafContexts() []*Context {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var result []*Context
	for _, ctx := range bt.contexts {
		if ctx.IsLeaf {
			result = append(result, ctx)
		}
	}
	return result
}

// BuildTreeFromContexts builds a tree structure from a flat list of contexts.
func BuildTreeFromContexts(contexts []*Context) *BuildingTree {
	tree := NewBuildingTree()

	// First pass: add all contexts
	for _, ctx := range contexts {
		tree.AddContext(ctx)
	}

	// Second pass: find root (context with no parent)
	for _, ctx := range contexts {
		if ctx.ParentURI == "" {
			tree.rootURI = ctx.URI
			break
		}
	}

	return tree
}
