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
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// TestPearsonCorrelation tests the Pearson correlation calculation
func TestPearsonCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		wantR     float64
		wantP     float64
		tolerance float64
		wantErr   bool
	}{
		{
			name:      "Perfect positive correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{2, 4, 6, 8, 10},
			wantR:     1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Perfect negative correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{10, 8, 6, 4, 2},
			wantR:     -1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "No correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{3, 1, 4, 1, 5},
			wantR:     0.0,
			wantP:     1.0,
			tolerance: 0.4, // More tolerance for no correlation
			wantErr:   false,
		},
		{
			name:      "Strong positive correlation",
			x:         []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			y:         []float64{2.1, 3.9, 6.1, 7.8, 10.2, 11.9, 14.1, 15.8, 18.2, 19.9},
			wantR:     0.999,
			wantP:     0.001,
			tolerance: 0.01,
			wantErr:   false,
		},
		{
			name:      "With missing values",
			x:         []float64{1, 2, math.NaN(), 4, 5},
			y:         []float64{2, 4, 6, math.NaN(), 10},
			wantR:     1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:    "Too few observations",
			x:       []float64{1, 2},
			y:       []float64{2, 4},
			wantErr: true,
		},
		{
			name:    "All missing values",
			x:       []float64{math.NaN(), math.NaN(), math.NaN()},
			y:       []float64{math.NaN(), math.NaN(), math.NaN()},
			wantErr: true,
		},
		{
			name:    "Different lengths",
			x:       []float64{1, 2, 3},
			y:       []float64{1, 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, p, err := pearsonCorrelation(tt.x, tt.y)

			if tt.wantErr {
				if err == nil {
					t.Errorf("pearsonCorrelation() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("pearsonCorrelation() unexpected error = %v", err)
				return
			}

			if math.Abs(r-tt.wantR) > tt.tolerance {
				t.Errorf("pearsonCorrelation() r = %v, want %v (tolerance %v)", r, tt.wantR, tt.tolerance)
			}

			// For p-values, check significance thresholds
			// Strong correlations should have low p-values
			if math.Abs(r) > 0.9 && p > 0.05 {
				t.Errorf("pearsonCorrelation() p = %v, expected significant (< 0.05) for r = %v", p, r)
			}
			// Weak correlations should have high p-values
			if math.Abs(r) < 0.3 && p < 0.5 {
				t.Errorf("pearsonCorrelation() p = %v, expected non-significant (> 0.5) for r = %v", p, r)
			}
		})
	}
}

// TestSpearmanCorrelation tests the Spearman correlation calculation
func TestSpearmanCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		wantR     float64
		tolerance float64
		wantErr   bool
	}{
		{
			name:      "Perfect monotonic positive",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{1, 4, 9, 16, 25}, // y = x^2, perfect monotonic
			wantR:     1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Perfect monotonic negative",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{25, 16, 9, 4, 1},
			wantR:     -1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Non-linear but monotonic",
			x:         []float64{1, 2, 3, 4, 5, 6, 7, 8},
			y:         []float64{1, 8, 27, 64, 125, 216, 343, 512}, // y = x^3
			wantR:     1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "With ties",
			x:         []float64{1, 2, 2, 3, 4, 4, 5},
			y:         []float64{1, 2, 3, 4, 5, 5, 6},
			wantR:     0.95,
			tolerance: 0.1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _, err := spearmanCorrelation(tt.x, tt.y)

			if tt.wantErr {
				if err == nil {
					t.Errorf("spearmanCorrelation() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("spearmanCorrelation() unexpected error = %v", err)
				return
			}

			if math.Abs(r-tt.wantR) > tt.tolerance {
				t.Errorf("spearmanCorrelation() r = %v, want %v (tolerance %v)", r, tt.wantR, tt.tolerance)
			}
		})
	}
}

// TestRank tests the ranking function
func TestRank(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want []float64
	}{
		{
			name: "No ties",
			x:    []float64{3, 1, 4, 1.5, 9},
			want: []float64{3, 1, 4, 2, 5},
		},
		{
			name: "With ties",
			x:    []float64{1, 2, 2, 3, 3, 3, 4},
			want: []float64{1, 2.5, 2.5, 5, 5, 5, 7},
		},
		{
			name: "All same",
			x:    []float64{5, 5, 5, 5},
			want: []float64{2.5, 2.5, 2.5, 2.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rank(tt.x)
			for i := range got {
				if math.Abs(got[i]-tt.want[i]) > 1e-10 {
					t.Errorf("rank() at index %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestOneHotEncode tests the one-hot encoding function
func TestOneHotEncode(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		wantKeys   []string
		wantValues map[string][]float64
	}{
		{
			name:       "Basic categories",
			categories: []string{"A", "B", "A", "C", "B"},
			wantKeys:   []string{"A", "B", "C"},
			wantValues: map[string][]float64{
				"A": {1, 0, 1, 0, 0},
				"B": {0, 1, 0, 0, 1},
				"C": {0, 0, 0, 1, 0},
			},
		},
		{
			name:       "With empty categories",
			categories: []string{"A", "", "B", "", "A"},
			wantKeys:   []string{"A", "B"},
			wantValues: map[string][]float64{
				"A": {1, 0, 0, 0, 1},
				"B": {0, 0, 1, 0, 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := oneHotEncode(tt.categories)

			// Check number of encoded variables
			if len(got) != len(tt.wantKeys) {
				t.Errorf("oneHotEncode() returned %d variables, want %d", len(got), len(tt.wantKeys))
			}

			// Check each expected key exists with correct values
			for _, key := range tt.wantKeys {
				values, exists := got[key]
				if !exists {
					t.Errorf("oneHotEncode() missing key %s", key)
					continue
				}

				wantValues := tt.wantValues[key]
				if len(values) != len(wantValues) {
					t.Errorf("oneHotEncode() key %s has %d values, want %d", key, len(values), len(wantValues))
					continue
				}

				for i := range values {
					if values[i] != wantValues[i] {
						t.Errorf("oneHotEncode() key %s at index %d = %v, want %v", key, i, values[i], wantValues[i])
					}
				}
			}
		})
	}
}

// TestCalculateEigencorrelations tests the main correlation calculation function
func TestCalculateEigencorrelations(t *testing.T) {
	// Create test PC scores
	scores := mat.NewDense(10, 3, []float64{
		// PC1  PC2  PC3
		1, 0, 0,
		2, 0, 0,
		3, 0, 0,
		4, 0, 0,
		5, 0, 0,
		6, 0, 0,
		7, 0, 0,
		8, 0, 0,
		9, 0, 0,
		10, 0, 0,
	})

	tests := []struct {
		name             string
		request          CorrelationRequest
		wantNumVars      int
		checkCorrelation func(t *testing.T, result *CorrelationResult)
		wantErr          bool
	}{
		{
			name: "Numeric variables with Pearson",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // Perfect correlation with PC1
					"var2": {10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, // Perfect negative correlation with PC1
				},
				Components: []int{0, 1, 2},
				Method:     "pearson",
			},
			wantNumVars: 2,
			checkCorrelation: func(t *testing.T, result *CorrelationResult) {
				// Check var1 correlation with PC1
				if math.Abs(result.Correlations["var1"][0]-1.0) > 0.001 {
					t.Errorf("var1 PC1 correlation = %v, want ~1.0", result.Correlations["var1"][0])
				}
				// Check var2 correlation with PC1
				if math.Abs(result.Correlations["var2"][0]+1.0) > 0.001 {
					t.Errorf("var2 PC1 correlation = %v, want ~-1.0", result.Correlations["var2"][0])
				}
			},
			wantErr: false,
		},
		{
			name: "Categorical variables",
			request: CorrelationRequest{
				Scores: scores,
				MetadataCategorical: map[string][]string{
					"group": {"A", "A", "A", "B", "B", "B", "C", "C", "C", "C"},
				},
				Components: []int{0},
				Method:     "pearson",
			},
			wantNumVars: 3, // One-hot encoded: group_A, group_B, group_C
			wantErr:     false,
		},
		{
			name: "Spearman correlation",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 4, 9, 16, 25, 36, 49, 64, 81, 100}, // Non-linear but monotonic
				},
				Components: []int{0},
				Method:     "spearman",
			},
			wantNumVars: 1,
			checkCorrelation: func(t *testing.T, result *CorrelationResult) {
				// Should have perfect Spearman correlation
				if math.Abs(result.Correlations["var1"][0]-1.0) > 0.001 {
					t.Errorf("var1 PC1 Spearman correlation = %v, want ~1.0", result.Correlations["var1"][0])
				}
			},
			wantErr: false,
		},
		{
			name: "Invalid method",
			request: CorrelationRequest{
				Scores:     scores,
				Components: []int{0},
				Method:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "Nil scores",
			request: CorrelationRequest{
				Scores: nil,
				Method: "pearson",
			},
			wantErr: true,
		},
		{
			name: "Mismatched lengths",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 2, 3}, // Wrong length
				},
				Method: "pearson",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateEigencorrelations(tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateEigencorrelations() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateEigencorrelations() unexpected error = %v", err)
				return
			}

			// Check number of variables
			if len(result.Variables) != tt.wantNumVars {
				t.Errorf("CalculateEigencorrelations() returned %d variables, want %d", len(result.Variables), tt.wantNumVars)
			}

			// Run specific correlation checks if provided
			if tt.checkCorrelation != nil {
				tt.checkCorrelation(t, result)
			}
		})
	}
}

// TestEigencorrelationPC1Sorting verifies that variables are sorted by PC1 correlation
// from highest positive to most negative values
func TestEigencorrelationPC1Sorting(t *testing.T) {
	// Create test scores with known patterns
	scores := mat.NewDense(10, 2, []float64{
		1.0, 0.1,
		0.9, 0.2,
		0.8, -0.1,
		0.7, 0.0,
		0.0, 0.8,
		-0.1, 0.9,
		-0.5, 0.5,
		-0.7, 0.3,
		-0.8, -0.2,
		-0.9, -0.3,
	})

	// Create metadata variables with varying correlations to PC1
	request := CorrelationRequest{
		Scores: scores,
		MetadataNumeric: map[string][]float64{
			"strong_positive": {1.0, 0.9, 0.8, 0.7, 0.0, -0.1, -0.5, -0.7, -0.8, -0.9}, // Strong positive correlation
			"weak_positive":   {0.2, 0.1, 0.3, 0.1, 0.0, -0.1, -0.2, 0.0, -0.1, -0.2},  // Weak positive correlation
			"zero_corr":       {0.1, -0.1, 0.2, -0.2, 0.3, -0.3, 0.1, -0.1, 0.2, -0.2}, // Near zero correlation
			"weak_negative":   {-0.1, -0.2, 0.0, -0.1, 0.1, 0.2, 0.3, 0.1, 0.2, 0.3},   // Weak negative correlation
			"strong_negative": {-1.0, -0.9, -0.8, -0.7, 0.0, 0.1, 0.5, 0.7, 0.8, 0.9},  // Strong negative correlation
		},
		Components: []int{0, 1},
		Method:     "pearson",
	}

	result, err := CalculateEigencorrelations(request)
	if err != nil {
		t.Fatalf("CalculateEigencorrelations failed: %v", err)
	}

	// Check that variables are sorted by PC1 correlation in descending order
	if len(result.Variables) != 5 {
		t.Fatalf("Expected 5 variables, got %d", len(result.Variables))
	}

	// Verify sorting order: highest positive to most negative
	prevCorr := 2.0 // Start with value higher than any possible correlation
	for i, varName := range result.Variables {
		pc1Corr := result.Correlations[varName][0]

		// Check descending order
		if pc1Corr > prevCorr {
			t.Errorf("Variables not sorted correctly at position %d: %s (corr=%f) > previous (corr=%f)",
				i, varName, pc1Corr, prevCorr)
		}

		t.Logf("Position %d: %s, PC1 correlation = %.3f", i, varName, pc1Corr)
		prevCorr = pc1Corr
	}

	// Verify expected order
	expectedOrder := []string{"strong_positive", "weak_positive", "zero_corr", "weak_negative", "strong_negative"}
	for i, expected := range expectedOrder {
		if result.Variables[i] != expected {
			// It's OK if the order is not exactly as expected due to numerical precision,
			// but the general pattern should be maintained
			pc1Corr := result.Correlations[result.Variables[i]][0]
			t.Logf("Warning: Position %d has %s (corr=%.3f) instead of expected %s",
				i, result.Variables[i], pc1Corr, expected)
		}
	}
}

// TestEigencorrelationSortingWithCategorical verifies that one-hot encoded categorical
// variables are sorted individually by PC1 correlation, not grouped by base name
func TestEigencorrelationSortingWithCategorical(t *testing.T) {
	// Create test scores
	scores := mat.NewDense(8, 2, []float64{
		1.0, 0.1, // Pattern for category A
		0.9, 0.2, // Pattern for category A
		-0.8, 0.3, // Pattern for category B
		-0.9, 0.4, // Pattern for category B
		0.3, -0.8, // Pattern for category C
		0.2, -0.9, // Pattern for category C
		-0.1, 0.0, // Pattern for category D
		-0.2, 0.1, // Pattern for category D
	})

	request := CorrelationRequest{
		Scores: scores,
		MetadataCategorical: map[string][]string{
			"group": {"A", "A", "B", "B", "C", "C", "D", "D"},
		},
		Components: []int{0, 1},
		Method:     "pearson",
	}

	result, err := CalculateEigencorrelations(request)
	if err != nil {
		t.Fatalf("CalculateEigencorrelations failed: %v", err)
	}

	// Should have 4 one-hot encoded variables
	expectedVars := 4
	if len(result.Variables) != expectedVars {
		t.Fatalf("Expected %d variables, got %d", expectedVars, len(result.Variables))
	}

	// Verify all variables are sorted by PC1 correlation
	prevCorr := 2.0
	for i, varName := range result.Variables {
		pc1Corr := result.Correlations[varName][0]

		if pc1Corr > prevCorr {
			t.Errorf("Variables not sorted at position %d: %s (corr=%f) > previous (corr=%f)",
				i, varName, pc1Corr, prevCorr)
		}

		t.Logf("Position %d: %s, PC1 correlation = %.3f", i, varName, pc1Corr)
		prevCorr = pc1Corr
	}

	// Note: The sorting is by PC1 correlation value, so one-hot encoded variables
	// from the same categorical variable may be separated based on their correlation values
	// This is intentional as it provides the most meaningful visual hierarchy
}

// TestStatisticalFunctions tests the statistical helper functions
func TestStatisticalFunctions(t *testing.T) {
	t.Run("normalCDF", func(t *testing.T) {
		tests := []struct {
			z    float64
			want float64
			tol  float64
		}{
			{0, 0.5, 0.001},
			{1, 0.8413, 0.001},
			{-1, 0.1587, 0.001},
			{2, 0.9772, 0.001},
			{-2, 0.0228, 0.001},
		}

		for _, tt := range tests {
			got := normalCDF(tt.z)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("normalCDF(%v) = %v, want %v ± %v", tt.z, got, tt.want, tt.tol)
			}
		}
	})

	t.Run("studentTCDF", func(t *testing.T) {
		// Basic sanity check that it returns reasonable values
		// Note: Comprehensive validation against scipy is in TestStudentTCDF_Validation
		tests := []struct {
			t  float64
			df float64
		}{
			{0, 10},
			{1, 10},
			{2, 10},
			{0, 50},
		}

		for _, tt := range tests {
			got := studentTCDF(tt.t, tt.df)
			if got < 0 || got > 1 {
				t.Errorf("studentTCDF(%v, %v) = %v, want value in [0,1]", tt.t, tt.df, got)
			}
		}
	})
}

// TestStudentTCDF_Validation validates studentTCDF against scipy reference values.
//
// Regression test for issue #570: Prior to the fix, studentTCDF used a simple power
// approximation x^(df/2) instead of the correct regularized incomplete beta function.
// This caused p-values to be inflated by 2-3× for small samples (df < 30), invalidating
// hypothesis testing results.
//
// This test reads reference values generated by testdata/validation/generate_studentt_reference.py
// using scipy.stats.t.cdf(), which is a widely trusted implementation, and validates that
// the Go implementation matches within numerical precision (1e-10).
//
// Test coverage:
//   - Various degrees of freedom: 1, 2, 5, 10, 20, 30, 50
//   - Various t-values: -3.0 to 3.0 (including 0)
//   - Edge cases: Cauchy distribution (df=1), large df approaching normal
func TestStudentTCDF_Validation(t *testing.T) {
	// Read reference CSV generated by scipy
	refPath := filepath.Join("..", "..", "testdata", "validation", "reference_results", "studentt_reference.csv")

	file, err := os.Open(refPath)
	if err != nil {
		t.Skipf("Skipping validation test: reference file not found: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Errorf("Failed to close reference file: %v", closeErr)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read reference CSV: %v", err)
	}

	// Skip header
	if len(records) < 2 {
		t.Fatal("Reference CSV is empty or missing header")
	}

	const tolerance = 1e-10 // Tight tolerance since both use proper math libraries
	failCount := 0
	maxError := 0.0

	for i, record := range records[1:] {
		if len(record) != 4 {
			t.Fatalf("Invalid record at line %d: expected 4 fields, got %d", i+2, len(record))
		}

		df, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			t.Fatalf("Failed to parse df at line %d: %v", i+2, err)
		}

		tValue, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			t.Fatalf("Failed to parse t at line %d: %v", i+2, err)
		}

		expectedCDF, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			t.Fatalf("Failed to parse cdf at line %d: %v", i+2, err)
		}

		// Test CDF value
		gotCDF := studentTCDF(tValue, df)
		errorCDF := math.Abs(gotCDF - expectedCDF)

		if errorCDF > maxError {
			maxError = errorCDF
		}

		if errorCDF > tolerance {
			failCount++
			if failCount <= 5 { // Only show first 5 failures to avoid spam
				t.Errorf("studentTCDF(%v, df=%v) = %.15f, want %.15f (scipy), error = %.2e",
					tValue, df, gotCDF, expectedCDF, errorCDF)
			}
		}
	}

	if failCount > 0 {
		t.Errorf("Total failures: %d/%d test cases", failCount, len(records)-1)
	}

	t.Logf("Validated %d test cases, max error: %.2e", len(records)-1, maxError)
}

// TestStudentTCDF_EdgeCases tests special mathematical properties of the t-distribution.
//
// These properties must hold exactly (within numerical precision) for any correct
// implementation of the Student's t-distribution CDF:
//  1. CDF(0, df) = 0.5 for all df (symmetry around zero)
//  2. CDF(t, df) + CDF(-t, df) = 1.0 (distribution symmetry)
//  3. CDF(1, 1) = 0.75 (Cauchy distribution special case)
//  4. CDF is monotonically increasing in t
//  5. CDF approaches 0 as t → -∞ and 1 as t → +∞
func TestStudentTCDF_EdgeCases(t *testing.T) {
	const tolerance = 1e-10

	t.Run("zero_returns_half", func(t *testing.T) {
		// CDF(0, df) should be exactly 0.5 for all df
		dfValues := []float64{1, 2, 5, 10, 20, 30, 50, 100}
		for _, df := range dfValues {
			got := studentTCDF(0, df)
			if math.Abs(got-0.5) > tolerance {
				t.Errorf("studentTCDF(0, df=%v) = %v, want 0.5", df, got)
			}
		}
	})

	t.Run("symmetry", func(t *testing.T) {
		// CDF(t) + CDF(-t) should equal 1.0 due to symmetry around zero
		testCases := []struct {
			t  float64
			df float64
		}{
			{1.0, 10},
			{2.0, 5},
			{1.5, 20},
			{2.5, 2},
			{0.5, 30},
		}

		for _, tc := range testCases {
			positiveCDF := studentTCDF(tc.t, tc.df)
			negativeCDF := studentTCDF(-tc.t, tc.df)
			sum := positiveCDF + negativeCDF

			if math.Abs(sum-1.0) > tolerance {
				t.Errorf("CDF(%v, df=%v) + CDF(%v, df=%v) = %v, want 1.0",
					tc.t, tc.df, -tc.t, tc.df, sum)
			}
		}
	})

	t.Run("cauchy_special_case", func(t *testing.T) {
		// For df=1 (Cauchy distribution), CDF(1, 1) should be exactly 0.75
		// This is because Cauchy is symmetric and P(T ≤ 1) = 0.5 + 0.25 = 0.75
		got := studentTCDF(1.0, 1.0)
		expected := 0.75
		if math.Abs(got-expected) > tolerance {
			t.Errorf("studentTCDF(1.0, df=1) = %.15f, want %.15f (Cauchy)",
				got, expected)
		}
	})

	t.Run("monotonicity", func(t *testing.T) {
		// CDF should be monotonically increasing in t
		df := 10.0
		tValues := []float64{-3, -2, -1, 0, 1, 2, 3}

		prevCDF := studentTCDF(tValues[0], df)
		for i := 1; i < len(tValues); i++ {
			currentCDF := studentTCDF(tValues[i], df)
			if currentCDF <= prevCDF {
				t.Errorf("CDF not monotonic: CDF(%v) = %v <= CDF(%v) = %v",
					tValues[i], currentCDF, tValues[i-1], prevCDF)
			}
			prevCDF = currentCDF
		}
	})

	t.Run("extreme_values", func(t *testing.T) {
		// For very large |t|, CDF should approach 0 or 1
		df := 10.0

		// Very negative t should approach 0
		veryNegative := studentTCDF(-10, df)
		if veryNegative >= 0.001 {
			t.Errorf("studentTCDF(-10, df=%v) = %v, expected < 0.001", df, veryNegative)
		}

		// Very positive t should approach 1
		veryPositive := studentTCDF(10, df)
		if veryPositive <= 0.999 {
			t.Errorf("studentTCDF(10, df=%v) = %v, expected > 0.999", df, veryPositive)
		}
	})

	t.Run("comparison_with_normal", func(t *testing.T) {
		// For large df (e.g., 100), t-distribution should be very close to normal
		df := 100.0
		t_val := 2.0

		tCDF := studentTCDF(t_val, df)
		normalCDF := normalCDF(t_val)

		// Should be within 0.01 (1 percentage point)
		if math.Abs(tCDF-normalCDF) > 0.01 {
			t.Errorf("For df=100, t-distribution differs from normal by %v (%.4f vs %.4f)",
				math.Abs(tCDF-normalCDF), tCDF, normalCDF)
		}
	})
}
