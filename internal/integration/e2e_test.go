package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2EBasicWorkflow tests the complete workflow from CSV to results
func TestE2EBasicWorkflow(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	datasets := tc.CreateSampleDatasets(t)

	testCases := []struct {
		name        string
		dataset     string
		method      string
		components  int
		preprocess  string
		expectError bool
	}{
		{
			name:       "SVD with standard scaling",
			dataset:    "small",
			method:     "svd",
			components: 2,
			preprocess: "standard",
		},
		{
			name:       "NIPALS with missing data",
			dataset:    "missing",
			method:     "nipals",
			components: 3,
			preprocess: "mean-center",
		},
		{
			name:       "Kernel PCA with RBF",
			dataset:    "medium",
			method:     "kernel",
			components: 2,
			preprocess: "standard",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			dataset := datasets[test.dataset]
			outputDir := filepath.Join(tc.TempDir, test.name)

			// Run PCA analysis
			args := []string{
				"analyze",
				dataset.Path,
				"--method", test.method,
				"--components", fmt.Sprintf("%d", test.components),
				"--preprocessing", test.preprocess,
				"--output", outputDir,
				"--format", "json",
			}

			if test.method == "kernel" {
				args = append(args, "--kernel", "rbf", "--gamma", "0.1")
			}

			_, err := tc.RunCLI(t, args...)

			if test.expectError {
				AssertError(t, err, "Expected error")
				return
			}

			AssertNoError(t, err, "PCA analysis failed")

			// Verify output files exist
			jsonPath := filepath.Join(outputDir, "pca_results.json")
			CheckFileExists(t, jsonPath)

			// Load and validate results
			results := tc.LoadJSONResult(t, jsonPath)

			// Validate structure
			if _, ok := results["scores"]; !ok {
				t.Error("Missing scores in results")
			}

			if test.method != "kernel" {
				if _, ok := results["loadings"]; !ok {
					t.Error("Missing loadings in results")
				}
			}

			if _, ok := results["explainedVariance"]; !ok {
				t.Error("Missing explainedVariance in results")
			}

			// Validate dimensions
			scores, ok := results["scores"].([]interface{})
			if !ok {
				t.Fatal("Invalid scores format")
			}

			expectedRows := dataset.Rows
			if dataset.HasMissing && test.preprocess != "nipals" {
				expectedRows-- // Some rows might be dropped
			}

			if len(scores) != expectedRows && !dataset.HasMissing {
				t.Errorf("Expected %d scores, got %d", expectedRows, len(scores))
			}
		})
	}
}

// TestE2EAllPreprocessingMethods tests all preprocessing combinations
func TestE2EAllPreprocessingMethods(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	dataset := tc.CreateTestCSV(t, "preprocess.csv", GenerateTestMatrix(30, 10, 5.0))

	preprocessMethods := []string{
		"mean-center",
		"standard",
		"robust",
		"variance",
	}

	rowPreprocess := []string{"", "snv", "l2-norm"}

	for _, colPrep := range preprocessMethods {
		for _, rowPrep := range rowPreprocess {
			name := colPrep
			if rowPrep != "" {
				name = rowPrep + "+" + colPrep
			}

			t.Run(name, func(t *testing.T) {
				outputDir := filepath.Join(tc.TempDir, name)

				args := []string{
					"analyze",
					dataset,
					"--method", "svd",
					"--components", "3",
					"--preprocessing", colPrep,
					"--output", outputDir,
					"--format", "json",
				}

				if rowPrep == "snv" {
					args = append(args, "--snv")
				} else if rowPrep == "l2-norm" {
					args = append(args, "--l2-norm")
				}

				_, err := tc.RunCLI(t, args...)
				AssertNoError(t, err, "Preprocessing test failed")

				// Verify results exist
				jsonPath := filepath.Join(outputDir, "pca_results.json")
				CheckFileExists(t, jsonPath)

				results := tc.LoadJSONResult(t, jsonPath)

				// Verify preprocessing was applied
				if config, ok := results["config"].(map[string]interface{}); ok {
					if prep, ok := config["preprocessing"].(string); ok {
						if prep != colPrep {
							t.Errorf("Expected preprocessing %s, got %s", colPrep, prep)
						}
					}
				}
			})
		}
	}
}

// TestE2ELargeDataset tests performance with larger datasets
func TestE2ELargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Create large dataset
	largeData := GenerateTestMatrix(1000, 100, 10.0)
	largePath := tc.CreateTestCSV(t, "large.csv", largeData)

	outputDir := filepath.Join(tc.TempDir, "large_output")

	// Benchmark the analysis
	result := tc.BenchmarkCLI(t,
		"analyze",
		largePath,
		"--method", "svd",
		"--components", "10",
		"--preprocessing", "standard",
		"--output", outputDir,
		"--format", "json",
	)

	// Check performance
	if result.Duration.Seconds() > 30 {
		t.Errorf("Large dataset took too long: %v", result.Duration)
	}

	// Verify results
	jsonPath := filepath.Join(outputDir, "pca_results.json")
	CheckFileExists(t, jsonPath)
}

// TestE2EExportFormats tests all export formats
func TestE2EExportFormats(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	dataset := tc.CreateTestCSV(t, "export.csv", GenerateTestMatrix(20, 8, 7.0))

	formats := []struct {
		format   string
		ext      string
		validate func(t *testing.T, path string)
	}{
		{
			format: "json",
			ext:    ".json",
			validate: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				AssertNoError(t, err, "Failed to read JSON")

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				AssertNoError(t, err, "Invalid JSON format")
			},
		},
		{
			format: "csv",
			ext:    ".csv",
			validate: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				AssertNoError(t, err, "Failed to read CSV")

				if !strings.Contains(string(data), ",") {
					t.Error("CSV file does not contain commas")
				}
			},
		},
		{
			format: "tsv",
			ext:    ".tsv",
			validate: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				AssertNoError(t, err, "Failed to read TSV")

				if !strings.Contains(string(data), "\t") {
					t.Error("TSV file does not contain tabs")
				}
			},
		},
	}

	for _, f := range formats {
		t.Run(f.format, func(t *testing.T) {
			outputDir := filepath.Join(tc.TempDir, f.format)

			args := []string{
				"analyze",
				dataset,
				"--method", "svd",
				"--components", "2",
				"--output", outputDir,
				"--format", f.format,
			}

			_, err := tc.RunCLI(t, args...)
			AssertNoError(t, err, "Export failed")

			// Find output file
			files, err := filepath.Glob(filepath.Join(outputDir, "*"+f.ext))
			AssertNoError(t, err, "Failed to find output files")

			if len(files) == 0 {
				t.Fatalf("No %s files found in output", f.ext)
			}

			// Validate format
			f.validate(t, files[0])
		})
	}
}

// TestE2EModelExportImport tests model export and transformation
func TestE2EModelExportImport(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Create training data
	trainData := GenerateTestMatrix(50, 15, 8.0)
	trainPath := tc.CreateTestCSV(t, "train.csv", trainData)

	// Create test data (same structure)
	testData := GenerateTestMatrix(20, 15, 9.0)
	testPath := tc.CreateTestCSV(t, "test.csv", testData)

	modelPath := filepath.Join(tc.TempDir, "model.json")
	outputDir1 := filepath.Join(tc.TempDir, "train_output")
	outputDir2 := filepath.Join(tc.TempDir, "test_output")

	// Train model and export
	_, err := tc.RunCLI(t,
		"analyze",
		trainPath,
		"--method", "svd",
		"--components", "5",
		"--preprocessing", "standard",
		"--output", outputDir1,
		"--format", "json",
		"--export-model", modelPath,
	)
	AssertNoError(t, err, "Training failed")

	// Check model file exists
	CheckFileExists(t, modelPath)

	// Transform new data using model
	_, err = tc.RunCLI(t,
		"transform",
		modelPath,
		testPath,
		"--output", outputDir2,
		"--format", "json",
	)
	AssertNoError(t, err, "Transform failed")

	// Verify transform results
	transformPath := filepath.Join(outputDir2, "transform_results.json")
	CheckFileExists(t, transformPath)

	results := tc.LoadJSONResult(t, transformPath)

	// Check that scores exist and have correct dimensions
	if scores, ok := results["scores"].([]interface{}); ok {
		if len(scores) != 20 { // test data has 20 rows
			t.Errorf("Expected 20 transformed samples, got %d", len(scores))
		}

		// Check first row has 5 components
		if firstRow, ok := scores[0].([]interface{}); ok {
			if len(firstRow) != 5 {
				t.Errorf("Expected 5 components, got %d", len(firstRow))
			}
		}
	} else {
		t.Error("Missing or invalid scores in transform results")
	}
}

// TestE2EErrorHandling tests error handling scenarios
func TestE2EErrorHandling(t *testing.T) {
	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	testCases := []struct {
		name        string
		setupFunc   func() string
		args        []string
		expectInErr string
	}{
		{
			name: "Invalid CSV file",
			setupFunc: func() string {
				path := filepath.Join(tc.TempDir, "invalid.csv")
				os.WriteFile(path, []byte("not,a,valid\ncsv,file"), 0644)
				return path
			},
			args:        []string{"analyze", "", "--method", "svd"},
			expectInErr: "parse",
		},
		{
			name: "Non-existent file",
			setupFunc: func() string {
				return filepath.Join(tc.TempDir, "nonexistent.csv")
			},
			args:        []string{"analyze", "", "--method", "svd"},
			expectInErr: "no such file",
		},
		{
			name: "Invalid method",
			setupFunc: func() string {
				return tc.CreateTestCSV(t, "valid.csv", GenerateTestMatrix(10, 5, 1.0))
			},
			args:        []string{"analyze", "", "--method", "invalid"},
			expectInErr: "invalid method",
		},
		{
			name: "Too many components",
			setupFunc: func() string {
				return tc.CreateTestCSV(t, "small.csv", GenerateTestMatrix(5, 3, 1.0))
			},
			args:        []string{"analyze", "", "--method", "svd", "--components", "10"},
			expectInErr: "components",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			inputPath := test.setupFunc()

			// Replace empty string with actual path
			args := make([]string, len(test.args))
			for i, arg := range test.args {
				if arg == "" {
					args[i] = inputPath
				} else {
					args[i] = arg
				}
			}

			_, err := tc.RunCLI(t, args...)

			if err == nil {
				t.Fatal("Expected error but got none")
			}

			errStr := strings.ToLower(err.Error())
			if !strings.Contains(errStr, test.expectInErr) {
				t.Errorf("Expected error containing '%s', got: %v", test.expectInErr, err)
			}
		})
	}
}

// TestE2EDiagnosticMetrics tests diagnostic metrics calculation
func TestE2EDiagnosticMetrics(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	dataset := tc.CreateTestCSV(t, "diagnostics.csv", GenerateTestMatrix(30, 10, 11.0))
	outputDir := filepath.Join(tc.TempDir, "diagnostics")

	// Run with diagnostics enabled
	_, err := tc.RunCLI(t,
		"analyze",
		dataset,
		"--method", "svd",
		"--components", "3",
		"--preprocessing", "standard",
		"--diagnostics",
		"--output", outputDir,
		"--format", "json",
	)
	AssertNoError(t, err, "Diagnostics analysis failed")

	// Load results
	jsonPath := filepath.Join(outputDir, "pca_results.json")
	results := tc.LoadJSONResult(t, jsonPath)

	// Check for diagnostic metrics
	if _, ok := results["tSquared"]; !ok {
		t.Error("Missing T-squared values in results")
	}

	if _, ok := results["qResiduals"]; !ok {
		t.Error("Missing Q residuals in results")
	}

	// Verify dimensions
	if tSquared, ok := results["tSquared"].([]interface{}); ok {
		if len(tSquared) != 30 {
			t.Errorf("Expected 30 T-squared values, got %d", len(tSquared))
		}
	}
}
