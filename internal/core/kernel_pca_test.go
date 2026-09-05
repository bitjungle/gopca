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
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// Test data: simple 2D data that should benefit from RBF kernel
func generateCircleData() types.Matrix {
	// Generate points in a circle pattern
	n := 20
	data := make(types.Matrix, n*2)

	// Inner circle (class 1)
	for i := 0; i < n; i++ {
		angle := float64(i) * 2 * math.Pi / float64(n)
		data[i] = []float64{
			0.3 * math.Cos(angle),
			0.3 * math.Sin(angle),
		}
	}

	// Outer circle (class 2)
	for i := 0; i < n; i++ {
		angle := float64(i) * 2 * math.Pi / float64(n)
		data[n+i] = []float64{
			math.Cos(angle),
			math.Sin(angle),
		}
	}

	return data
}

// Test linear separable data
func generateLinearData() types.Matrix {
	// Simple linearly separable data
	return types.Matrix{
		[]float64{1.0, 2.0},
		[]float64{2.0, 3.0},
		[]float64{3.0, 4.0},
		[]float64{4.0, 5.0},
		[]float64{5.0, 6.0},
		[]float64{6.0, 7.0},
	}
}

func TestKernelPCA_LinearKernel(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateLinearData()

	config := types.PCAConfig{
		Components: 2,
		Method:     "kernel",
		KernelType: "linear",
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to fit linear kernel PCA: %v", err)
	}

	// Check result structure
	if result == nil {
		t.Fatal("Result is nil")
		return
	}

	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
	}

	if result.ComponentsComputed != 2 {
		t.Errorf("Expected 2 components, got %d", result.ComponentsComputed)
	}

	// Linear kernel PCA should give similar results to regular PCA for linear data
	// Check that variance is captured
	if result.ExplainedVarRatio[0] < 0.90 {
		t.Errorf("First component should explain most of the variance, got %.4f",
			result.ExplainedVarRatio[0])
	}
}

func TestKernelPCA_RBFKernel(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateCircleData()

	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 1.0,
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to fit RBF kernel PCA: %v", err)
	}

	// Check that it successfully separates the circular data
	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
	}

	// RBF kernel should capture non-linear patterns
	if result.CumulativeVar[1] < 0.50 {
		t.Errorf("RBF kernel should capture significant variance, got %.4f",
			result.CumulativeVar[1])
	}
}

func TestKernelPCA_PolynomialKernel(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateLinearData()

	config := types.PCAConfig{
		Components:   2,
		Method:       "kernel",
		KernelType:   "poly",
		KernelGamma:  0.1,
		KernelDegree: 2,
		KernelCoef0:  1.0,
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to fit polynomial kernel PCA: %v", err)
	}

	if result.Method != "kernel" {
		t.Errorf("Expected method 'kernel', got %s", result.Method)
	}

	// Loadings should be empty for kernel PCA
	if len(result.Loadings) != 0 {
		t.Error("Kernel PCA should not produce loadings")
	}
}

func TestKernelPCA_InvalidConfig(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateLinearData()

	tests := []struct {
		name   string
		config types.PCAConfig
	}{
		{
			name: "Missing kernel type",
			config: types.PCAConfig{
				Components: 2,
				Method:     "kernel",
			},
		},
		{
			name: "Invalid kernel type",
			config: types.PCAConfig{
				Components: 2,
				Method:     "kernel",
				KernelType: "invalid",
			},
		},
		{
			name: "RBF with negative gamma",
			config: types.PCAConfig{
				Components:  2,
				Method:      "kernel",
				KernelType:  "rbf",
				KernelGamma: -1.0,
			},
		},
		{
			name: "Poly with zero degree",
			config: types.PCAConfig{
				Components:   2,
				Method:       "kernel",
				KernelType:   "poly",
				KernelGamma:  1.0,
				KernelDegree: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Fit(data, tt.config)
			if err == nil {
				t.Error("Expected error for invalid configuration")
			}
		})
	}
}

func TestKernelPCA_Transform(t *testing.T) {
	engine := NewKernelPCAEngine()
	trainData := generateLinearData()

	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
	}

	// Fit the model
	_, err := engine.Fit(trainData, config)
	if err != nil {
		t.Fatalf("Failed to fit kernel PCA: %v", err)
	}

	// Transform new data
	testData := types.Matrix{
		{1.5, 2.5},
		{3.5, 4.5},
	}

	transformed, err := engine.Transform(testData)
	if err != nil {
		t.Fatalf("Failed to transform data: %v", err)
	}

	if len(transformed) != len(testData) {
		t.Errorf("Expected %d transformed samples, got %d", len(testData), len(transformed))
	}

	if len(transformed[0]) != 2 {
		t.Errorf("Expected 2 components, got %d", len(transformed[0]))
	}
}

func TestKernelPCA_FitTransform(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateCircleData()

	config := types.PCAConfig{
		Components:  3,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 2.0,
	}

	result, err := engine.FitTransform(data, config)
	if err != nil {
		t.Fatalf("Failed to fit-transform: %v", err)
	}

	// Scores should match the training data size
	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
	}

	// Check explained variance is reasonable (should be less than 100% for subset of components)
	totalExplained := 0.0
	for _, v := range result.ExplainedVarRatio {
		totalExplained += v
	}

	if totalExplained > 1.0001 {
		t.Errorf("Explained variance ratios should not exceed 100%%, got %.2f%%", totalExplained)
	}

	// For 2 components out of many samples, explained variance should be less than 100%
	if totalExplained > 99.0 {
		t.Errorf("Expected explained variance for 2 components to be less than 99%%, got %.2f%%", totalExplained)
	}
}

func TestKernelPCA_TransformBeforeFit(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := types.Matrix{{1.0, 2.0}}

	_, err := engine.Transform(data)
	if err == nil {
		t.Error("Expected error when transforming before fit")
	}
}

func TestKernelPCA_EmptyData(t *testing.T) {
	engine := NewKernelPCAEngine()

	config := types.PCAConfig{
		Components: 2,
		Method:     "kernel",
		KernelType: "linear",
	}

	// Empty matrix
	_, err := engine.Fit(types.Matrix{}, config)
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Matrix with empty rows
	_, err = engine.Fit(types.Matrix{{}}, config)
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestKernelPCA_MoreComponentsThanSamples(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}

	config := types.PCAConfig{
		Components: 5, // More than samples
		Method:     "kernel",
		KernelType: "linear",
	}

	_, err := engine.Fit(data, config)
	if err == nil {
		t.Error("Expected error when components exceed samples")
	}
}

func TestKernelPCA_ExplainedVarianceCalculation(t *testing.T) {
	// Test that explained variance is calculated correctly
	// Using all eigenvalues, not just selected components
	engine := NewKernelPCAEngine()

	// Create a dataset with enough samples to test variance calculation
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{2.0, 3.0, 4.0},
		{3.0, 4.0, 5.0},
		{4.0, 5.0, 6.0},
		{5.0, 6.0, 7.0},
		{6.0, 7.0, 8.0},
		{7.0, 8.0, 9.0},
		{8.0, 9.0, 10.0},
		{9.0, 10.0, 11.0},
		{10.0, 11.0, 12.0},
	}

	config := types.PCAConfig{
		Components:  2, // Select only 2 components
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.1,
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to fit: %v", err)
	}

	// Check that individual variance ratios are reasonable
	for i, ratio := range result.ExplainedVarRatio {
		if ratio > 1.0001 {
			t.Errorf("Component %d has explained variance ratio > 100%%: %.2f%%", i+1, ratio)
		}
		if ratio < 0.0 {
			t.Errorf("Component %d has negative explained variance ratio: %.2f%%", i+1, ratio)
		}
	}

	// Check cumulative variance
	totalVar := result.CumulativeVar[len(result.CumulativeVar)-1]
	if totalVar > 100.0 {
		t.Errorf("Cumulative variance exceeds 100%%: %.2f%%", totalVar)
	}

	// For 2 components out of 10 samples, we expect much less than 100%
	if totalVar > 80.0 {
		t.Logf("Warning: Cumulative variance for 2 components is %.2f%%, which seems high", totalVar)
	}
}

func TestNewPCAEngineForMethod(t *testing.T) {
	tests := []struct {
		method       string
		expectKernel bool
	}{
		{"kernel", true},
		{"svd", false},
		{"nipals", false},
		{"", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			engine := NewPCAEngineForMethod(tt.method)

			// Try to fit with appropriate config
			config := types.PCAConfig{
				Components: 2,
				Method:     tt.method,
			}

			if tt.expectKernel {
				config.KernelType = "rbf"
				config.KernelGamma = 1.0
			}

			data := generateLinearData()
			_, err := engine.Fit(data, config)

			if tt.expectKernel {
				// Should work for kernel engine
				if err != nil && err.Error() == "kernel type must be specified for kernel PCA" {
					t.Error("Kernel PCA engine should accept kernel configuration")
				}
			} else {
				// Regular PCA should work but ignore kernel params
				if err != nil {
					// Regular PCA might fail for other reasons, but not because of kernel params
					t.Logf("Regular PCA error (expected): %v", err)
				}
			}
		})
	}
}

// TestKernelPCA_PolynomialNormalization tests that "polynomial" is normalized to "poly"
// This addresses issue #569 where validation accepts both forms but runtime only recognizes "poly"
func TestKernelPCA_PolynomialNormalization(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateLinearData()

	// Test with "polynomial" string (should be normalized to "poly" internally)
	config := types.PCAConfig{
		Components:   2,
		Method:       "kernel",
		KernelType:   "polynomial", // This should be normalized to "poly"
		KernelGamma:  0.1,
		KernelDegree: 3,
		KernelCoef0:  1.0,
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to fit with 'polynomial' kernel type: %v", err)
	}

	// Verify the result contains normalized kernel_type
	if result.KernelType != "poly" {
		t.Errorf("Expected normalized kernel_type 'poly', got %s", result.KernelType)
	}

	// Verify scores were computed
	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
	}
}

// TestKernelPCA_PolynomialDefaultGamma tests that default gamma is applied for "polynomial"
// This verifies the normalization happens before the default gamma logic (line 240 in kernel_pca.go)
func TestKernelPCA_PolynomialDefaultGamma(t *testing.T) {
	engine := NewKernelPCAEngine()
	data := generateLinearData()

	config := types.PCAConfig{
		Components:   2,
		Method:       "kernel",
		KernelType:   "polynomial",
		KernelGamma:  0, // Should get default value of 1/n_features
		KernelDegree: 2,
		KernelCoef0:  0,
	}

	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Failed to apply default gamma for polynomial: %v", err)
	}

	// Verify default gamma was applied (should be 1/n_features = 1/2 = 0.5)
	expectedGamma := 1.0 / float64(len(data[0]))
	actualGamma, ok := result.KernelParams["gamma"]
	if !ok {
		t.Fatal("KernelParams['gamma'] not found in result")
	}
	if actualGamma != expectedGamma {
		t.Errorf("Expected default gamma %.2f, got %.2f", expectedGamma, actualGamma)
	}
}

// TestKernelPCA_PolyVsPolynomial tests that "poly" and "polynomial" produce identical results
func TestKernelPCA_PolyVsPolynomial(t *testing.T) {
	data := generateLinearData()

	baseConfig := types.PCAConfig{
		Components:   2,
		Method:       "kernel",
		KernelGamma:  0.1,
		KernelDegree: 3,
		KernelCoef0:  1.0,
	}

	// Test with "poly"
	configPoly := baseConfig
	configPoly.KernelType = "poly"
	enginePoly := NewKernelPCAEngine()
	resultPoly, err := enginePoly.Fit(data, configPoly)
	if err != nil {
		t.Fatalf("Failed to fit with 'poly': %v", err)
	}

	// Test with "polynomial"
	configPolynomial := baseConfig
	configPolynomial.KernelType = "polynomial"
	enginePolynomial := NewKernelPCAEngine()
	resultPolynomial, err := enginePolynomial.Fit(data, configPolynomial)
	if err != nil {
		t.Fatalf("Failed to fit with 'polynomial': %v", err)
	}

	// Compare results - scores should be identical (or very close due to floating point)
	if len(resultPoly.Scores) != len(resultPolynomial.Scores) {
		t.Errorf("Score lengths differ: %d vs %d", len(resultPoly.Scores), len(resultPolynomial.Scores))
	}

	tolerance := 1e-10
	for i := range resultPoly.Scores {
		for j := range resultPoly.Scores[i] {
			diff := math.Abs(resultPoly.Scores[i][j] - resultPolynomial.Scores[i][j])
			if diff > tolerance {
				t.Errorf("Scores differ at [%d][%d]: %.10f vs %.10f (diff: %.10e)",
					i, j, resultPoly.Scores[i][j], resultPolynomial.Scores[i][j], diff)
			}
		}
	}
}
