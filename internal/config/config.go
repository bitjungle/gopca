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
