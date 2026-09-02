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

package core_test

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// loadCalibrationSet reads a dataset and its numeric target, dropping rows with
// incomplete predictors as the command line will.
//
// columnStride subsamples the predictors, keeping every nth one. Spectra are
// heavily oversampled, so a stride preserves the structure these tests assert
// while cutting the cost of the decomposition by more than an order of magnitude.
// That matters because the pre-commit hook budgets five minutes for the entire
// suite under the race detector, and a full 855 by 1001 sweep alone exceeds it.
func loadCalibrationSet(t *testing.T, relative, response string, columnStride int) (types.Matrix, []float64) {
	t.Helper()

	opts := pkgcsv.DefaultOptions()
	opts.ParseMode = pkgcsv.ParseMixedWithTargets
	parsed, err := pkgcsv.NewReader(opts).ReadFile(
		filepath.Join("..", "..", "testdata", relative))
	if err != nil {
		t.Skipf("dataset %s unavailable: %v", relative, err)
	}

	y := parsed.NumericTargetColumns[response]
	if y == nil {
		t.Fatalf("%s has no numeric target column %q", relative, response)
	}

	if columnStride < 1 {
		columnStride = 1
	}

	data := make(types.Matrix, 0, len(parsed.Matrix))
	kept := make([]float64, 0, len(y))
	for i := range parsed.Matrix {
		complete := true
		for _, v := range parsed.Matrix[i] {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				complete = false
				break
			}
		}
		if !complete {
			continue
		}
		row := make([]float64, 0, len(parsed.Matrix[i])/columnStride+1)
		for j := 0; j < len(parsed.Matrix[i]); j += columnStride {
			row = append(row, parsed.Matrix[i][j])
		}
		data = append(data, row)
		kept = append(kept, y[i])
	}
	return data, kept
}

// TestPCROnPartiallyLabelledCalibrationSet exercises the semi-supervised path on
// real data rather than a construction.
//
// bronir2 is a polymer near-infrared set in which no row carries all four
// responses and each is measured on a different subset. Predicting density means
// regressing on roughly half the rows while the decomposition can still use every
// spectrum, which is the situation the design exists for. A change that quietly
// restricted the decomposition to labelled rows would still produce a plausible
// model, so the row counts are asserted directly.
func TestPCROnPartiallyLabelledCalibrationSet(t *testing.T) {
	data, y := loadCalibrationSet(t, "bronir2/bronir2.csv", "Dens#target", 8)
	if len(data) == 0 {
		t.Skip("no usable rows")
	}

	unobserved := 0
	for _, v := range y {
		if math.IsNaN(v) {
			unobserved++
		}
	}
	if unobserved == 0 {
		t.Fatal("this test needs a response with unobserved values; the dataset may have changed")
	}

	config := types.PCRConfig{
		PCA: types.PCAConfig{
			Components: 8, MeanCenter: true, StandardScale: true, Method: "svd",
		},
		Response: "Dens#target",
		Selection: types.SelectionConfig{
			Mode: "cv", Metric: "rmse", Rule: types.SelectOneSE,
			CV: types.CVConfig{Scheme: types.CVRandom, Folds: 4, Seed: 1},
		},
	}

	result, err := core.NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	if len(result.ExcludedRows) != unobserved {
		t.Errorf("excluded %d rows, want the %d without an observed response",
			len(result.ExcludedRows), unobserved)
	}
	if len(result.LabelledRows) != len(data)-unobserved {
		t.Errorf("regressed on %d rows, want %d", len(result.LabelledRows), len(data)-unobserved)
	}

	// The decisive assertion: the decomposition used every row, not just the
	// labelled ones. Restricting it here would discard hundreds of usable spectra.
	if got := len(result.PCA.Scores); got != len(data) {
		t.Errorf("the decomposition used %d rows, want all %d: rows without an "+
			"observed response still carry predictor structure", got, len(data))
	}

	if result.CV == nil {
		t.Fatal("expected a cross-validation report")
	}
	if result.CV.NSamples != len(result.LabelledRows) {
		t.Errorf("the report covers %d samples, want the %d labelled rows",
			result.CV.NSamples, len(result.LabelledRows))
	}
	if result.Components < 1 {
		t.Errorf("selected %d components on a dataset with real signal", result.Components)
	}

	selected := result.CV.Selected
	baseline := result.CV.RMSECV[0]
	if result.CV.RMSECV[selected] >= baseline {
		t.Errorf("the selected model (RMSECV %.4f) does no better than the "+
			"intercept-only baseline (%.4f)", result.CV.RMSECV[selected], baseline)
	}

	// Every labelled row must receive a held-out prediction.
	for i, v := range result.CV.OutOfFold {
		if math.IsNaN(v) {
			t.Fatalf("labelled row %d received no held-out prediction", i)
		}
	}
}

// TestPCRSelectionRulesDisagreeAsExpected records how the rules behave on a real
// calibration curve, so that a change in any of them is visible rather than
// absorbed silently.
//
// The corn moisture curve falls, plateaus, then collapses at seven components.
// That shape separates the rules: the minimum chases the last small gain, the
// one-standard-error rule stops earlier, and Wold's greedy criterion stops on the
// first plateau and badly underfits. The last is a documented weakness of the
// published rule, not a defect here; see SelectComponents.
func TestPCRSelectionRulesDisagreeAsExpected(t *testing.T) {
	data, y := loadCalibrationSet(t, "corn/corn.csv", "Moisture#target", 1)
	if len(data) == 0 {
		t.Skip("no usable rows")
	}

	fit := func(rule string) *types.PCRResult {
		t.Helper()
		config := types.PCRConfig{
			PCA: types.PCAConfig{
				Components: 20, MeanCenter: true, StandardScale: true, Method: "svd",
			},
			Response: "Moisture#target",
			Selection: types.SelectionConfig{
				Mode: "cv", Metric: "rmse", Rule: rule, WoldR: 0.95,
				CV: types.CVConfig{Scheme: types.CVContiguous, Folds: 5},
			},
		}
		result, err := core.NewPCREngine().Fit(data, y, config)
		if err != nil {
			t.Fatalf("rule %s: %v", rule, err)
		}
		return result
	}

	minimum := fit(types.SelectMin)
	oneSE := fit(types.SelectOneSE)
	wold := fit(types.SelectWold)

	if oneSE.Components > minimum.Components {
		t.Errorf("the one-standard-error rule chose %d components, more than the "+
			"minimum's %d; it must never be less parsimonious",
			oneSE.Components, minimum.Components)
	}
	if wold.Components > oneSE.Components {
		t.Errorf("Wold's rule chose %d components, more than the one-standard-error "+
			"rule's %d on a curve with an early plateau", wold.Components, oneSE.Components)
	}

	// The chosen model must beat the intercept-only baseline by a wide margin:
	// near-infrared moisture is close to deterministic.
	i := minimum.CV.Selected
	if q2 := minimum.CV.Q2[i]; q2 < 0.9 {
		t.Errorf("cross-validated Q2 = %.4f at %d components; near-infrared moisture "+
			"should be predicted far better than this", q2, minimum.Components)
	}
	if minimum.RMSEC > minimum.CV.RMSECV[i] {
		t.Errorf("RMSEC %.5f exceeds RMSECV %.5f; the training fit cannot be worse "+
			"than the held-out estimate", minimum.RMSEC, minimum.CV.RMSECV[i])
	}
}
