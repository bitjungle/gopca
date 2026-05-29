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
	"gonum.org/v1/gonum/mat"
)

// InitializeScoresAndLoadings creates new score and loading matrices for PCA algorithms
func InitializeScoresAndLoadings(nSamples, nFeatures, nComponents int) (*mat.Dense, *mat.Dense) {
	scores := mat.NewDense(nSamples, nComponents, nil)
	loadings := mat.NewDense(nFeatures, nComponents, nil)
	return scores, loadings
}

// CreateWorkingCopy creates a working copy of a matrix for deflation operations
func CreateWorkingCopy(original *mat.Dense) *mat.Dense {
	r, c := original.Dims()
	copy := mat.NewDense(r, c, nil)
	copy.Copy(original)
	return copy
}

// InitializeVector creates a new vector of specified size
func InitializeVector(size int) *mat.VecDense {
	return mat.NewVecDense(size, nil)
}

// InitializeMatrix creates a new matrix with specified dimensions
func InitializeMatrix(rows, cols int) *mat.Dense {
	return mat.NewDense(rows, cols, nil)
}

// InitializeSquareMatrix creates a new square matrix
func InitializeSquareMatrix(size int) *mat.Dense {
	return mat.NewDense(size, size, nil)
}

// CreateFloatSlice creates a new float64 slice of specified size
func CreateFloatSlice(size int) []float64 {
	return make([]float64, size)
}

// CreateFloat2DSlice creates a new 2D float64 slice with specified dimensions
func CreateFloat2DSlice(rows, cols int) [][]float64 {
	result := make([][]float64, rows)
	for i := range result {
		result[i] = make([]float64, cols)
	}
	return result
}

// CopyMatrixData creates a flat array copy of matrix data for gonum operations
func CopyMatrixData(source *mat.Dense) []float64 {
	r, c := source.Dims()
	data := make([]float64, r*c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			data[i*c+j] = source.At(i, j)
		}
	}
	return data
}

// ExtractColumn extracts a column from a matrix as a slice
func ExtractColumn(m *mat.Dense, col int) []float64 {
	r, _ := m.Dims()
	result := make([]float64, r)
	for i := 0; i < r; i++ {
		result[i] = m.At(i, col)
	}
	return result
}

// ExtractRow extracts a row from a matrix as a slice
func ExtractRow(m *mat.Dense, row int) []float64 {
	_, c := m.Dims()
	result := make([]float64, c)
	for j := 0; j < c; j++ {
		result[j] = m.At(row, j)
	}
	return result
}
