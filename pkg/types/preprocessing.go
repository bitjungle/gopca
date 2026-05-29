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
