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
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// issue736Data is a small fixed dataset used for the kernel-PCA score-scaling
// regression tests (#736).
var issue736Data = types.Matrix{
	{0.0, 0.0},
	{1.0, 0.0},
	{0.0, 1.0},
	{1.0, 1.0},
	{2.0, 2.0},
	{0.5, 1.5},
}

// TestIssue736_ScoresMatchSklearn validates kernel-PCA training scores against
// scikit-learn's KernelPCA.fit_transform, the ground-truth reference.
//
// The kernel-PCA score of a training point on component i is √λ_i·v_i (Schölkopf,
// Smola & Müller 1998). The pre-fix code returned v_i/√λ_i (the expansion
// coefficient α), off by a factor of λ per component.
//
// Reference generated with scikit-learn:
//
//	KernelPCA(n_components=2, kernel="rbf", gamma=0.5).fit_transform(X)
//
// Our RBF kernel is exp(-γ‖x-y‖²), identical to scikit-learn's, and neither
// implementation centers the input X (kernel centering only). Eigenvector sign is
// arbitrary per component, so each component is sign-aligned before comparison.
func TestIssue736_ScoresMatchSklearn(t *testing.T) {
	sklearn := types.Matrix{
		{-0.5354526141, -0.2727148707},
		{-0.3710946418, -0.4489715871},
		{-0.2202052520, 0.4365759904},
		{0.1191983013, 0.1469531938},
		{0.8308467781, -0.3950694946},
		{0.1767074284, 0.5332267682},
	}
	// scikit-learn eigenvalues_ for the same fit (eigenvalues of the centered
	// kernel matrix); our ExplainedVar should match these.
	sklearnEigenvalues := []float64{1.2086512071, 0.9285534151}

	engine := NewKernelPCAEngine()
	result, err := engine.Fit(issue736Data, types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
	})
	if err != nil {
		t.Fatalf("Fit failed: %v", err)
	}

	// Eigenvalues must match scikit-learn (validates the eigendecomposition path).
	for c := 0; c < 2; c++ {
		if math.Abs(result.ExplainedVar[c]-sklearnEigenvalues[c]) > 1e-4 {
			t.Errorf("ExplainedVar[%d] = %.6f, sklearn eigenvalue = %.6f",
				c, result.ExplainedVar[c], sklearnEigenvalues[c])
		}
	}

	// Scores must match scikit-learn after per-component sign alignment.
	for c := 0; c < 2; c++ {
		sign := signAlignColumn(result.Scores, sklearn, c)
		for r := range issue736Data {
			got := sign * result.Scores[r][c]
			want := sklearn[r][c]
			if math.Abs(got-want) > 1e-6 {
				t.Errorf("score[%d][%d] = %.10f (sign-aligned), sklearn = %.10f (diff %.2e)",
					r, c, got, want, math.Abs(got-want))
			}
		}
	}
}

// TestIssue736_FitTransformMatchesTransform is a direct regression for the bug:
// FitTransform (which returns Fit's training scores) must equal Transform applied
// to the same training data. Before the fix, Fit returned √λ·v scaled by 1/λ, so
// the two disagreed by a factor of λ per component.
func TestIssue736_FitTransformMatchesTransform(t *testing.T) {
	config := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
	}
	engine := NewKernelPCAEngine()

	fit, err := engine.FitTransform(issue736Data, config)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}
	transformed, err := engine.Transform(issue736Data)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(transformed) != len(fit.Scores) {
		t.Fatalf("row count mismatch: FitTransform=%d, Transform=%d", len(fit.Scores), len(transformed))
	}
	for r := range fit.Scores {
		for c := range fit.Scores[r] {
			diff := math.Abs(fit.Scores[r][c] - transformed[r][c])
			if diff > 1e-9 {
				t.Errorf("score[%d][%d]: FitTransform=%.12f, Transform=%.12f (diff %.2e)",
					r, c, fit.Scores[r][c], transformed[r][c], diff)
			}
		}
	}
}

// signAlignColumn returns +1 or -1 so that sign*got[:,c] best matches ref[:,c],
// resolving the per-component eigenvector sign ambiguity. The sign is decided from
// the row with the largest reference magnitude in that column.
func signAlignColumn(got, ref types.Matrix, c int) float64 {
	maxAbs, pivot := -1.0, 0
	for r := range ref {
		if a := math.Abs(ref[r][c]); a > maxAbs {
			maxAbs, pivot = a, r
		}
	}
	if got[pivot][c]*ref[pivot][c] < 0 {
		return -1.0
	}
	return 1.0
}
