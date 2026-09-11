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
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// A workbook with several sheets has no defensible default. These tests pin the
// two halves of the handoff to the import wizard, which only works if both hold:
// loadExcel has to refuse, and suggestExcelImport has to say the wizard can read
// the file. Refusing without the suggestion would show the user an error where
// they used to get data (#905).

// writeSheets builds a workbook. The first grid is written to the default sheet;
// the rest get sheets of their own. Names in hidden are marked not-visible.
func writeSheets(t *testing.T, path string, sheets []string, grids [][][]string, hidden ...string) {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName(f.GetSheetName(0), sheets[0]); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	for _, name := range sheets[1:] {
		if _, err := f.NewSheet(name); err != nil {
			t.Fatalf("NewSheet(%s): %v", name, err)
		}
	}
	for i, grid := range grids {
		for r, row := range grid {
			for c, v := range row {
				if v == "" {
					continue
				}
				cell, err := excelize.CoordinatesToCellName(c+1, r+1)
				if err != nil {
					t.Fatalf("CoordinatesToCellName: %v", err)
				}
				if err := f.SetCellValue(sheets[i], cell, v); err != nil {
					t.Fatalf("SetCellValue: %v", err)
				}
			}
		}
	}
	for _, name := range hidden {
		if err := f.SetSheetVisible(name, false); err != nil {
			t.Fatalf("SetSheetVisible(%s): %v", name, err)
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
}

func table() [][]string {
	return [][]string{{"ID", "A", "B"}, {"S1", "1", "2"}, {"S2", "3", "4"}}
}

func TestMultiSheetWorkbookIsHandedToTheWizard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "two.xlsx")
	writeSheets(t, path, []string{"Data", "Notes"}, [][][]string{table(), {{"ignore me"}}})

	app := &App{}

	// Half one: it must not quietly load the first sheet.
	if _, err := app.loadExcel(path); err == nil {
		t.Fatal("loadExcel accepted a two-sheet workbook; the second sheet would be discarded silently")
	} else {
		if !strings.Contains(err.Error(), "Data") || !strings.Contains(err.Error(), "Notes") {
			t.Errorf("error should name the sheets so the message is actionable, got: %v", err)
		}
	}

	// Half two: without this the user sees an error instead of the wizard.
	suggestion, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("suggestExcelImport: %v", err)
	}
	if !suggestion.NeedsWizard {
		t.Fatal("refused the file but did not offer the wizard: the user would see a dead end")
	}
	if suggestion.Sheet != "Data" {
		t.Errorf("Sheet = %q, want the first visible sheet %q", suggestion.Sheet, "Data")
	}
}

func TestSingleSheetWorkbookStillLoadsDirectly(t *testing.T) {
	// The control. A refusal that fired for every workbook would satisfy the
	// test above just as well.
	path := filepath.Join(t.TempDir(), "one.xlsx")
	writeSheets(t, path, []string{"Sheet1"}, [][][]string{table()})

	app := &App{}
	data, err := app.loadExcel(path)
	if err != nil {
		t.Fatalf("loadExcel refused a single-sheet workbook: %v", err)
	}
	if data.Rows != 2 {
		t.Errorf("rows = %d, want 2", data.Rows)
	}

	suggestion, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("suggestExcelImport: %v", err)
	}
	if suggestion.NeedsWizard {
		t.Error("an unambiguous sheet should not interrupt the user with the wizard")
	}
}

func TestHiddenSheetsDoNotCountAsAChoice(t *testing.T) {
	// Tools emit hidden sheets for their own bookkeeping. Prompting for a
	// choice between one real sheet and one the user cannot see is noise.
	path := filepath.Join(t.TempDir(), "hidden.xlsx")
	writeSheets(t, path, []string{"Data", "_config"}, [][][]string{table(), {{"internal"}}}, "_config")

	app := &App{}
	data, err := app.loadExcel(path)
	if err != nil {
		t.Fatalf("a hidden second sheet triggered the wizard: %v", err)
	}
	if data.Rows != 2 {
		t.Errorf("rows = %d, want 2", data.Rows)
	}
}

func TestMultiSheetAndTitleBlockBothReported(t *testing.T) {
	// The two reasons compose: the wizard needs to be told which sheet *and*
	// where the table starts.
	path := filepath.Join(t.TempDir(), "both.xlsx")
	titled := [][]string{{"Experiment 7"}, {"ID", "A", "B"}, {"S1", "1", "2"}}
	writeSheets(t, path, []string{"Data", "Notes"}, [][][]string{titled, {{"x"}}})

	suggestion, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("suggestExcelImport: %v", err)
	}
	if !suggestion.NeedsWizard {
		t.Fatal("NeedsWizard = false")
	}
	if suggestion.SkipRows != 1 {
		t.Errorf("SkipRows = %d, want 1 so the wizard still locates the table", suggestion.SkipRows)
	}
}
