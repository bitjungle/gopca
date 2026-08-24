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

package cobra

import (
	"fmt"
	"math"
	"strings"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"gonum.org/v1/gonum/mat"
)

// ProjectData projects data onto principal components using loadings.
//
// The loadings must be a dense [nFeatures][nComponents] matrix matching the
// width of the data. A model file can fail that — kernel PCA stores no loadings
// at all, and temporal PCA stores them over the lagged embedding rather than the
// original variables — so the shape is checked rather than assumed. Model files
// can also be hand-edited or produced elsewhere, and a panic is never the right
// answer to one that is malformed (#809).
func ProjectData(data, loadings [][]float64) ([][]float64, error) {
	nSamples := len(data)
	if nSamples == 0 {
		return nil, fmt.Errorf("no data rows to project")
	}
	nFeatures := len(data[0])
	if nFeatures == 0 {
		return nil, fmt.Errorf("data rows have no columns")
	}
	if len(loadings) == 0 {
		return nil, fmt.Errorf("the model contains no loadings, so data cannot be projected onto its components")
	}
	if len(loadings) != nFeatures {
		return nil, fmt.Errorf("the model expects %d variables but the data has %d", len(loadings), nFeatures)
	}
	nComponents := len(loadings[0])
	if nComponents == 0 {
		return nil, fmt.Errorf("the model contains no components")
	}
	for i, row := range loadings {
		if len(row) != nComponents {
			return nil, fmt.Errorf("loadings row %d has %d components, expected %d", i, len(row), nComponents)
		}
	}
	for i, row := range data {
		if len(row) != nFeatures {
			return nil, fmt.Errorf("data row %d has %d columns, expected %d", i, len(row), nFeatures)
		}
	}

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

	return result, nil
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

	fmt.Fprintf(&sb, "Data dimensions: %d rows × %d columns\n", data.Rows, data.Columns)

	if len(data.Headers) > 0 {
		fmt.Fprintf(&sb, "Column names: %s", strings.Join(data.Headers[:min(5, len(data.Headers))], ", "))
		if len(data.Headers) > 5 {
			fmt.Fprintf(&sb, " ... (showing first 5 of %d)\n", len(data.Headers))
		} else {
			sb.WriteString("\n")
		}
	}

	if len(data.RowNames) > 0 {
		fmt.Fprintf(&sb, "Row names: %s", strings.Join(data.RowNames[:min(5, len(data.RowNames))], ", "))
		if len(data.RowNames) > 5 {
			fmt.Fprintf(&sb, " ... (showing first 5 of %d)\n", len(data.RowNames))
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
	fmt.Fprintf(&sb, "Missing values: %d (%.1f%%)\n", missingCount, missingPercent)

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
