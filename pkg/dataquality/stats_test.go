// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

import (
	"math"
	"sort"
	"testing"
)

const tol = 1e-9

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		{"integers", []float64{1, 2, 3, 4, 5}, 3.0},
		{"single", []float64{7.0}, 7.0},
		{"negative", []float64{-3, -1, 1, 3}, 0.0},
		{"floats", []float64{1.5, 2.5, 3.5}, 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMean(tt.input)
			if !almostEqual(got, tt.expect, tol) {
				t.Errorf("calculateMean(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		{"odd count sorted", []float64{1, 2, 3, 4, 5}, 3.0},
		{"even count sorted", []float64{1, 2, 3, 4}, 2.5},
		{"single", []float64{42.0}, 42.0},
		{"two elements", []float64{3, 7}, 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// calculateMedian requires a sorted slice
			sorted := make([]float64, len(tt.input))
			copy(sorted, tt.input)
			sort.Float64s(sorted)
			got := calculateMedian(sorted)
			if !almostEqual(got, tt.expect, tol) {
				t.Errorf("calculateMedian(%v) = %v, want %v", sorted, got, tt.expect)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		// Population std dev: sqrt(2) for {1,2,3,4,5}, mean=3
		{"five elements", []float64{1, 2, 3, 4, 5}, math.Sqrt(2)},
		{"single element", []float64{5.0}, 0.0},
		{"constant", []float64{3, 3, 3, 3}, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean := calculateMean(tt.input)
			got := calculateStdDev(tt.input, mean)
			if !almostEqual(got, tt.expect, 1e-9) {
				t.Errorf("calculateStdDev(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	// Using sorted [1,2,3,4,5,6,7,8,9,10]
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		name       string
		percentile float64
		expect     float64
	}{
		{"P0 (min)", 0, 1.0},
		{"P100 (max)", 100, 10.0},
		{"P50 (median)", 50, 5.5},
		{"P25", 25, 3.25},
		{"P75", 75, 7.75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePercentile(values, tt.percentile)
			if !almostEqual(got, tt.expect, 1e-9) {
				t.Errorf("calculatePercentile(P%.0f) = %v, want %v", tt.percentile, got, tt.expect)
			}
		})
	}
}

func TestCalculateSkewness(t *testing.T) {
	t.Run("symmetric distribution is near zero", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
		mean := calculateMean(values)
		stdDev := calculateStdDev(values, mean)
		skew := calculateSkewness(values, mean, stdDev)
		if math.Abs(skew) > 0.1 {
			t.Errorf("symmetric distribution skewness should be ~0, got %v", skew)
		}
	})

	t.Run("zero stddev returns zero", func(t *testing.T) {
		values := []float64{5, 5, 5, 5}
		mean := calculateMean(values)
		stdDev := calculateStdDev(values, mean) // = 0
		skew := calculateSkewness(values, mean, stdDev)
		if skew != 0 {
			t.Errorf("zero stdDev should give 0 skewness, got %v", skew)
		}
	})
}

func TestCalculateKurtosis(t *testing.T) {
	t.Run("zero stddev returns zero", func(t *testing.T) {
		values := []float64{3, 3, 3, 3, 3}
		mean := calculateMean(values)
		stdDev := calculateStdDev(values, mean)
		kurt := calculateKurtosis(values, mean, stdDev)
		if kurt != 0 {
			t.Errorf("zero stdDev should give 0 kurtosis, got %v", kurt)
		}
	})

	t.Run("large sample normal-ish data is near zero excess kurtosis", func(t *testing.T) {
		// Values drawn from a nearly normal distribution
		values := []float64{
			-2.3, -1.9, -1.5, -1.2, -0.9, -0.6, -0.4, -0.2, 0.0,
			0.2, 0.4, 0.6, 0.9, 1.2, 1.5, 1.9, 2.3,
		}
		mean := calculateMean(values)
		stdDev := calculateStdDev(values, mean)
		kurt := calculateKurtosis(values, mean, stdDev)
		// Excess kurtosis for a normal distribution ≈ 0; allow ±3
		if math.Abs(kurt) > 3.0 {
			t.Errorf("near-normal kurtosis should be close to 0, got %v", kurt)
		}
	})
}

func TestCountUnique(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect int
	}{
		{"all same", []float64{1, 1, 1}, 1},
		{"all different", []float64{1, 2, 3, 4}, 4},
		{"some duplicates", []float64{1, 2, 2, 3, 3, 3}, 3},
		{"empty", []float64{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countUnique(tt.input); got != tt.expect {
				t.Errorf("countUnique(%v) = %d, want %d", tt.input, got, tt.expect)
			}
		})
	}
}
