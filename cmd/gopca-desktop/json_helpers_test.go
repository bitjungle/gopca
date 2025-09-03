// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"math"
	"testing"
)

// TestToJSONSafe_EmptyData verifies handling of empty datasets
func TestToJSONSafe_EmptyData(t *testing.T) {
	fd := &FileData{
		Headers:  []string{"A", "B"},
		RowNames: []string{},
		Data:     [][]float64{},
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result for empty data")
	}
	if result.MissingMask != nil {
		t.Error("Expected nil MissingMask for empty data")
	}
}

// TestToJSONSafe_NoMissing verifies optimization for data without missing values
func TestToJSONSafe_NoMissing(t *testing.T) {
	fd := &FileData{
		Headers:  []string{"A", "B"},
		RowNames: []string{"R1", "R2"},
		Data: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.MissingMask != nil {
		t.Error("Expected nil MissingMask when no missing values - this is the key optimization")
	}
	if len(result.Data) != 2 || len(result.Data[0]) != 2 {
		t.Error("Data dimensions incorrect")
	}
}

// TestToJSONSafe_WithMissing verifies correct handling of missing values
func TestToJSONSafe_WithMissing(t *testing.T) {
	fd := &FileData{
		Headers:  []string{"A", "B", "C"},
		RowNames: []string{"R1", "R2"},
		Data: [][]float64{
			{1.0, math.NaN(), 3.0},
			{4.0, 5.0, math.NaN()},
		},
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.MissingMask == nil {
		t.Fatal("Expected non-nil MissingMask when missing values exist")
	}
	if !result.MissingMask[0][1] {
		t.Error("Expected MissingMask[0][1] to be true for NaN at position [0][1]")
	}
	if !result.MissingMask[1][2] {
		t.Error("Expected MissingMask[1][2] to be true for NaN at position [1][2]")
	}
	if result.MissingMask[0][0] || result.MissingMask[1][0] {
		t.Error("Expected non-missing values to be false in mask")
	}
}

// TestToJSONSafe_LargeDatasetNoMissing tests performance optimization effectiveness
func TestToJSONSafe_LargeDatasetNoMissing(t *testing.T) {
	// Simulate a large dataset like MET (1000 rows × 100 columns)
	rows, cols := 1000, 100
	data := make([][]float64, rows)
	rowNames := make([]string, rows)
	headers := make([]string, cols)
	
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		rowNames[i] = "Row" + string(rune(i))
		for j := 0; j < cols; j++ {
			data[i][j] = float64(i*cols + j)
		}
	}
	
	for j := 0; j < cols; j++ {
		headers[j] = "Col" + string(rune(j))
	}
	
	fd := &FileData{
		Headers:  headers,
		RowNames: rowNames,
		Data:     data,
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.MissingMask != nil {
		t.Error("PERFORMANCE: Expected nil MissingMask for large dataset without missing values")
		t.Error("This is the key optimization - avoiding allocation of 100,000 booleans")
	}
}

// TestToJSONSafe_EarlyNaN verifies early exit optimization in scanning
func TestToJSONSafe_EarlyNaN(t *testing.T) {
	fd := &FileData{
		Headers:  []string{"A", "B"},
		RowNames: []string{"R1", "R2"},
		Data: [][]float64{
			{math.NaN(), 2.0},  // NaN in first position should trigger early exit
			{3.0, 4.0},
		},
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.MissingMask == nil {
		t.Fatal("Expected non-nil MissingMask")
	}
	if !result.MissingMask[0][0] {
		t.Error("Expected MissingMask[0][0] to be true for NaN in first position")
	}
}

// TestToJSONSafe_NilInput verifies nil safety
func TestToJSONSafe_NilInput(t *testing.T) {
	var fd *FileData = nil
	result := fd.ToJSONSafe()
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
}

// TestToJSONSafe_NumericTargetsWithNaN verifies numeric target column handling
func TestToJSONSafe_NumericTargetsWithNaN(t *testing.T) {
	fd := &FileData{
		Headers:  []string{"A", "B"},
		RowNames: []string{"R1", "R2"},
		Data: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
		NumericTargetColumns: map[string][]float64{
			"target": {1.0, math.NaN(), 3.0},
		},
	}
	
	result := fd.ToJSONSafe()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// The main data has no NaN, so MissingMask should be nil
	if result.MissingMask != nil {
		t.Error("Expected nil MissingMask when main data has no missing values")
	}
	if result.NumericTargetColumns == nil {
		t.Fatal("Expected NumericTargetColumns to be converted")
	}
	if len(result.NumericTargetColumns["target"]) != 3 {
		t.Error("NumericTargetColumns should have correct length")
	}
}

// Benchmark to verify performance improvement
func BenchmarkToJSONSafe_NoMissing(b *testing.B) {
	// Setup large dataset without missing values
	rows, cols := 500, 50
	data := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			data[i][j] = float64(i*cols + j)
		}
	}
	
	fd := &FileData{
		Data: data,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fd.ToJSONSafe()
	}
}

func BenchmarkToJSONSafe_WithMissing(b *testing.B) {
	// Setup large dataset with 10% missing values
	rows, cols := 500, 50
	data := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			if (i*cols+j)%10 == 0 {
				data[i][j] = math.NaN()
			} else {
				data[i][j] = float64(i*cols + j)
			}
		}
	}
	
	fd := &FileData{
		Data: data,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fd.ToJSONSafe()
	}
}