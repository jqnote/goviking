// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"math"
	"time"
)

// HotnessConfig holds configuration for hotness scoring.
type HotnessConfig struct {
	Alpha          float64       // Weight for hotness (0-1), default 0.2
	HalfLifeDays   float64       // Half-life for time decay in days, default 7
}

// DefaultHotnessConfig returns default hotness configuration.
func DefaultHotnessConfig() HotnessConfig {
	return HotnessConfig{
		Alpha:        0.2,
		HalfLifeDays: 7,
	}
}

// HotnessScorer calculates hotness scores for contexts.
type HotnessScorer struct {
	config HotnessConfig
}

// NewHotnessScorer creates a new hotness scorer.
func NewHotnessScorer(config HotnessConfig) *HotnessScorer {
	if config.Alpha == 0 {
		config.Alpha = 0.2
	}
	if config.HalfLifeDays == 0 {
		config.HalfLifeDays = 7
	}
	return &HotnessScorer{
		config: config,
	}
}

// CalculateHotness calculates hotness score based on access count and last access time.
// Returns value between 0 and 1.
func (h *HotnessScorer) CalculateHotness(accessCount int, lastAccess time.Time) float64 {
	// Frequency component: use sigmoid function on log1p of access count
	frequencyScore := h.sigmoid(float64(accessCount))

	// Recency component: exponential decay based on time since last access
	recencyScore := h.exponentialDecay(lastAccess)

	// Combine frequency and recency (equal weight)
	hotness := (frequencyScore + recencyScore) / 2

	return hotness
}

// sigmoid applies sigmoid function for smooth frequency scoring.
func (h *HotnessScorer) sigmoid(x float64) float64 {
	// Use sigmoid: 1 / (1 + exp(-x))
	// Shift x to make it more useful (x-3 means counts < 20 give low scores)
	return 1 / (1 + math.Exp(-(x - 3)))
}

// exponentialDecay calculates exponential decay based on time since last access.
func (h *HotnessScorer) exponentialDecay(lastAccess time.Time) float64 {
	// Calculate hours since last access
	hoursSince := time.Since(lastAccess).Hours()
	if hoursSince < 0 {
		hoursSince = 0
	}

	// Convert half-life to hours
	halfLifeHours := h.config.HalfLifeDays * 24

	// Exponential decay: exp(-ln(2) * hours / halfLife)
	// At half-life, score = 0.5
	decay := math.Exp(-math.Ln2 * hoursSince / halfLifeHours)

	// Clamp to 0-1
	if decay > 1 {
		decay = 1
	}
	if decay < 0 {
		decay = 0
	}

	return decay
}

// HybridScore combines semantic similarity with hotness score.
func (h *HotnessScorer) HybridScore(semanticScore, hotnessScore float64) float64 {
	// final = (1 - alpha) * semantic + alpha * hotness
	return (1-h.config.Alpha)*semanticScore + h.config.Alpha*hotnessScore
}

// CombineScores combines semantic and hotness scores with custom alpha.
func CombineScores(semanticScore, hotnessScore float64, alpha float64) float64 {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	return (1-alpha)*semanticScore + alpha*hotnessScore
}

// ContextHotness represents hotness data for a context.
type ContextHotness struct {
	ContextID    string    `json:"context_id"`
	AccessCount int       `json:"access_count"`
	LastAccess   time.Time `json:"last_access"`
	HotnessScore float64  `json:"hotness_score"`
}

// UpdateHotness updates hotness score for a context.
func (h *HotnessScorer) UpdateHotness(ctx *ContextHotness) {
	ctx.HotnessScore = h.CalculateHotness(ctx.AccessCount, ctx.LastAccess)
}
