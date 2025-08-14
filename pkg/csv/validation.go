// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"fmt"
	"math"

	"github.com/bitjungle/gopca/pkg/types"
)

// NewValidator creates a new CSV validator with the given options
func NewValidator(opts Options) *Validator {
	return &Validator{opts: opts}
}

// Validate performs comprehensive validation on CSV data
func (v *Validator) Validate(data *Data) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		ColumnStats: []ColumnStatistics{},
	}

	// Basic structure validation
	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "nil CSV data")
		return result
	}

	if data.Matrix == nil && data.StringData == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "no data present (neither numeric nor string)")
		return result
	}

	// Validate numeric data if present
	if data.Matrix != nil {
		v.validateNumericData(data, result)
	}

	// Validate string data if present
	if data.StringData != nil {
		v.validateStringData(data, result)
	}

	// Check headers consistency
	if len(data.Headers) > 0 {
		expectedCols := data.Columns
		if len(data.Headers) != expectedCols {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("header count (%d) doesn't match column count (%d)",
					len(data.Headers), expectedCols))
		}
	}

	// Check row names consistency
	if len(data.RowNames) > 0 && len(data.RowNames) != data.Rows {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("row name count (%d) doesn't match row count (%d)",
				len(data.RowNames), data.Rows))
	}

	return result
}

// validateNumericData validates numeric matrix data
func (v *Validator) validateNumericData(data *Data, result *ValidationResult) {
	if len(data.Matrix) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "empty data matrix")
		return
	}

	// Check dimensions consistency
	if data.Rows != len(data.Matrix) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("row count mismatch: reported %d, actual %d",
				data.Rows, len(data.Matrix)))
	}

	// Check for consistent column count and collect statistics
	expectedCols := data.Columns
	if expectedCols == 0 && len(data.Matrix) > 0 && len(data.Matrix[0]) > 0 {
		expectedCols = len(data.Matrix[0])
	}

	// Initialize column statistics
	result.ColumnStats = make([]ColumnStatistics, expectedCols)
	for j := 0; j < expectedCols; j++ {
		result.ColumnStats[j] = ColumnStatistics{
			Index:    j,
			DataType: "numeric",
		}
		if j < len(data.Headers) {
			result.ColumnStats[j].Name = data.Headers[j]
		} else {
			result.ColumnStats[j].Name = fmt.Sprintf("Column_%d", j+1)
		}
	}

	// Validate each row and collect statistics
	for i, row := range data.Matrix {
		if len(row) != expectedCols {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("row %d has %d columns, expected %d",
					i+1, len(row), expectedCols))
			continue
		}

		// Process each value for statistics
		for j, val := range row {
			if j >= len(result.ColumnStats) {
				continue
			}

			stats := &result.ColumnStats[j]

			// Check for missing values
			isMissing := false
			if data.MissingMask != nil && i < len(data.MissingMask) &&
				j < len(data.MissingMask[i]) {
				isMissing = data.MissingMask[i][j]
			}

			if isMissing || math.IsNaN(val) {
				stats.Missing++
			} else {
				stats.NonMissing++

				// Update min/max
				if stats.NonMissing == 1 {
					stats.Min = val
					stats.Max = val
				} else {
					if val < stats.Min {
						stats.Min = val
					}
					if val > stats.Max {
						stats.Max = val
					}
				}
			}
		}
	}

	// Calculate statistics and check for issues
	for j := range result.ColumnStats {
		stats := &result.ColumnStats[j]

		// Calculate missing percentage
		total := stats.Missing + stats.NonMissing
		if total > 0 {
			stats.MissingPercent = float64(stats.Missing) / float64(total) * 100
		}

		// Check for all missing values
		if stats.NonMissing == 0 {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("column '%s' contains only missing values", stats.Name))
			continue
		}

		// Calculate mean and variance for numeric columns
		if stats.NonMissing > 0 {
			// First pass: calculate mean
			sum := 0.0
			count := 0
			for i := range data.Matrix {
				if i >= len(data.Matrix) || j >= len(data.Matrix[i]) {
					continue
				}
				val := data.Matrix[i][j]
				if !math.IsNaN(val) {
					sum += val
					count++
				}
			}
			if count > 0 {
				stats.Mean = sum / float64(count)
			}

			// Second pass: calculate variance
			if count > 1 {
				sumSquares := 0.0
				for i := range data.Matrix {
					if i >= len(data.Matrix) || j >= len(data.Matrix[i]) {
						continue
					}
					val := data.Matrix[i][j]
					if !math.IsNaN(val) {
						diff := val - stats.Mean
						sumSquares += diff * diff
					}
				}
				variance := sumSquares / float64(count-1)
				stats.StdDev = math.Sqrt(variance)

				// Check for zero variance
				if variance < 1e-10 {
					stats.HasZeroVariance = true
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("column '%s' has near-zero variance", stats.Name))
				}
			}
		}

		// Warn about high missing percentage
		if stats.MissingPercent > 50 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("column '%s' has %.1f%% missing values",
					stats.Name, stats.MissingPercent))
		}
	}
}

// validateStringData validates string matrix data
func (v *Validator) validateStringData(data *Data, result *ValidationResult) {
	if len(data.StringData) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "empty string data")
		return
	}

	// Check dimensions consistency
	if data.Rows != len(data.StringData) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("row count mismatch: reported %d, actual %d",
				data.Rows, len(data.StringData)))
	}

	// Check for consistent column count
	expectedCols := data.Columns
	if expectedCols == 0 && len(data.StringData) > 0 && len(data.StringData[0]) > 0 {
		expectedCols = len(data.StringData[0])
	}

	for i, row := range data.StringData {
		if len(row) != expectedCols {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("row %d has %d columns, expected %d",
					i+1, len(row), expectedCols))
		}
	}
}

// ValidateFile validates a CSV file
func ValidateFile(filename string, opts Options) (*ValidationResult, error) {
	reader := NewReader(opts)
	data, err := reader.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	validator := NewValidator(opts)
	return validator.Validate(data), nil
}

// ValidateStructure performs basic structural validation
func ValidateStructure(data *Data) error {
	if data == nil {
		return fmt.Errorf("nil CSV data")
	}

	if data.Matrix == nil && data.StringData == nil {
		return fmt.Errorf("no data present")
	}

	if data.Matrix != nil {
		if len(data.Matrix) == 0 {
			return fmt.Errorf("empty data matrix")
		}

		// Check consistent columns
		cols := len(data.Matrix[0])
		for i, row := range data.Matrix {
			if len(row) != cols {
				return fmt.Errorf("inconsistent columns at row %d: expected %d, got %d",
					i+1, cols, len(row))
			}
		}
	}

	return nil
}

// AnalyzeMissingValues analyzes missing value patterns
func AnalyzeMissingValues(data *Data) map[string]interface{} {
	analysis := make(map[string]interface{})

	if data == nil || data.Matrix == nil {
		return analysis
	}

	totalCells := data.Rows * data.Columns
	missingCells := 0

	// Count missing values
	for i := range data.Matrix {
		for j := range data.Matrix[i] {
			if math.IsNaN(data.Matrix[i][j]) {
				missingCells++
			} else if data.MissingMask != nil && i < len(data.MissingMask) &&
				j < len(data.MissingMask[i]) && data.MissingMask[i][j] {
				missingCells++
			}
		}
	}

	analysis["total_cells"] = totalCells
	analysis["missing_cells"] = missingCells
	analysis["missing_percentage"] = float64(missingCells) / float64(totalCells) * 100

	// Analyze by column
	columnMissing := make([]int, data.Columns)
	for i := range data.Matrix {
		for j := range data.Matrix[i] {
			if math.IsNaN(data.Matrix[i][j]) {
				columnMissing[j]++
			} else if data.MissingMask != nil && i < len(data.MissingMask) &&
				j < len(data.MissingMask[i]) && data.MissingMask[i][j] {
				columnMissing[j]++
			}
		}
	}
	analysis["missing_by_column"] = columnMissing

	// Analyze by row
	rowMissing := make([]int, data.Rows)
	for i := range data.Matrix {
		for j := range data.Matrix[i] {
			if math.IsNaN(data.Matrix[i][j]) {
				rowMissing[i]++
			} else if data.MissingMask != nil && i < len(data.MissingMask) &&
				j < len(data.MissingMask[i]) && data.MissingMask[i][j] {
				rowMissing[i]++
			}
		}
	}
	analysis["missing_by_row"] = rowMissing

	return analysis
}

// GetMissingValueInfo returns information about missing values in selected columns
func (d *Data) GetMissingValueInfo(selectedColumns []int) *types.MissingValueInfo {
	info := &types.MissingValueInfo{
		ColumnIndices:   []int{},
		RowsAffected:    []int{},
		TotalMissing:    0,
		MissingByColumn: make(map[int]int),
	}

	// If no columns selected, analyze all columns
	if len(selectedColumns) == 0 {
		selectedColumns = make([]int, d.Columns)
		for i := range selectedColumns {
			selectedColumns[i] = i
		}
	}

	// Track which rows have missing values in selected columns
	rowHasMissing := make(map[int]bool)

	// Analyze each selected column
	for _, col := range selectedColumns {
		if col < 0 || col >= d.Columns {
			continue
		}

		columnMissingCount := 0
		for row := 0; row < d.Rows; row++ {
			if d.MissingMask != nil && d.MissingMask[row][col] {
				columnMissingCount++
				rowHasMissing[row] = true
				info.TotalMissing++
			} else if math.IsNaN(d.Matrix[row][col]) {
				columnMissingCount++
				rowHasMissing[row] = true
				info.TotalMissing++
			}
		}

		if columnMissingCount > 0 {
			info.ColumnIndices = append(info.ColumnIndices, col)
			info.MissingByColumn[col] = columnMissingCount
		}
	}

	// Collect affected rows
	for row := range rowHasMissing {
		info.RowsAffected = append(info.RowsAffected, row)
	}

	return info
}
