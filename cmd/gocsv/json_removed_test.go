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
)

// TestIssue719_JSONImportRemoved verifies the stubbed JSON import path is gone:
// .json files are no longer advertised/detected as an importable "json" format,
// and PreviewFile/ImportFile no longer route into the "JSON import not yet
// implemented" dead end. Regression test for #719.
func TestIssue719_JSONImportRemoved(t *testing.T) {
	app := NewApp()
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "data.json")
	if err := os.WriteFile(jsonPath, []byte(`[{"a":1}]`), 0644); err != nil {
		t.Fatalf("write temp json: %v", err)
	}

	// A .json file must be rejected explicitly, not silently classified as an
	// importable format (e.g. CSV). GetFileInfo returns an error for .json.
	if _, err := app.GetFileInfo(jsonPath); err == nil {
		t.Fatal("GetFileInfo(.json): expected an error (JSON is not importable), got nil")
	} else if strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("GetFileInfo(.json): stale stub error, got %v", err)
	}

	// The old stub returned "JSON import not yet implemented"; that dead end must be
	// gone — an explicit json format now yields a clean unsupported-format error.
	if _, err := app.PreviewFile(jsonPath, ImportOptions{Format: "json"}); err == nil ||
		strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("PreviewFile(json): expected an unsupported-format error, got %v", err)
	}
	if _, err := app.ImportFile(jsonPath, ImportOptions{Format: "json"}); err == nil ||
		strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("ImportFile(json): expected an unsupported-format error, got %v", err)
	}

	// JSON content under a non-.json extension must not be misclassified as an
	// importable format; it is recognized as the non-importable "json" format.
	txtPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(txtPath, []byte(`[{"a":1}]`), 0644); err != nil {
		t.Fatalf("write temp txt: %v", err)
	}
	if f := app.detectFileFormat(txtPath); f == "csv" || f == "tsv" || f == "excel" {
		t.Errorf("JSON content classified as importable format %q; should be rejected", f)
	}
}
