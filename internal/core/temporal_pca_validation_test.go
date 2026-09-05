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

package core_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// Temporal PCA compared against the reference generator.
//
// Read the next paragraph before trusting this test further than it deserves.
//
// Unlike the Kernel PCA and PCR comparisons, this is NOT validation against an
// independent implementation. scikit-learn has no Singular Spectrum Analysis, so
// generate_temporal_pca_reference.py builds the trajectory matrix and decomposes
// it with scipy itself. Agreement therefore says that two implementations of the
// same algorithm, written from the same understanding, agree -- which catches
// drift and transcription errors but cannot catch a shared misreading of SSA.
// Validating against R's Rssa would be the stronger thing and is not what this
// is.
//
// It is still worth having. The trajectory construction, the lag handling and
// the component ordering are exactly the places where an implementation quietly
// diverges, and nothing else in the suite compares them against anything.
//
// What is compared, and what is not:
//
//   - Singular values, tightly. These are the physical quantity of the
//     decomposition and carry no sign or normalisation convention. They agree
//     to floating point.
//   - The shape of the variance profile, after renormalising both sides.
//   - The trajectory dimensions, since a lag or window off by one would change
//     them and is the easiest thing to get wrong.
//   - NOT explained_variance. GoPCA's temporal path puts a *fraction* in that
//     field (temporal_pca.go divides each eigenvalue by the total), while its
//     linear and kernel paths put the raw eigenvalue there and the reference
//     reports the eigenvalue. That inconsistency between GoPCA's own three
//     methods is a wire-format problem tracked for V2 in #848, not something to
//     paper over with a scaling factor here.
//   - NOT scores or loadings directly. A component is defined only up to sign,
//     and SSA pairs oscillatory components into two-dimensional subspaces where
//     the individual directions are arbitrary -- the same degeneracy the kernel
//     comparison handles, and far more common here because pairing is what SSA
//     does with any periodic signal.

// temporalReference is one SSA fit from the reference generator.
type temporalReference struct {
	WindowLength           int         `json:"window_length"`
	Preprocessing          string      `json:"preprocessing"`
	NTimepoints            int         `json:"n_timepoints"`
	NVariables             int         `json:"n_variables"`
	NComponents            int         `json:"n_components"`
	TrajectoryMatrixShape  []int       `json:"trajectory_matrix_shape"`
	InputSeries            [][]float64 `json:"input_series"`
	SingularValues         []float64   `json:"singular_values"`
	ExplainedVarianceRatio []float64   `json:"explained_variance_ratio"`
}

// TestValidateTemporalPCAAgainstReference compares every temporal reference.
//
// The references carry their own input series. They have to: the synthetic ones
// are built from seeded numpy expressions, so before that field existed a Go
// test had no way to reconstruct the input and nothing to compare against. That
// was the structural reason these files went unread for as long as they did.
func TestValidateTemporalPCAAgainstReference(t *testing.T) {
	references := []string{
		"synthetic_sine_wave_temporal_pca.json",
		"synthetic_damped_oscillation_temporal_pca.json",
		"synthetic_multi_frequency_temporal_pca.json",
		"synthetic_trend_seasonal_temporal_pca.json",
		"synthetic_multivariate_temporal_pca.json",
		"iris_temporal_pca_multivariate.json",
	}

	refDir := filepath.Join("..", "..", "testdata", "validation", "reference_results")
	paths := make([]string, 0, len(references))
	for _, name := range references {
		paths = append(paths, filepath.Join(refDir, name))
	}
	requireReferences(t, "Temporal PCA", paths...)

	for _, name := range references {
		t.Run(name, func(t *testing.T) {
			ref := loadTemporalReference(t, filepath.Join(refDir, name))

			config := types.PCAConfig{
				Components:   ref.NComponents,
				Method:       "temporal",
				TemporalLags: ref.WindowLength,
				MeanCenter:   ref.Preprocessing == "mean_center" || ref.Preprocessing == "standardize",
			}
			if ref.Preprocessing == "standardize" {
				// MeanCenter stays on. In GoPCA, StandardScale means "divide by
				// the standard deviation" and centring is controlled separately
				// by MeanCenter -- in every method, linear included. Setting only
				// StandardScale leaves the data uncentred, which puts the mean
				// into the first component: on iris that inflated the leading
				// singular value from 78 to 528.
				config.StandardScale = true
			}

			result, err := core.NewPCAEngineForMethod("temporal").Fit(
				types.Matrix(ref.InputSeries), config)
			if err != nil {
				t.Fatalf("Fit: %v", err)
			}

			compareTrajectoryShape(t, ref, result)
			compareTemporalSingularValues(t, ref, result)
			compareTemporalVarianceShape(t, ref, result)
		})
	}
}

// compareTrajectoryShape checks the embedding.
//
// A window or lag off by one changes these dimensions and nothing else in the
// suite would notice. The row count is T-L+1 and the column count is p*L; both
// implementations must agree on that or they are decomposing different matrices.
func compareTrajectoryShape(t *testing.T, ref *temporalReference, result *types.PCAResult) {
	t.Helper()

	if len(ref.TrajectoryMatrixShape) != 2 {
		t.Fatalf("the reference records a trajectory shape of %v, expected two dimensions",
			ref.TrajectoryMatrixShape)
	}
	wantRows := ref.TrajectoryMatrixShape[0]
	if got := len(result.Scores); got != wantRows {
		t.Errorf("GoPCA produced %d score rows; the reference embedded %d windows "+
			"(T=%d, L=%d gives T-L+1=%d)",
			got, wantRows, ref.NTimepoints, ref.WindowLength,
			ref.NTimepoints-ref.WindowLength+1)
	}
}

// compareTemporalSingularValues checks the decomposition itself.
//
// Singular values carry no sign and no normalisation convention, which makes
// them the one quantity here that can be compared without first agreeing about
// anything else.
func compareTemporalSingularValues(t *testing.T, ref *temporalReference, result *types.PCAResult) {
	t.Helper()

	if len(result.SingularValues) == 0 {
		t.Fatal("GoPCA returned no singular values, so there is nothing to compare")
	}

	n := ref.NComponents
	if len(result.SingularValues) < n {
		n = len(result.SingularValues)
	}
	if len(ref.SingularValues) < n {
		n = len(ref.SingularValues)
	}
	if n == 0 {
		t.Fatal("no singular values in common between GoPCA and the reference")
	}

	// Values at or near zero carry no information and their relative difference
	// is meaningless; the largest value sets the scale for what counts as zero.
	scale := ref.SingularValues[0]
	for i := 0; i < n; i++ {
		if math.Abs(ref.SingularValues[i]) < 1e-9*scale {
			continue
		}
		if d := relativeDifference(result.SingularValues[i], ref.SingularValues[i]); d > 1e-9 {
			t.Errorf("singular value %d: GoPCA %.12g, reference %.12g (relative %.3g)",
				i, result.SingularValues[i], ref.SingularValues[i], d)
		}
	}
}

// compareTemporalVarianceShape compares how variance is distributed, with the
// conventions removed.
//
// GoPCA reports percentages and the reference reports fractions, so both are
// renormalised over the components in common. What survives is the relative
// weight of each component, which is the part that carries meaning.
func compareTemporalVarianceShape(t *testing.T, ref *temporalReference, result *types.PCAResult) {
	t.Helper()

	n := ref.NComponents
	if len(result.ExplainedVarRatio) < n {
		n = len(result.ExplainedVarRatio)
	}
	if len(ref.ExplainedVarianceRatio) < n {
		n = len(ref.ExplainedVarianceRatio)
	}

	share := func(values []float64) []float64 {
		total := 0.0
		for i := 0; i < n; i++ {
			total += values[i]
		}
		if total == 0 {
			return nil
		}
		out := make([]float64, n)
		for i := 0; i < n; i++ {
			out[i] = values[i] / total
		}
		return out
	}

	want, got := share(ref.ExplainedVarianceRatio), share(result.ExplainedVarRatio)
	if want == nil || got == nil {
		t.Fatal("a variance profile summed to zero, which cannot be right")
	}
	for i := range want {
		if d := math.Abs(got[i] - want[i]); d > 1e-9 {
			t.Errorf("component %d holds %.9f of the retained variance in GoPCA "+
				"and %.9f in the reference", i, got[i], want[i])
		}
	}
}

func loadTemporalReference(t *testing.T, path string) *temporalReference {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var ref temporalReference
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	if len(ref.InputSeries) == 0 {
		t.Fatalf("%s carries no input_series, so GoPCA cannot be given the same "+
			"input and there is nothing to compare; regenerate it", path)
	}
	if ref.WindowLength == 0 || ref.NComponents == 0 {
		t.Fatalf("%s records no window length or component count", path)
	}
	return &ref
}
