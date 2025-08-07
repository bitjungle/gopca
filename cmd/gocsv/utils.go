// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"strconv"

	"github.com/bitjungle/gopca/pkg/utils"
)

// Common missing value indicators for GoCSV
// This extends the default set with additional indicators
var missingValueIndicators = append(utils.DefaultMissingValues(), "-", "?", "none", "missing")

// isMissingValue checks if a value is considered missing
func isMissingValue(value string) bool {
	return utils.IsMissingValue(value, missingValueIndicators)
}

// parseNumericValue attempts to parse a string as a float64
// Returns the parsed value and true if successful, or 0 and false if not
func parseNumericValue(value string) (float64, bool) {
	val, isMissing, err := utils.ParseNumericValueWithMissing(value, '.', missingValueIndicators)
	if isMissing || err != nil {
		return 0, false
	}
	return val, true
}

// getNumericValues extracts all numeric values from a column
// Skips missing values and non-numeric values
func getNumericValues(data [][]string, colIdx int) []float64 {
	values := make([]float64, 0)
	for rowIdx := 0; rowIdx < len(data); rowIdx++ {
		if colIdx >= len(data[rowIdx]) {
			continue
		}
		
		if num, ok := parseNumericValue(data[rowIdx][colIdx]); ok {
			values = append(values, num)
		}
	}
	return values
}

// getColumnMean calculates the mean of numeric values in a column
// Returns 0 if no numeric values are found
func getColumnMean(data [][]string, colIdx int) float64 {
	values := getNumericValues(data, colIdx)
	if len(values) == 0 {
		return 0
	}
	
	return calculateMean(values)
}

// getColumnMedian calculates the median of numeric values in a column
// Returns 0 if no numeric values are found
func getColumnMedian(data [][]string, colIdx int) float64 {
	values := getNumericValues(data, colIdx)
	if len(values) == 0 {
		return 0
	}
	
	return calculateMedian(values)
}

// fillMissingWithValue fills missing values in a column with the specified value
func fillMissingWithValue(data [][]string, colIdx int, fillValue string) {
	for rowIdx := 0; rowIdx < len(data); rowIdx++ {
		if colIdx >= len(data[rowIdx]) {
			continue
		}
		
		if isMissingValue(data[rowIdx][colIdx]) {
			data[rowIdx][colIdx] = fillValue
		}
	}
}

// fillMissingWithMean fills missing values in a column with the column mean
func fillMissingWithMean(data [][]string, colIdx int) {
	mean := getColumnMean(data, colIdx)
	fillValue := strconv.FormatFloat(mean, 'f', -1, 64)
	fillMissingWithValue(data, colIdx, fillValue)
}

// fillMissingWithMedian fills missing values in a column with the column median
func fillMissingWithMedian(data [][]string, colIdx int) {
	median := getColumnMedian(data, colIdx)
	fillValue := strconv.FormatFloat(median, 'f', -1, 64)
	fillMissingWithValue(data, colIdx, fillValue)
}