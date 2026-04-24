// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// CombineResult holds the merged column data produced by CombineColumns.
// The caller is responsible for converting NumericTargetData to any
// application-specific JSON-safe type (e.g. []types.JSONFloat64).
type CombineResult struct {
	// Headers is the ordered list of column names after merging.
	Headers []string
	// RowNames holds per-row labels (may be nil).
	RowNames []string
	// Data is the row-major data matrix with all columns as strings.
	Data [][]string
	// Rows is the number of data rows.
	Rows int
	// Columns is the number of columns (= len(Headers)).
	Columns int
	// CategoricalColumns maps each categorical column name to its value slice.
	CategoricalColumns map[string][]string
	// NumericTargetData maps each target column name to its float64 value slice.
	NumericTargetData map[string][]float64
	// ColumnTypes maps each column name to "numeric", "categorical", or "target".
	ColumnTypes map[string]string
}

// GetOriginalHeaders reads the first line of the CSV content string and returns
// the column headers in their original order. When the format has both
// HasHeaders and HasRowNames set, the leading row-name header column is skipped.
// Returns an empty slice if the content is empty or cannot be read.
func GetOriginalHeaders(content string, format types.CSVFormat) []string {
	r := csv.NewReader(strings.NewReader(content))
	r.Comma = format.FieldDelimiter
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.Read()
	if err != nil || len(records) == 0 {
		return []string{}
	}

	if format.HasHeaders && format.HasRowNames {
		return records[1:]
	}
	return records
}

// CombineColumns merges numeric, categorical, and target columns from a parsed
// CSV dataset into a single unified CombineResult, preserving the column order
// given by originalHeaders.
//
// When originalHeaders is empty the columns are ordered: numeric first,
// then categorical, then target — map iteration order applies within each group.
//
// Columns listed in originalHeaders that do not appear in any of the three data
// sources are silently skipped.
func CombineColumns(
	csvData *types.CSVData,
	categoricalData map[string][]string,
	numericTargetData map[string][]float64,
	originalHeaders []string,
) *CombineResult {
	// If no original headers were provided, build a fallback order.
	if len(originalHeaders) == 0 {
		originalHeaders = make([]string, 0, len(csvData.Headers)+len(categoricalData)+len(numericTargetData))
		originalHeaders = append(originalHeaders, csvData.Headers...)
		for colName := range categoricalData {
			originalHeaders = append(originalHeaders, colName)
		}
		for colName := range numericTargetData {
			originalHeaders = append(originalHeaders, colName)
		}
	}

	// Index numeric columns for O(1) lookup.
	numericColIndex := make(map[string]int, len(csvData.Headers))
	for idx, header := range csvData.Headers {
		numericColIndex[header] = idx
	}

	allHeaders := make([]string, 0, len(originalHeaders))
	allData := make([][]string, csvData.Rows)
	columnTypes := make(map[string]string, len(originalHeaders))

	for i := range allData {
		allData[i] = make([]string, 0, len(originalHeaders))
	}

	for _, header := range originalHeaders {
		if colIdx, isNumeric := numericColIndex[header]; isNumeric {
			allHeaders = append(allHeaders, header)
			columnTypes[header] = "numeric"
			for rowIdx := 0; rowIdx < csvData.Rows; rowIdx++ {
				if csvData.MissingMask != nil && csvData.MissingMask[rowIdx][colIdx] {
					allData[rowIdx] = append(allData[rowIdx], "")
				} else {
					allData[rowIdx] = append(allData[rowIdx], strconv.FormatFloat(csvData.Matrix[rowIdx][colIdx], 'g', -1, 64))
				}
			}
		} else if values, isCategorical := categoricalData[header]; isCategorical {
			allHeaders = append(allHeaders, header)
			columnTypes[header] = "categorical"
			for rowIdx, value := range values {
				if rowIdx < len(allData) {
					allData[rowIdx] = append(allData[rowIdx], value)
				}
			}
		} else if values, isTarget := numericTargetData[header]; isTarget {
			allHeaders = append(allHeaders, header)
			columnTypes[header] = "target"
			for rowIdx, value := range values {
				if rowIdx < len(allData) {
					allData[rowIdx] = append(allData[rowIdx], strconv.FormatFloat(value, 'g', -1, 64))
				}
			}
		}
		// Headers not found in any source are silently skipped.
	}

	return &CombineResult{
		Headers:            allHeaders,
		RowNames:           csvData.RowNames,
		Data:               allData,
		Rows:               csvData.Rows,
		Columns:            len(allHeaders),
		CategoricalColumns: categoricalData,
		NumericTargetData:  numericTargetData,
		ColumnTypes:        columnTypes,
	}
}
