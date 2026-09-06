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

func numericCol(name string, mean, stdDev, min, max float64, count int) ColumnAnalysis {
	return ColumnAnalysis{
		Name: name,
		Type: "numeric",
		Stats: ColumnStatistics{
			Count:  count,
			Mean:   &mean,
			StdDev: &stdDev,
			Min:    &min,
			Max:    &max,
		},
	}
}

func varianceIssuesFor(cols ...ColumnAnalysis) []QualityIssue {
	return varianceIssues(&DataQualityReport{ColumnAnalysis: cols})
}

func describe(issues []QualityIssue) string {
	parts := make([]string, 0, len(issues))
	for _, i := range issues {
		parts = append(parts, i.Severity+":"+i.Description)
	}
	return strings.Join(parts, " | ")
}

// TestVarianceIssuesAreScaleFree is the reason the previous check was replaced.
//
// It tested StdDev < 0.01, a threshold on a dimensional quantity. The same
// measurements expressed in different units gave different answers: a distance
// column recorded in kilometres was reported as low variance, and the identical
// data recorded in metres was not. A threshold on a number whose unit the
// software does not know cannot mean anything.
//
// The coefficient of variation is dimensionless, so both must now agree.
func TestVarianceIssuesAreScaleFree(t *testing.T) {
	// 5 metres of real variation about a 2 km level, expressed two ways.
	inKilometres := numericCol("Distance_km", 2.0, 0.005, 1.99, 2.01, 100)
	inMetres := numericCol("Distance_m", 2000.0, 5.0, 1990, 2010, 100)

	km := varianceIssuesFor(inKilometres)
	m := varianceIssuesFor(inMetres)

	if len(km) != len(m) {
		t.Errorf("the same data in different units gave different answers:\n  km: %s\n  m:  %s",
			describe(km), describe(m))
	}
	// Neither should be flagged: 5 m of spread in 2 km is 0.25%, real variation.
	if len(km) != 0 {
		t.Errorf("kilometres flagged real variation: %s", describe(km))
	}

	// The old absolute rule would have flagged the kilometre column.
	if *inKilometres.Stats.StdDev >= 0.01 {
		t.Fatal("fixture no longer exercises the old threshold; σ must be below 0.01")
	}
}

func TestVarianceIssuesConstantColumns(t *testing.T) {
	tests := []struct {
		name    string
		col     ColumnAnalysis
		want    bool
		wantSev string
	}{
		{
			name:    "numeric with min == max",
			col:     numericCol("Setting", 7, 0, 7, 7, 50),
			want:    true,
			wantSev: "warning",
		},
		{
			name: "categorical with one distinct value",
			col: ColumnAnalysis{
				Name:  "Instrument",
				Type:  "categorical",
				Stats: ColumnStatistics{Count: 50, Unique: 1},
			},
			want:    true,
			wantSev: "warning",
		},
		{
			name: "categorical with two values is not constant",
			col: ColumnAnalysis{
				Name:  "Site",
				Type:  "categorical",
				Stats: ColumnStatistics{Count: 50, Unique: 2},
			},
			want: false,
		},
		{
			name: "an empty column is empty, not constant",
			col: ColumnAnalysis{
				Name:  "Blank",
				Type:  "categorical",
				Stats: ColumnStatistics{Count: 0, Unique: 0},
			},
			want: false,
		},
		{
			name: "ordinary variation is not reported",
			col:  numericCol("Score", 10, 2.5, 4, 16, 100),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := varianceIssuesFor(tt.col)
			got := len(issues) > 0
			if got != tt.want {
				t.Fatalf("reported = %v, want %v (%s)", got, tt.want, describe(issues))
			}
			if tt.want && issues[0].Severity != tt.wantSev {
				t.Errorf("severity = %q, want %q", issues[0].Severity, tt.wantSev)
			}
			if tt.want && len(issues[0].Affected) == 0 {
				t.Error("the issue must name the column it is about")
			}
		})
	}
}

// TestVarianceIssuesGroupsConstantColumns checks several constant columns are
// reported once, naming them all, rather than as a wall of separate entries.
func TestVarianceIssuesGroupsConstantColumns(t *testing.T) {
	issues := varianceIssuesFor(
		numericCol("A", 1, 0, 1, 1, 10),
		numericCol("B", 2, 0, 2, 2, 10),
		numericCol("C", 3, 1.5, 1, 5, 10),
	)

	if len(issues) != 1 {
		t.Fatalf("expected one grouped issue, got %d: %s", len(issues), describe(issues))
	}
	if len(issues[0].Affected) != 2 {
		t.Errorf("affected = %v, want both constant columns", issues[0].Affected)
	}
	if !strings.Contains(issues[0].Description, "Columns") {
		t.Errorf("a multi-column issue should read as plural: %q", issues[0].Description)
	}
}

func TestVarianceIssuesNearConstant(t *testing.T) {
	// Spread is 0.005% of the level -- far below the stated 0.1%.
	issues := varianceIssuesFor(numericCol("Pressure", 1000, 0.05, 999.8, 1000.2, 100))
	if len(issues) != 1 {
		t.Fatalf("expected a near-constant report, got %s", describe(issues))
	}
	if issues[0].Severity != "info" {
		t.Errorf("severity = %q, want info -- this is a judgement, not a fact", issues[0].Severity)
	}
	// The threshold must be visible, so the user can disagree with it.
	if !strings.Contains(issues[0].Impact, "0.1%") {
		t.Errorf("the message should state the threshold, got %q", issues[0].Impact)
	}
}

// TestVarianceIssuesSkipsZeroMean covers the case the coefficient of variation
// cannot judge.
//
// CV is undefined at a mean of zero and unstable near it. A mean-centred column
// is not judged rather than judged badly: saying nothing is better than a
// number that means nothing.
func TestVarianceIssuesSkipsZeroMean(t *testing.T) {
	if got := varianceIssuesFor(numericCol("Centred", 0, 0.0001, -0.001, 0.001, 100)); len(got) != 0 {
		t.Errorf("a zero-mean column should not be judged by CV, got %s", describe(got))
	}
	// But a genuinely constant zero column is still constant.
	if got := varianceIssuesFor(numericCol("AllZero", 0, 0, 0, 0, 100)); len(got) != 1 {
		t.Errorf("an all-zero column is constant and should be reported, got %s", describe(got))
	}
}

// TestQualityScoreAccountsForConstantColumns closes a contradiction that
// reporting constant columns created.
//
// The score deducted for missing values, duplicates and outliers but knew
// nothing about constant columns, so a dataset with half its variables
// carrying no information scored 100 and read "excellent" while the issues
// list said those columns contribute nothing to any component. A headline that
// disagrees with the detail is worse than either alone.
//
// Weighted like missing data, and for the same reason: a constant column
// carries as little information as an empty one.
func TestQualityScoreAccountsForConstantColumns(t *testing.T) {
	varying := func(name string) ColumnAnalysis {
		return numericCol(name, 10, 2.5, 4, 16, 100)
	}
	constant := func(name string) ColumnAnalysis {
		return numericCol(name, 7, 0, 7, 7, 100)
	}

	clean := &DataQualityReport{
		ColumnAnalysis: []ColumnAnalysis{varying("A"), varying("B"), varying("C"), varying("D")},
		DataProfile:    DataProfile{Rows: 100, NumericColumns: 4},
	}
	half := &DataQualityReport{
		ColumnAnalysis: []ColumnAnalysis{varying("A"), varying("B"), constant("C"), constant("D")},
		DataProfile:    DataProfile{Rows: 100, NumericColumns: 4},
	}

	cleanScore := calculateQualityScore(clean)
	halfScore := calculateQualityScore(half)

	if cleanScore != 100 {
		t.Errorf("a clean dataset scored %.1f, want 100", cleanScore)
	}
	if halfScore >= cleanScore {
		t.Errorf("half the columns constant scored %.1f, not below the clean %.1f",
			halfScore, cleanScore)
	}

	// Proportional: one dead column among many should barely register.
	many := make([]ColumnAnalysis, 0, 200)
	for i := 0; i < 199; i++ {
		many = append(many, varying("v"))
	}
	many = append(many, constant("dead"))
	sparse := calculateQualityScore(&DataQualityReport{
		ColumnAnalysis: many,
		DataProfile:    DataProfile{Rows: 100, NumericColumns: 200},
	})
	if sparse < 99 {
		t.Errorf("one constant column in 200 cost %.2f points, which is disproportionate",
			100-sparse)
	}
}
