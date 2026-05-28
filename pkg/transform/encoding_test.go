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
	// 3 rows, 2 unique values ("a", "b") → 2 new columns, original removed.
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

	// Original column removed; 2 new columns added.
	if res.Columns != 2 {
		t.Errorf("expected 2 columns after one-hot, got %d", res.Columns)
	}
	if len(res.NewColumns) != 2 {
		t.Errorf("expected 2 new columns, got %v", res.NewColumns)
	}
	if len(res.TransformedColumns) != 1 || res.TransformedColumns[0] != "Cat" {
		t.Errorf("expected TransformedColumns=[Cat], got %v", res.TransformedColumns)
	}

	// Original column must be absent from ColumnTypes.
	if _, exists := res.ColumnTypes["Cat"]; exists {
		t.Error("original column 'Cat' should be removed from ColumnTypes")
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

func TestApply_OneHot_OriginalColumnRemoved(t *testing.T) {
	in := Input{
		Data:               [][]string{{"a"}, {"b"}},
		Headers:            []string{"Cat"},
		ColumnTypes:        map[string]string{"Cat": "categorical"},
		CategoricalColumns: map[string][]string{"Cat": {"a", "b"}},
		Rows:               2,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: OneHot, Columns: []string{"Cat"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, h := range res.Headers {
		if h == "Cat" {
			t.Error("original column 'Cat' should have been removed from Headers")
		}
	}
	if _, exists := res.CategoricalColumns["Cat"]; exists {
		t.Error("original column 'Cat' should have been removed from CategoricalColumns")
	}
}
