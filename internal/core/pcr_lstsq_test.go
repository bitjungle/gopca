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

	"gonum.org/v1/gonum/mat"
)

// randomDesign builds an n×c design matrix and a response, deterministically.
func randomDesign(n, c int, seed uint64) (*mat.Dense, []float64) {
	r := rand.New(rand.NewPCG(seed, 0x2545F4914F6CDD1D))
	m := mat.NewDense(n, c, nil)
	for i := 0; i < n; i++ {
		m.Set(i, 0, 1) // intercept column, as PCR uses
		for j := 1; j < c; j++ {
			m.Set(i, j, r.NormFloat64())
		}
	}
	y := make([]float64, n)
	for i := range y {
		y[i] = r.NormFloat64()
	}
	return m, y
}

// solveWithGonum is an independent reference: it factorizes the first k columns
// from scratch using gonum's QR and solves. If our incremental solver disagrees
// with it, one of the two is wrong.
func solveWithGonum(t *testing.T, m *mat.Dense, y []float64, k int) []float64 {
	t.Helper()
	n, _ := m.Dims()
	sub := mat.NewDense(n, k, nil)
	for j := 0; j < k; j++ {
		for i := 0; i < n; i++ {
			sub.Set(i, j, m.At(i, j))
		}
	}
	var qr mat.QR
	qr.Factorize(sub)
	b := mat.NewVecDense(k, nil)
	if err := qr.SolveVecTo(b, false, mat.NewVecDense(n, y)); err != nil {
		t.Fatalf("gonum QR solve for k=%d: %v", k, err)
	}
	out := make([]float64, k)
	for i := range out {
		out[i] = b.AtVec(i)
	}
	return out
}

// TestNestedLeastSquaresMatchesGonum is the load-bearing check for the
// estimator. The incremental solver reuses one factorization across every
// candidate component count; gonum refactorizes each subset independently. They
// must agree, and if the nested-prefix reasoning is wrong they will not.
func TestNestedLeastSquaresMatchesGonum(t *testing.T) {
	const n, c = 60, 12
	m, y := randomDesign(n, c, 11)

	solver, err := newNestedLeastSquares(m, y)
	if err != nil {
		t.Fatalf("newNestedLeastSquares: %v", err)
	}
	if solver.Rank() != c {
		t.Fatalf("rank = %d, want %d for a random design", solver.Rank(), c)
	}

	for k := 1; k <= c; k++ {
		got, err := solver.Coefficients(k)
		if err != nil {
			t.Fatalf("Coefficients(%d): %v", k, err)
		}
		want := solveWithGonum(t, m, y, k)
		for i := range want {
			if math.Abs(got[i]-want[i]) > 1e-9*(1+math.Abs(want[i])) {
				t.Errorf("k=%d coefficient %d: got %.12g, gonum %.12g", k, i, got[i], want[i])
			}
		}
	}
}

// TestNestedLeastSquaresFittedAccumulates checks the running-sum identity that
// makes the all-k sweep cheap: the fitted values for k columns are the sum of
// the projections onto the first k orthonormal directions, and must equal the
// design matrix times the coefficients for that k.
func TestNestedLeastSquaresFittedAccumulates(t *testing.T) {
	const n, c = 40, 8
	m, y := randomDesign(n, c, 23)

	solver, err := newNestedLeastSquares(m, y)
	if err != nil {
		t.Fatalf("newNestedLeastSquares: %v", err)
	}

	accumulated := make([]float64, n)
	for k := 1; k <= c; k++ {
		if err := solver.FittedInto(accumulated, k-1); err != nil {
			t.Fatalf("FittedInto(%d): %v", k-1, err)
		}

		coef, err := solver.Coefficients(k)
		if err != nil {
			t.Fatalf("Coefficients(%d): %v", k, err)
		}
		for i := 0; i < n; i++ {
			var direct float64
			for j := 0; j < k; j++ {
				direct += m.At(i, j) * coef[j]
			}
			if math.Abs(accumulated[i]-direct) > 1e-9*(1+math.Abs(direct)) {
				t.Fatalf("k=%d row %d: accumulated %.12g, direct %.12g", k, i, accumulated[i], direct)
			}
		}
	}
}

// TestNestedLeastSquaresOrthogonality guards the reorthogonalization pass. A
// single Gram-Schmidt sweep loses orthogonality on collinear designs, and the
// resulting coefficients drift silently rather than failing.
func TestNestedLeastSquaresOrthogonality(t *testing.T) {
	const n, c = 50, 10
	m, y := randomDesign(n, c, 5)

	// Make the design badly conditioned: each column is mostly a copy of the one
	// before it, which is what strongly correlated spectra look like.
	for j := 2; j < c; j++ {
		for i := 0; i < n; i++ {
			m.Set(i, j, m.At(i, j-1)+1e-6*m.At(i, j))
		}
	}

	solver, err := newNestedLeastSquares(m, y)
	if err != nil {
		t.Fatalf("newNestedLeastSquares: %v", err)
	}

	for a := 0; a < solver.Rank(); a++ {
		for b := a; b < solver.Rank(); b++ {
			var dot float64
			for i := 0; i < n; i++ {
				dot += solver.q.At(i, a) * solver.q.At(i, b)
			}
			want := 0.0
			if a == b {
				want = 1.0
			}
			if math.Abs(dot-want) > 1e-10 {
				t.Errorf("q_%d . q_%d = %.3g, want %.0f", a, b, dot, want)
			}
		}
	}
}

// TestNestedLeastSquaresRankDeficiency checks that an exactly dependent column
// is detected rather than producing an arbitrary member of a solution set.
func TestNestedLeastSquaresRankDeficiency(t *testing.T) {
	const n, c = 30, 6
	m, y := randomDesign(n, c, 77)

	// Column 4 is an exact duplicate of column 2.
	for i := 0; i < n; i++ {
		m.Set(i, 4, m.At(i, 2))
	}

	solver, err := newNestedLeastSquares(m, y)
	if err != nil {
		t.Fatalf("newNestedLeastSquares: %v", err)
	}
	if solver.Rank() != 4 {
		t.Errorf("rank = %d, want 4 (the duplicate column is the fifth)", solver.Rank())
	}
	if _, err := solver.Coefficients(5); err == nil {
		t.Error("expected an error when solving beyond the detected rank")
	}
	if _, err := solver.Coefficients(4); err != nil {
		t.Errorf("solving within the rank should succeed: %v", err)
	}
}

func TestNestedLeastSquaresErrors(t *testing.T) {
	m, y := randomDesign(10, 3, 1)

	if _, err := newNestedLeastSquares(m, y[:5]); err == nil {
		t.Error("expected an error for mismatched response length")
	}
	if _, err := newNestedLeastSquares(mat.NewDense(10, 1, nil), y); err == nil {
		t.Error("expected an error for an all-zero design column")
	}

	solver, err := newNestedLeastSquares(m, y)
	if err != nil {
		t.Fatalf("newNestedLeastSquares: %v", err)
	}
	if _, err := solver.Coefficients(-1); err == nil {
		t.Error("expected an error for a negative column count")
	}
	if err := solver.FittedInto(make([]float64, 3), 0); err == nil {
		t.Error("expected an error for a wrongly sized destination")
	}
}
