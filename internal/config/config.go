// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package config

// CLIConfig holds configuration for the CLI application
type CLIConfig struct {
	// CSV parsing configuration
	CSV CSVConfig `json:"csv"`

	// Output configuration
	Output OutputConfig `json:"output"`

	// Analysis configuration
	Analysis AnalysisConfig `json:"analysis"`
}

// CSVConfig holds CSV parsing configuration
type CSVConfig struct {
	// Number of rows to sample for column type detection
	TypeDetectionSampleSize int `json:"type_detection_sample_size"`

	// Default null value strings
	DefaultNullValues []string `json:"default_null_values"`
}

// OutputConfig holds output file configuration
type OutputConfig struct {
	// Suffix for output files
	FileSuffix string `json:"file_suffix"`

	// Whether to create output directory if it doesn't exist
	CreateOutputDir bool `json:"create_output_dir"`
}

// AnalysisConfig holds analysis configuration
type AnalysisConfig struct {
	// Default number of components if not specified
	DefaultComponents int `json:"default_components"`

	// Whether to show preview of transformed data
	ShowPreview bool `json:"show_preview"`

	// Maximum number of rows to show in preview
	PreviewMaxRows int `json:"preview_max_rows"`
}

// AlgorithmConfig holds tunable parameters for PCA algorithms.
// These values affect the accuracy/performance trade-off of iterative methods.
type AlgorithmConfig struct {
	// NIPALS algorithm parameters
	NIPALS NIPALSConfig `json:"nipals"`

	// Kernel PCA algorithm parameters
	KernelPCA KernelPCAConfig `json:"kernel_pca"`
}

// NIPALSConfig holds parameters for the NIPALS (Nonlinear Iterative Partial Least Squares)
// power-iteration algorithm used for standard and missing-value PCA.
//
// Reference: Wold (1966), "Estimation of principal components and related models
// by iterative least squares", in Multivariate Analysis, pp. 391-420.
type NIPALSConfig struct {
	// Convergence tolerance: iteration stops when the change in the score vector
	// (Euclidean norm of t_new - t_old) falls below this threshold.
	// Smaller values give more precise components at the cost of more iterations.
	Tolerance float64 `json:"tolerance"`

	// Maximum number of power iterations per component.
	// For well-conditioned data, convergence typically occurs in <50 iterations.
	// Increase only if you see "did not converge" warnings on ill-conditioned data.
	MaxIterations int `json:"max_iterations"`
}

// KernelPCAConfig holds parameters for the Kernel PCA algorithm.
type KernelPCAConfig struct {
	// Minimum eigenvalue: eigenvalues below this threshold are clamped to it
	// to avoid division-by-zero when normalizing eigenvectors.
	// This is a numerical stability guard, not a truncation criterion.
	MinEigenvalue float64 `json:"min_eigenvalue"`
}

// DefaultAlgorithmConfig returns algorithm parameters suitable for typical
// well-conditioned datasets. Override these only when working with
// ill-conditioned data or when tighter numerical precision is required.
func DefaultAlgorithmConfig() AlgorithmConfig {
	return AlgorithmConfig{
		NIPALS: NIPALSConfig{
			Tolerance:     1e-8,
			MaxIterations: 1000,
		},
		KernelPCA: KernelPCAConfig{
			MinEigenvalue: 1e-10,
		},
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *CLIConfig {
	return &CLIConfig{
		CSV: CSVConfig{
			TypeDetectionSampleSize: 10,
			DefaultNullValues:       []string{"", "NA", "N/A", "null", "NULL", "NaN", "nan"},
		},
		Output: OutputConfig{
			FileSuffix:      "_pca",
			CreateOutputDir: true,
		},
		Analysis: AnalysisConfig{
			DefaultComponents: 0, // 0 means auto-detect
			ShowPreview:       true,
			PreviewMaxRows:    10,
		},
	}
}
