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
	"math"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
)

// FitPreprocessorForExport fits a Preprocessor to data for the given
// configuration and returns it alongside the preprocessed matrix.
//
// It exists because the fitted preprocessing parameters are needed after the
// fact, by callers that no longer hold the engine which computed them: the CLI's
// model exporter and GoPCA Desktop's ExportPCAModel, which is a separate call
// from the one that ran the analysis. Both previously constructed a Preprocessor
// and called FitTransform on the raw data.
//
// That is wrong whenever the data contains missing values. Preprocessor has no
// NaN handling, so FitTransform returns no error and produces NaN means and
// standard deviations, which then either poison the exported model or fail JSON
// marshalling — Go refuses to encode NaN. This function routes the
// missing-value case through the same NaN-aware column statistics the NIPALS
// engine applies, so an exported model carries the parameters that were actually
// used and `pca transform` can reproduce the fit.
//
// The returned matrix is the preprocessed data, for callers that compute
// diagnostics from it. It is nil when native missing-value handling is in play:
// per-sample reconstruction diagnostics are ill-defined when entries have no
// ground truth, which is the same reason PCAImpl leaves PCAResult.PreprocessedData
// unset in that case.
//
// A nil preprocessor with a nil error means the configuration requested no
// preprocessing at all.
//
// Algorithm complexity: O(n*m) for centering and scaling, O(n*m*log n) when
// robust scaling requires per-column medians.
func FitPreprocessorForExport(data types.Matrix, config types.PCAConfig) (*Preprocessor, types.Matrix, error) {
	native := usesNativeMissingHandling(data, config)

	if !config.MeanCenter && !config.StandardScale && !config.RobustScale &&
		!config.ScaleOnly && !config.SNV && !config.VectorNorm {
		if native {
			// No preprocessing to record, but the data still carries the missing
			// entries, so the matrix must be withheld all the same. Handing it
			// back would let the caller compute diagnostics from NaN — and the
			// only gate downstream is a nil check (pkg/csv/output.go), so a
			// non-nil NaN-bearing matrix produces NaN metrics and an export that
			// cannot be marshalled.
			return nil, nil, nil
		}
		return nil, data, nil
	}

	if !native {
		preprocessor := NewPreprocessorWithScaleOnly(
			config.MeanCenter, config.StandardScale, config.RobustScale,
			config.ScaleOnly, config.SNV, config.VectorNorm)
		processed, err := preprocessor.FitTransform(data)
		if err != nil {
			return nil, nil, err
		}
		return preprocessor, processed, nil
	}
	preprocessor := NewPreprocessorWithScaleOnly(
		config.MeanCenter, config.StandardScale, config.RobustScale,
		config.ScaleOnly, false, false)

	// Native missing-value handling: derive the column statistics the way the
	// engine does, skipping absent entries, and hand them to the preprocessor
	// rather than letting it compute its own from NaN-bearing columns.
	// Means and standard deviations are computed for every column whatever the
	// scaling, mirroring the complete-data Fit — Preprocessor.Transform takes its
	// feature count from len(mean), so a preprocessor without means cannot
	// transform anything even when the branch in use would never subtract them.
	X := utils.MatrixToDense(data)
	means := computeColumnMeansWithMissing(X)
	stdDevs := computeColumnStdDevsWithMissing(X, means)
	var medians, mads []float64
	if config.RobustScale {
		medians = computeColumnMediansWithMissing(X)
		mads = computeColumnMADsWithMissing(X, medians)
	}
	if err := preprocessor.SetFittedParameters(means, stdDevs, medians, mads, nil, nil); err != nil {
		return nil, nil, err
	}
	// Diagnostics are skipped rather than computed against imputed values.
	return preprocessor, nil, nil
}

// usesNativeMissingHandling reports whether this configuration will take the
// NIPALS native missing-value path, which requires the data to actually contain
// missing values — the engine makes the same check before diverging.
func usesNativeMissingHandling(data types.Matrix, config types.PCAConfig) bool {
	if config.Method != "nipals" || config.MissingStrategy != types.MissingNative {
		return false
	}
	for _, row := range data {
		for _, v := range row {
			if math.IsNaN(v) {
				return true
			}
		}
	}
	return false
}
