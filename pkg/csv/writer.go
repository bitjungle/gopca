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

	"github.com/bitjungle/gopca/pkg/types"
)

// NewWriter creates a new CSV writer with the given options
func NewWriter(opts Options) *Writer {
	return &Writer{opts: opts}
}

// WriteFile writes CSV data to a file
func (w *Writer) WriteFile(filename string, data *Data) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return w.Write(file, data)
}

// Write writes CSV data to an io.Writer
func (w *Writer) Write(output io.Writer, data *Data) error {
	writer := csv.NewWriter(output)
	writer.Comma = w.opts.Delimiter
	defer writer.Flush()

	// Determine what type of data to write
	if data.StringData != nil && len(data.StringData) > 0 {
		return w.writeStringData(writer, data)
	}
	return w.writeNumericData(writer, data)
}

// writeNumericData writes numeric matrix data
func (w *Writer) writeNumericData(writer *csv.Writer, data *Data) error {
	// Write headers
	if w.opts.HasHeaders && len(data.Headers) > 0 {
		headers := data.Headers
		if w.opts.HasRowNames && len(data.RowNames) > 0 {
			// Add empty header for row names column
			headers = append([]string{""}, headers...)
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Write data rows
	for i, row := range data.Matrix {
		record := make([]string, 0, len(row)+1)

		// Add row name if present
		if w.opts.HasRowNames && i < len(data.RowNames) {
			record = append(record, data.RowNames[i])
		}

		// Convert numeric values to strings
		for j, val := range row {
			var str string

			// Check missing mask if available
			if data.MissingMask != nil && i < len(data.MissingMask) &&
				j < len(data.MissingMask[i]) && data.MissingMask[i][j] {
				// Use first null value from options or "NA"
				if len(w.opts.NullValues) > 0 {
					str = w.opts.NullValues[0]
				} else {
					str = "NA"
				}
			} else if math.IsNaN(val) {
				str = "NaN"
			} else if math.IsInf(val, 1) {
				str = "Inf"
			} else if math.IsInf(val, -1) {
				str = "-Inf"
			} else {
				// Format the float value
				if w.opts.Precision >= 0 {
					str = strconv.FormatFloat(val, w.opts.FloatFormat, w.opts.Precision, 64)
				} else {
					str = strconv.FormatFloat(val, w.opts.FloatFormat, -1, 64)
				}

				// Handle decimal separator for European format
				if w.opts.DecimalSeparator == ',' {
					str = replaceDecimalSeparator(str, '.', ',')
				}
			}

			record = append(record, str)
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+1, err)
		}
	}

	return nil
}

// writeStringData writes string matrix data (for GoCSV)
func (w *Writer) writeStringData(writer *csv.Writer, data *Data) error {
	// Write headers
	if w.opts.HasHeaders && len(data.Headers) > 0 {
		headers := data.Headers
		if w.opts.HasRowNames && len(data.RowNames) > 0 {
			// Add empty header for row names column
			headers = append([]string{""}, headers...)
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Write string data rows
	for i, row := range data.StringData {
		record := make([]string, 0, len(row)+1)

		// Add row name if present
		if w.opts.HasRowNames && i < len(data.RowNames) {
			record = append(record, data.RowNames[i])
		}

		// Append string values
		record = append(record, row...)

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+1, err)
		}
	}

	return nil
}

// WriteMatrix writes a numeric matrix to CSV
func (w *Writer) WriteMatrix(output io.Writer, matrix types.Matrix, headers []string, rowNames []string) error {
	data := &Data{
		Matrix:   matrix,
		Headers:  headers,
		RowNames: rowNames,
		Rows:     len(matrix),
	}
	if len(matrix) > 0 {
		data.Columns = len(matrix[0])
	}

	return w.Write(output, data)
}

// WriteMatrixFile writes a numeric matrix to a CSV file
func (w *Writer) WriteMatrixFile(filename string, matrix types.Matrix, headers []string, rowNames []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return w.WriteMatrix(file, matrix, headers, rowNames)
}

// replaceDecimalSeparator replaces decimal separators in a string
func replaceDecimalSeparator(s string, old, new rune) string {
	runes := []rune(s)
	for i, r := range runes {
		if r == old {
			runes[i] = new
		}
	}
	return string(runes)
}

// SaveFile is a convenience function for simple CSV writing
func SaveFile(filename string, data *Data, opts Options) error {
	writer := NewWriter(opts)
	return writer.WriteFile(filename, data)
}

// Save is a convenience function for writing CSV to an io.Writer
func Save(w io.Writer, data *Data, opts Options) error {
	writer := NewWriter(opts)
	return writer.Write(w, data)
}

// SaveMatrix is a convenience function for writing a matrix to CSV
func SaveMatrix(filename string, matrix types.Matrix, headers []string, rowNames []string, opts Options) error {
	writer := NewWriter(opts)
	return writer.WriteMatrixFile(filename, matrix, headers, rowNames)
}
