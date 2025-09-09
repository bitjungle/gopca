// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"fmt"
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
)

// compareMethodResults compares two PCA results with tolerance-aware checking.
// It handles sign ambiguity in eigenvectors and accounts for numerical differences.
//
// Reference: Golub, G. H., & Van Loan, C. F. (2013).
// Matrix computations (4th ed.). Johns Hopkins University Press.
func compareMethodResults(result1, result2 *types.PCAResult, tolerance float64, description string) error {
	// Check dimensions match
	if result1.ComponentsComputed != result2.ComponentsComputed {
		return fmt.Errorf("%s: component count mismatch: %d vs %d",
			description, result1.ComponentsComputed, result2.ComponentsComputed)
	}

	// Compare explained variance
	if err := compareVectors(result1.ExplainedVar, result2.ExplainedVar,
		tolerance, fmt.Sprintf("%s: explained variance", description)); err != nil {
		return err
	}

	// Compare cumulative variance
	if err := compareVectors(result1.CumulativeVar, result2.CumulativeVar,
		tolerance, fmt.Sprintf("%s: cumulative variance", description)); err != nil {
		return err
	}

	// Compare scores (handle sign ambiguity)
	scores1Aligned := resolveSignAmbiguity(result1.Scores, result2.Scores)
	if err := compareMatrices(scores1Aligned, result2.Scores,
		tolerance, fmt.Sprintf("%s: scores", description)); err != nil {
		return err
	}

	// Compare loadings (handle sign ambiguity)
	loadings1Aligned := resolveSignAmbiguity(result1.Loadings, result2.Loadings)
	if err := compareMatrices(loadings1Aligned, result2.Loadings,
		tolerance, fmt.Sprintf("%s: loadings", description)); err != nil {
		return err
	}

	return nil
}

// getMethodComparisonTolerance returns appropriate tolerance based on matrix condition number
// and the methods being compared.
//
// Algorithm complexity: O(1)
func getMethodComparisonTolerance(condition float64, method1, method2 string) float64 {
	// Base tolerance from condition number
	baseTolerance := getToleranceForCondition(condition)

	// Adjust for specific method pairs
	if (method1 == "nipals" || method2 == "nipals") && condition > 1000 {
		// NIPALS may have larger errors for ill-conditioned matrices
		return baseTolerance * 10
	}

	if method1 == "kernel" || method2 == "kernel" {
		// Kernel methods may have additional numerical errors
		return baseTolerance * 5
	}

	if method1 == "temporal" || method2 == "temporal" {
		// Temporal methods involve additional matrix operations
		return baseTolerance * 3
	}

	return baseTolerance
}

// verifyOrthogonality checks if the columns of a matrix are orthonormal.
//
// Mathematical property: For orthonormal matrix Q, Q^T * Q = I
// Reference: Golub & Van Loan (2013), Section 2.2
func verifyOrthogonality(t *testing.T, matrix types.Matrix, tolerance float64, description string) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return
	}

	n := len(matrix)
	p := len(matrix[0])

	// Check orthogonality between all pairs of columns
	for i := 0; i < p; i++ {
		for j := i; j < p; j++ {
			dotProduct := 0.0
			for k := 0; k < n; k++ {
				dotProduct += matrix[k][i] * matrix[k][j]
			}

			expectedValue := 0.0
			if i == j {
				expectedValue = 1.0 // Columns should be normalized
			}

			assert.InDelta(t, expectedValue, dotProduct, tolerance,
				"%s: columns %d and %d orthogonality check", description, i, j)
		}
	}
}

// documentExpectedDivergence logs and documents expected differences between methods
// for specific conditions.
func documentExpectedDivergence(condition float64, method1, method2 string, actualDifference float64) string {
	var reason string

	if condition > 10000 {
		reason = fmt.Sprintf("Extremely ill-conditioned matrix (κ=%.1e)", condition)
	} else if condition > 1000 {
		reason = fmt.Sprintf("Ill-conditioned matrix (κ=%.1f)", condition)
	} else if method1 == "nipals" || method2 == "nipals" {
		if actualDifference > 1e-6 {
			reason = "NIPALS iterative convergence vs direct decomposition"
		}
	} else if (method1 == "kernel" && method2 != "kernel") ||
		(method2 == "kernel" && method1 != "kernel") {
		reason = "Kernel space transformation numerical differences"
	} else if (method1 == "temporal" && method2 != "temporal") ||
		(method2 == "temporal" && method1 != "temporal") {
		reason = "Temporal embedding vs direct analysis"
	}

	if reason != "" {
		return fmt.Sprintf("Expected divergence between %s and %s: %s (difference: %.2e)",
			method1, method2, reason, actualDifference)
	}

	return ""
}

// resolveMethodDifferences provides detailed explanation of why two methods
// might produce different results for specific data characteristics.
func resolveMethodDifferences(method1, method2 string, dataCharacteristics map[string]interface{}) string {
	condition, hasCondition := dataCharacteristics["condition"].(float64)
	hasMissing, _ := dataCharacteristics["has_missing"].(bool)
	isTimeSeries, _ := dataCharacteristics["is_timeseries"].(bool)

	var explanations []string

	// SVD vs NIPALS differences
	if (method1 == "svd" && method2 == "nipals") || (method1 == "nipals" && method2 == "svd") {
		if hasCondition && condition > 1000 {
			explanations = append(explanations,
				"SVD uses direct decomposition while NIPALS uses iterative extraction, "+
					"leading to accumulation of numerical errors for ill-conditioned matrices")
		}
		if hasMissing {
			explanations = append(explanations,
				"NIPALS can handle missing values natively while SVD requires imputation")
		}
	}

	// Kernel PCA differences
	if method1 == "kernel" || method2 == "kernel" {
		explanations = append(explanations,
			"Kernel PCA operates in feature space via kernel trick, "+
				"introducing additional numerical operations")
		if method1 == "kernel" && method2 == "kernel" {
			explanations = append(explanations,
				"Different kernel parameters can lead to different feature spaces")
		}
	}

	// Temporal PCA differences
	if method1 == "temporal" || method2 == "temporal" {
		if isTimeSeries {
			explanations = append(explanations,
				"Temporal PCA uses SSA decomposition on trajectory matrix, "+
					"which is fundamentally different from standard PCA on raw data")
		}
		explanations = append(explanations,
			"Window length and embedding dimension affect the decomposition")
	}

	if len(explanations) == 0 {
		return "Methods should produce equivalent results for this data"
	}

	result := "Method differences explained:\n"
	for i, exp := range explanations {
		result += fmt.Sprintf("  %d. %s\n", i+1, exp)
	}
	return result
}

// calculateMaxDifference computes the maximum absolute difference between two matrices.
//
// Algorithm complexity: O(n*m) where n,m are matrix dimensions
func calculateMaxDifference(matrix1, matrix2 types.Matrix) (float64, error) {
	if len(matrix1) != len(matrix2) {
		return 0, fmt.Errorf("matrix dimension mismatch: %d vs %d rows",
			len(matrix1), len(matrix2))
	}

	maxDiff := 0.0
	for i := range matrix1 {
		if len(matrix1[i]) != len(matrix2[i]) {
			return 0, fmt.Errorf("matrix dimension mismatch at row %d: %d vs %d columns",
				i, len(matrix1[i]), len(matrix2[i]))
		}

		for j := range matrix1[i] {
			diff := math.Abs(matrix1[i][j] - matrix2[i][j])
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	return maxDiff, nil
}

// validateMethodSelection provides guidance on which PCA method to use
// based on data characteristics.
func validateMethodSelection(dataCharacteristics map[string]interface{}) string {
	hasMissing, _ := dataCharacteristics["has_missing"].(bool)
	isLarge, _ := dataCharacteristics["is_large"].(bool)
	isTimeSeries, _ := dataCharacteristics["is_timeseries"].(bool)
	isNonlinear, _ := dataCharacteristics["is_nonlinear"].(bool)
	condition, hasCondition := dataCharacteristics["condition"].(float64)

	recommendations := []string{}

	if hasMissing {
		recommendations = append(recommendations,
			"NIPALS: Best for data with missing values (native handling)")
	}

	if isTimeSeries {
		recommendations = append(recommendations,
			"Temporal PCA: Designed for time series analysis (SSA decomposition)")
	}

	if isNonlinear {
		recommendations = append(recommendations,
			"Kernel PCA: Can capture nonlinear relationships")
	}

	if !hasMissing && !isTimeSeries && !isNonlinear {
		if isLarge {
			recommendations = append(recommendations,
				"SVD: Fast and efficient for large complete datasets")
		} else if hasCondition && condition > 1000 {
			recommendations = append(recommendations,
				"SVD with careful preprocessing: Direct method more stable for ill-conditioned matrices")
		} else {
			recommendations = append(recommendations,
				"SVD: Standard choice for complete, well-conditioned data")
		}
	}

	if len(recommendations) == 0 {
		return "SVD: Default choice for most datasets"
	}

	result := "Method selection recommendations:\n"
	for _, rec := range recommendations {
		result += fmt.Sprintf("  • %s\n", rec)
	}
	return result
}
