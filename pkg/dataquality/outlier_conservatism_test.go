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
	if len(farOut.Outliers) == 1 && farOut.Outliers[0].Method != "detached" {
		t.Errorf("method = %q, want detached", farOut.Outliers[0].Method)
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

// The three tests below are the argument for requiring two conditions. Each
// shows one of them refusing a case the other would have flagged, so removing
// either makes a named test fail rather than merely changing a number.

// TestTopOfATailIsNotAnOutlier is condition 1 doing the refusing.
//
// Modelled on Ti in the aluminium alloy dataset, where the report picked a
// value 0.0001 past the far-out fence that sat just above two dozen identical
// ones and called it an outlier. Zero inflation crushes the IQR, which drags
// the fence down inside the populated range, so the largest value of a smooth
// tail clears it while being no more remarkable than its neighbours.
func TestTopOfATailIsNotAnOutlier(t *testing.T) {
	// Q1=0, Q3=2, so the fence sits at 8 and the values 9 and 10 are beyond it.
	// The widest gap anywhere is 1, a tenth of the range: nothing is detached.
	values := repeatVals("0", 50)
	for _, v := range []string{"1", "2", "3"} {
		values = append(values, repeatVals(v, 10)...)
	}
	values = append(values, "4", "5", "6", "7", "8", "9", "10")

	col := numericColumnOf(t, "tail", values...)
	if *col.Stats.Q3+3*(*col.Stats.IQR) >= 10 {
		t.Fatalf("fixture no longer puts values beyond the fence (Q3=%v IQR=%v); test has gone stale",
			*col.Stats.Q3, *col.Stats.IQR)
	}
	if len(col.Outliers) != 0 {
		t.Errorf("the top of a continuous tail must not be flagged; got %d outliers", len(col.Outliers))
	}
}

// TestAbsentOrPresentIsNotAnOutlier is condition 2 doing the refusing.
//
// A column holding two values is detached by construction -- the only gap is
// the whole range -- so the gap test alone flags the smaller group. On the
// aluminium alloy dataset that meant the non-zero values of sparse columns like
// B, Cd and V: the element is simply absent from most alloys, and its presence
// is a measurement rather than an anomaly.
func TestAbsentOrPresentIsNotAnOutlier(t *testing.T) {
	col := numericColumnOf(t, "present", append(repeatVals("0", 400), repeatVals("10", 600)...)...)
	if len(col.Outliers) != 0 {
		t.Errorf("a two-valued column must not be flagged; got %d outliers", len(col.Outliers))
	}
}

// TestObviouslyWrongValueIsFlagged is the case both conditions agree on, and
// the reason any of this exists.
//
// Modelled on eeg_eye_state, where single sensor readings of 309231 and 715897
// sit among values of a few thousand, and on met.csv, where a dew point of 100
// appears where the next largest is 19 -- a sentinel left in the data.
func TestObviouslyWrongValueIsFlagged(t *testing.T) {
	values := make([]string, 0, 201)
	for i := 0; i < 200; i++ {
		values = append(values, strconv.Itoa(95+i%11))
	}
	values = append(values, "50000")

	col := numericColumnOf(t, "sensor", values...)
	if len(col.Outliers) != 1 {
		t.Fatalf("got %d outliers, want exactly 1: 50000 among values near 100", len(col.Outliers))
	}
	if col.Outliers[0].Value != "50000" {
		t.Errorf("flagged %q, want 50000", col.Outliers[0].Value)
	}
	if col.Outliers[0].RowIndex != 200 {
		t.Errorf("RowIndex = %d, want 200: the outlier keeps its position in the file",
			col.Outliers[0].RowIndex)
	}
	if !hasReportableOutliers(col) {
		t.Errorf("1 in 201 (%.2f%%) must be reportable", outlierShare(col))
	}
}

// TestGapThresholdKeepsTheQualifyingGapUnique guards the reason the threshold
// is one half rather than some smaller number that also looks conservative.
//
// Two gaps each spanning more than half the range cannot both exist, so at this
// threshold the widest gap is the only one that can qualify. That gives two
// properties the behavioural tests above do not pin, because they pass for any
// threshold between roughly 0.1 and 0.6: the chosen gap is unique rather than
// an arbitrary pick among near-ties, and removing a true outlier cannot shrink
// the range enough to make an ordinary neighbour look detached in turn.
func TestGapThresholdKeepsTheQualifyingGapUnique(t *testing.T) {
	if minGapShare < 0.5 {
		t.Errorf("minGapShare = %v; below 0.5 more than one gap can qualify, so the "+
			"widest is an arbitrary choice and the rule can cascade", minGapShare)
	}
}
