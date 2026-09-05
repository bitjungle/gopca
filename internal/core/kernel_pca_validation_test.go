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
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// Kernel PCA validated against scikit-learn.
//
// CI has generated these reference files on every run for a long time and no Go
// test ever opened one (#845). Kernel PCA's only scikit-learn comparison was
// TestIssue736_ScoresMatchSklearn, which hardcodes expected values for a single
// six-sample fixture at RBF gamma 0.5 -- a real check, and a narrow one. The
// references cover iris, wine and the Swiss Roll across linear, sigmoid,
// polynomial degree 2 and 3, and RBF at gamma from 0.01 to 10.
//
// The gamma sweep is the valuable part. Gamma controls how local the kernel is,
// and 0.01 against 10 are numerically very different regimes; a defect that
// appears only at one end is exactly what a single fixture at 0.5 cannot see.
//
// Two things had to be put right before any comparison could pass, and they are
// the likeliest reason nobody wired these up before:
//
//  1. The generator selected numeric columns without excluding #target ones, so
//     on iris it fed species#target in as a fifth predictor -- decomposing the
//     class label alongside the measurements. GoPCA excludes those columns by
//     design, so the reference described a different problem.
//
//  2. The generator standardized with StandardScaler, which divides by the
//     population standard deviation, while GoPCA divides by the sample one. For
//     linear PCA that difference cancels in the quantities worth comparing. For
//     a kernel it does not: gamma multiplies squared distances, so rescaling the
//     inputs changes exp(-gamma*d^2) non-linearly. On iris it moved the leading
//     eigenvalue by 0.4% -- far above floating-point noise, and far below what a
//     loosened tolerance would flag as a bug. With the divisor matched, the
//     eigenvalues agree exactly.

// kernelReference is one scikit-learn Kernel PCA fit.
type kernelReference struct {
	Kernel                 string             `json:"kernel"`
	KernelParams           map[string]float64 `json:"kernel_params"`
	Preprocessing          string             `json:"preprocessing"`
	NSamples               int                `json:"n_samples"`
	NFeatures              int                `json:"n_features"`
	NComponents            int                `json:"n_components"`
	Scores                 [][]float64        `json:"scores"`
	Eigenvalues            []float64          `json:"eigenvalues"`
	ExplainedVarianceRatio []float64          `json:"explained_variance_ratio"`
}

// TestValidateKernelPCAAgainstSklearn compares every reference that has a real
// dataset behind it.
//
// The name begins with TestValidate so the CI validation filter catches it.
func TestValidateKernelPCAAgainstSklearn(t *testing.T) {
	cases := []struct {
		reference string
		dataset   string
	}{
		// GoPCA implements rbf, linear and poly. There is deliberately no
		// sigmoid case: the generator no longer produces one, because a
		// reference for a kernel the engine does not have is a file nothing
		// can consume, which is the shape of problem #845 is about.
		{"iris_kpca_linear.json", "iris/iris.csv"},
		{"iris_kpca_poly_deg2.json", "iris/iris.csv"},
		{"iris_kpca_poly_deg3.json", "iris/iris.csv"},
		{"iris_kpca_rbf_gamma0.1.json", "iris/iris.csv"},
		{"iris_kpca_rbf_gamma1.0.json", "iris/iris.csv"},
		{"iris_kpca_rbf_gamma10.0.json", "iris/iris.csv"},
		{"wine_kpca_linear.json", "wine/wine.csv"},
		{"wine_kpca_poly_deg2.json", "wine/wine.csv"},
		{"wine_kpca_poly_deg3.json", "wine/wine.csv"},
		{"wine_kpca_rbf_gamma0.1.json", "wine/wine.csv"},
		{"wine_kpca_rbf_gamma1.0.json", "wine/wine.csv"},
		{"wine_kpca_rbf_gamma10.0.json", "wine/wine.csv"},
		{"swiss_roll_kpca_rbf_gamma0.01.json", "swiss_roll/swiss_roll.csv"},
		{"swiss_roll_kpca_rbf_gamma0.1.json", "swiss_roll/swiss_roll.csv"},
		{"swiss_roll_kpca_rbf_gamma1.0.json", "swiss_roll/swiss_roll.csv"},
		{"swiss_roll_kpca_rbf_gamma10.0.json", "swiss_roll/swiss_roll.csv"},
	}

	refDir := filepath.Join("..", "..", "testdata", "validation", "reference_results")
	if _, err := os.Stat(filepath.Join(refDir, cases[0].reference)); os.IsNotExist(err) {
		t.Skip("Kernel PCA reference files not found. Generate them with: " +
			"cd testdata/validation && python generate_kernel_pca_reference.py")
	}

	for _, tc := range cases {
		t.Run(tc.reference, func(t *testing.T) {
			ref := loadKernelReference(t, filepath.Join(refDir, tc.reference))
			data := loadKernelDataset(t, tc.dataset)

			if len(data) != ref.NSamples || len(data[0]) != ref.NFeatures {
				t.Fatalf("GoPCA reads %dx%d from %s but the reference was built on %dx%d; "+
					"the two are not describing the same problem",
					len(data), len(data[0]), tc.dataset, ref.NSamples, ref.NFeatures)
			}

			result, err := core.NewKernelPCAEngine().Fit(data, kernelConfigFor(ref))
			if err != nil {
				t.Fatalf("Fit: %v", err)
			}

			compareKernelEigenvalues(t, ref, result)
			compareKernelScores(t, ref, result)
			compareKernelVarianceShape(t, ref, result)
		})
	}
}

func kernelConfigFor(ref *kernelReference) types.PCAConfig {
	config := types.PCAConfig{
		Components:   ref.NComponents,
		Method:       "kernel",
		KernelType:   ref.Kernel,
		KernelGamma:  ref.KernelParams["gamma"],
		KernelDegree: int(ref.KernelParams["degree"]),
		KernelCoef0:  ref.KernelParams["coef0"],
	}
	// Kernel PCA centres in kernel space, so the predictor-side preprocessing is
	// scaling only -- which is what the reference applies before computing its
	// kernel matrix.
	if ref.Preprocessing == "standardize" {
		config.StandardScale = true
		config.ScaleOnly = true
	}
	return config
}

// compareKernelEigenvalues checks the eigenvalues of the centred kernel matrix.
//
// These are the physical quantity: they do not depend on the sign convention and
// they are what scikit-learn exposes as eigenvalues_. GoPCA reports them as
// ExplainedVar. With the standardization divisor matched they agree to floating
// point, so the tolerance is tight enough to catch a real defect.
func compareKernelEigenvalues(t *testing.T, ref *kernelReference, result *types.PCAResult) {
	t.Helper()

	if len(result.ExplainedVar) < len(ref.Eigenvalues) {
		t.Fatalf("GoPCA returned %d eigenvalues, the reference has %d",
			len(result.ExplainedVar), len(ref.Eigenvalues))
	}
	for i, want := range ref.Eigenvalues {
		got := result.ExplainedVar[i]
		if d := relativeDifference(got, want); d > 1e-9 {
			t.Errorf("eigenvalue %d: GoPCA %.12g, scikit-learn %.12g (relative %.3g)",
				i, got, want, d)
		}
	}
}

// compareKernelScores checks the projections, up to the sign of each component.
//
// A component is defined only up to sign, and the two implementations resolve
// that independently, so each column is aligned before comparison. The alignment
// is per component and by the largest-magnitude entry, which is stable; aligning
// on an entry near zero would pick a sign from noise.
func compareKernelScores(t *testing.T, ref *kernelReference, result *types.PCAResult) {
	t.Helper()

	if len(result.Scores) != len(ref.Scores) {
		t.Fatalf("GoPCA produced %d score rows, the reference has %d",
			len(result.Scores), len(ref.Scores))
	}

	for c := 0; c < ref.NComponents; c++ {
		// A component whose eigenvalue is shared with a neighbour has no
		// individually determined eigenvector: any rotation within the
		// degenerate subspace is an equally valid answer, and the two
		// implementations have no reason to choose the same one. Comparing such
		// a column measures which arbitrary basis each happened to pick.
		//
		// This is not hypothetical. On wine at RBF gamma 10, thirteen features
		// put every pair far enough apart that exp(-gamma*d^2) collapses: the
		// kernel matrix is the identity to six decimals and all ten eigenvalues
		// are 1.0. That the two implementations still agree to 2.6e-8 there is
		// remarkable rather than required.
		//
		// The eigenvalues are still compared for these components, and so is the
		// share of variance. Only the individual directions are skipped.
		if degenerateWith(ref.Eigenvalues, c) {
			continue
		}

		sign := alignmentSign(ref.Scores, result.Scores, c)
		worst, at := 0.0, -1
		for row := range ref.Scores {
			got := sign * result.Scores[row][c]
			if d := math.Abs(got - ref.Scores[row][c]); d > worst {
				worst, at = d, row
			}
		}
		if worst > 1e-8 {
			t.Errorf("component %d: worst score difference %.3g at row %d "+
				"(GoPCA %.12g against scikit-learn %.12g)",
				c, worst, at, sign*result.Scores[at][c], ref.Scores[at][c])
		}
	}
}

// alignmentSign returns +1 or -1 to bring a GoPCA component into the reference's
// sign convention, decided by the entry with the largest reference magnitude.
func alignmentSign(reference, got [][]float64, component int) float64 {
	best, at := 0.0, 0
	for row := range reference {
		if m := math.Abs(reference[row][component]); m > best {
			best, at = m, row
		}
	}
	if reference[at][component]*got[at][component] < 0 {
		return -1
	}
	return 1
}

// compareKernelVarianceShape checks how variance is distributed across the
// retained components.
//
// The absolute ratios are not comparable: scikit-learn normalises over the
// components it retained, so its values sum to 1, while GoPCA normalises over
// the whole spectrum and reports a percentage. Neither is wrong. Renormalising
// both over the retained set removes the convention and leaves the quantity that
// actually carries meaning -- the relative weight of each component.
func compareKernelVarianceShape(t *testing.T, ref *kernelReference, result *types.PCAResult) {
	t.Helper()

	share := func(values []float64, n int) []float64 {
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

	want := share(ref.ExplainedVarianceRatio, ref.NComponents)
	got := share(result.ExplainedVarRatio, ref.NComponents)
	if want == nil || got == nil {
		t.Fatal("a variance profile summed to zero, which cannot be right")
	}

	for i := range want {
		if d := math.Abs(got[i] - want[i]); d > 1e-9 {
			t.Errorf("component %d holds %.9f of the retained variance in GoPCA "+
				"and %.9f in scikit-learn", i, got[i], want[i])
		}
	}
}

func loadKernelReference(t *testing.T, path string) *kernelReference {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var ref kernelReference
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	if ref.NComponents == 0 || len(ref.Eigenvalues) == 0 {
		t.Fatalf("%s carries no components, so this case would check nothing", path)
	}
	return &ref
}

// loadKernelDataset reads through the same parser the applications use, so the
// test exercises the real path from file to engine.
func loadKernelDataset(t *testing.T, relative string) types.Matrix {
	t.Helper()

	opts := pkgcsv.DefaultOptions()
	opts.ParseMode = pkgcsv.ParseMixedWithTargets
	parsed, err := pkgcsv.NewReader(opts).ReadFile(
		filepath.Join("..", "..", "testdata", relative))
	if err != nil {
		t.Fatalf("reading %s: %v", relative, err)
	}
	return types.Matrix(parsed.Matrix)
}

// degenerateWith reports whether a component's eigenvalue is indistinguishable
// from that of an adjacent component, leaving its eigenvector undetermined.
//
// The threshold is relative, because eigenvalue scale varies by orders of
// magnitude across these fixtures -- from about 1 on a collapsed RBF kernel to
// 2.5e7 on a polynomial one.
func degenerateWith(eigenvalues []float64, c int) bool {
	const separation = 1e-6

	near := func(a, b float64) bool {
		scale := math.Max(math.Abs(a), math.Abs(b))
		if scale == 0 {
			return true
		}
		return math.Abs(a-b)/scale < separation
	}

	if c > 0 && near(eigenvalues[c], eigenvalues[c-1]) {
		return true
	}
	if c+1 < len(eigenvalues) && near(eigenvalues[c], eigenvalues[c+1]) {
		return true
	}
	return false
}
