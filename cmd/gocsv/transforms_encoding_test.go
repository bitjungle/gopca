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

// TestOrdinalCategoryOrderCrossesTheWire follows the category order the dialog
// sends through to the codes that come out.
//
// CategoryOrder is a map copied by hand from TransformOptions into
// transform.Options, exactly like RemoveOriginal above. Dropping the copy
// leaves both sides compiling and every unit test on either side green, and
// the only visible symptom is that the codes come out alphabetical -- which
// looks like a plausible result rather than a failure.
func TestOrdinalCategoryOrderCrossesTheWire(t *testing.T) {
	// What the dialog sends after the user confirms the suggested order.
	const payload = `{"type":"ordinal","columns":["Quality"],` +
		`"categoryOrder":{"Quality":["lav","middels","høy"]}}`

	var options TransformOptions
	if err := json.Unmarshal([]byte(payload), &options); err != nil {
		t.Fatalf("unmarshalling the dialog payload: %v", err)
	}

	data := &FileData{
		Headers:            []string{"Quality"},
		Data:               [][]string{{"høy"}, {"lav"}, {"middels"}},
		Rows:               3,
		Columns:            1,
		ColumnTypes:        map[string]string{"Quality": "categorical"},
		CategoricalColumns: map[string][]string{"Quality": {"høy", "lav", "middels"}},
	}

	app := &App{}
	res, err := app.applyTransformationInternal(data, options)
	if err != nil {
		t.Fatalf("applyTransformationInternal: %v", err)
	}
	if res.Data == nil {
		t.Fatal("transformation returned no data")
	}

	index := -1
	for i, h := range res.Data.Headers {
		if h == "Quality_code" {
			index = i
		}
	}
	if index == -1 {
		t.Fatalf("no Quality_code column in %v", res.Data.Headers)
	}

	// lav=0, middels=1, høy=2. Alphabetically it would be høy=0, lav=1,
	// middels=2, giving 0,1,2 for these rows -- so the wrong answer is not
	// obviously wrong from the shape of the column alone.
	want := []string{"2", "0", "1"}
	for i, row := range res.Data.Data {
		if row[index] != want[i] {
			t.Errorf("row %d (%q): got code %q, want %q; the requested order was "+
				"lav,middels,høy but the column came out %v",
				i, data.Data[i][0], row[index], want[i], columnOf(res.Data.Data, index))
		}
	}
}

// TestSuggestCategoryOrderBinding checks what the dialog pre-fills its control
// with, including the empty cases it would otherwise have to guard.
func TestSuggestCategoryOrderBinding(t *testing.T) {
	data := &FileData{
		Headers:     []string{"Quality"},
		Data:        [][]string{{"høy"}, {"lav"}, {"lav"}, {"middels"}},
		Rows:        4,
		Columns:     1,
		ColumnTypes: map[string]string{"Quality": "categorical"},
	}

	app := &App{}
	got := app.SuggestCategoryOrder(data, "Quality")
	want := []string{"lav", "middels", "høy"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d: got %q, want %q; full order %v", i, got[i], want[i], got)
		}
	}

	// A missing column and a nil FileData must return an empty list rather than
	// null, so the dialog has nothing to special-case.
	if got := app.SuggestCategoryOrder(data, "Nope"); got == nil || len(got) != 0 {
		t.Errorf("unknown column should give an empty list, got %v", got)
	}
	if got := app.SuggestCategoryOrder(nil, "Quality"); got == nil || len(got) != 0 {
		t.Errorf("nil data should give an empty list, got %v", got)
	}
}

func columnOf(data [][]string, index int) []string {
	out := make([]string, 0, len(data))
	for _, row := range data {
		if index < len(row) {
			out = append(out, row[index])
		}
	}
	return out
}
