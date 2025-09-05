// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"
)

// TestTemporalPCALagMatrixConstruction tests the lag matrix building
func TestTemporalPCALagMatrixConstruction(t *testing.T) {
	tests := []struct {
		name        string
		data        [][]float64
		numLags     int
		expectedDim [2]int // [rows, cols]
		shouldError bool
	}{
		{
			name: "simple 2-lag matrix",
			data: [][]float64{
				{1.0, 2.0},
				{3.0, 4.0},
				{5.0, 6.0},
				{7.0, 8.0},
			},
			numLags:     2,
			expectedDim: [2]int{3, 4}, // (4-2+1) × (2*2)
		},
		{
			name: "single lag",
			data: [][]float64{
				{1.0, 2.0, 3.0},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			numLags:     1,
			expectedDim: [2]int{3, 3}, // (3-1+1) × (3*1)
		},
		{
			name: "more lags than samples",
			data: [][]float64{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			numLags:     3,
			shouldError: true,
		},
		{
			name: "equal lags and samples",
			data: [][]float64{
				{1.0},
				{2.0},
				{3.0},
			},
			numLags:     3,
			expectedDim: [2]int{1, 3}, // (3-3+1) × (1*3)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &TemporalPCAImpl{
				numLags: tt.numLags,
			}

			dataMatrix := mat.NewDense(len(tt.data), len(tt.data[0]), nil)
			for i := range tt.data {
				for j := range tt.data[i] {
					dataMatrix.Set(i, j, tt.data[i][j])
				}
			}

			lagMatrix, err := engine.buildLagMatrix(dataMatrix)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, lagMatrix)

				rows, cols := lagMatrix.Dims()
				assert.Equal(t, tt.expectedDim[0], rows, "unexpected number of rows")
				assert.Equal(t, tt.expectedDim[1], cols, "unexpected number of columns")

				// Verify content for simple case
				if tt.name == "simple 2-lag matrix" {
					// First row should be [1, 2, 3, 4] (t=0 and t=1 for both variables)
					assert.Equal(t, 1.0, lagMatrix.At(0, 0))
					assert.Equal(t, 2.0, lagMatrix.At(0, 1))
					assert.Equal(t, 3.0, lagMatrix.At(0, 2))
					assert.Equal(t, 4.0, lagMatrix.At(0, 3))
				}
			}
		})
	}
}

// TestTemporalPCAFit tests the fitting process
func TestTemporalPCAFit(t *testing.T) {
	// Generate synthetic time series data
	// Simple linear trend with noise
	data := make([][]float64, 50)
	for i := range data {
		data[i] = []float64{
			float64(i) + 0.1*math.Sin(float64(i)*0.5),
			float64(i)*0.5 + 0.2*math.Cos(float64(i)*0.3),
		}
	}

	matrix := types.Matrix(data)

	config := types.PCAConfig{
		Components:   2,
		TemporalLags: 5,
		MeanCenter:   true,
	}

	engine := NewTemporalPCAEngine()
	result, err := engine.Fit(matrix, config)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check dimensions
	assert.Equal(t, 2, result.ComponentsComputed)
	assert.Equal(t, 46, len(result.Scores))      // 50 - 5 + 1 = 46 effective samples
	assert.Equal(t, 2, len(result.Scores[0]))    // 2 components
	assert.Equal(t, 2, len(result.Loadings))     // 2 components
	assert.Equal(t, 10, len(result.Loadings[0])) // 2 vars * 5 lags = 10

	// Check explained variance
	assert.Equal(t, 2, len(result.ExplainedVar))
	assert.Equal(t, 2, len(result.ExplainedVarRatio))
	assert.Equal(t, 2, len(result.CumulativeVar))

	// First component should explain most variance for trend data
	assert.Greater(t, result.ExplainedVar[0], 0.7)

	// Cumulative variance should be increasing (now in percentage scale 0-100)
	assert.Less(t, result.CumulativeVar[0], result.CumulativeVar[1])
	assert.LessOrEqual(t, result.CumulativeVar[1], 100.0)
}

// TestTemporalPCAVarianceExplained tests the variance explained criterion
func TestTemporalPCAVarianceExplained(t *testing.T) {
	// Generate data with clear structure
	data := make([][]float64, 100)
	for i := range data {
		data[i] = []float64{
			float64(i), // Strong trend
			0.1*float64(i) + math.Sin(float64(i)*0.1), // Weak trend + oscillation
			0.01 * float64(i),                         // Very weak trend
		}
	}

	matrix := types.Matrix(data)

	tests := []struct {
		name              string
		varianceExplained float64
		expectedComps     int
	}{
		{"90% variance", 0.90, 1},
		{"95% variance", 0.95, 2},
		{"99% variance", 0.99, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.PCAConfig{
				VarianceExplained: tt.varianceExplained,
				TemporalLags:      3,
				StandardScale:     true,
			}

			engine := NewTemporalPCAEngine()
			result, err := engine.Fit(matrix, config)

			require.NoError(t, err)
			assert.LessOrEqual(t, result.ComponentsComputed, tt.expectedComps+1,
				"should not use too many components")
			assert.GreaterOrEqual(t, result.CumulativeVar[result.ComponentsComputed-1],
				tt.varianceExplained, "should meet variance criterion")
		})
	}
}

// TestTemporalPCATransform tests transformation of new data
func TestTemporalPCATransform(t *testing.T) {
	// Training data
	trainData := make([][]float64, 30)
	for i := range trainData {
		trainData[i] = []float64{
			float64(i) + math.Sin(float64(i)*0.2),
			float64(i)*0.5 + math.Cos(float64(i)*0.3),
		}
	}

	trainMatrix := types.Matrix(
		trainData)

	config := types.PCAConfig{
		Components:   2,
		TemporalLags: 3,
		MeanCenter:   true,
	}

	engine := NewTemporalPCAEngine()
	_, err := engine.Fit(trainMatrix, config)
	require.NoError(t, err)

	// Test data (continuation of time series)
	testData := make([][]float64, 20)
	for i := range testData {
		j := i + 30
		testData[i] = []float64{
			float64(j) + math.Sin(float64(j)*0.2),
			float64(j)*0.5 + math.Cos(float64(j)*0.3),
		}
	}

	testMatrix := types.Matrix(
		testData)

	// Transform test data
	transformed, err := engine.Transform(testMatrix)
	require.NoError(t, err)
	require.NotNil(t, transformed)

	// Check dimensions
	// types.Matrix is [][]float64
	transformedData := [][]float64(transformed)
	assert.Equal(t, 18, len(transformedData))   // 20 - 3 + 1
	assert.Equal(t, 2, len(transformedData[0])) // 2 components
}

// TestTemporalPCABasicEdgeCases tests basic edge cases
func TestTemporalPCABasicEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        [][]float64
		config      types.PCAConfig
		shouldError bool
		errorMsg    string
	}{
		{
			name: "zero lags",
			data: [][]float64{{1, 2}, {3, 4}},
			config: types.PCAConfig{
				Components:   1,
				TemporalLags: 0,
			},
			shouldError: true,
			errorMsg:    "positive number of lags",
		},
		{
			name: "negative lags",
			data: [][]float64{{1, 2}, {3, 4}},
			config: types.PCAConfig{
				Components:   1,
				TemporalLags: -1,
			},
			shouldError: true,
			errorMsg:    "positive number of lags",
		},
		{
			name: "empty data",
			data: [][]float64{},
			config: types.PCAConfig{
				Components:   1,
				TemporalLags: 2,
			},
			shouldError: true,
			errorMsg:    "empty",
		},
		{
			name: "single sample",
			data: [][]float64{{1, 2, 3}},
			config: types.PCAConfig{
				Components:   1,
				TemporalLags: 1,
			},
			shouldError: false, // Should work with L=1 and T=1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := types.Matrix(tt.data)

			engine := NewTemporalPCAEngine()
			result, err := engine.Fit(matrix, tt.config)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestGetLoadingForLag tests the loading accessor method
func TestGetLoadingForLag(t *testing.T) {
	// Create simple data
	data := make([][]float64, 20)
	for i := range data {
		data[i] = []float64{float64(i), float64(i) * 2}
	}

	matrix := types.Matrix(data)

	config := types.PCAConfig{
		Components:   2,
		TemporalLags: 3,
	}

	engine := &TemporalPCAImpl{}
	_, err := engine.Fit(matrix, config)
	require.NoError(t, err)

	// Test valid access
	loading, err := engine.GetLoadingForLag(0, 0, 0) // var 0, lag 0, comp 0
	assert.NoError(t, err)
	assert.IsType(t, float64(0), loading)

	// Test invalid variable
	_, err = engine.GetLoadingForLag(5, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable index")

	// Test invalid lag
	_, err = engine.GetLoadingForLag(0, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lag index")

	// Test invalid component
	_, err = engine.GetLoadingForLag(0, 0, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component index")
}

// TestReconstructionError tests reconstruction error computation
func TestReconstructionError(t *testing.T) {
	// Generate data with noise
	data := make([][]float64, 50)
	for i := range data {
		data[i] = []float64{
			float64(i) + 0.1*(float64(i%5)-2.5), // Trend with small noise
			float64(i)*0.5 + 0.2*(float64(i%3)-1.0),
		}
	}

	matrix := types.Matrix(data)

	config := types.PCAConfig{
		Components:    1, // Use only 1 component to ensure reconstruction error
		TemporalLags:  3,
		StandardScale: true,
	}

	engine := &TemporalPCAImpl{}
	_, err := engine.Fit(matrix, config)
	require.NoError(t, err)

	// Compute reconstruction errors
	errors, err := engine.ReconstructionError(matrix)
	require.NoError(t, err)
	require.NotNil(t, errors)

	// Should have error for each effective sample
	assert.Equal(t, 48, len(errors)) // 50 - 3 + 1

	// All errors should be non-negative
	for i, err := range errors {
		assert.GreaterOrEqual(t, err, 0.0, "error %d should be non-negative", i)
	}

	// Errors should be relatively small for trend data
	// Since we're using only 1 component, expect some reconstruction error
	avgError := 0.0
	for _, e := range errors {
		avgError += e
	}
	avgError /= float64(len(errors))
	assert.Less(t, avgError, 2.0, "average reconstruction error should be reasonable for 1 component")
}

// TestLagContributions tests the lag contribution calculation
func TestLagContributions(t *testing.T) {
	// Create data where recent lags should be more important
	data := make([][]float64, 100)
	for i := range data {
		data[i] = []float64{
			float64(i) + 0.5*float64(i-1),     // AR(1) process
			float64(i)*0.3 + 0.7*float64(i-1), // Another AR(1)
		}
	}
	// Fix first row
	data[0] = []float64{0, 0}

	matrix := types.Matrix(data)

	config := types.PCAConfig{
		Components:   2,
		TemporalLags: 4,
	}

	engine := &TemporalPCAImpl{}
	_, err := engine.Fit(matrix, config)
	require.NoError(t, err)

	contributions, err := engine.GetLagContributions()
	require.NoError(t, err)
	require.NotNil(t, contributions)

	// Check dimensions
	assert.Equal(t, 2, len(contributions))    // 2 components
	assert.Equal(t, 4, len(contributions[0])) // 4 lags

	// Each component's contributions should sum to 1
	for comp, contribs := range contributions {
		sum := 0.0
		for _, val := range contribs {
			sum += val
			assert.GreaterOrEqual(t, val, 0.0, "contributions should be non-negative")
		}
		assert.InDelta(t, 1.0, sum, 1e-10, "component %d contributions should sum to 1", comp)
	}
}

// TestAutoCorrelation tests the autocorrelation computation
func TestAutoCorrelation(t *testing.T) {
	// Create periodic data
	data := make([][]float64, 100)
	for i := range data {
		data[i] = []float64{
			math.Sin(float64(i) * 2 * math.Pi / 10), // Period 10
			math.Cos(float64(i) * 2 * math.Pi / 20), // Period 20
		}
	}

	acf, err := ComputeAutoCorrelation(data, 30)
	require.NoError(t, err)
	require.NotNil(t, acf)

	// Check dimensions
	assert.Equal(t, 2, len(acf))     // 2 variables
	assert.Equal(t, 31, len(acf[0])) // maxLag + 1

	// ACF at lag 0 should be 1
	assert.InDelta(t, 1.0, acf[0][0], 1e-10)
	assert.InDelta(t, 1.0, acf[1][0], 1e-10)

	// For periodic data, ACF should show periodicity
	// Variable 0 (period 10) should have high correlation at lag 10, 20, 30
	assert.Greater(t, acf[0][10], 0.9, "should have high correlation at period")
	assert.Greater(t, acf[0][20], 0.9, "should have high correlation at 2*period")

	// Variable 1 (period 20) should have high correlation at lag 20
	assert.Greater(t, acf[1][20], 0.9, "should have high correlation at period")
}

// TestMemoryEstimation tests the memory estimation function
func TestMemoryEstimation(t *testing.T) {
	tests := []struct {
		name      string
		samples   int
		variables int
		lags      int
		wantWarn  bool
	}{
		{
			name:      "small dataset",
			samples:   100,
			variables: 10,
			lags:      5,
			wantWarn:  false,
		},
		{
			name:      "large dataset",
			samples:   100000,
			variables: 100,
			lags:      50,
			wantWarn:  true,
		},
		{
			name:      "more lags than samples",
			samples:   10,
			variables: 5,
			lags:      20,
			wantWarn:  false, // Should get different warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, warning := EstimateTemporalPCAMemory(tt.samples, tt.variables, tt.lags)

			if tt.lags > tt.samples {
				assert.Contains(t, warning, "exceeds")
				assert.Equal(t, int64(0), bytes)
			} else {
				if tt.wantWarn {
					assert.Contains(t, warning, "GB")
				} else {
					assert.Empty(t, warning)
				}
				assert.Greater(t, bytes, int64(0))

				// Verify calculation
				effectiveSamples := tt.samples - tt.lags + 1
				expectedBase := int64(effectiveSamples) * int64(tt.variables*tt.lags) * 8
				expectedTotal := expectedBase + expectedBase/4
				assert.Equal(t, expectedTotal, bytes)
			}
		})
	}
}

// TestNumericalStability tests numerical stability with ill-conditioned data
func TestNumericalStability(t *testing.T) {
	// Create data with very different scales
	data := make([][]float64, 50)
	for i := range data {
		data[i] = []float64{
			1e-10 * float64(i), // Very small scale
			1e10 * float64(i),  // Very large scale
			float64(i),         // Normal scale
		}
	}

	matrix := types.Matrix(data)

	config := types.PCAConfig{
		Components:    2,
		TemporalLags:  3,
		StandardScale: true, // Should handle scale differences
	}

	engine := NewTemporalPCAEngine()
	result, err := engine.Fit(matrix, config)

	// Should complete without error despite scale differences
	require.NoError(t, err)
	require.NotNil(t, result)

	// Results should be finite
	for i, score := range result.Scores {
		for j, val := range score {
			assert.False(t, math.IsNaN(val), "score[%d][%d] is NaN", i, j)
			assert.False(t, math.IsInf(val, 0), "score[%d][%d] is Inf", i, j)
		}
	}
}

// TestTemporalPCAEdgeCases tests edge cases and validation in temporal PCA
func TestTemporalPCAEdgeCases(t *testing.T) {
	engine := NewTemporalPCAEngine()

	t.Run("transform with unfitted model", func(t *testing.T) {
		data := types.Matrix{{1.0, 2.0}, {3.0, 4.0}}
		_, err := engine.Transform(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model not fitted")
	})

	t.Run("transform with empty data", func(t *testing.T) {
		// First fit the model
		trainingData := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
		}
		config := types.PCAConfig{
			Components:   1,
			MeanCenter:   true,
			TemporalLags: 2,
		}
		_, err := engine.Fit(trainingData, config)
		require.NoError(t, err)

		// Test transform with empty data
		_, err = engine.Transform(types.Matrix{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty data matrix")
	})

	t.Run("transform with wrong number of variables", func(t *testing.T) {
		// First fit the model
		trainingData := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
		}
		config := types.PCAConfig{
			Components:   1,
			MeanCenter:   true,
			TemporalLags: 2,
		}
		_, err := engine.Fit(trainingData, config)
		require.NoError(t, err)

		// Test transform with wrong number of variables
		wrongData := types.Matrix{
			{1.0, 2.0, 3.0}, // 3 variables instead of 2
			{4.0, 5.0, 6.0},
		}
		_, err = engine.Transform(wrongData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "doesn't match training data")
	})

	t.Run("transform with insufficient samples for lags", func(t *testing.T) {
		// First fit the model with 3 lags
		trainingData := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
			{7.0, 8.0},
		}
		config := types.PCAConfig{
			Components:   1,
			MeanCenter:   true,
			TemporalLags: 3,
		}
		_, err := engine.Fit(trainingData, config)
		require.NoError(t, err)

		// Test transform with only 2 samples (insufficient for 3 lags)
		smallData := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
		}
		_, err = engine.Transform(smallData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient samples")
	})

	t.Run("lags equal to samples minus 1", func(t *testing.T) {
		// Edge case: L = T - 1 (results in 2 effective samples)
		data := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
			{7.0, 8.0},
			{9.0, 10.0},
		}
		config := types.PCAConfig{
			Components:   1,
			MeanCenter:   true,
			TemporalLags: 4, // T=5, L=4, effective samples = 2
		}
		result, err := engine.Fit(data, config)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result.Scores)) // 5-4+1 = 2 effective samples
	})

	t.Run("all eigenvalues stored", func(t *testing.T) {
		// Verify that AllEigenvalues field is populated
		data := types.Matrix{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
			{7.0, 8.0, 9.0},
			{10.0, 11.0, 12.0},
			{13.0, 14.0, 15.0},
		}
		config := types.PCAConfig{
			Components:   2,
			MeanCenter:   true,
			TemporalLags: 2,
		}
		result, err := engine.Fit(data, config)
		require.NoError(t, err)
		require.NotNil(t, result)

		// AllEigenvalues should contain all eigenvalues, not just the retained ones
		assert.NotEmpty(t, result.AllEigenvalues)
		assert.GreaterOrEqual(t, len(result.AllEigenvalues), result.ComponentsComputed)

		// Verify that eigenvalues are in descending order
		for i := 1; i < len(result.AllEigenvalues); i++ {
			assert.GreaterOrEqual(t, result.AllEigenvalues[i-1], result.AllEigenvalues[i])
		}
	})

	t.Run("variance explained calculation", func(t *testing.T) {
		// Test that variance explained sums correctly
		data := types.Matrix{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
			{7.0, 8.0},
			{9.0, 10.0},
			{11.0, 12.0},
		}
		config := types.PCAConfig{
			Components:   3,
			MeanCenter:   true,
			TemporalLags: 2,
		}
		result, err := engine.Fit(data, config)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Sum of explained variance ratio should be <= 100
		totalVar := 0.0
		for _, v := range result.ExplainedVarRatio {
			totalVar += v
		}
		assert.LessOrEqual(t, totalVar, 100.01) // Allow tiny floating point error

		// Cumulative variance should be monotonically increasing
		for i := 1; i < len(result.CumulativeVar); i++ {
			assert.GreaterOrEqual(t, result.CumulativeVar[i], result.CumulativeVar[i-1])
		}
	})
}
