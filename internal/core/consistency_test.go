// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMethodConsistencyAcrossDatasets verifies that each PCA method produces
// consistent results across different datasets with similar characteristics.
func TestMethodConsistencyAcrossDatasets(t *testing.T) {
	methods := []string{"svd", "nipals"}

	datasets := []struct {
		name     string
		file     string
		expected map[string]float64 // Expected variance for first component
	}{
		{
			name: "iris",
			file: "../../testdata/iris/iris.csv",
			expected: map[string]float64{
				"svd":    72.9, // Expected ~72.9% for first PC with standardization
				"nipals": 72.9,
			},
		},
		{
			name: "wine",
			file: "../../testdata/wine/wine.csv",
			expected: map[string]float64{
				"svd":    36.2, // Expected ~36.2% for first PC with standardization
				"nipals": 36.2,
			},
		},
	}

	for _, dataset := range datasets {
		t.Run(dataset.name, func(t *testing.T) {
			// Load data using the existing function from sklearn_validation_test.go
			data, err := loadTestDataAsMatrix(dataset.file)
			require.NoError(t, err)

			for _, method := range methods {
				t.Run(method, func(t *testing.T) {
					config := types.PCAConfig{
						Components:    3,
						StandardScale: true,
						MeanCenter:    true,
						Method:        method,
					}

					// Create PCA engine
					engine := NewPCAEngine()

					// Run PCA
					result, err := engine.Fit(data, config)
					require.NoError(t, err)

					// Check first component variance is close to expected
					if expected, ok := dataset.expected[method]; ok {
						assert.InDelta(t, expected, result.ExplainedVarRatio[0], 1.0,
							"First component variance should match expected for %s on %s",
							method, dataset.name)
					}

					// Verify properties
					verifyPCAProperties(t, result, method, dataset.name)
				})
			}
		})
	}
}

// TestPreprocessingReversibility verifies that preprocessing can be correctly
// reversed during the transform operation.
func TestPreprocessingReversibility(t *testing.T) {
	// Load test data
	data, err := loadTestDataAsMatrix("../../testdata/iris/iris.csv")
	require.NoError(t, err)

	preprocessingMethods := []struct {
		name          string
		meanCenter    bool
		standardScale bool
		robustScale   bool
	}{
		{"mean_center", true, false, false},
		{"standard_scaling", true, true, false},
		{"robust_scaling", true, false, true},
	}

	for _, prep := range preprocessingMethods {
		t.Run(prep.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    3,
				MeanCenter:    prep.meanCenter,
				StandardScale: prep.standardScale,
				RobustScale:   prep.robustScale,
				Method:        "svd",
			}

			// Run PCA
			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)
			require.NoError(t, err)

			// Transform the original data
			scores, err := engine.Transform(data)
			require.NoError(t, err)

			// Inverse transform to reconstruct
			reconstructed := reconstructFromScores(scores, result)

			// Reverse preprocessing
			original := reversePreprocessing(reconstructed, result)

			// Check reconstruction error (note: with 3 components, perfect reconstruction isn't possible)
			mse := calculateMSE(data, original)
			// Tolerance depends on how many components we use vs data dimensions
			// With 3 components for 4-dimensional iris data, some error is expected
			assert.Less(t, mse, 1.0, "Reconstruction error should be reasonable")
		})
	}
}

// TestTransformConsistency ensures that the transform operation produces
// consistent results across different methods
func TestTransformConsistency(t *testing.T) {
	// Load training data
	trainData, err := loadTestDataAsMatrix("../../testdata/wine/wine.csv")
	require.NoError(t, err)

	// Split data into train and test
	splitIdx := len(trainData) * 3 / 4
	testData := trainData[splitIdx:]
	trainData = trainData[:splitIdx]

	methods := []string{"svd", "nipals"}

	var engines []types.PCAEngine

	// Fit models with different methods
	for _, method := range methods {
		config := types.PCAConfig{
			Components:    5,
			MeanCenter:    true,
			StandardScale: true,
			Method:        method,
		}

		engine := NewPCAEngine()
		_, err := engine.Fit(trainData, config)
		require.NoError(t, err)

		engines = append(engines, engine)
	}

	// Transform test data with each model
	var transformedData []types.Matrix
	for _, engine := range engines {
		transformed, err := engine.Transform(testData)
		require.NoError(t, err)
		transformedData = append(transformedData, transformed)
	}

	// Compare transformed results (with sign ambiguity)
	for i := 1; i < len(transformedData); i++ {
		compareTransformedData(t, transformedData[0], transformedData[i], 1e-6,
			"Transform results should be consistent between %s and %s", methods[0], methods[i])
	}
}

// TestMissingValueHandling tests that missing values are handled consistently
func TestMissingValueHandling(t *testing.T) {
	// Test that PCA properly rejects data with NaN values
	dataWithNaN := types.Matrix{
		{1.0, 2.0, 3.0, 4.0},
		{math.NaN(), 6.0, 7.0, 8.0},
		{9.0, 10.0, 11.0, 12.0},
		{13.0, 14.0, 15.0, 16.0},
	}

	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}

	engine := NewPCAEngine()

	// PCA should reject data with NaN values
	_, err := engine.Fit(dataWithNaN, config)
	assert.Error(t, err, "PCA should reject data with NaN values")
	assert.Contains(t, err.Error(), "NaN", "Error should mention NaN values")

	// Test with clean data
	cleanData := types.Matrix{
		{1.0, 2.0, 3.0, 4.0},
		{5.0, 6.0, 7.0, 8.0},
		{9.0, 10.0, 11.0, 12.0},
		{13.0, 14.0, 15.0, 16.0},
	}

	result, err := engine.Fit(cleanData, config)
	require.NoError(t, err, "PCA should work with clean data")
	assert.NotNil(t, result)
}

// Helper function to verify PCA properties
func verifyPCAProperties(t *testing.T, result *types.PCAResult, method, dataset string) {
	t.Helper()

	// Check explained variance sums to <= 100
	totalVar := result.CumulativeVar[len(result.CumulativeVar)-1]
	assert.LessOrEqual(t, totalVar, 100.0,
		"%s on %s: Total variance should not exceed 100%%", method, dataset)

	// Check explained variance is decreasing
	for i := 1; i < len(result.ExplainedVar); i++ {
		assert.GreaterOrEqual(t, result.ExplainedVar[i-1], result.ExplainedVar[i],
			"%s on %s: Explained variance should decrease", method, dataset)
	}

	// Check cumulative variance is increasing
	for i := 1; i < len(result.CumulativeVar); i++ {
		assert.Greater(t, result.CumulativeVar[i], result.CumulativeVar[i-1],
			"%s on %s: Cumulative variance should increase", method, dataset)
	}

	// Check loadings orthogonality (if applicable)
	if len(result.Loadings) > 0 {
		checkLoadingsOrthogonality(t, result.Loadings, 1e-10, method, dataset)
	}
}

// Helper function to check loadings orthogonality
func checkLoadingsOrthogonality(t *testing.T, loadings types.Matrix, tolerance float64, method, dataset string) {
	t.Helper()

	if len(loadings) == 0 || len(loadings[0]) < 2 {
		return
	}

	nComponents := len(loadings[0])

	// Check that loading vectors are orthogonal
	for i := 0; i < nComponents; i++ {
		for j := i + 1; j < nComponents; j++ {
			dotProduct := 0.0
			for k := 0; k < len(loadings); k++ {
				dotProduct += loadings[k][i] * loadings[k][j]
			}
			assert.InDelta(t, 0.0, dotProduct, tolerance,
				"%s on %s: Loadings %d and %d should be orthogonal", method, dataset, i, j)
		}
	}
}

// Helper function to reverse preprocessing
func reversePreprocessing(data types.Matrix, result *types.PCAResult) types.Matrix {
	n := len(data)
	p := len(data[0])

	output := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		output[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			val := data[i][j]

			// Reverse standardization
			if result.StdDevs != nil && j < len(result.StdDevs) && result.StdDevs[j] > 0 {
				val *= result.StdDevs[j]
			}

			// Reverse mean-centering
			if result.Means != nil && j < len(result.Means) {
				val += result.Means[j]
			}

			output[i][j] = val
		}
	}

	return output
}

// Helper function to reconstruct data from scores
func reconstructFromScores(scores types.Matrix, result *types.PCAResult) types.Matrix {
	n := len(scores)
	p := len(result.Loadings)

	reconstructed := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		reconstructed[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			for k := 0; k < len(scores[i]); k++ {
				reconstructed[i][j] += scores[i][k] * result.Loadings[j][k]
			}
		}
	}

	return reconstructed
}

// Helper function to calculate mean squared error
func calculateMSE(data1, data2 types.Matrix) float64 {
	if len(data1) != len(data2) || len(data1[0]) != len(data2[0]) {
		return math.Inf(1)
	}

	sum := 0.0
	count := 0
	for i := range data1 {
		for j := range data1[i] {
			diff := data1[i][j] - data2[i][j]
			sum += diff * diff
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// Helper function to compare transformed data with sign ambiguity
func compareTransformedData(t *testing.T, data1, data2 types.Matrix, tolerance float64, msg string, args ...interface{}) {
	t.Helper()

	require.Equal(t, len(data1), len(data2), "Different number of rows")
	require.Equal(t, len(data1[0]), len(data2[0]), "Different number of columns")

	// Check each component
	for j := 0; j < len(data1[0]); j++ {
		// Determine sign by checking correlation
		sum := 0.0
		for i := 0; i < len(data1); i++ {
			sum += data1[i][j] * data2[i][j]
		}

		sign := 1.0
		if sum < 0 {
			sign = -1.0
		}

		// Compare with sign adjustment
		for i := 0; i < len(data1); i++ {
			expected := sign * data2[i][j]
			assert.InDelta(t, data1[i][j], expected, tolerance)
		}
	}
}
