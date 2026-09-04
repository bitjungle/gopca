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

	"github.com/bitjungle/gopca/internal/crossval"
	"github.com/bitjungle/gopca/pkg/types"
)

// makeRegressionData builds a matrix whose response is a linear function of a
// few latent directions plus noise.
func makeRegressionData(n, p, latent int, noise float64, seed uint64) (types.Matrix, []float64) {
	r := rand.New(rand.NewPCG(seed, 0x5DEECE66D))

	loadings := make([][]float64, latent)
	for i := range loadings {
		loadings[i] = make([]float64, p)
		for j := range loadings[i] {
			loadings[i][j] = r.NormFloat64()
		}
	}
	weights := make([]float64, latent)
	for i := range weights {
		weights[i] = r.NormFloat64() * 2
	}

	data := make(types.Matrix, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		scores := make([]float64, latent)
		for l := range scores {
			scores[l] = r.NormFloat64()
		}
		data[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			var v float64
			for l := 0; l < latent; l++ {
				v += scores[l] * loadings[l][j]
			}
			data[i][j] = v + 0.05*r.NormFloat64()
		}
		for l := 0; l < latent; l++ {
			y[i] += scores[l] * weights[l]
		}
		y[i] += noise * r.NormFloat64()
	}
	return data, y
}

func fixedConfig(components int) types.PCRConfig {
	return types.PCRConfig{
		PCA: types.PCAConfig{
			Components:    components,
			MeanCenter:    true,
			StandardScale: true,
			Method:        "svd",
		},
		Response:  "y",
		Selection: types.SelectionConfig{Mode: "fixed", Fixed: components, Metric: "rmse"},
	}
}

func cvConfig(maxComponents, folds int) types.PCRConfig {
	c := fixedConfig(maxComponents)
	c.Selection = types.SelectionConfig{
		Mode:   "cv",
		Metric: "rmse",
		Rule:   types.SelectMin,
		CV: types.CVConfig{
			Scheme: types.CVRandom,
			Folds:  folds,
			Seed:   17,
		},
	}
	return c
}

// TestPCRZeroComponentsPredictsTheMean pins the baseline every positive component
// count must beat. With nothing retained the model is the intercept alone, and
// its prediction must be the training mean exactly, not merely close to it.
func TestPCRZeroComponentsPredictsTheMean(t *testing.T) {
	data, y := makeRegressionData(40, 6, 3, 0.5, 1)

	engine := NewPCREngine()
	result, err := engine.Fit(data, y, fixedConfig(0))
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	var mean float64
	for _, v := range y {
		mean += v
	}
	mean /= float64(len(y))

	for i, f := range result.Fitted {
		if math.Abs(f-mean) > 1e-12 {
			t.Errorf("row %d fitted %.15g, want the response mean %.15g", i, f, mean)
		}
	}
	if len(result.ScoreCoefficients) != 0 {
		t.Errorf("expected no score coefficients at k=0, got %d", len(result.ScoreCoefficients))
	}

	// Every original-scale coefficient must be zero, so the collapsed form also
	// reduces to the mean.
	for j, b := range result.Coefficients {
		if b != 0 {
			t.Errorf("coefficient %d = %v, want 0 at k=0", j, b)
		}
	}
	if math.Abs(result.InterceptOriginal-mean) > 1e-12 {
		t.Errorf("original-scale intercept %.15g, want %.15g", result.InterceptOriginal, mean)
	}
}

// TestPCRFullRankMatchesOLS checks the identity that anchors the estimator:
// retaining every available component makes PCR ordinary least squares on the
// preprocessed design. If the score-space solve or the mapping back to original
// variables is wrong, this diverges.
func TestPCRFullRankMatchesOLS(t *testing.T) {
	const n, p = 50, 5
	data, y := makeRegressionData(n, p, p, 0.3, 9)

	engine := NewPCREngine()
	result, err := engine.Fit(data, y, fixedConfig(p))
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	if !result.OriginalScaleValid {
		t.Fatal("expected original-scale coefficients to be available")
	}

	ols := ordinaryLeastSquares(t, data, y)
	for j := range ols.beta {
		if math.Abs(result.Coefficients[j]-ols.beta[j]) > 1e-8*(1+math.Abs(ols.beta[j])) {
			t.Errorf("coefficient %d: PCR %.12g, OLS %.12g", j, result.Coefficients[j], ols.beta[j])
		}
	}
	if math.Abs(result.InterceptOriginal-ols.intercept) > 1e-8*(1+math.Abs(ols.intercept)) {
		t.Errorf("intercept: PCR %.12g, OLS %.12g", result.InterceptOriginal, ols.intercept)
	}
}

// TestPCRPipelineEquivalence checks that the collapsed original-scale form and
// the explicit pipeline agree. They are two routes to the same prediction, and
// only the collapsed one is deployed, so a discrepancy would ship silently.
func TestPCRPipelineEquivalence(t *testing.T) {
	data, y := makeRegressionData(45, 8, 4, 0.4, 3)

	for _, k := range []int{1, 3, 6} {
		engine := NewPCREngine()
		result, err := engine.Fit(data, y, fixedConfig(k))
		if err != nil {
			t.Fatalf("k=%d Fit: %v", k, err)
		}
		if !result.OriginalScaleValid {
			t.Fatalf("k=%d: expected original-scale coefficients", k)
		}

		viaPipeline, err := engine.Predict(data)
		if err != nil {
			t.Fatalf("k=%d Predict: %v", k, err)
		}

		for i := range data {
			collapsed := result.InterceptOriginal
			for j := range data[i] {
				collapsed += data[i][j] * result.Coefficients[j]
			}
			if math.Abs(collapsed-viaPipeline[i]) > 1e-8*(1+math.Abs(collapsed)) {
				t.Fatalf("k=%d row %d: collapsed %.12g, pipeline %.12g", k, i, collapsed, viaPipeline[i])
			}
		}
	}
}

// TestPCRPredictReproducesFittedValues closes the loop between fitting and
// predicting: applying the model to its own training rows must return the fitted
// values it reported.
func TestPCRPredictReproducesFittedValues(t *testing.T) {
	data, y := makeRegressionData(35, 7, 3, 0.3, 12)

	engine := NewPCREngine()
	result, err := engine.Fit(data, y, fixedConfig(4))
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	predictions, err := engine.Predict(data)
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	for i, row := range result.LabelledRows {
		if math.Abs(predictions[row]-result.Fitted[i]) > 1e-9*(1+math.Abs(result.Fitted[i])) {
			t.Errorf("row %d: Predict %.12g, Fitted %.12g", row, predictions[row], result.Fitted[i])
		}
	}
}

// TestDecompositionRowsExcludeHeldOutRows is the direct guard against the
// subtlest leak in the whole estimator: letting the held-out rows into a fold's
// decomposition.
//
// It is asserted structurally rather than behaviourally, and deliberately so.
// Measured against the running code, a behavioural check cannot see this bug: the
// decomposition is unsupervised, so giving it the test rows does not help it
// predict a response it never saw, and the pure-noise Q2 check below stays green
// with the leak present. A test that cannot go red for the bug it names is not
// evidence about that bug, so the row selection is checked where it is decided.
func TestDecompositionRowsExcludeHeldOutRows(t *testing.T) {
	labelled := []int{0, 1, 2, 3, 4, 5, 6, 7}
	unlabelled := []int{8, 9, 10}

	splitter := &crossval.GroupKFold{K: 4, Shuffle: true, Seed: 5}
	folds, err := splitter.Split(labelled)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}

	for f, fold := range folds {
		rows := decompositionRows(fold.Train, unlabelled)

		held := make(map[int]struct{}, len(fold.Test))
		for _, row := range fold.Test {
			held[row] = struct{}{}
		}
		for _, row := range rows {
			if _, bad := held[row]; bad {
				t.Errorf("fold %d: held-out row %d was given to the decomposition", f, row)
			}
		}

		if len(rows) != len(fold.Train)+len(unlabelled) {
			t.Errorf("fold %d: decomposition got %d rows, want %d training plus %d unlabelled",
				f, len(rows), len(fold.Train), len(unlabelled))
		}
		for i := 1; i < len(rows); i++ {
			if rows[i] <= rows[i-1] {
				t.Errorf("fold %d: rows are not strictly increasing at %d", f, i)
			}
		}

		// Unlabelled rows must be present: dropping them would quietly discard
		// predictor structure the decomposition is entitled to use.
		present := make(map[int]struct{}, len(rows))
		for _, row := range rows {
			present[row] = struct{}{}
		}
		for _, row := range unlabelled {
			if _, ok := present[row]; !ok {
				t.Errorf("fold %d: unlabelled row %d was withheld from the decomposition", f, row)
			}
		}
	}
}

// TestPCRDetectsLeakage checks that the response cannot reach a fold's model.
//
// The response here is pure noise, independent of every predictor, so no honest
// model can beat predicting the mean and cross-validated Q2 must stay at or below
// zero. The design is deliberately sharp: two folds, so a leak doubles the
// training set, and enough components to interpolate the training rows. Under a
// leak the model fits the held-out rows directly and Q2 climbs well clear of the
// threshold; measured against the running code, an honest fit sits far below it
// while a full leak reaches roughly 0.9.
//
// Note what this does and does not cover. It is sensitive to the response
// reaching a fold's training set. It is NOT sensitive to the held-out predictors
// reaching the decomposition, because the decomposition never uses the response;
// that leak is covered structurally by TestDecompositionRowsExcludeHeldOutRows.
func TestPCRDetectsLeakage(t *testing.T) {
	const n, p = 60, 40
	r := rand.New(rand.NewPCG(4242, 99))

	data := make(types.Matrix, n)
	for i := range data {
		data[i] = make([]float64, p)
		for j := range data[i] {
			data[i][j] = r.NormFloat64()
		}
	}
	y := make([]float64, n)
	for i := range y {
		y[i] = r.NormFloat64()
	}

	engine := NewPCREngine()
	config := cvConfig(25, 2)
	result, err := engine.Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	if result.CV == nil {
		t.Fatal("expected a cross-validation report")
	}

	for i, k := range result.CV.Candidates {
		if q2 := result.CV.Q2[i]; q2 > 0.3 {
			t.Errorf("Q2 = %.3f at %d components on a response that is pure noise; "+
				"a value this high means the response reached the fold's training set", q2, k)
		}
	}

	if q0 := result.CV.Q2[0]; q0 > 0.02 || q0 < -0.2 {
		t.Errorf("Q2 at 0 components = %.4f, want approximately 0", q0)
	}
}

// TestPCRExcludesRowsWithoutAResponse covers the case that is normal rather than
// exceptional in calibration data: a response measured on only part of the rows.
// Those rows must leave the regression but stay in the decomposition, and the
// exclusion must be recorded rather than silent.
func TestPCRExcludesRowsWithoutAResponse(t *testing.T) {
	data, y := makeRegressionData(50, 6, 3, 0.3, 21)

	missing := map[int]bool{2: true, 7: true, 19: true, 33: true, 44: true}
	for row := range missing {
		y[row] = math.NaN()
	}

	engine := NewPCREngine()
	result, err := engine.Fit(data, y, fixedConfig(3))
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	if len(result.ExcludedRows) != len(missing) {
		t.Errorf("recorded %d excluded rows, want %d", len(result.ExcludedRows), len(missing))
	}
	for _, row := range result.ExcludedRows {
		if !missing[row] {
			t.Errorf("row %d was recorded as excluded but its response was observed", row)
		}
	}
	if len(result.LabelledRows) != len(data)-len(missing) {
		t.Errorf("kept %d labelled rows, want %d", len(result.LabelledRows), len(data)-len(missing))
	}
	for _, row := range result.LabelledRows {
		if missing[row] {
			t.Errorf("row %d has no response but was used in the regression", row)
		}
	}
	if len(result.Fitted) != len(result.LabelledRows) {
		t.Errorf("Fitted has %d values but there are %d labelled rows",
			len(result.Fitted), len(result.LabelledRows))
	}

	// The decomposition still saw every row, which is the point of keeping them.
	if rows := len(result.PCA.Scores); rows != len(data) {
		t.Errorf("decomposition used %d rows, want all %d", rows, len(data))
	}
}

// TestPCRRowWisePreprocessingHasNoOriginalScale checks that a model which cannot
// be collapsed into fixed coefficients says so, rather than reporting numbers
// that look usable and are not.
func TestPCRRowWisePreprocessingHasNoOriginalScale(t *testing.T) {
	data, y := makeRegressionData(40, 10, 3, 0.3, 31)

	config := fixedConfig(3)
	config.PCA.SNV = true

	engine := NewPCREngine()
	result, err := engine.Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	if result.OriginalScaleValid {
		t.Error("SNV makes the preprocessing sample dependent; no fixed coefficient vector exists")
	}
	if result.Coefficients != nil {
		t.Errorf("expected no original-scale coefficients, got %d", len(result.Coefficients))
	}

	// Prediction must still work, through the explicit pipeline.
	predictions, err := engine.Predict(data)
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	for i, row := range result.LabelledRows {
		if math.Abs(predictions[row]-result.Fitted[i]) > 1e-9*(1+math.Abs(result.Fitted[i])) {
			t.Errorf("row %d: Predict %.12g, Fitted %.12g", row, predictions[row], result.Fitted[i])
		}
	}
}

func TestPCRRefusesUnsupportedMethods(t *testing.T) {
	data, y := makeRegressionData(30, 5, 2, 0.3, 41)

	for _, method := range []string{"kernel", "temporal", "nonsense"} {
		t.Run(method, func(t *testing.T) {
			config := fixedConfig(2)
			config.PCA.Method = method
			if _, err := NewPCREngine().Fit(data, y, config); err == nil {
				t.Errorf("expected %s to be refused", method)
			}
		})
	}
}

func TestPCRInputValidation(t *testing.T) {
	data, y := makeRegressionData(30, 5, 2, 0.3, 51)

	tests := []struct {
		name   string
		data   types.Matrix
		y      []float64
		mutate func(*types.PCRConfig)
	}{
		{"empty data", types.Matrix{}, y, nil},
		{"response length mismatch", data, y[:10], nil},
		{"negative components", data, y, func(c *types.PCRConfig) { c.Selection.Fixed = -1 }},
		{"unknown selection mode", data, y, func(c *types.PCRConfig) { c.Selection.Mode = "guess" }},
		{"grouped without groups", data, y, func(c *types.PCRConfig) {
			c.Selection.Mode = "cv"
			c.Selection.CV.GroupBy = "batch"
			c.Selection.CV.Folds = 3
		}},
		{"group length mismatch", data, y, func(c *types.PCRConfig) {
			c.Selection.Mode = "cv"
			c.Selection.CV.GroupBy = "batch"
			c.Selection.CV.Folds = 3
			c.Selection.CV.Groups = []int{0, 1}
		}},
		{"unknown scheme", data, y, func(c *types.PCRConfig) {
			c.Selection.Mode = "cv"
			c.Selection.CV.Scheme = "sideways"
			c.Selection.CV.Folds = 3
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := fixedConfig(2)
			if tt.mutate != nil {
				tt.mutate(&config)
			}
			if _, err := NewPCREngine().Fit(tt.data, tt.y, config); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestPCRRejectsInfiniteResponse(t *testing.T) {
	data, y := makeRegressionData(30, 5, 2, 0.3, 61)
	y[4] = math.Inf(1)
	if _, err := NewPCREngine().Fit(data, y, fixedConfig(2)); err == nil {
		t.Error("expected an infinite response value to be refused")
	}
}

func TestPCRRejectsTooFewLabelledRows(t *testing.T) {
	data, y := makeRegressionData(30, 5, 2, 0.3, 71)
	for i := 2; i < len(y); i++ {
		y[i] = math.NaN()
	}
	if _, err := NewPCREngine().Fit(data, y, fixedConfig(1)); err == nil {
		t.Error("expected a fit with two labelled rows to be refused")
	}
}

// TestPCRRepeatedLeaveOneOutRefused guards a claim the configuration makes: a
// repeat count above one implies the partition varies, which it cannot at
// leave-one-out. Silently doing identical work would waste time and report a
// spread of zero as though it meant something.
func TestPCRRepeatedLeaveOneOutRefused(t *testing.T) {
	data, y := makeRegressionData(25, 5, 2, 0.3, 81)

	config := cvConfig(3, 0)
	config.Selection.CV.Repeats = 3
	if _, err := NewPCREngine().Fit(data, y, config); err == nil {
		t.Error("expected repeated leave-one-out to be refused")
	}
}

// ordinaryLeastSquares solves the standardized regression directly, as an
// independent reference for the full-rank identity.
type olsFit struct {
	beta      []float64
	intercept float64
}

func ordinaryLeastSquares(t *testing.T, data types.Matrix, y []float64) olsFit {
	t.Helper()
	n, p := len(data), len(data[0])

	pre := NewPreprocessorWithScaleOnly(true, true, false, false, false, false)
	z, err := pre.FitTransform(data)
	if err != nil {
		t.Fatalf("preprocessing: %v", err)
	}
	center, divisor, err := pre.ColumnAffine()
	if err != nil {
		t.Fatalf("ColumnAffine: %v", err)
	}

	allRows := make([]int, n)
	for i := range allRows {
		allRows[i] = i
	}
	solver, err := newNestedLeastSquares(designMatrix(z, allRows, p), y)
	if err != nil {
		t.Fatalf("least squares: %v", err)
	}
	coefficients, err := solver.Coefficients(p + 1)
	if err != nil {
		t.Fatalf("solve: %v", err)
	}

	beta := make([]float64, p)
	intercept := coefficients[0]
	for j := 0; j < p; j++ {
		beta[j] = coefficients[j+1] / divisor[j]
		intercept -= center[j] * beta[j]
	}
	return olsFit{beta: beta, intercept: intercept}
}

// TestPCRMAESelectionUsesItsOwnStandardError checks that selecting on MAE draws
// its spread from the MAE curve rather than the RMSE one.
//
// The one-standard-error rule adds a standard error to the minimum of a curve.
// Taking that spread from RMSE while applying it to MAE combines two quantities
// with different units of sensitivity to large residuals, and the resulting
// threshold means nothing. The two standard errors genuinely differ, so a
// crossed wiring changes the answer.
func TestPCRMAESelectionUsesItsOwnStandardError(t *testing.T) {
	// This seed and noise level are chosen because the two standard errors lead
	// to different answers here: applying the RMSE spread to the MAE curve picks
	// one component where the MAE spread picks three. On quieter data the two
	// agree and the test would pass whichever way the wiring ran, proving nothing.
	data, y := makeRegressionData(50, 8, 3, 2.5, 13)

	config := cvConfig(6, 5)
	config.Selection.Metric = "mae"
	config.Selection.Rule = types.SelectOneSE

	result, err := NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	report := result.CV

	if len(report.MAESE) != len(report.Candidates) {
		t.Fatalf("MAESE has %d values but there are %d candidates",
			len(report.MAESE), len(report.Candidates))
	}

	differs := false
	for i := range report.MAESE {
		if report.MAESE[i] != report.RMSECVSE[i] {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("the MAE and RMSE standard errors are identical at every candidate, " +
			"which suggests one is being copied from the other")
	}

	// Selecting on MAE with the MAE spread must reproduce what the rule gives when
	// applied directly to those two arrays.
	want, err := SelectComponents(report.Candidates, report.MAE, report.MAESE,
		types.SelectOneSE, 0, 0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if report.Selected != want {
		t.Errorf("selected %d components, but the MAE curve with its own standard "+
			"error gives %d", report.Selected, want)
	}

	// Confirm the check is capable of failing: on this data the crossed wiring
	// gives a different answer, so passing above is evidence rather than luck.
	crossed, err := SelectComponents(report.Candidates, report.MAE, report.RMSECVSE,
		types.SelectOneSE, 0, 0)
	if err != nil {
		t.Fatalf("SelectComponents: %v", err)
	}
	if crossed == want {
		t.Fatal("this dataset no longer distinguishes the two standard errors, so the " +
			"test cannot detect the wiring it exists to check; choose different data")
	}
}

// TestPCRAlternateMetricIsTheOtherOne checks that the reported alternative is
// genuinely the other measure, whichever was primary. Naming the field after MAE
// would have made it wrong exactly when MAE was the selection metric.
// TestCVReportNamesTheCurveItSelectedOn checks that the report says which of the
// two error curves the rule was applied to.
//
// Selected, LowestError, CurveStillFalling and OutOfFold all refer to one curve,
// and until this field existed a consumer had to guess which. The desktop panel
// guessed RMSECV, so choosing MAE produced a screen that plotted one curve and
// credited it with a decision taken on the other. Nothing in the engine's own
// tests could see that, because the engine was right and only its report was
// ambiguous.
func TestCVReportNamesTheCurveItSelectedOn(t *testing.T) {
	data, y := makeRegressionData(40, 6, 3, 1.0, 7)

	for _, tc := range []struct {
		metric string
		want   string
	}{
		{types.MetricRMSE, types.MetricRMSE},
		{types.MetricMAE, types.MetricMAE},
		{"", types.MetricRMSE}, // unset means RMSE, as SelectionConfig documents
	} {
		config := cvConfig(4, 5)
		config.Selection.Metric = tc.metric

		result, err := NewPCREngine().Fit(data, y, config)
		if err != nil {
			t.Fatalf("Fit with metric %q: %v", tc.metric, err)
		}
		if result.CV.Metric != tc.want {
			t.Errorf("metric %q: report says %q, want %q",
				tc.metric, result.CV.Metric, tc.want)
		}

		// The name has to match the curve the numbers actually came from, not
		// merely echo the request. Selected must be reproducible from the curve
		// the report names.
		curve, se := result.CV.RMSECV, result.CV.RMSECVSE
		if result.CV.Metric == types.MetricMAE {
			curve, se = result.CV.MAE, result.CV.MAESE
		}
		want, err := SelectComponents(result.CV.Candidates, curve, se,
			result.CV.Rule, config.Selection.Tolerance, config.Selection.WoldR)
		if err != nil {
			t.Fatalf("SelectComponents: %v", err)
		}
		if result.CV.Selected != want {
			t.Errorf("metric %q: report names %q but Selected=%d, while that curve gives %d",
				tc.metric, result.CV.Metric, result.CV.Selected, want)
		}
	}
}

// TestCurveStillFallingUsesTheSelectedMetricsSpread checks that the margin
// deciding "the error was still falling" is drawn from the curve being read.
//
// CurveStillFalling compares the last step down against the noise around it. The
// noise has to be that curve's own: measuring the scatter of RMSE across folds
// and subtracting it from a step in MAE compares two different quantities, and
// the RMSE spread is systematically the larger of the two, so the error is not
// symmetric. It suppresses the advice to raise the ceiling on exactly the runs
// where the curve really was still descending.
//
// The data below is chosen so the two answers differ: the final step is larger
// than the MAE standard error and smaller than the RMSE one. On quieter data both
// wirings agree and this test would pass without looking.
func TestCurveStillFallingUsesTheSelectedMetricsSpread(t *testing.T) {
	data, y := makeRegressionData(50, 10, 3, 0.5, 57)

	config := cvConfig(6, 5)
	config.Selection.Metric = types.MetricMAE

	result, err := NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}
	report := result.CV

	n := len(report.MAE)
	if n < 2 {
		t.Fatalf("need at least two candidates, got %d", n)
	}
	gap := report.MAE[n-2] - report.MAE[n-1]

	// Guard the fixture before trusting the assertion. If the two standard errors
	// ever stop straddling the gap, the assertion below holds for both the right
	// and the wrong wiring and proves nothing.
	if !(report.MAESE[n-1] < gap && gap < report.RMSECVSE[n-1]) {
		t.Fatalf("this dataset no longer distinguishes the two spreads "+
			"(gap %.6g, MAE SE %.6g, RMSE SE %.6g); choose different data",
			gap, report.MAESE[n-1], report.RMSECVSE[n-1])
	}

	if !report.CurveStillFalling {
		t.Error("CurveStillFalling is false, but the final step in the MAE curve " +
			"exceeds the MAE standard error; the margin appears to have been taken " +
			"from the RMSE curve, which is not the curve being read")
	}
}

func TestPCRAlternateMetricIsTheOtherOne(t *testing.T) {
	data, y := makeRegressionData(60, 8, 3, 0.6, 103)

	for _, metric := range []string{"rmse", "mae"} {
		t.Run(metric, func(t *testing.T) {
			config := cvConfig(6, 5)
			config.Selection.Metric = metric
			config.Selection.Rule = types.SelectMin

			result, err := NewPCREngine().Fit(data, y, config)
			if err != nil {
				t.Fatalf("Fit: %v", err)
			}
			report := result.CV

			other, otherSE := report.MAE, report.MAESE
			if metric == "mae" {
				other, otherSE = report.RMSECV, report.RMSECVSE
			}
			want, err := SelectComponents(report.Candidates, other, otherSE,
				types.SelectMin, 0, 0)
			if err != nil {
				t.Fatalf("SelectComponents: %v", err)
			}
			if report.SelectedByAlternateMetric != want {
				t.Errorf("SelectedByAlternateMetric = %d, want %d (the choice the other "+
					"measure would have made)", report.SelectedByAlternateMetric, want)
			}
		})
	}
}

// TestPCRLeaveOneOutComponentCeiling checks that leave-one-out bounds the
// component count by its training partition, which is one row shorter than the
// dataset.
//
// The ceiling is resolved from a fold count of zero, which means one fold per
// group. Treating zero as "no fold constraint" would let the ceiling be computed
// from the full row count and overstate what any fold can actually fit.
func TestPCRLeaveOneOutComponentCeiling(t *testing.T) {
	const n, p = 12, 20
	data, y := makeRegressionData(n, p, 4, 0.3, 107)

	// Ask for far more components than the data can support, so the ceiling is
	// what decides the answer rather than the request.
	config := cvConfig(n, 0)
	config.Selection.Rule = types.SelectMin

	result, err := NewPCREngine().Fit(data, y, config)
	if err != nil {
		t.Fatalf("Fit: %v", err)
	}

	// Each training partition holds n-1 rows. Centring costs one degree of
	// freedom and the score-space intercept another, leaving n-3.
	ceiling := n - 3
	last := result.CV.Candidates[len(result.CV.Candidates)-1]
	if last > ceiling {
		t.Errorf("the sweep evaluated %d components, but a leave-one-out training "+
			"partition of %d rows supports at most %d", last, n-1, ceiling)
	}
}
