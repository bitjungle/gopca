// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package types

import (
	"math"
	"testing"
)

func TestConvertFloat64SliceToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []JSONFloat64
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []float64{},
			expected: []JSONFloat64{},
		},
		{
			name:     "normal values",
			input:    []float64{1.0, 2.5, 3.7},
			expected: []JSONFloat64{1.0, 2.5, 3.7},
		},
		{
			name:     "with NaN and Inf",
			input:    []float64{1.0, math.NaN(), math.Inf(1), math.Inf(-1)},
			expected: []JSONFloat64{1.0, JSONFloat64(math.NaN()), JSONFloat64(math.Inf(1)), JSONFloat64(math.Inf(-1))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFloat64SliceToJSON(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Length mismatch: expected %d, got %d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				// Special handling for NaN
				if math.IsNaN(float64(tt.expected[i])) {
					if !math.IsNaN(float64(result[i])) {
						t.Errorf("Index %d: expected NaN, got %v", i, result[i])
					}
				} else if result[i] != tt.expected[i] {
					t.Errorf("Index %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestConvertFloat64MatrixToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]float64
		expected [][]JSONFloat64
	}{
		{
			name:     "nil matrix",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty matrix",
			input:    [][]float64{},
			expected: [][]JSONFloat64{},
		},
		{
			name:     "normal matrix",
			input:    [][]float64{{1.0, 2.0}, {3.0, 4.0}},
			expected: [][]JSONFloat64{{1.0, 2.0}, {3.0, 4.0}},
		},
		{
			name:     "with NaN",
			input:    [][]float64{{1.0, math.NaN()}, {math.Inf(1), 4.0}},
			expected: [][]JSONFloat64{{1.0, JSONFloat64(math.NaN())}, {JSONFloat64(math.Inf(1)), 4.0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFloat64MatrixToJSON(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Row count mismatch: expected %d, got %d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				if len(result[i]) != len(tt.expected[i]) {
					t.Errorf("Row %d length mismatch: expected %d, got %d", i, len(tt.expected[i]), len(result[i]))
					continue
				}

				for j := range result[i] {
					// Special handling for NaN
					if math.IsNaN(float64(tt.expected[i][j])) {
						if !math.IsNaN(float64(result[i][j])) {
							t.Errorf("Position [%d][%d]: expected NaN, got %v", i, j, result[i][j])
						}
					} else if result[i][j] != tt.expected[i][j] {
						t.Errorf("Position [%d][%d]: expected %v, got %v", i, j, tt.expected[i][j], result[i][j])
					}
				}
			}
		})
	}
}

func TestConvertFloat64MapToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input map[string][]float64
	}{
		{
			name:  "nil map",
			input: nil,
		},
		{
			name:  "empty map",
			input: map[string][]float64{},
		},
		{
			name: "normal map",
			input: map[string][]float64{
				"col1": {1.0, 2.0, 3.0},
				"col2": {4.0, 5.0, 6.0},
			},
		},
		{
			name: "with NaN and Inf",
			input: map[string][]float64{
				"col1": {1.0, math.NaN(), math.Inf(1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFloat64MapToJSON(tt.input)

			if tt.input == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.input) {
				t.Errorf("Map size mismatch: expected %d, got %d", len(tt.input), len(result))
				return
			}

			for key, expectedValues := range tt.input {
				resultValues, exists := result[key]
				if !exists {
					t.Errorf("Key %s missing in result", key)
					continue
				}

				if len(resultValues) != len(expectedValues) {
					t.Errorf("Key %s: length mismatch, expected %d, got %d", key, len(expectedValues), len(resultValues))
					continue
				}

				for i := range resultValues {
					if math.IsNaN(expectedValues[i]) {
						if !math.IsNaN(float64(resultValues[i])) {
							t.Errorf("Key %s[%d]: expected NaN, got %v", key, i, resultValues[i])
						}
					} else if float64(resultValues[i]) != expectedValues[i] {
						t.Errorf("Key %s[%d]: expected %v, got %v", key, i, expectedValues[i], resultValues[i])
					}
				}
			}
		})
	}
}

func TestConvertFloat64ParamsMapToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]float64
	}{
		{
			name:  "nil map",
			input: nil,
		},
		{
			name:  "empty map",
			input: map[string]float64{},
		},
		{
			name: "normal params",
			input: map[string]float64{
				"gamma":  0.5,
				"degree": 3.0,
			},
		},
		{
			name: "with special values",
			input: map[string]float64{
				"value":  1.0,
				"nanval": math.NaN(),
				"infval": math.Inf(1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFloat64ParamsMapToJSON(tt.input)

			if tt.input == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.input) {
				t.Errorf("Map size mismatch: expected %d, got %d", len(tt.input), len(result))
				return
			}

			for key, expectedVal := range tt.input {
				resultVal, exists := result[key]
				if !exists {
					t.Errorf("Key %s missing in result", key)
					continue
				}

				if math.IsNaN(expectedVal) {
					if !math.IsNaN(float64(resultVal)) {
						t.Errorf("Key %s: expected NaN, got %v", key, resultVal)
					}
				} else if float64(resultVal) != expectedVal {
					t.Errorf("Key %s: expected %v, got %v", key, expectedVal, resultVal)
				}
			}
		})
	}
}
