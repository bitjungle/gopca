// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

package core

import (
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
)

// RunPCAWithDiagnostics fits a PCA model for the configured method and, for
// linear PCA (svd/nipals), attaches per-sample diagnostic metrics (Q-residuals,
// Hotelling's T²) and their 95%/99% confidence limits.
//
// This is the single shared entry point used by both the CLI and the Desktop
// app. Previously each computed diagnostics independently by re-creating the
// preprocessor to recover the preprocessed matrix — logic that was duplicated
// and had drifted (native-missing NIPALS was mean-centered in one app but not
// the other), producing different metrics for identical input (#716). The
// engine now exposes the exact matrix it operated on via
// PCAResult.PreprocessedData, so diagnostics are computed against a single,
// authoritative reference in one place.
//
// Diagnostic computation is best-effort and non-fatal: a metrics error leaves
// the diagnostic fields unset but still returns the fitted result, matching the
// historical behaviour of both entry points. A fatal fitting error is returned.
func RunPCAWithDiagnostics(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	engine := NewPCAEngineForMethod(config.Method)
	result, err := engine.Fit(data, config)
	if err != nil {
		return nil, err
	}
	// Best-effort: a diagnostics failure must not fail the whole analysis.
	_ = AttachDiagnostics(result)
	return result, nil
}

// AttachDiagnostics computes per-sample diagnostic metrics (Q/T²) and their
// confidence limits for a fitted linear-PCA result and attaches them in place.
//
// Diagnostics are computed against result.PreprocessedData — the exact matrix
// the engine operated on, in the same space as the reconstruction
// (scores·loadingsᵀ). For kernel and temporal PCA that field is nil (their
// feature-space / lag-windowed reconstructions don't correspond to per-sample
// residuals in the original data space), so this function is a no-op for them.
//
// It returns the metrics error (nil on success or when skipped) so callers may
// log it; the historical behaviour is to treat such failures as non-fatal.
func AttachDiagnostics(result *types.PCAResult) error {
	if result == nil || result.PreprocessedData == nil {
		return nil // kernel/temporal PCA, or nothing to compute against
	}

	metrics, err := CalculateMetricsFromPCAResult(result, result.PreprocessedData)
	if err != nil {
		return err
	}
	result.Metrics = metrics

	// Confidence limits use the same scores/loadings as the metrics.
	scores := utils.MatrixToDense(result.Scores)
	loadings := utils.MatrixToDense(result.Loadings)
	calculator := NewPCAMetricsCalculator(scores, loadings, result.Means, result.StdDevs)

	result.T2Limit95, result.T2Limit99 = calculator.CalculateT2Limits()

	// Q-residual limits need the full eigenvalue spectrum (retained + discarded).
	if result.AllEigenvalues != nil && len(result.AllEigenvalues) > result.ComponentsComputed {
		result.QLimit95, result.QLimit99 = calculator.CalculateQLimits(result.AllEigenvalues, len(result.AllEigenvalues))
	}

	return nil
}
