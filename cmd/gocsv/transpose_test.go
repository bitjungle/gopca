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
	"strings"
	"testing"
)

// transposeFixture is a 3x2 table with row names and a named corner.
func transposeFixture() *FileData {
	return &FileData{
		Headers:            []string{"A", "B"},
		RowNames:           []string{"P1", "P2", "P3"},
		RowNamesHeader:     "Prove",
		Data:               [][]string{{"1", "x"}, {"2", "y"}, {"3", "z"}},
		Rows:               3,
		Columns:            2,
		ColumnTypes:        map[string]string{"A": "numeric", "B": "categorical"},
		CategoricalColumns: map[string][]string{"B": {"x", "y", "z"}},
	}
}

func gridOf(data *FileData) string {
	var b strings.Builder
	b.WriteString(data.RowNamesHeader + "|" + strings.Join(data.Headers, "|") + "\n")
	for i, row := range data.Data {
		name := ""
		if i < len(data.RowNames) {
			name = data.RowNames[i]
		}
		b.WriteString(name + "|" + strings.Join(row, "|") + "\n")
	}
	return b.String()
}

// TestTransposeMapping pins the layout, including which cell ends up where.
func TestTransposeMapping(t *testing.T) {
	data := transposeFixture()
	cmd, err := NewTransposeCommand(&App{}, data)
	if err != nil {
		t.Fatalf("NewTransposeCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "Prove|P1|P2|P3\nA|1|2|3\nB|x|y|z\n"
	if got := gridOf(data); got != want {
		t.Errorf("transposed to:\n%s\nwant:\n%s", got, want)
	}

	// The corner cell is unchanged -- that symmetry is what makes the operation
	// its own inverse.
	if data.RowNamesHeader != "Prove" {
		t.Errorf("RowNamesHeader = %q, want Prove", data.RowNamesHeader)
	}
	if data.Rows != 2 || data.Columns != 3 {
		t.Errorf("shape = %dx%d, want 2x3", data.Rows, data.Columns)
	}
}

// TestTransposeIsItsOwnInverse is the property that catches an off-by-one or a
// swapped index in a way that checking one direction cannot.
func TestTransposeIsItsOwnInverse(t *testing.T) {
	data := transposeFixture()
	original := gridOf(data)

	for i := 0; i < 2; i++ {
		cmd, err := NewTransposeCommand(&App{}, data)
		if err != nil {
			t.Fatalf("pass %d: %v", i, err)
		}
		if err := cmd.Execute(data); err != nil {
			t.Fatalf("pass %d: %v", i, err)
		}
	}

	if got := gridOf(data); got != original {
		t.Errorf("transposing twice gave:\n%s\nwant the original:\n%s", got, original)
	}
}

// TestTransposeRecomputesTypes covers the decision not to carry types across.
//
// A transposed dataset has entirely different columns. The old column "A" was
// numeric; after transposition "A" is a row, and the new column "P1" holds
// whatever the first row held -- here a number and a letter, so text.
func TestTransposeRecomputesTypes(t *testing.T) {
	data := transposeFixture()
	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, header := range []string{"P1", "P2", "P3"} {
		if data.ColumnTypes[header] != "categorical" {
			t.Errorf("column %q typed %q, want categorical (it holds a number and a letter)",
				header, data.ColumnTypes[header])
		}
		if _, ok := data.CategoricalColumns[header]; !ok {
			t.Errorf("column %q is categorical but absent from CategoricalColumns", header)
		}
	}

	// Types belonging to the old columns must not survive.
	for _, stale := range []string{"A", "B"} {
		if _, ok := data.ColumnTypes[stale]; ok {
			t.Errorf("type for old column %q survived transposition", stale)
		}
		if _, ok := data.CategoricalColumns[stale]; ok {
			t.Errorf("old column %q survived in CategoricalColumns", stale)
		}
	}
}

// TestTransposeNumericColumnsSurviveAsNumeric checks the recompute is a real
// classification and not a blanket "everything becomes text".
func TestTransposeNumericColumnsSurviveAsNumeric(t *testing.T) {
	data := &FileData{
		Headers:     []string{"A", "B"},
		RowNames:    []string{"P1", "P2"},
		Data:        [][]string{{"1", "2"}, {"3", "4"}},
		Rows:        2,
		Columns:     2,
		ColumnTypes: map[string]string{"A": "numeric", "B": "numeric"},
	}
	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, header := range []string{"P1", "P2"} {
		if data.ColumnTypes[header] != "numeric" {
			t.Errorf("column %q typed %q, want numeric", header, data.ColumnTypes[header])
		}
	}
}

// TestTransposeUndo restores the original exactly, types included.
func TestTransposeUndo(t *testing.T) {
	data := transposeFixture()
	before := gridOf(data)

	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if got := gridOf(data); got != before {
		t.Errorf("undo gave:\n%s\nwant:\n%s", got, before)
	}
	if data.ColumnTypes["A"] != "numeric" || data.ColumnTypes["B"] != "categorical" {
		t.Errorf("types not restored: %v", data.ColumnTypes)
	}
	if got := data.CategoricalColumns["B"]; strings.Join(got, ",") != "x,y,z" {
		t.Errorf("CategoricalColumns not restored: %v", got)
	}
}

// TestTransposeUndoIsNotAliased guards the deep copy.
//
// A shallow capture would leave the undo state sharing slices with the live
// data, so transposing would edit the "before" too and undo would restore the
// transposed values -- an undo that silently does nothing.
func TestTransposeUndoIsNotAliased(t *testing.T) {
	data := transposeFixture()
	cmd, _ := NewTransposeCommand(&App{}, data)

	// Mutate the live data before executing; the captured state must not move.
	data.Data[0][0] = "MUTATED"
	data.Headers[0] = "MUTATED"

	if cmd.before.Data[0][0] == "MUTATED" || cmd.before.Headers[0] == "MUTATED" {
		t.Fatal("the captured state shares memory with the live data")
	}
}

// TestTransposeNamesUnnamedRows covers a file loaded without row names.
func TestTransposeNamesUnnamedRows(t *testing.T) {
	data := transposeFixture()
	data.RowNames = nil

	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := []string{"Row_1", "Row_2", "Row_3"}
	for i, header := range want {
		if data.Headers[i] != header {
			t.Errorf("header %d = %q, want %q (a blank column header is not addressable)",
				i, data.Headers[i], header)
		}
	}
}

// TestTransposeSuffixesDuplicateRowNames covers row names that are not unique.
//
// The load path assigns the first column as row names without checking
// uniqueness (#859), so duplicates reach here. Headers addressed by name must
// be distinct, but the user asked to transpose and a repeated label is not a
// reason to decline.
func TestTransposeSuffixesDuplicateRowNames(t *testing.T) {
	data := transposeFixture()
	data.RowNames = []string{"P1", "P1", "P1"}

	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	seen := map[string]bool{}
	for _, header := range data.Headers {
		if seen[header] {
			t.Errorf("duplicate header %q after transposition: %v", header, data.Headers)
		}
		seen[header] = true
	}
	if len(data.Headers) != 3 {
		t.Errorf("expected 3 headers, got %v", data.Headers)
	}
}

// TestTransposeRaggedRows checks short rows read as empty rather than shifting
// the remaining values into the wrong columns.
func TestTransposeRaggedRows(t *testing.T) {
	data := &FileData{
		Headers:     []string{"A", "B", "C"},
		RowNames:    []string{"P1", "P2"},
		Data:        [][]string{{"1", "2", "3"}, {"4"}},
		Rows:        2,
		Columns:     3,
		ColumnTypes: map[string]string{},
	}

	cmd, _ := NewTransposeCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "|P1|P2\nA|1|4\nB|2|\nC|3|\n"
	if got := gridOf(data); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestTransposeRefusesEmpty checks the guard exists.
func TestTransposeRefusesEmpty(t *testing.T) {
	for _, data := range []*FileData{
		nil,
		{Headers: []string{"A"}},
		{Data: [][]string{{"1"}}},
	} {
		if _, err := NewTransposeCommand(&App{}, data); err == nil {
			t.Errorf("expected a refusal for %+v", data)
		}
	}
}

// TestTransposeWarnings covers what the user is told before committing.
func TestTransposeWarnings(t *testing.T) {
	data := transposeFixture()
	data.Headers = []string{"A", "Score#target"}
	data.RowNames = []string{"P1", "P1", ""}

	got := strings.Join((&App{}).TransposeWarnings(data), " ")

	// 2 headers and 3 data rows become 2 rows and 3 columns.
	for _, want := range []string{"Score#target", "repeat", "no name", "2 rows and 3 columns"} {
		if !strings.Contains(got, want) {
			t.Errorf("warnings should mention %q, got: %s", want, got)
		}
	}

	// A clean dataset gets the shape note and nothing alarming.
	clean := (&App{}).TransposeWarnings(transposeFixture())
	if len(clean) != 1 {
		t.Errorf("a clean dataset should warn only about the shape, got %v", clean)
	}
	if !strings.Contains(clean[0], "2 rows and 3 columns") {
		t.Errorf("shape note wrong: %v", clean)
	}
}

// TestTransposeWideDatasetHeadersUnique exercises the suffixing at the scale
// this feature exists for, in its worst case.
//
// Transposition is for wide files -- a 2000-column spectrum becomes 2000 rows
// -- and every generated header has to be checked against every other. With
// every row name identical, the suffix search is maximally contended. The
// original implementation rebuilt its lookup map on each of those 2000 calls.
func TestTransposeWideDatasetHeadersUnique(t *testing.T) {
	const n = 2000
	rows := make([][]string, n)
	names := make([]string, n)
	for i := range rows {
		rows[i] = []string{"1"}
		names[i] = "dup"
	}
	data := &FileData{
		Headers: []string{"A"}, RowNames: names, Data: rows,
		Rows: n, Columns: 1, ColumnTypes: map[string]string{"A": "numeric"},
	}

	cmd, err := NewTransposeCommand(&App{}, data)
	if err != nil {
		t.Fatalf("NewTransposeCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	seen := map[string]bool{}
	for _, header := range data.Headers {
		if seen[header] {
			t.Fatalf("duplicate header %q after transposing %d identical row names", header, n)
		}
		seen[header] = true
	}
	if len(data.Headers) != n {
		t.Errorf("got %d headers, want %d", len(data.Headers), n)
	}
}
