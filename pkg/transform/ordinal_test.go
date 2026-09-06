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

package transform

import (
	"strings"
	"testing"
)

// ordinalInput builds a one-categorical-column input from the given cells.
func ordinalInput(values ...string) Input {
	data := make([][]string, len(values))
	present := make([]string, 0, len(values))
	for i, v := range values {
		data[i] = []string{v}
		present = append(present, v)
	}
	return Input{
		Data:               data,
		Headers:            []string{"Cat"},
		ColumnTypes:        map[string]string{"Cat": "categorical"},
		CategoricalColumns: map[string][]string{"Cat": present},
		Rows:               len(values),
		Columns:            1,
	}
}

// codeColumn returns the values of the generated code column.
func codeColumn(t *testing.T, res *Result, name string) []string {
	t.Helper()
	index := findColumn(res.Headers, name)
	if index == -1 {
		t.Fatalf("no column %q in %v", name, res.Headers)
	}
	got := make([]string, 0, len(res.Data))
	for _, row := range res.Data {
		got = append(got, row[index])
	}
	return got
}

// TestApply_Ordinal_MatchesSklearnLabelEncoderByDefault pins the default to
// scikit-learn's behaviour, using the example from the LabelEncoder docs.
//
// LabelEncoder sorts its classes_, so fitting on paris/paris/tokyo/amsterdam
// gives amsterdam=0, paris=1, tokyo=2. A caller who supplies no order gets
// exactly that, which is what makes the ordering control a superset of
// LabelEncoder rather than a different feature wearing its name.
func TestApply_Ordinal_MatchesSklearnLabelEncoderByDefault(t *testing.T) {
	in := ordinalInput("paris", "paris", "tokyo", "amsterdam")

	res, err := Apply(in, Options{Type: Ordinal, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	want := []string{"1", "1", "2", "0"}
	got := codeColumn(t, res, "Cat_code")
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got code %q, want %q (sklearn classes_ are sorted: "+
				"amsterdam=0, paris=1, tokyo=2); full column %v", i, got[i], want[i], got)
		}
	}

	if res.ColumnTypes["Cat_code"] != "numeric" {
		t.Errorf("the code column must be numeric for PCA to use it, got %q",
			res.ColumnTypes["Cat_code"])
	}
	if res.ColumnTypes["Cat"] != "categorical" {
		t.Error("the source column should be kept by default, as with one-hot")
	}
}

// TestApply_Ordinal_HonoursRequestedOrder is the reason this is not a straight
// port of LabelEncoder.
//
// low/medium/high is the canonical case: sorted alphabetically it becomes
// high=0, low=1, medium=2, which scrambles the very ordering the codes exist to
// carry. The assertion below fails if the requested order is ignored, and the
// alphabetical result is spelled out in the message so a failure is legible.
func TestApply_Ordinal_HonoursRequestedOrder(t *testing.T) {
	in := ordinalInput("low", "high", "medium", "low")

	res, err := Apply(in, Options{
		Type:          Ordinal,
		Columns:       []string{"Cat"},
		CategoryOrder: map[string][]string{"Cat": {"low", "medium", "high"}},
	})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	want := []string{"0", "2", "1", "0"}
	got := codeColumn(t, res, "Cat_code")
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("row %d: got %q, want %q. Requested order was low,medium,high; "+
				"alphabetical would give high=0,low=1,medium=2 -> %v",
				i, got[i], want[i], []string{"1", "0", "2", "1"})
		}
	}

	// The message must state the mapping actually used. A user who cannot see
	// which order was applied cannot tell these two cases apart.
	joined := strings.Join(res.Messages, " ")
	if !strings.Contains(joined, "low=0") || !strings.Contains(joined, "high=2") {
		t.Errorf("the result message should report the code mapping, got %v", res.Messages)
	}
}

// TestApply_Ordinal_BlanksStayBlank guards the choice not to encode missing
// values.
//
// Treating "" as a category would give it code 0 under alphabetical ordering --
// the low end of the scale -- and PCA would consume that as a real measurement
// rather than an absence.
func TestApply_Ordinal_BlanksStayBlank(t *testing.T) {
	in := ordinalInput("low", "", "high", "   ")

	res, err := Apply(in, Options{
		Type:          Ordinal,
		Columns:       []string{"Cat"},
		CategoryOrder: map[string][]string{"Cat": {"low", "high"}},
	})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	want := []string{"0", "", "1", ""}
	got := codeColumn(t, res, "Cat_code")
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %q, want %q; full column %v", i, got[i], want[i], got)
		}
	}
}

// TestApply_Ordinal_OrderResolution covers the two ways a supplied order can
// fail to line up with the data.
func TestApply_Ordinal_OrderResolution(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		order  []string
		want   []string
		why    string
	}{
		{
			name:   "partial order completed alphabetically",
			values: []string{"high", "low", "extra"},
			order:  []string{"low", "high"},
			want:   []string{"1", "0", "2"},
			why: "values missing from the requested order must still be encoded, " +
				"appended alphabetically after the ones that were named",
		},
		{
			name:   "requested value absent from the data",
			values: []string{"low", "high"},
			order:  []string{"low", "medium", "high"},
			want:   []string{"0", "1"},
			why: "codes must stay contiguous when a named category has no rows, " +
				"so filtering rows out does not leave a gap in the scale",
		},
		{
			name:   "duplicate in the requested order",
			values: []string{"low", "high"},
			order:  []string{"low", "low", "high"},
			want:   []string{"0", "1"},
			why:    "a repeated category must not consume two codes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Apply(ordinalInput(tt.values...), Options{
				Type:          Ordinal,
				Columns:       []string{"Cat"},
				CategoryOrder: map[string][]string{"Cat": tt.order},
			})
			if err != nil {
				t.Fatalf("Apply ordinal: %v", err)
			}
			got := codeColumn(t, res, "Cat_code")
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("row %d: got %q, want %q (%s); full column %v",
						i, got[i], tt.want[i], tt.why, got)
				}
			}
		})
	}
}

// TestApply_Ordinal_RemoveOriginal mirrors the one-hot behaviour: keeping the
// source is the default, removing it is opt-in.
func TestApply_Ordinal_RemoveOriginal(t *testing.T) {
	in := Input{
		Data:               [][]string{{"low", "1"}, {"high", "2"}},
		Headers:            []string{"Cat", "Num"},
		ColumnTypes:        map[string]string{"Cat": "categorical", "Num": "numeric"},
		CategoricalColumns: map[string][]string{"Cat": {"low", "high"}},
		Rows:               2,
		Columns:            2,
	}

	res, err := Apply(in, Options{Type: Ordinal, Columns: []string{"Cat"}, RemoveOriginal: true})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	if findColumn(res.Headers, "Cat") != -1 {
		t.Errorf("'Cat' should have been removed, headers were %v", res.Headers)
	}
	if _, exists := res.CategoricalColumns["Cat"]; exists {
		t.Error("'Cat' should have been removed from CategoricalColumns")
	}
	// The removal must shift the row cells too, not just the headers.
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers: %v",
				i, len(row), len(res.Headers), row)
		}
	}
	if res.Data[0][0] != "1" {
		t.Errorf("after removing 'Cat' the first cell should be Num's \"1\", got %q",
			res.Data[0][0])
	}
	// Removing the source must not take the codes with it.
	if got := codeColumn(t, res, "Cat_code"); got[0] != "1" || got[1] != "0" {
		t.Errorf("codes lost after removing the source: %v", got)
	}
}

// TestApply_Ordinal_NameCollision checks the generated name does not overwrite
// an existing column.
func TestApply_Ordinal_NameCollision(t *testing.T) {
	in := Input{
		Data:               [][]string{{"low", "x"}, {"high", "y"}},
		Headers:            []string{"Cat", "Cat_code"},
		ColumnTypes:        map[string]string{"Cat": "categorical", "Cat_code": "categorical"},
		CategoricalColumns: map[string][]string{"Cat": {"low", "high"}, "Cat_code": {"x", "y"}},
		Rows:               2,
		Columns:            2,
	}

	res, err := Apply(in, Options{Type: Ordinal, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	if len(res.NewColumns) != 1 {
		t.Fatalf("expected one new column, got %v", res.NewColumns)
	}
	if res.NewColumns[0] == "Cat_code" {
		t.Fatal("the new column reused the name of an existing column, overwriting it")
	}
	// The pre-existing column must be untouched.
	if got := codeColumn(t, res, "Cat_code"); got[0] != "x" || got[1] != "y" {
		t.Errorf("the existing 'Cat_code' column was overwritten: %v", got)
	}
}

// TestApply_Ordinal_SkipsNonCategorical checks the guard rather than assuming it.
func TestApply_Ordinal_SkipsNonCategorical(t *testing.T) {
	in := Input{
		Data:        [][]string{{"1"}, {"2"}},
		Headers:     []string{"Num"},
		ColumnTypes: map[string]string{"Num": "numeric"},
		Rows:        2,
		Columns:     1,
	}

	res, err := Apply(in, Options{Type: Ordinal, Columns: []string{"Num"}})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}
	if len(res.NewColumns) != 0 {
		t.Errorf("a numeric column should not be ordinal encoded, got %v", res.NewColumns)
	}
	if len(res.Messages) == 0 {
		t.Error("skipping a column should be reported")
	}
}

// TestGetTransformableColumns_Ordinal checks the dialog is offered the right
// columns. An option the UI cannot reach is not a feature.
func TestGetTransformableColumns_Ordinal(t *testing.T) {
	in := Input{
		Headers: []string{"Cat", "Num", "Score#target"},
		ColumnTypes: map[string]string{
			"Cat": "categorical", "Num": "numeric", "Score#target": "numeric",
		},
	}

	got := GetTransformableColumns(in, Ordinal)
	if len(got) != 1 || got[0] != "Cat" {
		t.Errorf("expected [Cat] to be offered for ordinal encoding, got %v", got)
	}
}
