package main

import (
	"strings"
	"testing"
)

// numberedFile returns a file with row numbers and three small numeric columns,
// the shape that makes the hazard in #942 visible: identifiers of 1..n against
// measurements of order 0.01.
func numberedFile(rows int) *FileData {
	data := make([][]string, rows)
	names := make([]string, rows)
	for i := range data {
		data[i] = []string{"0.01", "0.02", "0.03"}
		names[i] = itoa(i + 1)
	}
	return &FileData{
		Headers:        []string{"a", "b", "c"},
		Data:           data,
		RowNames:       names,
		RowNamesHeader: "Sample_ID",
		Rows:           rows,
		Columns:        3,
		ColumnTypes:    map[string]string{},
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func moveRowNamesIn(t *testing.T, data *FileData) *FileData {
	t.Helper()
	cmd, err := NewMoveRowNamesIntoTableCommand(nil, data)
	if err != nil {
		t.Fatalf("NewMoveRowNamesIntoTableCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	return data
}

// TestMovedRowNumbersDoNotBecomeAVariable is the round trip that mattered.
//
// Numbering the rows and then moving them into the table used to leave a numeric
// column of 1..n. On weight fractions that column takes 100% of PC1 and every
// real variable reads 0.0000 -- a completed-looking analysis of nothing but the
// row number (#942). The column must be held out instead.
func TestMovedRowNumbersDoNotBecomeAVariable(t *testing.T) {
	data := moveRowNamesIn(t, numberedFile(200))

	if data.Headers[0] != "Sample_ID#category" {
		t.Fatalf("column 0 is %q, want Sample_ID#category", data.Headers[0])
	}
	if got := data.ColumnTypes["Sample_ID#category"]; got != "categorical" {
		t.Errorf("moved row numbers are typed %q, want categorical; they would enter the PCA", got)
	}
	// Both maps, or the column is half converted and the encoders miss it.
	if values, ok := data.CategoricalColumns["Sample_ID#category"]; !ok {
		t.Error("the column is categorical by type but absent from CategoricalColumns")
	} else if len(values) != 200 || values[0] != "1" || values[199] != "200" {
		t.Errorf("CategoricalColumns holds %d values starting %q", len(values), values[0])
	}
	if len(data.RowNames) != 0 {
		t.Errorf("row names survived the move: %v", data.RowNames[:3])
	}
}

// TestMovedTextRowNamesAreLeftAlone keeps the fix narrow. Text row names are
// already categorical, so marking them would be noise.
func TestMovedTextRowNamesAreLeftAlone(t *testing.T) {
	data := numberedFile(5)
	for i := range data.RowNames {
		data.RowNames[i] = "se_" + data.RowNames[i]
	}
	moveRowNamesIn(t, data)

	if strings.Contains(data.Headers[0], "#category") {
		t.Errorf("text row names were marked: %q", data.Headers[0])
	}
	if data.Headers[0] != "Sample_ID" {
		t.Errorf("column 0 is %q, want Sample_ID", data.Headers[0])
	}
}

// TestMovingRowNamesIsUndoable checks the marker does not leak past an undo.
func TestMovingRowNamesIsUndoable(t *testing.T) {
	data := numberedFile(10)
	cmd, err := NewMoveRowNamesIntoTableCommand(nil, data)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("undo: %v", err)
	}

	if len(data.Headers) != 3 || data.Headers[0] != "a" {
		t.Errorf("headers after undo: %v", data.Headers)
	}
	if len(data.RowNames) != 10 || data.RowNames[0] != "1" {
		t.Errorf("row names not restored: %v", data.RowNames)
	}
	if data.RowNamesHeader != "Sample_ID" {
		t.Errorf("row-name header is %q, want Sample_ID", data.RowNamesHeader)
	}
	if _, marked := data.ColumnTypes["Sample_ID#category"]; marked {
		t.Errorf("the marker outlived the undo: %v", data.ColumnTypes)
	}
}

// TestBlankRowNamesAreNotCalledNumeric guards the same trap classifyColumn
// documents: an all-blank set must not be treated as numbers.
func TestBlankRowNamesAreNotCalledNumeric(t *testing.T) {
	if allNumeric([]string{"", "  ", ""}) {
		t.Error("an all-blank column was called numeric")
	}
	if !allNumeric([]string{"1", "", "3"}) {
		t.Error("numbers with a blank among them were not called numeric")
	}
	if allNumeric([]string{"1", "two", "3"}) {
		t.Error("a column containing text was called numeric")
	}
}

// TestMarkerSurvivesAHeaderCollision guards the hole review found in the first
// version of this fix.
//
// The marked name was built first and passed to uniqueHeader, which appends its
// suffix at the end: against a file that already had Sample_ID#category, that
// produced "Sample_ID#category_2". A marker is recognised by its suffix, so the
// column read back as an ordinary numeric one on export -- restoring the exact
// hazard the marker exists to prevent, for anyone whose file happened to carry
// a similarly named column.
func TestMarkerSurvivesAHeaderCollision(t *testing.T) {
	data := &FileData{
		Headers:        []string{"Sample_ID#category", "a"},
		Data:           [][]string{{"x", "0.01"}, {"y", "0.02"}},
		RowNames:       []string{"1", "2"},
		RowNamesHeader: "Sample_ID",
		Rows:           2,
		Columns:        2,
		ColumnTypes:    map[string]string{},
	}
	moveRowNamesIn(t, data)

	got := data.Headers[0]
	if !strings.HasSuffix(got, "#category") {
		t.Errorf("inserted header %q does not end in #category, so it would be read "+
			"back as a numeric column on export", got)
	}
	if got == "Sample_ID#category" {
		t.Errorf("inserted header %q collides with the existing column", got)
	}
	if got != "Sample_ID_2#category" {
		t.Errorf("inserted header = %q, want Sample_ID_2#category", got)
	}
}

// TestMarkedHeaderStaysUniqueAcrossSeveralCollisions checks the suffix keeps
// counting rather than giving up after one attempt.
func TestMarkedHeaderStaysUniqueAcrossSeveralCollisions(t *testing.T) {
	taken := []string{"S#category", "S_2#category", "S_3#category"}
	got := uniqueMarkedHeader(taken, "S", "#category")
	if got != "S_4#category" {
		t.Errorf("uniqueMarkedHeader = %q, want S_4#category", got)
	}
}
