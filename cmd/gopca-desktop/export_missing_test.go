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
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIssue739_ExportPreservesMissingValuesWithRowNames verifies that
// exportDataToCSV writes a blank cell (not a raw NaN/0 sentinel) for missing
// values when row names are present. The old code computed the data-column index
// as len(rowStrings) - len(data.RowNames) (total rows) instead of the row-name
// prefix width, so the missing-value branch was mis-aligned. Regression for #739.
func TestIssue739_ExportPreservesMissingValuesWithRowNames(t *testing.T) {
	app := &App{}
	data := &FileData{
		Headers:  []string{"A", "B", "C"},
		RowNames: []string{"r1", "r2"},
		Data: [][]float64{
			{1.0, math.NaN(), 3.0}, // B missing in row r1
			{4.0, 5.0, 6.0},
		},
		MissingMask: [][]bool{
			{false, true, false},
			{false, false, false},
		},
	}

	path := filepath.Join(t.TempDir(), "out.csv")
	if err := app.exportDataToCSV(data, path); err != nil {
		t.Fatalf("export failed: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	got := string(content)

	if strings.Contains(got, "NaN") {
		t.Errorf("export wrote a raw NaN sentinel instead of a blank for a missing value:\n%s", got)
	}

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header + 2 data rows, got %d lines:\n%s", len(lines), got)
	}
	// r1: B missing -> blank; A and C must be intact (the bug also blanked/garbled C).
	if lines[1] != "r1,1,,3" {
		t.Errorf("row r1 = %q, want %q (B blank, A/C intact)", lines[1], "r1,1,,3")
	}
	if lines[2] != "r2,4,5,6" {
		t.Errorf("row r2 = %q, want %q", lines[2], "r2,4,5,6")
	}
}
