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
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// rankTolerance is the relative threshold below which a column is treated as
// linearly dependent on those before it. Expressed relative to the column's
// original norm so that it is scale free.
const rankTolerance = 1e-12

// nestedLeastSquares solves a family of least-squares problems that share a
// design matrix prefix: for every k, the fit of y on the first k+1 columns of M.
//
// PCR needs exactly this. Sweeping candidate component counts means fitting the
// response on the first 1, 2, ... K score columns, and the naive approach of
// refactorizing for every k costs O(nK³) overall. Because the thin QR of the
// leading k columns of M is the leading k columns of Q and the leading k×k block
// of R, one factorization serves every k, and each solution costs only a
// back-substitution. Total cost is O(nK² + K³), the same order as a single
// factorization.
//
// The general least-squares formulation is used rather than the per-component
// shortcut γⱼ = tⱼᵀy / tⱼᵀtⱼ because that shortcut is valid only when the design
// columns are orthogonal over exactly the rows being regressed. That holds when
// the decomposition and the regression use the same rows, but not when the
// decomposition also saw rows whose response was unobserved, which is the normal
// situation for the calibration data this tool targets. Solving properly keeps
// one code path correct in both cases instead of two paths that can disagree.
//
// Reference: Björck (1996), Numerical Methods for Least Squares Problems, §2.4.
type nestedLeastSquares struct {
	q    *mat.Dense // n × c, orthonormal columns
	r    *mat.Dense // c × c, upper triangular
	qty  []float64  // Qᵀy, length c
	rank int        // number of numerically independent leading columns
}

// newNestedLeastSquares factorizes M and projects y onto its column space.
//
// M is n × c and is not modified. Columns beyond the detected rank are
// unusable; callers must not request a solution wider than Rank.
//
// Algorithm complexity: O(n c²).
func newNestedLeastSquares(m *mat.Dense, y []float64) (*nestedLeastSquares, error) {
	n, c := m.Dims()
	if len(y) != n {
		return nil, fmt.Errorf("design matrix has %d rows but response has %d values", n, len(y))
	}
	if c == 0 {
		return nil, fmt.Errorf("design matrix has no columns")
	}

	q := mat.NewDense(n, c, nil)
	r := mat.NewDense(c, c, nil)

	rank := c
	column := make([]float64, n)

	for j := 0; j < c; j++ {
		mat.Col(column, j, m)

		originalNorm := 0.0
		for _, v := range column {
			originalNorm += v * v
		}
		originalNorm = math.Sqrt(originalNorm)

		// Modified Gram-Schmidt, then a second pass over the same columns.
		// Reorthogonalization costs one extra sweep and recovers the accuracy of
		// a Householder factorization; without it, orthogonality degrades badly
		// once the design columns are close to dependent, which is precisely the
		// situation with many components on collinear spectra.
		for pass := 0; pass < 2; pass++ {
			for i := 0; i < j; i++ {
				var dot float64
				for row := 0; row < n; row++ {
					dot += q.At(row, i) * column[row]
				}
				for row := 0; row < n; row++ {
					column[row] -= dot * q.At(row, i)
				}
				r.Set(i, j, r.At(i, j)+dot)
			}
		}

		norm := 0.0
		for _, v := range column {
			norm += v * v
		}
		norm = math.Sqrt(norm)

		if originalNorm == 0 || norm <= rankTolerance*originalNorm {
			// This column adds no direction the earlier ones do not already span.
			// Record where independence ran out and leave the remaining columns
			// of Q zero; solutions wider than this are refused rather than
			// returned as an arbitrary member of a solution set.
			rank = j
			break
		}

		r.Set(j, j, norm)
		for row := 0; row < n; row++ {
			q.Set(row, j, column[row]/norm)
		}
	}

	if rank == 0 {
		// Nothing in the design spans any direction at all. This cannot arise from
		// a PCR design, whose first column is the all-ones intercept, so reaching
		// here means the caller built the matrix wrongly. Fail at construction
		// rather than returning a solver that answers every query with nothing.
		return nil, fmt.Errorf("design matrix has no numerically independent columns")
	}

	qty := make([]float64, c)
	for j := 0; j < rank; j++ {
		var dot float64
		for row := 0; row < n; row++ {
			dot += q.At(row, j) * y[row]
		}
		qty[j] = dot
	}

	return &nestedLeastSquares{q: q, r: r, qty: qty, rank: rank}, nil
}

// Rank reports how many leading columns of the design were numerically
// independent.
func (s *nestedLeastSquares) Rank() int { return s.rank }

// Coefficients returns the least-squares solution using the first k columns of
// the design matrix.
//
// Algorithm complexity: O(k²).
func (s *nestedLeastSquares) Coefficients(k int) ([]float64, error) {
	if k < 0 || k > s.rank {
		return nil, fmt.Errorf("cannot solve with %d columns: only %d are independent", k, s.rank)
	}
	b := make([]float64, k)
	// Back substitution on the leading k×k block of R.
	for i := k - 1; i >= 0; i-- {
		sum := s.qty[i]
		for j := i + 1; j < k; j++ {
			sum -= s.r.At(i, j) * b[j]
		}
		diag := s.r.At(i, i)
		if diag == 0 {
			return nil, fmt.Errorf("singular triangular factor at column %d", i)
		}
		b[i] = sum / diag
	}
	return b, nil
}

// FittedInto writes the fitted values using the first k columns into dst.
//
// Fitted values accumulate across k, since projecting onto the span of the first
// k columns is the sum of the projections onto each orthonormal direction:
// ŷ_k = Σ_{j<k} qⱼ (qⱼᵀy). A caller sweeping k in increasing order can therefore
// reuse dst and add one term per step rather than recomputing, which is what
// makes an all-k sweep O(nK) rather than O(nK²).
//
// Algorithm complexity: O(n) per call when accumulating.
func (s *nestedLeastSquares) FittedInto(dst []float64, column int) error {
	if column < 0 || column >= s.rank {
		return fmt.Errorf("column %d is outside the independent range of %d", column, s.rank)
	}
	n, _ := s.q.Dims()
	if len(dst) != n {
		return fmt.Errorf("destination has %d values but the design has %d rows", len(dst), n)
	}
	coef := s.qty[column]
	for row := 0; row < n; row++ {
		dst[row] += coef * s.q.At(row, column)
	}
	return nil
}
