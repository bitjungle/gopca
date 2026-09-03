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

package core

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestRegressionMetricsDecomposition asserts the exact relationship between the
// three error measures:
//
//	RMSE² = bias² + (n-1)/n · SEP²
//
// It is an identity, so the tolerance can be tight, and it fails loudly if any of
// the three is computed with the wrong divisor. That matters because a wrong
// divisor shifts a reported error by only a few percent, which is exactly the
// size of difference a reader would attribute to the data rather than to a bug.
func TestRegressionMetricsDecomposition(t *testing.T) {
	r := rand.New(rand.NewPCG(2024, 7))

	for _, offset := range []float64{0, 0.4, -2.5} {
		n := 137
		measured := make([]float64, n)
		predicted := make([]float64, n)
		for i := 0; i < n; i++ {
			measured[i] = r.NormFloat64()
			predicted[i] = measured[i] + offset + 0.7*r.NormFloat64()
		}

		m, err := ComputeRegressionMetrics(predicted, measured)
		if err != nil {
			t.Fatalf("ComputeRegressionMetrics: %v", err)
		}

		nf := float64(n)
		want := m.RMSE * m.RMSE
		got := m.Bias*m.Bias + (nf-1)/nf*m.SEP*m.SEP
		if math.Abs(want-got) > 1e-12*(1+want) {
			t.Errorf("offset %v: RMSE² = %.15g but bias² + (n-1)/n·SEP² = %.15g", offset, want, got)
		}

		// A deliberate offset must show up as bias and leave SEP roughly alone.
		if math.Abs(m.Bias-offset) > 0.2 {
			t.Errorf("offset %v: measured bias %.4f is not close to the offset applied", offset, m.Bias)
		}
	}
}

// TestRegressionMetricsPerfectAndBaseline covers the two anchors of the scale.
func TestRegressionMetricsPerfectAndBaseline(t *testing.T) {
	measured := []float64{1, 2, 3, 4, 5, 6}

	perfect, err := ComputeRegressionMetrics(measured, measured)
	if err != nil {
		t.Fatalf("perfect fit: %v", err)
	}
	if perfect.RMSE != 0 || perfect.Bias != 0 || perfect.MAE != 0 || perfect.SEP != 0 {
		t.Errorf("a perfect fit should have zero error, got %+v", perfect)
	}
	if math.Abs(perfect.R2-1) > 1e-15 {
		t.Errorf("a perfect fit should have R2 = 1, got %v", perfect.R2)
	}

	mean := 3.5
	baseline := []float64{mean, mean, mean, mean, mean, mean}
	flat, err := ComputeRegressionMetrics(baseline, measured)
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	if math.Abs(flat.R2) > 1e-15 {
		t.Errorf("predicting the mean should give R2 = 0, got %v", flat.R2)
	}
	if math.Abs(flat.Bias) > 1e-15 {
		t.Errorf("predicting the mean should be unbiased, got %v", flat.Bias)
	}
}

// TestRegressionMetricsConstantResponse checks that an undefined R2 is reported as
// zero rather than as an infinity that would then propagate into a selection rule.
func TestRegressionMetricsConstantResponse(t *testing.T) {
	measured := []float64{5, 5, 5, 5}
	predicted := []float64{5.1, 4.9, 5.2, 4.8}

	m, err := ComputeRegressionMetrics(predicted, measured)
	if err != nil {
		t.Fatalf("ComputeRegressionMetrics: %v", err)
	}
	if m.R2 != 0 {
		t.Errorf("R2 against a constant response = %v, want 0", m.R2)
	}
	if m.RMSE == 0 {
		t.Error("RMSE should still be positive")
	}
}

func TestRegressionMetricsErrors(t *testing.T) {
	tests := []struct {
		name                string
		predicted, measured []float64
	}{
		{"no observations", []float64{}, []float64{}},
		{"length mismatch", []float64{1, 2}, []float64{1, 2, 3}},
		{"NaN prediction", []float64{1, math.NaN()}, []float64{1, 2}},
		{"infinite measurement", []float64{1, 2}, []float64{1, math.Inf(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ComputeRegressionMetrics(tt.predicted, tt.measured); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

// TestSelectComponentsRules exercises each rule against a curve whose shape makes
// the differences between them visible: a sharp drop, then a long flat tail where
// further components buy almost nothing.
func TestSelectComponentsRules(t *testing.T) {
	candidates := []int{0, 1, 2, 3, 4, 5, 6}
	curve := []float64{1.00, 0.50, 0.30, 0.29, 0.285, 0.284, 0.2835}
	stderr := []float64{0.05, 0.04, 0.02, 0.02, 0.02, 0.02, 0.02}

	tests := []struct {
		name      string
		rule      string
		tolerance float64
		woldR     float64
		want      int
	}{
		{"minimum takes the last small gain", types.SelectMin, 0, 0, 6},
		{"one standard error stops at the knee", types.SelectOneSE, 0, 0, 2},
		{"tolerance of 0.02 stops at the knee", types.SelectTolerance, 0.02, 0, 2},
		{"tolerance of 0 matches the minimum", types.SelectTolerance, 0, 0, 6},
		{"empty rule defaults to the minimum", "", 0, 0, 6},
		{"wold stops where improvement stalls", types.SelectWold, 0, 0.95, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectComponents(candidates, curve, stderr, tt.rule, tt.tolerance, tt.woldR)
			if err != nil {
				t.Fatalf("SelectComponents: %v", err)
			}
			if got != tt.want {
				t.Errorf("selected %d components, want %d", got, tt.want)
			}
		})
	}
}

// TestSelectComponentsWoldUsesPress pins the convention. Wold's R is defined on
// PRESS, the sum of squared residuals, so the threshold applies to the square of
// the RMSE ratio. Applying it to the RMSE ratio directly would make a nominal 0.95
// demand a 10% improvement in PRESS instead of 5%, stopping earlier than the
// published criterion and than the user asked for.
func TestSelectComponentsWoldUsesPress(t *testing.T) {
	candidates := []int{0, 1, 2}

	// The RMSE ratio from index 1 to 2 is 0.96, so the PRESS ratio is 0.9216.
	curve := []float64{2.0, 1.0, 0.96}

	// A PRESS threshold of 0.95 is satisfied by 0.9216, so the sweep continues.
	got, err := SelectComponents(candidates, curve, nil, types.SelectWold, 0, 0.95)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if got != 2 {
		t.Errorf("selected %d, want 2: a PRESS ratio of 0.9216 clears a 0.95 threshold", got)
	}

	// A stricter threshold of 0.90 is not satisfied, so it stops at the previous.
	got, err = SelectComponents(candidates, curve, nil, types.SelectWold, 0, 0.90)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if got != 1 {
		t.Errorf("selected %d, want 1: a PRESS ratio of 0.9216 fails a 0.90 threshold", got)
	}
}

// TestSelectComponentsWoldStopsEarlyOnPlateaus documents a real weakness rather
// than hiding it. Wold's rule is greedy: it stops the first time improvement
// stalls, so a curve that plateaus before falling again is cut short. The shape
// here is the one testdata/corn/corn.csv actually produces.
func TestSelectComponentsWoldStopsEarlyOnPlateaus(t *testing.T) {
	candidates := []int{0, 1, 2, 3, 4, 5, 6, 7}
	curve := []float64{0.397, 0.327, 0.337, 0.254, 0.256, 0.266, 0.152, 0.060}

	got, err := SelectComponents(candidates, curve, nil, types.SelectWold, 0, 1.0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if got != 1 {
		t.Errorf("selected %d, want 1: the rule stops at the first plateau even though "+
			"the error later falls by a factor of five", got)
	}

	// The minimum, by contrast, finds the real optimum on the same curve.
	best, err := SelectComponents(candidates, curve, nil, types.SelectMin, 0, 0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if best != 7 {
		t.Errorf("the minimum rule selected %d, want 7", best)
	}
}

func TestSelectComponentsErrors(t *testing.T) {
	candidates := []int{0, 1, 2}
	curve := []float64{1.0, 0.5, 0.4}

	tests := []struct {
		name       string
		candidates []int
		curve      []float64
		stderr     []float64
		rule       string
		tolerance  float64
	}{
		{"no candidates", nil, nil, nil, types.SelectMin, 0},
		{"curve length mismatch", candidates, []float64{1.0}, nil, types.SelectMin, 0},
		{"one-se without standard errors", candidates, curve, nil, types.SelectOneSE, 0},
		{"negative tolerance", candidates, curve, nil, types.SelectTolerance, -1},
		{"unknown rule", candidates, curve, nil, "vibes", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := SelectComponents(tt.candidates, tt.curve, tt.stderr,
				tt.rule, tt.tolerance, 0); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

// TestPCRRefusesMissingPredictors checks that incomplete predictors are refused
// rather than quietly imputed. Filling them by column mean estimates a value from
// the data, so doing it before cross-validation would let the held-out rows shape
// what the model trains on and make every reported error optimistic.
func TestPCRRefusesMissingPredictors(t *testing.T) {
	data, y := makeRegressionData(30, 5, 2, 0.3, 91)
	data[7][2] = math.NaN()

	_, err := NewPCREngine().Fit(data, y, fixedConfig(2))
	if err == nil {
		t.Fatal("expected missing predictors to be refused")
	}

	// NIPALS with native handling is the one method that may see them.
	config := fixedConfig(2)
	config.PCA.Method = "nipals"
	config.PCA.MissingStrategy = types.MissingNative
	if _, err := NewPCREngine().Fit(data, y, config); err != nil {
		t.Errorf("NIPALS with native missing-value handling should accept incomplete data: %v", err)
	}
}

// TestSelectComponentsFirstMinimum covers the rule a practitioner applies by eye:
// take the first point the curve turns upward from, rather than the lowest point
// anywhere in the range.
func TestSelectComponentsFirstMinimum(t *testing.T) {
	tests := []struct {
		name       string
		candidates []int
		curve      []float64
		stderr     []float64
		want       int
	}{
		{
			name:       "stops at the first real turn",
			candidates: []int{0, 1, 2, 3, 4, 5},
			curve:      []float64{1.0, 0.6, 0.4, 0.7, 0.3, 0.2},
			stderr:     []float64{0.02, 0.02, 0.02, 0.02, 0.02, 0.02},
			want:       2,
		},
		{
			name:       "a wiggle smaller than the noise is not a turn",
			candidates: []int{0, 1, 2, 3, 4},
			curve:      []float64{1.0, 0.50, 0.51, 0.30, 0.20},
			stderr:     []float64{0.1, 0.1, 0.1, 0.1, 0.1},
			want:       4,
		},
		{
			name:       "a monotone curve runs to the end of the range",
			candidates: []int{0, 1, 2, 3},
			curve:      []float64{1.0, 0.8, 0.6, 0.4},
			stderr:     []float64{0.01, 0.01, 0.01, 0.01},
			want:       3,
		},
		{
			name:       "a curve that only rises stops immediately",
			candidates: []int{0, 1, 2},
			curve:      []float64{0.2, 0.5, 0.9},
			stderr:     []float64{0.01, 0.01, 0.01},
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectComponents(tt.candidates, tt.curve, tt.stderr,
				types.SelectFirstMin, 0, 0)
			if err != nil {
				t.Fatalf("SelectComponents: %v", err)
			}
			if got != tt.want {
				t.Errorf("selected %d, want %d", got, tt.want)
			}
		})
	}
}

// TestSelectComponentsFirstMinimumPassesOverADeeperOne is the property that makes
// the rule worth having and worth warning about at the same time.
//
// The curve here is the shape testdata/bronir2/bronir2.csv actually produces for
// Dens#target: an early shoulder, a long unstable stretch, then a much lower
// minimum far to the right. Stopping at the shoulder is the conservative reading
// and it gives up a great deal, which is why the interface reports where the
// lowest point was rather than only what was chosen.
func TestSelectComponentsFirstMinimumPassesOverADeeperOne(t *testing.T) {
	candidates := []int{0, 1, 2, 3, 4, 5, 6, 7}
	curve := []float64{8.42, 7.43, 6.57, 6.52, 6.54, 6.51, 6.80, 3.48}
	stderr := []float64{0.20, 0.20, 0.20, 0.20, 0.20, 0.20, 0.20, 0.20}

	first, err := SelectComponents(candidates, curve, stderr, types.SelectFirstMin, 0, 0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	lowest, err := SelectComponents(candidates, curve, stderr, types.SelectMin, 0, 0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}

	if first >= lowest {
		t.Errorf("the first-minimum rule selected %d and the lowest point is %d; "+
			"on this curve the rule is supposed to stop earlier", first, lowest)
	}
	if curve[first] <= curve[lowest] {
		t.Errorf("the rule stopped somewhere no worse than the minimum, so this curve " +
			"does not exercise the trade-off it exists to make")
	}
}
