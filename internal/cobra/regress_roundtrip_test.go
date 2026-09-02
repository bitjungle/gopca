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
	"math/rand/v2"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// TestPCRModelSurvivesTheJSONRoundTrip follows a prediction the whole way from the
// engine, through the exported model file, and back out of pca transform.
//
// This is the check the layering demands. A PCR result is declared three times
// over: the engine type in pkg/types, the exported RegressionModel, and the JSON
// on disk. Each boundary is a hand-maintained copy, and a field dropped at any of
// them compiles cleanly and passes every test written on one side of the gap.
// Predictions that come back wrong, or not at all, are the symptom this catches.
func TestPCRModelSurvivesTheJSONRoundTrip(t *testing.T) {
	const n, p, latent = 60, 10, 3
	r := rand.New(rand.NewPCG(31, 17))

	loadings := make([][]float64, latent)
	for i := range loadings {
		loadings[i] = make([]float64, p)
		for j := range loadings[i] {
			loadings[i][j] = r.NormFloat64()
		}
	}
	weights := make([]float64, latent)
	for i := range weights {
		weights[i] = r.NormFloat64() * 3
	}

	matrix := make(types.Matrix, n)
	y := make([]float64, n)
	headers := make([]string, p)
	for j := range headers {
		headers[j] = string(rune('A' + j))
	}
	for i := 0; i < n; i++ {
		scores := make([]float64, latent)
		for l := range scores {
			scores[l] = r.NormFloat64()
		}
		matrix[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			var v float64
			for l := 0; l < latent; l++ {
				v += scores[l] * loadings[l][j]
			}
			matrix[i][j] = v + 0.05*r.NormFloat64()
		}
		for l := 0; l < latent; l++ {
			y[i] += scores[l] * weights[l]
		}
	}

	data := &pkgcsv.Data{
		Matrix: matrix, Headers: headers, Rows: n, Columns: p,
	}

	pcaConfig := types.PCAConfig{
		Components: 4, MeanCenter: true, StandardScale: true, Method: "svd",
	}
	config := types.PCRConfig{
		PCA:       pcaConfig,
		Response:  "y#target",
		Selection: types.SelectionConfig{Mode: "fixed", Fixed: 4, Metric: "rmse"},
	}

	engine := core.NewPCREngine()
	result, err := engine.Fit(matrix, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	impl, ok := engine.(*core.PCRImpl)
	if !ok {
		t.Fatal("expected the concrete engine type")
	}

	exported := pkgcsv.ConvertToPCROutputData(result, data, false, pcaConfig,
		impl.Preprocessor(), nil, map[string][]float64{"y#target": y}, nil)
	if exported.Regression == nil {
		t.Fatal("the exported model carries no regression block")
	}

	encoded, err := json.Marshal(exported)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored types.PCAOutputData
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.Regression == nil {
		t.Fatal("the regression block did not survive the round trip")
	}

	// Every field the prediction path depends on must have made the journey.
	got, want := restored.Regression, exported.Regression
	if got.Response != want.Response {
		t.Errorf("response = %q, want %q", got.Response, want.Response)
	}
	if got.Components != want.Components {
		t.Errorf("components = %d, want %d", got.Components, want.Components)
	}
	if len(got.ScoreCoefficients) != len(want.ScoreCoefficients) {
		t.Fatalf("score coefficients: %d survived of %d",
			len(got.ScoreCoefficients), len(want.ScoreCoefficients))
	}
	if len(got.Coefficients) != p {
		t.Errorf("original-scale coefficients: %d survived of %d", len(got.Coefficients), p)
	}
	if !got.OriginalScaleValid {
		t.Error("original_scale_valid did not survive as true")
	}
	if got.Validation == nil && want.Validation != nil {
		t.Error("the validation report did not survive")
	}

	// The decisive check: predicting from the restored model must reproduce the
	// fitted values the engine reported.
	predictions, err := predictFromModel(restored.Regression, result.PCA.Scores)
	if err != nil {
		t.Fatalf("predictFromModel: %v", err)
	}
	if len(predictions) != n {
		t.Fatalf("got %d predictions, want %d", len(predictions), n)
	}
	for i, row := range result.LabelledRows {
		if diff := math.Abs(predictions[row] - result.Fitted[i]); diff > 1e-9 {
			t.Errorf("row %d: the round trip predicts %.12g but the engine fitted %.12g",
				row, predictions[row], result.Fitted[i])
		}
	}

	// The collapsed original-scale form must agree with the score-space form, since
	// consumers may use either.
	for i := 0; i < n; i++ {
		collapsed := restored.Regression.InterceptOriginal
		for j := 0; j < p; j++ {
			collapsed += matrix[i][j] * restored.Regression.Coefficients[j]
		}
		if diff := math.Abs(collapsed - predictions[i]); diff > 1e-8*(1+math.Abs(collapsed)) {
			t.Fatalf("row %d: collapsed form gives %.12g, score-space form %.12g",
				i, collapsed, predictions[i])
		}
	}
}

// TestPredictFromModelRefusesTooFewComponents checks that a model asking for more
// components than the data can supply fails rather than reading past the end.
func TestPredictFromModelRefusesTooFewComponents(t *testing.T) {
	model := &types.RegressionModel{
		ScoreCoefficients: []float64{1, 2, 3},
		Intercept:         0.5,
	}
	if _, err := predictFromModel(model, types.Matrix{{1, 2}}); err == nil {
		t.Error("expected a model needing 3 components to refuse 2-column scores")
	}
}

// TestPredictFromModelInterceptOnly covers the k=0 baseline surviving export.
func TestPredictFromModelInterceptOnly(t *testing.T) {
	model := &types.RegressionModel{ScoreCoefficients: nil, Intercept: 7.25}
	predictions, err := predictFromModel(model, types.Matrix{{1, 2}, {3, 4}})
	if err != nil {
		t.Fatalf("predictFromModel: %v", err)
	}
	for i, v := range predictions {
		if v != 7.25 {
			t.Errorf("row %d predicted %v, want the intercept 7.25", i, v)
		}
	}
}
