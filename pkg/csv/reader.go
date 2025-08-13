// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// NewReader creates a new CSV reader with the given options
func NewReader(opts Options) *Reader {
	return &Reader{opts: opts}
}

// ReadFile reads and parses a CSV file
func (r *Reader) ReadFile(filename string) (*Data, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return r.Read(file)
}

// Read parses CSV data from an io.Reader
func (r *Reader) Read(input io.Reader) (*Data, error) {
	reader := csv.NewReader(input)
	reader.Comma = r.opts.Delimiter
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable fields initially
	reader.ReuseRecord = r.opts.StreamingMode

	// Create null value map for fast lookup
	nullMap := make(map[string]bool)
	for _, nv := range r.opts.NullValues {
		nullMap[nv] = true
	}

	// Read all records or stream based on options
	var records [][]string
	var err error

	if r.opts.StreamingMode {
		// Stream processing for large files
		return r.readStreaming(reader, nullMap)
	}

	// Read all records at once
	records, err = reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Skip rows if needed
	if r.opts.SkipRows > 0 {
		if r.opts.SkipRows >= len(records) {
			return nil, fmt.Errorf("skip rows (%d) exceeds total rows (%d)",
				r.opts.SkipRows, len(records))
		}
		records = records[r.opts.SkipRows:]
	}

	// Process based on parse mode
	switch r.opts.ParseMode {
	case ParseString:
		return r.parseAsString(records, nullMap)
	case ParseMixed:
		return r.parseAsMixed(records, nullMap)
	case ParseMixedWithTargets:
		return r.parseAsMixedWithTargets(records, nullMap)
	default: // ParseNumeric
		return r.parseAsNumeric(records, nullMap)
	}
}

// parseAsNumeric parses all data as numeric values
func (r *Reader) parseAsNumeric(records [][]string, nullMap map[string]bool) (*Data, error) {
	data := &Data{}
	currentRow := 0

	// Handle headers
	if r.opts.HasHeaders {
		if len(records) <= currentRow {
			return nil, fmt.Errorf("no data rows after header")
		}

		headerRow := records[currentRow]
		currentRow++

		// Extract column names, skipping first if it's a row name column
		startCol := 0
		if r.opts.HasRowNames {
			startCol = 1
		}

		if startCol >= len(headerRow) {
			return nil, fmt.Errorf("no data columns found")
		}

		data.Headers = make([]string, len(headerRow)-startCol)
		copy(data.Headers, headerRow[startCol:])
	}

	// Process data rows
	dataRows := records[currentRow:]
	if len(dataRows) == 0 {
		return nil, fmt.Errorf("no data rows found")
	}

	// Apply max rows limit if specified
	if r.opts.MaxRows > 0 && len(dataRows) > r.opts.MaxRows {
		dataRows = dataRows[:r.opts.MaxRows]
	}

	// Determine dimensions
	startCol := 0
	if r.opts.HasRowNames {
		startCol = 1
		data.RowNames = make([]string, len(dataRows))
	}

	// Validate consistent column count
	expectedCols := len(records[currentRow]) - startCol
	if expectedCols <= 0 {
		return nil, fmt.Errorf("no data columns found")
	}

	// Apply column selection if specified
	selectedCols := r.getSelectedColumns(expectedCols)
	actualCols := len(selectedCols)

	// Initialize matrix and missing mask
	data.Matrix = make(types.Matrix, len(dataRows))
	data.MissingMask = make([][]bool, len(dataRows))

	// Parse data
	for i, row := range dataRows {
		if len(row) < startCol {
			return nil, fmt.Errorf("row %d has insufficient columns", i+1)
		}

		// Extract row name if present
		if r.opts.HasRowNames {
			data.RowNames[i] = row[0]
		}

		// Validate column count
		actualRowCols := len(row) - startCol
		if actualRowCols != expectedCols {
			return nil, fmt.Errorf("row %d has %d data columns, expected %d",
				i+1, actualRowCols, expectedCols)
		}

		// Parse numerical data for selected columns
		data.Matrix[i] = make([]float64, actualCols)
		data.MissingMask[i] = make([]bool, actualCols)

		for j, colIdx := range selectedCols {
			value := strings.TrimSpace(row[startCol+colIdx])

			// Check for null values
			if nullMap[value] {
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			// Handle decimal separator if needed
			if r.opts.DecimalSeparator == ',' {
				value = strings.ReplaceAll(value, ",", ".")
			}

			// Try to parse as float
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				// Try special cases
				switch strings.ToLower(value) {
				case "inf", "+inf", "infinity":
					val = math.Inf(1)
				case "-inf", "-infinity":
					val = math.Inf(-1)
				default:
					return nil, fmt.Errorf("cannot parse '%s' at row %d, column %d as number",
						value, i+1, selectedCols[j]+1)
				}
			}

			data.Matrix[i][j] = val
			data.MissingMask[i][j] = false
		}
	}

	// Set dimensions
	data.Rows = len(data.Matrix)
	data.Columns = actualCols

	// Update headers for selected columns
	if len(data.Headers) > 0 && len(selectedCols) != expectedCols {
		selectedHeaders := make([]string, len(selectedCols))
		for i, colIdx := range selectedCols {
			if colIdx < len(data.Headers) {
				selectedHeaders[i] = data.Headers[colIdx]
			} else {
				selectedHeaders[i] = fmt.Sprintf("Column_%d", colIdx+1)
			}
		}
		data.Headers = selectedHeaders
	}

	// Validate column names if present
	if len(data.Headers) > 0 && len(data.Headers) != data.Columns {
		return nil, fmt.Errorf("column name count (%d) doesn't match data columns (%d)",
			len(data.Headers), data.Columns)
	}

	return data, nil
}

// parseAsString parses all data as strings (for GoCSV)
func (r *Reader) parseAsString(records [][]string, nullMap map[string]bool) (*Data, error) {
	data := &Data{}
	currentRow := 0

	// Handle headers
	if r.opts.HasHeaders {
		if len(records) <= currentRow {
			return nil, fmt.Errorf("no data rows after header")
		}

		headerRow := records[currentRow]
		currentRow++

		// Extract column names
		startCol := 0
		if r.opts.HasRowNames {
			startCol = 1
		}

		if startCol >= len(headerRow) {
			return nil, fmt.Errorf("no data columns found")
		}

		data.Headers = make([]string, len(headerRow)-startCol)
		copy(data.Headers, headerRow[startCol:])
	}

	// Process data rows
	dataRows := records[currentRow:]
	if len(dataRows) == 0 {
		return nil, fmt.Errorf("no data rows found")
	}

	// Apply max rows limit
	if r.opts.MaxRows > 0 && len(dataRows) > r.opts.MaxRows {
		dataRows = dataRows[:r.opts.MaxRows]
	}

	// Determine dimensions
	startCol := 0
	if r.opts.HasRowNames {
		startCol = 1
		data.RowNames = make([]string, len(dataRows))
	}

	expectedCols := len(records[currentRow]) - startCol
	if expectedCols <= 0 {
		return nil, fmt.Errorf("no data columns found")
	}

	// Initialize string data
	data.StringData = make([][]string, len(dataRows))

	// Copy string data
	for i, row := range dataRows {
		if len(row) < startCol {
			return nil, fmt.Errorf("row %d has insufficient columns", i+1)
		}

		// Extract row name if present
		if r.opts.HasRowNames {
			data.RowNames[i] = row[0]
		}

		// Copy string values
		data.StringData[i] = make([]string, expectedCols)
		copy(data.StringData[i], row[startCol:])
	}

	data.Rows = len(data.StringData)
	data.Columns = expectedCols

	return data, nil
}

// parseAsMixed automatically detects column types
func (r *Reader) parseAsMixed(records [][]string, nullMap map[string]bool) (*Data, error) {
	// Convert to types.CSVFormat for compatibility with existing mixed parser
	format := types.CSVFormat{
		FieldDelimiter:   r.opts.Delimiter,
		DecimalSeparator: r.opts.DecimalSeparator,
		HasHeaders:       r.opts.HasHeaders,
		HasRowNames:      r.opts.HasRowNames,
		NullValues:       r.opts.NullValues,
	}

	// Use existing mixed parser
	csvData, categoricalData, err := types.ParseCSVMixed(
		strings.NewReader(recordsToString(records)),
		format,
	)
	if err != nil {
		return nil, err
	}

	// Convert to unified Data structure
	data := &Data{
		Matrix:             csvData.Matrix,
		Headers:            csvData.Headers,
		RowNames:           csvData.RowNames,
		MissingMask:        csvData.MissingMask,
		Rows:               csvData.Rows,
		Columns:            csvData.Columns,
		CategoricalColumns: categoricalData,
	}

	return data, nil
}

// parseAsMixedWithTargets detects columns and identifies target columns
func (r *Reader) parseAsMixedWithTargets(records [][]string, nullMap map[string]bool) (*Data, error) {
	// Convert to types.CSVFormat
	format := types.CSVFormat{
		FieldDelimiter:   r.opts.Delimiter,
		DecimalSeparator: r.opts.DecimalSeparator,
		HasHeaders:       r.opts.HasHeaders,
		HasRowNames:      r.opts.HasRowNames,
		NullValues:       r.opts.NullValues,
	}

	// Use existing parser with target detection
	csvData, categoricalData, targetData, err := types.ParseCSVMixedWithTargets(
		strings.NewReader(recordsToString(records)),
		format,
		nil, // Auto-detect targets based on suffix
	)
	if err != nil {
		return nil, err
	}

	// Convert to unified Data structure
	data := &Data{
		Matrix:               csvData.Matrix,
		Headers:              csvData.Headers,
		RowNames:             csvData.RowNames,
		MissingMask:          csvData.MissingMask,
		Rows:                 csvData.Rows,
		Columns:              csvData.Columns,
		CategoricalColumns:   categoricalData,
		NumericTargetColumns: targetData,
	}

	return data, nil
}

// readStreaming handles streaming read for large files
func (r *Reader) readStreaming(reader *csv.Reader, nullMap map[string]bool) (*Data, error) {
	// Implementation for streaming large files
	// This is a placeholder - full implementation would be more complex
	return nil, fmt.Errorf("streaming mode not yet implemented")
}

// getSelectedColumns returns the indices of columns to parse
func (r *Reader) getSelectedColumns(totalCols int) []int {
	if len(r.opts.Columns) == 0 {
		// Select all columns
		cols := make([]int, totalCols)
		for i := range cols {
			cols[i] = i
		}
		return cols
	}

	// Filter valid column indices
	var selected []int
	for _, col := range r.opts.Columns {
		if col >= 0 && col < totalCols {
			selected = append(selected, col)
		}
	}
	return selected
}

// recordsToString converts records back to CSV string for compatibility
func recordsToString(records [][]string) string {
	var sb strings.Builder
	for _, record := range records {
		for i, field := range record {
			if i > 0 {
				sb.WriteRune(',')
			}
			// Simple CSV escaping
			if strings.ContainsAny(field, ",\"\n") {
				sb.WriteRune('"')
				sb.WriteString(strings.ReplaceAll(field, "\"", "\"\""))
				sb.WriteRune('"')
			} else {
				sb.WriteString(field)
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// ParseFile is a convenience function for simple CSV parsing
func ParseFile(filename string, opts Options) (*Data, error) {
	reader := NewReader(opts)
	return reader.ReadFile(filename)
}

// Parse is a convenience function for parsing CSV from a reader
func Parse(r io.Reader, opts Options) (*Data, error) {
	reader := NewReader(opts)
	return reader.Read(r)
}
