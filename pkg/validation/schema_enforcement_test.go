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
	"io/fs"
	"path"
	"strings"
	"testing"
)

// TestSchemaGraphCompiles is the check that the schemas are actually usable.
//
// Before #834 they were embedded, two were read into strings, and none was ever
// parsed. A broken $ref, a malformed document or a missing file would have gone
// unnoticed indefinitely, because nothing ever asked the schemas to work.
func TestSchemaGraphCompiles(t *testing.T) {
	if _, err := compileSchemaGraph("v1"); err != nil {
		t.Fatalf("the v1 schema graph does not compile: %v", err)
	}
}

// TestEveryRefResolvesToAnEmbeddedFile keeps validation offline.
//
// The $id of every schema is an https:// URL. gojsonschema resolves a $ref from
// its pre-loaded set when it can and falls back to dereferencing the URL when it
// cannot -- so a $ref naming a file that is not in the embedded directory turns
// this package into something that reaches the network to validate a local file.
// That fails in an air-gapped build and stalls in a slow one, and it would do so
// only on the machines least able to debug it.
//
// This is a static check on the schema sources, so it fails on the edit that
// introduces the problem rather than on the deployment that suffers from it.
func TestEveryRefResolvesToAnEmbeddedFile(t *testing.T) {
	dir := path.Join("schemas", "v1")
	entries, err := fs.ReadDir(schemaFS, dir)
	if err != nil {
		t.Fatalf("reading embedded schemas: %v", err)
	}

	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name()] = true
	}
	if len(present) == 0 {
		t.Fatal("no embedded schemas found, so this test is passing by not looking")
	}

	refs := 0
	for name := range present {
		raw, err := fs.ReadFile(schemaFS, path.Join(dir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		var doc interface{}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("%s is not valid JSON: %v", name, err)
		}
		for _, ref := range collectRefs(doc) {
			refs++
			// A fragment-only ref points inside the same document.
			if strings.HasPrefix(ref, "#") {
				continue
			}
			file := strings.SplitN(ref, "#", 2)[0]
			if !present[file] {
				t.Errorf("%s refers to %q, which is not among the embedded schemas; "+
					"validation would have to fetch it over the network", name, ref)
			}
		}
	}
	if refs == 0 {
		t.Error("no $ref was found in any schema, which cannot be right and means " +
			"this test is not looking at what it thinks it is")
	}
}

func collectRefs(node interface{}) []string {
	var out []string
	switch v := node.(type) {
	case map[string]interface{}:
		for key, child := range v {
			if key == "$ref" {
				if s, ok := child.(string); ok {
					out = append(out, s)
				}
				continue
			}
			out = append(out, collectRefs(child)...)
		}
	case []interface{}:
		for _, child := range v {
			out = append(out, collectRefs(child)...)
		}
	}
	return out
}

// TestVarianceIsAPercentageNotAFraction pins the fix that turning validation on
// made necessary.
//
// The schema declared explained_variance_ratio and cumulative_variance as
// fractions with maximum 1. GoPCA has always emitted percentages: corn's first
// component is 97.495, not 0.97495. Every model the software has ever written
// violated its own published schema, and nothing noticed, because nothing ever
// checked. The schema now describes what the software does.
//
// The name is a genuine wart -- scikit-learn's explained_variance_ratio_ is a
// fraction, so a consumer comparing the two is out by a factor of 100 -- but the
// wire format is what it is, and changing it would break every saved model and
// every consumer. The schema descriptions say so explicitly instead.
func TestVarianceIsAPercentageNotAFraction(t *testing.T) {
	model := validModel(t, func(m map[string]interface{}) {
		components := m["model"].(map[string]interface{})
		components["explained_variance_ratio"] = []interface{}{97.495, 2.108}
		components["cumulative_variance"] = []interface{}{97.495, 99.603}
	})

	v, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("NewModelValidator: %v", err)
	}
	if err := v.ValidateModel(model); err != nil {
		t.Errorf("a real corn variance profile must validate, but: %v", err)
	}
}

// TestValidationRejectsWhatTheSchemaForbids proves the enforcement is real.
//
// Each case violates one rule that lives only in the JSON schema and in no Go
// code, so before #834 every one of them was accepted.
func TestValidationRejectsWhatTheSchemaForbids(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]interface{})
		wantErr string
	}{
		{
			name: "unknown PCA method",
			mutate: func(m map[string]interface{}) {
				cfg := m["metadata"].(map[string]interface{})["config"].(map[string]interface{})
				cfg["method"] = "quantum"
			},
			wantErr: "method",
		},
		{
			name: "negative component count",
			mutate: func(m map[string]interface{}) {
				cfg := m["metadata"].(map[string]interface{})["config"].(map[string]interface{})
				cfg["n_components"] = -3
			},
			wantErr: "n_components",
		},
		{
			name: "analysis id that is not a uuid",
			mutate: func(m map[string]interface{}) {
				m["metadata"].(map[string]interface{})["analysis_id"] = "model-7"
			},
			wantErr: "analysis_id",
		},
		{
			name: "variance above 100 percent",
			mutate: func(m map[string]interface{}) {
				m["model"].(map[string]interface{})["cumulative_variance"] =
					[]interface{}{140.0}
			},
			wantErr: "cumulative_variance",
		},
	}

	v, err := NewModelValidator("v1")
	if err != nil {
		t.Fatalf("NewModelValidator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The unmutated fixture must pass, or the case proves nothing about
			// the mutation.
			if err := v.ValidateModel(validModel(t, nil)); err != nil {
				t.Fatalf("the baseline fixture does not validate: %v", err)
			}

			err := v.ValidateModel(validModel(t, tt.mutate))
			if err == nil {
				t.Fatalf("%s was accepted; the schema forbids it", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error does not name %q: %v", tt.wantErr, err)
			}
		})
	}
}

// validModel returns a minimal model that satisfies the v1 schema, optionally
// mutated. Built fresh each call so one test cannot alter another's fixture.
func validModel(t *testing.T, mutate func(map[string]interface{})) []byte {
	t.Helper()

	const base = `{
      "$schema": "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
      "metadata": {
        "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
        "software_version": "1.7.0",
        "created_at": "2026-01-01T00:00:00Z",
        "software": "gopca",
        "config": {"method": "svd", "n_components": 2}
      },
      "preprocessing": {
        "mean_center": true, "standard_scale": true, "robust_scale": false,
        "scale_only": false, "snv": false, "vector_norm": false,
        "parameters": {}
      },
      "model": {
        "loadings": [[0.7, 0.7], [0.7, -0.7]],
        "explained_variance": [1.5, 0.5],
        "explained_variance_ratio": [75.0, 25.0],
        "cumulative_variance": [75.0, 100.0],
        "component_labels": ["PC1", "PC2"],
        "feature_labels": ["a", "b"]
      },
      "results": {
        "samples": {
          "names": ["s1", "s2"],
          "scores": [[1.0, 0.2], [-1.0, -0.2]]
        }
      }
    }`

	var model map[string]interface{}
	if err := json.Unmarshal([]byte(base), &model); err != nil {
		t.Fatalf("the fixture is not valid JSON: %v", err)
	}
	if mutate != nil {
		mutate(model)
	}
	out, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("re-encoding the fixture: %v", err)
	}
	return out
}
