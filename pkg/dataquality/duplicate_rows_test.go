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

// TestSeparatorInACellIsNotADuplicate guards against a false match that would
// now cost data.
//
// The key used to be the cells joined with "|", so ["a|b", "c"] and ["a",
// "b|c"] both produced "a|b|c" and the second row counted as a copy of the
// first. Harmless while the result was a number nobody could act on; not
// harmless once the rows are selected and Delete Row acts on the selection.
func TestSeparatorInACellIsNotADuplicate(t *testing.T) {
	data := [][]string{
		{"a|b", "c"},
		{"a", "b|c"},
	}
	if got := findDuplicateRows(data, len(data)); len(got) != 0 {
		t.Errorf("findDuplicateRows = %v, want none: these rows differ, they merely "+
			"join to the same string", got)
	}
}

// TestGenuineDuplicatesStillMatchWithSeparators keeps the test above honest. A
// key that never matched anything would satisfy it while breaking the feature.
func TestGenuineDuplicatesStillMatchWithSeparators(t *testing.T) {
	data := [][]string{
		{"a|b", "c"},
		{"x", "y"},
		{"a|b", "c"},
	}
	got := findDuplicateRows(data, len(data))
	if len(got) != 1 || got[0] != 3 {
		t.Errorf("findDuplicateRows = %v, want [3]: row 3 is an exact copy of row 1", got)
	}
}

// TestEmptyCellsDoNotShiftTheKey covers the other way a delimiter scheme goes
// wrong: a run of empty cells must not be interchangeable with a different run.
func TestEmptyCellsDoNotShiftTheKey(t *testing.T) {
	data := [][]string{
		{"", "", "ab"},
		{"", "ab", ""},
	}
	if got := findDuplicateRows(data, len(data)); len(got) != 0 {
		t.Errorf("findDuplicateRows = %v, want none: the same values in different "+
			"columns are different rows", got)
	}
}

// TestRepeatedRowsPhraseAgreesInNumber checks the user-facing text for both
// counts. Getting the noun right and leaving the verb wrong -- "1 row repeat"
// -- reads no better than "1 rows".
func TestRepeatedRowsPhraseAgreesInNumber(t *testing.T) {
	one := repeatedRowsPhrase(1)
	if !strings.HasPrefix(one, "1 row repeats") {
		t.Errorf("singular phrase = %q, want it to begin \"1 row repeats\"", one)
	}
	if strings.Contains(one, "1 rows") || strings.Contains(one, "row repeat ") {
		t.Errorf("singular phrase does not agree in number: %q", one)
	}

	many := repeatedRowsPhrase(97)
	if !strings.HasPrefix(many, "97 rows repeat ") {
		t.Errorf("plural phrase = %q, want it to begin \"97 rows repeat \"", many)
	}
}

// TestDuplicateFindingReadsCorrectlyForOneRow follows the phrase through to the
// issue and the recommendation, since both render it and either could
// reintroduce the disagreement on its own.
func TestDuplicateFindingReadsCorrectlyForOneRow(t *testing.T) {
	report := &DataQualityReport{DataProfile: DataProfile{DuplicateRows: 1}}

	issue, found := issueOfCategory(generateQualityIssues(report, nil, nil, []int{5}), "duplicate")
	if !found {
		t.Fatal("no duplicate issue raised")
	}
	if !strings.HasPrefix(issue.Description, "1 row repeats") {
		t.Errorf("issue description = %q", issue.Description)
	}

	for _, rec := range generateRecommendations(report) {
		if rec.Category == "duplicate" && !strings.HasPrefix(rec.Description, "1 row repeats") {
			t.Errorf("recommendation description = %q", rec.Description)
		}
	}
}

// TestLastColumnStillCounts closes a hole found by cross-checking the shipped
// function against a brute-force implementation on the real dataset.
//
// That check agreed exactly -- 97 rows, no false positives or negatives -- and
// caught a key built from the first column alone, and a key made unique per
// row. It did not catch a key built from every column except the last, because
// the final column of that file never distinguishes a near-pair, so ignoring it
// changed nothing. None of the other tests here covered it either: they all
// differ in an earlier column.
func TestLastColumnStillCounts(t *testing.T) {
	data := [][]string{
		{"same", "same", "different"},
		{"same", "same", "values"},
	}
	if got := findDuplicateRows(data, len(data)); len(got) != 0 {
		t.Errorf("findDuplicateRows = %v, want none: these rows differ in the last "+
			"column, which the key must still read", got)
	}
}

// TestFirstColumnStillCounts is the same hole at the other end, so a key built
// from the tail of each row fails too.
func TestFirstColumnStillCounts(t *testing.T) {
	data := [][]string{
		{"different", "same", "same"},
		{"values", "same", "same"},
	}
	if got := findDuplicateRows(data, len(data)); len(got) != 0 {
		t.Errorf("findDuplicateRows = %v, want none: these rows differ in the first "+
			"column", got)
	}
}

// TestDuplicateImpactSaysOriginalsAreNotSelected guards the message that made
// the finding checkable. Without it a reader compares a highlighted row against
// the unhighlighted row below, finds them different, and concludes the
// detection is broken (#936) -- when the row it duplicates is the one above.
func TestDuplicateImpactSaysOriginalsAreNotSelected(t *testing.T) {
	report := &DataQualityReport{DataProfile: DataProfile{DuplicateRows: 2}}
	issue, found := issueOfCategory(generateQualityIssues(report, nil, nil, []int{4, 9}), "duplicate")
	if !found {
		t.Fatal("no duplicate issue raised")
	}
	for _, want := range []string{"unselected", "unhighlighted"} {
		if !strings.Contains(issue.Impact, want) {
			t.Errorf("impact never mentions %q, so a reader cannot tell which row of a "+
				"pair is selected: %q", want, issue.Impact)
		}
	}
}
