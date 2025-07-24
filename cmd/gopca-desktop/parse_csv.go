//go:build desktop || wails
// +build desktop wails

package main

import (
	"fmt"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// ParseCSV parses CSV content and returns data matrix and headers
func (a *App) ParseCSV(content string) (result *FileDataJSON, err error) {
	// Recover from any panic to prevent app crash
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected error while parsing file: %v", r)
			result = nil
		}
	}()

	// Validate input
	if content == "" {
		return nil, fmt.Errorf("empty file content")
	}

	// Try to detect CSV format
	sample := []byte(content)
	if len(sample) > 1000 {
		sample = sample[:1000] // Use first 1KB for detection
	}

	// Try multiple formats
	defaultFormat := types.DefaultCSVFormat()
	formats := []types.CSVFormat{
		defaultFormat, // Comma with dot decimal
		{
			FieldDelimiter:   ';',
			DecimalSeparator: ',',
			HasHeaders:       true,
			HasRowNames:      true,
			NullValues:       defaultFormat.NullValues,
		},
		{
			FieldDelimiter:   '\t',
			DecimalSeparator: '.',
			HasHeaders:       true,
			HasRowNames:      true,
			NullValues:       defaultFormat.NullValues,
		},
	}

	var csvData *types.CSVData
	var categoricalData map[string][]string
	var lastErr error

	for _, format := range formats {
		// Use the mixed parser that can handle both numeric and categorical columns
		data, catData, err := types.ParseCSVMixed(strings.NewReader(content), format)
		if err == nil && data != nil && data.Columns > 0 {
			csvData = data
			categoricalData = catData
			break
		}
		if err != nil {
			lastErr = err
		}
	}

	if csvData == nil {
		if lastErr != nil {
			return nil, fmt.Errorf("invalid file format: %w", lastErr)
		}
		return nil, fmt.Errorf("no numeric data columns found in file")
	}

	fileResult := &FileData{
		Headers:     csvData.Headers,
		RowNames:    csvData.RowNames,
		Data:        csvData.Matrix,
		MissingMask: csvData.MissingMask,
	}

	// Add categorical columns if there are any
	if categoricalData != nil && len(categoricalData) > 0 {
		fileResult.CategoricalColumns = categoricalData
	}

	return fileResult.ToJSONSafe(), nil
}
