// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package utils

import (
	"gonum.org/v1/gonum/mat"
)

// MinInt returns the minimum of two integers
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SliceToMatrix converts a [][]float64 to a gonum Dense matrix
func SliceToMatrix(data [][]float64) *mat.Dense {
	if len(data) == 0 || len(data[0]) == 0 {
		return mat.NewDense(0, 0, nil)
	}

	rows, cols := len(data), len(data[0])
	flat := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			flat[i*cols+j] = data[i][j]
		}
	}
	return mat.NewDense(rows, cols, flat)
}

// MatrixToSlice converts a gonum Dense matrix to [][]float64
func MatrixToSlice(m *mat.Dense) [][]float64 {
	r, c := m.Dims()
	result := make([][]float64, r)
	for i := 0; i < r; i++ {
		result[i] = make([]float64, c)
		for j := 0; j < c; j++ {
			result[i][j] = m.At(i, j)
		}
	}
	return result
}
