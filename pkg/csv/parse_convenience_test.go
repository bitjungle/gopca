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

package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── Parse (convenience wrapper) ─────────────────────────────────────────────

func TestParse_NumericCSV(t *testing.T) {
	input := "A,B,C\n1,2,3\n4,5,6\n"
	opts := DefaultOptions()
	opts.HasRowNames = false

	data, err := Parse(strings.NewReader(input), opts)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.Rows != 2 {
		t.Errorf("Rows: got %d, want 2", data.Rows)
	}
	if data.Columns != 3 {
		t.Errorf("Columns: got %d, want 3", data.Columns)
	}
	if len(data.Headers) != 3 || data.Headers[0] != "A" {
		t.Errorf("Headers: got %v", data.Headers)
	}
}

func TestParse_WithRowNames(t *testing.T) {
	input := "\"\",X,Y\nrow1,10,20\nrow2,30,40\n"
	opts := DefaultOptions()

	data, err := Parse(strings.NewReader(input), opts)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(data.RowNames) != 2 || data.RowNames[0] != "row1" {
		t.Errorf("RowNames: got %v", data.RowNames)
	}
}

func TestParse_EmptyReader(t *testing.T) {
	opts := DefaultOptions()
	_, err := Parse(strings.NewReader(""), opts)
	// Empty content should produce an error (no data)
	if err == nil {
		t.Error("expected error for empty reader, got nil")
	}
}

func TestParse_TabDelimited(t *testing.T) {
	input := "col1\tcol2\n1\t2\n3\t4\n"
	opts := TabDelimitedOptions()
	opts.HasRowNames = false

	data, err := Parse(strings.NewReader(input), opts)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.Rows != 2 || data.Columns != 2 {
		t.Errorf("got %dx%d, want 2x2", data.Rows, data.Columns)
	}
}

// ─── ParseFile (convenience wrapper) ─────────────────────────────────────────

func TestParseFile_ValidCSV(t *testing.T) {
	content := "A,B\n1,2\n3,4\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := DefaultOptions()
	opts.HasRowNames = false

	data, err := ParseFile(path, opts)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if data.Rows != 2 || data.Columns != 2 {
		t.Errorf("got %dx%d, want 2x2", data.Rows, data.Columns)
	}
}

func TestParseFile_NonexistentFile(t *testing.T) {
	// Use a path under t.TempDir() that is guaranteed never to be created,
	// avoiding reliance on a hard-coded path that may exist on some hosts.
	path := filepath.Join(t.TempDir(), "does-not-exist.csv")
	opts := DefaultOptions()
	_, err := ParseFile(path, opts)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
