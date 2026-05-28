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

package utils

import (
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// MatrixToDense converts a types.Matrix to a gonum Dense matrix
func MatrixToDense(m types.Matrix) *mat.Dense {
	if len(m) == 0 || len(m[0]) == 0 {
		return mat.NewDense(0, 0, nil)
	}

	rows, cols := len(m), len(m[0])
	data := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			data[i*cols+j] = m[i][j]
		}
	}
	return mat.NewDense(rows, cols, data)
}

// DenseToMatrix converts a gonum Dense matrix to types.Matrix
func DenseToMatrix(d *mat.Dense) types.Matrix {
	r, c := d.Dims()
	m := make(types.Matrix, r)
	for i := 0; i < r; i++ {
		m[i] = make([]float64, c)
		for j := 0; j < c; j++ {
			m[i][j] = d.At(i, j)
		}
	}
	return m
}
