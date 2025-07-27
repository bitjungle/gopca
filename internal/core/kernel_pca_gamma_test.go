package core

import (
	"math"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/internal/datasets"
	"github.com/bitjungle/gopca/pkg/types"
)

// TestKernelPCA_DefaultGamma tests that gamma is set to 1/n_features when not specified
func TestKernelPCA_DefaultGamma(t *testing.T) {
	tests := []struct {
		name        string
		data        types.Matrix
		kernelType  string
		expectGamma float64
	}{
		{
			name:        "2D data",
			data:        generateCircleData(),
			kernelType:  "rbf",
			expectGamma: 0.5, // 1/2 features
		},
		{
			name: "3D data",
			data: types.Matrix{
				[]float64{1.0, 2.0, 3.0},
				[]float64{4.0, 5.0, 6.0},
				[]float64{7.0, 8.0, 9.0},
			},
			kernelType:  "rbf",
			expectGamma: 1.0 / 3.0, // 1/3 features
		},
		{
			name: "5D data",
			data: types.Matrix{
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{6.0, 7.0, 8.0, 9.0, 10.0},
			},
			kernelType:  "poly",
			expectGamma: 0.2, // 1/5 features
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewKernelPCAEngine()
			config := types.PCAConfig{
				Components:  1,
				Method:      "kernel",
				KernelType:  tt.kernelType,
				KernelGamma: 0, // Not specified, should use default
			}

			// For polynomial kernel, add required parameters
			if tt.kernelType == "poly" {
				config.KernelDegree = 3
				config.KernelCoef0 = 0
			}

			// Fit should set default gamma
			result, err := engine.Fit(tt.data, config)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check that gamma was set correctly in the config
			kpca := engine.(*KernelPCAImpl)
			if math.Abs(kpca.config.KernelGamma-tt.expectGamma) > 1e-9 {
				t.Errorf("Expected gamma %f, got %f", tt.expectGamma, kpca.config.KernelGamma)
			}

			// Result should be valid
			if result == nil || len(result.Scores) == 0 {
				t.Error("Expected valid PCA result")
			}
		})
	}
}

// TestKernelPCA_SwissRollDataset tests Kernel PCA with the Swiss Roll dataset
func TestKernelPCA_SwissRollDataset(t *testing.T) {
	// Load Swiss Roll dataset
	swissRollCSV, ok := datasets.GetDataset("swiss_roll.csv")
	if !ok {
		t.Fatal("Failed to load Swiss Roll dataset")
	}

	// Parse the CSV data
	csvData, _, err := types.ParseCSVMixed(strings.NewReader(swissRollCSV), types.DefaultCSVFormat())
	if err != nil {
		t.Fatalf("Failed to parse Swiss Roll CSV: %v", err)
	}

	// Create Kernel PCA engine
	engine := NewKernelPCAEngine()

	// Test with default gamma (should be 1/3 for 3 features)
	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0, // Should default to 1/3
	}

	result, err := engine.Fit(csvData.Matrix, config)
	if err != nil {
		t.Fatalf("Failed to fit Kernel PCA on Swiss Roll: %v", err)
	}

	// Check that gamma was set to 1/3
	kpca := engine.(*KernelPCAImpl)
	expectedGamma := 1.0 / 3.0
	if math.Abs(kpca.config.KernelGamma-expectedGamma) > 1e-9 {
		t.Errorf("Expected gamma %f, got %f", expectedGamma, kpca.config.KernelGamma)
	}

	// Verify results are reasonable
	if result.ComponentsComputed != 2 {
		t.Errorf("Expected 2 components, got %d", result.ComponentsComputed)
	}

	// Check that explained variance values are positive and decreasing
	for i := 0; i < len(result.ExplainedVar)-1; i++ {
		if result.ExplainedVar[i] < 0 {
			t.Errorf("Explained variance %d is negative: %f", i, result.ExplainedVar[i])
		}
		if result.ExplainedVar[i] < result.ExplainedVar[i+1] {
			t.Errorf("Explained variance not in descending order: %f < %f",
				result.ExplainedVar[i], result.ExplainedVar[i+1])
		}
	}

	// Check that scores are valid (no NaN or Inf)
	for i, row := range result.Scores {
		for j, val := range row {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				t.Errorf("Invalid score at [%d,%d]: %f", i, j, val)
			}
		}
	}
}

// TestKernelPCA_ExplicitGamma tests that explicitly set gamma is not overridden
func TestKernelPCA_ExplicitGamma(t *testing.T) {
	data := generateCircleData()
	engine := NewKernelPCAEngine()

	explicitGamma := 0.1
	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: explicitGamma, // Explicitly set
	}

	_, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that gamma was NOT changed
	kpca := engine.(*KernelPCAImpl)
	if math.Abs(kpca.config.KernelGamma-explicitGamma) > 1e-9 {
		t.Errorf("Expected gamma to remain %f, got %f", explicitGamma, kpca.config.KernelGamma)
	}
}
