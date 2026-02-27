// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"strings"
	"sync"
)

// ContextSource represents a source of context.
type ContextSource interface {
	GetContexts() []*Context
	GetType() string
}

// MemorySource represents a memory context source.
type MemorySource struct {
	contexts []*Context
}

func NewMemorySource(contexts []*Context) *MemorySource {
	return &MemorySource{contexts: contexts}
}

func (s *MemorySource) GetContexts() []*Context {
	return s.contexts
}

func (s *MemorySource) GetType() string {
	return "memory"
}

// ResourceSource represents a resource context source.
type ResourceSource struct {
	contexts []*Context
}

func NewResourceSource(contexts []*Context) *ResourceSource {
	return &ResourceSource{contexts: contexts}
}

func (s *ResourceSource) GetContexts() []*Context {
	return s.contexts
}

func (s *ResourceSource) GetType() string {
	return "resource"
}

// SkillSource represents a skill context source.
type SkillSource struct {
	contexts []*Context
}

func NewSkillSource(contexts []*Context) *SkillSource {
	return &SkillSource{contexts: contexts}
}

func (s *SkillSource) GetContexts() []*Context {
	return s.contexts
}

func (s *SkillSource) GetType() string {
	return "skill"
}

// ContextBuilder builds context from multiple sources.
type ContextBuilder struct {
	sources []ContextSource
	mu      sync.RWMutex
}

// NewContextBuilder creates a new ContextBuilder.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		sources: []ContextSource{},
	}
}

// AddSource adds a context source to the builder.
func (b *ContextBuilder) AddSource(source ContextSource) *ContextBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sources = append(b.sources, source)
	return b
}

// AddMemorySource adds a memory source.
func (b *ContextBuilder) AddMemorySource(contexts []*Context) *ContextBuilder {
	return b.AddSource(NewMemorySource(contexts))
}

// AddResourceSource adds a resource source.
func (b *ContextBuilder) AddResourceSource(contexts []*Context) *ContextBuilder {
	return b.AddSource(NewResourceSource(contexts))
}

// AddSkillSource adds a skill source.
func (b *ContextBuilder) AddSkillSource(contexts []*Context) *ContextBuilder {
	return b.AddSource(NewSkillSource(contexts))
}

// AddTiered adds a tiered context as source.
func (b *ContextBuilder) AddTiered(tc *TieredContext) *ContextBuilder {
	return b.AddMemorySource(tc.GetAll())
}

// Build builds and returns the merged context.
func (b *ContextBuilder) Build() []*Context {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []*Context
	seen := make(map[string]bool)

	for _, source := range b.sources {
		for _, ctx := range source.GetContexts() {
			if !seen[ctx.URI] {
				seen[ctx.URI] = true
				result = append(result, ctx)
			}
		}
	}

	return result
}

// BuildTiered builds context organized by tier.
func (b *ContextBuilder) BuildTiered() *TieredContext {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tc := NewTieredContext()
	seen := make(map[string]bool)

	for _, source := range b.sources {
		for _, ctx := range source.GetContexts() {
			if !seen[ctx.URI] {
				seen[ctx.URI] = true
				tc.Add(ctx)
			}
		}
	}

	return tc
}

// BuildWithPrioritization builds context with tier prioritization.
func (b *ContextBuilder) BuildWithPrioritization(maxTokens int, tokenCounter TokenCounter) ([]*Context, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Collect all contexts
	var allItems []tieredItem
	seen := make(map[string]bool)

	for _, source := range b.sources {
		for i, ctx := range source.GetContexts() {
			if !seen[ctx.URI] {
				seen[ctx.URI] = true
				allItems = append(allItems, tieredItem{
					ctx:   ctx,
					tier:  ctx.Tier,
					order: i,
				})
			}
		}
	}

	// Sort by tier (L0 first), then by active count, then by order
	sortByPriority(allItems)

	var result []*Context
	currentTokens := 0

	for _, item := range allItems {
		tokens := tokenCounter.CountTokens(item.ctx.Abstract)
		if currentTokens+tokens > maxTokens {
			// Skip if adding would exceed limit, but always include L0
			if item.tier != TierL0 {
				continue
			}
		}
		result = append(result, item.ctx)
		currentTokens += tokens
	}

	return result, nil
}

func sortByPriority(items []tieredItem) {
	// Sort by tier first (L0 < L1 < L2), then by active count descending, then by order
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if compareTieredItem(items[i], items[j]) > 0 {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func compareTieredItem(a, b tieredItem) int {
	// First compare by tier
	if a.tier != b.tier {
		if a.tier < b.tier {
			return -1
		}
		return 1
	}
	// Then by active count (descending)
	if a.ctx.ActiveCount != b.ctx.ActiveCount {
		if a.ctx.ActiveCount > b.ctx.ActiveCount {
			return -1
		}
		return 1
	}
	// Finally by order
	if a.order < b.order {
		return -1
	}
	if a.order > b.order {
		return 1
	}
	return 0
}

// TokenCounter counts tokens in text.
type TokenCounter interface {
	CountTokens(text string) int
}

// tieredItem is used for sorting contexts by priority.
type tieredItem struct {
	ctx   *Context
	tier  ContextTier
	order int
}

// SimpleTokenCounter counts tokens using simple word-based estimation.
type SimpleTokenCounter struct {
	AverageTokenLength float64
}

// NewSimpleTokenCounter creates a SimpleTokenCounter.
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{AverageTokenLength: 4.0}
}

// CountTokens estimates token count using word count / average token length.
func (c *SimpleTokenCounter) CountTokens(text string) int {
	words := strings.Fields(text)
	if words == nil {
		return 0
	}
	// Rough estimate: ~4 characters per token
	return (len(text) + 3) / 4
}

// BuildString builds a context string for LLM input.
func (b *ContextBuilder) BuildString() string {
	contexts := b.Build()
	return FormatContextsForLLM(contexts)
}

// FormatContextsForLLM formats contexts for LLM input.
func FormatContextsForLLM(contexts []*Context) string {
	var sb strings.Builder

	// Group by type
	memoryCtxs := filterByType(contexts, ContextTypeMemory)
	resourceCtxs := filterByType(contexts, ContextTypeResource)
	skillCtxs := filterByType(contexts, ContextTypeSkill)

	if len(memoryCtxs) > 0 {
		sb.WriteString("## Memories\n\n")
		for _, ctx := range memoryCtxs {
			sb.WriteString(formatContextItem(ctx))
		}
		sb.WriteString("\n")
	}

	if len(resourceCtxs) > 0 {
		sb.WriteString("## Resources\n\n")
		for _, ctx := range resourceCtxs {
			sb.WriteString(formatContextItem(ctx))
		}
		sb.WriteString("\n")
	}

	if len(skillCtxs) > 0 {
		sb.WriteString("## Skills\n\n")
		for _, ctx := range skillCtxs {
			sb.WriteString(formatContextItem(ctx))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func filterByType(contexts []*Context, ctxType ContextType) []*Context {
	var result []*Context
	for _, ctx := range contexts {
		if ctx.ContextType == ctxType {
			result = append(result, ctx)
		}
	}
	return result
}

func formatContextItem(ctx *Context) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s\n", ctx.URI))
	if ctx.Abstract != "" {
		sb.WriteString(fmt.Sprintf("%s\n", ctx.Abstract))
	}
	if len(ctx.Meta) > 0 {
		for k, v := range ctx.Meta {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}
	return sb.String()
}
