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

	// Try to detect CSV format by attempting different formats

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
	var numericTargetData map[string][]float64
	var lastErr error

	for _, format := range formats {
		// Use the enhanced parser that can handle numeric, categorical, and target columns
		// Columns ending with "#target" (with or without space) are automatically detected as target columns
		data, catData, targetData, err := types.ParseCSVMixedWithTargets(strings.NewReader(content), format, nil)
		if err == nil && data != nil && data.Columns > 0 {
			csvData = data
			categoricalData = catData
			numericTargetData = targetData
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
	if len(categoricalData) > 0 {
		fileResult.CategoricalColumns = categoricalData
	}

	// Add numeric target columns if there are any
	if len(numericTargetData) > 0 {
		fileResult.NumericTargetColumns = numericTargetData
	}

	return fileResult.ToJSONSafe(), nil
}
