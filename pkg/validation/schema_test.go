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

package validation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestNewModelValidator tests creating a new model validator
func TestNewModelValidator(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "Default version",
			version: "",
			wantErr: false,
		},
		{
			name:    "Explicit v1",
			version: "v1",
			wantErr: false,
		},
		{
			name:    "Invalid version",
			version: "v99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewModelValidator(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewModelValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateModel tests model validation with valid and invalid data
func TestValidateModel(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid complete model",
			data:    createValidPCAOutputData(),
			wantErr: false,
		},
		{
			// The table marshals every case, so this one arrives as the valid
			// JSON document `"not json"` -- a string where an object belongs.
			// That is a real rejection but not a parse failure; the name used to
			// claim otherwise. Malformed bytes are covered by
			// TestValidateModelRejectsMalformedJSON, which cannot be expressed
			// here because anything this table holds is valid JSON by
			// construction.
			name:    "JSON that is not an object",
			data:    "not json",
			wantErr: true,
			errMsg:  "Expected: object, given: string",
		},
		{
			name: "Missing required metadata",
			data: map[string]interface{}{
				"preprocessing": createValidPreprocessing(),
				"model":         createValidModel(),
				"results":       createValidResults(),
			},
			wantErr: true,
			errMsg:  "metadata is required",
		},
		{
			name: "Missing required preprocessing",
			data: map[string]interface{}{
				"metadata": createValidMetadata(),
				"model":    createValidModel(),
				"results":  createValidResults(),
			},
			wantErr: true,
			errMsg:  "preprocessing is required",
		},
		{
			name: "Invalid metadata structure",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"analysis_id":      "123e4567-e89b-12d3-a456-426614174000",
					"software_version": "dev",
					"software":         "wrong-software", // Should be "gopca"
					"created_at":       time.Now().Format(time.RFC3339),
					"config": map[string]interface{}{
						"method":       "svd",
						"n_components": 2,
					},
				},
				"preprocessing": createValidPreprocessing(),
				"model":         createValidModel(),
				"results":       createValidResults(),
			},
			wantErr: true,
			errMsg:  "metadata.software: metadata.software does not match: \"gopca\"",
		},
		{
			name: "Invalid loadings structure",
			data: map[string]interface{}{
				"metadata":      createValidMetadata(),
				"preprocessing": createValidPreprocessing(),
				"model": map[string]interface{}{
					"loadings":                 "not-an-array", // Should be 2D array
					"explained_variance":       []float64{0.5, 0.3},
					"explained_variance_ratio": []float64{0.5, 0.3},
					"cumulative_variance":      []float64{0.5, 0.8},
					"component_labels":         []string{"PC1", "PC2"},
					"feature_labels":           []string{"Feature1", "Feature2"},
				},
				"results": createValidResults(),
			},
			wantErr: true,
			errMsg:  "model.loadings: Invalid type. Expected: array, given: string",
		},
		{
			name: "Invalid scores structure",
			data: map[string]interface{}{
				"metadata":      createValidMetadata(),
				"preprocessing": createValidPreprocessing(),
				"model":         createValidModel(),
				"results": map[string]interface{}{
					"samples": map[string]interface{}{
						"names":  []string{"Sample1", "Sample2"},
						"scores": "not-an-array", // Should be 2D array
					},
				},
			},
			wantErr: true,
			errMsg:  "results.samples.scores: Invalid type. Expected: array, given: string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal test data to JSON
			jsonData, err := json.Marshal(tt.data)
			if err != nil && !tt.wantErr {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			// Validate the model
			err = validator.ValidateModel(jsonData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateModel() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestValidateRealPCAOutputData tests validation with real PCAOutputData structures
func TestValidateRealPCAOutputData(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Create a real PCAOutputData structure
	outputData := &types.PCAOutputData{
		Schema: "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
		Metadata: types.ModelMetadata{
			AnalysisID:      "123e4567-e89b-12d3-a456-426614174000",
			SoftwareVersion: "dev",
			CreatedAt:       time.Now().Format(time.RFC3339),
			Software:        "gopca",
			Config: types.ModelConfig{
				Method:          "svd",
				NComponents:     2,
				MissingStrategy: types.MissingError,
			},
		},
		Preprocessing: types.PreprocessingInfo{
			MeanCenter:    true,
			StandardScale: true,
			RobustScale:   false,
			ScaleOnly:     false,
			SNV:           false,
			VectorNorm:    false,
			Parameters: types.PreprocessingParams{
				FeatureMeans:   []float64{5.0, 3.0},
				FeatureStdDevs: []float64{1.0, 0.5},
			},
		},
		Model: types.ModelComponents{
			Loadings: [][]float64{
				{0.707, -0.707},
				{0.707, 0.707},
			},
			ExplainedVariance:      []float64{2.0, 1.0},
			ExplainedVarianceRatio: []float64{0.67, 0.33},
			CumulativeVariance:     []float64{0.67, 1.0},
			ComponentLabels:        []string{"PC1", "PC2"},
			FeatureLabels:          []string{"Feature1", "Feature2"},
		},
		Results: types.ResultsData{
			Samples: types.SamplesResults{
				Names: []string{"Sample1", "Sample2", "Sample3"},
				Scores: [][]float64{
					{1.0, 0.5},
					{-1.0, 0.5},
					{0.0, -1.0},
				},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(outputData)
	if err != nil {
		t.Fatalf("Failed to marshal PCAOutputData: %v", err)
	}

	// Validate
	if err := validator.ValidateModel(jsonData); err != nil {
		t.Errorf("ValidateModel() failed for valid PCAOutputData: %v", err)
	}
}

// TestValidateSchemaField tests validation of the $schema field
func TestValidateSchemaField(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{
			name:    "Valid GitHub URL schema",
			schema:  "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
			wantErr: false,
		},
		{
			name:    "Valid relative schema",
			schema:  "../schemas/v1/pca-output.schema.json",
			wantErr: false,
		},
		{
			name:    "Invalid schema version",
			schema:  "https://github.com/bitjungle/gopca/schemas/v99/pca-output.schema.json",
			wantErr: true,
		},
		{
			name:    "No schema field",
			schema:  "",
			wantErr: false, // Schema field is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputData := createValidPCAOutputData()
			dataMap := outputData.(map[string]interface{})

			if tt.schema != "" {
				dataMap["$schema"] = tt.schema
			}

			jsonData, err := json.Marshal(dataMap)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			err = validator.ValidateModel(jsonData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateWithMetrics tests validation with diagnostic metrics included
func TestValidateWithMetrics(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Create model with metrics
	outputData := createValidPCAOutputData()
	dataMap := outputData.(map[string]interface{})

	// Add metrics to results
	results := dataMap["results"].(map[string]interface{})
	samples := results["samples"].(map[string]interface{})
	samples["metrics"] = map[string]interface{}{
		"hotelling_t2": []float64{1.5, 2.0, 0.8},
		"mahalanobis":  []float64{1.2, 1.8, 0.6},
		"rss":          []float64{0.1, 0.2, 0.05},
		"is_outlier":   []bool{false, true, false},
	}

	// Add diagnostic limits
	dataMap["diagnostics"] = map[string]interface{}{
		"t2_limit_95": 5.99,
		"t2_limit_99": 9.21,
		"q_limit_95":  2.5,
		"q_limit_99":  3.8,
	}

	// Marshal and validate
	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := validator.ValidateModel(jsonData); err != nil {
		t.Errorf("ValidateModel() failed for model with metrics: %v", err)
	}
}

// Helper functions

func createValidMetadata() map[string]interface{} {
	return map[string]interface{}{
		"analysis_id":      "123e4567-e89b-12d3-a456-426614174000",
		"software_version": "dev",
		"created_at":       time.Now().Format(time.RFC3339),
		"software":         "gopca",
		"config": map[string]interface{}{
			"method":       "svd",
			"n_components": 2,
		},
	}
}

func createValidPreprocessing() map[string]interface{} {
	return map[string]interface{}{
		"mean_center":    true,
		"standard_scale": true,
		"robust_scale":   false,
		"scale_only":     false,
		"snv":            false,
		"vector_norm":    false,
		"parameters": map[string]interface{}{
			"feature_means":   []float64{5.0, 3.0},
			"feature_stddevs": []float64{1.0, 0.5},
		},
	}
}

func createValidModel() map[string]interface{} {
	return map[string]interface{}{
		"loadings": [][]float64{
			{0.707, -0.707},
			{0.707, 0.707},
		},
		"explained_variance":       []float64{2.0, 1.0},
		"explained_variance_ratio": []float64{0.67, 0.33},
		"cumulative_variance":      []float64{0.67, 1.0},
		"component_labels":         []string{"PC1", "PC2"},
		"feature_labels":           []string{"Feature1", "Feature2"},
	}
}

func createValidResults() map[string]interface{} {
	return map[string]interface{}{
		"samples": map[string]interface{}{
			"names": []string{"Sample1", "Sample2", "Sample3"},
			"scores": [][]float64{
				{1.0, 0.5},
				{-1.0, 0.5},
				{0.0, -1.0},
			},
		},
	}
}

func createValidPCAOutputData() interface{} {
	return map[string]interface{}{
		"metadata":      createValidMetadata(),
		"preprocessing": createValidPreprocessing(),
		"model":         createValidModel(),
		"results":       createValidResults(),
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

// TestValidateModelRejectsMalformedJSON covers the parse-error branch.
//
// The table-driven test above cannot reach it: it marshals each case, so
// whatever it holds is valid JSON by the time ValidateModel sees it. Its
// "Invalid JSON" case therefore exercised a type mismatch and left the
// `invalid JSON:` path untested, which is why this takes raw bytes instead.
func TestValidateModelRejectsMalformedJSON(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("NewModelValidator: %v", err)
	}

	for _, tc := range []struct{ name, data string }{
		{"truncated object", `{"metadata": {`},
		{"trailing comma", `{"metadata": {},}`},
		{"not json at all", `this is not json`},
		{"empty input", ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateModel([]byte(tc.data))
			if err == nil {
				t.Fatal("malformed JSON was accepted")
			}
			if !contains(err.Error(), "invalid JSON") {
				t.Errorf("expected a parse error naming the input as invalid JSON, got: %v", err)
			}
		})
	}
}
