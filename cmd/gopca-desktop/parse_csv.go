package main

import (
	"fmt"
	"strings"
	
	"github.com/bitjungle/gopca/pkg/types"
)

// ParseCSV parses CSV content and returns data matrix and headers
func (a *App) ParseCSV(content string) (*FileData, error) {
	// Try to detect CSV format
	sample := []byte(content)
	if len(sample) > 1000 {
		sample = sample[:1000] // Use first 1KB for detection
	}
	
	format, err := types.DetectFormat(sample)
	if err != nil {
		// Use default format if detection fails
		defaultFormat := types.DefaultCSVFormat()
		format = &defaultFormat
	}
	
	// Parse using unified parser
	parser := types.NewCSVParser(*format)
	csvData, err := parser.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	
	if csvData.Rows < 1 {
		return nil, fmt.Errorf("CSV file must have at least one data row")
	}
	
	// Detect categorical columns
	categoricalData := make(map[string][]string)
	numericHeaders := []string{}
	numericData := [][]float64{}
	
	// For now, include all columns as numeric if they contain valid float data
	// The unified parser already handles NaN values appropriately
	for j := 0; j < csvData.Columns; j++ {
		isNumeric := true
		hasSomeNumeric := false
		
		// Check if column has at least some numeric values
		for i := 0; i < csvData.Rows && i < 10; i++ {
			if !csvData.MissingMask[i][j] {
				hasSomeNumeric = true
				break
			}
		}
		
		if hasSomeNumeric {
			// Include this as a numeric column
			numericHeaders = append(numericHeaders, csvData.Headers[j])
		} else {
			// Treat as categorical - extract string values
			isNumeric = false
			colName := csvData.Headers[j]
			categoricalData[colName] = make([]string, csvData.Rows)
			// Note: For now, we'll skip categorical columns in the GUI
			// This can be enhanced later to properly handle categorical data
		}
		
		if !isNumeric {
			continue
		}
	}
	
	// Build numeric data matrix with all columns
	// The PCA engine will validate and handle any remaining NaN values
	numericData = csvData.Matrix
	
	result := &FileData{
		Headers:  csvData.Headers,
		RowNames: csvData.RowNames,
		Data:     numericData,
	}
	
	// Only add categorical columns if there are any
	if len(categoricalData) > 0 {
		result.CategoricalColumns = categoricalData
	}
	
	return result, nil
}