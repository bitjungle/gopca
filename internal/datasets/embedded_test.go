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

package datasets

import (
	"strings"
	"testing"
)

func TestGetDataset(t *testing.T) {
	// All embedded datasets must decompress to non-empty CSV content.
	validFiles := []string{
		"corn.csv",
		"iris.csv",
		"wine.csv",
		"swiss_roll.csv",
		"eeg_eye_state.csv",
		"cstr.csv",
	}
	for _, name := range validFiles {
		t.Run(name, func(t *testing.T) {
			content, ok := GetDataset(name)
			if !ok {
				t.Fatalf("GetDataset(%q) returned ok=false, want true", name)
			}
			if content == "" {
				t.Errorf("GetDataset(%q) returned empty content", name)
			}
			// Decompressed content should look like CSV: at least one row
			// separator and one field separator.
			if !strings.Contains(content, "\n") {
				t.Errorf("GetDataset(%q) content has no newline, not CSV-like", name)
			}
		})
	}
}

func TestGetDatasetUnknown(t *testing.T) {
	content, ok := GetDataset("does_not_exist.csv")
	if ok {
		t.Error("GetDataset(unknown) returned ok=true, want false")
	}
	if content != "" {
		t.Errorf("GetDataset(unknown) returned %q, want empty string", content)
	}
}
