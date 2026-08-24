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

	"github.com/bitjungle/gopca/pkg/types"
)

// CalculateVariableCorrelations returns the Pearson correlation between each
// preprocessed variable and each principal component, as [variables][components].
//
// This is the quantity a Circle of Correlations displays, and it is not the same
// as the loadings. Writing p_jk for the loading of variable j on component k,
// lambda_k for the component's variance and s_j for the standard deviation of
// variable j in the matrix that was decomposed:
//
//	r_jk = p_jk * sqrt(lambda_k) / s_j
//
// which follows from cov(x_j, t_k) = lambda_k * p_jk and sd(t_k) = sqrt(lambda_k).
// On standardised data s_j = 1 and every leading lambda_k exceeds 1, so the
// correlations are strictly larger than the loadings — plotting loadings in their
// place makes every arrow too short and puts the unit circle out of reach.
//
// Summing r_jk^2 across components gives variable j's communality — the share of
// its variance those components capture. Over the FULL basis that sum is exactly
// 1; over the retained components it is less, and callers should not treat the
// returned rows as unit vectors. A truncated fit returns a truncated sum.
//
// This is what gives arrow length its meaning: the squared length of a variable's
// arrow in a two-component plot is the fraction of that variable's variance those
// two components capture. An arrow at the unit circle is fully represented; one at
// 0.707 has half its variance there.
//
// Rather than assemble the formula from stored eigenvalues and standard
// deviations, this computes the correlation directly from the two matrices it is
// defined over. That is one fewer place for a scaling factor to go missing, it
// needs no special case for the preprocessing in force, and it is verifiable
// against any statistics package.
//
// References: Jolliffe, I. T. (2002). Principal Component Analysis (2nd ed.),
// Springer, on the correlation between a variable and a component; Abdi, H. &
// Williams, L. J. (2010). Principal component analysis. WIREs Computational
// Statistics, 2(4), 433-459, on the distinction between loadings and correlations
// and on the correlation circle.
//
// Algorithm complexity: O(n*m*k) for n samples, m variables, k components.
func CalculateVariableCorrelations(preprocessed, scores types.Matrix) (types.Matrix, error) {
	if len(preprocessed) == 0 || len(scores) == 0 {
		return nil, fmt.Errorf("variable correlations need both a preprocessed matrix and scores")
	}
	if len(preprocessed) != len(scores) {
		return nil, fmt.Errorf("row count mismatch: preprocessed has %d, scores has %d",
			len(preprocessed), len(scores))
	}
	n := len(preprocessed)
	if n < 2 {
		return nil, fmt.Errorf("correlations need at least 2 samples, got %d", n)
	}
	m := len(preprocessed[0])
	k := len(scores[0])

	// Column means and centred sums of squares, computed once per matrix.
	varMean, varSS := columnStats(preprocessed, n, m)
	scoreMean, scoreSS := columnStats(scores, n, k)

	out := make(types.Matrix, m)
	for j := 0; j < m; j++ {
		out[j] = make([]float64, k)
		for c := 0; c < k; c++ {
			// A constant variable or a degenerate component has no correlation to
			// report. Zero is the honest answer: it plots as no arrow at all,
			// rather than as a spurious direction.
			if varSS[j] <= 0 || scoreSS[c] <= 0 {
				continue
			}
			cov := 0.0
			for i := 0; i < n; i++ {
				cov += (preprocessed[i][j] - varMean[j]) * (scores[i][c] - scoreMean[c])
			}
			r := cov / math.Sqrt(varSS[j]*scoreSS[c])
			// Guard the boundary: accumulated rounding can carry |r| a few ulps
			// past 1, which would draw an arrow outside a circle it should touch.
			out[j][c] = math.Max(-1, math.Min(1, r))
		}
	}
	return out, nil
}

// columnStats returns the mean and centred sum of squares of each column.
func columnStats(data types.Matrix, n, cols int) (means, ss []float64) {
	means = make([]float64, cols)
	ss = make([]float64, cols)
	for c := 0; c < cols; c++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			sum += data[i][c]
		}
		means[c] = sum / float64(n)
	}
	for c := 0; c < cols; c++ {
		for i := 0; i < n; i++ {
			d := data[i][c] - means[c]
			ss[c] += d * d
		}
	}
	return means, ss
}
