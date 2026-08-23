package csv

import (
	"strings"
	"testing"
)

// TestIssue778_DimensionLimitsAreEnforcedAtRead confirms the read path applies
// security.ValidateDataDimensions rather than local helpers restating the same
// constants. The function was previously called from nowhere outside its own
// tests, while pkg/csv re-implemented two of its three checks inline — leaving
// the third, the memory guard on rows x cols, enforced nowhere at all.
func TestIssue778_DimensionLimitsAreEnforcedAtRead(t *testing.T) {
	// A normal file still reads.
	r := NewReader(DefaultOptions())
	data, err := r.Read(strings.NewReader("name,a,b\nr1,1,2\nr2,3,4\n"))
	if err != nil {
		t.Fatalf("a valid file was rejected: %v", err)
	}
	if data.Rows != 2 || data.Columns != 2 {
		t.Errorf("got %dx%d, want 2x2 (first column is row names)", data.Rows, data.Columns)
	}

	// Too many columns is still refused, now via the shared validator.
	wide := strings.Repeat("c,", 10001)
	if _, err := r.Read(strings.NewReader(wide[:len(wide)-1] + "\n")); err == nil {
		t.Error("a file exceeding the column limit was accepted")
	} else if !strings.Contains(err.Error(), "too many columns") {
		t.Errorf("unexpected error for an over-wide file: %v", err)
	}
}
