// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/bitjungle/gopca/pkg/security"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKernelPCAVisualizationFields tests that kernel PCA properly exposes
// visualization-related fields in the PCAResult
func TestKernelPCAVisualizationFields(t *testing.T) {
	// Create simple test data
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{2.0, 3.0, 4.0},
		{3.0, 4.0, 5.0},
		{4.0, 5.0, 6.0},
		{5.0, 6.0, 7.0},
	}

	tests := []struct {
		name         string
		kernelType   string
		checkParams  map[string]float64
		expectMatrix bool // Should kernel matrix be included (data size <= 500)
	}{
		{
			name:       "RBF kernel with default gamma",
			kernelType: "rbf",
			checkParams: map[string]float64{
				"gamma": 1.0 / 3.0, // 1/n_features
			},
			expectMatrix: true,
		},
		{
			name:         "Linear kernel",
			kernelType:   "linear",
			checkParams:  map[string]float64{},
			expectMatrix: true,
		},
		{
			name:       "Polynomial kernel",
			kernelType: "poly",
			checkParams: map[string]float64{
				"gamma":  0.5,
				"degree": 3.0,
				"coef0":  1.0,
			},
			expectMatrix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:   2,
				Method:       "kernel",
				KernelType:   tt.kernelType,
				KernelGamma:  0.5,
				KernelDegree: 3,
				KernelCoef0:  1.0,
			}

			// For RBF kernel, let it use default gamma
			if tt.kernelType == "rbf" {
				config.KernelGamma = 0
			}

			engine := NewKernelPCAEngine()
			result, err := engine.Fit(data, config)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check kernel type is properly set
			assert.Equal(t, tt.kernelType, result.KernelType, "Kernel type should be set")

			// Check kernel parameters
			assert.NotNil(t, result.KernelParams, "Kernel parameters should not be nil")
			for key, expectedValue := range tt.checkParams {
				actualValue, exists := result.KernelParams[key]
				assert.True(t, exists, "Parameter %s should exist", key)
				assert.InDelta(t, expectedValue, actualValue, 1e-10, "Parameter %s value mismatch", key)
			}

			// Check kernel matrix (should be included for small datasets)
			if tt.expectMatrix {
				assert.NotNil(t, result.KernelMatrix, "Kernel matrix should be included for small datasets")
				assert.Equal(t, len(data), len(result.KernelMatrix), "Kernel matrix should be n×n")
				if len(result.KernelMatrix) > 0 {
					assert.Equal(t, len(data), len(result.KernelMatrix[0]), "Kernel matrix should be square")
				}
			}

			// Check eigenvectors are included
			assert.NotNil(t, result.KernelEigenvectors, "Eigenvectors should be included")
			assert.Equal(t, len(data), len(result.KernelEigenvectors), "Should have eigenvectors for each sample")
			if len(result.KernelEigenvectors) > 0 {
				assert.Equal(t, config.Components, len(result.KernelEigenvectors[0]), "Should have eigenvector values for each component")
			}

			// Verify eigenvectors are non-zero
			hasNonZero := false
			for i := range result.KernelEigenvectors {
				for j := range result.KernelEigenvectors[i] {
					if result.KernelEigenvectors[i][j] != 0 {
						hasNonZero = true
						break
					}
				}
			}
			assert.True(t, hasNonZero, "Eigenvectors should contain non-zero values")
		})
	}
}

// TestKernelMatrixSizeLimit tests that kernel matrix is not included for large datasets
func TestKernelMatrixSizeLimit(t *testing.T) {
	// Test that maximum allowed samples includes the matrix
	t.Run("max samples - should include matrix", func(t *testing.T) {
		nSamples := security.MaxKernelMatrixVisualization
		nFeatures := 3
		data := make(types.Matrix, nSamples)
		for i := 0; i < nSamples; i++ {
			data[i] = make([]float64, nFeatures)
			for j := 0; j < nFeatures; j++ {
				data[i][j] = float64(i*nFeatures + j)
			}
		}

		config := types.PCAConfig{
			Components: 2,
			Method:     "kernel",
			KernelType: "rbf",
		}

		engine := NewKernelPCAEngine()
		result, err := engine.Fit(data, config)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Kernel matrix SHOULD be included for max allowed samples
		assert.NotNil(t, result.KernelMatrix, "Kernel matrix should be included for datasets with max allowed samples")
		assert.Equal(t, security.MaxKernelMatrixVisualization, len(result.KernelMatrix), "Kernel matrix should have correct number of rows")
	})

	// Test that exceeding max samples excludes the matrix
	t.Run("max+1 samples - should exclude matrix", func(t *testing.T) {
		nSamples := security.MaxKernelMatrixVisualization + 1
		nFeatures := 3
		data := make(types.Matrix, nSamples)
		for i := 0; i < nSamples; i++ {
			data[i] = make([]float64, nFeatures)
			for j := 0; j < nFeatures; j++ {
				data[i][j] = float64(i*nFeatures + j)
			}
		}

		config := types.PCAConfig{
			Components: 2,
			Method:     "kernel",
			KernelType: "rbf",
		}

		engine := NewKernelPCAEngine()
		result, err := engine.Fit(data, config)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Kernel matrix should NOT be included for large datasets
		assert.Nil(t, result.KernelMatrix, "Kernel matrix should not be included for datasets exceeding the limit")

		// But other fields should still be present
		assert.Equal(t, "rbf", result.KernelType)
		assert.NotNil(t, result.KernelParams)
		assert.NotNil(t, result.KernelEigenvectors)
	})
}

// TestKernelPCAVisualizationConsistency tests that visualization data is consistent
func TestKernelPCAVisualizationConsistency(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{2.0, 3.0},
		{3.0, 1.0},
		{4.0, 2.0},
	}

	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
	}

	engine := NewKernelPCAEngine()
	result, err := engine.Fit(data, config)
	require.NoError(t, err)

	// Check consistency between scores and eigenvectors dimensions
	assert.Equal(t, len(data), len(result.Scores), "Scores should have one row per sample")
	assert.Equal(t, len(data), len(result.KernelEigenvectors), "Eigenvectors should have one row per sample")

	// Check that kernel matrix is symmetric (if included)
	if result.KernelMatrix != nil {
		n := len(result.KernelMatrix)
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				assert.InDelta(t, result.KernelMatrix[i][j], result.KernelMatrix[j][i], 1e-10,
					"Kernel matrix should be symmetric at (%d,%d) and (%d,%d)", i, j, j, i)
			}
		}
	}

	// Check that explained variance ratios sum to <= 100
	totalVariance := 0.0
	for _, v := range result.ExplainedVarRatio {
		totalVariance += v
	}
	assert.LessOrEqual(t, totalVariance, 100.0, "Total explained variance ratio should not exceed 100%")
}
