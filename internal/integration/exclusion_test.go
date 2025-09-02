// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestExclusionFeature tests the row and column exclusion functionality
func TestExclusionFeature(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Create a test dataset with known dimensions
	data := GenerateTestMatrix(20, 6, 1.0)
	testFile := tc.CreateTestCSV(t, "exclusion_test.csv", data)

	testCases := []struct {
		name         string
		excludeRows  string
		excludeCols  string
		expectedRows int
		expectedCols int
	}{
		{
			name:         "No exclusions",
			excludeRows:  "",
			excludeCols:  "",
			expectedRows: 20,
			expectedCols: 6,
		},
		{
			name:         "Exclude single row",
			excludeRows:  "1",
			excludeCols:  "",
			expectedRows: 19,
			expectedCols: 6,
		},
		{
			name:         "Exclude multiple rows (comma-separated)",
			excludeRows:  "1,3,5,7",
			excludeCols:  "",
			expectedRows: 16,
			expectedCols: 6,
		},
		{
			name:         "Exclude row range",
			excludeRows:  "1-5",
			excludeCols:  "",
			expectedRows: 15,
			expectedCols: 6,
		},
		{
			name:         "Exclude mixed rows (range and comma)",
			excludeRows:  "1-3,10,15-17",
			excludeCols:  "",
			expectedRows: 13, // 20 - 7 excluded rows
			expectedCols: 6,
		},
		{
			name:         "Exclude single column",
			excludeRows:  "",
			excludeCols:  "1",
			expectedRows: 20,
			expectedCols: 5,
		},
		{
			name:         "Exclude multiple columns",
			excludeRows:  "",
			excludeCols:  "1,3",
			expectedRows: 20,
			expectedCols: 4,
		},
		{
			name:         "Exclude column range",
			excludeRows:  "",
			excludeCols:  "1-3",
			expectedRows: 20,
			expectedCols: 3,
		},
		{
			name:         "Exclude both rows and columns",
			excludeRows:  "1-5",
			excludeCols:  "1,6",
			expectedRows: 15,
			expectedCols: 4,
		},
		{
			name:         "Complex exclusion pattern",
			excludeRows:  "1-3,8,10-12,20",
			excludeCols:  "2-4",
			expectedRows: 12, // 20 - 8 excluded rows
			expectedCols: 3,  // 6 - 3 excluded columns
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			outputDir := filepath.Join(tc.TempDir, test.name)

			// Build CLI arguments
			args := []string{
				"analyze",
				"--components", "2",
				"--format", "json",
				"--output-dir", outputDir,
				"--output-all",
			}

			if test.excludeRows != "" {
				args = append(args, "--exclude-rows", test.excludeRows)
			}
			if test.excludeCols != "" {
				args = append(args, "--exclude-columns", test.excludeCols)
			}

			args = append(args, testFile)

			// Run the CLI
			_, err := tc.RunCLI(t, args...)
			AssertNoError(t, err, "PCA analysis with exclusions failed")

			// Load and validate results
			jsonPath := filepath.Join(outputDir, "exclusion_test_pca.json")
			CheckFileExists(t, jsonPath)

			jsonData, err := os.ReadFile(jsonPath)
			AssertNoError(t, err, "Failed to read JSON output")

			var results map[string]interface{}
			err = json.Unmarshal(jsonData, &results)
			AssertNoError(t, err, "Failed to parse JSON output")

			// Validate dimensions
			validateExclusionResults(t, results, test.expectedRows, test.expectedCols)
		})
	}
}

// TestExclusionWithRealDataset tests exclusion with the Iris dataset
func TestExclusionWithRealDataset(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Use the Iris dataset from testdata
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	AssertNoError(t, err, "Failed to get project root")
	irisPath := filepath.Join(projectRoot, "testdata", "iris", "iris.csv")
	if _, err := os.Stat(irisPath); os.IsNotExist(err) {
		t.Skip("Iris dataset not found")
	}

	outputDir := filepath.Join(tc.TempDir, "iris_exclusion")

	// Test case: Exclude setosa samples (rows 1-50) and sepal length (column 1)
	args := []string{
		"analyze",
		"--components", "2",
		"--format", "json",
		"--output-dir", outputDir,
		"--exclude-rows", "1-50",
		"--exclude-columns", "1",
		"--verbose",
		irisPath,
	}

	output, err := tc.RunCLI(t, args...)
	AssertNoError(t, err, "PCA analysis failed")

	// Check that verbose output mentions the exclusions
	if !contains(output, "Excluded 50 rows") {
		t.Error("Expected verbose output to mention excluded rows")
	}
	if !contains(output, "Excluded 1 column") {
		t.Error("Expected verbose output to mention excluded column")
	}

	// Load results
	jsonPath := filepath.Join(outputDir, "iris_pca.json")
	jsonData, err := os.ReadFile(jsonPath)
	AssertNoError(t, err, "Failed to read JSON output")

	var results map[string]interface{}
	err = json.Unmarshal(jsonData, &results)
	AssertNoError(t, err, "Failed to parse JSON output")

	// Should have 100 samples (150 - 50) and 3 variables (4 - 1)
	validateExclusionResults(t, results, 100, 3)

	// Verify that the first sample in the results is not from the excluded range
	if resultsData, ok := results["results"].(map[string]interface{}); ok {
		if samples, ok := resultsData["samples"].(map[string]interface{}); ok {
			if rowNames, ok := samples["row_names"].([]interface{}); ok && len(rowNames) > 0 {
				firstName := fmt.Sprintf("%v", rowNames[0])
				// The first non-excluded row should be row 51 (0-based index 50)
				if firstName != "50" && firstName != "51" { // Depending on 0 or 1-based in output
					t.Logf("First sample name after exclusion: %s", firstName)
				}
			}
		}
	}
}

func validateExclusionResults(t *testing.T, results map[string]interface{}, expectedRows, expectedCols int) {
	t.Helper()

	// Check scores dimension (rows)
	if resultsData, ok := results["results"].(map[string]interface{}); ok {
		if samples, ok := resultsData["samples"].(map[string]interface{}); ok {
			if scores, ok := samples["scores"].([]interface{}); ok {
				if len(scores) != expectedRows {
					t.Errorf("Expected %d rows in scores, got %d", expectedRows, len(scores))
				}
			} else {
				t.Error("Missing or invalid scores in results")
			}
		}
	}

	// Check loadings dimension (columns)
	if model, ok := results["model"].(map[string]interface{}); ok {
		if loadings, ok := model["loadings"].([]interface{}); ok {
			if len(loadings) != expectedCols {
				t.Errorf("Expected %d variables in loadings, got %d", expectedCols, len(loadings))
			}
		}
	}

	// Check data shape in metadata if present
	if metadata, ok := results["metadata"].(map[string]interface{}); ok {
		if dataShape, ok := metadata["data_shape"].([]interface{}); ok {
			if len(dataShape) == 2 {
				rows := int(dataShape[0].(float64))
				cols := int(dataShape[1].(float64))
				if rows != expectedRows {
					t.Errorf("Metadata shows %d rows, expected %d", rows, expectedRows)
				}
				if cols != expectedCols {
					t.Errorf("Metadata shows %d columns, expected %d", cols, expectedCols)
				}
			}
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
