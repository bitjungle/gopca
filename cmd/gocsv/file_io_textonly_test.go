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
)

// TestParseCSVContentAcceptsTextOnly covers #801: GoCSV refused to open any CSV
// with no numeric column, reporting "no data found in file".
//
// The parser was never the problem -- it returns the text columns as
// categorical with no error. The load gate tested the numeric column count, so
// a perfectly well-formed file was rejected and the user was told something
// untrue about it. Preparing data for PCA regularly starts from a file that is
// not numeric yet; with one-hot and ordinal encoding available, such a file is
// a supported starting point rather than a dead end.
func TestParseCSVContentAcceptsTextOnly(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantHeaders []string
		wantTypes   map[string]string
	}{
		{
			name:        "text only",
			content:     "Name,City,Country\nAnn,Oslo,NO\nBo,Bergen,SE\n",
			wantHeaders: []string{"City", "Country"},
			wantTypes:   map[string]string{"City": "categorical", "Country": "categorical"},
		},
		{
			name:        "text with one numeric column still works",
			content:     "Name,City,Score\nAnn,Oslo,10\nBo,Bergen,12\n",
			wantHeaders: []string{"City", "Score"},
			wantTypes:   map[string]string{"City": "categorical", "Score": "numeric"},
		},
		{
			name:        "all numeric unchanged",
			content:     "Name,A,B\nr1,1,2\nr2,3,4\n",
			wantHeaders: []string{"A", "B"},
			wantTypes:   map[string]string{"A": "numeric", "B": "numeric"},
		},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := app.parseCSVContent(tt.content, ".csv")
			if err != nil {
				t.Fatalf("parseCSVContent: %v", err)
			}
			if len(data.Headers) != len(tt.wantHeaders) {
				t.Fatalf("headers = %v, want %v", data.Headers, tt.wantHeaders)
			}
			for i, want := range tt.wantHeaders {
				if data.Headers[i] != want {
					t.Errorf("header %d = %q, want %q (full: %v)", i, data.Headers[i], want, data.Headers)
				}
			}
			for column, want := range tt.wantTypes {
				if got := data.ColumnTypes[column]; got != want {
					t.Errorf("column %q typed %q, want %q", column, got, want)
				}
			}

			// The cells must arrive, not just the headers. A load that reports
			// the right shape and an empty grid is the same failure one layer on.
			if data.Rows != 2 {
				t.Errorf("rows = %d, want 2", data.Rows)
			}
			for i, row := range data.Data {
				if len(row) != len(data.Headers) {
					t.Errorf("row %d has %d cells, want %d: %v",
						i, len(row), len(data.Headers), row)
				}
				for j, cell := range row {
					if strings.TrimSpace(cell) == "" {
						t.Errorf("row %d column %d (%s) is empty: %v",
							i, j, data.Headers[j], row)
					}
				}
			}
		})
	}
}

// TestParseCSVContentDelimiterFallback is the regression guard for the change
// that made the above pass.
//
// The numeric-column test being relaxed was doing two jobs: deciding whether
// the file was usable, and deciding whether the delimiter was right. Accepting
// categorical-only parses risks settling on the first format that yields
// anything at all, which for a semicolon file would be the comma attempt.
//
// It does not, for two reasons this test pins down: a wrong delimiter either
// makes the field counts inconsistent (an error), or collapses every line to
// one field that is then consumed as row names, leaving no columns behind.
// Both send the loop on to the next format.
func TestParseCSVContentDelimiterFallback(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantHeaders []string
		why         string
	}{
		{
			name:        "semicolon with decimal comma",
			content:     "Name;City;Score\nAnn;Oslo;10,5\nBo;Bergen;12,25\n",
			wantHeaders: []string{"City", "Score"},
			why:         "the comma attempt splits rows inconsistently and errors",
		},
		{
			name:        "semicolon text only",
			content:     "Name;City;Country\nAnn;Oslo;NO\nBo;Bergen;SE\n",
			wantHeaders: []string{"City", "Country"},
			why: "the comma attempt yields one field per line, consumed as row " +
				"names, so no column survives to be mistaken for a good parse",
		},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := app.parseCSVContent(tt.content, ".csv")
			if err != nil {
				t.Fatalf("parseCSVContent: %v", err)
			}
			if len(data.Headers) != len(tt.wantHeaders) {
				t.Fatalf("headers = %v, want %v -- the semicolon format should have "+
					"won because %s", data.Headers, tt.wantHeaders, tt.why)
			}
			for i, want := range tt.wantHeaders {
				if data.Headers[i] != want {
					t.Errorf("header %d = %q, want %q (%s)", i, data.Headers[i], want, tt.why)
				}
			}
		})
	}
}

// TestParseCSVContentStillRejectsEmpty checks the gate did not become a
// rubber stamp. Relaxing an acceptance test is only safe if something is still
// refused.
func TestParseCSVContentStillRejectsEmpty(t *testing.T) {
	app := &App{}
	for _, content := range []string{"", "\n", "   \n\n"} {
		if _, err := app.parseCSVContent(content, ".csv"); err == nil {
			t.Errorf("parseCSVContent(%q) succeeded; empty input must still be refused", content)
		}
	}
}

// TestTextOnlyFileIsUsableAfterLoading checks what happens once such a file is
// open, which is the half of #801 that opening it only makes reachable.
//
// The issue argued the existing validation would report the situation clearly.
// That is a claim about behaviour, so it is checked rather than assumed: the
// file must load, must be refused for PCA, and the refusal must say why.
func TestTextOnlyFileIsUsableAfterLoading(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent("Name,City,Country\nAnn,Oslo,NO\nBo,Bergen,SE\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}

	result := app.ValidateForGoPCA(data)
	if result.IsValid {
		t.Error("a file with no numeric column should not validate for PCA")
	}
	if len(result.Messages) == 0 {
		t.Fatal("the refusal must explain itself; there were no messages")
	}
	joined := strings.ToLower(strings.Join(result.Messages, " "))
	if !strings.Contains(joined, "numeric") {
		t.Errorf("the message should say a numeric column is what is missing, got %v",
			result.Messages)
	}

	// And an encoder must be offered on it -- opening the file is only useful
	// because there is a way to make it numeric. (Ordinal encoding joins one-hot
	// here via #856; this branch checks the encoder that exists on develop.)
	offered := app.GetTransformableColumns(data, TransformOneHot)
	if len(offered) != 2 {
		t.Errorf("expected both text columns to be offered for encoding, got %v", offered)
	}
}

// TestTextOnlyExcelSheetLoads is the complement to TestIssue799_TitleBlockIsReportedActionably,
// and the two only make sense read together.
//
// loadExcel distinguishes a sheet that is all text because it genuinely is
// from one that is all text because a title block is being read as the header.
// The discriminator is a preamble of narrow leading rows, not the absence of
// numbers -- so this sheet, which has no preamble, must open.
//
// Without this test the obvious "fix" for any future trouble in that branch is
// to require a numeric column outright, which would restore #801 for
// spreadsheets while leaving every CSV test passing.
func TestTextOnlyExcelSheetLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "textonly.xlsx")
	setCells(t, path, [][]string{
		{"Name", "City", "Country"},
		{"Ann", "Oslo", "NO"},
		{"Bo", "Bergen", "SE"},
	})

	data, err := NewApp().loadExcel(path)
	if err != nil {
		t.Fatalf("a text-only sheet with no title block must open: %v", err)
	}
	if data.Rows != 2 {
		t.Errorf("rows = %d, want 2", data.Rows)
	}
	for _, header := range []string{"City", "Country"} {
		if data.ColumnTypes[header] != "categorical" {
			t.Errorf("column %q typed %q, want categorical", header, data.ColumnTypes[header])
		}
	}
}

// TestTextOnlyFileBecomesPCAReadyByEncoding follows the whole point of #801:
// a text-only CSV is a valid starting point because there is now a way to make
// it numeric.
//
// The three steps are checked together because each is uninteresting alone.
// Loading a file nothing can be done with is not a fix; encoding a file that
// cannot be opened is unreachable. This asserts the arc: the file opens, is
// correctly refused for PCA, and after encoding is accepted.
func TestTextOnlyFileBecomesPCAReadyByEncoding(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent(
		"Prove,By,Kvalitet\nP1,Oslo,høy\nP2,Bergen,lav\nP3,Trondheim,middels\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}

	if before := app.ValidateForGoPCA(data); before.IsValid {
		t.Fatal("a text-only file should not be PCA-ready before encoding")
	}

	res, err := app.applyTransformationInternal(data, TransformOptions{
		Type:    TransformOneHot,
		Columns: []string{"Kvalitet"},
	})
	if err != nil {
		t.Fatalf("one-hot encoding a column of the loaded file: %v", err)
	}
	if res.Data == nil {
		t.Fatal("the transformation returned no data")
	}

	after := app.ValidateForGoPCA(res.Data)
	if !after.IsValid {
		t.Errorf("after encoding, the file should be valid for PCA; validation said %v",
			after.Messages)
	}
}
