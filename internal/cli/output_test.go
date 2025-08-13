// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cli

import (
	"encoding/json"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

func TestConvertToPCAOutputData(t *testing.T) {
	// Create test data
	result := &types.PCAResult{
		Scores: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
		Loadings: [][]float64{
			{0.5, 0.6},
			{0.7, 0.8},
		},
		ExplainedVar:         []float64{50.0, 30.0},
		ExplainedVarRatio:    []float64{0.5, 0.3},
		CumulativeVar:        []float64{50.0, 80.0},
		ComponentLabels:      []string{"PC1", "PC2"},
		VariableLabels:       []string{"Var1", "Var2"},
		ComponentsComputed:   2,
		Method:               "svd",
		PreprocessingApplied: true,
		Means:                []float64{5.0, 6.0},
		StdDevs:              []float64{1.0, 1.5},
		T2Limit95:            10.0,
		T2Limit99:            15.0,
		QLimit95:             5.0,
		QLimit99:             7.5,
	}

	csvData := &CSVData{
		CSVData: &types.CSVData{
			Headers:  []string{"Feature1", "Feature2"},
			RowNames: []string{"Sample1", "Sample2"},
			Matrix: [][]float64{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			Rows:    2,
			Columns: 2,
		},
	}

	config := types.PCAConfig{
		Components:      2,
		MeanCenter:      true,
		StandardScale:   true,
		RobustScale:     false,
		ScaleOnly:       false,
		SNV:             false,
		VectorNorm:      false,
		Method:          "svd",
		MissingStrategy: types.MissingError,
	}

	// Create preprocessor with test data
	preprocessor := core.NewPreprocessorWithScaleOnly(
		config.MeanCenter,
		config.StandardScale,
		config.RobustScale,
		config.ScaleOnly,
		config.SNV,
		config.VectorNorm,
	)

	// Convert to output data
	outputData := ConvertToPCAOutputData(result, csvData, false, config, preprocessor, nil, nil)

	// Test metadata
	if outputData.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", outputData.Metadata.Version)
	}
	if outputData.Metadata.Software != "gopca" {
		t.Errorf("Expected software gopca, got %s", outputData.Metadata.Software)
	}

	// Test config
	if outputData.Metadata.Config.Method != "svd" {
		t.Errorf("Expected method svd, got %s", outputData.Metadata.Config.Method)
	}
	if outputData.Metadata.Config.NComponents != 2 {
		t.Errorf("Expected 2 components, got %d", outputData.Metadata.Config.NComponents)
	}

	// Test preprocessing
	if !outputData.Preprocessing.MeanCenter {
		t.Error("Expected mean center to be true")
	}
	if !outputData.Preprocessing.StandardScale {
		t.Error("Expected standard scale to be true")
	}

	// Test model components
	if len(outputData.Model.Loadings) != 2 {
		t.Errorf("Expected 2 loadings, got %d", len(outputData.Model.Loadings))
	}
	if len(outputData.Model.ComponentLabels) != 2 {
		t.Errorf("Expected 2 component labels, got %d", len(outputData.Model.ComponentLabels))
	}

	// Test results
	if len(outputData.Results.Samples.Scores) != 2 {
		t.Errorf("Expected 2 scores, got %d", len(outputData.Results.Samples.Scores))
	}
	if len(outputData.Results.Samples.Names) != 2 {
		t.Errorf("Expected 2 sample names, got %d", len(outputData.Results.Samples.Names))
	}

	// Test diagnostics
	if outputData.Diagnostics.T2Limit95 != 10.0 {
		t.Errorf("Expected T2Limit95 = 10.0, got %f", outputData.Diagnostics.T2Limit95)
	}
}

func TestJSONSerialization(t *testing.T) {
	// Create minimal test data
	result := &types.PCAResult{
		Scores:             [][]float64{{1.0}},
		Loadings:           [][]float64{{1.0}},
		ExplainedVar:       []float64{100.0},
		ExplainedVarRatio:  []float64{1.0},
		CumulativeVar:      []float64{100.0},
		ComponentLabels:    []string{"PC1"},
		ComponentsComputed: 1,
		Method:             "svd",
	}

	csvData := &CSVData{
		CSVData: &types.CSVData{
			Headers:  []string{"Feature1"},
			RowNames: []string{"Sample1"},
			Matrix:   [][]float64{{1.0}},
			Rows:     1,
			Columns:  1,
		},
	}

	config := types.PCAConfig{
		Components: 1,
		MeanCenter: true,
		Method:     "svd",
	}

	preprocessor := core.NewPreprocessor(true, false, false)

	// Convert to output data
	outputData := ConvertToPCAOutputData(result, csvData, false, config, preprocessor, nil, nil)

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledData types.PCAOutputData
	err = json.Unmarshal(jsonData, &unmarshaledData)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Verify key fields survived round-trip
	if unmarshaledData.Metadata.Version != outputData.Metadata.Version {
		t.Error("Version mismatch after JSON round-trip")
	}
	if unmarshaledData.Model.ComponentLabels[0] != "PC1" {
		t.Error("Component label mismatch after JSON round-trip")
	}
}

func TestKernelParametersOnlyForKernelPCA(t *testing.T) {
	// Create minimal test data
	csvData := &CSVData{
		CSVData: &types.CSVData{
			Headers:  []string{"Feature1"},
			RowNames: []string{"Sample1"},
			Matrix:   [][]float64{{1.0}},
			Rows:     1,
			Columns:  1,
		},
	}

	preprocessor := core.NewPreprocessorWithScaleOnly(true, false, false, false, false, false)

	tests := []struct {
		name               string
		method             string
		expectKernelParams bool
		kernelType         string
		kernelGamma        float64
		kernelDegree       int
		kernelCoef0        float64
		expectGamma        bool
		expectDegree       bool
		expectCoef0        bool
	}{
		{
			name:               "SVD method should not include kernel parameters",
			method:             "svd",
			expectKernelParams: false,
			kernelType:         "rbf",
			kernelGamma:        0.001,
			kernelDegree:       3,
			kernelCoef0:        1.0,
		},
		{
			name:               "NIPALS method should not include kernel parameters",
			method:             "nipals",
			expectKernelParams: false,
			kernelType:         "linear",
			kernelGamma:        0.5,
			kernelDegree:       2,
			kernelCoef0:        0.0,
		},
		{
			name:               "Kernel method with RBF should include only type and gamma",
			method:             "kernel",
			expectKernelParams: true,
			kernelType:         "rbf",
			kernelGamma:        0.001,
			kernelDegree:       3,
			kernelCoef0:        1.0,
			expectGamma:        true,
			expectDegree:       false,
			expectCoef0:        false,
		},
		{
			name:               "Kernel method with polynomial should include all params",
			method:             "kernel",
			expectKernelParams: true,
			kernelType:         "poly",
			kernelGamma:        0.1,
			kernelDegree:       3,
			kernelCoef0:        2.0,
			expectGamma:        true,
			expectDegree:       true,
			expectCoef0:        true,
		},
		{
			name:               "Kernel method with linear should include only type",
			method:             "kernel",
			expectKernelParams: true,
			kernelType:         "linear",
			kernelGamma:        0.5,
			kernelDegree:       2,
			kernelCoef0:        0.0,
			expectGamma:        false,
			expectDegree:       false,
			expectCoef0:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create result with appropriate method and loadings
			result := &types.PCAResult{
				Scores:             [][]float64{{1.0}},
				Loadings:           [][]float64{}, // Empty for kernel PCA
				ExplainedVar:       []float64{100.0},
				ExplainedVarRatio:  []float64{1.0},
				CumulativeVar:      []float64{100.0},
				ComponentLabels:    []string{"PC1"},
				ComponentsComputed: 1,
				Method:             tt.method,
			}

			// Add loadings for non-kernel methods
			if tt.method != "kernel" {
				result.Loadings = [][]float64{{1.0}}
			}

			// Create config with kernel parameters
			config := types.PCAConfig{
				Components:      1,
				MeanCenter:      true,
				Method:          tt.method,
				MissingStrategy: types.MissingError,
				KernelType:      tt.kernelType,
				KernelGamma:     tt.kernelGamma,
				KernelDegree:    tt.kernelDegree,
				KernelCoef0:     tt.kernelCoef0,
			}

			// Convert to output data
			outputData := ConvertToPCAOutputData(result, csvData, false, config, preprocessor, nil, nil)

			// Marshal to JSON
			jsonData, err := json.Marshal(outputData)
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			// Parse JSON to check if kernel fields are present
			var jsonMap map[string]interface{}
			if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Navigate to metadata.config
			metadata, ok := jsonMap["metadata"].(map[string]interface{})
			if !ok {
				t.Fatal("metadata not found in JSON")
			}
			configMap, ok := metadata["config"].(map[string]interface{})
			if !ok {
				t.Fatal("config not found in metadata")
			}

			// Verify method is correct
			if configMap["method"] != tt.method {
				t.Errorf("Expected method %s, got %v", tt.method, configMap["method"])
			}

			// Check if kernel parameters are present/absent as expected
			_, hasKernelType := configMap["kernel_type"]
			_, hasKernelGamma := configMap["kernel_gamma"]
			_, hasKernelDegree := configMap["kernel_degree"]
			_, hasKernelCoef0 := configMap["kernel_coef0"]

			if tt.expectKernelParams {
				// Kernel method - type should always be present
				if !hasKernelType {
					t.Error("kernel_type should be present for kernel method")
				}

				// Check specific parameters based on kernel type
				if tt.expectGamma && !hasKernelGamma {
					t.Errorf("kernel_gamma should be present for %s kernel", tt.kernelType)
				} else if !tt.expectGamma && hasKernelGamma {
					t.Errorf("kernel_gamma should not be present for %s kernel", tt.kernelType)
				}

				if tt.expectDegree && !hasKernelDegree {
					t.Errorf("kernel_degree should be present for %s kernel", tt.kernelType)
				} else if !tt.expectDegree && hasKernelDegree {
					t.Errorf("kernel_degree should not be present for %s kernel", tt.kernelType)
				}

				if tt.expectCoef0 && !hasKernelCoef0 {
					t.Errorf("kernel_coef0 should be present for %s kernel", tt.kernelType)
				} else if !tt.expectCoef0 && hasKernelCoef0 {
					t.Errorf("kernel_coef0 should not be present for %s kernel", tt.kernelType)
				}
			} else {
				// Non-kernel methods - no parameters should be present
				if hasKernelType || hasKernelGamma || hasKernelDegree || hasKernelCoef0 {
					t.Errorf("No kernel parameters should be present for %s method", tt.method)
				}
			}
		})
	}
}
