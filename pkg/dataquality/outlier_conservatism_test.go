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

// TestFenceIsFarOutNotOutside is the substance of the change.
//
// Tukey separates "outside" values beyond 1.5*IQR from "far out" values beyond
// 3*IQR. GoCSV reports only the latter, because the narrower fence flags a
// large share of any skewed column -- it was calling 21% of a composition
// column outliers. A value between the two fences must not be reported.
func TestFenceIsFarOutNotOutside(t *testing.T) {
	// 1..9 gives Q1=3, Q3=7, IQR=4: outside beyond 13, far out beyond 19.
	base := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

	outside := numericColumnOf(t, "x", append(append([]string{}, base...), "15")...)
	if len(outside.Outliers) != 0 {
		t.Errorf("15 is beyond 1.5*IQR but not 3*IQR and must not be flagged; got %d outliers",
			len(outside.Outliers))
	}

	farOut := numericColumnOf(t, "x", append(append([]string{}, base...), "40")...)
	if len(farOut.Outliers) != 1 {
		t.Errorf("40 is beyond the far-out fence and must be flagged; got %d outliers",
			len(farOut.Outliers))
	}
	if len(farOut.Outliers) == 1 && farOut.Outliers[0].Method != "far-out" {
		t.Errorf("method = %q, want far-out", farOut.Outliers[0].Method)
	}
}

// TestNoRobustSpreadReportsNothing pins the deliberate blind spot. A column
// whose middle half is a single repeated value has no robust scale, so "far"
// has no meaning and detection stays silent rather than guessing.
func TestNoRobustSpreadReportsNothing(t *testing.T) {
	col := numericColumnOf(t, "x", append(repeatVals("5", 9), "9999")...)
	if len(col.Outliers) != 0 {
		t.Errorf("a column with zero IQR must report nothing; got %d outliers", len(col.Outliers))
	}
}

// TestManyExtremeValuesAreNotReportedAsOutliers is the inverted gate, fixed.
//
// The old rule reported a column only when MORE than 10% of it was flagged,
// so it spoke exactly when its own method had broken down and stayed silent
// for the handful of extreme values a user can act on. Both halves are
// asserted here, because fixing one without the other still leaves the report
// misleading.
func TestManyExtremeValuesAreNotReportedAsOutliers(t *testing.T) {
	// A fifth of the column sits far from a tight core: a shape, not outliers.
	// Q1=1 and Q3=2, so the IQR is real and the far-out fence sits at 5 --
	// the twenty values of 100 are genuinely beyond it, and there are far too
	// many of them to be called outliers. This mirrors Zn in the aluminium
	// alloy dataset, where 19% of the column lies beyond the fence.
	values := append(repeatVals("1", 60), append(repeatVals("2", 20), repeatVals("100", 20)...)...)
	many := numericColumnOf(t, "wide", values...)
	if len(many.Outliers) == 0 {
		t.Fatal("fixture no longer produces extreme values; the test has gone stale")
	}
	if hasReportableOutliers(many) {
		t.Errorf("%.0f%% of the column is beyond the fence and must not be reported as outliers",
			outlierShare(many))
	}

	// Three in a thousand: rare enough to be outliers, and the case that was
	// silently dropped before.
	few := numericColumnOf(t, "narrow",
		append(repeatVals("1", 499), append(repeatVals("2", 498), "900", "901", "902")...)...)
	if !hasReportableOutliers(few) {
		t.Errorf("%d extreme values in %d (%.2f%%) must be reported",
			len(few.Outliers), few.Stats.Count, outlierShare(few))
	}

	issues := generateQualityIssues(&DataQualityReport{ColumnAnalysis: []ColumnAnalysis{many, few}}, nil, nil)
	named := []string{}
	for _, issue := range issues {
		if issue.Category == "outlier" {
			named = append(named, issue.Affected...)
		}
	}
	if strings.Join(named, ",") != "narrow" {
		t.Errorf("outlier issues name %v, want only [narrow]", named)
	}
}

// TestOutlierIssueHandsOffToGoPCA keeps the division of labour in the message.
// GoCSV flags only the obvious; deciding which samples are unusual is
// multivariate work done on the fitted model.
func TestOutlierIssueHandsOffToGoPCA(t *testing.T) {
	few := numericColumnOf(t, "narrow",
		append(repeatVals("1", 499), append(repeatVals("2", 498), "900", "901", "902")...)...)

	issue, found := issueOfCategory(generateQualityIssues(
		&DataQualityReport{ColumnAnalysis: []ColumnAnalysis{few}}, nil, nil), "outlier")
	if !found {
		t.Fatal("no outlier issue raised")
	}
	if !strings.Contains(issue.Impact, "GoPCA") {
		t.Errorf("impact does not point at GoPCA for the real assessment: %q", issue.Impact)
	}
	if issue.Severity != "info" {
		t.Errorf("severity = %q, want info: an extreme value is not a defect", issue.Severity)
	}
}
