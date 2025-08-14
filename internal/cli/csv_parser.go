// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cli

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
)

// Type aliases for backward compatibility
type CSVData = pkgcsv.Data
type CSVParseOptions = pkgcsv.Options

// NewCSVParseOptions creates default parse options
func NewCSVParseOptions() CSVParseOptions {
	return pkgcsv.DefaultOptions()
}

// ParseCSV reads and parses a CSV file according to the given options
func ParseCSV(filename string, options CSVParseOptions) (*CSVData, error) {
	// Set parse mode to Mixed for CLI usage
	options.ParseMode = pkgcsv.ParseMixed

	// Use unified CSV reader
	reader := pkgcsv.NewReader(options)
	data, err := reader.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Notify about categorical columns if any
	if len(data.CategoricalColumns) > 0 {
		fmt.Fprintf(os.Stderr, "\nNote: Detected and excluded %d categorical column(s):\n", len(data.CategoricalColumns))
		for colName := range data.CategoricalColumns {
			fmt.Fprintf(os.Stderr, "  - %s\n", colName)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	return data, nil
}

// ParseCSVReader parses CSV data from an io.Reader
func ParseCSVReader(r io.Reader, options CSVParseOptions) (*CSVData, error) {
	// Set parse mode to Mixed for CLI usage
	options.ParseMode = pkgcsv.ParseMixed

	// Use unified CSV reader
	reader := pkgcsv.NewReader(options)
	data, err := reader.Read(r)
	if err != nil {
		return nil, err
	}

	// Notify about categorical columns if any
	if len(data.CategoricalColumns) > 0 {
		fmt.Fprintf(os.Stderr, "\nNote: Detected and excluded %d categorical column(s):\n", len(data.CategoricalColumns))
		for colName := range data.CategoricalColumns {
			fmt.Fprintf(os.Stderr, "  - %s\n", colName)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	return data, nil
}

// ValidateCSVData performs basic validation on parsed CSV data
func ValidateCSVData(data *CSVData) error {
	if data == nil {
		return fmt.Errorf("nil CSV data")
	}

	if len(data.Matrix) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	if data.Rows != len(data.Matrix) {
		return fmt.Errorf("row count mismatch")
	}

	// Check for consistent column count
	for i, row := range data.Matrix {
		if len(row) != data.Columns {
			return fmt.Errorf("row %d has %d columns, expected %d",
				i+1, len(row), data.Columns)
		}
	}

	// Check for all NaN columns
	for j := 0; j < data.Columns; j++ {
		allNaN := true
		for i := 0; i < data.Rows; i++ {
			if !math.IsNaN(data.Matrix[i][j]) {
				allNaN = false
				break
			}
		}
		if allNaN {
			colName := fmt.Sprintf("%d", j+1)
			if j < len(data.Headers) {
				colName = data.Headers[j]
			}
			return fmt.Errorf("column '%s' contains only missing values", colName)
		}
	}

	return nil
}

// GetDataSummary returns a summary of the CSV data
func GetDataSummary(data *CSVData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Data dimensions: %d rows × %d columns\n", data.Rows, data.Columns))

	if len(data.Headers) > 0 {
		sb.WriteString(fmt.Sprintf("Column names: %s", strings.Join(data.Headers, ", ")))
		if len(data.Headers) > 5 {
			sb.WriteString(fmt.Sprintf(" (showing first 5 of %d)\n", len(data.Headers)))
		} else {
			sb.WriteString("\n")
		}
	}

	if len(data.RowNames) > 0 {
		sb.WriteString(fmt.Sprintf("Row names: %s", strings.Join(data.RowNames[:min(5, len(data.RowNames))], ", ")))
		if len(data.RowNames) > 5 {
			sb.WriteString(fmt.Sprintf(" ... (showing first 5 of %d)\n", len(data.RowNames)))
		} else {
			sb.WriteString("\n")
		}
	}

	// Count missing values
	missingCount := 0
	for i := 0; i < data.Rows; i++ {
		for j := 0; j < data.Columns; j++ {
			if math.IsNaN(data.Matrix[i][j]) {
				missingCount++
			}
		}
	}

	totalValues := data.Rows * data.Columns
	missingPercent := float64(missingCount) / float64(totalValues) * 100
	sb.WriteString(fmt.Sprintf("Missing values: %d (%.1f%%)\n", missingCount, missingPercent))

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseCSVMixedWithTargets reads and parses a CSV file with support for target columns
func ParseCSVMixedWithTargets(filename string, options CSVParseOptions, targetColumns []string) (*CSVData, map[string][]string, map[string][]float64, error) {
	// Set parse mode to Mixed with targets
	options.ParseMode = pkgcsv.ParseMixedWithTargets

	// Set target suffix if we have target columns
	if len(targetColumns) > 0 {
		// Assume target columns are identified by suffix
		options.TargetSuffix = "#target"
	}

	// Use unified CSV reader
	reader := pkgcsv.NewReader(options)
	data, err := reader.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, err
	}

	return data, data.CategoricalColumns, data.NumericTargetColumns, nil
}
