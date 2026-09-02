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

package main

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestPCRResultJSONMirrorsEngineType fails when the engine type gains a field the
// transport type does not.
//
// A PCR result is declared three times: the engine type in pkg/types, this
// transport struct, and TypeScript in the frontend. Each is a hand-maintained
// copy, so adding a field to the engine alone compiles, passes every test written
// against the engine, and never reaches the interface. Nothing but a check like
// this notices.
//
// Comparing JSON tag names rather than Go field names lets the two differ in
// naming while still being held to the same contract.
func TestPCRResultJSONMirrorsEngineType(t *testing.T) {
	engineTags := jsonTagsOf(reflect.TypeOf(types.PCRResult{}))
	transportTags := jsonTagsOf(reflect.TypeOf(PCRResultJSON{}))

	for tag := range engineTags {
		if _, present := transportTags[tag]; !present {
			t.Errorf("types.PCRResult has %q but PCRResultJSON does not: "+
				"a field added to the engine alone never reaches the interface", tag)
		}
	}
	for tag := range transportTags {
		if _, present := engineTags[tag]; !present {
			t.Errorf("PCRResultJSON has %q with no counterpart on types.PCRResult", tag)
		}
	}
}

// TestCVReportJSONMirrorsEngineType does the same for the validation report,
// which is where most of the numbers a user reads actually live.
func TestCVReportJSONMirrorsEngineType(t *testing.T) {
	engineTags := jsonTagsOf(reflect.TypeOf(types.CVReport{}))
	transportTags := jsonTagsOf(reflect.TypeOf(CVReportJSON{}))

	for tag := range engineTags {
		if _, present := transportTags[tag]; !present {
			t.Errorf("types.CVReport has %q but CVReportJSON does not", tag)
		}
	}
	for tag := range transportTags {
		if _, present := engineTags[tag]; !present {
			t.Errorf("CVReportJSON has %q with no counterpart on types.CVReport", tag)
		}
	}
}

// jsonTagsOf collects the JSON names a struct serialises under, skipping fields
// explicitly excluded with a "-" tag.
func jsonTagsOf(t reflect.Type) map[string]struct{} {
	tags := make(map[string]struct{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := tag
		for j := 0; j < len(tag); j++ {
			if tag[j] == ',' {
				name = tag[:j]
				break
			}
		}
		if name != "" {
			tags[name] = struct{}{}
		}
	}
	return tags
}

// TestPCRResultCrossesTheTransportBoundary follows real values the whole way from
// the engine type into marshalled JSON and back, rather than checking each side
// against its own declaration.
func TestPCRResultCrossesTheTransportBoundary(t *testing.T) {
	engineResult := &types.PCRResult{
		PCA: &types.PCAResult{
			Scores:          types.Matrix{{1, 2}, {3, 4}},
			Loadings:        types.Matrix{{0.7, 0.7}, {0.7, -0.7}},
			ComponentLabels: []string{"PC1", "PC2"},
			Method:          "svd",
		},
		Response:           "Moisture#target",
		Components:         2,
		ScoreCoefficients:  []float64{1.5, -0.25},
		Intercept:          10.5,
		Coefficients:       []float64{0.125, -0.5},
		InterceptOriginal:  9.75,
		OriginalScaleValid: true,
		ResponseMean:       10.2,
		Fitted:             []float64{10.1, 10.3},
		Residuals:          []float64{0.1, -0.1},
		RMSEC:              0.1,
		R2C:                0.93,
		LabelledRows:       []int{0, 1},
		ExcludedRows:       []int{2},
		CV: &types.CVReport{
			Scheme: "random", Design: "5-fold by row (shuffled)", Folds: 5, Seed: 42,
			NSamples: 2, Candidates: []int{0, 1, 2},
			RMSECV:     []float64{1.0, 0.5, 0.3},
			RMSECVMean: []float64{1.0, 0.5, 0.3},
			RMSECVSE:   []float64{0.1, 0.1, 0.1},
			Bias:       []float64{0, 0.01, 0.02},
			SEP:        []float64{1.0, 0.5, 0.3},
			MAE:        []float64{0.8, 0.4, 0.24},
			MAESE:      []float64{0.05, 0.05, 0.05},
			Q2:         []float64{0, 0.75, 0.91},
			Selected:   2, Rule: "one-se", SelectedByAlternateMetric: 1,
			OutOfFold: []float64{10.0, 10.4},
		},
	}

	encoded, err := json.Marshal(ConvertPCRResultToJSON(engineResult))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	checks := []struct {
		path string
		want interface{}
	}{
		{"response", "Moisture#target"},
		{"components", 2.0},
		{"intercept", 10.5},
		{"intercept_original", 9.75},
		{"original_scale_valid", true},
		{"response_mean", 10.2},
		{"rmsec", 0.1},
		{"r2c", 0.93},
	}
	for _, check := range checks {
		got, present := decoded[check.path]
		if !present {
			t.Errorf("%s did not survive the crossing", check.path)
			continue
		}
		if !reflect.DeepEqual(got, check.want) {
			t.Errorf("%s = %v, want %v", check.path, got, check.want)
		}
	}

	if decoded["pca"] == nil {
		t.Error("the decomposition did not survive; the interface needs it for the scree plot")
	}

	cv, ok := decoded["cv"].(map[string]interface{})
	if !ok {
		t.Fatal("the validation report did not survive")
	}
	if cv["selected"] != 2.0 {
		t.Errorf("cv.selected = %v, want 2", cv["selected"])
	}
	if cv["selected_by_alternate_metric"] != 1.0 {
		t.Errorf("cv.selected_by_alternate_metric = %v, want 1", cv["selected_by_alternate_metric"])
	}
	if mae, ok := cv["mae_se"].([]interface{}); !ok || len(mae) != 3 {
		t.Errorf("cv.mae_se did not survive as a three-element array: %v", cv["mae_se"])
	}
}

// TestPCRResultJSONSurvivesNonFiniteValues checks that a NaN does not take the
// whole response down.
//
// Missing responses produce NaN legitimately, and a plain float64 holding one is
// not representable in JSON: encoding/json fails the entire marshal, so a single
// unmeasured sample would turn into a blank screen rather than a gap in a plot.
func TestPCRResultJSONSurvivesNonFiniteValues(t *testing.T) {
	result := &types.PCRResult{
		PCA:        &types.PCAResult{Method: "svd"},
		Response:   "y",
		Components: 1,
		Fitted:     []float64{1.0, math.NaN()},
		Residuals:  []float64{0.0, math.NaN()},
		RMSEC:      math.NaN(),
		CV: &types.CVReport{
			Candidates: []int{0, 1},
			RMSECV:     []float64{1.0, math.NaN()},
			OutOfFold:  []float64{math.NaN(), 2.0},
		},
	}

	encoded, err := json.Marshal(ConvertPCRResultToJSON(result))
	if err != nil {
		t.Fatalf("a NaN broke the whole response: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("empty encoding")
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("the encoded result is not valid JSON: %v", err)
	}
}

// TestConvertPCRResultToJSONOmitsCoefficientsWithoutCollapsedForm checks that a
// model with no fixed original-scale form carries no coefficients, so the
// interface cannot draw numbers that do not describe it.
func TestConvertPCRResultToJSONOmitsCoefficientsWithoutCollapsedForm(t *testing.T) {
	result := &types.PCRResult{
		PCA:                &types.PCAResult{Method: "svd"},
		Response:           "y",
		Components:         1,
		Coefficients:       []float64{1, 2, 3}, // present on the engine type
		OriginalScaleValid: false,              // but not usable
	}

	converted := ConvertPCRResultToJSON(result)
	if converted.Coefficients != nil {
		t.Error("coefficients were carried across despite there being no collapsed form; " +
			"the interface would plot numbers that do not describe the model")
	}
	if converted.OriginalScaleValid {
		t.Error("OriginalScaleValid should have crossed as false")
	}
}

func TestConvertPCRResultToJSONHandlesNil(t *testing.T) {
	if ConvertPCRResultToJSON(nil) != nil {
		t.Error("expected nil for a nil result")
	}
}
