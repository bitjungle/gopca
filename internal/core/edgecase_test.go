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

// TestPCAWithEmptyData tests PCA behavior with empty or invalid data
func TestPCAWithEmptyData(t *testing.T) {
	testCases := []struct {
		name      string
		data      types.Matrix
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Empty matrix (0x0)",
			data:      types.Matrix{},
			expectErr: true,
			errMsg:    "empty data",
		},
		{
			name:      "Empty rows (0xN)",
			data:      types.Matrix{},
			expectErr: true,
			errMsg:    "empty data",
		},
		{
			name:      "Empty columns (Nx0)",
			data:      types.Matrix{{}, {}, {}},
			expectErr: true,
			errMsg:    "empty columns",
		},
		{
			name:      "Nil data",
			data:      nil,
			expectErr: true,
			errMsg:    "nil data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: false,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				}
			}
		})
	}
}

// TestPCAWithSingleRowColumn tests PCA with single row or column data
func TestPCAWithSingleRowColumn(t *testing.T) {
	testCases := []struct {
		name   string
		data   types.Matrix
		method string
	}{
		{
			name:   "Single row (1xN) SVD",
			data:   types.Matrix{{1.0, 2.0, 3.0, 4.0, 5.0}},
			method: "svd",
		},
		{
			name:   "Single row (1xN) NIPALS",
			data:   types.Matrix{{1.0, 2.0, 3.0, 4.0, 5.0}},
			method: "nipals",
		},
		{
			name:   "Single column (Nx1) SVD",
			data:   types.Matrix{{1.0}, {2.0}, {3.0}, {4.0}, {5.0}},
			method: "svd",
		},
		{
			name:   "Single column (Nx1) NIPALS",
			data:   types.Matrix{{1.0}, {2.0}, {3.0}, {4.0}, {5.0}},
			method: "nipals",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    1,
				MeanCenter:    true,
				StandardScale: false,
				Method:        tc.method,
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			// Single row after mean centering becomes all zeros - should handle gracefully
			if len(tc.data) == 1 {
				// Expect error or handle specially
				if err == nil && result != nil {
					// Check that explained variance is near zero or NaN
					if len(result.ExplainedVarRatio) > 0 && result.ExplainedVarRatio[0] > 1e-10 {
						t.Errorf("Non-zero explained variance for single row: %.6f", result.ExplainedVarRatio[0])
					}
				}
			} else if len(tc.data[0]) == 1 {
				// Single column should work
				if err != nil {
					t.Errorf("Unexpected error for single column: %v", err)
				}
				if result != nil {
					// Should have one component that explains all variance
					if len(result.ExplainedVarRatio) > 0 && result.ExplainedVarRatio[0] < 99.9 {
						t.Errorf("Single column should explain ~100%% variance, got %.2f%%", result.ExplainedVarRatio[0])
					}
				}
			}
		})
	}
}

// TestPCAWithConstantColumns tests PCA with zero variance columns
func TestPCAWithConstantColumns(t *testing.T) {
	testCases := []struct {
		name   string
		data   types.Matrix
		method string
	}{
		{
			name: "All constant columns",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
				{1.0, 2.0, 3.0},
				{1.0, 2.0, 3.0},
				{1.0, 2.0, 3.0},
			},
			method: "svd",
		},
		{
			name: "Mixed constant and variable columns",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
				{1.0, 2.5, 3.0},
				{1.0, 3.0, 3.0},
				{1.0, 3.5, 3.0},
			},
			method: "svd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" "+tc.method, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: false,
				Method:        tc.method,
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			if err != nil {
				// It's acceptable to error on constant columns
				t.Logf("PCA returned error for constant columns (acceptable): %v", err)
			} else if result != nil {
				// Check that explained variance makes sense
				totalVar := 0.0
				for _, v := range result.ExplainedVarRatio {
					totalVar += v
				}
				if totalVar > 100.1 {
					t.Errorf("Total explained variance %.2f%% > 100%%", totalVar)
				}
			}
		})
	}
}

// TestPCAWithExtremeValues tests PCA with very small or large values
func TestPCAWithExtremeValues(t *testing.T) {
	testCases := []struct {
		name   string
		data   types.Matrix
		method string
	}{
		{
			name: "Very small values (near epsilon)",
			data: types.Matrix{
				{1e-15, 2e-15, 3e-15},
				{4e-15, 5e-15, 6e-15},
				{7e-15, 8e-15, 9e-15},
			},
			method: "svd",
		},
		{
			name: "Very large values",
			data: types.Matrix{
				{1e10, 2e10, 3e10},
				{4e10, 5e10, 6e10},
				{7e10, 8e10, 9e10},
			},
			method: "svd",
		},
		{
			name: "Mixed scale values",
			data: types.Matrix{
				{1e-10, 1e0, 1e10},
				{2e-10, 2e0, 2e10},
				{3e-10, 3e0, 3e10},
			},
			method: "svd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: true, // Standardization is important for mixed scales
				Method:        tc.method,
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			if err != nil {
				t.Errorf("PCA failed with extreme values: %v", err)
			}

			if result != nil {
				// Check for NaN or Inf in results
				for i, score := range result.Scores {
					for j, val := range score {
						if math.IsNaN(val) || math.IsInf(val, 0) {
							t.Errorf("Invalid score[%d][%d]: %v", i, j, val)
						}
					}
				}
			}
		})
	}
}

// TestPCAWithWideData tests PCA with more variables than samples
func TestPCAWithWideData(t *testing.T) {
	// Create wide data (5 samples, 20 variables)
	data := make(types.Matrix, 5)
	for i := range data {
		data[i] = make([]float64, 20)
		for j := range data[i] {
			data[i][j] = float64(i+1)*float64(j+1) + float64(i*j)
		}
	}

	methods := []string{"svd", "nipals"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    4, // Max meaningful components is min(n-1, p) = 4
				MeanCenter:    true,
				StandardScale: false,
				Method:        method,
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)

			if err != nil {
				t.Errorf("PCA failed with wide data: %v", err)
			}

			if result != nil {
				// Should have at most 4 non-zero components
				if result.ComponentsComputed > 4 {
					t.Errorf("Too many components computed: %d (expected <= 4)", result.ComponentsComputed)
				}

				// Check that later components have very small variance
				if len(result.ExplainedVar) > 4 {
					for i := 4; i < len(result.ExplainedVar); i++ {
						if result.ExplainedVar[i] > 1e-10 {
							t.Errorf("Component %d has non-zero variance %.6e for wide data", i+1, result.ExplainedVar[i])
						}
					}
				}
			}
		})
	}
}

// TestPCAWithTooManyComponents tests requesting more components than possible
func TestPCAWithTooManyComponents(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{10.0, 11.0, 12.0},
	}

	config := types.PCAConfig{
		Components:    10, // Requesting 10 components for 4x3 data
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}

	engine := NewPCAEngine()
	result, err := engine.Fit(data, config)

	if err != nil {
		// It's ok to error on too many components
		t.Logf("PCA returned error for too many components (acceptable): %v", err)
	} else if result != nil {
		// Should automatically limit to max possible components
		maxComponents := min(len(data)-1, len(data[0]))
		if result.ComponentsComputed > maxComponents {
			t.Errorf("Computed %d components, expected at most %d", result.ComponentsComputed, maxComponents)
		}
	}
}

// TestPCAWithNaNInfValues tests PCA behavior with NaN and Inf values
func TestPCAWithNaNInfValues(t *testing.T) {
	testCases := []struct {
		name      string
		data      types.Matrix
		expectErr bool
	}{
		{
			name: "Data with NaN",
			data: types.Matrix{
				{1.0, 2.0, math.NaN()},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			expectErr: true,
		},
		{
			name: "Data with Inf",
			data: types.Matrix{
				{1.0, 2.0, math.Inf(1)},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			expectErr: true,
		},
		{
			name: "Data with -Inf",
			data: types.Matrix{
				{1.0, 2.0, math.Inf(-1)},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: false,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
				if result == nil {
					t.Errorf("Expected result for %s but got nil", tc.name)
				}
			}
		})
	}
}

// TestPCAPreprocessingStability tests preprocessing with edge cases
func TestPCAPreprocessingStability(t *testing.T) {
	testCases := []struct {
		name          string
		data          types.Matrix
		standardScale bool
		desc          string
	}{
		{
			name: "Zero variance column with standardization",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
				{1.0, 3.0, 4.0},
				{1.0, 4.0, 5.0},
				{1.0, 5.0, 6.0},
			},
			standardScale: true,
			desc:          "First column has zero variance",
		},
		{
			name: "Near-zero variance with standardization",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
				{1.0 + 1e-15, 3.0, 4.0},
				{1.0 - 1e-15, 4.0, 5.0},
				{1.0, 5.0, 6.0},
			},
			standardScale: true,
			desc:          "First column has near-zero variance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: tc.standardScale,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(tc.data, config)

			// Should handle gracefully without panicking
			if err != nil {
				t.Logf("PCA returned error for %s (acceptable): %v", tc.desc, err)
			} else if result != nil {
				// Check that results are valid
				for _, ratio := range result.ExplainedVarRatio {
					if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
						t.Errorf("Invalid explained variance ratio for %s", tc.desc)
					}
				}
			}
		})
	}
}
