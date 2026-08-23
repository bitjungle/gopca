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

	"gonum.org/v1/gonum/mat"
)

// A principal component is defined only up to its sign: if p is a unit
// eigenvector of the covariance matrix then so is -p, with the same eigenvalue,
// and the pair (-t, -p) reconstructs the data exactly as well as (t, p). Nothing
// in the mathematics prefers one over the other, so the sign a component comes
// out with is decided by whichever numerical routine produced it.
//
// That is a problem when one program offers several routines. GoPCA's SVD and
// NIPALS implementations agree on every component to within NIPALS's convergence
// tolerance, yet on the Wine dataset they returned PC1, PC2, PC4 and PC5 with
// opposite signs, so switching method — for instance to handle missing values —
// mirrored the scores plot. The components were identical; only the arbitrary
// choice differed.
//
// normalizeComponentSigns removes that choice by fixing it. The rule adopted is
// the one used by scikit-learn (`svd_flip`, loadings-based) and by MATLAB's
// `pca`: within each loading vector, the element of largest magnitude is made
// positive. It is no more "correct" than any other rule — R's prcomp applies
// none at all and documents its signs as arbitrary — but it is deterministic,
// depends only on the loadings rather than on which sample happens to be
// extreme, and costs nothing to match two widely used implementations.
//
// Reference for the underlying indeterminacy and a more principled (data-driven)
// alternative: Bro, R., Acar, E., & Kolda, T. G. (2008). Resolving the sign
// ambiguity in the singular value decomposition. Journal of Chemometrics, 22(2),
// 135-140. https://doi.org/10.1002/cem.1122
//
// Both matrices are modified in place. Scores and loadings must be flipped
// together: negating one alone would break X ≈ T·Pᵀ.
//
// Algorithm complexity: O(n*k + m*k) for n samples, m variables, k components.
func normalizeComponentSigns(scores, loadings *mat.Dense) {
	if scores == nil || loadings == nil {
		return
	}
	nSamples, nScoreCols := scores.Dims()
	nVars, nComponents := loadings.Dims()
	if nScoreCols != nComponents {
		// Shapes disagree; leave the caller's data untouched rather than
		// corrupting one matrix relative to the other.
		return
	}

	for k := 0; k < nComponents; k++ {
		// Locate the loading of largest magnitude. Ties resolve to the lowest
		// variable index, which keeps the result deterministic.
		maxAbs := 0.0
		maxVal := 0.0
		for j := 0; j < nVars; j++ {
			v := loadings.At(j, k)
			if math.IsNaN(v) {
				continue
			}
			if a := math.Abs(v); a > maxAbs {
				maxAbs = a
				maxVal = v
			}
		}
		// A component whose largest loading is zero (or all-NaN) carries no
		// direction to orient; leave it alone rather than multiplying by zero.
		if maxAbs == 0 || maxVal >= 0 {
			continue
		}

		for j := 0; j < nVars; j++ {
			loadings.Set(j, k, -loadings.At(j, k))
		}
		for i := 0; i < nSamples; i++ {
			scores.Set(i, k, -scores.At(i, k))
		}
	}
}
