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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// setCells writes a grid into a new sheet, skipping empty strings so that the
// cells are genuinely absent rather than blank. That distinction matters: it is
// what makes excelize's GetRows trim the row and return a short slice.
func setCells(t *testing.T, path string, grid [][]string) {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	for r, row := range grid {
		for c, v := range row {
			if v == "" {
				continue
			}
			cell, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				t.Fatalf("CoordinatesToCellName: %v", err)
			}
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatalf("SetCellValue(%s): %v", cell, err)
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
}

// assertRagged fails if the fixture does not actually produce rows of differing
// widths. Without this the tests below could pass against a rectangular sheet and
// prove nothing about the regression.
func assertRagged(t *testing.T, path string) {
	t.Helper()
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	widths := map[int]bool{}
	for _, row := range rows {
		widths[len(row)] = true
	}
	if len(widths) < 2 {
		t.Fatalf("fixture is not ragged (widths %v); the regression cannot occur", widths)
	}
}

// TestIssue799_RaggedTableLoads covers the core defect. GetRows trims trailing
// empty cells, so a table with gaps in its last column comes back with rows of
// differing widths; serialising those verbatim made encoding/csv reject the first
// disagreeing row with "wrong number of fields". The table here starts at row 1,
// so once the rows are padded it loads normally.
func TestIssue799_RaggedTableLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ragged.xlsx")
	setCells(t, path, [][]string{
		{"Sample", "A", "B", "C", "D"},
		{"S1", "1", "2", "3", "4"},
		{"S2", "5", "6", "7", ""}, // trailing gap: GetRows returns 4 fields
		{"S3", "8", "9", "10", "11"},
		{"S4", "12", "13", "", ""}, // two trailing gaps: 3 fields
	})
	assertRagged(t, path)

	data, err := NewApp().loadExcel(path)
	if err != nil {
		t.Fatalf("loadExcel on a ragged table failed: %v", err)
	}
	for i, row := range data.Data {
		if len(row) != data.Columns {
			t.Errorf("row %d has %d fields, want %d — the table is not rectangular",
				i, len(row), data.Columns)
		}
	}
	if data.Rows != 4 {
		t.Errorf("Rows = %d, want 4", data.Rows)
	}
}

// TestIssue799_TitleBlockIsReportedActionably covers the layout that was actually
// reported: a title block above the table. Padding alone does not rescue it —
// the preamble becomes the header and first data rows, so no column reads as a
// consistent type. loadExcel cannot safely guess where the table starts, so it
// must at least say what it found and name the tool that can be told.
func TestIssue799_TitleBlockIsReportedActionably(t *testing.T) {
	path := filepath.Join(t.TempDir(), "titleblock.xlsx")
	grid := [][]string{
		{}, // blank spacer
		{"", "Excel Sample Data"},
		{}, // blank spacer
		{"", "Inventory Records Data"},
		{}, // blank spacer
		{"", "Product ID", "Name", "Stock", "Sold"},
	}
	for r := 0; r < 3; r++ {
		grid = append(grid, []string{"", fmt.Sprintf("P%03d", r+101), "Widget", "50", "10"})
	}
	setCells(t, path, grid)
	assertRagged(t, path)

	_, err := NewApp().loadExcel(path)
	if err == nil {
		t.Fatal("a sheet with a title block loaded silently; the header would be a blank row")
	}
	// The message must tell the user where the table starts and what to do.
	if !strings.Contains(err.Error(), "row 6") {
		t.Errorf("error does not locate the table (want it to name row 6): %v", err)
	}
	if !strings.Contains(err.Error(), "Import with Wizard") {
		t.Errorf("error does not point at the tool that can express the answer: %v", err)
	}
}

// TestIssue799_EmptySheetIsRejected guards the width==0 branch: a sheet with no
// populated cells must be reported as having no data rather than serialised into
// blank lines and failing later with a less clear message.
func TestIssue799_EmptySheetIsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.xlsx")
	f := excelize.NewFile()
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	_ = f.Close()

	if _, err := NewApp().loadExcel(path); err == nil {
		t.Error("an empty sheet loaded without error; expected a no-data failure")
	}
}

// TestIssue799_SuggestExcelImportLocatesTheTable checks the structured signal the
// frontend uses to offer the wizard handoff. It must name the exact number of
// rows to skip, since that value is pre-filled into the wizard and a wrong count
// would silently import the wrong header row.
func TestIssue799_SuggestExcelImportLocatesTheTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "titleblock.xlsx")
	setCells(t, path, [][]string{
		{},
		{"", "Excel Sample Data"},
		{},
		{"", "Inventory Records Data"},
		{},
		{"", "Product ID", "Name", "Stock", "Sold"},
		{"", "P101", "Widget", "50", "10"},
	})

	got, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("SuggestExcelImport: %v", err)
	}
	if !got.NeedsWizard {
		t.Fatal("NeedsWizard is false; the handoff would never be offered")
	}
	if got.SkipRows != 5 {
		t.Errorf("SkipRows = %d, want 5 — the wizard would start on the wrong row", got.SkipRows)
	}
	if got.DataRow != 6 {
		t.Errorf("DataRow = %d, want 6", got.DataRow)
	}
	if got.Sheet == "" {
		t.Error("Sheet is empty; the wizard needs to know which sheet to read")
	}
}

// TestIssue799_OrdinarySheetNeedsNoWizard is the other half: a sheet that is
// nothing but its table must not trigger the handoff, or every failed load would
// send the user into the wizard for no reason.
func TestIssue799_OrdinarySheetNeedsNoWizard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.xlsx")
	setCells(t, path, [][]string{
		{"Sample", "A", "B"},
		{"S1", "1", "2"},
		{"S2", "3", ""}, // still ragged, but the table starts at row 1
	})

	got, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("SuggestExcelImport: %v", err)
	}
	if got.NeedsWizard {
		t.Errorf("NeedsWizard is true for a sheet with no preamble (SkipRows=%d)", got.SkipRows)
	}
}

// TestIssue799_SuggestionFollowsTheFailedLoad covers the path the frontend
// actually uses. LoadCSV opens the file dialog in the backend, so the frontend
// never learns the path; the suggestion has to follow the recorded failure.
// A stale or empty record must not send the user into the wizard.
func TestIssue799_SuggestionFollowsTheFailedLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "titleblock.xlsx")
	setCells(t, path, [][]string{
		{},
		{"", "Excel Sample Data"},
		{},
		{"", "Inventory Records Data"},
		{},
		{"", "Product ID", "Name", "Stock"},
		{"", "P101", "Widget", "50"},
	})

	app := NewApp()

	// Nothing has failed yet, so nothing should be suggested.
	if got, err := app.SuggestImportForFailedLoad(); err != nil {
		t.Fatalf("SuggestImportForFailedLoad on a fresh app: %v", err)
	} else if got.NeedsWizard {
		t.Error("a fresh app suggested the wizard; the user would be sent there unprompted")
	}

	if _, err := app.LoadCSV(path); err == nil {
		t.Fatal("the title-block sheet loaded; this test needs the failure it follows")
	}

	got, err := app.SuggestImportForFailedLoad()
	if err != nil {
		t.Fatalf("SuggestImportForFailedLoad: %v", err)
	}
	if !got.NeedsWizard {
		t.Fatal("no wizard suggested after the load that should trigger it")
	}
	if got.SkipRows != 5 {
		t.Errorf("SkipRows = %d, want 5", got.SkipRows)
	}

	// A subsequent successful load must clear the record.
	plain := filepath.Join(t.TempDir(), "plain.xlsx")
	setCells(t, plain, [][]string{
		{"Sample", "A", "B"},
		{"S1", "1", "2"},
		{"S2", "3", "4"},
	})
	if _, err := app.LoadCSV(plain); err != nil {
		t.Fatalf("plain sheet failed to load: %v", err)
	}
	if got, err := app.SuggestImportForFailedLoad(); err != nil {
		t.Fatalf("SuggestImportForFailedLoad after success: %v", err)
	} else if got.NeedsWizard {
		t.Error("a stale failure still suggests the wizard after a successful load")
	}
}

// TestIssue799_SuggestionCarriesThePath guards the value the frontend needs most:
// without FilePath the wizard cannot be opened on the file, and the handoff
// silently does nothing. It crosses the Go/JSON boundary, so check it there too.
func TestIssue799_SuggestionCarriesThePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "titleblock.xlsx")
	setCells(t, path, [][]string{
		{},
		{"", "Title"},
		{"", "Product ID", "Name", "Stock"},
		{"", "P101", "Widget", "50"},
	})

	app := NewApp()
	if _, err := app.LoadCSV(path); err == nil {
		t.Fatal("the title-block sheet loaded; this test needs the failure it follows")
	}
	got, err := app.SuggestImportForFailedLoad()
	if err != nil {
		t.Fatalf("SuggestImportForFailedLoad: %v", err)
	}
	if got.FilePath != path {
		t.Errorf("FilePath = %q, want %q — the wizard could not be opened on the file", got.FilePath, path)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded struct {
		FilePath string `json:"filePath"`
		SkipRows int    `json:"skipRows"`
		Needs    bool   `json:"needsWizard"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.FilePath != path || decoded.SkipRows != got.SkipRows || !decoded.Needs {
		t.Errorf("the frontend would receive %+v, want path=%q skipRows=%d needsWizard=true",
			decoded, path, got.SkipRows)
	}
}
