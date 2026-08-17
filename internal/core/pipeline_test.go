// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// wellConditionedData returns a small, well-conditioned dataset for diagnostics.
func wellConditionedData() types.Matrix {
	return types.Matrix{
		{2.5, 2.4, 1.1},
		{0.5, 0.7, 3.2},
		{2.2, 2.9, 1.5},
		{1.9, 2.2, 1.8},
		{3.1, 3.0, 0.9},
		{2.3, 2.7, 1.3},
		{2.0, 1.6, 2.1},
		{1.0, 1.1, 2.9},
	}
}

// TestRunPCAWithDiagnostics_AttachesLinearDiagnostics verifies the shared
// pipeline attaches per-sample metrics and confidence limits for linear PCA,
// and exposes the preprocessed matrix used for them.
func TestRunPCAWithDiagnostics_AttachesLinearDiagnostics(t *testing.T) {
	data := wellConditionedData()
	config := types.PCAConfig{Method: "svd", Components: 2, MeanCenter: true}

	result, err := RunPCAWithDiagnostics(data, config)
	if err != nil {
		t.Fatalf("RunPCAWithDiagnostics failed: %v", err)
	}
	if result.PreprocessedData == nil {
		t.Fatal("expected PreprocessedData to be populated for linear PCA")
	}
	if len(result.Metrics) != len(data) {
		t.Fatalf("expected %d per-sample metrics, got %d", len(data), len(result.Metrics))
	}
	if result.T2Limit95 <= 0 || result.T2Limit99 <= 0 {
		t.Errorf("expected positive T² limits, got T2Limit95=%g T2Limit99=%g", result.T2Limit95, result.T2Limit99)
	}
}

// TestIssue716_PreprocessedDataMatchesEngineSpace verifies the engine exposes
// exactly the mean-centered/scaled matrix it operated on — the same matrix a
// caller would previously re-derive by re-creating the preprocessor.
func TestIssue716_PreprocessedDataMatchesEngineSpace(t *testing.T) {
	data := wellConditionedData()
	config := types.PCAConfig{Method: "svd", Components: 2, MeanCenter: true, StandardScale: true}

	result, err := RunPCAWithDiagnostics(data, config)
	if err != nil {
		t.Fatalf("RunPCAWithDiagnostics failed: %v", err)
	}

	// Independently reproduce the preprocessing the old callers used.
	want, err := NewPreprocessorWithScaleOnly(true, true, false, false, false, false).FitTransform(data)
	if err != nil {
		t.Fatalf("reference preprocessing failed: %v", err)
	}
	assertMatrixClose(t, result.PreprocessedData, want, 1e-9, "PreprocessedData vs reference preprocessing")
}

// TestIssue716_NativeMissingSkipsDiagnostics verifies that NIPALS with genuine
// missing values skips per-sample diagnostics (PreprocessedData nil, Metrics
// empty). Reconstruction diagnostics (Q/T²) are ill-defined when entries are
// missing, and the generic Preprocessor cannot reproduce NIPALS' NaN-aware
// centering (it would yield NaN column means, corrupting the whole matrix). So
// rather than centering with the generic preprocessor, diagnostics are skipped —
// and, crucially, PreprocessedData must never contain NaN-filled columns.
func TestIssue716_NativeMissingSkipsDiagnostics(t *testing.T) {
	data := wellConditionedData()
	data[3][1] = math.NaN() // introduce a genuine missing value

	config := types.PCAConfig{
		Method:          "nipals",
		Components:      2,
		MeanCenter:      true,
		MissingStrategy: types.MissingNative,
	}

	result, err := RunPCAWithDiagnostics(data, config)
	if err != nil {
		t.Fatalf("RunPCAWithDiagnostics failed: %v", err)
	}
	if result.PreprocessedData != nil {
		t.Errorf("expected PreprocessedData nil for native-missing NIPALS (diagnostics skipped), got %d rows", len(result.PreprocessedData))
	}
	if len(result.Metrics) != 0 {
		t.Errorf("expected no diagnostic metrics for native-missing NIPALS, got %d", len(result.Metrics))
	}
}

// TestRunPCAWithDiagnostics_SkipsNonlinearMethods verifies kernel and temporal
// PCA leave PreprocessedData nil so per-sample reconstruction metrics are
// skipped (their reconstructions don't correspond to residuals in the original
// data space). Temporal in particular must not leave callers with a nil matrix
// that later panics when metrics are requested.
func TestRunPCAWithDiagnostics_SkipsKernelDiagnostics(t *testing.T) {
	data := wellConditionedData()
	config := types.PCAConfig{Method: "kernel", Components: 2, MeanCenter: true, KernelType: "rbf", KernelGamma: 0.5}

	result, err := RunPCAWithDiagnostics(data, config)
	if err != nil {
		t.Fatalf("RunPCAWithDiagnostics (kernel) failed: %v", err)
	}
	if result.PreprocessedData != nil {
		t.Error("expected kernel PCA to leave PreprocessedData nil (diagnostics skipped)")
	}
	if len(result.Metrics) != 0 {
		t.Errorf("expected no diagnostic metrics for kernel PCA, got %d", len(result.Metrics))
	}
}

// TestRunPCAWithDiagnostics_SkipsTemporalDiagnostics guards the CLI regression
// from #716: temporal PCA reduces the sample count, so per-sample diagnostics
// against the original data don't apply. PreprocessedData must be nil (and
// Metrics empty) so the table/JSON output paths never attempt to recompute
// metrics against a nil matrix.
func TestRunPCAWithDiagnostics_SkipsTemporalDiagnostics(t *testing.T) {
	data := wellConditionedData()
	config := types.PCAConfig{Method: "temporal", Components: 2, MeanCenter: true, TemporalLags: 2}

	result, err := RunPCAWithDiagnostics(data, config)
	if err != nil {
		t.Fatalf("RunPCAWithDiagnostics (temporal) failed: %v", err)
	}
	if result.PreprocessedData != nil {
		t.Error("expected temporal PCA to leave PreprocessedData nil (diagnostics skipped)")
	}
	if len(result.Metrics) != 0 {
		t.Errorf("expected no diagnostic metrics for temporal PCA, got %d", len(result.Metrics))
	}
}

// assertMatrixClose fails if two matrices differ by more than tol elementwise.
func assertMatrixClose(t *testing.T, got, want types.Matrix, tol float64, msg string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: row count %d != %d", msg, len(got), len(want))
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("%s: row %d col count %d != %d", msg, i, len(got[i]), len(want[i]))
		}
		for j := range got[i] {
			if math.Abs(got[i][j]-want[i][j]) > tol {
				t.Errorf("%s: [%d][%d] = %g, want %g (Δ=%g)", msg, i, j, got[i][j], want[i][j], math.Abs(got[i][j]-want[i][j]))
			}
		}
	}
}
