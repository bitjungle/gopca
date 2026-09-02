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

// TestColumnAffineReproducesTransform is the check that keeps ColumnAffine and
// Transform from drifting apart.
//
// ColumnAffine restates, as two vectors, the branch structure Transform applies
// in code. Nothing but a test forces the two to agree, and a disagreement would
// not announce itself: PCR would simply report original-scale coefficients that
// are wrong by a per-column factor, and every one of them would look plausible.
func TestColumnAffineReproducesTransform(t *testing.T) {
	configs := []struct {
		name                                                        string
		meanCenter, standardScale, robustScale, scaleOnly, snv, vec bool
	}{
		{"mean center only", true, false, false, false, false, false},
		{"standardize", true, true, false, false, false, false},
		{"scale without centering", false, true, false, false, false, false},
		{"robust", false, false, true, false, false, false},
		{"scale only", false, false, false, true, false, false},
		{"robust with mean center flag set", true, false, true, false, false, false},
		{"snv then standardize", true, true, false, false, true, false},
		{"vector norm then mean center", true, false, false, false, false, true},
	}

	r := rand.New(rand.NewPCG(19, 87))
	const n, m = 40, 6
	data := make(types.Matrix, n)
	for i := range data {
		data[i] = make([]float64, m)
		for j := range data[i] {
			// Deliberately different scales per column, so a wrong divisor shows up.
			data[i][j] = (r.NormFloat64() + float64(j)) * math.Pow(10, float64(j-2))
		}
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			p := NewPreprocessorWithScaleOnly(cfg.meanCenter, cfg.standardScale,
				cfg.robustScale, cfg.scaleOnly, cfg.snv, cfg.vec)

			transformed, err := p.FitTransform(data)
			if err != nil {
				t.Fatalf("FitTransform: %v", err)
			}

			center, divisor, err := p.ColumnAffine()
			if err != nil {
				t.Fatalf("ColumnAffine: %v", err)
			}
			if len(center) != m || len(divisor) != m {
				t.Fatalf("ColumnAffine returned %d/%d values, want %d", len(center), len(divisor), m)
			}
			for j, d := range divisor {
				if d == 0 {
					t.Fatalf("divisor for column %d is zero", j)
				}
			}

			if p.IsRowWiseEnabled() != (cfg.snv || cfg.vec) {
				t.Errorf("IsRowWiseEnabled() = %v, want %v", p.IsRowWiseEnabled(), cfg.snv || cfg.vec)
			}

			// The affine map describes the column-wise stage only, so for a
			// row-wise configuration it must be applied to the row-normalized
			// data. Recover that stage by running the row transform alone.
			source := data
			if p.IsRowWiseEnabled() {
				rowOnly := NewPreprocessorWithScaleOnly(false, false, false, false, cfg.snv, cfg.vec)
				source, err = rowOnly.FitTransform(data)
				if err != nil {
					t.Fatalf("row-wise FitTransform: %v", err)
				}
			}

			for i := 0; i < n; i++ {
				for j := 0; j < m; j++ {
					want := transformed[i][j]
					got := (source[i][j] - center[j]) / divisor[j]
					if math.Abs(got-want) > 1e-9*(1+math.Abs(want)) {
						t.Fatalf("row %d column %d: affine map gives %.12g, Transform gives %.12g",
							i, j, got, want)
					}
				}
			}
		})
	}
}

// TestColumnAffineRejectsUnfitted checks that the accessor refuses to describe a
// map that has not been estimated, rather than returning zeros that would read as
// "no centering, divide by nothing".
func TestColumnAffineRejectsUnfitted(t *testing.T) {
	p := NewPreprocessor(true, true, false)
	if _, _, err := p.ColumnAffine(); err == nil {
		t.Error("expected an error from an unfitted preprocessor")
	}
}

// TestColumnAffineHandlesConstantColumn covers the clamp: a near-constant column
// has its divisor forced to 1 so that scaling cannot divide by zero. The affine
// map must report the clamped value, not the measured standard deviation, or
// coefficients for that column would be scaled by a vanishing number.
func TestColumnAffineHandlesConstantColumn(t *testing.T) {
	data := types.Matrix{
		{1, 5}, {2, 5}, {3, 5}, {4, 5},
	}
	p := NewPreprocessorWithScaleOnly(true, true, false, false, false, false)
	if _, err := p.FitTransform(data); err != nil {
		t.Fatalf("FitTransform: %v", err)
	}

	_, divisor, err := p.ColumnAffine()
	if err != nil {
		t.Fatalf("ColumnAffine: %v", err)
	}
	if divisor[1] != 1.0 {
		t.Errorf("divisor for the constant column = %v, want the clamped value 1.0", divisor[1])
	}
	if sd := p.GetStdDevs()[1]; sd == divisor[1] {
		t.Log("note: measured std dev happens to equal the clamp here; " +
			"the distinction this test guards is still real")
	}
}
