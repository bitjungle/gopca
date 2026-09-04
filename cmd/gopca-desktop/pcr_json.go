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

import "github.com/bitjungle/gopca/pkg/types"

// PCRResultJSON is a JSON-safe view of types.PCRResult.
//
// This is a hand-maintained mirror of the engine type, and the third place a PCR
// result is declared: once in pkg/types, once here, once in the frontend's
// TypeScript. A field added to the engine alone compiles, passes its tests, and
// never reaches the interface, so all three have to move together. See
// TestPCRResultCrossesTheTransportBoundary.
//
// The float fields are types.JSONFloat64 because a plain float64 carrying NaN or
// an infinity is not representable in JSON and would fail the marshal, taking the
// whole response with it. Missing responses and undefined metrics both produce
// NaN legitimately.
type PCRResultJSON struct {
	// PCA is the decomposition the regression was fitted on, so the interface can
	// show a scree plot beside the selection curve without a second request.
	PCA *PCAResultJSON `json:"pca"`

	Response   string `json:"response"`
	Components int    `json:"components"`

	ScoreCoefficients []types.JSONFloat64 `json:"score_coefficients"`
	Intercept         types.JSONFloat64   `json:"intercept"`

	// Coefficients and InterceptOriginal are the collapsed original-scale form.
	// OriginalScaleValid says whether it exists: row-wise preprocessing scales
	// each sample by a statistic of that same sample, so no fixed coefficient
	// vector reproduces it. The interface must hide the coefficient plot rather
	// than draw meaningless numbers.
	Coefficients       []types.JSONFloat64 `json:"coefficients,omitempty"`
	InterceptOriginal  types.JSONFloat64   `json:"intercept_original"`
	OriginalScaleValid bool                `json:"original_scale_valid"`

	ResponseMean types.JSONFloat64   `json:"response_mean"`
	Fitted       []types.JSONFloat64 `json:"fitted"`
	Residuals    []types.JSONFloat64 `json:"residuals"`

	// RMSEC describes the training fit and is not an estimate of performance.
	// The interface labels it as such; see the Error panel.
	RMSEC types.JSONFloat64 `json:"rmsec"`
	R2C   types.JSONFloat64 `json:"r2c"`

	CV *CVReportJSON `json:"cv,omitempty"`

	// LabelledRows indexes Fitted, Residuals and CV.OutOfFold against the original
	// data rows, so the interface can label points with the right sample names.
	LabelledRows []int `json:"labelled_rows,omitempty"`
	ExcludedRows []int `json:"excluded_rows,omitempty"`
}

// CVReportJSON is a JSON-safe view of types.CVReport.
type CVReportJSON struct {
	Scheme   string `json:"scheme"`
	Design   string `json:"design"`
	Folds    int    `json:"folds"`
	Repeats  int    `json:"repeats"`
	GroupBy  string `json:"group_by,omitempty"`
	Seed     int64  `json:"seed"`
	NSamples int    `json:"n_samples"`

	Candidates []int               `json:"candidates"`
	RMSECV     []types.JSONFloat64 `json:"rmsecv"`
	RMSECVMean []types.JSONFloat64 `json:"rmsecv_mean"`
	RMSECVSE   []types.JSONFloat64 `json:"rmsecv_se"`
	Bias       []types.JSONFloat64 `json:"bias"`
	SEP        []types.JSONFloat64 `json:"sep"`
	MAE        []types.JSONFloat64 `json:"mae"`
	MAESE      []types.JSONFloat64 `json:"mae_se"`
	Q2         []types.JSONFloat64 `json:"q2"`

	Selected                  int    `json:"selected"`
	Rule                      string `json:"rule"`
	Metric                    string `json:"metric"`
	SelectedByAlternateMetric int    `json:"selected_by_alternate_metric"`
	LowestError               int    `json:"lowest_error"`
	CurveStillFalling         bool   `json:"curve_still_falling"`

	OutOfFold []types.JSONFloat64 `json:"out_of_fold_predictions"`
}

// ConvertPCRResultToJSON converts an engine result into its transport form.
func ConvertPCRResultToJSON(result *types.PCRResult) *PCRResultJSON {
	if result == nil {
		return nil
	}

	out := &PCRResultJSON{
		PCA:                ConvertPCAResultToJSON(result.PCA),
		Response:           result.Response,
		Components:         result.Components,
		ScoreCoefficients:  toJSONFloats(result.ScoreCoefficients),
		Intercept:          types.JSONFloat64(result.Intercept),
		InterceptOriginal:  types.JSONFloat64(result.InterceptOriginal),
		OriginalScaleValid: result.OriginalScaleValid,
		ResponseMean:       types.JSONFloat64(result.ResponseMean),
		Fitted:             toJSONFloats(result.Fitted),
		Residuals:          toJSONFloats(result.Residuals),
		RMSEC:              types.JSONFloat64(result.RMSEC),
		R2C:                types.JSONFloat64(result.R2C),
		LabelledRows:       result.LabelledRows,
		ExcludedRows:       result.ExcludedRows,
	}

	// Only carried when the collapsed form exists, so the interface can test for
	// its presence as well as reading the flag.
	if result.OriginalScaleValid {
		out.Coefficients = toJSONFloats(result.Coefficients)
	}

	if result.CV != nil {
		out.CV = &CVReportJSON{
			Scheme:                    result.CV.Scheme,
			Design:                    result.CV.Design,
			Folds:                     result.CV.Folds,
			Repeats:                   result.CV.Repeats,
			GroupBy:                   result.CV.GroupBy,
			Seed:                      result.CV.Seed,
			NSamples:                  result.CV.NSamples,
			Candidates:                result.CV.Candidates,
			RMSECV:                    toJSONFloats(result.CV.RMSECV),
			RMSECVMean:                toJSONFloats(result.CV.RMSECVMean),
			RMSECVSE:                  toJSONFloats(result.CV.RMSECVSE),
			Bias:                      toJSONFloats(result.CV.Bias),
			SEP:                       toJSONFloats(result.CV.SEP),
			MAE:                       toJSONFloats(result.CV.MAE),
			MAESE:                     toJSONFloats(result.CV.MAESE),
			Q2:                        toJSONFloats(result.CV.Q2),
			Selected:                  result.CV.Selected,
			Rule:                      result.CV.Rule,
			Metric:                    result.CV.Metric,
			LowestError:               result.CV.LowestError,
			CurveStillFalling:         result.CV.CurveStillFalling,
			SelectedByAlternateMetric: result.CV.SelectedByAlternateMetric,
			OutOfFold:                 toJSONFloats(result.CV.OutOfFold),
		}
	}

	return out
}

// toJSONFloats wraps a float slice so that NaN and infinities survive marshalling.
func toJSONFloats(values []float64) []types.JSONFloat64 {
	if values == nil {
		return nil
	}
	out := make([]types.JSONFloat64, len(values))
	for i, v := range values {
		out[i] = types.JSONFloat64(v)
	}
	return out
}
