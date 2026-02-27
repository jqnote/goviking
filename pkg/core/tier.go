// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"sort"
	"sync"
)

// TieredContext holds contexts organized by tier.
type TieredContext struct {
	L0 []*Context // Essential context, always loaded
	L1 []*Context // Context loaded on demand
	L2 []*Context // Archive context, loaded when needed

	mu sync.RWMutex
}

// NewTieredContext creates a new TieredContext.
func NewTieredContext() *TieredContext {
	return &TieredContext{
		L0: []*Context{},
		L1: []*Context{},
		L2: []*Context{},
	}
}

// Add adds a context to the appropriate tier.
func (tc *TieredContext) Add(ctx *Context) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	switch ctx.Tier {
	case TierL0:
		tc.L0 = append(tc.L0, ctx)
	case TierL1:
		tc.L1 = append(tc.L1, ctx)
	case TierL2:
		tc.L2 = append(tc.L2, ctx)
	default:
		tc.L1 = append(tc.L1, ctx)
	}
}

// GetL0 returns all L0 contexts.
func (tc *TieredContext) GetL0() []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.L0
}

// GetL1 returns all L1 contexts.
func (tc *TieredContext) GetL1() []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.L1
}

// GetL2 returns all L2 contexts.
func (tc *TieredContext) GetL2() []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.L2
}

// GetAll returns all contexts from all tiers.
func (tc *TieredContext) GetAll() []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	result := make([]*Context, 0, len(tc.L0)+len(tc.L1)+len(tc.L2))
	result = append(result, tc.L0...)
	result = append(result, tc.L1...)
	result = append(result, tc.L2...)
	return result
}

// GetByTier returns contexts for a specific tier.
func (tc *TieredContext) GetByTier(tier ContextTier) []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	switch tier {
	case TierL0:
		return tc.L0
	case TierL1:
		return tc.L1
	case TierL2:
		return tc.L2
	}
	return nil
}

// GetByURI finds a context by URI across all tiers.
func (tc *TieredContext) GetByURI(uri string) *Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	for _, ctx := range tc.L0 {
		if ctx.URI == uri {
			return ctx
		}
	}
	for _, ctx := range tc.L1 {
		if ctx.URI == uri {
			return ctx
		}
	}
	for _, ctx := range tc.L2 {
		if ctx.URI == uri {
			return ctx
		}
	}
	return nil
}

// Remove removes a context by URI.
func (tc *TieredContext) Remove(uri string) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if removeFromSlice(&tc.L0, uri) {
		return true
	}
	if removeFromSlice(&tc.L1, uri) {
		return true
	}
	if removeFromSlice(&tc.L2, uri) {
		return true
	}
	return false
}

func removeFromSlice(slice *[]*Context, uri string) bool {
	for i, ctx := range *slice {
		if ctx.URI == uri {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return true
		}
	}
	return false
}

// MoveToTier moves a context to a different tier.
func (tc *TieredContext) MoveToTier(uri string, tier ContextTier) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Remove from current tier
	var ctx *Context
	if removeFromSlice(&tc.L0, uri) {
		ctx = findContextByURIInSlices([][]*Context{tc.L0}, uri)
		// Need to re-find - it was already removed
	} else if removeFromSlice(&tc.L1, uri) {
		// Already removed
	} else if removeFromSlice(&tc.L2, uri) {
		// Already removed
	} else {
		return false
	}

	// Actually find and remove the context
	if ctx == nil {
		ctx = findContextByURIInSlices([][]*Context{tc.L0, tc.L1, tc.L2}, uri)
		if ctx == nil {
			return false
		}
	}

	// Remove from all slices to be safe
	removeFromSlice(&tc.L0, uri)
	removeFromSlice(&tc.L1, uri)
	removeFromSlice(&tc.L2, uri)

	// Add to new tier
	ctx.Tier = tier
	switch tier {
	case TierL0:
		tc.L0 = append(tc.L0, ctx)
	case TierL1:
		tc.L1 = append(tc.L1, ctx)
	case TierL2:
		tc.L2 = append(tc.L2, ctx)
	}

	return true
}

func findContextByURIInSlices(slices [][]*Context, uri string) *Context {
	for _, slice := range slices {
		for _, ctx := range slice {
			if ctx.URI == uri {
				return ctx
			}
		}
	}
	return nil
}

// Count returns the total count of contexts.
func (tc *TieredContext) Count() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.L0) + len(tc.L1) + len(tc.L2)
}

// CountByTier returns the count for a specific tier.
func (tc *TieredContext) CountByTier(tier ContextTier) int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	switch tier {
	case TierL0:
		return len(tc.L0)
	case TierL1:
		return len(tc.L1)
	case TierL2:
		return len(tc.L2)
	}
	return 0
}

// GetContextsByType returns contexts filtered by context type.
func (tc *TieredContext) GetContextsByType(ctxType ContextType) []*Context {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var result []*Context
	for _, ctx := range tc.L0 {
		if ctx.ContextType == ctxType {
			result = append(result, ctx)
		}
	}
	for _, ctx := range tc.L1 {
		if ctx.ContextType == ctxType {
			result = append(result, ctx)
		}
	}
	for _, ctx := range tc.L2 {
		if ctx.ContextType == ctxType {
			result = append(result, ctx)
		}
	}
	return result
}

// SortByActivity sorts contexts by active count (descending).
func (tc *TieredContext) SortByActivity() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	sort.Slice(tc.L0, func(i, j int) bool {
		return tc.L0[i].ActiveCount > tc.L0[j].ActiveCount
	})
	sort.Slice(tc.L1, func(i, j int) bool {
		return tc.L1[i].ActiveCount > tc.L1[j].ActiveCount
	})
	sort.Slice(tc.L2, func(i, j int) bool {
		return tc.L2[i].ActiveCount > tc.L2[j].ActiveCount
	})
}

// String returns a string representation of the tiered context.
func (tc *TieredContext) String() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return fmt.Sprintf("TieredContext{L0: %d, L1: %d, L2: %d}", len(tc.L0), len(tc.L1), len(tc.L2))
}

// TierLoader defines an interface for loading contexts from storage.
type TierLoader interface {
	LoadTier(tier ContextTier, sessionID string) ([]*Context, error)
	LoadAll(sessionID string) (*TieredContext, error)
}

// TierLoaderFunc is a function type that implements TierLoader.
type TierLoaderFunc func(tier ContextTier, sessionID string) ([]*Context, error)

func (f TierLoaderFunc) LoadTier(tier ContextTier, sessionID string) ([]*Context, error) {
	return f(tier, sessionID)
}
