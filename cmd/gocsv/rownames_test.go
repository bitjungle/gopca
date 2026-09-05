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

func TestCheckRowNameCandidate(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		wantOK  bool
		wantHas string
	}{
		{name: "unique and complete", values: []string{"P1", "P2", "P3"}, wantOK: true},
		{name: "numeric ids are fine", values: []string{"1", "2", "3"}, wantOK: true},
		{
			name:    "a repeat is rejected",
			values:  []string{"P1", "P2", "P1"},
			wantHas: "unique",
		},
		{
			name:    "an empty cell is rejected",
			values:  []string{"P1", "", "P3"},
			wantHas: "complete",
		},
		{
			name:    "whitespace-only counts as empty",
			values:  []string{"P1", "   ", "P3"},
			wantHas: "complete",
		},
		{
			// They would render identically as a plot label, so calling them
			// distinct would defeat the point of the rule.
			name:    "values differing only by surrounding space collide",
			values:  []string{"P1", "P1 "},
			wantHas: "unique",
		},
		{
			// IDs are commonly case-bearing. Treating these as equal would
			// reject data that is in fact well-formed.
			name:   "case is significant",
			values: []string{"P1", "p1"},
			wantOK: true,
		},
		{
			name:    "empty column",
			values:  []string{},
			wantHas: "no rows",
		},
		{
			name:    "both problems are reported together",
			values:  []string{"P1", "P1", ""},
			wantHas: "unique and complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRowNameCandidate(tt.values)
			if got.OK != tt.wantOK {
				t.Fatalf("OK = %v, want %v (reason: %q)", got.OK, tt.wantOK, got.Reason)
			}
			if tt.wantOK {
				if got.Reason != "" {
					t.Errorf("an accepted column should carry no reason, got %q", got.Reason)
				}
				return
			}
			if !strings.Contains(got.Reason, tt.wantHas) {
				t.Errorf("reason %q should mention %q", got.Reason, tt.wantHas)
			}
		})
	}
}

// TestCheckRowNameCandidateNamesTheOffender checks the message is actionable.
// "row names must be unique" alone leaves the user to hunt for the duplicate.
func TestCheckRowNameCandidateNamesTheOffender(t *testing.T) {
	got := checkRowNameCandidate([]string{"P1", "P2", "P1"})
	if !strings.Contains(got.Reason, `"P1"`) {
		t.Errorf("the reason should name the repeated value, got %q", got.Reason)
	}

	counted := checkRowNameCandidate([]string{"a", "", "", "b"})
	if !strings.Contains(counted.Reason, "2 cells") {
		t.Errorf("the reason should count the empty cells, got %q", counted.Reason)
	}
}

// rowNameFixture is a table with row names already set, one text column and one
// numeric column.
func rowNameFixture() *FileData {
	return &FileData{
		Headers:            []string{"By", "Score"},
		RowNames:           []string{"P1", "P2", "P3"},
		RowNamesHeader:     "Prove",
		Data:               [][]string{{"Oslo", "10"}, {"Bergen", "12"}, {"Tromsø", "14"}},
		Rows:               3,
		Columns:            2,
		ColumnTypes:        map[string]string{"By": "categorical", "Score": "numeric"},
		CategoricalColumns: map[string][]string{"By": {"Oslo", "Bergen", "Tromsø"}},
	}
}

func headerIndex(headers []string, name string) int {
	for i, h := range headers {
		if h == name {
			return i
		}
	}
	return -1
}

// TestSetRowNamesIsASwap covers the central design decision: the outgoing row
// names come back as a column rather than being discarded.
//
// Correcting a bad guess by the loader is the reason the menu item exists, and
// a correction that destroys the previous identifier is not one. This is the
// same principle applied to one-hot encoding in #854.
func TestSetRowNamesIsASwap(t *testing.T) {
	data := rowNameFixture()
	app := &App{}

	cmd, err := NewSetRowNamesCommand(app, data, headerIndex(data.Headers, "By"))
	if err != nil {
		t.Fatalf("NewSetRowNamesCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// "By" is now the row names, carrying its header.
	if got := strings.Join(data.RowNames, ","); got != "Oslo,Bergen,Tromsø" {
		t.Errorf("RowNames = %q", got)
	}
	if data.RowNamesHeader != "By" {
		t.Errorf("RowNamesHeader = %q, want By", data.RowNamesHeader)
	}
	if headerIndex(data.Headers, "By") != -1 {
		t.Errorf("'By' should have left the table, headers are %v", data.Headers)
	}

	// The previous row names are back as column 0, under their own name.
	if data.Headers[0] != "Prove" {
		t.Fatalf("headers = %v, want 'Prove' first", data.Headers)
	}
	for i, want := range []string{"P1", "P2", "P3"} {
		if data.Data[i][0] != want {
			t.Errorf("row %d column 0 = %q, want %q", i, data.Data[i][0], want)
		}
	}
	if data.ColumnTypes["Prove"] != "categorical" {
		t.Errorf("the demoted column should be typed, got %q", data.ColumnTypes["Prove"])
	}
	if data.Columns != len(data.Headers) {
		t.Errorf("Columns = %d but there are %d headers", data.Columns, len(data.Headers))
	}
	for i, row := range data.Data {
		if len(row) != len(data.Headers) {
			t.Errorf("row %d has %d cells, want %d: %v", i, len(row), len(data.Headers), row)
		}
	}
	if _, stale := data.CategoricalColumns["By"]; stale {
		t.Error("'By' is still in CategoricalColumns after being promoted")
	}
}

// TestSetRowNamesUndoRestoresEverything is the other half of "safe to try".
func TestSetRowNamesUndoRestoresEverything(t *testing.T) {
	data := rowNameFixture()
	before := rowNameFixture()
	app := &App{}

	cmd, err := NewSetRowNamesCommand(app, data, headerIndex(data.Headers, "By"))
	if err != nil {
		t.Fatalf("NewSetRowNamesCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if strings.Join(data.Headers, ",") != strings.Join(before.Headers, ",") {
		t.Errorf("headers = %v, want %v", data.Headers, before.Headers)
	}
	if strings.Join(data.RowNames, ",") != strings.Join(before.RowNames, ",") {
		t.Errorf("RowNames = %v, want %v", data.RowNames, before.RowNames)
	}
	if data.RowNamesHeader != before.RowNamesHeader {
		t.Errorf("RowNamesHeader = %q, want %q", data.RowNamesHeader, before.RowNamesHeader)
	}
	for i := range before.Data {
		if strings.Join(data.Data[i], ",") != strings.Join(before.Data[i], ",") {
			t.Errorf("row %d = %v, want %v", i, data.Data[i], before.Data[i])
		}
	}
	for header, want := range before.ColumnTypes {
		if data.ColumnTypes[header] != want {
			t.Errorf("ColumnTypes[%q] = %q, want %q", header, data.ColumnTypes[header], want)
		}
	}
	if got := data.CategoricalColumns["By"]; strings.Join(got, ",") != "Oslo,Bergen,Tromsø" {
		t.Errorf("CategoricalColumns[By] = %v after undo", got)
	}
	if _, leftover := data.ColumnTypes["Prove"]; leftover {
		t.Error("the demoted column's type survived the undo")
	}
}

// TestSetRowNamesRefusesNonUniqueColumn is the requirement itself, at the layer
// that has to hold it. The dialog also greys the item out, but that is not
// where the rule lives.
func TestSetRowNamesRefusesNonUniqueColumn(t *testing.T) {
	data := rowNameFixture()
	data.Headers = append(data.Headers, "Region")
	data.ColumnTypes["Region"] = "categorical"
	for i, value := range []string{"Nord", "Sør", "Nord"} {
		data.Data[i] = append(data.Data[i], value)
	}
	data.Columns = len(data.Headers)

	app := &App{}
	_, err := app.ExecuteSetRowNames(data, headerIndex(data.Headers, "Region"))
	if err == nil {
		t.Fatal("a column with a repeated value was accepted as row names")
	}
	if !strings.Contains(err.Error(), "unique") || !strings.Contains(err.Error(), "Nord") {
		t.Errorf("the error should say what is wrong and name the value, got %v", err)
	}

	// And the data must be untouched by the refusal.
	if data.RowNamesHeader != "Prove" || len(data.Headers) != 3 {
		t.Errorf("a refused operation modified the data: headers=%v rowNamesHeader=%q",
			data.Headers, data.RowNamesHeader)
	}
}

// TestSetRowNamesOnFileWithoutRowNames covers the case with nothing to demote.
func TestSetRowNamesOnFileWithoutRowNames(t *testing.T) {
	data := rowNameFixture()
	data.RowNames = nil
	data.RowNamesHeader = ""

	app := &App{}
	cmd, err := NewSetRowNamesCommand(app, data, headerIndex(data.Headers, "By"))
	if err != nil {
		t.Fatalf("NewSetRowNamesCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(data.Headers) != 1 || data.Headers[0] != "Score" {
		t.Errorf("headers = %v, want just [Score]", data.Headers)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if len(data.Headers) != 2 || data.Headers[0] != "By" {
		t.Errorf("after undo headers = %v, want [By Score]", data.Headers)
	}
	if len(data.RowNames) != 0 {
		t.Errorf("undo invented row names: %v", data.RowNames)
	}
}

// TestMoveRowNamesIntoTable covers the reverse operation.
func TestMoveRowNamesIntoTable(t *testing.T) {
	data := rowNameFixture()
	app := &App{}

	cmd, err := NewMoveRowNamesIntoTableCommand(app, data)
	if err != nil {
		t.Fatalf("NewMoveRowNamesIntoTableCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(data.RowNames) != 0 || data.RowNamesHeader != "" {
		t.Errorf("row names should be cleared, got %v / %q", data.RowNames, data.RowNamesHeader)
	}
	if data.Headers[0] != "Prove" {
		t.Fatalf("headers = %v, want 'Prove' first", data.Headers)
	}
	if data.Data[0][0] != "P1" {
		t.Errorf("column 0 = %q, want P1", data.Data[0][0])
	}

	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if strings.Join(data.RowNames, ",") != "P1,P2,P3" || data.RowNamesHeader != "Prove" {
		t.Errorf("undo did not restore row names: %v / %q", data.RowNames, data.RowNamesHeader)
	}
	if len(data.Headers) != 2 {
		t.Errorf("undo left a column behind: %v", data.Headers)
	}
}

// TestMoveRowNamesIntoTableNamesAnUnnamedColumn covers the blank-header
// convention, which is what every dataset in testdata/ uses.
func TestMoveRowNamesIntoTableNamesAnUnnamedColumn(t *testing.T) {
	data := rowNameFixture()
	data.RowNamesHeader = ""

	app := &App{}
	cmd, err := NewMoveRowNamesIntoTableCommand(app, data)
	if err != nil {
		t.Fatalf("NewMoveRowNamesIntoTableCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// A blank column header is not addressable in the grid or in a dialog.
	if strings.TrimSpace(data.Headers[0]) == "" {
		t.Errorf("the restored column needs a name, headers are %v", data.Headers)
	}
}

// TestMoveRowNamesIntoTableAvoidsCollision guards the generated name.
func TestMoveRowNamesIntoTableAvoidsCollision(t *testing.T) {
	data := rowNameFixture()
	data.RowNamesHeader = "By" // same name as an existing column

	app := &App{}
	cmd, err := NewMoveRowNamesIntoTableCommand(app, data)
	if err != nil {
		t.Fatalf("NewMoveRowNamesIntoTableCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if data.Headers[0] == data.Headers[1] {
		t.Errorf("two columns share the name %q: %v", data.Headers[0], data.Headers)
	}
	if data.Data[0][0] != "P1" {
		t.Errorf("the restored column lost its values: %v", data.Data[0])
	}
}

// TestMoveRowNamesIntoTableRefusesWhenThereAreNone checks the guard exists.
func TestMoveRowNamesIntoTableRefusesWhenThereAreNone(t *testing.T) {
	data := rowNameFixture()
	data.RowNames = nil

	if _, err := (&App{}).ExecuteMoveRowNamesIntoTable(data); err == nil {
		t.Error("moving row names of a file that has none should fail")
	}
}

// TestClassifyColumnOnBlanks covers the two ways a column can contain no
// numbers.
//
// The numeric test used to start from true and skip blanks, so a column of
// nothing but empty cells never met a value that failed to parse and came out
// typed numeric. A demoted row-name column can be exactly that -- files whose
// first column is empty do exist -- and PCA would then be offered a variable
// with nothing in it.
func TestClassifyColumnOnBlanks(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{name: "all blank is not numeric", values: []string{"", "", ""}, want: "categorical"},
		{name: "whitespace only is not numeric", values: []string{" ", "\t"}, want: "categorical"},
		{name: "no values at all", values: []string{}, want: "categorical"},
		{name: "real numbers are numeric", values: []string{"1", "2.5", "-3"}, want: "numeric"},
		{name: "numbers with a gap are still numeric", values: []string{"1", "", "3"}, want: "numeric"},
		{name: "text is categorical", values: []string{"a", "b"}, want: "categorical"},
		{name: "one bad value makes it categorical", values: []string{"1", "2", "x"}, want: "categorical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &FileData{ColumnTypes: map[string]string{}}
			classifyColumn(data, "C", tt.values)
			if got := data.ColumnTypes["C"]; got != tt.want {
				t.Errorf("classified as %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSetRowNamesWithBlankPreviousRowNames is the path that reaches the case
// above through the command rather than directly.
func TestSetRowNamesWithBlankPreviousRowNames(t *testing.T) {
	data := rowNameFixture()
	data.RowNames = []string{"", "", ""}
	data.RowNamesHeader = ""

	cmd, err := NewSetRowNamesCommand(&App{}, data, headerIndex(data.Headers, "By"))
	if err != nil {
		t.Fatalf("NewSetRowNamesCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := data.ColumnTypes[data.Headers[0]]; got != "categorical" {
		t.Errorf("a demoted column of blanks was typed %q, want categorical", got)
	}
}
