package integration

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

// TestSimpleParitySVD tests that CLI produces consistent results for SVD
func TestSimpleParitySVD(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Create test dataset
	testData := GenerateTestMatrix(25, 10, 42.0)
	csvPath := tc.CreateTestCSV(t, "parity_svd.csv", testData)

	// Run the same analysis twice and compare results
	output1 := runAnalysis(t, tc, csvPath, "svd", 3, "standard", "run1")
	output2 := runAnalysis(t, tc, csvPath, "svd", 3, "standard", "run2")

	// Results should be identical
	compareJSONResults(t, output1, output2, 1e-10)
}

// TestSimpleParityMethods tests consistency across different methods
func TestSimpleParityMethods(t *testing.T) {
	SkipIfShort(t)

	tc := NewTestConfig(t)
	tc.BuildCLI(t)

	// Create test dataset without missing values for SVD
	testData := GenerateTestMatrix(20, 8, 33.0)
	csvPath := tc.CreateTestCSV(t, "methods.csv", testData)

	// Run SVD and NIPALS on same data - results should be very similar
	svdOutput := runAnalysis(t, tc, csvPath, "svd", 2, "standard", "svd_run")
	nipalsOutput := runAnalysis(t, tc, csvPath, "nipals", 2, "standard", "nipals_run")

	// Compare explained variance - should be very similar
	svdVar := extractExplainedVariance(t, svdOutput)
	nipalsVar := extractExplainedVariance(t, nipalsOutput)

	if len(svdVar) != len(nipalsVar) {
		t.Errorf("Different number of components: SVD=%d, NIPALS=%d", len(svdVar), len(nipalsVar))
	}

	for i := range svdVar {
		diff := math.Abs(svdVar[i] - nipalsVar[i])
		if diff > 0.01 { // 1% tolerance for explained variance
			t.Errorf("Explained variance mismatch at PC%d: SVD=%.4f, NIPALS=%.4f (diff=%.4f)",
				i+1, svdVar[i], nipalsVar[i], diff)
		}
	}
}

// Helper to run analysis and return JSON results
func runAnalysis(t *testing.T, tc *TestConfig, csvPath, method string, components int, preprocessing, runName string) map[string]interface{} {
	t.Helper()

	outputDir := filepath.Join(tc.TempDir, runName)

	args := []string{
		"analyze",
		"--method", method,
		"--components", fmt.Sprintf("%d", components),
		"--scale", preprocessing,
		"--output-dir", outputDir,
		"--format", "json",
		"--output-all",
		csvPath,
	}

	if method == "nipals" {
		args = append(args, "--missing-strategy", "drop")
	}

	_, err := tc.RunCLI(t, args...)
	AssertNoError(t, err, fmt.Sprintf("%s analysis failed", method))

	// The CLI outputs files as {basename}_pca.json
	baseName := strings.TrimSuffix(filepath.Base(csvPath), ".csv")
	jsonPath := filepath.Join(outputDir, baseName+"_pca.json")
	return tc.LoadJSONResult(t, jsonPath)
}

// Compare two JSON results for consistency
func compareJSONResults(t *testing.T, result1, result2 map[string]interface{}, tolerance float64) {
	t.Helper()

	// Check that both have the same keys
	for key := range result1 {
		if _, ok := result2[key]; !ok {
			t.Errorf("Key '%s' missing in second result", key)
		}
	}

	for key := range result2 {
		if _, ok := result1[key]; !ok {
			t.Errorf("Key '%s' missing in first result", key)
		}
	}

	// Compare scores if present (nested in results.samples.scores)
	if results1, ok1 := result1["results"].(map[string]interface{}); ok1 {
		if results2, ok2 := result2["results"].(map[string]interface{}); ok2 {
			if samples1, ok1 := results1["samples"].(map[string]interface{}); ok1 {
				if samples2, ok2 := results2["samples"].(map[string]interface{}); ok2 {
					if scores1, ok1 := samples1["scores"]; ok1 {
						if scores2, ok2 := samples2["scores"]; ok2 {
							compareMatrixValues(t, "scores", scores1, scores2, tolerance)
						}
					}
				}
			}
		}
	}

	// Compare loadings if present (nested in model.loadings)
	if model1, ok1 := result1["model"].(map[string]interface{}); ok1 {
		if model2, ok2 := result2["model"].(map[string]interface{}); ok2 {
			if loadings1, ok1 := model1["loadings"]; ok1 {
				if loadings2, ok2 := model2["loadings"]; ok2 {
					compareMatrixValues(t, "loadings", loadings1, loadings2, tolerance)
				}
			}

			// Compare explained variance (nested in model.explained_variance)
			if var1, ok1 := model1["explained_variance"]; ok1 {
				if var2, ok2 := model2["explained_variance"]; ok2 {
					compareVectorValues(t, "explained_variance", var1, var2, tolerance)
				}
			}
		}
	}
}

// Compare matrix values
func compareMatrixValues(t *testing.T, name string, val1, val2 interface{}, tolerance float64) {
	t.Helper()

	arr1, ok1 := val1.([]interface{})
	arr2, ok2 := val2.([]interface{})

	if !ok1 || !ok2 {
		t.Errorf("Invalid %s format", name)
		return
	}

	if len(arr1) != len(arr2) {
		t.Errorf("%s row count mismatch: %d vs %d", name, len(arr1), len(arr2))
		return
	}

	for i := range arr1 {
		row1, ok1 := arr1[i].([]interface{})
		row2, ok2 := arr2[i].([]interface{})

		if !ok1 || !ok2 {
			t.Errorf("Invalid %s row format at index %d", name, i)
			continue
		}

		if len(row1) != len(row2) {
			t.Errorf("%s column count mismatch at row %d: %d vs %d",
				name, i, len(row1), len(row2))
			continue
		}

		for j := range row1 {
			v1 := toFloat64(row1[j])
			v2 := toFloat64(row2[j])

			// Check for sign flips in eigenvectors (both signs are valid)
			if math.Abs(v1+v2) < tolerance {
				// Values are opposite signs, this is OK for eigenvectors
				continue
			}

			if math.Abs(v1-v2) > tolerance {
				t.Errorf("%s mismatch at [%d,%d]: %.10f vs %.10f (diff=%.10f)",
					name, i, j, v1, v2, math.Abs(v1-v2))
			}
		}
	}
}

// Compare vector values
func compareVectorValues(t *testing.T, name string, val1, val2 interface{}, tolerance float64) {
	t.Helper()

	arr1 := toFloatSlice(val1)
	arr2 := toFloatSlice(val2)

	if len(arr1) != len(arr2) {
		t.Errorf("%s length mismatch: %d vs %d", name, len(arr1), len(arr2))
		return
	}

	for i := range arr1 {
		if math.Abs(arr1[i]-arr2[i]) > tolerance {
			t.Errorf("%s mismatch at [%d]: %.10f vs %.10f (diff=%.10f)",
				name, i, arr1[i], arr2[i], math.Abs(arr1[i]-arr2[i]))
		}
	}
}

// Extract explained variance from results
func extractExplainedVariance(t *testing.T, results map[string]interface{}) []float64 {
	t.Helper()

	// Look for explained variance in model.explained_variance
	if model, ok := results["model"].(map[string]interface{}); ok {
		if varInterface, ok := model["explained_variance"]; ok {
			return toFloatSlice(varInterface)
		}
	}

	t.Fatal("No explained_variance in model")
	return nil
}

// Convert interface to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

// Convert interface to float slice
func toFloatSlice(v interface{}) []float64 {
	switch val := v.(type) {
	case []float64:
		return val
	case []interface{}:
		result := make([]float64, len(val))
		for i, item := range val {
			result[i] = toFloat64(item)
		}
		return result
	default:
		return nil
	}
}
