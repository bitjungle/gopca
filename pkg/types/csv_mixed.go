// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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
	rowNamesHeader := ""
	if format.HasRowNames {
		startCol = 1
		// Extract row names
		for i := startRow; i < len(records); i++ {
			if len(records[i]) > 0 {
				rowNames = append(rowNames, records[i][0])
			}
		}
		// Remove row name from headers if present, keeping its name (#859).
		if len(headers) > 0 && format.HasHeaders {
			rowNamesHeader = headers[0]
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
		Headers:        numericHeaders,
		RowNames:       rowNames,
		RowNamesHeader: rowNamesHeader,
		Matrix:         make([][]float64, numRows),
		MissingMask:    make([][]bool, numRows),
		Rows:           numRows,
		Columns:        len(numericCols),
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

// isCategoryColumn reports whether a column is marked as categorical by name.
//
// The marker exists for the one thing the type of a column cannot tell us: that
// its numbers are labels. A processing code holding 1..11, a site number, a
// batch identifier -- each parses as numeric and would otherwise enter the PCA
// as a measurement, where its variance is arbitrary and can dwarf everything
// else. Marking it says "these are categories", and the column is then held out
// and offered for colouring, exactly as a text column would be.
//
// On a column that already holds text the marker is redundant rather than
// wrong: it agrees with what the type says. That is deliberate -- a marker
// should never contradict itself.
//
// A space before the marker is accepted because the suffix test does not care
// what precedes it: "site #category" ends with "#category" just as
// "site#category" does. swiss_roll.csv ships "color #target" written that way,
// so the form does occur. A space *after* the hash -- "site# category" -- is not
// recognised, matching isTargetColumn, which has the same limitation.
func isCategoryColumn(columnName string) bool {
	return strings.HasSuffix(strings.ToLower(columnName), "#category")
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
	rowNamesHeader := ""
	if format.HasRowNames {
		startCol = 1
		for i := startRow; i < len(records); i++ {
			if len(records[i]) > 0 {
				rowNames = append(rowNames, records[i][0])
			}
		}
		// Remove row name header if present, keeping its name (#859).
		if format.HasHeaders && len(headers) > 0 {
			rowNamesHeader = headers[0]
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

		colName := ""
		if j < len(headers) {
			colName = headers[j]
		}

		switch {
		case isCategoryColumn(colName):
			// Marked as categories, whatever the values look like. This is the
			// only case where the name overrides what the column contains, and
			// it is why the marker exists (#914).
			categoricalCols = append(categoricalCols, j)
		case !hasAnyValue || isNumeric:
			// An empty column is treated as numeric, as it always has been:
			// there is nothing in it to suggest otherwise.
			if isTargetColumn(colName, targetColumns) {
				numericTargetCols = append(numericTargetCols, j)
			} else {
				numericDataCols = append(numericDataCols, j)
			}
		default:
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
		Headers:        numericHeaders,
		RowNames:       rowNames,
		RowNamesHeader: rowNamesHeader,
		Matrix:         make([][]float64, numRows),
		MissingMask:    make([][]bool, numRows),
		Rows:           numRows,
		Columns:        len(numericDataCols),
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
