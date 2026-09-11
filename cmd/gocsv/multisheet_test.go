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
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
	if len(sheets) == 0 {
		t.Fatal("writeSheets: need at least one sheet name")
	}
	if len(grids) != len(sheets) {
		t.Fatalf("writeSheets: %d grids for %d sheets", len(grids), len(sheets))
	}
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

// hideEverySheet rewrites a workbook so that every sheet is marked hidden.
//
// excelize will not produce this through SetSheetVisible -- like Excel, it keeps
// one sheet visible -- so the state has to be written into xl/workbook.xml
// directly. Files in the wild do reach it, and the first version of this test
// tried to build the case with SetSheetVisible, quietly got one visible sheet
// and two hidden, and proved nothing about the case it was named for.
func hideEverySheet(t *testing.T, path string) {
	t.Helper()
	in, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = in.Close() }()

	var buf bytes.Buffer
	out := zip.NewWriter(&buf)
	pattern := regexp.MustCompile(`(<sheet name="[^"]*" sheetId="\d+")`)
	for _, file := range in.File {
		r, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		content, err := io.ReadAll(r)
		_ = r.Close()
		if err != nil {
			t.Fatalf("read %s: %v", file.Name, err)
		}
		if file.Name == "xl/workbook.xml" {
			content = pattern.ReplaceAll(content, []byte(`$1 state="hidden"`))
		}
		w, err := out.Create(file.Name)
		if err != nil {
			t.Fatalf("create %s: %v", file.Name, err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("write %s: %v", file.Name, err)
		}
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

// assertAllHidden fails if the fixture did not actually end up all-hidden, so a
// silently ineffective rewrite cannot make the tests below pass vacuously.
func assertAllHidden(t *testing.T, path string) {
	t.Helper()
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = f.Close() }()
	for _, sheet := range f.GetSheetList() {
		visible, err := f.GetSheetVisible(sheet)
		if err != nil {
			t.Fatalf("GetSheetVisible(%s): %v", sheet, err)
		}
		if visible {
			t.Fatalf("fixture sheet %s is still visible; the test would not exercise the case", sheet)
		}
	}
}

func TestEverySheetHiddenStillOffersTheChoice(t *testing.T) {
	// Filtering hidden sheets can remove every candidate. When it does, the
	// filter has deleted the question rather than answered it: several hidden
	// sheets would otherwise fall through to an arbitrary first-sheet default,
	// which is the guess this change exists to remove. Found by review on #907.
	path := filepath.Join(t.TempDir(), "allhidden.xlsx")
	writeSheets(t, path, []string{"A", "B", "C"}, [][][]string{table(), table(), table()})
	hideEverySheet(t, path)
	assertAllHidden(t, path)

	app := &App{}
	err := errFromLoad(app, path)
	if err == nil {
		t.Fatal("loaded one of three hidden sheets arbitrarily instead of asking")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C") {
		t.Errorf("error should still name the sheets, got: %v", err)
	}

	suggestion, err := suggestExcelImport(path)
	if err != nil {
		t.Fatalf("suggestExcelImport: %v", err)
	}
	if !suggestion.NeedsWizard {
		t.Error("refused the file but did not offer the wizard: a dead end")
	}
}

func TestOneHiddenSheetLoadsWithoutAsking(t *testing.T) {
	// The companion control. One sheet is unambiguous whether or not it is
	// hidden, so the fallback above must not turn a hidden sheet into a prompt.
	path := filepath.Join(t.TempDir(), "onehidden.xlsx")
	writeSheets(t, path, []string{"Only"}, [][][]string{table()})
	hideEverySheet(t, path)
	assertAllHidden(t, path)

	app := &App{}
	data, err := app.loadExcel(path)
	if err != nil {
		t.Fatalf("a single hidden sheet should still load: %v", err)
	}
	if data.Rows != 2 {
		t.Errorf("rows = %d, want 2", data.Rows)
	}
}

func errFromLoad(app *App, path string) error {
	_, err := app.loadExcel(path)
	return err
}
