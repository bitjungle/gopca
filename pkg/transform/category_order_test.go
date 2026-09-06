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

package transform

import (
	"strings"
	"testing"
)

func TestSuggestCategoryOrder(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []string
		why    string
	}{
		{
			name:   "ordinal vocabulary beats alphabetical",
			values: []string{"high", "low", "medium"},
			want:   []string{"low", "medium", "high"},
			why:    "alphabetical would give high, low, medium -- the scrambled order",
		},
		{
			name:   "case and padding are ignored when matching",
			values: []string{"HIGH", " Low ", "Medium"},
			want:   []string{"Low", "Medium", "HIGH"},
			why:    "matching is case-insensitive, but the values are returned as written",
		},
		{
			name:   "subset of a vocabulary still orders",
			values: []string{"high", "low"},
			want:   []string{"low", "high"},
			why:    "a scale need not use every rung",
		},
		{
			name:   "months",
			values: []string{"March", "January", "February"},
			want:   []string{"January", "February", "March"},
			why:    "alphabetical would give February, January, March",
		},
		{
			name:   "one unrecognised value falls back to alphabetical",
			values: []string{"low", "high", "unknown"},
			want:   []string{"high", "low", "unknown"},
			why: "a partial match means this is not that vocabulary; guessing an " +
				"order for the rest would be a fabricated scale",
		},
		{
			name:   "unrelated categories sort alphabetically",
			values: []string{"tokyo", "paris", "amsterdam"},
			want:   []string{"amsterdam", "paris", "tokyo"},
			why:    "no vocabulary applies, so LabelEncoder's ordering is used",
		},
		{
			name:   "duplicates and blanks are dropped",
			values: []string{"low", "", "low", "high", "   "},
			want:   []string{"low", "high"},
			why:    "the order lists categories, not rows",
		},
		{
			name:   "norwegian lav/middels/høy",
			values: []string{"høy", "lav", "middels"},
			want:   []string{"lav", "middels", "høy"},
			why:    "alphabetical would give høy, lav, middels",
		},
		{
			name:   "norwegian is matched case-insensitively through the ø",
			values: []string{"HØY", "Lav", "Middels"},
			want:   []string{"Lav", "Middels", "HØY"},
			why:    "lowercasing must handle non-ASCII, and the original spelling is kept",
		},
		{
			name:   "nynorsk høg",
			values: []string{"høg", "lav", "middels"},
			want:   []string{"lav", "middels", "høg"},
			why:    "both written standards are covered",
		},
		{
			name:   "norwegian weekdays",
			values: []string{"onsdag", "mandag", "fredag"},
			want:   []string{"mandag", "onsdag", "fredag"},
			why:    "alphabetical would give fredag, mandag, onsdag",
		},
		{
			name:   "norwegian months keep their own order",
			values: []string{"desember", "mars", "mai"},
			want:   []string{"mars", "mai", "desember"},
			why:    "alphabetical would give desember, mai, mars",
		},
		{
			name:   "norwegian likert scale",
			values: []string{"enig", "helt uenig", "nøytral", "uenig", "helt enig"},
			want:   []string{"helt uenig", "uenig", "nøytral", "enig", "helt enig"},
			why:    "a five-point scale whose alphabetical order is meaningless",
		},
		{
			name:   "mixed languages fall back to alphabetical",
			values: []string{"lav", "high"},
			want:   []string{"high", "lav"},
			why: "one vocabulary must cover every value; mixing them is a sign " +
				"the words are not being used as a scale",
		},
		{
			name:   "a single value needs no vocabulary",
			values: []string{"medium"},
			want:   []string{"medium"},
			why:    "nothing to order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestCategoryOrder(tt.values)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("got %v, want %v (%s)", got, tt.want, tt.why)
			}
		})
	}
}

// TestSuggestCategoryOrderFeedsApply checks the suggestion is usable as the
// order Apply accepts -- the two halves are only useful joined up.
func TestSuggestCategoryOrderFeedsApply(t *testing.T) {
	values := []string{"high", "low", "medium", "low"}

	res, err := Apply(ordinalInput(values...), Options{
		Type:          Ordinal,
		Columns:       []string{"Cat"},
		CategoryOrder: map[string][]string{"Cat": SuggestCategoryOrder(values)},
	})
	if err != nil {
		t.Fatalf("Apply ordinal: %v", err)
	}

	want := []string{"2", "0", "1", "0"}
	got := codeColumn(t, res, "Cat_code")
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d (%q): got %q, want %q; full column %v",
				i, values[i], got[i], want[i], got)
		}
	}
}
