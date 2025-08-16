// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"fmt"
	"math"
	"strings"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"gonum.org/v1/gonum/mat"
)

// ProjectData projects data onto principal components using loadings
func ProjectData(data, loadings [][]float64) [][]float64 {
	// Convert to matrices for computation
	nSamples := len(data)
	nFeatures := len(data[0])
	nComponents := len(loadings[0])

	// Create data matrix
	dataFlat := make([]float64, nSamples*nFeatures)
	for i := 0; i < nSamples; i++ {
		for j := 0; j < nFeatures; j++ {
			dataFlat[i*nFeatures+j] = data[i][j]
		}
	}
	X := mat.NewDense(nSamples, nFeatures, dataFlat)

	// Create loadings matrix
	loadingsFlat := make([]float64, nFeatures*nComponents)
	for i := 0; i < nFeatures; i++ {
		for j := 0; j < nComponents; j++ {
			loadingsFlat[i*nComponents+j] = loadings[i][j]
		}
	}
	L := mat.NewDense(nFeatures, nComponents, loadingsFlat)

	// Project: scores = X * L
	scores := mat.NewDense(nSamples, nComponents, nil)
	scores.Mul(X, L)

	// Convert back to [][]float64
	result := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		result[i] = make([]float64, nComponents)
		for j := 0; j < nComponents; j++ {
			result[i][j] = scores.At(i, j)
		}
	}

	return result
}

// validateCSVData performs basic validation on parsed CSV data
func validateCSVData(data *pkgcsv.Data) error {
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

// getDataSummary returns a summary of the CSV data
func getDataSummary(data *pkgcsv.Data) string {
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
