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
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
)

// TestRowNameHeaderSurvivesRoundTrip is the prerequisite for #859 and a bug in
// its own right.
//
// A file read and written back had its row-name column renamed to nothing:
// "Prove,By,Kvalitet" became ",By,Kvalitet". The name was dropped by the parser
// and no field existed to carry it, so promoting a column to row names could
// never have been undone -- there would be no name to restore.
//
// It went unnoticed because every dataset in this repo's testdata leaves that
// header blank, which is a common CSV convention. Files from instruments and
// LIMS systems routinely name it, and those are the files this loses.
func TestRowNameHeaderSurvivesRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			// Mixed text and numeric, which routes through CombineColumns --
			// a second place the header has to be carried across.
			name:    "named row-name column",
			content: "Prove,By,Score\nP1,Oslo,10\nP2,Bergen,12\n",
			want:    "Prove,By,Score",
		},
		{
			name:    "blank row-name header stays blank",
			content: ",By,Score\nP1,Oslo,10\nP2,Bergen,12\n",
			want:    ",By,Score",
		},
		{
			name:    "numeric data with a named id column",
			content: "SampleID,A,B\nS1,1,2\nS2,3,4\n",
			want:    "SampleID,A,B",
		},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := app.parseCSVContent(tt.content, ".csv")
			if err != nil {
				t.Fatalf("parseCSVContent: %v", err)
			}

			out := filepath.Join(t.TempDir(), "out.csv")
			opts := pkgcsv.DefaultOptions()
			opts.HasHeaders = true
			opts.HasRowNames = len(data.RowNames) > 0
			written := &pkgcsv.Data{
				Headers:        data.Headers,
				RowNames:       data.RowNames,
				RowNamesHeader: data.RowNamesHeader,
				StringData:     data.Data,
				Rows:           data.Rows,
				Columns:        data.Columns,
			}
			if err := pkgcsv.SaveFile(out, written, opts); err != nil {
				t.Fatalf("SaveFile: %v", err)
			}

			raw, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("reading back: %v", err)
			}
			got := strings.SplitN(strings.TrimSpace(string(raw)), "\n", 2)[0]
			got = strings.TrimRight(got, "\r")
			if got != tt.want {
				t.Errorf("header line round-tripped as %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExtractRowNameColumnOnRaggedData covers a crash introduced while making
// the row-name header survive the import wizard.
//
// Two lengths have to be kept apart: the data rows say whether the column
// exists, and the headers say whether it has a name. On a ragged sheet they
// disagree -- excelize's GetRows trims trailing empty cells, so a short header
// row leaves fewer headers than data columns -- and indexing the headers by a
// column number that was validated against the data panics.
//
// The CSV path cannot reach this, because encoding/csv enforces a consistent
// field count and rejects a ragged file first. The Excel path can. Neither
// import function can be called from a test -- both end by emitting a Wails
// event, which is fatal without a live runtime -- which is why the logic was
// lifted into a helper this test can reach.
func TestExtractRowNameColumnOnRaggedData(t *testing.T) {
	// Two headers, four data columns: the shape excelize produces for a sheet
	// whose header row is shorter than its data rows.
	for _, rowNameColumn := range []int{0, 1, 2, 3} {
		fileData := &FileData{Headers: []string{"ID", "A"}}
		data := [][]string{
			{"S1", "1", "2", "3"},
			{"S2", "4", "5", "6"},
		}

		// A panic fails the test by crashing it, which is the guarded behaviour.
		got := extractRowNameColumn(fileData, data, rowNameColumn)

		if len(fileData.RowNames) != 2 {
			t.Errorf("rowNameColumn=%d: got %d row names, want 2",
				rowNameColumn, len(fileData.RowNames))
		}
		for i, row := range got {
			if len(row) != 3 {
				t.Errorf("rowNameColumn=%d row %d: %d cells left, want 3: %v",
					rowNameColumn, i, len(row), row)
			}
		}
	}
}

// TestExtractRowNameColumn covers the ordinary, non-ragged behaviour.
func TestExtractRowNameColumn(t *testing.T) {
	fileData := &FileData{Headers: []string{"Prove", "A", "B"}}
	data := [][]string{{"P1", "1", "2"}, {"P2", "3", "4"}}

	got := extractRowNameColumn(fileData, data, 0)

	if fileData.RowNamesHeader != "Prove" {
		t.Errorf("RowNamesHeader = %q, want Prove", fileData.RowNamesHeader)
	}
	if strings.Join(fileData.RowNames, ",") != "P1,P2" {
		t.Errorf("RowNames = %v", fileData.RowNames)
	}
	if strings.Join(fileData.Headers, ",") != "A,B" {
		t.Errorf("Headers = %v, want [A B]", fileData.Headers)
	}
	if strings.Join(got[0], ",") != "1,2" {
		t.Errorf("row 0 = %v, want [1 2]", got[0])
	}

	// A column index outside the data is a no-op, not a panic.
	untouched := &FileData{Headers: []string{"A"}}
	if out := extractRowNameColumn(untouched, [][]string{{"1"}}, 5); len(out) != 1 {
		t.Errorf("out-of-range index changed the data: %v", out)
	}
	if len(untouched.RowNames) != 0 {
		t.Errorf("out-of-range index invented row names: %v", untouched.RowNames)
	}
	if out := extractRowNameColumn(untouched, [][]string{{"1"}}, -1); len(out) != 1 {
		t.Errorf("negative index changed the data: %v", out)
	}
}

// TestExtractRowNameColumnFirstRowShorter covers a column that exists in the
// data but not in the first row.
//
// excelize's GetRows trims trailing empty cells per row, so a sheet whose first
// data row happens to end early is narrower than the ones below it. Testing the
// column number against data[0] alone would silently ignore a row-name column
// that is genuinely present, doing nothing and reporting nothing.
func TestExtractRowNameColumnFirstRowShorter(t *testing.T) {
	fileData := &FileData{Headers: []string{"A", "B", "C"}}
	data := [][]string{
		{"1", "2"},       // short: no third cell
		{"3", "4", "P2"}, // the row-name column lives here
		{"5", "6", "P3"},
	}

	got := extractRowNameColumn(fileData, data, 2)

	if len(fileData.RowNames) != 3 {
		t.Fatalf("expected 3 row names, got %v -- the column was ignored because "+
			"the first row is shorter than the rest", fileData.RowNames)
	}
	if fileData.RowNames[0] != "" || fileData.RowNames[1] != "P2" {
		t.Errorf("row names = %q, want [\"\" P2 P3]", fileData.RowNames)
	}
	if fileData.RowNamesHeader != "C" {
		t.Errorf("RowNamesHeader = %q, want C", fileData.RowNamesHeader)
	}
	if strings.Join(fileData.Headers, ",") != "A,B" {
		t.Errorf("headers = %v, want [A B]", fileData.Headers)
	}
	// The short row must not have a cell removed it never had.
	if strings.Join(got[0], ",") != "1,2" {
		t.Errorf("short row = %v, want [1 2]", got[0])
	}
}
