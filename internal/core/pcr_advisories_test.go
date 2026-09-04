// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

package core

import (
	"math"
	"strings"
	"testing"
)

// TestCategoricalResponseAdvisory pins both edges of the detection.
//
// The threshold matters in both directions. Too eager and it cries wolf on every
// coarse measurement, which trains readers to ignore it; too lax and it stays
// silent on iris's species#target, which is the case it exists for.
func TestCategoricalResponseAdvisory(t *testing.T) {
	repeat := func(pattern []float64, times int) []float64 {
		out := make([]float64, 0, len(pattern)*times)
		for i := 0; i < times; i++ {
			out = append(out, pattern...)
		}
		return out
	}

	tests := []struct {
		name string
		y    []float64
		want bool
	}{
		{
			// The shape of testdata/iris/iris.csv: three codes, fifty rows each.
			name: "three class codes over 150 rows",
			y:    repeat([]float64{0, 1, 2}, 50),
			want: true,
		},
		{
			name: "a binary indicator",
			y:    repeat([]float64{0, 1}, 60),
			want: true,
		},
		{
			// A continuous response has as many distinct values as rows.
			name: "a genuine measurement",
			y:    []float64{1.1, 2.4, 3.9, 4.2, 5.7, 6.3, 7.8, 8.1, 9.6, 10.2, 11.5, 12.8},
			want: false,
		},
		{
			// Three distinct values among nine rows is three per value, below the
			// ten-per-value threshold. On a dataset this small, coarseness says
			// more about the sample size than about the column, and warning here
			// would fire on almost every small pilot study.
			name: "coarse but tiny",
			y:    repeat([]float64{5, 10, 15}, 3),
			want: false,
		},
		{
			name: "more distinct values than the limit",
			y:    repeat([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 20),
			want: false,
		},
		{
			name: "no observed values at all",
			y:    []float64{math.NaN(), math.NaN()},
			want: false,
		},
		{
			// Gaps must not count towards the row total, or a mostly-unmeasured
			// continuous response would look coarse.
			name: "class codes with most rows unmeasured",
			y: append(repeat([]float64{0, 1, 2}, 20),
				repeat([]float64{math.NaN()}, 200)...),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoricalResponseAdvisory("y#target", tt.y)
			if (got != "") != tt.want {
				t.Errorf("advisory %q, want present=%v", got, tt.want)
			}
			if tt.want {
				// The message has to name the column and both counts, or a reader
				// cannot check the judgement against their own data.
				for _, want := range []string{`"y#target"`, "distinct values", "classification"} {
					if !strings.Contains(got, want) {
						t.Errorf("the advisory does not mention %q: %s", want, got)
					}
				}
			}
		})
	}
}

// TestResponseAdvisoriesIsNilWhenThereIsNothingToSay lets callers range over the
// result unconditionally.
func TestResponseAdvisoriesIsNilWhenThereIsNothingToSay(t *testing.T) {
	y := make([]float64, 50)
	for i := range y {
		y[i] = float64(i) * 1.7
	}
	if got := ResponseAdvisories("y", y); got != nil {
		t.Errorf("expected no advisories on a continuous response, got %v", got)
	}
}

// TestAdvisoriesAreUnwrapped guards the contract the two front ends depend on.
//
// The CLI wraps for a terminal and the desktop lays the text out in a panel, so
// the shared text must carry no line breaks of its own. A newline here would
// appear as a hard break in the middle of a sentence in the GUI, and would also
// defeat the parity test's whitespace-normalised comparison in a way that looks
// like a divergence rather than a formatting choice.
func TestAdvisoriesAreUnwrapped(t *testing.T) {
	y := make([]float64, 90)
	for i := range y {
		y[i] = float64(i % 3)
	}
	for _, advisory := range ResponseAdvisories("y#target", y) {
		if strings.ContainsAny(advisory, "\n\r\t") {
			t.Errorf("advisory contains its own line breaks, which the callers format:\n%q",
				advisory)
		}
	}
}
