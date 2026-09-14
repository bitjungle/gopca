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

package dataquality

import (
	"strconv"
	"strings"
	"testing"
)

// numericColumnOf builds a column from raw values, running the same statistics
// and detection the report uses, so these tests exercise the real pipeline
// rather than a hand-filled ColumnAnalysis that could drift from it.
func numericColumnOf(t *testing.T, name string, values ...string) ColumnAnalysis {
	t.Helper()
	data := make([][]string, len(values))
	for i, v := range values {
		data[i] = []string{v}
	}
	stats := analyzeNumericStats(data, len(data), 0)
	return ColumnAnalysis{
		Name:     name,
		Type:     "numeric",
		Stats:    stats,
		Outliers: detectOutliers(data, len(data), 0, stats),
	}
}

func repeatVals(v string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = v
	}
	return out
}

// nearHundred returns 200 values spread between 95 and 105.
func nearHundred() []string {
	out := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
		out = append(out, strconv.Itoa(95+i%11))
	}
	return out
}

// TestHundredfoldJumpIsFlagged is the case the check exists for: a value two
// orders of magnitude past everything else, which is what a misplaced decimal
// point or a sentinel looks like.
func TestHundredfoldJumpIsFlagged(t *testing.T) {
	col := numericColumnOf(t, "sensor", append(nearHundred(), "50000")...)
	if len(col.Outliers) != 1 {
		t.Fatalf("got %d outliers, want 1: 50000 is ~476x the next value", len(col.Outliers))
	}
	if col.Outliers[0].Value != "50000" || col.Outliers[0].RowIndex != 200 {
		t.Errorf("flagged %q at row %d, want 50000 at row 200",
			col.Outliers[0].Value, col.Outliers[0].RowIndex)
	}
	if !hasReportableOutliers(col) {
		t.Errorf("1 in 201 must be reportable")
	}
}

// TestTenfoldJumpIsNotFlagged pins the threshold from below. A value ten times
// the rest is conspicuous, and might well be an error -- but it is also what a
// legitimately different sample looks like, so GoCSV says nothing.
func TestTenfoldJumpIsNotFlagged(t *testing.T) {
	col := numericColumnOf(t, "sensor", append(nearHundred(), "1000")...)
	if len(col.Outliers) != 0 {
		t.Errorf("a tenfold jump must not be flagged; got %d outliers", len(col.Outliers))
	}
}

// TestAlloyingElementIsNotFlagged is the case that drove this rule, kept as a
// test because it is the one a statistical fence gets wrong every time.
//
// Rows 570 and 571 of the aluminium alloy dataset are an Al-4Cr-1Fe alloy:
// 4.11% chromium where most aluminium alloys hold none. Every robust fence
// flags it, and the value is the defining property of the material and the most
// correct number in the row. At roughly twelve times the next value it is
// nowhere near a hundredfold, so the magnitude rule leaves it alone.
func TestAlloyingElementIsNotFlagged(t *testing.T) {
	values := repeatVals("0", 900)
	for _, v := range []string{"0.0005", "0.001", "0.002", "0.003", "0.0035"} {
		values = append(values, repeatVals(v, 50)...)
	}
	values = append(values, "0.0411", "0.0411")

	col := numericColumnOf(t, "Cr", values...)
	if len(col.Outliers) != 0 {
		t.Errorf("an alloying element at ~12x the next value must not be flagged; got %d outliers",
			len(col.Outliers))
	}
}

// TestSmoothDecadesAreNotFlagged checks that a column spanning several orders
// of magnitude is safe. The ratio is measured against the neighbouring value,
// not the middle of the data, so a value can be far from the median and still
// be close to its neighbour.
func TestSmoothDecadesAreNotFlagged(t *testing.T) {
	col := numericColumnOf(t, "conc",
		"0.001", "0.003", "0.01", "0.03", "0.1", "0.3", "1", "3", "10", "30", "100", "300", "1000")
	if len(col.Outliers) != 0 {
		t.Errorf("a column spanning six decades smoothly must not be flagged; got %d outliers",
			len(col.Outliers))
	}
}

// TestZeroNeighbourIsNotAYardstick guards the sparse-column case. Everything is
// infinitely larger than nothing, so a column that is mostly zeros would report
// its only real measurements as errors.
func TestZeroNeighbourIsNotAYardstick(t *testing.T) {
	col := numericColumnOf(t, "trace", append(repeatVals("0", 500), repeatVals("0.0001", 20)...)...)
	if len(col.Outliers) != 0 {
		t.Errorf("values next to a zero neighbour must not be flagged; got %d outliers",
			len(col.Outliers))
	}
}

// TestRepeatedGlitchIsNotHiddenByItsOwnRepetition checks the tie handling. Two
// identical extreme readings would otherwise be compared against each other,
// giving a ratio of one and concealing the jump beneath them.
func TestRepeatedGlitchIsNotHiddenByItsOwnRepetition(t *testing.T) {
	col := numericColumnOf(t, "sensor", append(nearHundred(), "50000", "50000")...)
	if len(col.Outliers) != 2 {
		t.Errorf("got %d outliers, want 2: identical extremes are taken together", len(col.Outliers))
	}
}

// TestNegativeSentinelIsFlagged checks the lower end, where the classic case is
// a negative sentinel standing in for a missing reading.
func TestNegativeSentinelIsFlagged(t *testing.T) {
	col := numericColumnOf(t, "depth", append([]string{"-9999"}, nearHundred()...)...)
	if len(col.Outliers) != 1 {
		t.Fatalf("got %d outliers, want 1: -9999 against values near 100", len(col.Outliers))
	}
	if col.Outliers[0].Value != "-9999" {
		t.Errorf("flagged %q, want -9999", col.Outliers[0].Value)
	}
}

// TestSmallFileCanStillReport covers the absolute alternative in the reporting
// gate. One value in forty is 2.5% of the column, so a share ceiling alone
// would silence every file shorter than a hundred rows.
func TestSmallFileCanStillReport(t *testing.T) {
	values := make([]string, 0, 40)
	for i := 0; i < 39; i++ {
		values = append(values, strconv.Itoa(100+i%5))
	}
	values = append(values, "99999")

	col := numericColumnOf(t, "small", values...)
	if len(col.Outliers) != 1 {
		t.Fatalf("got %d outliers, want 1", len(col.Outliers))
	}
	if outlierShare(col) <= 1.0 {
		t.Fatalf("share is %.2f%%, so this no longer tests the absolute alternative",
			outlierShare(col))
	}
	if !hasReportableOutliers(col) {
		t.Errorf("1 flagged value in a 40-row file must be reportable, share %.2f%%",
			outlierShare(col))
	}
}

// TestOutlierIssueExplainsTheLikelyCause keeps the message useful: a jump this
// large has a short list of mundane explanations, and the real judgement of
// which samples are unusual happens in GoPCA.
func TestOutlierIssueExplainsTheLikelyCause(t *testing.T) {
	col := numericColumnOf(t, "sensor", append(nearHundred(), "50000")...)

	issue, found := issueOfCategory(generateQualityIssues(
		&DataQualityReport{ColumnAnalysis: []ColumnAnalysis{col}}, nil, nil), "outlier")
	if !found {
		t.Fatal("no outlier issue raised")
	}
	for _, want := range []string{"decimal", "unit", "sentinel", "GoPCA"} {
		if !strings.Contains(issue.Impact, want) {
			t.Errorf("impact does not mention %q: %q", want, issue.Impact)
		}
	}
	if issue.Severity != "info" {
		t.Errorf("severity = %q, want info", issue.Severity)
	}
}
