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

package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CSV.TypeDetectionSampleSize != 10 {
		t.Errorf("TypeDetectionSampleSize = %d, want 10", cfg.CSV.TypeDetectionSampleSize)
	}
	if len(cfg.CSV.DefaultNullValues) == 0 {
		t.Error("DefaultNullValues should not be empty")
	}
	if cfg.Output.FileSuffix != "_pca" {
		t.Errorf("FileSuffix = %q, want %q", cfg.Output.FileSuffix, "_pca")
	}
	if !cfg.Output.CreateOutputDir {
		t.Error("CreateOutputDir should default to true")
	}
	if cfg.Analysis.DefaultComponents != 0 {
		t.Errorf("DefaultComponents = %d, want 0 (auto-detect)", cfg.Analysis.DefaultComponents)
	}
}

func TestDefaultAlgorithmConfig(t *testing.T) {
	cfg := DefaultAlgorithmConfig()
	if cfg.NIPALS.Tolerance != 1e-8 {
		t.Errorf("NIPALS.Tolerance = %g, want 1e-8", cfg.NIPALS.Tolerance)
	}
	if cfg.NIPALS.MaxIterations != 1000 {
		t.Errorf("NIPALS.MaxIterations = %d, want 1000", cfg.NIPALS.MaxIterations)
	}
	if cfg.KernelPCA.MinEigenvalue != 1e-10 {
		t.Errorf("KernelPCA.MinEigenvalue = %g, want 1e-10", cfg.KernelPCA.MinEigenvalue)
	}
}

func TestDefaultGUIConfig(t *testing.T) {
	cfg := DefaultGUIConfig()
	if cfg.Visualization.DefaultConfidenceLevel != 0.95 {
		t.Errorf("DefaultConfidenceLevel = %g, want 0.95", cfg.Visualization.DefaultConfidenceLevel)
	}
	if cfg.Visualization.BiplotMaxVariables != 100 {
		t.Errorf("BiplotMaxVariables = %d, want 100", cfg.Visualization.BiplotMaxVariables)
	}
	if cfg.UI.DataPreviewMaxRows != 10 {
		t.Errorf("DataPreviewMaxRows = %d, want 10", cfg.UI.DataPreviewMaxRows)
	}
}
