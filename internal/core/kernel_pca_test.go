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
	}

	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
	}

	if result.ComponentsComputed != 2 {
		t.Errorf("Expected 2 components, got %d", result.ComponentsComputed)
	}

	// Linear kernel PCA should give similar results to regular PCA for linear data
	// Check that variance is captured
	if result.ExplainedVarRatio[0] < 90.0 {
		t.Errorf("First component should explain most variance, got %.2f%%", result.ExplainedVarRatio[0])
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
	if result.CumulativeVar[1] < 50.0 {
		t.Errorf("RBF kernel should capture significant variance, got %.2f%%", result.CumulativeVar[1])
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
			name: "RBF with zero gamma",
			config: types.PCAConfig{
				Components:  2,
				Method:      "kernel",
				KernelType:  "rbf",
				KernelGamma: 0.0,
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

	// Check explained variance adds up
	totalExplained := 0.0
	for _, v := range result.ExplainedVarRatio {
		totalExplained += v
	}

	if math.Abs(totalExplained-100.0) > 0.1 {
		t.Errorf("Explained variance ratios should sum to 100%%, got %.2f%%", totalExplained)
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
