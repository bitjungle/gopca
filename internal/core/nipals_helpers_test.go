// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package core

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestFindMaxVarianceColumn(t *testing.T) {
	tests := []struct {
		name           string
		data           [][]float64
		expectedCol    int
		expectedVarMin float64 // minimum expected variance
	}{
		{
			name: "simple 3x3",
			data: [][]float64{
				{1, 2, 10},
				{2, 3, 20},
				{3, 4, 30},
			},
			expectedCol:    2, // third column has highest variance
			expectedVarMin: 60.0,
		},
		{
			name: "all equal variance",
			data: [][]float64{
				{1, 1, 1},
				{2, 2, 2},
				{3, 3, 3},
			},
			expectedCol:    0, // returns first column when all equal
			expectedVarMin: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			X := mat.NewDense(len(tt.data), len(tt.data[0]), nil)
			for i, row := range tt.data {
				for j, val := range row {
					X.Set(i, j, val)
				}
			}

			col, variance := findMaxVarianceColumn(X)
			if col != tt.expectedCol {
				t.Errorf("Expected column %d, got %d", tt.expectedCol, col)
			}
			if variance < tt.expectedVarMin {
				t.Errorf("Expected variance >= %f, got %f", tt.expectedVarMin, variance)
			}
		})
	}
}

func TestFindMaxVarianceColumnWithMissing(t *testing.T) {
	tests := []struct {
		name        string
		data        [][]float64
		expectedCol int
	}{
		{
			name: "with NaN values",
			data: [][]float64{
				{1, math.NaN(), 10},
				{2, 3, 20},
				{3, math.NaN(), 30},
			},
			expectedCol: 2, // third column has highest variance even with NaN
		},
		{
			name: "column with all NaN",
			data: [][]float64{
				{1, math.NaN(), 10},
				{2, math.NaN(), 20},
				{3, math.NaN(), 30},
			},
			expectedCol: 2, // should skip column 1 (all NaN)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			X := mat.NewDense(len(tt.data), len(tt.data[0]), nil)
			for i, row := range tt.data {
				for j, val := range row {
					X.Set(i, j, val)
				}
			}

			col, variance := findMaxVarianceColumnWithMissing(X)
			if col != tt.expectedCol {
				t.Errorf("Expected column %d, got %d", tt.expectedCol, col)
			}
			if variance <= 0 {
				t.Errorf("Expected positive variance, got %f", variance)
			}
		})
	}
}

func TestComputeColumnMeansWithMissing(t *testing.T) {
	data := [][]float64{
		{1, math.NaN(), 10},
		{2, 3, 20},
		{3, math.NaN(), 30},
	}

	X := mat.NewDense(3, 3, nil)
	for i, row := range data {
		for j, val := range row {
			X.Set(i, j, val)
		}
	}

	means := computeColumnMeansWithMissing(X)

	// Column 0: (1+2+3)/3 = 2
	if math.Abs(means[0]-2.0) > 1e-10 {
		t.Errorf("Column 0 mean: expected 2.0, got %f", means[0])
	}

	// Column 1: only value 3 (others are NaN)
	if math.Abs(means[1]-3.0) > 1e-10 {
		t.Errorf("Column 1 mean: expected 3.0, got %f", means[1])
	}

	// Column 2: (10+20+30)/3 = 20
	if math.Abs(means[2]-20.0) > 1e-10 {
		t.Errorf("Column 2 mean: expected 20.0, got %f", means[2])
	}
}

func TestCenterMatrixWithMissing(t *testing.T) {
	data := [][]float64{
		{1, math.NaN(), 10},
		{2, 3, 20},
		{3, math.NaN(), 30},
	}

	X := mat.NewDense(3, 3, nil)
	for i, row := range data {
		for j, val := range row {
			X.Set(i, j, val)
		}
	}

	means := []float64{2, 3, 20}
	centerMatrixWithMissing(X, means)

	// Check column 0: should be [-1, 0, 1]
	if math.Abs(X.At(0, 0)-(-1.0)) > 1e-10 {
		t.Errorf("X[0,0]: expected -1.0, got %f", X.At(0, 0))
	}

	// Check that NaN values remain NaN
	if !math.IsNaN(X.At(0, 1)) {
		t.Errorf("X[0,1]: expected NaN, got %f", X.At(0, 1))
	}

	// Check column 2: should be [-10, 0, 10]
	if math.Abs(X.At(1, 2)-0.0) > 1e-10 {
		t.Errorf("X[1,2]: expected 0.0, got %f", X.At(1, 2))
	}
}

func TestInitializeScoreVector(t *testing.T) {
	data := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	X := mat.NewDense(3, 3, nil)
	for i, row := range data {
		for j, val := range row {
			X.Set(i, j, val)
		}
	}

	t.Run("initialize with column 1", func(t *testing.T) {
		vec := initializeScoreVector(X, 1, 3)

		// Should be [2, 5, 8]
		expected := []float64{2, 5, 8}
		for i, exp := range expected {
			if math.Abs(vec.AtVec(i)-exp) > 1e-10 {
				t.Errorf("vec[%d]: expected %f, got %f", i, exp, vec.AtVec(i))
			}
		}
	})
}

func TestInitializeScoreVectorWithMissing(t *testing.T) {
	data := [][]float64{
		{1, math.NaN(), 10},
		{2, 3, 20},
		{3, math.NaN(), 30},
	}

	X := mat.NewDense(3, 3, nil)
	for i, row := range data {
		for j, val := range row {
			X.Set(i, j, val)
		}
	}

	t.Run("initialize with column 1 (has NaN)", func(t *testing.T) {
		vec := initializeScoreVectorWithMissing(X, 1, 3)

		// Column 1 has [NaN, 3, NaN], mean = 3
		// Should be [3, 3, 3] (NaN replaced with mean)
		for i := 0; i < 3; i++ {
			if math.Abs(vec.AtVec(i)-3.0) > 1e-10 {
				t.Errorf("vec[%d]: expected 3.0, got %f", i, vec.AtVec(i))
			}
		}
	})
}

func TestComputeLoadingVector(t *testing.T) {
	// Create simple test data
	X := mat.NewDense(3, 2, []float64{
		1, 0,
		0, 1,
		1, 1,
	})

	tVec := mat.NewVecDense(3, []float64{1, 1, 1})

	p, err := computeLoadingVector(X, tVec, 1e-8)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that p is normalized (unit length)
	norm := math.Sqrt(mat.Dot(p, p))
	if math.Abs(norm-1.0) > 1e-10 {
		t.Errorf("Expected unit norm, got %f", norm)
	}
}

func TestComputeLoadingVectorWithMissing(t *testing.T) {
	// Create test data with NaN
	X := mat.NewDense(3, 2, []float64{
		1, math.NaN(),
		0, 1,
		1, 1,
	})

	tVec := mat.NewVecDense(3, []float64{1, 1, 1})

	p, err := computeLoadingVectorWithMissing(X, tVec, 1e-8)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that p is normalized
	pNorm := 0.0
	for j := 0; j < p.Len(); j++ {
		pVal := p.AtVec(j)
		if !math.IsNaN(pVal) {
			pNorm += pVal * pVal
		}
	}
	pNorm = math.Sqrt(pNorm)
	if math.Abs(pNorm-1.0) > 1e-10 {
		t.Errorf("Expected unit norm, got %f", pNorm)
	}
}

func TestUpdateScoreVector(t *testing.T) {
	X := mat.NewDense(3, 2, []float64{
		1, 0,
		0, 1,
		1, 1,
	})

	p := mat.NewVecDense(2, []float64{0.707, 0.707}) // approximately normalized

	tNew, err := updateScoreVector(X, p, 1e-8)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Just check that we get a valid result
	if tNew.Len() != 3 {
		t.Errorf("Expected length 3, got %d", tNew.Len())
	}
}

func TestUpdateScoreVectorWithMissing(t *testing.T) {
	X := mat.NewDense(3, 2, []float64{
		1, math.NaN(),
		0, 1,
		1, 1,
	})

	p := mat.NewVecDense(2, []float64{0.707, 0.707})
	tOld := mat.NewVecDense(3, []float64{1, 1, 1})

	tNew := updateScoreVectorWithMissing(X, p, tOld, 1e-8)

	// Check that we get a valid result
	if tNew.Len() != 3 {
		t.Errorf("Expected length 3, got %d", tNew.Len())
	}

	// Check that no values are NaN
	for i := 0; i < tNew.Len(); i++ {
		if math.IsNaN(tNew.AtVec(i)) {
			t.Errorf("t[%d] is NaN", i)
		}
	}
}

func TestCheckConvergence(t *testing.T) {
	tests := []struct {
		name      string
		t         []float64
		tOld      []float64
		tolerance float64
		expected  bool
	}{
		{
			name:      "converged",
			t:         []float64{1.0, 2.0, 3.0},
			tOld:      []float64{1.0000001, 2.0000001, 3.0000001},
			tolerance: 1e-6,
			expected:  true,
		},
		{
			name:      "not converged",
			t:         []float64{1.0, 2.0, 3.0},
			tOld:      []float64{1.1, 2.1, 3.1},
			tolerance: 1e-6,
			expected:  false,
		},
		{
			name:      "exactly same",
			t:         []float64{1.0, 2.0, 3.0},
			tOld:      []float64{1.0, 2.0, 3.0},
			tolerance: 1e-10,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tVec := mat.NewVecDense(len(tt.t), tt.t)
			tOldVec := mat.NewVecDense(len(tt.tOld), tt.tOld)

			result := checkConvergence(tVec, tOldVec, tt.tolerance)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDeflateMatrix(t *testing.T) {
	// Create simple test data
	X := mat.NewDense(3, 2, []float64{
		2, 2,
		2, 2,
		2, 2,
	})

	tVec := mat.NewVecDense(3, []float64{1, 1, 1})
	p := mat.NewVecDense(2, []float64{1, 1})

	Xorig := mat.DenseCopyOf(X)
	deflateMatrix(X, tVec, p)

	// After deflation, X should be X - t*p^T
	// Each element should be 2 - 1*1 = 1
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			expected := Xorig.At(i, j) - 1.0
			if math.Abs(X.At(i, j)-expected) > 1e-10 {
				t.Errorf("X[%d,%d]: expected %f, got %f", i, j, expected, X.At(i, j))
			}
		}
	}
}

func TestDeflateMatrixWithMissing(t *testing.T) {
	X := mat.NewDense(3, 2, []float64{
		2, math.NaN(),
		2, 2,
		2, 2,
	})

	tData := []float64{1, 1, 1}
	pData := []float64{1, 1}

	deflateMatrixWithMissing(X, tData, pData)

	// Check that NaN remains NaN
	if !math.IsNaN(X.At(0, 1)) {
		t.Errorf("X[0,1]: expected NaN, got %f", X.At(0, 1))
	}

	// Check that non-NaN values are deflated
	// Each should be 2 - 1*1 = 1
	expected := 1.0
	if math.Abs(X.At(1, 1)-expected) > 1e-10 {
		t.Errorf("X[1,1]: expected %f, got %f", expected, X.At(1, 1))
	}
}

func TestExtractVectorData(t *testing.T) {
	vec := mat.NewVecDense(4, []float64{1, 2, 3, 4})
	data := extractVectorData(vec)

	if len(data) != 4 {
		t.Fatalf("Expected length 4, got %d", len(data))
	}

	for i, expected := range []float64{1, 2, 3, 4} {
		if math.Abs(data[i]-expected) > 1e-10 {
			t.Errorf("data[%d]: expected %f, got %f", i, expected, data[i])
		}
	}
}

// Edge case tests
func TestComputeLoadingVectorZeroVariance(t *testing.T) {
	// Create data where t has zero variance
	X := mat.NewDense(3, 2, []float64{
		1, 0,
		0, 1,
		1, 1,
	})

	tVec := mat.NewVecDense(3, []float64{0, 0, 0}) // zero vector

	_, err := computeLoadingVector(X, tVec, 1e-8)
	if err == nil {
		t.Error("Expected error for zero variance, got nil")
	}
}

func TestComputeLoadingVectorWithMissingZeroVariance(t *testing.T) {
	X := mat.NewDense(3, 2, []float64{
		1, math.NaN(),
		0, 1,
		1, 1,
	})

	tVec := mat.NewVecDense(3, []float64{0, 0, 0})

	_, err := computeLoadingVectorWithMissing(X, tVec, 1e-8)
	if err == nil {
		t.Error("Expected error for zero variance, got nil")
	}
}
