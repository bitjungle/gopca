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

// TestIssue793_CorrelationsAreNotLoadings is the defect. The Circle of
// Correlations drew pcaResult.loadings under a label promising correlations.
// The two differ by sqrt(lambda)/sd, and on standardised data every leading
// eigenvalue exceeds 1, so the substitution shortens every arrow and puts the
// unit circle out of reach.
func TestIssue793_CorrelationsAreNotLoadings(t *testing.T) {
	data := correlationFixture(60, 5)
	res, err := NewPCAEngine().Fit(data, types.PCAConfig{
		Method: "svd", Components: 3, MeanCenter: true, StandardScale: true,
	})
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	if len(res.VariableCorrelations) == 0 {
		t.Fatal("no variable correlations produced")
	}

	differs := false
	for j := range res.Loadings {
		for k := range res.Loadings[j] {
			if math.Abs(res.Loadings[j][k]-res.VariableCorrelations[j][k]) > 1e-6 {
				differs = true
			}
		}
	}
	if !differs {
		t.Error("correlations are identical to loadings; the scaling is missing")
	}

	// On standardised data r = p*sqrt(lambda), and lambda > 1 for the leading
	// components, so correlations must be the larger of the two.
	for k := 0; k < 3; k++ {
		if res.ExplainedVar[k] <= 1 {
			continue
		}
		for j := range res.Loadings {
			if math.Abs(res.VariableCorrelations[j][k]) < math.Abs(res.Loadings[j][k])-1e-9 {
				t.Errorf("component %d, variable %d: |r|=%.4f is smaller than |loading|=%.4f "+
					"despite eigenvalue %.3f > 1", k+1, j,
					math.Abs(res.VariableCorrelations[j][k]), math.Abs(res.Loadings[j][k]), res.ExplainedVar[k])
			}
		}
	}
}

// TestIssue793_MatchesTheClosedForm checks the direct computation against
// r_jk = p_jk * sqrt(lambda_k) / s_j, the standard identity. Two independent
// routes to the same number: if either is wrong they will disagree.
func TestIssue793_MatchesTheClosedForm(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config types.PCAConfig
	}{
		{"standardised", types.PCAConfig{MeanCenter: true, StandardScale: true}},
		{"mean centred only", types.PCAConfig{MeanCenter: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := correlationFixture(80, 4)
			cfg := tc.config
			cfg.Method, cfg.Components = "svd", 4

			engine := NewPCAEngine().(*PCAImpl)
			res, err := engine.Fit(data, cfg)
			if err != nil {
				t.Fatalf("fit failed: %v", err)
			}
			pre := res.PreprocessedData
			n := len(pre)

			for j := range res.Loadings {
				// sd of variable j in the matrix that was decomposed
				mean := 0.0
				for i := 0; i < n; i++ {
					mean += pre[i][j]
				}
				mean /= float64(n)
				ss := 0.0
				for i := 0; i < n; i++ {
					ss += (pre[i][j] - mean) * (pre[i][j] - mean)
				}
				sd := math.Sqrt(ss / float64(n-1))

				for k := range res.Loadings[j] {
					want := res.Loadings[j][k] * math.Sqrt(res.ExplainedVar[k]) / sd
					if got := res.VariableCorrelations[j][k]; math.Abs(got-want) > 1e-9 {
						t.Errorf("variable %d, PC%d: got %.10f, closed form gives %.10f",
							j, k+1, got, want)
					}
				}
			}
		})
	}
}

// TestIssue793_CommunalityIdentity is what makes the unit circle mean something:
// summed over every component, a variable's squared correlations total 1. So the
// squared length of an arrow in a two-component plot is the share of that
// variable's variance those components capture.
func TestIssue793_CommunalityIdentity(t *testing.T) {
	data := correlationFixture(70, 4)
	res, err := NewPCAEngine().Fit(data, types.PCAConfig{
		Method: "svd", Components: 4, MeanCenter: true, StandardScale: true,
	})
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	for j := range res.VariableCorrelations {
		total := 0.0
		for _, r := range res.VariableCorrelations[j] {
			total += r * r
		}
		if math.Abs(total-1) > 1e-9 {
			t.Errorf("variable %d: squared correlations sum to %.10f, want 1", j, total)
		}
	}
}

// TestIssue793_NoCorrelationsWithoutAPreprocessedMatrix covers the case the
// frontend now refuses to draw rather than falling back to loadings.
func TestIssue793_NoCorrelationsWithoutAPreprocessedMatrix(t *testing.T) {
	data := correlationFixture(40, 4)
	data[6][2] = math.NaN()
	res, err := NewPCAEngine().Fit(data, types.PCAConfig{
		Method: "nipals", Components: 2, MeanCenter: true,
		MissingStrategy: types.MissingNative,
	})
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	if res.PreprocessedData != nil {
		t.Fatal("fixture no longer exercises the nil-preprocessed-matrix path")
	}
	if len(res.VariableCorrelations) != 0 {
		t.Error("correlations reported without a matrix to correlate against")
	}
}

// TestIssue793_ConstantVariableYieldsZero guards the degenerate case: a variable
// with no variance has no correlation, and zero plots as no arrow rather than as
// a spurious direction.
func TestIssue793_ConstantVariableYieldsZero(t *testing.T) {
	data := correlationFixture(30, 3)
	for i := range data {
		data[i][1] = 7.0
	}
	pre := types.Matrix{}
	for _, row := range data {
		pre = append(pre, append([]float64(nil), row...))
	}
	scores := types.Matrix{}
	for i := range data {
		scores = append(scores, []float64{data[i][0], data[i][2]})
	}
	got, err := CalculateVariableCorrelations(pre, scores)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for k, r := range got[1] {
		if r != 0 {
			t.Errorf("constant variable, PC%d: got %v, want 0", k+1, r)
		}
	}
}

func correlationFixture(n, m int) types.Matrix {
	data := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		row := make([]float64, m)
		f1 := math.Sin(float64(i) * 0.13)
		f2 := math.Cos(float64(i) * 0.09)
		for j := 0; j < m; j++ {
			w := float64(j+1) / float64(m)
			row[j] = 10*w*f1 + (2-w)*f2 + 0.3*math.Sin(float64(i*(j+2))*0.41)
		}
		data[i] = row
	}
	return data
}
