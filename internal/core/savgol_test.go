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
	"math/rand"
	"strings"
	"testing"
)

// evalPoly evaluates sum_j coeffs[j] * t^j.
func evalPoly(coeffs []float64, t float64) float64 {
	result, power := 0.0, 1.0
	for _, c := range coeffs {
		result += c * power
		power *= t
	}
	return result
}

// derivPoly returns the coefficients of the d-th derivative of a polynomial.
func derivPoly(coeffs []float64, d int) []float64 {
	out := append([]float64(nil), coeffs...)
	for ; d > 0; d-- {
		if len(out) <= 1 {
			return []float64{0}
		}
		next := make([]float64, len(out)-1)
		for j := 1; j < len(out); j++ {
			next[j-1] = out[j] * float64(j)
		}
		out = next
	}
	return out
}

// TestSavGolReproducesPolynomialsExactly is the defining property of the filter,
// and the check that would fail if any coefficient were wrong.
//
// A Savitzky-Golay filter of polynomial order p fits a degree-p polynomial by
// least squares. When the data *is* a polynomial of degree at most p, that fit
// is exact -- the residual is zero at every point -- so the filter must return
// the polynomial's own d-th derivative, everywhere, including the edge
// positions where a different window is used. Nothing weaker distinguishes a
// correct coefficient vector from a plausible one: a filter with subtly wrong
// weights still smooths, still looks like a spectrum, and still passes any test
// that only asks whether the output is finite and the right length.
func TestSavGolReproducesPolynomialsExactly(t *testing.T) {
	tests := []struct {
		name      string
		poly      []float64 // ascending powers
		window    int
		polyOrder int
		deriv     int
	}{
		{"constant, smoothed", []float64{7.5}, 5, 2, 0},
		{"constant, first derivative is zero", []float64{7.5}, 5, 2, 1},
		{"line, smoothed", []float64{3, -2}, 7, 2, 0},
		{"line, first derivative is the slope", []float64{3, -2}, 7, 2, 1},
		{"line, second derivative is zero", []float64{3, -2}, 7, 2, 2},
		{"quadratic, smoothed", []float64{3, -2, 0.5}, 9, 2, 0},
		{"quadratic, first derivative", []float64{3, -2, 0.5}, 9, 2, 1},
		{"quadratic, second derivative is constant", []float64{3, -2, 0.5}, 9, 2, 2},
		{"cubic with cubic fit, first derivative", []float64{1, 2, -0.3, 0.05}, 11, 3, 1},
		{"quartic with quartic fit, second derivative", []float64{-2, 1, 0.4, -0.02, 0.003}, 13, 4, 2},
		{"wide window still exact", []float64{3, -2, 0.5}, 31, 2, 1},
		{"window equal to the spectrum length", []float64{3, -2, 0.5}, 41, 2, 1},
	}

	const nVars = 41
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := make([]float64, nVars)
			for i := range row {
				row[i] = evalPoly(tt.poly, float64(i))
			}

			filter, err := NewSavGol(SavGolConfig{tt.window, tt.polyOrder, tt.deriv}, nVars)
			if err != nil {
				t.Fatalf("building the filter failed: %v", err)
			}
			got, err := filter.Apply(row)
			if err != nil {
				t.Fatalf("applying the filter failed: %v", err)
			}

			want := derivPoly(tt.poly, tt.deriv)
			for i := range got {
				expected := evalPoly(want, float64(i))
				// Scale the tolerance with the magnitude involved: a wide window
				// raises positions to the polynomial order, so the intermediate
				// sums are large even when the answer is not.
				tol := 1e-9 * math.Max(1, math.Abs(expected))
				if math.Abs(got[i]-expected) > tol {
					t.Errorf("variable %d: got %.12g, want %.12g (difference %.3g)",
						i, got[i], expected, got[i]-expected)
				}
			}
		})
	}
}

// TestSavGolIdentityWhenPolynomialInterpolates checks the degenerate case where
// the polynomial has as many free parameters as the window has points. The fit
// then interpolates rather than smooths, so smoothing must return the input
// untouched -- for any input at all, not merely a polynomial one.
func TestSavGolIdentityWhenPolynomialInterpolates(t *testing.T) {
	rng := rand.New(rand.NewSource(20260907))
	row := make([]float64, 25)
	for i := range row {
		row[i] = rng.NormFloat64() * 10
	}

	filter, err := NewSavGol(SavGolConfig{WindowLength: 7, PolyOrder: 6, Deriv: 0}, len(row))
	if err != nil {
		t.Fatalf("building the filter failed: %v", err)
	}
	got, err := filter.Apply(row)
	if err != nil {
		t.Fatalf("applying the filter failed: %v", err)
	}
	for i := range row {
		if math.Abs(got[i]-row[i]) > 1e-9*math.Max(1, math.Abs(row[i])) {
			t.Errorf("variable %d: interpolating fit changed the value: got %.12g, want %.12g", i, got[i], row[i])
		}
	}
}

// TestSavGolCoefficientSymmetry checks a structural property that follows from
// the window being symmetric about its centre: smoothing weights are even,
// odd-derivative weights are odd. A sign or index error in the coefficient
// solve breaks this even when the magnitudes happen to look reasonable.
func TestSavGolCoefficientSymmetry(t *testing.T) {
	for _, deriv := range []int{0, 1, 2} {
		filter, err := NewSavGol(SavGolConfig{WindowLength: 9, PolyOrder: 3, Deriv: deriv}, 50)
		if err != nil {
			t.Fatalf("deriv %d: building the filter failed: %v", deriv, err)
		}
		c := filter.centre
		sign := 1.0
		if deriv%2 == 1 {
			sign = -1.0
		}
		for i := range c {
			mirror := c[len(c)-1-i]
			if math.Abs(c[i]-sign*mirror) > 1e-12 {
				t.Errorf("deriv %d: coefficient %d (%.12g) and its mirror (%.12g) violate the expected symmetry",
					deriv, i, c[i], mirror)
			}
		}
		// Smoothing must preserve a constant, so its weights sum to one; every
		// derivative annihilates a constant, so theirs sum to zero.
		var sum float64
		for _, v := range c {
			sum += v
		}
		want := 0.0
		if deriv == 0 {
			want = 1.0
		}
		if math.Abs(sum-want) > 1e-12 {
			t.Errorf("deriv %d: coefficients sum to %.12g, want %.12g", deriv, sum, want)
		}
	}
}

// TestSavGolApplyTransposeMatchesApply checks S^T against S through the identity
// <Sx, v> = <x, S^T v>, which holds for every x and v exactly when the transpose
// is the transpose. The PCR original-scale collapse depends on this, and an
// error in it would surface as regression coefficients that are wrong in a way
// no plot would reveal.
func TestSavGolApplyTransposeMatchesApply(t *testing.T) {
	rng := rand.New(rand.NewSource(4711))
	const nVars = 37

	for _, cfg := range []SavGolConfig{
		{WindowLength: 5, PolyOrder: 2, Deriv: 0},
		{WindowLength: 11, PolyOrder: 2, Deriv: 1},
		{WindowLength: 15, PolyOrder: 3, Deriv: 2},
	} {
		filter, err := NewSavGol(cfg, nVars)
		if err != nil {
			t.Fatalf("%+v: building the filter failed: %v", cfg, err)
		}
		for trial := 0; trial < 5; trial++ {
			x := make([]float64, nVars)
			v := make([]float64, nVars)
			for i := range x {
				x[i] = rng.NormFloat64()
				v[i] = rng.NormFloat64()
			}
			sx, err := filter.Apply(x)
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}
			stv, err := filter.ApplyTranspose(v)
			if err != nil {
				t.Fatalf("ApplyTranspose failed: %v", err)
			}
			left, right := dot(sx, v), dot(x, stv)
			if math.Abs(left-right) > 1e-9*math.Max(1, math.Abs(left)) {
				t.Errorf("%+v: <Sx,v> = %.12g but <x,S^T v> = %.12g", cfg, left, right)
			}
		}
	}
}

func TestSavGolConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SavGolConfig
		nVars    int
		wantErr  bool
		contains string
	}{
		{"valid", SavGolConfig{11, 2, 1}, 100, false, ""},
		{"window too small", SavGolConfig{1, 0, 0}, 100, true, "at least 3"},
		{"even window", SavGolConfig{10, 2, 1}, 100, true, "must be odd"},
		{"order not below window", SavGolConfig{5, 5, 0}, 100, true, "must be less than the window length"},
		{"negative order", SavGolConfig{5, -1, 0}, 100, true, "must not be negative"},
		{"negative deriv", SavGolConfig{5, 2, -1}, 100, true, "must not be negative"},
		{"deriv above order", SavGolConfig{7, 2, 3}, 100, true, "identically zero"},
		{"window wider than spectrum", SavGolConfig{101, 2, 1}, 50, true, "exceeds the number of variables"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate(tt.nVars)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				if !strings.Contains(err.Error(), tt.contains) {
					t.Errorf("error %q does not mention %q", err.Error(), tt.contains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestSavGolRefusesMissingValues checks that a hole is refused rather than
// smeared. A single NaN inside a window would poison WindowLength outputs, and
// the result would still look like a spectrum.
func TestSavGolRefusesMissingValues(t *testing.T) {
	row := make([]float64, 20)
	for i := range row {
		row[i] = float64(i)
	}
	row[7] = math.NaN()

	filter, err := NewSavGol(SavGolConfig{WindowLength: 5, PolyOrder: 2, Deriv: 1}, len(row))
	if err != nil {
		t.Fatalf("building the filter failed: %v", err)
	}
	if _, err := filter.Apply(row); err == nil {
		t.Fatal("expected Apply to refuse a row containing NaN, but it returned a result")
	} else if !strings.Contains(err.Error(), "missing value") {
		t.Errorf("error should name the missing value, got: %v", err)
	}
}

func TestSavGolRejectsWrongLengthInput(t *testing.T) {
	filter, err := NewSavGol(SavGolConfig{WindowLength: 5, PolyOrder: 2, Deriv: 0}, 20)
	if err != nil {
		t.Fatalf("building the filter failed: %v", err)
	}
	if _, err := filter.Apply(make([]float64, 19)); err == nil {
		t.Error("Apply accepted a row of the wrong length")
	}
	if _, err := filter.ApplyTranspose(make([]float64, 21)); err == nil {
		t.Error("ApplyTranspose accepted a vector of the wrong length")
	}
}
