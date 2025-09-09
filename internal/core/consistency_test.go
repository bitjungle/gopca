// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"fmt"
	"math"
	"os"
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
				"svd":    92.4, // Expected ~92.4% for first PC
				"nipals": 92.4,
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
			// Load data
			csvData, err := loadTestDataFromCSV(dataset.file)
			require.NoError(t, err)

			for _, method := range methods {
				t.Run(method, func(t *testing.T) {
					config := types.PCAConfig{
						Components:    3,
						StandardScale: true,
					}

					// Create PCA instance based on method
					var pca types.PCAEngine
					switch method {
					case "svd":
						pca = NewPCAEngine()
					case "nipals":
						pca = NewPCAEngine()
					}

					// Run PCA with method in config
					config.Method = method
					result, err := pca.Fit(csvData.Matrix, config)
					require.NoError(t, err)

					// Check first component variance is close to expected
					if expected, ok := dataset.expected[method]; ok {
						assert.InDelta(t, expected, result.ExplainedVar[0], 1.0,
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
	csvData, err := loadTestDataFromCSV("../../testdata/iris/iris.csv")
	require.NoError(t, err)

	preprocessingMethods := []types.PreprocessingType{
		types.PreprocessingTypeMeanCenter,
		types.PreprocessingTypeStandardScaling,
		types.PreprocessingTypeRobustScaling,
	}

	for _, preprocessing := range preprocessingMethods {
		t.Run(string(preprocessing), func(t *testing.T) {
			config := types.PCAConfig{
				Components: 4, // Use all components for perfect reconstruction
				Method:     "svd",
			}
			// Apply preprocessing
			switch preprocessing {
			case types.PreprocessingTypeMeanCenter:
				config.MeanCenter = true
			case types.PreprocessingTypeStandardScaling:
				config.StandardScale = true
			case types.PreprocessingTypeRobustScaling:
				config.RobustScale = true
			}

			// Run PCA
			pca := NewPCAEngine()
			result, err := pca.Fit(csvData.Matrix, config)
			require.NoError(t, err)

			// Transform the same data
			scores, err := pca.Transform(csvData.Matrix)
			require.NoError(t, err)

			// Inverse transform to reconstruct original data
			reconstructed := reconstructFromScores(scores, result.Loadings, result)

			// Compare with original (should be very close with all components)
			maxError := 0.0
			for i := range csvData.Matrix {
				for j := range csvData.Matrix[i] {
					diff := math.Abs(csvData.Matrix[i][j] - reconstructed[i][j])
					if diff > maxError {
						maxError = diff
					}
				}
			}

			// With all components, reconstruction should be near-perfect
			assert.Less(t, maxError, 1e-10,
				"Reconstruction error should be minimal with all components for %s",
				preprocessing)
		})
	}
}

// TestTransformConsistency verifies that the transform operation produces
// consistent results across different PCA methods.
func TestTransformConsistency(t *testing.T) {
	// Load training and test data
	csvData, err := loadTestDataFromCSV("../../testdata/wine/wine.csv")
	require.NoError(t, err)

	// Split data into train and test
	trainSize := len(csvData.Matrix) * 3 / 4
	trainData := csvData.Matrix[:trainSize]
	testData := csvData.Matrix[trainSize:]

	// Fit models with different methods
	svdConfig := types.PCAConfig{
		Components:    3,
		StandardScale: true,
		Method:        "svd",
	}
	svdPCA := NewPCAEngine()
	svdResult, err := svdPCA.Fit(trainData, svdConfig)
	require.NoError(t, err)
	assert.NotNil(t, svdResult)

	nipalsConfig := types.PCAConfig{
		Components:    3,
		StandardScale: true,
		Method:        "nipals",
	}
	nipalsPCA := NewPCAEngine()
	nipalsResult, err := nipalsPCA.Fit(trainData, nipalsConfig)
	require.NoError(t, err)
	assert.NotNil(t, nipalsResult)

	// Transform test data with both models
	svdTransformed, err := svdPCA.Transform(testData)
	require.NoError(t, err)

	nipalsTransformed, err := nipalsPCA.Transform(testData)
	require.NoError(t, err)

	// Results should be very similar (after resolving sign ambiguity)
	svdAligned := resolveSignAmbiguity(svdTransformed, nipalsTransformed)
	err = compareMatrices(svdAligned, nipalsTransformed, 1e-4,
		"Transform results between SVD and NIPALS")
	assert.NoError(t, err)
}

// TestIterativeConvergence verifies that iterative methods (NIPALS) converge
// properly and produce stable results.
func TestIterativeConvergence(t *testing.T) {
	// Generate well-conditioned synthetic data
	data := generateWellConditionedData(100, 10)

	config := types.PCAConfig{
		Components: 5,
		MeanCenter: true,
		Method:     "nipals",
	}

	// Run NIPALS multiple times
	nRuns := 5
	results := make([]*types.PCAResult, nRuns)

	for i := 0; i < nRuns; i++ {
		pca := NewPCAEngine()
		result, err := pca.Fit(data, config)
		require.NoError(t, err)
		results[i] = result
	}

	// All runs should produce identical results (deterministic)
	for i := 1; i < nRuns; i++ {
		err := compareMethodResults(results[0], results[i], 1e-12,
			fmt.Sprintf("NIPALS run %d vs run 0", i))
		assert.NoError(t, err, "NIPALS should produce deterministic results")
	}
}

// TestMissingValueHandling verifies that methods handle missing values correctly
// and consistently where applicable.
func TestMissingValueHandling(t *testing.T) {
	// Create data with missing values
	data := generateDataWithMissing(100, 10, 0.1) // 10% missing

	strategies := []types.MissingValueStrategy{
		types.MissingMean,
		types.MissingMedian,
		types.MissingDrop,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			config := types.PCAConfig{
				Components:      3,
				MeanCenter:      true,
				MissingStrategy: strategy,
				Method:          "svd",
			}

			// SVD with imputation
			svdPCA := NewPCAEngine()
			svdResult, err := svdPCA.Fit(data, config)

			if strategy == types.MissingError {
				assert.Error(t, err, "Should error with missing values when strategy is 'error'")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, svdResult)

				// Verify imputation was applied - check scores don't contain NaN
				for i := range svdResult.Scores {
					for j := range svdResult.Scores[i] {
						assert.False(t, math.IsNaN(svdResult.Scores[i][j]),
							"Scores should not contain NaN after imputation")
					}
				}
			}
		})
	}

	// Test NIPALS with native missing value handling
	t.Run("nipals_native", func(t *testing.T) {
		config := types.PCAConfig{
			Components:      3,
			MeanCenter:      true,
			MissingStrategy: types.MissingNative,
			Method:          "nipals",
		}

		nipalsPCA := NewPCAEngine()
		nipalsResult, err := nipalsPCA.Fit(data, config)
		require.NoError(t, err)
		assert.NotNil(t, nipalsResult)

		// NIPALS should handle missing values without imputation
		verifyPCAProperties(t, nipalsResult, "nipals", "data with missing")
	})
}

// Helper function to verify basic PCA properties
func verifyPCAProperties(t *testing.T, result *types.PCAResult, method, dataset string) {
	// Check explained variance sums to <= 100
	totalVariance := 0.0
	for _, v := range result.ExplainedVar {
		totalVariance += v
	}
	assert.LessOrEqual(t, totalVariance, 100.0+1e-6,
		"%s on %s: total variance should not exceed 100%%", method, dataset)

	// Check cumulative variance is monotonically increasing
	for i := 1; i < len(result.CumulativeVar); i++ {
		assert.GreaterOrEqual(t, result.CumulativeVar[i], result.CumulativeVar[i-1],
			"%s on %s: cumulative variance should be monotonically increasing", method, dataset)
	}

	// Check explained variance is non-negative and decreasing
	for i := 0; i < len(result.ExplainedVar); i++ {
		assert.GreaterOrEqual(t, result.ExplainedVar[i], 0.0,
			"%s on %s: explained variance should be non-negative", method, dataset)

		if i > 0 {
			assert.GreaterOrEqual(t, result.ExplainedVar[i-1], result.ExplainedVar[i]-1e-6,
				"%s on %s: explained variance should be decreasing", method, dataset)
		}
	}

	// Check dimensions
	assert.Equal(t, result.ComponentsComputed, len(result.ExplainedVar),
		"%s on %s: explained variance length should match num components", method, dataset)
	assert.Equal(t, result.ComponentsComputed, len(result.CumulativeVar),
		"%s on %s: cumulative variance length should match num components", method, dataset)

	if len(result.Scores) > 0 {
		assert.Equal(t, result.ComponentsComputed, len(result.Scores[0]),
			"%s on %s: scores width should match num components", method, dataset)
	}

	if len(result.Loadings) > 0 {
		assert.Equal(t, result.ComponentsComputed, len(result.Loadings[0]),
			"%s on %s: loadings width should match num components", method, dataset)
	}
}

// Helper function to generate well-conditioned synthetic data
func generateWellConditionedData(n, p int) types.Matrix {
	// Use the existing function from stability_test.go
	m := generateMatrixWithConditionNumber(n, p, 10.0) // condition number = 10
	return matrixToTypes(m)
}

// Helper function to generate data with missing values
func generateDataWithMissing(n, p int, missingRate float64) types.Matrix {
	data := make([][]float64, n)
	for i := 0; i < n; i++ {
		data[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			if uniformRandom() < missingRate {
				data[i][j] = math.NaN()
			} else {
				data[i][j] = standardNormal()
			}
		}
	}
	return types.Matrix(data)
}

// Helper function to reconstruct data from scores and loadings
func reconstructFromScores(scores types.Matrix, loadings types.Matrix, result *types.PCAResult) types.Matrix {
	n := len(scores)
	p := len(loadings)

	// Reconstruct: X_reconstructed = Scores * Loadings^T
	reconstructed := make([][]float64, n)
	for i := 0; i < n; i++ {
		reconstructed[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			sum := 0.0
			for k := 0; k < len(scores[i]); k++ {
				sum += scores[i][k] * loadings[j][k]
			}
			reconstructed[i][j] = sum
		}
	}

	// Reverse preprocessing
	if result.PreprocessingApplied {
		reconstructed = reversePreprocessing(reconstructed, result)
	}

	return types.Matrix(reconstructed)
}

// Helper function to reverse preprocessing
func reversePreprocessing(data [][]float64, result *types.PCAResult) [][]float64 {
	n := len(data)
	p := len(data[0])

	output := make([][]float64, n)
	for i := 0; i < n; i++ {
		output[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			val := data[i][j]

			// Reverse standardization
			if result.StdDevs != nil && j < len(result.StdDevs) {
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

// Helper function to load test data from CSV
func loadTestDataFromCSV(path string) (*types.CSVData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	format := types.DefaultCSVFormat()
	format.FieldDelimiter = ','
	parser := types.NewCSVParser(format)
	return parser.Parse(file)
}
