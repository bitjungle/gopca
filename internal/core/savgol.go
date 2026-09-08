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

// Savitzky-Golay smoothing and differentiation along the variable axis.
//
// The filter slides a window across the variables, fits a low-order polynomial
// to the points inside it by least squares, and evaluates that polynomial -- or
// one of its derivatives -- at the window's centre. Smoothing and
// differentiation happen in the same pass, which is the whole point: the
// derivative of a noisy spectrum taken by plain differencing is dominated by
// the noise, because differencing amplifies exactly the high-frequency content
// that noise lives in.
//
// In near-infrared spectroscopy this is the standard next step after scatter
// correction. A first derivative removes an additive baseline offset, a second
// derivative removes a baseline slope as well -- both are constants under
// differentiation and simply vanish. That is why derivative preprocessing
// reaches the same predictive accuracy with markedly fewer components.
//
// References:
//   - Savitzky, A. & Golay, M.J.E. (1964). Smoothing and Differentiation of Data
//     by Simplified Least Squares Procedures. Analytical Chemistry 36(8),
//     1627-1639.
//   - Gorry, P.A. (1990). General least-squares smoothing and differentiation by
//     the convolution (Savitzky-Golay) method. Analytical Chemistry 62(6),
//     570-573. Covers evaluation away from the window centre, which is what the
//     edge handling below rests on.
//   - Rinnan, Å., van den Berg, F. & Engelsen, S.B. (2009). Review of the most
//     common pre-processing techniques for near-infrared spectra. TrAC Trends in
//     Analytical Chemistry 28(10), 1201-1222.

// float64Epsilon is the gap between 1 and the next representable float64. The
// standard library does not export it, and it is the scale NumPy and SciPy use
// when deciding which singular values carry information.
const float64Epsilon = 2.220446049250313e-16

// SavGolConfig describes a Savitzky-Golay filter.
//
// Spacing between variables is taken to be uniform and equal to one. Exposing a
// physical spacing would be misleading here: it enters the d-th derivative only
// as the constant factor 1/spacing^d, identical for every variable, which PCA is
// insensitive to -- standard scaling absorbs it entirely, and any other column
// scaling rescales every score by the same amount without moving a single point
// relative to another. A knob that cannot change a result should not be offered.
type SavGolConfig struct {
	// WindowLength is the number of variables in the sliding window. Must be
	// odd, at least 3, and greater than PolyOrder.
	WindowLength int
	// PolyOrder is the degree of the polynomial fitted within each window.
	PolyOrder int
	// Deriv is the derivative order to evaluate: 0 smooths without
	// differentiating. Any order up to and including PolyOrder is accepted,
	// matching scipy.signal.savgol_filter; 0, 1 and 2 are the ones reached for
	// in practice, since a spectrum rarely carries usable structure in its
	// third derivative.
	Deriv int
}

// Validate reports whether the configuration is usable for a spectrum of
// nVars variables.
func (c SavGolConfig) Validate(nVars int) error {
	if err := c.ValidateShape(); err != nil {
		return err
	}
	if c.WindowLength > nVars {
		return fmt.Errorf("Savitzky-Golay window length (%d) exceeds the number of variables (%d)",
			c.WindowLength, nVars)
	}
	return nil
}

// ValidateShape checks everything that can be judged without knowing how wide
// the data is, so a command line can reject a nonsensical combination before it
// has opened the file.
func (c SavGolConfig) ValidateShape() error {
	if c.WindowLength < 3 {
		return fmt.Errorf("Savitzky-Golay window length must be at least 3, got %d", c.WindowLength)
	}
	if c.WindowLength%2 == 0 {
		return fmt.Errorf("Savitzky-Golay window length must be odd so the window has a centre, got %d", c.WindowLength)
	}
	if c.PolyOrder < 0 {
		return fmt.Errorf("Savitzky-Golay polynomial order must not be negative, got %d", c.PolyOrder)
	}
	if c.PolyOrder >= c.WindowLength {
		return fmt.Errorf("Savitzky-Golay polynomial order (%d) must be less than the window length (%d)",
			c.PolyOrder, c.WindowLength)
	}
	if c.Deriv < 0 {
		return fmt.Errorf("Savitzky-Golay derivative order must not be negative, got %d", c.Deriv)
	}
	// A polynomial of degree p has no non-zero derivative beyond the p-th, so
	// this request has exactly one answer: zero everywhere. Returning that
	// silently would look like a filter that had destroyed the data rather than
	// a configuration that could not mean anything.
	if c.Deriv > c.PolyOrder {
		return fmt.Errorf("Savitzky-Golay derivative order (%d) must not exceed the polynomial order (%d): "+
			"a degree-%d polynomial has no non-zero derivative of order %d, so the result would be zero everywhere",
			c.Deriv, c.PolyOrder, c.PolyOrder, c.Deriv)
	}
	return nil
}

// SavGol is a Savitzky-Golay filter compiled for a fixed number of variables.
//
// The filter is a fixed linear operator on the variable axis: it depends only on
// the configuration and the variable count, never on the data. That is what
// separates it from SNV and vector normalisation, which scale each sample by a
// statistic of that same sample. Because the operator is fixed, it can be
// transposed, and a regression fitted on filtered variables can be collapsed
// back to the original ones.
type SavGol struct {
	cfg SavGolConfig
	n   int

	// centre holds the coefficients applied to a window centred on the output
	// variable; it serves every interior position.
	centre []float64
	// left[i] and right[i] hold the coefficients for the i-th output position
	// from each end, where no centred window fits. Each is applied to the first
	// (respectively last) WindowLength variables.
	left  [][]float64
	right [][]float64
}

// NewSavGol compiles the filter for a spectrum of nVars variables.
func NewSavGol(cfg SavGolConfig, nVars int) (*SavGol, error) {
	if err := cfg.Validate(nVars); err != nil {
		return nil, err
	}

	half := cfg.WindowLength / 2

	centre, err := savGolCoeffs(cfg.WindowLength, cfg.PolyOrder, cfg.Deriv, float64(half))
	if err != nil {
		return nil, err
	}

	// Near the ends there is no centred window, and padding the spectrum would
	// invent measurements that were never made -- at exactly the wavelengths
	// where absorption bands most often sit. Instead fit the polynomial to the
	// end window and evaluate it at the edge positions, which is what
	// scipy.signal.savgol_filter does under its default mode="interp" and what
	// spectroscopists therefore expect.
	left := make([][]float64, half)
	right := make([][]float64, half)
	for i := 0; i < half; i++ {
		if left[i], err = savGolCoeffs(cfg.WindowLength, cfg.PolyOrder, cfg.Deriv, float64(i)); err != nil {
			return nil, err
		}
		// Output position n-half+i sits at offset WindowLength-half+i inside the
		// final window.
		pos := float64(cfg.WindowLength - half + i)
		if right[i], err = savGolCoeffs(cfg.WindowLength, cfg.PolyOrder, cfg.Deriv, pos); err != nil {
			return nil, err
		}
	}

	return &SavGol{cfg: cfg, n: nVars, centre: centre, left: left, right: right}, nil
}

// Config returns the configuration the filter was compiled from.
func (s *SavGol) Config() SavGolConfig { return s.cfg }

// NumVars returns the variable count the filter was compiled for.
func (s *SavGol) NumVars() int { return s.n }

// Apply filters one sample, returning a new slice.
//
// Complexity: O(nVars * WindowLength).
func (s *SavGol) Apply(row []float64) ([]float64, error) {
	if len(row) != s.n {
		return nil, fmt.Errorf("Savitzky-Golay filter was built for %d variables but received %d", s.n, len(row))
	}
	// A window spanning a hole would spread that hole across WindowLength
	// output variables, and a derivative across a hole means nothing at all.
	// Refusing beats returning a spectrum quietly poisoned near every gap.
	for j, v := range row {
		if math.IsNaN(v) {
			return nil, fmt.Errorf("Savitzky-Golay cannot filter a sample with a missing value "+
				"(variable %d): impute or drop missing values first", j)
		}
	}

	half := s.cfg.WindowLength / 2
	out := make([]float64, s.n)

	for i := 0; i < half; i++ {
		out[i] = dot(s.left[i], row[:s.cfg.WindowLength])
	}
	for i := half; i < s.n-half; i++ {
		out[i] = dot(s.centre, row[i-half:i-half+s.cfg.WindowLength])
	}
	for i := 0; i < half; i++ {
		out[s.n-half+i] = dot(s.right[i], row[s.n-s.cfg.WindowLength:])
	}
	return out, nil
}

// ApplyTranspose computes S^T v, where S is the filter's matrix.
//
// This is what collapses a model fitted on filtered variables back onto the
// original ones: if a linear model predicts from Sx, then the same predictions
// are given by (S^T beta) . x. The transpose is formed from the same
// coefficients rather than by building S, so it costs the same as Apply.
func (s *SavGol) ApplyTranspose(v []float64) ([]float64, error) {
	if len(v) != s.n {
		return nil, fmt.Errorf("Savitzky-Golay filter was built for %d variables but received %d", s.n, len(v))
	}

	half := s.cfg.WindowLength / 2
	w := s.cfg.WindowLength
	out := make([]float64, s.n)

	for i := 0; i < half; i++ {
		scatter(out[:w], s.left[i], v[i])
	}
	for i := half; i < s.n-half; i++ {
		scatter(out[i-half:i-half+w], s.centre, v[i])
	}
	for i := 0; i < half; i++ {
		scatter(out[s.n-w:], s.right[i], v[s.n-half+i])
	}
	return out, nil
}

// dot returns the inner product of two equal-length slices.
func dot(coeffs, values []float64) float64 {
	var sum float64
	for k, c := range coeffs {
		sum += c * values[k]
	}
	return sum
}

// scatter adds scale*coeffs into dst.
func scatter(dst, coeffs []float64, scale float64) {
	for k, c := range coeffs {
		dst[k] += c * scale
	}
}

// savGolCoeffs returns the convolution coefficients that evaluate the deriv-th
// derivative of the least-squares polynomial at position pos within a window of
// windowLength points, where positions are numbered 0..windowLength-1.
//
// The coefficients are found from the moment conditions rather than by forming a
// fit and differentiating it. Measure position from the centre of the window,
// u_i = i - (windowLength-1)/2. A vector h reproduces the deriv-th derivative at
// the output position exactly for every polynomial of degree at most polyOrder
// precisely when
//
//	sum_i h_i * u_i^j = d^deriv/du^deriv (u^j) evaluated at u = pos - centre,
//
// for j = 0..polyOrder. That is an underdetermined linear system, and its
// minimum-norm solution is the Savitzky-Golay filter. Writing it this way keeps
// the centre and the edges as one formula -- only the evaluation point moves --
// and it is the same characterisation scipy.signal.savgol_coeffs solves.
//
// The basis is centred on the window rather than on the evaluation point, and
// that choice is numerical, not cosmetic. Centring on the evaluation point makes
// the coordinates run from -(windowLength-1) to 0 at the outermost edge
// position; raised to the polynomial order those reach 8e5 for a 31-point window
// of order 4, and the resulting Vandermonde is ill-conditioned enough to cost
// several digits. Measured against scipy on a cubic, where the answer is exact
// and the discrepancy is therefore all error, the worst variable moved:
//
//	centred on the evaluation point   2.8e-10
//	centred on the window             2.1e-11
//	plus row normalisation below      6.4e-13  (worst over every case)
func savGolCoeffs(windowLength, polyOrder, deriv int, pos float64) ([]float64, error) {
	rows := polyOrder + 1
	centre := float64(windowLength-1) / 2

	// a[j][i] = u_i^j
	a := mat.NewDense(rows, windowLength, nil)
	for i := 0; i < windowLength; i++ {
		u := float64(i) - centre
		p := 1.0
		for j := 0; j < rows; j++ {
			a.Set(j, i, p)
			p *= u
		}
	}

	// The deriv-th derivative of u^j is j!/(j-deriv)! * u^(j-deriv), and it
	// vanishes for j < deriv. At the window centre only the j = deriv term
	// survives, leaving the familiar deriv! on its own.
	target := pos - centre
	rhs := mat.NewVecDense(rows, nil)
	for j := deriv; j < rows; j++ {
		falling := 1.0
		for k := 0; k < deriv; k++ {
			falling *= float64(j - k)
		}
		rhs.SetVec(j, falling*math.Pow(target, float64(j-deriv)))
	}

	// Scale each equation to unit norm. Row j has entries u_i^j, so the rows
	// span many orders of magnitude for a wide window -- and it is the ratio
	// between them, not their absolute size, that the SVD has to resolve.
	// Dividing a row of A and the matching entry of the right-hand side by the
	// same number leaves the solution set untouched, so this is exact rather
	// than approximate; it is the same normalisation numpy.polyfit applies to
	// its design matrix, and it is why SciPy is accurate on the hardest case
	// here. A row cannot be all zeros: row 0 is all ones.
	for j := 0; j < rows; j++ {
		var norm float64
		for i := 0; i < windowLength; i++ {
			v := a.At(j, i)
			norm += v * v
		}
		norm = math.Sqrt(norm)
		for i := 0; i < windowLength; i++ {
			a.Set(j, i, a.At(j, i)/norm)
		}
		rhs.SetVec(j, rhs.AtVec(j)/norm)
	}

	// Minimum-norm least-squares solution via SVD, matching what
	// scipy.linalg.lstsq does inside savgol_coeffs. Normal equations would be a
	// Hankel matrix of power sums -- Hilbert-like, and needlessly ill-conditioned
	// for the wide windows spectroscopists reach for.
	var svd mat.SVD
	if ok := svd.Factorize(a, mat.SVDThin); !ok {
		return nil, fmt.Errorf("Savitzky-Golay coefficients could not be computed: "+
			"the SVD of the %dx%d design matrix did not converge", rows, windowLength)
	}

	// Drop singular values that carry no information, on the same relative
	// threshold NumPy and SciPy use by default.
	values := svd.Values(nil)
	rank := 0
	if len(values) > 0 {
		tol := float64(max(rows, windowLength)) * values[0] * float64Epsilon
		for _, sv := range values {
			if sv > tol {
				rank++
			}
		}
	}
	if rank == 0 {
		return nil, fmt.Errorf("Savitzky-Golay coefficients could not be computed: design matrix has rank 0")
	}

	var sol mat.VecDense
	svd.SolveVecTo(&sol, rhs, rank)

	coeffs := make([]float64, windowLength)
	for i := range coeffs {
		coeffs[i] = sol.AtVec(i)
	}
	return coeffs, nil
}
