// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"fmt"
	"strings"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
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

	// Try multiple formats
	formats := []pkgcsv.Options{
		pkgcsv.DefaultOptions(),      // Comma with dot decimal
		pkgcsv.EuropeanOptions(),     // Semicolon with comma decimal
		pkgcsv.TabDelimitedOptions(), // Tab delimited
	}

	var csvData *pkgcsv.Data
	var lastErr error

	for _, opts := range formats {
		// Use ParseMixedWithTargets mode to detect all column types
		opts.ParseMode = pkgcsv.ParseMixedWithTargets

		reader := pkgcsv.NewReader(opts)
		data, err := reader.Read(strings.NewReader(content))
		if err == nil && data != nil && data.Columns > 0 {
			csvData = data
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
	if len(csvData.CategoricalColumns) > 0 {
		fileResult.CategoricalColumns = csvData.CategoricalColumns
	}

	// Add numeric target columns if there are any
	if len(csvData.NumericTargetColumns) > 0 {
		fileResult.NumericTargetColumns = csvData.NumericTargetColumns
	}

	return fileResult.ToJSONSafe(), nil
}
