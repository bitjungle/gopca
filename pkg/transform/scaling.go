// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// applyStandardize applies z-score standardization (mean=0, std=1) to the
// specified numeric columns.
func applyStandardize(data [][]string, columnTypes map[string]string, headers []string, opts Options, result *Result) error {
	for _, colName := range opts.Columns {
		colIndex := findColumn(headers, colName)
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}

		if columnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}

		values, indices := collectNumeric(data, colIndex)

		if len(values) < 2 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has insufficient numeric values for standardization", colName))
			continue
		}

		mean := 0.0
		for _, v := range values {
			mean += v
		}
		mean /= float64(len(values))

		variance := 0.0
		for _, v := range values {
			variance += (v - mean) * (v - mean)
		}
		variance /= float64(len(values))
		stdDev := math.Sqrt(variance)

		if stdDev < 1e-10 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has zero variance, cannot standardize", colName))
			continue
		}

		for i, idx := range indices {
			data[idx][colIndex] = fmt.Sprintf("%.6g", (values[i]-mean)/stdDev)
		}

		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Standardized %d values in column '%s' (mean=%.3f, std=%.3f)", len(values), colName, mean, stdDev))
	}

	return nil
}

// applyMinMax applies min-max scaling to the specified numeric columns.
// Values are scaled to [opts.MinValue, opts.MaxValue]; defaults to [0, 1].
func applyMinMax(data [][]string, columnTypes map[string]string, headers []string, opts Options, result *Result) error {
	targetMin := opts.MinValue
	targetMax := opts.MaxValue
	if targetMax <= targetMin {
		targetMin = 0.0
		targetMax = 1.0
	}

	for _, colName := range opts.Columns {
		colIndex := findColumn(headers, colName)
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}

		if columnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}

		values, indices := collectNumeric(data, colIndex)

		if len(values) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no numeric values", colName))
			continue
		}

		minVal := math.Inf(1)
		maxVal := math.Inf(-1)
		for _, v := range values {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}

		if maxVal <= minVal {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has constant values, cannot scale", colName))
			continue
		}

		for i, idx := range indices {
			scaled := (values[i]-minVal)/(maxVal-minVal)*(targetMax-targetMin) + targetMin
			data[idx][colIndex] = fmt.Sprintf("%.6g", scaled)
		}

		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Scaled %d values in column '%s' to range [%.2f, %.2f]", len(values), colName, targetMin, targetMax))
	}

	return nil
}

// collectNumeric extracts the non-missing numeric values from a column and
// returns the values alongside their row indices in data.
func collectNumeric(data [][]string, colIndex int) (values []float64, indices []int) {
	for i := range data {
		if colIndex >= len(data[i]) {
			continue
		}
		v := strings.TrimSpace(data[i][colIndex])
		if v == "" {
			continue
		}
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		values = append(values, num)
		indices = append(indices, i)
	}
	return values, indices
}
