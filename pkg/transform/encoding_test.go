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
	"testing"
)

// ─── One-hot encoding ─────────────────────────────────────────────────────────

func TestApply_OneHot_Basic(t *testing.T) {
	// 3 rows, 2 unique values ("a", "b") → 2 new columns, source kept.
	in := Input{
		Data:               [][]string{{"a"}, {"b"}, {"a"}},
		Headers:            []string{"Cat"},
		ColumnTypes:        map[string]string{"Cat": "categorical"},
		CategoricalColumns: map[string][]string{"Cat": {"a", "b", "a"}},
		Rows:               3,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("Apply one-hot: %v", err)
	}

	// Source column kept alongside the 2 new ones.
	if res.Columns != 3 {
		t.Errorf("expected 3 columns after one-hot (source + 2 encoded), got %d", res.Columns)
	}
	if len(res.NewColumns) != 2 {
		t.Errorf("expected 2 new columns, got %v", res.NewColumns)
	}
	if len(res.TransformedColumns) != 1 || res.TransformedColumns[0] != "Cat" {
		t.Errorf("expected TransformedColumns=[Cat], got %v", res.TransformedColumns)
	}

	// The source column must still be there, and still categorical -- GoPCA
	// colours plots by categorical columns, which is what keeping it buys.
	if res.ColumnTypes["Cat"] != "categorical" {
		t.Errorf("source column 'Cat' should be kept as categorical by default, got %q",
			res.ColumnTypes["Cat"])
	}

	// New columns must be numeric.
	for _, col := range res.NewColumns {
		if res.ColumnTypes[col] != "numeric" {
			t.Errorf("new column %q should be numeric, got %q", col, res.ColumnTypes[col])
		}
	}
}

func TestApply_OneHot_CorrectEncoding(t *testing.T) {
	// Values: row0="a", row1="b", row2="a"
	// Expected: Cat_a=[1,0,1], Cat_b=[0,1,0]
	in := Input{
		Data:               [][]string{{"a"}, {"b"}, {"a"}},
		Headers:            []string{"Cat"},
		ColumnTypes:        map[string]string{"Cat": "categorical"},
		CategoricalColumns: map[string][]string{},
		Rows:               3,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find column indices for Cat_a and Cat_b.
	aIdx, bIdx := -1, -1
	for i, h := range res.Headers {
		switch h {
		case "Cat_a":
			aIdx = i
		case "Cat_b":
			bIdx = i
		}
	}
	if aIdx == -1 || bIdx == -1 {
		t.Fatalf("expected headers Cat_a and Cat_b, got %v", res.Headers)
	}

	expected := map[int]struct{ a, b string }{
		0: {"1", "0"},
		1: {"0", "1"},
		2: {"1", "0"},
	}
	for row, exp := range expected {
		if res.Data[row][aIdx] != exp.a {
			t.Errorf("row %d Cat_a: expected %q, got %q", row, exp.a, res.Data[row][aIdx])
		}
		if res.Data[row][bIdx] != exp.b {
			t.Errorf("row %d Cat_b: expected %q, got %q", row, exp.b, res.Data[row][bIdx])
		}
	}
}

func TestApply_OneHot_KColumnsForKUniqueValues(t *testing.T) {
	// 4 unique values → 4 new binary columns.
	in := Input{
		Data:               [][]string{{"x"}, {"y"}, {"z"}, {"w"}, {"x"}},
		Headers:            []string{"C"},
		ColumnTypes:        map[string]string{"C": "categorical"},
		CategoricalColumns: map[string][]string{},
		Rows:               5,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.NewColumns) != 4 {
		t.Errorf("expected 4 new columns for 4 unique values, got %d", len(res.NewColumns))
	}
}

func TestApply_OneHot_TooManyUniqueValues(t *testing.T) {
	// 21 unique values → should be skipped (>20 limit).
	data := make([][]string, 21)
	for i := range data {
		data[i] = []string{string(rune('A' + i))}
	}
	in := Input{
		Data:               data,
		Headers:            []string{"C"},
		ColumnTypes:        map[string]string{"C": "categorical"},
		CategoricalColumns: map[string][]string{},
		Rows:               21,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.NewColumns) != 0 {
		t.Errorf("expected no new columns when >20 unique values, got %v", res.NewColumns)
	}
	if len(res.Messages) == 0 {
		t.Error("expected a message explaining the skip")
	}
}

func TestApply_OneHot_AllMissingColumn(t *testing.T) {
	// All values empty → should produce a message and no new columns.
	in := Input{
		Data:               [][]string{{""}, {""}},
		Headers:            []string{"C"},
		ColumnTypes:        map[string]string{"C": "categorical"},
		CategoricalColumns: map[string][]string{},
		Rows:               2,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.NewColumns) != 0 {
		t.Errorf("expected no new columns for all-missing column, got %v", res.NewColumns)
	}
}

func TestApply_OneHot_NonCategoricalColumn(t *testing.T) {
	in := Input{
		Data:               [][]string{{"1"}},
		Headers:            []string{"N"},
		ColumnTypes:        map[string]string{"N": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               1,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"N"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.NewColumns) != 0 {
		t.Error("expected no new columns for numeric column")
	}
}

func TestApply_OneHot_DoesNotMutateCategoricalColumns(t *testing.T) {
	// Result.CategoricalColumns must be a deep copy: mutating it must not
	// affect the original Input.CategoricalColumns slice.
	origSlice := []string{"a", "b"}
	in := Input{
		Data:               [][]string{{"a"}, {"b"}},
		Headers:            []string{"Cat"},
		ColumnTypes:        map[string]string{"Cat": "categorical"},
		CategoricalColumns: map[string][]string{"Cat": origSlice},
		Rows:               2,
		Columns:            1,
	}

	// Apply does NOT one-hot "Cat" here — we just want to verify the deep copy
	// of CategoricalColumns during Apply's setup, independent of one-hot.
	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = res

	// The original slice must be unchanged.
	if origSlice[0] != "a" || origSlice[1] != "b" {
		t.Error("Apply mutated the original CategoricalColumns slice")
	}
}

// oneHotInput is the smallest input that exercises the source-column decision.
func oneHotInput() Input {
	return Input{
		Data:               [][]string{{"a", "1"}, {"b", "2"}},
		Headers:            []string{"Cat", "Num"},
		ColumnTypes:        map[string]string{"Cat": "categorical", "Num": "numeric"},
		CategoricalColumns: map[string][]string{"Cat": {"a", "b"}},
		Rows:               2,
		Columns:            2,
	}
}

// TestApply_OneHot_OriginalColumnRemoved covers the opt-in destructive path.
func TestApply_OneHot_OriginalColumnRemoved(t *testing.T) {
	res, err := Apply(oneHotInput(), Options{
		Type:           OneHot,
		Columns:        []string{"Cat"},
		RemoveOriginal: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, h := range res.Headers {
		if h == "Cat" {
			t.Error("'Cat' should have been removed from Headers when RemoveOriginal is set")
		}
	}
	if _, exists := res.CategoricalColumns["Cat"]; exists {
		t.Error("'Cat' should have been removed from CategoricalColumns when RemoveOriginal is set")
	}
	if _, exists := res.ColumnTypes["Cat"]; exists {
		t.Error("'Cat' should have been removed from ColumnTypes when RemoveOriginal is set")
	}

	// The row data must shrink with the headers, not merely the metadata. A
	// removal that updates the headers and leaves the cells in place shifts
	// every subsequent column by one -- silently, and only in the data.
	if got := len(res.Headers); got != 3 {
		t.Fatalf("expected 3 headers (Num + 2 encoded), got %d: %v", got, res.Headers)
	}
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers: %v",
				i, len(row), len(res.Headers), row)
		}
	}
	if res.Data[0][0] != "1" {
		t.Errorf("after removing 'Cat', the first cell should be Num's value \"1\", got %q",
			res.Data[0][0])
	}
}

// TestApply_OneHot_OriginalColumnKeptByDefault is the counterpart, and the
// reason RemoveOriginal is spelled that way round: an Options built without
// thinking about the field must not discard a column.
//
// The default was the other way until this test existed. One-hot encoding was
// the only transformation in the package that destroyed its input, and the
// dialog said nothing about it.
func TestApply_OneHot_OriginalColumnKeptByDefault(t *testing.T) {
	res, err := Apply(oneHotInput(), Options{Type: OneHot, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, h := range res.Headers {
		if h == "Cat" {
			found = true
		}
	}
	if !found {
		t.Errorf("'Cat' should be kept by default, headers were %v", res.Headers)
	}
	if _, exists := res.CategoricalColumns["Cat"]; !exists {
		t.Error("'Cat' should still be in CategoricalColumns when kept")
	}

	// Kept means the values survive too, not just the header.
	if res.Data[0][0] != "a" || res.Data[1][0] != "b" {
		t.Errorf("the kept column lost its values: %v", res.Data)
	}
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers: %v",
				i, len(row), len(res.Headers), row)
		}
	}

	// Keeping the source must not suppress the encoding itself.
	if len(res.NewColumns) != 2 {
		t.Errorf("expected 2 encoded columns, got %v", res.NewColumns)
	}
	if len(res.TransformedColumns) != 1 || res.TransformedColumns[0] != "Cat" {
		t.Errorf("expected TransformedColumns=[Cat], got %v", res.TransformedColumns)
	}
	if len(res.Messages) == 0 {
		t.Error("expected a message describing the encoding")
	}
}
