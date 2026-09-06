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
	"testing"
)

// TestOneHotSourceColumnCrossesTheWire follows the "keep original column"
// checkbox from the JSON the dialog sends to the data that comes back.
//
// The flag is declared three times -- as a checkbox in DataTransformDialog.tsx,
// as TransformOptions.RemoveOriginal here, and as transform.Options.RemoveOriginal
// in pkg/transform -- and copied by hand between the second and third. Tests on
// either side compile against their own declaration and stay green while the
// value is dropped in between, so this one starts from the wire format and
// asserts on the resulting columns.
//
// The payloads below are what the dialog actually sends: it emits
// `removeOriginal: !keepOriginal` for one-hot and `undefined` otherwise, and
// `undefined` is omitted from the JSON entirely. The omitted case is the one
// that matters, because it is what every other transformation sends -- and
// under the old behaviour it destroyed a column.
func TestOneHotSourceColumnCrossesTheWire(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantKept bool
	}{
		{
			name:     "checkbox ticked",
			payload:  `{"type":"onehot","columns":["Cat"],"removeOriginal":false}`,
			wantKept: true,
		},
		{
			name:     "checkbox unticked",
			payload:  `{"type":"onehot","columns":["Cat"],"removeOriginal":true}`,
			wantKept: false,
		},
		{
			name:     "field absent from the payload",
			payload:  `{"type":"onehot","columns":["Cat"]}`,
			wantKept: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var options TransformOptions
			if err := json.Unmarshal([]byte(tt.payload), &options); err != nil {
				t.Fatalf("unmarshalling the dialog payload: %v", err)
			}

			data := &FileData{
				Headers:            []string{"Cat", "Num"},
				Data:               [][]string{{"a", "1"}, {"b", "2"}},
				Rows:               2,
				Columns:            2,
				ColumnTypes:        map[string]string{"Cat": "categorical", "Num": "numeric"},
				CategoricalColumns: map[string][]string{"Cat": {"a", "b"}},
			}

			app := &App{}
			res, err := app.applyTransformationInternal(data, options)
			if err != nil {
				t.Fatalf("applyTransformationInternal: %v", err)
			}
			if res.Data == nil {
				t.Fatal("transformation returned no data")
			}

			kept := false
			for _, h := range res.Data.Headers {
				if h == "Cat" {
					kept = true
				}
			}
			if kept != tt.wantKept {
				t.Errorf("source column kept = %v, want %v; headers were %v",
					kept, tt.wantKept, res.Data.Headers)
			}

			// The encoding must happen either way -- a flag that also
			// suppressed the transformation would satisfy the check above.
			if len(res.Data.Headers) < 3 {
				t.Errorf("expected the encoded columns to be present, headers were %v",
					res.Data.Headers)
			}
			for i, row := range res.Data.Data {
				if len(row) != len(res.Data.Headers) {
					t.Errorf("row %d has %d cells but there are %d headers: %v",
						i, len(row), len(res.Data.Headers), row)
				}
			}
		})
	}
}
