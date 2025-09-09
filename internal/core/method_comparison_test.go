// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compareVectorsWithSign compares two vectors accounting for sign ambiguity
// PCA components can have arbitrary sign, so we check both orientations
func compareVectorsWithSign(t *testing.T, vec1, vec2 []float64, tolerance float64, description string) {
	t.Helper()

	if len(vec1) != len(vec2) {
		t.Errorf("%s: vectors have different lengths (%d vs %d)", description, len(vec1), len(vec2))
		return
	}

	// Calculate correlation to determine sign
	correlation := 0.0
	for i := range vec1 {
		correlation += vec1[i] * vec2[i]
	}

	// If correlation is negative, flip sign
	sign := 1.0
	if correlation < 0 {
		sign = -1.0
	}

	// Compare element by element with sign adjustment
	for i := range vec1 {
		expected := sign * vec2[i]
		assert.InDelta(t, vec1[i], expected, tolerance,
			"%s: element %d differs", description, i)
	}
}

// compareMatricesWithSign compares two matrices accounting for sign ambiguity in columns
func compareMatricesWithSign(t *testing.T, mat1, mat2 [][]float64, tolerance float64, description string) {
	t.Helper()

	if len(mat1) != len(mat2) {
		t.Errorf("%s: matrices have different row counts (%d vs %d)", description, len(mat1), len(mat2))
		return
	}

	if len(mat1) == 0 || len(mat1[0]) != len(mat2[0]) {
		t.Errorf("%s: matrices have different column counts", description)
		return
	}

	nCols := len(mat1[0])

	// Check each column independently for sign
	for j := 0; j < nCols; j++ {
		// Calculate correlation for this column
		correlation := 0.0
		for i := range mat1 {
			correlation += mat1[i][j] * mat2[i][j]
		}

		sign := 1.0
		if correlation < 0 {
			sign = -1.0
		}

		// Compare this column with sign adjustment
		for i := range mat1 {
			expected := sign * mat2[i][j]
			assert.InDelta(t, mat1[i][j], expected, tolerance,
				"%s: element [%d,%d] differs", description, i, j)
		}
	}
}

// calculateConditionNumber estimates the condition number of a matrix
// This is useful for determining appropriate tolerances
func calculateConditionNumber(data [][]float64) float64 {
	if len(data) == 0 || len(data[0]) == 0 {
		return math.Inf(1)
	}

	// For simplicity, estimate using the ratio of max to min non-zero singular values
	// In a real implementation, you'd compute actual singular values
	
	// Find max and min absolute values as a rough approximation
	maxVal := 0.0
	minVal := math.Inf(1)

	for i := range data {
		for j := range data[i] {
			absVal := math.Abs(data[i][j])
			if absVal > maxVal {
				maxVal = absVal
			}
			if absVal > 0 && absVal < minVal {
				minVal = absVal
			}
		}
	}

	if minVal == 0 || minVal == math.Inf(1) {
		return math.Inf(1)
	}

	// Very rough approximation - in practice you'd use SVD
	return maxVal / minVal
}

// getComparisonTolerance returns appropriate tolerance based on condition number
func getComparisonTolerance(condition float64) float64 {
	switch {
	case condition < 10:
		return 1e-10 // Well-conditioned
	case condition < 100:
		return 1e-8  // Moderately conditioned
	case condition < 1000:
		return 1e-6  // Poorly conditioned
	case condition < 10000:
		return 1e-4  // Very poorly conditioned
	default:
		return 1e-2  // Ill-conditioned
	}
}

// verifyOrthogonality checks if columns of a matrix are orthogonal
func verifyOrthogonality(t *testing.T, matrix [][]float64, tolerance float64, description string) {
	t.Helper()

	if len(matrix) == 0 || len(matrix[0]) < 2 {
		return // Nothing to check
	}

	nCols := len(matrix[0])

	for i := 0; i < nCols; i++ {
		for j := i + 1; j < nCols; j++ {
			dotProduct := 0.0
			for k := range matrix {
				dotProduct += matrix[k][i] * matrix[k][j]
			}

			assert.InDelta(t, 0.0, dotProduct, tolerance,
				"%s: columns %d and %d should be orthogonal", description, i, j)
		}
	}
}

// verifyUnitNorm checks if columns of a matrix have unit norm
func verifyUnitNorm(t *testing.T, matrix [][]float64, tolerance float64, description string) {
	t.Helper()

	if len(matrix) == 0 {
		return
	}

	nCols := len(matrix[0])

	for j := 0; j < nCols; j++ {
		norm := 0.0
		for i := range matrix {
			norm += matrix[i][j] * matrix[i][j]
		}
		norm = math.Sqrt(norm)

		assert.InDelta(t, 1.0, norm, tolerance,
			"%s: column %d should have unit norm", description, j)
	}
}

// compareExplainedVariance compares explained variance arrays with appropriate tolerance
func compareExplainedVariance(t *testing.T, var1, var2 []float64, tolerance float64, description string) {
	t.Helper()

	assert.Equal(t, len(var1), len(var2), "%s: different number of components", description)

	for i := range var1 {
		// Use relative tolerance for larger values
		relTolerance := tolerance
		if var1[i] > 1.0 {
			relTolerance = tolerance * var1[i]
		}

		assert.InDelta(t, var1[i], var2[i], relTolerance,
			"%s: explained variance for component %d", description, i+1)
	}
}

// matrixDifference calculates the Frobenius norm of the difference between two matrices
func matrixDifference(mat1, mat2 [][]float64) float64 {
	if len(mat1) != len(mat2) || (len(mat1) > 0 && len(mat1[0]) != len(mat2[0])) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := range mat1 {
		for j := range mat1[i] {
			diff := mat1[i][j] - mat2[i][j]
			sum += diff * diff
		}
	}

	return math.Sqrt(sum)
}