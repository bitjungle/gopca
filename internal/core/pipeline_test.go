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

// TestIssue716_NativeMissingMetricsUseCenteredSpace locks in the divergence fix:
// for NIPALS with native missing values + mean-centering, diagnostics must be
// computed in the mean-centered space (the reconstruction space), not against
// the raw data. Previously the Desktop app centered while the CLI used raw,
// producing different metrics for identical input.
func TestIssue716_NativeMissingMetricsUseCenteredSpace(t *testing.T) {
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
	if result.PreprocessedData == nil {
		t.Fatal("expected PreprocessedData for native-missing NIPALS")
	}

	// The exposed matrix must be mean-centered (column means ≈ 0, ignoring NaN),
	// not the raw data.
	nRows := len(result.PreprocessedData)
	nCols := len(result.PreprocessedData[0])
	for j := 0; j < nCols; j++ {
		sum, count := 0.0, 0
		for i := 0; i < nRows; i++ {
			v := result.PreprocessedData[i][j]
			if !math.IsNaN(v) {
				sum += v
				count++
			}
		}
		mean := sum / float64(count)
		if math.Abs(mean) > 1e-9 {
			t.Errorf("column %d not centered: mean=%g (expected ~0)", j, mean)
		}
	}

	// And it must differ from the raw data (a raw column mean is clearly non-zero).
	if math.Abs(result.PreprocessedData[0][0]-data[0][0]) < 1e-12 {
		t.Error("PreprocessedData appears to equal raw data; expected mean-centering")
	}
}

// TestRunPCAWithDiagnostics_SkipsNonlinearMethods verifies kernel/temporal PCA
// leave PreprocessedData nil so per-sample reconstruction metrics are skipped.
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
