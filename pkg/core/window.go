// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"sort"
	"sync"
)

// ContextWindowConfig holds configuration for context window management.
type ContextWindowConfig struct {
	MaxTokens         int
	MinL0Retention    int      // Minimum L0 contexts to always keep
	CompressionRatio  float64  // Ratio to compress when approaching limit
	PriorityTiers     []ContextTier // Tier priority order
}

// DefaultContextWindowConfig returns a default configuration.
func DefaultContextWindowConfig() *ContextWindowConfig {
	return &ContextWindowConfig{
		MaxTokens:      128000,
		MinL0Retention: 1,
		CompressionRatio: 0.9,
		PriorityTiers: []ContextTier{TierL0, TierL1, TierL2},
	}
}

// ContextWindow manages context within token limits.
type ContextWindow struct {
	config   *ContextWindowConfig
	tc       *TieredContext
	tokenCnt TokenCounter
	mu       sync.RWMutex
}

// NewContextWindow creates a new ContextWindow.
func NewContextWindow(config *ContextWindowConfig, tc *TieredContext, tokenCnt TokenCounter) *ContextWindow {
	if config == nil {
		config = DefaultContextWindowConfig()
	}
	if tokenCnt == nil {
		tokenCnt = NewSimpleTokenCounter()
	}
	return &ContextWindow{
		config:   config,
		tc:       tc,
		tokenCnt: tokenCnt,
	}
}

// CurrentTokens returns the current token count.
func (w *ContextWindow) CurrentTokens() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	contexts := w.tc.GetAll()
	total := 0
	for _, ctx := range contexts {
		total += w.tokenCnt.CountTokens(ctx.Abstract)
	}
	return total
}

// CurrentTokensByTier returns token count for each tier.
func (w *ContextWindow) CurrentTokensByTier() map[ContextTier]int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[ContextTier]int)
	for _, tier := range []ContextTier{TierL0, TierL1, TierL2} {
		contexts := w.tc.GetByTier(tier)
		total := 0
		for _, ctx := range contexts {
			total += w.tokenCnt.CountTokens(ctx.Abstract)
		}
		result[tier] = total
	}
	return result
}

// WithinLimit checks if current tokens are within the limit.
func (w *ContextWindow) WithinLimit() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentTokensUnsafe() <= w.config.MaxTokens
}

// ApproachingLimit checks if tokens are approaching the limit.
func (w *ContextWindow) ApproachingLimit() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	threshold := int(float64(w.config.MaxTokens) * w.config.CompressionRatio)
	return w.currentTokensUnsafe() >= threshold
}

// FitInWindow fits contexts into the window by priority.
func (w *ContextWindow) FitInWindow() ([]*Context, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var prioritized []*Context

	// First, add all L0 contexts (always included)
	l0Ctxs := w.tc.GetL0()
	for _, ctx := range l0Ctxs {
		prioritized = append(prioritized, ctx)
	}

	// Then add L1 contexts
	l1Ctxs := w.tc.GetL1()
	for _, ctx := range l1Ctxs {
		prioritized = append(prioritized, ctx)
	}

	// Then add L2 contexts
	l2Ctxs := w.tc.GetL2()
	for _, ctx := range l2Ctxs {
		prioritized = append(prioritized, ctx)
	}

	// Calculate current tokens
	currentTokens := 0
	for _, ctx := range prioritized {
		currentTokens += w.tokenCnt.CountTokens(ctx.Abstract)
	}

	// If within limit, return all
	if currentTokens <= w.config.MaxTokens {
		return prioritized, nil
	}

	// Need to trim - start from L2, then L1
	var result []*Context
	currentTokens = 0

	for _, ctx := range prioritized {
		tokens := w.tokenCnt.CountTokens(ctx.Abstract)
		if currentTokens+tokens <= w.config.MaxTokens {
			result = append(result, ctx)
			currentTokens += tokens
		} else if ctx.Tier == TierL0 && len(result) < w.config.MinL0Retention {
			// Always keep minimum L0 contexts
			result = append(result, ctx)
			currentTokens += tokens
		}
	}

	return result, nil
}

// Compress reduces context size using compression.
func (w *ContextWindow) Compress() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	compressed := 0

	// Compress L2 first, then L1
	tiers := []ContextTier{TierL2, TierL1}

	for _, tier := range tiers {
		contexts := w.tc.GetByTier(tier)
		for _, ctx := range contexts {
			originalLen := len(ctx.Abstract)
			compressedText := CompressText(ctx.Abstract)
			ctx.Abstract = compressedText
			compressed += originalLen - len(compressedText)
		}
	}

	return compressed, nil
}

// AddContext adds a context to the window.
func (w *ContextWindow) AddContext(ctx *Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if adding would exceed limit
	currentTokens := w.currentTokensUnsafe()
	newTokens := w.tokenCnt.CountTokens(ctx.Abstract)

	if currentTokens+newTokens > w.config.MaxTokens {
		return fmt.Errorf("context would exceed window limit: current=%d new=%d max=%d",
			currentTokens, newTokens, w.config.MaxTokens)
	}

	w.tc.Add(ctx)
	return nil
}

// RemoveContext removes a context from the window.
func (w *ContextWindow) RemoveContext(uri string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.tc.Remove(uri)
}

// GetWindowInfo returns information about the current window state.
func (w *ContextWindow) GetWindowInfo() *WindowInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()

	info := &WindowInfo{
		MaxTokens:    w.config.MaxTokens,
		CurrentTotal: w.currentTokensUnsafe(),
		TierCounts:   make(map[ContextTier]int),
		TierTokens:   make(map[ContextTier]int),
	}

	for _, tier := range []ContextTier{TierL0, TierL1, TierL2} {
		contexts := w.tc.GetByTier(tier)
		info.TierCounts[tier] = len(contexts)
		info.TierTokens[tier] = 0
		for _, ctx := range contexts {
			info.TierTokens[tier] += w.tokenCnt.CountTokens(ctx.Abstract)
		}
	}

	info.UsagePercent = float64(info.CurrentTotal) / float64(info.MaxTokens) * 100
	info.ApproachingLimit = info.UsagePercent >= w.config.CompressionRatio*100

	return info
}

func (w *ContextWindow) currentTokensUnsafe() int {
	contexts := w.tc.GetAll()
	total := 0
	for _, ctx := range contexts {
		total += w.tokenCnt.CountTokens(ctx.Abstract)
	}
	return total
}

// WindowInfo holds information about the context window.
type WindowInfo struct {
	MaxTokens         int              `json:"max_tokens"`
	CurrentTotal      int              `json:"current_total"`
	UsagePercent     float64          `json:"usage_percent"`
	ApproachingLimit bool             `json:"approaching_limit"`
	TierCounts       map[ContextTier]int `json:"tier_counts"`
	TierTokens       map[ContextTier]int `json:"tier_tokens"`
}

// String returns a string representation of WindowInfo.
func (wi *WindowInfo) String() string {
	return fmt.Sprintf("WindowInfo{Max: %d, Current: %d (%.1f%%), Approaching: %v, L0: %d/%d, L1: %d/%d, L2: %d/%d}",
		wi.MaxTokens, wi.CurrentTotal, wi.UsagePercent, wi.ApproachingLimit,
		wi.TierTokens[TierL0], wi.TierCounts[TierL0],
		wi.TierTokens[TierL1], wi.TierCounts[TierL1],
		wi.TierTokens[TierL2], wi.TierCounts[TierL2])
}

// OptimizeWindow optimizes the window by rebalancing contexts.
func (w *ContextWindow) OptimizeWindow() ([]*Context, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Sort contexts by activity within each tier
	sortL0ByActivity(w.tc.L0)
	sortL1ByActivity(w.tc.L1)
	sortL2ByActivity(w.tc.L2)

	return w.fitInWindowUnsafe()
}

func sortL0ByActivity(ctxs []*Context) {
	sort.Slice(ctxs, func(i, j int) bool {
		return ctxs[i].ActiveCount > ctxs[j].ActiveCount
	})
}

func sortL1ByActivity(ctxs []*Context) {
	sort.Slice(ctxs, func(i, j int) bool {
		return ctxs[i].ActiveCount > ctxs[j].ActiveCount
	})
}

func sortL2ByActivity(ctxs []*Context) {
	sort.Slice(ctxs, func(i, j int) bool {
		return ctxs[i].ActiveCount > ctxs[j].ActiveCount
	})
}

func (w *ContextWindow) fitInWindowUnsafe() ([]*Context, error) {
	var result []*Context
	currentTokens := 0

	// Add L0 contexts
	for _, ctx := range w.tc.L0 {
		tokens := w.tokenCnt.CountTokens(ctx.Abstract)
		result = append(result, ctx)
		currentTokens += tokens
	}

	// Add L1 contexts if space permits
	for _, ctx := range w.tc.L1 {
		tokens := w.tokenCnt.CountTokens(ctx.Abstract)
		if currentTokens+tokens <= w.config.MaxTokens {
			result = append(result, ctx)
			currentTokens += tokens
		}
	}

	// Add L2 contexts if space permits
	for _, ctx := range w.tc.L2 {
		tokens := w.tokenCnt.CountTokens(ctx.Abstract)
		if currentTokens+tokens <= w.config.MaxTokens {
			result = append(result, ctx)
			currentTokens += tokens
		}
	}

	return result, nil
}
