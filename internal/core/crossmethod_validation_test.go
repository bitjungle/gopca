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

package core

import (
	"math"
	"math/rand"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"
)

var rng = rand.New(rand.NewSource(42)) // Deterministic random for tests

// TestSVDvsNIPALS verifies that SVD and NIPALS produce consistent results
// for well-conditioned and moderately ill-conditioned matrices.
//
// Reference: Bro, R., & Smilde, A. K. (2014). Principal component analysis.
// Analytical Methods, 6(9), 2812-2831.
func TestSVDvsNIPALS(t *testing.T) {
	tests := []struct {
		name      string
		dataFile  string
		condition float64 // 0 means use real data
		tolerance float64
	}{
		{
			name:      "iris well-conditioned",
			dataFile:  "../../testdata/iris/iris.csv",
			condition: 0,
			tolerance: 1e-6,
		},
		{
			name:      "wine dataset",
			dataFile:  "../../testdata/wine/wine.csv",
			condition: 0,
			tolerance: 1e-5, // Wine has higher condition number
		},
		{
			name:      "synthetic κ=10",
			dataFile:  "",
			condition: 10,
			tolerance: 1e-7,
		},
		{
			name:      "synthetic κ=100",
			dataFile:  "",
			condition: 100,
			tolerance: 1e-6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data types.Matrix

			if tt.dataFile != "" {
				// Load real data using the existing function from sklearn_validation_test.go
				var err error
				data, err = loadTestDataAsMatrix(tt.dataFile)
				require.NoError(t, err)
			} else {
				// Generate synthetic data with specified condition number
				data = generateConditionedMatrix(50, 10, tt.condition)
			}

			// Run SVD
			config := types.PCAConfig{
				Components:    minInt(3, len(data[0])),
				MeanCenter:    true,
				StandardScale: true,
				Method:        "svd",
			}

			svdEngine := NewPCAEngine()
			svdResult, err := svdEngine.Fit(data, config)
			require.NoError(t, err)

			// Run NIPALS
			config.Method = "nipals"
			nipalsEngine := NewPCAEngine()
			nipalsResult, err := nipalsEngine.Fit(data, config)
			require.NoError(t, err)

			// Compare results
			compareMethodResults(t, svdResult, nipalsResult, tt.tolerance, "SVD vs NIPALS")
		})
	}
}

// TestPreprocessingConsistency verifies that different preprocessing methods
// are applied consistently across different PCA methods
func TestPreprocessingConsistency(t *testing.T) {
	// Load test data
	data, err := loadTestDataAsMatrix("../../testdata/iris/iris.csv")
	require.NoError(t, err)

	preprocessingTypes := []struct {
		name          string
		meanCenter    bool
		standardScale bool
	}{
		{"mean_center", true, false},
		{"standardize", true, true},
	}

	methods := []string{"svd", "nipals"}

	for _, prep := range preprocessingTypes {
		t.Run(prep.name, func(t *testing.T) {
			var results []*types.PCAResult

			for _, method := range methods {
				config := types.PCAConfig{
					Components:    3,
					MeanCenter:    prep.meanCenter,
					StandardScale: prep.standardScale,
					Method:        method,
				}

				engine := NewPCAEngine()
				result, err := engine.Fit(data, config)
				require.NoError(t, err, "Failed with method %s", method)
				results = append(results, result)
			}

			// All methods should have the same preprocessing parameters
			for i := 1; i < len(results); i++ {
				if results[0].Means != nil && results[i].Means != nil {
					for j := range results[0].Means {
						assert.InDelta(t, results[0].Means[j], results[i].Means[j], 1e-10,
							"Mean[%d] should be identical across methods", j)
					}
				}

				if prep.standardScale && results[0].StdDevs != nil && results[i].StdDevs != nil {
					for j := range results[0].StdDevs {
						assert.InDelta(t, results[0].StdDevs[j], results[i].StdDevs[j], 1e-10,
							"StdDev[%d] should be identical across methods", j)
					}
				}
			}
		})
	}
}

// TestTemporalPCAVsStandard tests that Temporal PCA with lag=1 on univariate data
// produces similar results to standard PCA on the lag matrix
func TestTemporalPCAVsStandard(t *testing.T) {
	// Create a simple time series
	n := 100
	timeSeries := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		timeSeries[i] = []float64{math.Sin(float64(i)*0.1) + 0.1*float64(i)}
	}

	// Run Temporal PCA
	temporalEngine := NewTemporalPCAEngine()
	temporalConfig := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		TemporalLags:  3,
	}
	temporalResult, err := temporalEngine.Fit(timeSeries, temporalConfig)
	require.NoError(t, err)

	// Create lag matrix manually and run standard PCA
	lagMatrix := createLagMatrix(timeSeries, 3)
	standardEngine := NewPCAEngine()
	standardConfig := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}
	standardResult, err := standardEngine.Fit(lagMatrix, standardConfig)
	require.NoError(t, err)

	// Compare explained variance ratio (should be reasonably similar)
	// Note: Temporal PCA and standard PCA on lag matrix won't be exactly identical
	// due to different preprocessing approaches, but should be in same ballpark
	if len(temporalResult.ExplainedVarRatio) > 0 && len(standardResult.ExplainedVarRatio) > 0 {
		// Check they're at least in the same order of magnitude
		ratio := temporalResult.ExplainedVarRatio[0] / standardResult.ExplainedVarRatio[0]
		assert.Greater(t, ratio, 0.5, "Variance ratio should be similar")
		assert.Less(t, ratio, 2.0, "Variance ratio should be similar")
	}
}

// TestComponentSelectionConsistency tests that different component selection methods
// produce consistent results
func TestComponentSelectionConsistency(t *testing.T) {
	data, err := loadTestDataAsMatrix("../../testdata/wine/wine.csv")
	require.NoError(t, err)

	tests := []struct {
		name             string
		config           types.PCAConfig
		expectedMinComps int
		expectedMaxComps int
	}{
		{
			name: "fixed_components",
			config: types.PCAConfig{
				Components:    3,
				MeanCenter:    true,
				StandardScale: true,
				Method:        "svd",
			},
			expectedMinComps: 3,
			expectedMaxComps: 3,
		},
		{
			name: "variance_explained_90",
			config: types.PCAConfig{
				Components:        10, // Set reasonable max
				MeanCenter:        true,
				StandardScale:     true,
				Method:            "svd",
				VarianceExplained: 0.90,
			},
			expectedMinComps: 1,
			expectedMaxComps: 10, // Wine typically needs ~8 components for 90%
		},
		{
			name: "variance_explained_95",
			config: types.PCAConfig{
				Components:        12, // Set reasonable max
				MeanCenter:        true,
				StandardScale:     true,
				Method:            "svd",
				VarianceExplained: 0.95,
			},
			expectedMinComps: 1,
			expectedMaxComps: 12, // Wine typically needs ~10 components for 95%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPCAEngine()
			result, err := engine.Fit(data, tt.config)
			require.NoError(t, err)

			actualComps := result.ComponentsComputed
			assert.GreaterOrEqual(t, actualComps, tt.expectedMinComps,
				"Should have at least %d components", tt.expectedMinComps)
			assert.LessOrEqual(t, actualComps, tt.expectedMaxComps,
				"Should have at most %d components", tt.expectedMaxComps)

			// Verify cumulative variance if threshold was set
			if tt.config.VarianceExplained > 0 {
				lastCompVariance := result.CumulativeVar[actualComps-1]
				assert.GreaterOrEqual(t, lastCompVariance, tt.config.VarianceExplained,
					"Cumulative variance should meet threshold")
			}
		})
	}
}

// Helper function to compare PCA results from different methods
func compareMethodResults(t *testing.T, result1, result2 *types.PCAResult, tolerance float64, description string) {
	t.Helper()

	// Compare explained variance
	require.Equal(t, len(result1.ExplainedVar), len(result2.ExplainedVar),
		"%s: Different number of components", description)

	for i := range result1.ExplainedVar {
		assert.InDelta(t, result1.ExplainedVar[i], result2.ExplainedVar[i], tolerance,
			"%s: Explained variance for component %d", description, i+1)
	}

	// Compare cumulative variance
	for i := range result1.CumulativeVar {
		assert.InDelta(t, result1.CumulativeVar[i], result2.CumulativeVar[i], tolerance,
			"%s: Cumulative variance for component %d", description, i+1)
	}

	// Compare scores (with sign ambiguity resolution)
	if result1.Scores != nil && result2.Scores != nil {
		// Check dimensions
		r1 := len(result1.Scores)
		c1 := len(result1.Scores[0])
		r2 := len(result2.Scores)
		c2 := len(result2.Scores[0])
		require.Equal(t, r1, r2, "%s: Different number of rows in scores", description)
		require.Equal(t, c1, c2, "%s: Different number of columns in scores", description)

		// Compare with sign ambiguity resolution
		for j := 0; j < c1; j++ {
			// Check if we need to flip sign for this component
			sum := 0.0
			for i := 0; i < r1; i++ {
				sum += result1.Scores[i][j] * result2.Scores[i][j]
			}

			sign := 1.0
			if sum < 0 {
				sign = -1.0
			}

			// Compare values
			for i := 0; i < r1; i++ {
				expected := sign * result2.Scores[i][j]
				assert.InDelta(t, result1.Scores[i][j], expected, tolerance,
					"%s: Scores[%d,%d]", description, i, j)
			}
		}
	}
}

// Helper function to generate a matrix with specified condition number
func generateConditionedMatrix(rows, cols int, condition float64) types.Matrix {
	// Create a matrix with controlled singular values
	data := make(types.Matrix, rows)

	// Generate random orthogonal matrices using QR decomposition
	A := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			A.Set(i, j, normalRand())
		}
	}

	var qr mat.QR
	qr.Factorize(A)
	var Q mat.Dense
	qr.QTo(&Q)

	// Create diagonal matrix with controlled singular values
	singularValues := make([]float64, cols)
	singularValues[0] = 1.0
	for i := 1; i < cols-1; i++ {
		// Linearly space singular values
		singularValues[i] = 1.0 - float64(i)*((1.0-1.0/condition)/float64(cols-1))
	}
	singularValues[cols-1] = 1.0 / condition

	// Construct the matrix: Q * S * Q^T
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			sum := 0.0
			for k := 0; k < cols; k++ {
				sum += Q.At(i, k) * singularValues[k] * Q.At(j, k)
			}
			data[i][j] = sum
		}
	}

	return data
}

// Helper function to create lag matrix for temporal PCA comparison
func createLagMatrix(timeSeries types.Matrix, lags int) types.Matrix {
	n := len(timeSeries)
	if n <= lags || len(timeSeries[0]) != 1 {
		return nil
	}

	// Create lag matrix where each row contains [x(t), x(t-1), ..., x(t-lags)]
	lagMatrix := make(types.Matrix, n-lags)
	for i := lags; i < n; i++ {
		row := make([]float64, lags+1)
		for j := 0; j <= lags; j++ {
			row[j] = timeSeries[i-j][0]
		}
		lagMatrix[i-lags] = row
	}

	return lagMatrix
}

// Simple normal random number generator for matrix generation
func normalRand() float64 {
	// Box-Muller transform for normal distribution
	u1 := 1.0 - rng.Float64()
	u2 := rng.Float64()
	return math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
}

// Helper function to get minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
