// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package types

// ConvertFloat64SliceToJSON converts a slice of float64 to JSONFloat64.
// This is useful for safely serializing float64 slices that may contain
// NaN or Inf values to JSON.
func ConvertFloat64SliceToJSON(values []float64) []JSONFloat64 {
	if values == nil {
		return nil
	}
	result := make([]JSONFloat64, len(values))
	for i, val := range values {
		result[i] = JSONFloat64(val)
	}
	return result
}

// ConvertFloat64MatrixToJSON converts a 2D slice of float64 to JSONFloat64.
// This is useful for safely serializing matrices that may contain
// NaN or Inf values to JSON.
func ConvertFloat64MatrixToJSON(matrix [][]float64) [][]JSONFloat64 {
	if matrix == nil {
		return nil
	}
	result := make([][]JSONFloat64, len(matrix))
	for i, row := range matrix {
		result[i] = make([]JSONFloat64, len(row))
		for j, val := range row {
			result[i][j] = JSONFloat64(val)
		}
	}
	return result
}

// ConvertFloat64MapToJSON converts a map of string keys to float64 slices
// into a map with JSONFloat64 slices.
// This is useful for safely serializing maps containing numeric data to JSON.
func ConvertFloat64MapToJSON(data map[string][]float64) map[string][]JSONFloat64 {
	if data == nil {
		return nil
	}
	result := make(map[string][]JSONFloat64, len(data))
	for key, values := range data {
		result[key] = ConvertFloat64SliceToJSON(values)
	}
	return result
}

// ConvertFloat64ParamsMapToJSON converts a map of string keys to float64 values
// into a map with JSONFloat64 values.
// This is useful for safely serializing parameter maps to JSON.
func ConvertFloat64ParamsMapToJSON(params map[string]float64) map[string]JSONFloat64 {
	if params == nil {
		return nil
	}
	result := make(map[string]JSONFloat64, len(params))
	for key, val := range params {
		result[key] = JSONFloat64(val)
	}
	return result
}
