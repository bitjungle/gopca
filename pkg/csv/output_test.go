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

package csv

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// minimalPCAResult returns a minimal PCAResult sufficient for output conversion.
func minimalPCAResult() *types.PCAResult {
	return &types.PCAResult{
		Scores:             types.Matrix{{1, 2}, {3, 4}},
		Loadings:           types.Matrix{{0.5, 0.5}, {-0.5, 0.5}},
		ExplainedVar:       []float64{3.0, 1.0},
		ExplainedVarRatio:  []float64{0.75, 0.25},
		CumulativeVar:      []float64{0.75, 1.0},
		ComponentLabels:    []string{"PC1", "PC2"},
		ComponentsComputed: 2,
		Method:             "svd",
	}
}

// minimalData returns a minimal *Data for output conversion.
func minimalData() *Data {
	return &Data{
		Headers:  []string{"x", "y"},
		RowNames: []string{"obs1", "obs2"},
		Rows:     2,
		Columns:  2,
	}
}

func TestConvertToPCAOutputData_BasicSVD(t *testing.T) {
	result := minimalPCAResult()
	data := minimalData()
	config := types.PCAConfig{Method: "svd", MeanCenter: true}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, nil)

	if out == nil {
		t.Fatal("ConvertToPCAOutputData returned nil")
		return
	}
	if out.Schema == "" {
		t.Error("expected Schema to be set")
	}
	if out.Metadata.AnalysisID == "" {
		t.Error("expected AnalysisID to be set")
	}
	if out.Metadata.Config.Method != "svd" {
		t.Errorf("Method: got %q, want \"svd\"", out.Metadata.Config.Method)
	}
	if len(out.Model.Loadings) != 2 {
		t.Errorf("Loadings rows: got %d, want 2", len(out.Model.Loadings))
	}
	if len(out.Results.Samples.Names) != 2 {
		t.Errorf("Sample names: got %v", out.Results.Samples.Names)
	}
}

func TestConvertToPCAOutputData_WithMetadata(t *testing.T) {
	result := minimalPCAResult()
	data := minimalData()
	config := types.PCAConfig{Method: "svd"}
	meta := &ExportMetadata{
		InputFilename: "wine.csv",
		Description:   "Wine dataset PCA",
		Tags:          []string{"wine", "demo"},
	}

	out := ConvertToPCAOutputDataWithMetadata(result, data, nil, false, config, nil, nil, nil, meta)

	if out.Metadata.Description != "Wine dataset PCA" {
		t.Errorf("Description: got %q", out.Metadata.Description)
	}
	if len(out.Metadata.Tags) != 2 {
		t.Errorf("Tags: got %v", out.Metadata.Tags)
	}
	if out.Metadata.DataSource == nil {
		t.Error("expected DataSource to be set when InputFilename is provided")
	}
	if out.Metadata.DataSource.Filename != "wine.csv" {
		t.Errorf("DataSource.Filename: got %q", out.Metadata.DataSource.Filename)
	}
}

func TestConvertToPCAOutputData_KernelPCAMethodBranch(t *testing.T) {
	result := minimalPCAResult()
	result.Method = "kernel"
	data := minimalData()
	config := types.PCAConfig{
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.1,
	}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, nil)

	if out.Metadata.Config.KernelType != "rbf" {
		t.Errorf("KernelType: got %q, want \"rbf\"", out.Metadata.Config.KernelType)
	}
}

func TestConvertToPCAOutputData_PolyKernelBranch(t *testing.T) {
	result := minimalPCAResult()
	result.Method = "kernel"
	data := minimalData()
	config := types.PCAConfig{
		Method:       "kernel",
		KernelType:   "poly",
		KernelGamma:  0.5,
		KernelDegree: 3,
		KernelCoef0:  1.0,
	}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, nil)

	if out.Metadata.Config.KernelDegree != 3 {
		t.Errorf("KernelDegree: got %d, want 3", out.Metadata.Config.KernelDegree)
	}
}

func TestConvertToPCAOutputData_WithPreservedColumns(t *testing.T) {
	result := minimalPCAResult()
	data := minimalData()
	config := types.PCAConfig{Method: "svd"}
	catData := map[string][]string{"label": {"A", "B"}}
	targetData := map[string][]float64{"score": {1.0, 2.0}}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, catData, targetData)

	if out.PreservedColumns == nil {
		t.Fatal("expected PreservedColumns to be set")
		return
	}
	if _, ok := out.PreservedColumns.Categorical["label"]; !ok {
		t.Error("expected 'label' in PreservedColumns.Categorical")
	}
	if _, ok := out.PreservedColumns.NumericTarget["score"]; !ok {
		t.Error("expected 'score' in PreservedColumns.NumericTarget")
	}
}

func TestConvertToPCAOutputData_VariableLabelsFromResult(t *testing.T) {
	result := minimalPCAResult()
	result.VariableLabels = []string{"lag_x_1", "lag_x_2"}
	data := minimalData() // data.Headers = ["x", "y"]
	config := types.PCAConfig{Method: "svd"}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, nil)

	// When VariableLabels is set in result, it takes priority over data.Headers
	if len(out.Model.FeatureLabels) != 2 || out.Model.FeatureLabels[0] != "lag_x_1" {
		t.Errorf("FeatureLabels: got %v, want [lag_x_1 lag_x_2]", out.Model.FeatureLabels)
	}
}

func TestConvertToPCAOutputData_SchemaURL(t *testing.T) {
	out := ConvertToPCAOutputData(minimalPCAResult(), minimalData(), nil, false,
		types.PCAConfig{Method: "svd"}, nil, nil, nil)

	if !strings.Contains(out.Schema, "pca-output.schema.json") {
		t.Errorf("unexpected schema URL: %q", out.Schema)
	}
}

// TestPreservedTargetsSurviveAnUnmeasuredValue is a regression test for a defect
// that made JSON export impossible for a whole class of ordinary dataset.
//
// PreservedColumns.NumericTarget was map[string][]float64, and encoding/json
// refuses to marshal a NaN under any circumstances. A #target column with a gap
// in it — a calibration set where only some samples went for reference analysis,
// which is the normal case in chemometrics, not an edge case — therefore aborted
// the entire export with "json: unsupported value: NaN". Both `pca analyze -f
// json` and `pca regress -o` failed outright and wrote no model at all.
//
// The bug survived because every fixture in this file, and every dataset the
// export was tried on, had complete target columns. It was found by a
// differential test comparing the CLI against the desktop on testdata/bronir2,
// whose Dens#target is unmeasured on 414 of 855 rows.
//
// The assertion is on json.Marshal, not on the struct: the failure was in the
// encoding, so a check that only inspected the converted value would pass while
// the export stayed broken.
func TestPreservedTargetsSurviveAnUnmeasuredValue(t *testing.T) {
	result := minimalPCAResult()
	data := minimalData()
	config := types.PCAConfig{Method: "svd"}
	targetData := map[string][]float64{
		"reference": {1.5, math.NaN()},
	}

	out := ConvertToPCAOutputData(result, data, nil, false, config, nil, nil, targetData)

	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("the model cannot be encoded, which is the whole defect: %v", err)
	}

	// Null, not zero and not omitted. Zero would be a measurement nobody made,
	// and omitting the entry would misalign the column against the sample names.
	var decoded types.PCAOutputData
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decoding what was written: %v", err)
	}
	got := decoded.PreservedColumns.NumericTarget["reference"]
	if len(got) != 2 {
		t.Fatalf("the target column has %d values after the round trip, want 2", len(got))
	}
	if float64(got[0]) != 1.5 {
		t.Errorf("the measured value became %v, want 1.5", float64(got[0]))
	}
	if !got[1].IsNaN() {
		t.Errorf("the unmeasured value came back as %v; it must remain absent, "+
			"since a real measurement of that number is indistinguishable from a gap",
			float64(got[1]))
	}
	if !bytes.Contains(encoded, []byte(`"reference":[1.5,null]`)) {
		t.Errorf("the gap should be written as null; got:\n%s", encoded)
	}
}
