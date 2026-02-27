// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package retrieval

import (
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		a        []float64
		b        []float64
		expected float64
	}{
		{[]float64{1, 0}, []float64{1, 0}, 1.0},
		{[]float64{1, 0}, []float64{0, 1}, 0.0},
		{[]float64{1, 1}, []float64{1, 1}, 1.0},
		{[]float64{1, 1}, []float64{1, -1}, 0.0},
		{[]float64{1, 2, 3}, []float64{4, 5, 6}, 0.974631846},
	}

	for _, tt := range tests {
		result := CosineSimilarity(tt.a, tt.b)
		if result < tt.expected-0.001 || result > tt.expected+0.001 {
			t.Errorf("CosineSimilarity(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestCosineSimilarityMismatchedLength(t *testing.T) {
	result := CosineSimilarity([]float64{1, 2}, []float64{1, 2, 3})
	if result != 0 {
		t.Error("Expected 0 for mismatched lengths")
	}
}

func TestEuclideanDistance(t *testing.T) {
	a := []float64{0, 0}
	b := []float64{3, 4}
	expected := 5.0

	result := EuclideanDistance(a, b)
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("EuclideanDistance(%v, %v) = %v, expected %v", a, b, result, expected)
	}
}

func TestDotProduct(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{4, 5, 6}
	expected := 32.0

	result := DotProduct(a, b)
	if result != expected {
		t.Errorf("DotProduct(%v, %v) = %v, expected %v", a, b, result, expected)
	}
}
