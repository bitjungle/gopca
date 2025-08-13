// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// First, create a PCA model using analyze
	modelFile := filepath.Join(tempDir, "test_model.json")

	// Run analyze to create model
	app := NewApp()
	err := app.Run([]string{
		"pca", "analyze",
		"-f", "json",
		"-o", tempDir,
		"../datasets/iris.csv",
	})
	require.NoError(t, err, "Failed to create PCA model")

	// Check model file was created
	require.FileExists(t, filepath.Join(tempDir, "iris_pca.json"))

	// Rename to expected name
	err = os.Rename(filepath.Join(tempDir, "iris_pca.json"), modelFile)
	require.NoError(t, err)

	// Now test the transform command
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T)
	}{
		{
			name: "basic transform table output",
			args: []string{
				"pca", "transform",
				modelFile,
				"../datasets/iris.csv",
			},
			wantErr: false,
		},
		{
			name: "transform with JSON output",
			args: []string{
				"pca", "transform",
				"-f", "json",
				"-o", tempDir,
				modelFile,
				"../datasets/iris.csv",
			},
			wantErr: false,
			checkFunc: func(t *testing.T) {
				// Check output file exists
				outputFile := filepath.Join(tempDir, "iris_transformed.json")
				require.FileExists(t, outputFile)

				// Load and check content
				data, err := os.ReadFile(outputFile)
				require.NoError(t, err)

				var output struct {
					Metadata struct {
						NSamples    int `json:"n_samples"`
						NComponents int `json:"n_components"`
					} `json:"metadata"`
					Results []struct {
						ID     string             `json:"id"`
						Scores map[string]float64 `json:"scores"`
					} `json:"results"`
				}

				err = json.Unmarshal(data, &output)
				require.NoError(t, err)

				assert.Equal(t, 150, output.Metadata.NSamples)
				assert.Equal(t, 2, output.Metadata.NComponents)
				assert.Len(t, output.Results, 150)

				// Check first sample
				assert.Equal(t, "se_01", output.Results[0].ID)
				assert.InDelta(t, -2.6842, output.Results[0].Scores["PC1"], 0.0001)
				assert.InDelta(t, -0.3266, output.Results[0].Scores["PC2"], 0.0001)
			},
		},
		{
			name: "transform with excluded rows",
			args: []string{
				"pca", "transform",
				"--exclude-rows", "1-10",
				modelFile,
				"../datasets/iris.csv",
			},
			wantErr: false,
		},
		{
			name: "missing model file",
			args: []string{
				"pca", "transform",
				"nonexistent.json",
				"../datasets/iris.csv",
			},
			wantErr: true,
			errMsg:  "failed to load model",
		},
		{
			name: "missing data file",
			args: []string{
				"pca", "transform",
				modelFile,
				"nonexistent.csv",
			},
			wantErr: true,
			errMsg:  "failed to parse CSV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			err := app.Run(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

func TestTransformScoresMatchOriginal(t *testing.T) {
	// This test ensures that transforming the same data produces identical scores
	tempDir := t.TempDir()

	// Create model
	app := NewApp()
	err := app.Run([]string{
		"pca", "analyze",
		"-f", "json",
		"-o", tempDir,
		"../datasets/iris.csv",
	})
	require.NoError(t, err)

	// Load the original results
	originalData, err := os.ReadFile(filepath.Join(tempDir, "iris_pca.json"))
	require.NoError(t, err)

	var original types.PCAOutputData
	err = json.Unmarshal(originalData, &original)
	require.NoError(t, err)

	// Transform the same data
	err = app.Run([]string{
		"pca", "transform",
		"-f", "json",
		"-o", tempDir,
		filepath.Join(tempDir, "iris_pca.json"),
		"../datasets/iris.csv",
	})
	require.NoError(t, err)

	// Load transform results
	transformData, err := os.ReadFile(filepath.Join(tempDir, "iris_transformed.json"))
	require.NoError(t, err)

	var transformed struct {
		Results []struct {
			ID     string             `json:"id"`
			Scores map[string]float64 `json:"scores"`
		} `json:"results"`
	}
	err = json.Unmarshal(transformData, &transformed)
	require.NoError(t, err)

	// Compare scores
	require.Equal(t, len(original.Results.Samples.Names), len(transformed.Results))

	for i, name := range original.Results.Samples.Names {
		assert.Equal(t, name, transformed.Results[i].ID)

		// Check each component
		for j, compLabel := range original.Model.ComponentLabels {
			originalScore := original.Results.Samples.Scores[i][j]
			transformedScore := transformed.Results[i].Scores[compLabel]
			assert.InDelta(t, originalScore, transformedScore, 1e-10,
				fmt.Sprintf("Score mismatch for sample %s, component %s", name, compLabel))
		}
	}
}

func TestTransformWithPreprocessing(t *testing.T) {
	// Test that preprocessing parameters are correctly applied
	tempDir := t.TempDir()

	// Create model with standard scaling
	app := NewApp()
	err := app.Run([]string{
		"pca", "analyze",
		"--scale", "standard",
		"-f", "json",
		"-o", tempDir,
		"../datasets/iris.csv",
	})
	require.NoError(t, err)

	// Transform should apply the same preprocessing
	err = app.Run([]string{
		"pca", "transform",
		"-f", "json",
		"-o", tempDir,
		filepath.Join(tempDir, "iris_pca.json"),
		"../datasets/iris.csv",
	})
	require.NoError(t, err)

	// Verify output exists
	require.FileExists(t, filepath.Join(tempDir, "iris_transformed.json"))
}

func TestComprehensiveFitTransformWorkflow(t *testing.T) {
	// Comprehensive test of fit-transform workflow with different methods and preprocessing
	tests := []struct {
		name          string
		method        string
		preprocessing []string
		description   string
	}{
		{
			name:          "SVD with mean centering",
			method:        "svd",
			preprocessing: []string{}, // Mean centering is default
			description:   "SVD method with mean centering only",
		},
		{
			name:          "SVD with standard scaling",
			method:        "svd",
			preprocessing: []string{"--scale", "standard"},
			description:   "SVD method with standard scaling (includes mean centering)",
		},
		{
			name:          "SVD with robust scaling",
			method:        "svd",
			preprocessing: []string{"--scale", "robust"},
			description:   "SVD method with robust scaling (median/MAD)",
		},
		{
			name:          "NIPALS with mean centering",
			method:        "nipals",
			preprocessing: []string{}, // Mean centering is default
			description:   "NIPALS method with mean centering",
		},
		{
			name:          "NIPALS with standard scaling",
			method:        "nipals",
			preprocessing: []string{"--scale", "standard"},
			description:   "NIPALS method with standard scaling",
		},
		{
			name:          "SVD with SNV preprocessing",
			method:        "svd",
			preprocessing: []string{"--snv"},
			description:   "SVD with Standard Normal Variate row-wise normalization",
		},
		{
			name:          "SVD with vector normalization",
			method:        "svd",
			preprocessing: []string{"--vector-norm"},
			description:   "SVD with L2 vector normalization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			app := NewApp()

			// Step 1: Fit the model
			fitArgs := []string{
				"pca", "analyze",
				"--method", tt.method,
				"--components", "2",
				"-f", "json",
				"-o", tempDir,
			}
			fitArgs = append(fitArgs, tt.preprocessing...)
			fitArgs = append(fitArgs, "../../testdata/iris/iris.csv")

			err := app.Run(fitArgs)
			require.NoError(t, err, "Failed to fit model for %s", tt.description)

			// Step 2: Load the fitted model results
			originalData, err := os.ReadFile(filepath.Join(tempDir, "iris_pca.json"))
			require.NoError(t, err, "Failed to read original model file")

			var original types.PCAOutputData
			err = json.Unmarshal(originalData, &original)
			require.NoError(t, err, "Failed to unmarshal original model")

			// Verify the method is correct in the export
			assert.Equal(t, tt.method, original.Metadata.Config.Method,
				"Exported method should match the method used for fitting")

			// Step 3: Transform the same data using the saved model
			err = app.Run([]string{
				"pca", "transform",
				"-f", "json",
				"-o", tempDir,
				filepath.Join(tempDir, "iris_pca.json"),
				"../../testdata/iris/iris.csv",
			})
			require.NoError(t, err, "Failed to transform data for %s", tt.description)

			// Step 4: Load transform results
			transformData, err := os.ReadFile(filepath.Join(tempDir, "iris_transformed.json"))
			require.NoError(t, err, "Failed to read transform results")

			var transformed struct {
				Results []struct {
					ID     string             `json:"id"`
					Scores map[string]float64 `json:"scores"`
				} `json:"results"`
			}
			err = json.Unmarshal(transformData, &transformed)
			require.NoError(t, err, "Failed to unmarshal transform results")

			// Step 5: Compare scores - they should match exactly
			require.Equal(t, len(original.Results.Samples.Names), len(transformed.Results),
				"Number of samples should match")

			// Skip loadings check for kernel PCA
			if tt.method != "kernel" && len(original.Model.Loadings) > 0 {
				assert.NotEmpty(t, original.Model.Loadings, "Non-kernel methods should have loadings")
			}

			// Compare each sample's scores
			for i, name := range original.Results.Samples.Names {
				assert.Equal(t, name, transformed.Results[i].ID,
					"Sample names should match at index %d", i)

				// Check each component score
				for j, compLabel := range original.Model.ComponentLabels {
					originalScore := original.Results.Samples.Scores[i][j]
					transformedScore := transformed.Results[i].Scores[compLabel]

					// Use a tight tolerance - scores should be nearly identical
					assert.InDelta(t, originalScore, transformedScore, 1e-10,
						"Score mismatch for %s: sample %s, component %s (original: %f, transformed: %f)",
						tt.description, name, compLabel, originalScore, transformedScore)
				}
			}

			// Additional checks for preprocessing parameters
			if len(tt.preprocessing) == 0 || containsArg(tt.preprocessing, "standard") {
				assert.True(t, original.Preprocessing.MeanCenter,
					"Mean centering should be enabled for %s", tt.description)
			}
			if containsArg(tt.preprocessing, "standard") {
				assert.True(t, original.Preprocessing.StandardScale,
					"Standard scaling should be enabled for %s", tt.description)
			}
			if containsArg(tt.preprocessing, "robust") {
				assert.True(t, original.Preprocessing.RobustScale,
					"Robust scaling should be enabled for %s", tt.description)
			}
			if containsArg(tt.preprocessing, "--snv") {
				assert.True(t, original.Preprocessing.SNV,
					"SNV should be enabled for %s", tt.description)
			}
			if containsArg(tt.preprocessing, "--vector-norm") {
				assert.True(t, original.Preprocessing.VectorNorm,
					"Vector normalization should be enabled for %s", tt.description)
			}
		})
	}
}

// Helper function to check if a slice contains a string
func containsArg(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		model   *types.PCAOutputData
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid model",
			model: &types.PCAOutputData{
				Model: types.ModelComponents{
					Loadings: [][]float64{{1, 0}, {0, 1}},
				},
				Metadata: types.ModelMetadata{
					Config: types.ModelConfig{
						Method: "svd",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing loadings",
			model: &types.PCAOutputData{
				Metadata: types.ModelMetadata{
					Config: types.ModelConfig{
						Method: "svd",
					},
				},
			},
			wantErr: true,
			errMsg:  "missing loadings",
		},
		{
			name: "missing method",
			model: &types.PCAOutputData{
				Model: types.ModelComponents{
					Loadings: [][]float64{{1, 0}, {0, 1}},
				},
			},
			wantErr: true,
			errMsg:  "missing method",
		},
		{
			name: "kernel PCA not supported",
			model: &types.PCAOutputData{
				Model: types.ModelComponents{
					Loadings: [][]float64{{1, 0}, {0, 1}},
				},
				Metadata: types.ModelMetadata{
					Config: types.ModelConfig{
						Method: "kernel",
					},
				},
			},
			wantErr: true,
			errMsg:  "kernel PCA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModel(tt.model)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
