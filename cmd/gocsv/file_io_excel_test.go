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
