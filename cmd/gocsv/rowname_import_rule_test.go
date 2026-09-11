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

import "testing"

// The loader takes the first column as row names unconditionally. These tests
// pin the rule that decides whether it keeps them, end to end through
// parseCSVContent -- the single path CSV, TSV and Excel all reach, since
// loadExcel converts the sheet to CSV and calls it (#904).

func TestImportKeepsRowNamesWhenTheyIdentifyRows(t *testing.T) {
	// The control. Without this, a demotion that fired unconditionally would
	// satisfy every other test in this file.
	app := &App{}
	data, err := app.parseCSVContent("SampleID,A,B\nS1,1,2\nS2,3,4\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if len(data.RowNames) != 2 {
		t.Fatalf("row names dropped: got %d, want 2 (%v)", len(data.RowNames), data.RowNames)
	}
	if data.RowNamesHeader != "SampleID" {
		t.Errorf("RowNamesHeader = %q, want %q", data.RowNamesHeader, "SampleID")
	}
	if got := data.Headers; len(got) != 2 || got[0] != "A" {
		t.Errorf("headers = %v, want the id column held out as row names", got)
	}
}

func TestImportKeepsNonIdentifyingFirstColumnAsData(t *testing.T) {
	tests := []struct {
		name    string
		content string
		reason  string
	}{
		{
			name:    "repeated values",
			content: "Name,A,B\nPaper one,1,2\nPaper one,3,4\nPaper two,5,6\n",
			reason:  "a repeated value cannot identify two different rows",
		},
		{
			name:    "a blank value",
			content: "Name,A,B\nP1,1,2\n,3,4\nP3,5,6\n",
			reason:  "a blank leaves its row unlabelled",
		},
		{
			name:    "values differing only by surrounding space",
			content: "Name,A,B\nP1,1,2\nP1 ,3,4\nP3,5,6\n",
			reason:  "they render identically as a plot label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			data, err := app.parseCSVContent(tt.content, ".csv")
			if err != nil {
				t.Fatalf("parseCSVContent: %v", err)
			}
			if len(data.RowNames) != 0 {
				t.Errorf("row names were assigned anyway (%v): %s", data.RowNames, tt.reason)
			}
			if data.RowNamesHeader != "" {
				t.Errorf("RowNamesHeader = %q, want empty", data.RowNamesHeader)
			}
			// Demoting must not lose the column. It carries real data.
			if len(data.Headers) == 0 || data.Headers[0] != "Name" {
				t.Fatalf("headers = %v, want the column back at index 0 under its own name", data.Headers)
			}
			if data.Columns != len(data.Headers) {
				t.Errorf("Columns = %d, want %d", data.Columns, len(data.Headers))
			}
		})
	}
}

func TestImportDemotionKeepsEveryValue(t *testing.T) {
	// The values have to survive the move, in order, or the column is worse
	// than useless -- it would silently misalign with its own rows.
	app := &App{}
	data, err := app.parseCSVContent("Name,A\nrepeat,1\nrepeat,2\nunique,3\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	want := []string{"repeat", "repeat", "unique"}
	if len(data.Data) != len(want) {
		t.Fatalf("got %d rows, want %d", len(data.Data), len(want))
	}
	for i, w := range want {
		if got := data.Data[i][0]; got != w {
			t.Errorf("row %d column 0 = %q, want %q", i, got, w)
		}
	}
}

func TestImportDemotionNamesAnUnnamedColumn(t *testing.T) {
	// A blank header is a common CSV convention for the row-name column, but a
	// blank column header in the grid cannot be referred to anywhere.
	app := &App{}
	data, err := app.parseCSVContent(",A,B\ndup,1,2\ndup,3,4\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if len(data.RowNames) != 0 {
		t.Fatalf("row names were assigned anyway: %v", data.RowNames)
	}
	if data.Headers[0] != "RowName" {
		t.Errorf("headers[0] = %q, want %q", data.Headers[0], "RowName")
	}
}

func TestImportDemotionAvoidsHeaderCollision(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent("Name,Name,B\ndup,x,2\ndup,y,4\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if len(data.Headers) < 2 {
		t.Fatalf("headers = %v", data.Headers)
	}
	if data.Headers[0] == data.Headers[1] {
		t.Errorf("headers collide: %v", data.Headers)
	}
}

func TestDemoteNonIdentifyingRowNamesIsANoOpWithoutRowNames(t *testing.T) {
	data := &FileData{Headers: []string{"A", "B"}, Data: [][]string{{"1", "2"}}, Columns: 2}
	if demoteNonIdentifyingRowNames(data) {
		t.Error("reported a demotion on a file that has no row names")
	}
	if len(data.Headers) != 2 {
		t.Errorf("headers changed: %v", data.Headers)
	}
}
