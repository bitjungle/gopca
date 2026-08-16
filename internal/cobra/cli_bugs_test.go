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

package cobra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// TestIssue740_ValidateEmptyDelimiterReturnsError verifies that `validate` routes
// --delimiter through parseDelimiter. Previously it did `rune(opts.Delimiter[0])`,
// which panicked (index out of range) on an empty delimiter; it must now return a
// clean error instead. Regression test for #740.
func TestIssue740_ValidateEmptyDelimiterReturnsError(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(csvPath, []byte("a,b,c\n1,2,3\n4,5,6\n"), 0o644); err != nil {
		t.Fatalf("write temp csv: %v", err)
	}

	err := runValidate(&ValidateOptions{Delimiter: ""}, csvPath)
	if err == nil {
		t.Fatal("expected an error for empty --delimiter, got nil")
	}
	if !strings.Contains(err.Error(), "delimiter") {
		t.Errorf("expected a delimiter-related error, got: %v", err)
	}
}

// TestIssue740_GetDataSummaryTruncatesHeaders verifies getDataSummary actually
// prints only the first five column names when it claims to. Previously it joined
// all headers while appending "(showing first 5 of N)". Regression test for #740.
func TestIssue740_GetDataSummaryTruncatesHeaders(t *testing.T) {
	data := &pkgcsv.Data{
		Headers: []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8"},
		Rows:    1,
		Columns: 8,
		Matrix:  types.Matrix{{1, 2, 3, 4, 5, 6, 7, 8}},
	}

	out := getDataSummary(data)

	if !strings.Contains(out, "(showing first 5 of 8)") {
		t.Errorf("expected the truncation note, got:\n%s", out)
	}
	if !strings.Contains(out, "c1, c2, c3, c4, c5") {
		t.Errorf("expected the first five column names, got:\n%s", out)
	}
	// The message claims "first 5", so later names must not appear in the column list.
	for _, late := range []string{"c6", "c7", "c8"} {
		if strings.Contains(out, late) {
			t.Errorf("column list should be truncated to first 5 but included %q:\n%s", late, out)
		}
	}
}
