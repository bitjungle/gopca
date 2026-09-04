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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

// finite reports whether v is a usable number. The engine has an equivalent
// helper, but it is unexported and this file is an external test package: it must
// import pkg/csv to exercise the real loading path, and pkg/csv already imports
// the engine.
func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// pcrReferenceFit is one fitted model from the scikit-learn reference.
type pcrReferenceFit struct {
	NComponents              int       `json:"n_components"`
	Preprocessing            string    `json:"preprocessing"`
	Coefficients             []float64 `json:"coefficients"`
	Intercept                float64   `json:"intercept"`
	Fitted                   []float64 `json:"fitted"`
	RMSEC                    float64   `json:"rmsec"`
	R2C                      float64   `json:"r2c"`
	ResponseMean             float64   `json:"response_mean"`
	CoefficientRecoveryError float64   `json:"coefficient_recovery_error"`
}

// pcrCVPoint is one candidate component count on a cross-validated error curve.
type pcrCVPoint struct {
	NComponents int     `json:"n_components"`
	RMSECV      float64 `json:"rmsecv"`
	Bias        float64 `json:"bias"`
	SEP         float64 `json:"sep"`
	MAE         float64 `json:"mae"`
	Q2          float64 `json:"q2"`

	// The per-fold statistics. RMSECVMean averages the per-fold RMSE where
	// RMSECV pools every residual first, and RMSECVSE is the spread of those
	// per-fold values -- the quantity the one-standard-error selection rule
	// reads. Both were reported by GoPCA and compared against nothing.
	RMSECVMean float64 `json:"rmsecv_mean"`
	RMSECVSE   float64 `json:"rmsecv_se"`
	MAESE      float64 `json:"mae_se"`
}

// pcrCVDesign is a reference curve for one fold layout.
//
// Only contiguous K-fold had a reference. GroupKFold is one code path in GoPCA
// serving K-fold, leave-one-out, grouped K-fold and leave-one-group-out, so a
// fold assignment disagreeing with scikit-learn's would have shown up in none
// of the existing assertions.
type pcrCVDesign struct {
	Design string       `json:"design"`
	NFolds int          `json:"n_folds"`
	Groups []int        `json:"groups,omitempty"`
	Curve  []pcrCVPoint `json:"curve"`
}

// pcrSemiSupervised is a reference for a fit where only some rows carry a
// response: the decomposition sees every row, the regression only the labelled
// ones. This is the path where a leak makes the result better rather than
// worse, so no ordinary assertion fails when one is present.
type pcrSemiSupervised struct {
	NComponents        int       `json:"n_components"`
	LabelledRows       []int     `json:"labelled_rows"`
	NLabelled          int       `json:"n_labelled"`
	NDecompositionRows int       `json:"n_decomposition_rows"`
	Coefficients       []float64 `json:"coefficients"`
	Intercept          float64   `json:"intercept"`
	Fitted             []float64 `json:"fitted"`
	RMSEC              float64   `json:"rmsec"`
	R2C                float64   `json:"r2c"`
}

// pcrCVReference is a cross-validated sweep from scikit-learn.
type pcrCVReference struct {
	NFolds int          `json:"n_folds"`
	Curve  []pcrCVPoint `json:"curve"`
}

// pcrReference is one reference file: a dataset, a preprocessing choice, and the
// fits at several component counts.
type pcrReference struct {
	Dataset       string            `json:"dataset"`
	Source        string            `json:"source"`
	Response      string            `json:"response"`
	Preprocessing string            `json:"preprocessing"`
	NSamples      int               `json:"n_samples"`
	NFeatures     int               `json:"n_features"`
	XChecksum     float64           `json:"x_checksum"`
	YChecksum     float64           `json:"y_checksum"`
	Fits          []pcrReferenceFit `json:"fits"`
	CrossValid    *pcrCVReference   `json:"cross_validation,omitempty"`

	CVDesigns      []pcrCVDesign      `json:"cv_designs,omitempty"`
	SemiSupervised *pcrSemiSupervised `json:"semi_supervised,omitempty"`
}

// TestValidatePCRAgainstSklearn checks the estimator against scikit-learn on real
// datasets shipped with the repository.
//
// What is compared, and why only these quantities:
//
//   - Original-scale coefficients, the intercept, fitted values, RMSEC and R2C.
//     These are invariant to the two conventions that legitimately differ between
//     the implementations.
//   - Not scores, singular values or eigenvalues. GoPCA standardizes with the
//     sample standard deviation (divisor n-1) and scikit-learn with the
//     population one (divisor n), so the standardized matrices differ by the
//     scalar sqrt(n/(n-1)). That factor cancels between theta and the inverse
//     scaling, leaving coefficients and predictions identical while scaling
//     score-space quantities. Comparing those directly would report a difference
//     that is a convention rather than an error.
//   - Not the score-space coefficients gamma, whose signs follow the arbitrary
//     signs of the components.
//
// The name begins with TestValidate so that the CI validation step, which filters
// on that prefix, actually runs it. A test CI never executes guards nothing.
func TestValidatePCRAgainstSklearn(t *testing.T) {
	references := []string{
		"corn_moisture_pcr_standardize.json",
		"corn_moisture_pcr_mean_center.json",
		"corn_protein_pcr_standardize.json",
		"corn_protein_pcr_mean_center.json",
		"wine_flavanoids_pcr_standardize.json",
		"wine_flavanoids_pcr_mean_center.json",
	}

	refDir := filepath.Join("..", "..", "testdata", "validation", "reference_results")
	if _, err := os.Stat(filepath.Join(refDir, references[0])); os.IsNotExist(err) {
		t.Skip("PCR reference files not found. Generate them with: " +
			"cd testdata/validation && python generate_reference_pcr.py")
	}

	for _, name := range references {
		t.Run(name, func(t *testing.T) {
			ref := loadPCRReference(t, filepath.Join(refDir, name))

			data, y := loadPCRDataset(t, ref)
			assertDatasetMatches(t, ref, data, y)

			for _, fit := range ref.Fits {
				t.Run(fmt.Sprintf("k=%d", fit.NComponents), func(t *testing.T) {
					comparePCRFit(t, data, y, ref, fit)
				})
			}

			if ref.CrossValid != nil {
				t.Run("cross-validation", func(t *testing.T) {
					comparePCRCrossValidation(t, data, y, ref)
				})
			}

			for _, design := range ref.CVDesigns {
				t.Run(design.Design, func(t *testing.T) {
					comparePCRCVDesign(t, data, y, ref, design)
				})
			}

			if ref.SemiSupervised != nil {
				t.Run("semi-supervised", func(t *testing.T) {
					comparePCRSemiSupervised(t, data, y, ref)
				})
			}
		})
	}
}

func loadPCRReference(t *testing.T, path string) *pcrReference {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var ref pcrReference
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	if len(ref.Fits) == 0 {
		t.Fatalf("%s contains no fits", path)
	}
	return &ref
}

// loadPCRDataset reads the dataset through the same parser the CLI uses, so that
// this test exercises the real path from file to estimator rather than a
// convenient shortcut around it.
func loadPCRDataset(t *testing.T, ref *pcrReference) (types.Matrix, []float64) {
	t.Helper()

	opts := pkgcsv.DefaultOptions()
	opts.ParseMode = pkgcsv.ParseMixedWithTargets
	reader := pkgcsv.NewReader(opts)

	parsed, err := reader.ReadFile(filepath.Join("..", "..", "testdata", ref.Source))
	if err != nil {
		t.Fatalf("reading %s: %v", ref.Source, err)
	}

	// The response is either a numeric #target column or, where the dataset has
	// none, an ordinary numeric column held out as the response.
	response, isTarget := parsed.NumericTargetColumns[ref.Response]
	responseColumn := -1
	if !isTarget {
		for j, header := range parsed.Headers {
			if header == ref.Response {
				responseColumn = j
				break
			}
		}
		if responseColumn < 0 {
			t.Fatalf("response %q is neither a numeric target nor a column of %s",
				ref.Response, ref.Source)
		}
		response = make([]float64, parsed.Rows)
		for i := 0; i < parsed.Rows; i++ {
			response[i] = parsed.Matrix[i][responseColumn]
		}
	}

	// Keep rows whose response and predictors are all finite, matching the
	// generator, and drop the response column from the predictors when it came
	// from the ordinary numeric block.
	data := make(types.Matrix, 0, parsed.Rows)
	y := make([]float64, 0, parsed.Rows)
	for i := 0; i < parsed.Rows; i++ {
		if !finite(response[i]) {
			continue
		}
		row := make([]float64, 0, parsed.Columns)
		complete := true
		for j := 0; j < parsed.Columns; j++ {
			if j == responseColumn {
				continue
			}
			if !finite(parsed.Matrix[i][j]) {
				complete = false
				break
			}
			row = append(row, parsed.Matrix[i][j])
		}
		if !complete {
			continue
		}
		data = append(data, row)
		y = append(y, response[i])
	}

	if len(data) == 0 {
		t.Fatalf("no usable rows loaded from %s", ref.Source)
	}
	return data, y
}

// assertDatasetMatches confirms Go loaded the same numbers Python did, before any
// model output is compared. Without this, a difference in CSV parsing would show
// up as an unexplained coefficient mismatch rather than as the loading problem it
// actually is.
func assertDatasetMatches(t *testing.T, ref *pcrReference, data types.Matrix, y []float64) {
	t.Helper()

	if len(data) != ref.NSamples {
		t.Fatalf("loaded %d rows, reference used %d", len(data), ref.NSamples)
	}
	if len(data[0]) != ref.NFeatures {
		t.Fatalf("loaded %d predictors, reference used %d", len(data[0]), ref.NFeatures)
	}

	var xSum, ySum float64
	for i := range data {
		for _, v := range data[i] {
			xSum += v
		}
		ySum += y[i]
	}

	if math.Abs(xSum-ref.XChecksum) > 1e-6*(1+math.Abs(ref.XChecksum)) {
		t.Fatalf("predictor checksum %.9g does not match the reference %.9g: "+
			"Go and Python did not load the same data", xSum, ref.XChecksum)
	}
	if math.Abs(ySum-ref.YChecksum) > 1e-9*(1+math.Abs(ref.YChecksum)) {
		t.Fatalf("response checksum %.9g does not match the reference %.9g",
			ySum, ref.YChecksum)
	}
}

func comparePCRFit(t *testing.T, data types.Matrix, y []float64, ref *pcrReference, fit pcrReferenceFit) {
	t.Helper()

	// The generator asserts its own coefficient recovery against the pipeline it
	// came from; re-check it here so that a broken reference cannot quietly become
	// the standard the Go code is held to.
	if fit.CoefficientRecoveryError > 1e-8 {
		t.Fatalf("the reference itself is inconsistent: recovering coefficients from the "+
			"pipeline differed by %g", fit.CoefficientRecoveryError)
	}

	config := types.PCRConfig{
		PCA: types.PCAConfig{
			Components:    fit.NComponents,
			MeanCenter:    true,
			StandardScale: ref.Preprocessing == "standardize",
			Method:        "svd",
		},
		Response: ref.Response,
		Selection: types.SelectionConfig{
			Mode:   "fixed",
			Fixed:  fit.NComponents,
			Metric: "rmse",
		},
	}

	result, err := core.NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	if !result.OriginalScaleValid {
		t.Fatal("expected original-scale coefficients to be available")
	}

	// Spectral data is strongly collinear, so a looser tolerance is appropriate
	// there than for the well-conditioned case. Both are far tighter than any
	// difference that would indicate a wrong estimator.
	tolerance := 1e-7
	if ref.NFeatures > 100 {
		tolerance = 1e-5
	}

	if len(result.Coefficients) != len(fit.Coefficients) {
		t.Fatalf("got %d coefficients, reference has %d",
			len(result.Coefficients), len(fit.Coefficients))
	}

	var maxCoefficientError float64
	for j := range fit.Coefficients {
		scaled := math.Abs(result.Coefficients[j]-fit.Coefficients[j]) /
			(1 + math.Abs(fit.Coefficients[j]))
		if scaled > maxCoefficientError {
			maxCoefficientError = scaled
		}
	}
	if maxCoefficientError > tolerance {
		t.Errorf("largest relative coefficient difference %.3g exceeds %.0e",
			maxCoefficientError, tolerance)
	}

	if diff := relativeDifference(result.InterceptOriginal, fit.Intercept); diff > tolerance {
		t.Errorf("intercept: Go %.12g, sklearn %.12g (relative %.3g)",
			result.InterceptOriginal, fit.Intercept, diff)
	}
	if diff := relativeDifference(result.RMSEC, fit.RMSEC); diff > tolerance {
		t.Errorf("RMSEC: Go %.12g, sklearn %.12g (relative %.3g)",
			result.RMSEC, fit.RMSEC, diff)
	}
	if diff := relativeDifference(result.R2C, fit.R2C); diff > tolerance {
		t.Errorf("R2C: Go %.12g, sklearn %.12g (relative %.3g)", result.R2C, fit.R2C, diff)
	}

	if len(result.Fitted) != len(fit.Fitted) {
		t.Fatalf("got %d fitted values, reference has %d", len(result.Fitted), len(fit.Fitted))
	}
	var maxFittedError float64
	for i := range fit.Fitted {
		if d := relativeDifference(result.Fitted[i], fit.Fitted[i]); d > maxFittedError {
			maxFittedError = d
		}
	}
	if maxFittedError > tolerance {
		t.Errorf("largest relative fitted-value difference %.3g exceeds %.0e",
			maxFittedError, tolerance)
	}
}

func relativeDifference(got, want float64) float64 {
	return math.Abs(got-want) / (1 + math.Abs(want))
}

// comparePCRCrossValidation checks the whole cross-validation machinery against
// scikit-learn: fold construction, refitting every learned step inside each fold,
// and pooling the out-of-fold residuals.
//
// Contiguous folds are used because they are the one design both implementations
// construct identically. A shuffled design would depend on each language's random
// generator, so the comparison would be measuring the shuffle rather than the
// validation. The reference generator asserts that the row count divides evenly by
// the fold count, since scikit-learn and this implementation distribute a
// remainder differently.
//
// Agreement here is strong evidence that no learned step escaped a fold: leakage
// would push the Go error below the reference, and the comparison is two-sided.
func comparePCRCrossValidation(t *testing.T, data types.Matrix, y []float64, ref *pcrReference) {
	t.Helper()

	reference := ref.CrossValid
	maxComponents := reference.Curve[len(reference.Curve)-1].NComponents

	config := types.PCRConfig{
		PCA: types.PCAConfig{
			Components:    maxComponents,
			MeanCenter:    true,
			StandardScale: ref.Preprocessing == "standardize",
			Method:        "svd",
		},
		Response: ref.Response,
		Selection: types.SelectionConfig{
			Mode:   "cv",
			Metric: "rmse",
			Rule:   types.SelectMin,
			CV: types.CVConfig{
				Scheme: types.CVContiguous,
				Folds:  reference.NFolds,
			},
		},
	}

	result, err := core.NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	if result.CV == nil {
		t.Fatal("expected a cross-validation report")
	}

	// One comparison routine serves every fold layout, so a check added for one
	// design cannot quietly be missing from another. The per-fold statistics were
	// first compared only in the alternative designs, which meant a defect in
	// RMSECVMean went undetected here on the very layout most runs use.
	comparePCRCurve(t, "contiguous", result.CV, reference.Curve, ref.NFeatures)

	// The identity relating the three error measures must hold at every candidate.
	for i, k := range result.CV.Candidates {
		n := float64(result.CV.NSamples)
		want := result.CV.RMSECV[i] * result.CV.RMSECV[i]
		have := result.CV.Bias[i]*result.CV.Bias[i] +
			(n-1)/n*result.CV.SEP[i]*result.CV.SEP[i]
		if math.Abs(want-have) > 1e-12*(1+want) {
			t.Errorf("k=%d: RMSECV^2 = %.15g but bias^2 + (n-1)/n SEP^2 = %.15g", k, want, have)
		}
	}
}

// comparePCRCVDesign checks a fold layout other than plain contiguous K-fold.
//
// Both designs the reference carries have a *unique* partition, which is why
// they are the ones chosen. Shuffled K-fold and balanced GroupKFold depend on
// assignment heuristics that differ between implementations, so a disagreement
// there would say nothing about the machinery being tested. With one row per
// fold, or one group per fold, there is only one possible answer and any
// difference is a real one.
//
// GoPCA expresses both through the same GroupKFold with Folds set to zero,
// meaning "as many folds as there are groups" -- with the default grouping of
// one row per group that is leave-one-out, and with an explicit grouping it is
// leave-one-group-out. That the two designs share a code path is exactly why
// checking them separately is worth doing.
func comparePCRCVDesign(t *testing.T, data types.Matrix, y []float64,
	ref *pcrReference, design pcrCVDesign) {
	t.Helper()

	maxComponents := design.Curve[len(design.Curve)-1].NComponents

	config := types.PCRConfig{
		PCA: types.PCAConfig{
			Components:    maxComponents,
			MeanCenter:    true,
			StandardScale: ref.Preprocessing == "standardize",
			Method:        "svd",
		},
		Response: ref.Response,
		Selection: types.SelectionConfig{
			Mode:   "cv",
			Metric: "rmse",
			Rule:   types.SelectMin,
			CV: types.CVConfig{
				Scheme: types.CVContiguous,
				Folds:  0, // one fold per group
				Groups: design.Groups,
			},
		},
	}

	result, err := core.NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	if result.CV == nil {
		t.Fatal("no cross-validation report")
	}
	// CVReport.Folds now records the number of folds actually built rather than
	// the number configured, so this comparison means something. It did not
	// before: the config value is zero for both designs here, and the assertion
	// carried an "|| Folds != 0" escape hatch that made it unfireable for
	// exactly the layouts it was written to check.
	if result.CV.Folds != design.NFolds {
		t.Errorf("design %s: GoPCA built %d folds, scikit-learn used %d",
			design.Design, result.CV.Folds, design.NFolds)
	}

	comparePCRCurve(t, design.Design, result.CV, design.Curve, ref.NFeatures)
}

// comparePCRCurve compares a GoPCA error curve against a reference one, point by
// point, including the per-fold statistics.
func comparePCRCurve(t *testing.T, label string, report *types.CVReport,
	curve []pcrCVPoint, nFeatures int) {
	t.Helper()

	// The same rule the point-fit comparison uses. Selecting the tolerance by
	// dataset name would drift from it the moment a second high-dimensional
	// fixture arrives under a different name, and the loosening is a property of
	// the conditioning rather than of the file.
	tolerance := cvTolerance(nFeatures)

	index := make(map[int]int, len(report.Candidates))
	for i, k := range report.Candidates {
		index[k] = i
	}

	for _, point := range curve {
		i, ok := index[point.NComponents]
		if !ok {
			t.Errorf("%s: GoPCA has no candidate k=%d", label, point.NComponents)
			continue
		}
		for _, c := range []struct {
			name string
			got  float64
			want float64
		}{
			{"rmsecv", report.RMSECV[i], point.RMSECV},
			{"rmsecv_mean", report.RMSECVMean[i], point.RMSECVMean},
			{"rmsecv_se", report.RMSECVSE[i], point.RMSECVSE},
			{"mae", report.MAE[i], point.MAE},
			{"mae_se", report.MAESE[i], point.MAESE},
			{"bias", report.Bias[i], point.Bias},
			{"sep", report.SEP[i], point.SEP},
			{"q2", report.Q2[i], point.Q2},
		} {
			if d := relativeDifference(c.got, c.want); d > tolerance {
				t.Errorf("%s k=%d %s: Go %.12g, scikit-learn %.12g (relative %.3g)",
					label, point.NComponents, c.name, c.got, c.want, d)
			}
		}
	}
}

// comparePCRSemiSupervised checks the partially-labelled fit.
//
// The reference masks every third row's response and fits the scaler and the
// decomposition on all rows, the regression on the labelled rows alone. GoPCA
// reaches the same arrangement from the other direction: it is handed a
// response with NaN in those positions and works out which rows are labelled
// itself.
//
// The row counts are asserted before the coefficients, because that is where a
// disagreement would actually show. A model fitted on the wrong subset produces
// coefficients that are perfectly self-consistent, and only the count reveals
// which rows produced them.
func comparePCRSemiSupervised(t *testing.T, data types.Matrix, y []float64, ref *pcrReference) {
	t.Helper()

	want := ref.SemiSupervised

	masked := make([]float64, len(y))
	copy(masked, y)
	labelled := make(map[int]bool, len(want.LabelledRows))
	for _, row := range want.LabelledRows {
		labelled[row] = true
	}
	for i := range masked {
		if !labelled[i] {
			masked[i] = math.NaN()
		}
	}

	config := types.PCRConfig{
		PCA: types.PCAConfig{
			Components:    want.NComponents,
			MeanCenter:    true,
			StandardScale: true,
			Method:        "svd",
		},
		Response: ref.Response,
		Selection: types.SelectionConfig{
			Mode:   "fixed",
			Fixed:  want.NComponents,
			Metric: "rmse",
		},
	}

	result, err := core.NewPCREngine().Fit(data, masked, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	if len(result.LabelledRows) != want.NLabelled {
		t.Fatalf("GoPCA fitted the regression on %d rows, the reference on %d",
			len(result.LabelledRows), want.NLabelled)
	}
	for i, row := range result.LabelledRows {
		if row != want.LabelledRows[i] {
			t.Fatalf("labelled row %d: GoPCA says %d, the reference says %d",
				i, row, want.LabelledRows[i])
		}
	}
	// Every row must have entered the decomposition, including the unlabelled
	// ones. Silently dropping them would be a defensible-looking choice that
	// changes the components and every number downstream.
	if got := len(result.PCA.Scores); got != want.NDecompositionRows {
		t.Errorf("the decomposition used %d rows, the reference used %d: an "+
			"unlabelled sample still carries usable predictor structure",
			got, want.NDecompositionRows)
	}

	tolerance := cvTolerance(ref.NFeatures)

	if d := relativeDifference(result.InterceptOriginal, want.Intercept); d > tolerance {
		t.Errorf("intercept: Go %.12g, scikit-learn %.12g (relative %.3g)",
			result.InterceptOriginal, want.Intercept, d)
	}
	if d := relativeDifference(result.RMSEC, want.RMSEC); d > tolerance {
		t.Errorf("RMSEC: Go %.12g, scikit-learn %.12g (relative %.3g)",
			result.RMSEC, want.RMSEC, d)
	}
	if d := relativeDifference(result.R2C, want.R2C); d > tolerance {
		t.Errorf("R2C: Go %.12g, scikit-learn %.12g (relative %.3g)",
			result.R2C, want.R2C, d)
	}

	if len(result.Coefficients) != len(want.Coefficients) {
		t.Fatalf("coefficient counts differ: %d against %d",
			len(result.Coefficients), len(want.Coefficients))
	}
	worst := 0.0
	for j := range want.Coefficients {
		if d := relativeDifference(result.Coefficients[j], want.Coefficients[j]); d > worst {
			worst = d
		}
	}
	if worst > tolerance {
		t.Errorf("worst coefficient disagreement %.3g exceeds tolerance %g", worst, tolerance)
	}
}

// cvTolerance loosens the comparison for wide, strongly collinear fixtures.
//
// Spectral data with hundreds of correlated columns is ill-conditioned enough
// that the two implementations' different orderings of the same arithmetic
// diverge in the last few digits. The threshold matches the one the point-fit
// comparison uses, so the two cannot drift apart.
func cvTolerance(nFeatures int) float64 {
	if nFeatures > 100 {
		return 1e-4
	}
	return 1e-6
}
