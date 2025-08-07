// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package types

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strings"
)

// ParseCSVMixed parses a CSV file that may contain both numeric and categorical columns
func ParseCSVMixed(r io.Reader, format CSVFormat) (*CSVData, map[string][]string, error) {
	// First, read all records as strings
	csvReader := csv.NewReader(r)
	csvReader.Comma = format.FieldDelimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("empty CSV file")
	}

	// Determine dimensions
	startRow := 0
	headers := []string{}
	if format.HasHeaders {
		headers = records[0]
		startRow = 1
	}

	if len(records) <= startRow {
		return nil, nil, fmt.Errorf("no data rows found")
	}

	startCol := 0
	rowNames := []string{}
	if format.HasRowNames {
		startCol = 1
		// Extract row names
		for i := startRow; i < len(records); i++ {
			if len(records[i]) > 0 {
				rowNames = append(rowNames, records[i][0])
			}
		}
		// Remove row name from headers if present
		if len(headers) > 0 && format.HasHeaders {
			headers = headers[1:]
		}
	}

	numRows := len(records) - startRow
	numCols := len(records[startRow]) - startCol

	if numCols <= 0 {
		return nil, nil, fmt.Errorf("no data columns found")
	}

	// Detect column types by checking first N data rows
	numericCols := []int{}
	categoricalCols := []int{}

	for j := 0; j < numCols; j++ {
		isNumeric := true
		hasAnyValue := false

		// Check first N rows to determine type
		for i := startRow; i < len(records) && i < startRow+DefaultColumnTypeDetectionSampleSize; i++ {
			if j+startCol >= len(records[i]) {
				continue
			}

			value := strings.TrimSpace(records[i][j+startCol])
			if value == "" {
				continue
			}

			hasAnyValue = true

			// Check if the value is numeric
			isNum, _ := isNumericValue(value, format)
			if !isNum {
				// Not a number - this is categorical
				isNumeric = false
				break
			}
		}

		if !hasAnyValue || isNumeric {
			numericCols = append(numericCols, j)
		} else {
			categoricalCols = append(categoricalCols, j)
		}
	}

	// Extract numeric data
	numericHeaders := make([]string, len(numericCols))
	for i, colIdx := range numericCols {
		if colIdx < len(headers) {
			numericHeaders[i] = headers[colIdx]
		}
	}

	data := &CSVData{
		Headers:     numericHeaders,
		RowNames:    rowNames,
		Matrix:      make([][]float64, numRows),
		MissingMask: make([][]bool, numRows),
		Rows:        numRows,
		Columns:     len(numericCols),
	}

	// Parse numeric columns
	for i := 0; i < numRows; i++ {
		data.Matrix[i] = make([]float64, len(numericCols))
		data.MissingMask[i] = make([]bool, len(numericCols))

		rowIdx := i + startRow
		if rowIdx >= len(records) {
			continue
		}

		for j, colIdx := range numericCols {
			if colIdx+startCol >= len(records[rowIdx]) {
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			value := strings.TrimSpace(records[rowIdx][colIdx+startCol])

			// Try to parse the value as numeric
			isNum, val := isNumericValue(value, format)
			if !isNum {
				// This shouldn't happen if column detection worked correctly
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			data.Matrix[i][j] = val
			data.MissingMask[i][j] = math.IsNaN(val)
		}
	}

	// Extract categorical data
	categoricalData := make(map[string][]string)
	for _, colIdx := range categoricalCols {
		colName := ""
		if colIdx < len(headers) {
			colName = headers[colIdx]
		} else {
			colName = fmt.Sprintf("Column%d", colIdx+1)
		}

		values := make([]string, numRows)
		for i := 0; i < numRows; i++ {
			rowIdx := i + startRow
			if rowIdx < len(records) && colIdx+startCol < len(records[rowIdx]) {
				values[i] = strings.TrimSpace(records[rowIdx][colIdx+startCol])
			}
		}

		categoricalData[colName] = values
	}

	return data, categoricalData, nil
}

// isTargetColumn checks if a column name indicates it should be a target column
// Target columns are marked with "#target" suffix (with or without space) or are in the provided target list
func isTargetColumn(columnName string, targetColumns []string) bool {
	lowerName := strings.ToLower(columnName)

	// Check if column ends with "#target" (no space) or " #target" (with space)
	if strings.HasSuffix(lowerName, "#target") || strings.HasSuffix(lowerName, " #target") {
		return true
	}

	// Check if column is in the explicit target list
	for _, target := range targetColumns {
		if strings.EqualFold(columnName, target) {
			return true
		}
	}

	return false
}

// ParseCSVMixedWithTargets parses CSV data with support for numeric target columns
// Target columns are numeric columns that should be available for visualization but not included in PCA
// Columns with "#target" suffix (with or without space) are automatically detected as target columns
func ParseCSVMixedWithTargets(r io.Reader, format CSVFormat, targetColumns []string) (*CSVData, map[string][]string, map[string][]float64, error) {
	// First, read all records as strings
	csvReader := csv.NewReader(r)
	csvReader.Comma = format.FieldDelimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, nil, fmt.Errorf("empty CSV file")
	}

	// Determine dimensions
	startRow := 0
	headers := []string{}
	if format.HasHeaders {
		headers = records[0]
		startRow = 1
	}

	if len(records) <= startRow {
		return nil, nil, nil, fmt.Errorf("no data rows found")
	}

	// Determine row names
	startCol := 0
	rowNames := []string{}
	if format.HasRowNames {
		startCol = 1
		for i := startRow; i < len(records); i++ {
			if len(records[i]) > 0 {
				rowNames = append(rowNames, records[i][0])
			}
		}
		// Remove row name header if present
		if format.HasHeaders && len(headers) > 0 {
			headers = headers[1:]
		}
	}

	numRows := len(records) - startRow
	numCols := len(records[startRow]) - startCol

	// Detect column types - now with target column support
	numericDataCols := []int{}
	numericTargetCols := []int{}
	categoricalCols := []int{}

	// Check each column
	for j := 0; j < numCols; j++ {
		isNumeric := true
		hasAnyValue := false

		// Check first N rows to determine type
		for i := startRow; i < len(records) && i < startRow+DefaultColumnTypeDetectionSampleSize; i++ {
			if j+startCol >= len(records[i]) {
				continue
			}

			value := strings.TrimSpace(records[i][j+startCol])
			if value == "" {
				continue
			}

			hasAnyValue = true

			// Check if the value is numeric
			isNum, _ := isNumericValue(value, format)
			if !isNum {
				// Not a number - this is categorical
				isNumeric = false
				break
			}
		}

		if !hasAnyValue {
			// Empty column - check if it's a target column by name
			colName := ""
			if j < len(headers) {
				colName = headers[j]
			}

			if isTargetColumn(colName, targetColumns) {
				numericTargetCols = append(numericTargetCols, j)
			} else {
				numericDataCols = append(numericDataCols, j)
			}
		} else if isNumeric {
			// Numeric column - check if it's a target
			colName := ""
			if j < len(headers) {
				colName = headers[j]
			}

			if isTargetColumn(colName, targetColumns) {
				numericTargetCols = append(numericTargetCols, j)
			} else {
				numericDataCols = append(numericDataCols, j)
			}
		} else {
			categoricalCols = append(categoricalCols, j)
		}
	}

	// Extract numeric data (non-target columns only)
	numericHeaders := make([]string, len(numericDataCols))
	for i, colIdx := range numericDataCols {
		if colIdx < len(headers) {
			numericHeaders[i] = headers[colIdx]
		}
	}

	data := &CSVData{
		Headers:     numericHeaders,
		RowNames:    rowNames,
		Matrix:      make([][]float64, numRows),
		MissingMask: make([][]bool, numRows),
		Rows:        numRows,
		Columns:     len(numericDataCols),
	}

	// Parse numeric data columns
	for i := 0; i < numRows; i++ {
		data.Matrix[i] = make([]float64, len(numericDataCols))
		data.MissingMask[i] = make([]bool, len(numericDataCols))

		rowIdx := i + startRow
		if rowIdx >= len(records) {
			continue
		}

		for j, colIdx := range numericDataCols {
			if colIdx+startCol >= len(records[rowIdx]) {
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			value := strings.TrimSpace(records[rowIdx][colIdx+startCol])

			// Try to parse the value as numeric
			isNum, val := isNumericValue(value, format)
			if !isNum {
				// This shouldn't happen if column detection worked correctly
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			data.Matrix[i][j] = val
			data.MissingMask[i][j] = math.IsNaN(val)
		}
	}

	// Extract categorical data
	categoricalData := make(map[string][]string)
	for _, colIdx := range categoricalCols {
		colName := ""
		if colIdx < len(headers) {
			colName = headers[colIdx]
		} else {
			colName = fmt.Sprintf("Column%d", colIdx+1)
		}

		values := make([]string, numRows)
		for i := 0; i < numRows; i++ {
			rowIdx := i + startRow
			if rowIdx < len(records) && colIdx+startCol < len(records[rowIdx]) {
				values[i] = strings.TrimSpace(records[rowIdx][colIdx+startCol])
			}
		}

		categoricalData[colName] = values
	}

	// Extract numeric target data
	numericTargetData := make(map[string][]float64)
	for _, colIdx := range numericTargetCols {
		colName := ""
		if colIdx < len(headers) {
			colName = headers[colIdx]
		} else {
			colName = fmt.Sprintf("Column%d", colIdx+1)
		}

		values := make([]float64, numRows)
		for i := 0; i < numRows; i++ {
			rowIdx := i + startRow
			if rowIdx < len(records) && colIdx+startCol < len(records[rowIdx]) {
				value := strings.TrimSpace(records[rowIdx][colIdx+startCol])
				isNum, val := isNumericValue(value, format)
				if isNum {
					values[i] = val
				} else {
					values[i] = math.NaN()
				}
			} else {
				values[i] = math.NaN()
			}
		}

		numericTargetData[colName] = values
	}

	return data, categoricalData, numericTargetData, nil
}
