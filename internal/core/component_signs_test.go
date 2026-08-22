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

// TestIssue779_SVDAndNIPALSAgreeOnSign guards the defect this convention exists
// for: on Wine, SVD and NIPALS returned PC1, PC2, PC4 and PC5 with opposite
// signs, so switching method mirrored the scores plot for identical components.
func TestIssue779_SVDAndNIPALSAgreeOnSign(t *testing.T) {
	data := syntheticCorrelatedData(120, 8)

	base := types.PCAConfig{Components: 5, MeanCenter: true, StandardScale: true}
	svdCfg, nipCfg := base, base
	svdCfg.Method, nipCfg.Method = "svd", "nipals"

	svd, err := NewPCAEngine().Fit(data, svdCfg)
	if err != nil {
		t.Fatalf("SVD fit failed: %v", err)
	}
	nip, err := NewPCAEngine().Fit(data, nipCfg)
	if err != nil {
		t.Fatalf("NIPALS fit failed: %v", err)
	}

	for k := 0; k < 5; k++ {
		for j := range svd.Loadings {
			a, b := svd.Loadings[j][k], nip.Loadings[j][k]
			if math.Abs(a-b) > 1e-6 {
				t.Errorf("PC%d loading %d: SVD %+.9f, NIPALS %+.9f — methods disagree",
					k+1, j, a, b)
				break
			}
		}
	}
}

// TestIssue779_SignConventionIsLargestLoadingPositive pins the rule itself, so
// that a future change to it is a deliberate act rather than a side effect.
func TestIssue779_SignConventionIsLargestLoadingPositive(t *testing.T) {
	data := syntheticCorrelatedData(80, 6)

	for _, method := range []string{"svd", "nipals"} {
		res, err := NewPCAEngine().Fit(data, types.PCAConfig{
			Method: method, Components: 4, MeanCenter: true, StandardScale: true,
		})
		if err != nil {
			t.Fatalf("%s fit failed: %v", method, err)
		}
		for k := 0; k < 4; k++ {
			maxAbs, maxVal := 0.0, 0.0
			for j := range res.Loadings {
				if a := math.Abs(res.Loadings[j][k]); a > maxAbs {
					maxAbs, maxVal = a, res.Loadings[j][k]
				}
			}
			if maxVal < 0 {
				t.Errorf("%s PC%d: largest-magnitude loading is %+.6f, want positive",
					method, k+1, maxVal)
			}
		}
	}
}

// TestIssue779_NativeMissingHonoursStandardScale is the substantive half of the
// bug: NIPALS with native missing-value handling accepted a scaling request and
// silently ignored it, returning an unscaled analysis. On Body Measures with 10%
// of cells removed that meant PC1 = 85% instead of 59%.
func TestIssue779_NativeMissingHonoursStandardScale(t *testing.T) {
	// Column 0 is given a much larger spread than the rest, so an unscaled fit
	// is dominated by it and a scaled fit is not.
	data := syntheticCorrelatedData(200, 5)
	for i := range data {
		data[i][0] *= 50
	}
	// Punch a reproducible hole pattern; every 7th cell of every 3rd row.
	for i := 3; i < len(data); i += 3 {
		data[i][(i/3)%5] = math.NaN()
	}

	cfg := types.PCAConfig{
		Method: "nipals", Components: 3, MeanCenter: true,
		MissingStrategy: types.MissingNative,
	}
	unscaled, err := NewPCAEngine().Fit(data, cfg)
	if err != nil {
		t.Fatalf("unscaled fit failed: %v", err)
	}
	cfg.StandardScale = true
	scaled, err := NewPCAEngine().Fit(data, cfg)
	if err != nil {
		t.Fatalf("scaled fit failed: %v", err)
	}

	if math.Abs(unscaled.ExplainedVarRatio[0]-scaled.ExplainedVarRatio[0]) < 1e-6 {
		t.Fatalf("StandardScale had no effect on the native-missing path: "+
			"PC1 %.4f%% both ways — the scaling request is being ignored",
			scaled.ExplainedVarRatio[0])
	}
	// The inflated column must dominate without scaling and not with it.
	if unscaled.ExplainedVarRatio[0] <= scaled.ExplainedVarRatio[0] {
		t.Errorf("expected the unscaled fit to concentrate more variance in PC1: "+
			"unscaled %.2f%%, scaled %.2f%%",
			unscaled.ExplainedVarRatio[0], scaled.ExplainedVarRatio[0])
	}
}

// TestIssue779_NativeMissingRejectsRowWisePreprocessing checks that the
// combinations that remain unsupported now fail loudly. Silently returning a
// different analysis than the one requested was the original defect; a warning
// on stdout does not reach a Desktop user.
func TestIssue779_NativeMissingRejectsRowWisePreprocessing(t *testing.T) {
	data := syntheticCorrelatedData(40, 4)
	data[5][2] = math.NaN()

	for _, tc := range []struct {
		name string
		cfg  types.PCAConfig
	}{
		{"SNV", types.PCAConfig{SNV: true}},
		{"vector normalization", types.PCAConfig{VectorNorm: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			cfg.Method, cfg.Components = "nipals", 2
			cfg.MeanCenter, cfg.MissingStrategy = true, types.MissingNative
			if _, err := NewPCAEngine().Fit(data, cfg); err == nil {
				t.Fatalf("%s with native missing handling was accepted; want an error", tc.name)
			}
		})
	}
}

// syntheticCorrelatedData builds a well-conditioned matrix whose columns share
// two latent factors, so that the leading components are stable and the sign
// question is well posed.
func syntheticCorrelatedData(n, m int) types.Matrix {
	data := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		row := make([]float64, m)
		f1 := math.Sin(float64(i) * 0.11)
		f2 := math.Cos(float64(i) * 0.07)
		for j := 0; j < m; j++ {
			w := float64(j+1) / float64(m)
			row[j] = w*f1 + (1-w)*f2 + 0.05*math.Sin(float64(i*j)*0.37)
		}
		data[i] = row
	}
	return data
}
