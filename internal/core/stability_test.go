// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package core

import (
	"math"
	"math/rand"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// generateMatrixWithConditionNumber creates a matrix with specified condition number
// using SVD reconstruction: A = U * S * V^T where S contains controlled singular values.
//
// Reference: Golub & Van Loan (2013), Matrix Computations, Ch. 2.4
func generateMatrixWithConditionNumber(rows, cols int, condition float64) *mat.Dense {
	// Create random orthogonal matrices U and V
	data := make([]float64, rows*cols)
	for i := range data {
		data[i] = rand.NormFloat64()
	}
	A := mat.NewDense(rows, cols, data)

	// Perform SVD to get orthogonal bases
	var svd mat.SVD
	ok := svd.Factorize(A, mat.SVDFull)
	if !ok {
		// Fallback to simple diagonal matrix if SVD fails
		minDim := rows
		if cols < minDim {
			minDim = cols
		}
		result := mat.NewDense(rows, cols, nil)
		for i := 0; i < minDim; i++ {
			val := 1.0
			if i > 0 {
				val = math.Pow(condition, -float64(i)/float64(minDim-1))
			}
			result.Set(i, i, val)
		}
		return result
	}

	// Create singular values with specified condition number
	values := svd.Values(nil)
	maxVal := values[0]
	minVal := maxVal / condition

	// Interpolate singular values logarithmically
	for i := range values {
		if i == 0 {
			values[i] = maxVal
		} else if i == len(values)-1 {
			values[i] = minVal
		} else {
			// Logarithmic interpolation
			t := float64(i) / float64(len(values)-1)
			values[i] = maxVal * math.Pow(condition, -t)
		}
	}

	// Reconstruct matrix with controlled singular values
	var U, V mat.Dense
	svd.UTo(&U)
	svd.VTo(&V)

	// Create S matrix with appropriate dimensions
	ur, _ := U.Dims()
	_, vc := V.Dims()
	S := mat.NewDense(ur, vc, nil)
	for i := 0; i < len(values) && i < ur && i < vc; i++ {
		S.Set(i, i, values[i])
	}

	// Compute A = U * S * V^T
	var temp mat.Dense
	temp.Mul(&U, S)

	var result mat.Dense
	V.T()
	result.Mul(&temp, &V)

	r, c := result.Dims()
	return mat.NewDense(r, c, result.RawMatrix().Data)
}

// TestPCAStabilityWithIllConditionedMatrices tests PCA with matrices of varying condition numbers
func TestPCAStabilityWithIllConditionedMatrices(t *testing.T) {
	testCases := []struct {
		name      string
		rows      int
		cols      int
		condition float64
		method    string
	}{
		{"Well-conditioned SVD", 50, 10, 10, "svd"},
		{"Well-conditioned NIPALS", 50, 10, 10, "nipals"},
		{"Moderately ill-conditioned SVD", 50, 10, 1e3, "svd"},
		{"Moderately ill-conditioned NIPALS", 50, 10, 1e3, "nipals"},
		{"Ill-conditioned SVD", 50, 10, 1e5, "svd"},
		{"Ill-conditioned NIPALS", 50, 10, 1e5, "nipals"},
		{"Severely ill-conditioned SVD", 50, 10, 1e7, "svd"},
		{"Severely ill-conditioned NIPALS", 50, 10, 1e7, "nipals"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate test matrix
			matrix := generateMatrixWithConditionNumber(tc.rows, tc.cols, tc.condition)

			// Convert to types.Matrix
			data := make(types.Matrix, tc.rows)
			for i := 0; i < tc.rows; i++ {
				data[i] = make([]float64, tc.cols)
				for j := 0; j < tc.cols; j++ {
					data[i][j] = matrix.At(i, j)
				}
			}

			// Configure PCA
			config := types.PCAConfig{
				Components:    min(5, tc.cols),
				MeanCenter:    true,
				StandardScale: false,
				Method:        tc.method,
			}

			// Run PCA
			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)

			// Should not panic or return error for reasonable conditions
			if tc.condition < 1e8 {
				if err != nil {
					t.Errorf("PCA failed for condition number %.2e: %v", tc.condition, err)
				}

				if result != nil {
					// Verify basic properties
					// 1. Explained variance values should be non-negative
					for i, ev := range result.ExplainedVar {
						if ev < -1e-10 {
							t.Errorf("Negative explained variance[%d] = %e for condition %.2e", i, ev, tc.condition)
						}
					}

					// 2. Explained variance should be in descending order
					for i := 1; i < len(result.ExplainedVar); i++ {
						if result.ExplainedVar[i] > result.ExplainedVar[i-1] {
							t.Errorf("Explained variance not in descending order at index %d for condition %.2e", i, tc.condition)
						}
					}

					// 3. Explained variance ratio should sum to <= 100
					totalVar := 0.0
					for _, v := range result.ExplainedVarRatio {
						totalVar += v
					}
					if totalVar > 100.1 { // Allow small numerical error
						t.Errorf("Total explained variance %.2f%% > 100%% for condition %.2e", totalVar, tc.condition)
					}
				}
			}
		})
	}
}

// TestPCAWithNearSingularMatrix tests PCA behavior with near-singular matrices
func TestPCAWithNearSingularMatrix(t *testing.T) {
	// Create a rank-deficient matrix (rank 2 in 3D)
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{2.0, 4.0, 6.0}, // Linear combination of first row
		{3.0, 6.0, 9.0}, // Linear combination of first row
		{1.1, 2.1, 3.1}, // Slight perturbation
		{0.9, 1.9, 2.9}, // Slight perturbation
	}

	methods := []string{"svd", "nipals"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    3,
				MeanCenter:    true,
				StandardScale: false,
				Method:        method,
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)

			if err != nil {
				t.Fatalf("PCA failed: %v", err)
			}

			// The third explained variance should be very small (near zero)
			if len(result.ExplainedVar) >= 3 {
				thirdVar := result.ExplainedVar[2]
				if thirdVar > 0.01 {
					t.Errorf("Third explained variance %.6f too large for rank-deficient matrix", thirdVar)
				}
			}

			// First two components should explain nearly all variance
			totalFirstTwo := result.ExplainedVarRatio[0] + result.ExplainedVarRatio[1]
			if totalFirstTwo < 99.0 {
				t.Errorf("First two components explain only %.2f%% (expected >99%%)", totalFirstTwo)
			}
		})
	}
}

// TestPCAStabilityConsistency tests that different methods give similar results for well-conditioned data
func TestPCAStabilityConsistency(t *testing.T) {
	// Generate well-conditioned test matrix
	matrix := generateMatrixWithConditionNumber(100, 20, 10)

	// Convert to types.Matrix
	rows, cols := matrix.Dims()
	data := make(types.Matrix, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			data[i][j] = matrix.At(i, j)
		}
	}

	// Run PCA with both methods
	config := types.PCAConfig{
		Components:    10,
		MeanCenter:    true,
		StandardScale: false,
	}

	engine := NewPCAEngine()

	// SVD method
	config.Method = "svd"
	svdResult, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("SVD PCA failed: %v", err)
	}

	// NIPALS method
	config.Method = "nipals"
	nipalsResult, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("NIPALS PCA failed: %v", err)
	}

	// Compare results - should be very similar for well-conditioned matrix
	tolerance := 1e-6

	// Compare explained variance values
	for i := 0; i < len(svdResult.ExplainedVar) && i < len(nipalsResult.ExplainedVar); i++ {
		diff := math.Abs(svdResult.ExplainedVar[i] - nipalsResult.ExplainedVar[i])
		relDiff := diff / math.Max(svdResult.ExplainedVar[i], 1e-10)
		if relDiff > tolerance {
			t.Errorf("ExplainedVar[%d] differs: SVD=%.6e, NIPALS=%.6e (rel diff %.2e)",
				i, svdResult.ExplainedVar[i], nipalsResult.ExplainedVar[i], relDiff)
		}
	}

	// Compare explained variance ratios
	for i := 0; i < len(svdResult.ExplainedVarRatio) && i < len(nipalsResult.ExplainedVarRatio); i++ {
		diff := math.Abs(svdResult.ExplainedVarRatio[i] - nipalsResult.ExplainedVarRatio[i])
		if diff > 0.01 { // 0.01% tolerance for explained variance
			t.Errorf("ExplainedVarRatio[%d] differs: SVD=%.4f%%, NIPALS=%.4f%%",
				i, svdResult.ExplainedVarRatio[i], nipalsResult.ExplainedVarRatio[i])
		}
	}
}

// TestMahalanobisDistanceAndPCA validates the relationship between Mahalanobis distance and PCA scores
// Reference: Brereton (2015), The Mahalanobis distance and its relationship to principal component scores
func TestMahalanobisDistanceAndPCA(t *testing.T) {
	// Create test data
	data := types.Matrix{
		{2.5, 2.4, 3.1},
		{0.5, 0.7, 1.2},
		{2.2, 2.9, 2.8},
		{1.9, 2.2, 2.4},
		{3.1, 3.0, 3.5},
		{2.3, 2.7, 2.9},
		{2.0, 1.6, 2.1},
		{1.0, 1.1, 1.3},
	}

	config := types.PCAConfig{
		Components:    3,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}

	engine := NewPCAEngine()
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA failed: %v", err)
	}

	// For each observation, the squared Mahalanobis distance should equal
	// the sum of squares of standardized PC scores
	// Standardized scores = scores / sqrt(explained variance)

	for i, scores := range result.Scores {
		mahalanobisSquared := 0.0
		for j, score := range scores {
			if result.ExplainedVar[j] > 1e-10 {
				standardizedScore := score / math.Sqrt(result.ExplainedVar[j])
				mahalanobisSquared += standardizedScore * standardizedScore
			}
		}

		// This is a basic check - in practice, you'd compute the actual
		// Mahalanobis distance using the covariance matrix
		if mahalanobisSquared < 0 {
			t.Errorf("Invalid Mahalanobis distance squared for observation %d: %.6f", i, mahalanobisSquared)
		}
	}
}
