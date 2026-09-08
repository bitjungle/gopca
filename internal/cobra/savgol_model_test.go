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

package cobra

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// Savitzky-Golay settings cross three hand-maintained representations between
// being chosen and being reapplied: the PCAConfig that fitted the model, the
// JSON written to disk, and the PreprocessingInfo parsed back out. Tests on
// either side of a hop compile against their own declaration of the "same"
// thing, so both stay green while the value is dropped in between.
//
// The failure is quiet in the worst way. A model whose filter was lost still
// transforms new data, still returns scores of the right shape and a plausible
// magnitude -- but projects unfiltered data onto loadings fitted on filtered
// data. Measured on corn, dropping the filter moved the scores by 167 times
// their own magnitude while every value still looked like a score.

func savGolTestData(rows, cols int) types.Matrix {
	data := make(types.Matrix, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			// A smooth peak on a sloping baseline, shifted per row: the shape
			// derivative preprocessing exists to isolate.
			t := float64(j)
			centre := 10.0 + float64(i)
			data[i][j] = 0.5 + 0.02*t + 3.0*math.Exp(-((t-centre)*(t-centre))/8.0)
		}
	}
	return data
}

// TestSavGolSurvivesTheModelFile follows the setting the whole way and compares
// what came out against what went in.
func TestSavGolSurvivesTheModelFile(t *testing.T) {
	config := types.PCAConfig{
		Components:      2,
		MeanCenter:      true,
		StandardScale:   true,
		Method:          "svd",
		SNV:             true,
		SavGolWindow:    9,
		SavGolPolyOrder: 2,
		SavGolDeriv:     1,
	}

	data := savGolTestData(12, 40)

	// What the engine actually does when fitting.
	fitted := core.NewPreprocessorWithScaleOnly(
		config.MeanCenter, config.StandardScale, config.RobustScale,
		config.ScaleOnly, config.SNV, config.VectorNorm)
	if err := core.ApplySavGolConfig(fitted, config); err != nil {
		t.Fatalf("configuring the fitting preprocessor: %v", err)
	}
	wantMatrix, err := fitted.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform: %v", err)
	}

	// The model file, built by the same writer the CLI uses, then serialised and
	// parsed back exactly as `pca transform` would receive it.
	info := types.PreprocessingInfo{
		MeanCenter:      config.MeanCenter,
		StandardScale:   config.StandardScale,
		SNV:             config.SNV,
		SavGolWindow:    config.SavGolWindow,
		SavGolPolyOrder: config.SavGolPolyOrder,
		SavGolDeriv:     config.SavGolDeriv,
		Parameters: types.PreprocessingParams{
			FeatureMeans:   fitted.GetMeans(),
			FeatureStdDevs: fitted.GetStdDevs(),
			FeatureMedians: fitted.GetMedians(),
			FeatureMADs:    fitted.GetMADs(),
		},
	}

	raw, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshalling the preprocessing block: %v", err)
	}
	var parsed types.PreprocessingInfo
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parsing the preprocessing block: %v", err)
	}

	// The hop most likely to lose the setting silently: a wrong or absent json
	// tag serialises to nothing and parses back as zero, which reads as "no
	// filter" rather than as an error.
	if parsed.SavGolWindow != config.SavGolWindow ||
		parsed.SavGolPolyOrder != config.SavGolPolyOrder ||
		parsed.SavGolDeriv != config.SavGolDeriv {
		t.Fatalf("the filter did not survive JSON: wrote window=%d order=%d deriv=%d, read back window=%d order=%d deriv=%d\nJSON: %s",
			config.SavGolWindow, config.SavGolPolyOrder, config.SavGolDeriv,
			parsed.SavGolWindow, parsed.SavGolPolyOrder, parsed.SavGolDeriv, raw)
	}

	// And the reconstruction the transform command actually performs.
	restored, err := preprocessorFromModel(parsed)
	if err != nil {
		t.Fatalf("preprocessorFromModel: %v", err)
	}
	gotMatrix, err := restored.Transform(data)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if len(gotMatrix) != len(wantMatrix) {
		t.Fatalf("restored preprocessor produced %d rows, want %d", len(gotMatrix), len(wantMatrix))
	}
	var worst float64
	for i := range wantMatrix {
		for j := range wantMatrix[i] {
			if d := math.Abs(gotMatrix[i][j] - wantMatrix[i][j]); d > worst {
				worst = d
			}
		}
	}
	if worst > 1e-12 {
		t.Errorf("data preprocessed through the model file differs from the fitted preprocessing by %.3g; "+
			"the transform path is not reproducing what the model was fitted with", worst)
	}
}

// TestSavGolAbsentFromModelMeansNoFilter checks the other direction: a model
// written before this feature existed, or one that legitimately used no filter,
// must not acquire one. Zero is the encoded absence, so this is the case that
// would break every older model if the enable condition were wrong.
func TestSavGolAbsentFromModelMeansNoFilter(t *testing.T) {
	data := savGolTestData(6, 20)

	plain := core.NewPreprocessorWithScaleOnly(true, false, false, false, false, false)
	want, err := plain.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform: %v", err)
	}

	restored, err := preprocessorFromModel(types.PreprocessingInfo{
		MeanCenter: true,
		Parameters: types.PreprocessingParams{
			FeatureMeans:   plain.GetMeans(),
			FeatureStdDevs: plain.GetStdDevs(),
			FeatureMedians: plain.GetMedians(),
			FeatureMADs:    plain.GetMADs(),
		},
	})
	if err != nil {
		t.Fatalf("preprocessorFromModel: %v", err)
	}
	if restored.SavitzkyGolayEnabled() {
		t.Fatal("a model with no Savitzky-Golay settings must not gain a filter")
	}
	got, err := restored.Transform(data)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > 1e-12 {
				t.Fatalf("row %d variable %d: %.12g vs %.12g", i, j, got[i][j], want[i][j])
			}
		}
	}
}

// TestSavGolReachesTheWrittenModel closes the remaining hop: the writer the CLI
// uses must copy the settings out of the config. Without this, the two tests
// above would both pass while `pca analyze` wrote a model that never mentioned
// the filter.
func TestSavGolReachesTheWrittenModel(t *testing.T) {
	config := types.PCAConfig{
		Components:      1,
		MeanCenter:      true,
		Method:          "svd",
		SavGolWindow:    7,
		SavGolPolyOrder: 3,
		SavGolDeriv:     2,
	}
	data := &csv.Data{
		Matrix:   savGolTestData(4, 12),
		Headers:  []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"},
		RowNames: []string{"r0", "r1", "r2", "r3"},
		Rows:     4,
		Columns:  12,
	}
	result := &types.PCAResult{
		Scores:               types.Matrix{{0.1}, {0.2}, {0.3}, {0.4}},
		Loadings:             types.Matrix{{0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}, {0.5}},
		ExplainedVar:         []float64{1},
		ExplainedVarRatio:    []float64{1},
		CumulativeVar:        []float64{1},
		ComponentLabels:      []string{"PC1"},
		ComponentsComputed:   1,
		Method:               "svd",
		PreprocessingApplied: true,
	}

	out := csv.ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, nil)
	if out == nil {
		t.Fatal("the writer returned nothing")
	}
	if out.Preprocessing.SavGolWindow != 7 ||
		out.Preprocessing.SavGolPolyOrder != 3 ||
		out.Preprocessing.SavGolDeriv != 2 {
		t.Errorf("the written model records window=%d order=%d deriv=%d, want 7/3/2 from the config",
			out.Preprocessing.SavGolWindow, out.Preprocessing.SavGolPolyOrder, out.Preprocessing.SavGolDeriv)
	}
}
