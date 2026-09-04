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

package validation

import (
	"encoding/json"
	"strings"
	"testing"
)

// baseModel is the smallest structure ValidateModel accepts, used as the ground
// each regression case is built on.
const baseModel = `{
  "metadata": {
    "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
    "software_version": "1", "created_at": "2026-01-01T00:00:00Z",
    "software": "gopca", "config": {"method": "svd", "n_components": 2}
  },
  "preprocessing": {
    "mean_center": true, "standard_scale": true, "robust_scale": false,
    "scale_only": false, "snv": false, "vector_norm": false, "parameters": {}
  },
  "model": {
    "loadings": [[1,0],[0,1]], "explained_variance": [1,1],
    "explained_variance_ratio": [50,50], "cumulative_variance": [50,100],
    "component_labels": ["PC1","PC2"], "feature_labels": ["a","b"]
  },
  "results": {"samples": {"names": ["s1"], "scores": [[1,1]]}}
}`

func withRegression(t *testing.T, regression string) []byte {
	t.Helper()
	var model map[string]interface{}
	if err := json.Unmarshal([]byte(baseModel), &model); err != nil {
		t.Fatalf("base model is not valid JSON: %v", err)
	}
	if regression != "" {
		var block interface{}
		if err := json.Unmarshal([]byte(regression), &block); err != nil {
			t.Fatalf("regression block is not valid JSON: %v", err)
		}
		model["regression"] = block
	}
	encoded, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return encoded
}

// TestValidateModelAcceptsModelsWithoutRegression pins the backward compatibility
// this design rests on: the block is additive, so every model produced before
// principal component regression existed must still validate untouched.
func TestValidateModelAcceptsModelsWithoutRegression(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("NewModelValidator: %v", err)
	}
	if err := validator.ValidateModel(withRegression(t, "")); err != nil {
		t.Errorf("a model without a regression block must still validate: %v", err)
	}
}

func TestValidateRegressionBlock(t *testing.T) {
	validator, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("NewModelValidator: %v", err)
	}

	tests := []struct {
		name       string
		regression string
		wantErr    string
	}{
		{
			name: "complete block",
			regression: `{"response":"y","components":2,"score_coefficients":[1.5,-0.5],
				"intercept":3.0,"original_scale_valid":true,"coefficients":[0.1,0.2],
				"intercept_original":2.5}`,
		},
		{
			name: "intercept-only model",
			regression: `{"response":"y","components":0,"score_coefficients":[],
				"intercept":3.0,"original_scale_valid":true,"coefficients":[0,0]}`,
		},
		{
			name: "row-wise preprocessing carries no collapsed form",
			regression: `{"response":"y","components":2,"score_coefficients":[1,2],
				"intercept":3.0,"original_scale_valid":false}`,
		},
		{
			name: "missing response",
			regression: `{"components":1,"score_coefficients":[1],"intercept":0,
				"original_scale_valid":false}`,
			wantErr: "response",
		},
		{
			name: "component count disagrees with coefficients",
			regression: `{"response":"y","components":5,"score_coefficients":[1,2],
				"intercept":0,"original_scale_valid":false}`,
			wantErr: "score coefficients",
		},
		{
			name: "claims a collapsed form it does not carry",
			regression: `{"response":"y","components":1,"score_coefficients":[1],
				"intercept":0,"original_scale_valid":true}`,
			wantErr: "does not carry",
		},
		{
			name: "negative component count",
			regression: `{"response":"y","components":-1,"score_coefficients":[],
				"intercept":0,"original_scale_valid":false}`,
			wantErr: "regression.components: Must be greater than or equal to 0",
		},
		{
			name:       "not an object",
			regression: `"a string"`,
			wantErr:    "object",
		},
		{
			// Absent, this field unmarshals to false and silently reclassifies a
			// model that does carry a collapsed form as one that does not. That is
			// a change of meaning, so it is required rather than defaulted.
			name:       "original_scale_valid absent",
			regression: `{"response":"y","components":1,"score_coefficients":[1],"intercept":0}`,
			wantErr:    "regression: original_scale_valid is required",
		},
		{
			name: "original_scale_valid not a boolean",
			regression: `{"response":"y","components":1,"score_coefficients":[1],
				"intercept":0,"original_scale_valid":"yes"}`,
			wantErr: "boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModel(withRegression(t, tt.regression))
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected the block to validate: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error mentioning %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}
