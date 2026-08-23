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
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func missingFixture() types.Matrix {
	data := syntheticCorrelatedData(60, 5)
	for i := range data {
		data[i][0] *= 40 // a column whose scale makes an omitted divisor obvious
	}
	data[7][2] = math.NaN()
	data[11][0] = math.NaN()
	data[23][4] = math.NaN()
	return data
}

func nativeConfig(scale string) types.PCAConfig {
	cfg := types.PCAConfig{
		Method: "nipals", Components: 3, MeanCenter: true,
		MissingStrategy: types.MissingNative,
	}
	switch scale {
	case "standard":
		cfg.StandardScale = true
	case "robust":
		cfg.RobustScale = true
	case "scaleonly":
		cfg.ScaleOnly = true
	}
	return cfg
}

// TestIssue783_ExportParametersAreFiniteWithMissingValues is the defect itself.
// Both model exporters previously called Preprocessor.FitTransform on the raw
// data. That returns no error on NaN input and produces NaN statistics, so the
// exported model either carried nonsense or could not be marshalled at all —
// Go's encoding/json refuses NaN.
func TestIssue783_ExportParametersAreFiniteWithMissingValues(t *testing.T) {
	data := missingFixture()

	for _, scale := range []string{"standard", "robust", "scaleonly"} {
		t.Run(scale, func(t *testing.T) {
			pre, _, err := FitPreprocessorForExport(data, nativeConfig(scale))
			if err != nil {
				t.Fatalf("fit failed: %v", err)
			}
			if pre == nil {
				t.Fatal("no preprocessor returned; the export would carry no parameters")
			}
			for name, vals := range map[string][]float64{
				"means": pre.GetMeans(), "stddevs": pre.GetStdDevs(),
				"medians": pre.GetMedians(), "mads": pre.GetMADs(),
			} {
				for i, v := range vals {
					if math.IsNaN(v) || math.IsInf(v, 0) {
						t.Errorf("%s[%d] = %v; a model carrying this cannot be marshalled to JSON", name, i, v)
					}
				}
			}
		})
	}
}

// TestIssue783_ExportParametersMatchWhatTheEngineApplied checks the numbers, not
// just their finiteness: a preprocessor fitted for export must reproduce the
// transformation the NIPALS engine performed, or a saved model will project new
// data differently from the fit.
func TestIssue783_ExportParametersMatchWhatTheEngineApplied(t *testing.T) {
	data := missingFixture()
	cfg := nativeConfig("standard")

	engine := NewPCAEngine().(*PCAImpl)
	if _, err := engine.Fit(data, cfg); err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	forExport, _, err := FitPreprocessorForExport(data, cfg)
	if err != nil {
		t.Fatalf("export fit failed: %v", err)
	}

	for name, pair := range map[string][2][]float64{
		"means":   {engine.preprocessor.GetMeans(), forExport.GetMeans()},
		"stddevs": {engine.preprocessor.GetStdDevs(), forExport.GetStdDevs()},
	} {
		got, want := pair[1], pair[0]
		if len(got) != len(want) {
			t.Fatalf("%s: length %d, want %d", name, len(got), len(want))
		}
		for i := range want {
			if math.Abs(got[i]-want[i]) > 1e-12 {
				t.Errorf("%s[%d] = %.12f, engine applied %.12f", name, i, got[i], want[i])
			}
		}
	}
}

// TestIssue783_DiagnosticsSuppressedForNativeMissing pins the second half: the
// export must not compute per-sample diagnostics against a matrix that still
// contains NaNs, which is why a nil matrix is returned. The engine makes the
// same choice for the same reason.
func TestIssue783_DiagnosticsSuppressedForNativeMissing(t *testing.T) {
	_, processed, err := FitPreprocessorForExport(missingFixture(), nativeConfig("standard"))
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	if processed != nil {
		t.Errorf("expected a nil matrix so diagnostics are skipped, got %d rows", len(processed))
	}
}

// TestIssue783_CompleteDataIsUnchanged guards the common path: for data without
// missing values the helper must behave exactly like the construct-and-fit it
// replaced, or every existing exported model changes.
func TestIssue783_CompleteDataIsUnchanged(t *testing.T) {
	data := syntheticCorrelatedData(40, 4)
	cfg := types.PCAConfig{Method: "svd", Components: 2, MeanCenter: true, StandardScale: true}

	old := NewPreprocessorWithScaleOnly(true, true, false, false, false, false)
	wantMatrix, err := old.FitTransform(data)
	if err != nil {
		t.Fatalf("reference fit failed: %v", err)
	}
	got, gotMatrix, err := FitPreprocessorForExport(data, cfg)
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	for i := range old.GetMeans() {
		if math.Abs(got.GetMeans()[i]-old.GetMeans()[i]) > 1e-12 ||
			math.Abs(got.GetStdDevs()[i]-old.GetStdDevs()[i]) > 1e-12 {
			t.Fatalf("column %d differs from the previous construct-and-fit behaviour", i)
		}
	}
	if gotMatrix == nil {
		t.Fatal("expected the preprocessed matrix so diagnostics still run on complete data")
	}
	for i := range wantMatrix {
		for j := range wantMatrix[i] {
			if math.Abs(gotMatrix[i][j]-wantMatrix[i][j]) > 1e-12 {
				t.Fatalf("preprocessed value [%d][%d] differs", i, j)
			}
		}
	}
}

// TestIssue783_NoPreprocessingRequested covers the pass-through case.
func TestIssue783_NoPreprocessingRequested(t *testing.T) {
	data := syntheticCorrelatedData(10, 3)
	pre, out, err := FitPreprocessorForExport(data, types.PCAConfig{Method: "svd", Components: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pre != nil {
		t.Error("expected no preprocessor when none was configured")
	}
	if len(out) != len(data) {
		t.Error("expected the data to pass through unchanged")
	}
}

// TestIssue783_PlainFitTransformStillPoisonsOnNaN documents why the helper
// exists. If Preprocessor ever gains NaN handling this will fail, and the
// special case can be revisited.
func TestIssue783_PlainFitTransformStillPoisonsOnNaN(t *testing.T) {
	pre := NewPreprocessorWithScaleOnly(true, true, false, false, false, false)
	if _, err := pre.FitTransform(missingFixture()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poisoned := false
	for _, v := range pre.GetMeans() {
		if math.IsNaN(v) {
			poisoned = true
		}
	}
	if !poisoned {
		t.Skip("Preprocessor now handles NaN; FitPreprocessorForExport's special case may be redundant")
	}
}
