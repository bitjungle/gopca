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

// skewedCol returns a numeric column the shape heuristic will not call normal.
func skewedCol(name string) ColumnAnalysis {
	return ColumnAnalysis{
		Name:         name,
		Type:         "numeric",
		Distribution: DistributionInfo{IsNormal: false},
	}
}

func distributionIssue(t *testing.T, cols ...ColumnAnalysis) (QualityIssue, bool) {
	t.Helper()
	report := &DataQualityReport{ColumnAnalysis: cols}
	for _, issue := range generateQualityIssues(report, nil, nil) {
		if issue.Category == "distribution" {
			return issue, true
		}
	}
	return QualityIssue{}, false
}

// TestDistributionIssueDoesNotClaimPCAAssumesNormality guards the correction in
// #929.
//
// The report used to tell users "PCA assumes normality; consider data
// transformations". PCA assumes nothing of the sort, and both end-user guides
// say so in as many words -- docs/intro_to_pca.md and docs/intro_to_data_prep.md
// -- so the application contradicted the documentation shipped beside it.
//
// Asserting the current wording verbatim would pass for any rephrasing,
// including a reintroduction of the claim in different words. This asserts the
// property that was wrong: the message must not tell the reader that PCA
// requires, assumes or expects a distribution.
func TestDistributionIssueDoesNotClaimPCAAssumesNormality(t *testing.T) {
	issue, found := distributionIssue(t, skewedCol("a"))
	if !found {
		t.Fatal("no distribution issue was raised for a skewed column")
	}

	text := strings.ToLower(issue.Description + " " + issue.Impact)
	for _, claim := range []string{
		"assumes normality",
		"assumes normal",
		"requires normal",
		"expects normal",
		"pca assumes a",
	} {
		if strings.Contains(text, claim) {
			t.Errorf("distribution issue claims PCA assumes a distribution (%q): %q", claim, text)
		}
	}

	// "normal" may still appear, but only in a sentence denying the assumption.
	// Catch the bare noun phrase the old message used.
	if strings.Contains(text, "non-normal") {
		t.Errorf("issue reports non-normality, which is not what IsNormal measures: %q", text)
	}
}

// TestDistributionIssueNamesWhatWasMeasured checks the message describes the
// heuristic that produced it. IsNormal is |skewness| < 0.5 && |kurtosis| < 1.0,
// a shape test, so the finding is about skew and tails.
func TestDistributionIssueNamesWhatWasMeasured(t *testing.T) {
	issue, found := distributionIssue(t, skewedCol("a"), skewedCol("b"))
	if !found {
		t.Fatal("no distribution issue was raised for skewed columns")
	}

	text := strings.ToLower(issue.Description)
	if !strings.Contains(text, "skew") && !strings.Contains(text, "tail") {
		t.Errorf("description names neither skew nor tails: %q", issue.Description)
	}
	if !strings.Contains(issue.Description, "2") {
		t.Errorf("description does not report the column count: %q", issue.Description)
	}
	if issue.Severity != "info" {
		t.Errorf("severity = %q, want info: a skewed column is not a defect", issue.Severity)
	}
}

// TestNoDistributionIssueWhenNothingIsSkewed keeps the two tests above honest:
// without it they would pass against a function that raises the issue
// unconditionally.
func TestNoDistributionIssueWhenNothingIsSkewed(t *testing.T) {
	normal := ColumnAnalysis{
		Name:         "a",
		Type:         "numeric",
		Distribution: DistributionInfo{IsNormal: true},
	}
	if issue, found := distributionIssue(t, normal); found {
		t.Errorf("distribution issue raised for an unskewed column: %q", issue.Description)
	}
}
