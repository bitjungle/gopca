// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package testutil

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

const (
	// DefaultTolerance is the default numerical tolerance for floating point comparisons
	DefaultTolerance = 1e-10
	// LooseTolerance is used for less strict comparisons
	LooseTolerance = 1e-6
	// StrictTolerance is used for very strict comparisons
	StrictTolerance = 1e-14
)

// AlmostEqual checks if two float64 values are approximately equal within tolerance
func AlmostEqual(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}
	return math.Abs(a-b) <= tolerance
}

// AssertAlmostEqual checks if two values are almost equal and fails the test if not
func AssertAlmostEqual(t *testing.T, expected, actual, tolerance float64, message string) {
	t.Helper()
	if !AlmostEqual(expected, actual, tolerance) {
		t.Errorf("%s: expected %v, got %v (tolerance %v)", message, expected, actual, tolerance)
	}
}

// AssertMatrixAlmostEqual checks if two matrices are almost equal element-wise
func AssertMatrixAlmostEqual(t *testing.T, expected, actual types.Matrix, tolerance float64, message string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s: row count mismatch - expected %d, got %d", message, len(expected), len(actual))
		return
	}

	if len(expected) == 0 {
		return
	}

	if len(expected[0]) != len(actual[0]) {
		t.Errorf("%s: column count mismatch - expected %d, got %d", message, len(expected[0]), len(actual[0]))
		return
	}

	for i := range expected {
		for j := range expected[i] {
			if !AlmostEqual(expected[i][j], actual[i][j], tolerance) {
				t.Errorf("%s: element [%d,%d] mismatch - expected %v, got %v",
					message, i, j, expected[i][j], actual[i][j])
				return
			}
		}
	}
}

// AssertSliceAlmostEqual checks if two slices are almost equal element-wise
func AssertSliceAlmostEqual(t *testing.T, expected, actual []float64, tolerance float64, message string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s: length mismatch - expected %d, got %d", message, len(expected), len(actual))
		return
	}

	for i := range expected {
		if !AlmostEqual(expected[i], actual[i], tolerance) {
			t.Errorf("%s: element [%d] mismatch - expected %v, got %v",
				message, i, expected[i], actual[i])
			return
		}
	}
}

// GenerateTestMatrix creates a test matrix with predictable values
func GenerateTestMatrix(rows, cols int, seed float64) types.Matrix {
	matrix := make(types.Matrix, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			// Simple deterministic formula for test data
			matrix[i][j] = seed + float64(i*cols+j)*0.1
		}
	}
	return matrix
}

// GenerateRandomMatrix creates a test matrix with pseudo-random values
func GenerateRandomMatrix(rows, cols int) types.Matrix {
	matrix := make(types.Matrix, rows)
	// Use a simple linear congruential generator for reproducibility
	var seed int64 = 12345
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			seed = (seed*1103515245 + 12345) & 0x7fffffff
			matrix[i][j] = float64(seed) / float64(0x7fffffff)
		}
	}
	return matrix
}

// GenerateIdentityMatrix creates an identity matrix
func GenerateIdentityMatrix(size int) types.Matrix {
	matrix := make(types.Matrix, size)
	for i := range matrix {
		matrix[i] = make([]float64, size)
		matrix[i][i] = 1.0
	}
	return matrix
}

// CreateTestPCAConfig creates a basic PCA configuration for testing
func CreateTestPCAConfig(components int) types.PCAConfig {
	return types.PCAConfig{
		Components:    components,
		Method:        "svd",
		MeanCenter:    true,
		StandardScale: false,
	}
}

// AssertNoError checks that an error is nil and fails the test if not
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError checks that an error is not nil and fails the test if it is
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", message)
	}
}

// CompareMatrixDimensions checks if two matrices have the same dimensions
func CompareMatrixDimensions(a, b types.Matrix) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	return len(a[0]) == len(b[0])
}
