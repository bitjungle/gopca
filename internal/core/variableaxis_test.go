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

package core_test

import (
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// loadAxisDataset reads through the same parser the applications use, so the
// test sees the variables the engine would.
func loadAxisDataset(t *testing.T, relative string) (types.Matrix, []string) {
	t.Helper()
	opts := pkgcsv.DefaultOptions()
	opts.ParseMode = pkgcsv.ParseMixedWithTargets
	parsed, err := pkgcsv.NewReader(opts).ReadFile(
		filepath.Join("..", "..", "testdata", relative))
	if err != nil {
		t.Fatalf("reading %s: %v", relative, err)
	}
	return types.Matrix(parsed.Matrix), parsed.Headers
}

// TestAnalyzeVariableAxisOnRealDatasets is the test the threshold was chosen
// from, run against the datasets it was measured on.
//
// The separation is not marginal: every spectral dataset sits below 0.007 and
// nothing else comes under 1.1, so this would fail loudly long before a change
// made the two groups merely close. The EEG case is the one worth keeping: it
// has 224 variables and is not a continuum, so any rule based on counting
// columns would classify it wrongly.
func TestAnalyzeVariableAxisOnRealDatasets(t *testing.T) {
	tests := []struct {
		dataset        string
		wantContinuous bool
		note           string
	}{
		{"corn/corn.csv", true, "NIR spectra, 700 wavelengths"},
		{"bronir2/bronir2.csv", true, "NIR spectra, 1001 wavelengths"},
		{"wine/wine.csv", false, "13 unrelated chemical measurements"},
		{"iris/iris.csv", false, "4 unrelated flower measurements"},
		{"nhanes/body_measures.csv", false, "unordered body measurements"},
		{"swiss_roll/swiss_roll_color_target.csv", false, "3 coordinates"},
	}

	for _, tt := range tests {
		t.Run(tt.dataset, func(t *testing.T) {
			data, headers := loadAxisDataset(t, tt.dataset)
			report := core.AnalyzeVariableAxis(data, headers)

			if report.IsContinuous != tt.wantContinuous {
				t.Errorf("%s (%s): IsContinuous = %v, want %v (ratio %.5g, %.1fx smoother than random)",
					tt.dataset, tt.note, report.IsContinuous, tt.wantContinuous,
					report.Continuity, report.SmoothnessFactor)
			}

			// The margin matters as much as the verdict. A threshold that the
			// data only just clears would be a coin toss on the next dataset.
			if tt.wantContinuous && report.Continuity > core.ContinuityThreshold/10 {
				t.Errorf("%s: ratio %.5g is within a factor of 10 of the threshold %.2g; "+
					"the separation this rule relies on has narrowed",
					tt.dataset, report.Continuity, core.ContinuityThreshold)
			}
			if !tt.wantContinuous && report.Continuity < core.ContinuityThreshold*2 {
				t.Errorf("%s: ratio %.5g is within a factor of 2 of the threshold %.2g",
					tt.dataset, report.Continuity, core.ContinuityThreshold)
			}
		})
	}
}

// TestAnalyzeVariableAxisSpacing covers the half of the report that the data
// cannot supply.
func TestAnalyzeVariableAxisSpacing(t *testing.T) {
	data, headers := loadAxisDataset(t, "corn/corn.csv")

	full := core.AnalyzeVariableAxis(data, headers)
	if !full.NamesNumeric || !full.SpacingUniform || full.DistinctSteps != 1 {
		t.Fatalf("intact corn: numeric=%v uniform=%v steps=%d, want true/true/1",
			full.NamesNumeric, full.SpacingUniform, full.DistinctSteps)
	}

	// Excluding a band from the middle makes non-adjacent wavelengths
	// neighbours. This is the case the continuity ratio cannot see, and the
	// reason the spacing check exists.
	keep := make([]int, 0, len(headers))
	for i, h := range headers {
		wl := parseFloatOrZero(h)
		if wl >= 1400 && wl <= 1500 {
			continue
		}
		keep = append(keep, i)
	}
	if len(keep) == len(headers) {
		t.Fatal("the exclusion removed nothing; the test would prove nothing")
	}

	gapped := make(types.Matrix, len(data))
	for i, row := range data {
		gapped[i] = make([]float64, len(keep))
		for k, j := range keep {
			gapped[i][k] = row[j]
		}
	}
	gappedHeaders := make([]string, len(keep))
	for k, j := range keep {
		gappedHeaders[k] = headers[j]
	}

	report := core.AnalyzeVariableAxis(gapped, gappedHeaders)
	if report.SpacingUniform {
		t.Error("a spectrum with an interior band removed was reported as evenly spaced")
	}
	if report.DistinctSteps < 2 {
		t.Errorf("expected more than one step size after the exclusion, got %d", report.DistinctSteps)
	}
	// And the point of having two checks: the data still looks perfectly smooth.
	if !report.IsContinuous {
		t.Error("the gapped spectrum should still read as continuous; " +
			"if it does not, the spacing check is no longer the only thing catching this")
	}
}

// TestAnalyzeVariableAxisSurvivesNoise guards against the opposite failure:
// rejecting real spectra because they are noisy.
func TestAnalyzeVariableAxisSurvivesNoise(t *testing.T) {
	data, headers := loadAxisDataset(t, "corn/corn.csv")

	// The standard deviation about the mean, not the root mean square about
	// zero. For absorbance spectra, which are entirely positive, the two differ
	// by a large factor, and the figure quoted as evidence for this threshold
	// was measured against the standard deviation.
	var sum float64
	var n int
	for _, row := range data {
		for _, v := range row {
			sum += v
			n++
		}
	}
	mean := sum / float64(n)
	var ss float64
	for _, row := range data {
		for _, v := range row {
			ss += (v - mean) * (v - mean)
		}
	}
	sd := math.Sqrt(ss / float64(n-1))

	rng := rand.New(rand.NewSource(20260908))
	for _, pct := range []float64{1, 5, 20} {
		noisy := make(types.Matrix, len(data))
		for i, row := range data {
			noisy[i] = make([]float64, len(row))
			for j, v := range row {
				noisy[i][j] = v + rng.NormFloat64()*sd*pct/100
			}
		}
		report := core.AnalyzeVariableAxis(noisy, headers)
		if !report.IsContinuous {
			t.Errorf("corn with %.0f%% noise was rejected as non-continuous (ratio %.5g)",
				pct, report.Continuity)
		}
		t.Logf("%.0f%% noise: ratio %.5g (threshold %.2g)", pct, report.Continuity, core.ContinuityThreshold)
	}
}

// TestAnalyzeVariableAxisShuffledColumns is the control that shows the statistic
// measures the ordering and not something else about the data. The same values
// in a different column order must stop being a continuum.
func TestAnalyzeVariableAxisShuffledColumns(t *testing.T) {
	data, _ := loadAxisDataset(t, "corn/corn.csv")

	rng := rand.New(rand.NewSource(11))
	order := rng.Perm(len(data[0]))
	shuffled := make(types.Matrix, len(data))
	for i, row := range data {
		shuffled[i] = make([]float64, len(row))
		for k, j := range order {
			shuffled[i][k] = row[j]
		}
	}

	report := core.AnalyzeVariableAxis(shuffled, nil)
	if report.IsContinuous {
		t.Errorf("corn with its columns shuffled was still reported as a continuum (ratio %.5g)",
			report.Continuity)
	}
	// Near the random-ordering expectation of 2, which is what makes the
	// statistic interpretable rather than merely comparative.
	if report.Continuity < 1.5 || report.Continuity > 2.5 {
		t.Errorf("shuffled ratio %.4g is far from the expected value of 2 under random ordering",
			report.Continuity)
	}
}

func TestAnalyzeVariableAxisEdgeCases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if core.AnalyzeVariableAxis(nil, nil).IsContinuous {
			t.Error("no data cannot be a continuum")
		}
	})

	t.Run("too few variables to hold a window", func(t *testing.T) {
		// Perfectly smooth, but a window needs at least three points.
		data := types.Matrix{{1, 2}, {2, 4}, {3, 6}}
		if core.AnalyzeVariableAxis(data, []string{"1", "2"}).IsContinuous {
			t.Error("two variables cannot carry a filter")
		}
	})

	t.Run("flat rows are skipped rather than counted", func(t *testing.T) {
		// A constant row has no variance to divide by. Counting it as either
		// smooth or rough would let padding rows decide the verdict.
		data := types.Matrix{{5, 5, 5, 5, 5}, {1, 2, 3, 4, 5}, {2, 4, 6, 8, 10}}
		report := core.AnalyzeVariableAxis(data, nil)
		if math.IsNaN(report.Continuity) {
			t.Error("a flat row produced a NaN ratio")
		}
	})

	t.Run("headers of the wrong length leave spacing unreported", func(t *testing.T) {
		data := types.Matrix{{1, 2, 3, 4}, {2, 3, 4, 5}}
		report := core.AnalyzeVariableAxis(data, []string{"1", "2"})
		if report.NamesNumeric {
			t.Error("spacing was reported from headers that do not match the data")
		}
	})

	t.Run("non-monotonic numeric names are not an axis", func(t *testing.T) {
		data := types.Matrix{{1, 2, 3, 4}, {2, 3, 4, 5}}
		report := core.AnalyzeVariableAxis(data, []string{"10", "20", "15", "30"})
		if report.SpacingUniform {
			t.Error("names that go up then down were reported as evenly spaced")
		}
	})
}

func parseFloatOrZero(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// TestAnalyzeVariableAxisUnknownIsNotNegative separates "nothing could be
// measured" from "these variables are not a continuum".
//
// They are different claims and only one of them is a finding. Reporting the
// first as the second would let a dataset of flat or incomplete rows disable a
// control on evidence that was never gathered.
func TestAnalyzeVariableAxisUnknownIsNotNegative(t *testing.T) {
	t.Run("all rows flat", func(t *testing.T) {
		data := types.Matrix{{5, 5, 5, 5, 5}, {2, 2, 2, 2, 2}}
		report := core.AnalyzeVariableAxis(data, nil)
		if report.Measurable {
			t.Error("rows with no variance cannot yield a ratio, so nothing was measurable")
		}
		if report.IsContinuous {
			t.Error("an unmeasurable axis must not be reported as continuous either")
		}
	})

	t.Run("all rows carry a missing value", func(t *testing.T) {
		data := types.Matrix{
			{1, math.NaN(), 3, 4, 5},
			{2, 3, math.NaN(), 5, 6},
		}
		report := core.AnalyzeVariableAxis(data, nil)
		if report.Measurable {
			t.Error("every row was skipped, so nothing was measurable")
		}
	})

	t.Run("a usable row makes it measurable", func(t *testing.T) {
		data := types.Matrix{
			{5, 5, 5, 5, 5},          // flat, skipped
			{1, math.NaN(), 3, 4, 5}, // incomplete, skipped
			{1, 2, 3, 4, 5},          // usable
		}
		report := core.AnalyzeVariableAxis(data, nil)
		if !report.Measurable {
			t.Error("one usable row is enough to measure")
		}
		if !report.IsContinuous {
			t.Errorf("a straight line is as continuous as an axis gets (ratio %.5g)", report.Continuity)
		}
	})
}
