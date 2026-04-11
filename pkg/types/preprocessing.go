// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package types

// PreprocessingType defines the type of preprocessing to apply
type PreprocessingType string

const (
	// PreprocessingTypeNone applies no preprocessing
	PreprocessingTypeNone PreprocessingType = "none"
	// PreprocessingTypeMeanCenter centers data by subtracting mean
	PreprocessingTypeMeanCenter PreprocessingType = "mean_center"
	// PreprocessingTypeStandardScaling standardizes to unit variance
	PreprocessingTypeStandardScaling PreprocessingType = "standard"
	// PreprocessingTypeRobustScaling uses median and MAD for robust scaling
	PreprocessingTypeRobustScaling PreprocessingType = "robust"
	// PreprocessingTypeSNV applies Standard Normal Variate
	PreprocessingTypeSNV PreprocessingType = "snv"
	// PreprocessingTypeVectorNorm applies L2 normalization
	PreprocessingTypeVectorNorm PreprocessingType = "vector_norm"
)

// PreprocessingConfig holds preprocessing configuration
type PreprocessingConfig struct {
	Method        PreprocessingType `json:"method"`
	MeanCenter    bool              `json:"mean_center"`
	StandardScale bool              `json:"standard_scale"`
	RobustScale   bool              `json:"robust_scale"`
	ScaleOnly     bool              `json:"scale_only"`
	SNV           bool              `json:"snv"`
	VectorNorm    bool              `json:"vector_norm"`
}

// MatrixImpl is a simple implementation of a Matrix-like interface
type MatrixImpl struct {
	Data [][]float64
	Rows int
	Cols int
}

// GetData returns the underlying data
func (m *MatrixImpl) GetData() [][]float64 {
	return m.Data
}

// GetRows returns the number of rows
func (m *MatrixImpl) GetRows() int {
	return m.Rows
}

// GetCols returns the number of columns
func (m *MatrixImpl) GetCols() int {
	return m.Cols
}
