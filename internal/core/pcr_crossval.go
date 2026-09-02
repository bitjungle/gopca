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
	"sort"

	"github.com/bitjungle/gopca/internal/crossval"
	"github.com/bitjungle/gopca/pkg/types"
)

// crossValidate estimates prediction error for every candidate component count
// and applies the selection rule.
//
// Every learned operation is refitted inside each fold: missing-value handling,
// centring, scaling and the decomposition itself. Doing any of them once over
// the whole dataset would let information from the held-out rows reach the model
// and make the estimate optimistic. That failure is invisible in ordinary
// testing, because the only symptom is that the numbers look better than they
// should, which is why TestPCRDetectsLeakage exists.
//
// The sweep costs one decomposition per fold, not one per fold per candidate.
// The decomposition does not depend on the component count, and the least-squares
// solutions for nested component counts all come from a single factorization; see
// nestedLeastSquares.
//
// Algorithm complexity: O(F R n p min(n,p)) for F folds and R repeats, dominated
// by the decompositions.
func (p *PCRImpl) crossValidate(data types.Matrix, y []float64, labelled []int,
	config types.PCRConfig, kMax int) (*types.CVReport, error) {

	cv := config.Selection.CV
	repeats := cv.Repeats
	if repeats < 1 {
		repeats = 1
	}

	groupCount := countGroups(labelled, cv.Groups)
	if repeats > 1 && (cv.Folds == 0 || cv.Folds == groupCount) {
		return nil, fmt.Errorf(
			"repeating a leave-one-out design has no effect: with %d folds over %d groups the "+
				"partition is unique, so every repeat would produce identical folds. "+
				"Use fewer folds, or a single repeat",
			groupCount, groupCount)
	}

	// Rows whose response was not observed. They inform each fold's decomposition
	// but never enter a test fold, having nothing to be scored against.
	unlabelled := complementRows(len(data), labelled)

	// predicted[c] and measured[c] accumulate every out-of-fold pair for candidate
	// c across all folds and repeats, so the pooled metrics weight each
	// observation equally regardless of fold size.
	predicted := make([][]float64, kMax+1)
	measured := make([][]float64, kMax+1)
	foldRMSE := make([][]float64, kMax+1)

	// firstRepeatOOF keeps one held-out prediction per labelled row, taken from
	// the first repeat, so that a caller can plot predicted against measured.
	firstRepeatOOF := make([][]float64, kMax+1)
	for c := range firstRepeatOOF {
		firstRepeatOOF[c] = make([]float64, len(labelled))
		for i := range firstRepeatOOF[c] {
			firstRepeatOOF[c][i] = math.NaN()
		}
	}

	positionOf := make(map[int]int, len(labelled))
	for i, row := range labelled {
		positionOf[row] = i
	}

	// maxEvaluable falls to the smallest component count every fold could
	// actually fit. Candidates beyond it are dropped from the report rather than
	// being compared on unequal evidence.
	maxEvaluable := kMax
	design := ""

	for repeat := 0; repeat < repeats; repeat++ {
		splitter, err := buildSplitter(cv, int64(repeat))
		if err != nil {
			return nil, err
		}
		if repeat == 0 {
			design = splitter.Name()
		}

		folds, err := splitter.Split(labelled)
		if err != nil {
			return nil, err
		}

		for foldIndex, fold := range folds {
			usable, err := p.evaluateFold(data, y, fold, unlabelled, config, kMax,
				predicted, measured, foldRMSE,
				firstRepeatOOF, positionOf, repeat == 0)
			if err != nil {
				return nil, fmt.Errorf("repeat %d fold %d: %w", repeat+1, foldIndex+1, err)
			}
			if usable < maxEvaluable {
				maxEvaluable = usable
			}
		}
	}

	if maxEvaluable < 0 {
		return nil, fmt.Errorf("no candidate component count could be fitted in every fold")
	}

	report := &types.CVReport{
		Scheme:   schemeName(cv.Scheme),
		Design:   design,
		Folds:    cv.Folds,
		Repeats:  repeats,
		GroupBy:  cv.GroupBy,
		Seed:     cv.Seed,
		NSamples: len(labelled),
		Rule:     config.Selection.Rule,
	}

	for k := 0; k <= maxEvaluable; k++ {
		metrics, err := ComputeRegressionMetrics(predicted[k], measured[k])
		if err != nil {
			return nil, fmt.Errorf("scoring candidate %d components: %w", k, err)
		}
		report.Candidates = append(report.Candidates, k)
		report.RMSECV = append(report.RMSECV, metrics.RMSE)
		report.Bias = append(report.Bias, metrics.Bias)
		report.SEP = append(report.SEP, metrics.SEP)
		report.MAE = append(report.MAE, metrics.MAE)
		report.Q2 = append(report.Q2, metrics.R2)

		mean, se := meanAndStandardError(foldRMSE[k])
		report.RMSECVMean = append(report.RMSECVMean, mean)
		report.RMSECVSE = append(report.RMSECVSE, se)
	}

	rule := config.Selection.Rule
	if rule == "" {
		rule = types.SelectOneSE
		report.Rule = rule
	}

	curve := report.RMSECV
	if config.Selection.Metric == "mae" {
		curve = report.MAE
	}

	selected, err := SelectComponents(report.Candidates, curve, report.RMSECVSE,
		rule, config.Selection.Tolerance, config.Selection.WoldR)
	if err != nil {
		return nil, err
	}
	report.Selected = selected

	// What the other metric would have chosen. When the two disagree, a few large
	// residuals are driving the choice, which is worth surfacing rather than
	// resolving silently.
	alternative := report.MAE
	if config.Selection.Metric == "mae" {
		alternative = report.RMSECV
	}
	if byOther, err := SelectComponents(report.Candidates, alternative, report.RMSECVSE,
		rule, config.Selection.Tolerance, config.Selection.WoldR); err == nil {
		report.SelectedByMAE = byOther
	}

	report.OutOfFold = firstRepeatOOF[selected]
	return report, nil
}

// evaluateFold fits one fold and records its held-out predictions for every
// candidate component count. It returns the largest component count this fold
// could fit.
func (p *PCRImpl) evaluateFold(data types.Matrix, y []float64, fold crossval.Fold,
	unlabelled []int, config types.PCRConfig, kMax int,
	predicted, measured, foldRMSE [][]float64,
	firstRepeatOOF [][]float64, positionOf map[int]int, recordOOF bool) (int, error) {

	if len(fold.Train) == 0 || len(fold.Test) == 0 {
		return 0, fmt.Errorf("fold has %d training and %d test rows", len(fold.Train), len(fold.Test))
	}

	pcaRows := decompositionRows(fold.Train, unlabelled)

	pcaConfig := config.PCA
	pcaConfig.Components = kMax

	engine := &PCAImpl{}
	foldResult, err := engine.Fit(subsetRows(data, pcaRows), pcaConfig)
	if err != nil {
		return 0, fmt.Errorf("decomposition failed: %w", err)
	}

	rowPosition := make(map[int]int, len(pcaRows))
	for i, row := range pcaRows {
		rowPosition[row] = i
	}
	trainPositions := make([]int, len(fold.Train))
	for i, row := range fold.Train {
		trainPositions[i] = rowPosition[row]
	}

	available := kMax
	if foldResult.ComponentsComputed < available {
		available = foldResult.ComponentsComputed
	}

	trainMeasured := gather(y, fold.Train)
	solver, err := newNestedLeastSquares(
		designMatrix(foldResult.Scores, trainPositions, available), trainMeasured)
	if err != nil {
		return 0, fmt.Errorf("score-space regression failed: %w", err)
	}
	// One design column is the intercept, so rank r supports r-1 components.
	if solver.Rank()-1 < available {
		available = solver.Rank() - 1
	}
	if available < 0 {
		return 0, fmt.Errorf("no usable components in this fold")
	}

	testScores, err := engine.Transform(subsetRows(data, fold.Test))
	if err != nil {
		return 0, fmt.Errorf("projecting the held-out rows failed: %w", err)
	}
	testMeasured := gather(y, fold.Test)

	for k := 0; k <= available; k++ {
		coefficients, err := solver.Coefficients(k + 1)
		if err != nil {
			return 0, err
		}

		var sumSq float64
		for i := range fold.Test {
			prediction := coefficients[0]
			for j := 0; j < k; j++ {
				prediction += coefficients[j+1] * testScores[i][j]
			}
			predicted[k] = append(predicted[k], prediction)
			measured[k] = append(measured[k], testMeasured[i])

			residual := prediction - testMeasured[i]
			sumSq += residual * residual

			if recordOOF {
				firstRepeatOOF[k][positionOf[fold.Test[i]]] = prediction
			}
		}
		foldRMSE[k] = append(foldRMSE[k], math.Sqrt(sumSq/float64(len(fold.Test))))
	}

	return available, nil
}

// meanAndStandardError summarises per-fold errors.
//
// The standard error across folds is a description of how much the estimate moved
// between partitions. It is not a valid standard error for deployment error,
// because the folds reuse the same observations and their errors are therefore
// correlated. It is used only by the one-standard-error selection rule, which is
// a heuristic for parsimony rather than an inferential statement.
func meanAndStandardError(values []float64) (mean, standardError float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}
	for _, v := range values {
		mean += v
	}
	mean /= float64(n)
	if n < 2 {
		return mean, 0
	}
	var sumSq float64
	for _, v := range values {
		d := v - mean
		sumSq += d * d
	}
	return mean, math.Sqrt(sumSq/float64(n-1)) / math.Sqrt(float64(n))
}

// decompositionRows returns the rows a fold's decomposition is allowed to see:
// the training rows, plus every row whose response was never observed.
//
// The held-out rows must not appear. Excluding them from the regression but not
// from the decomposition is the leak that is hardest to notice, because the
// decomposition is unsupervised: the resulting error estimate is only slightly
// too good, and no behavioural test on the numbers reliably separates it from an
// honest one. That is why the row selection is a named function with a direct
// test rather than three lines inlined into the fold loop; see
// TestDecompositionRowsExcludeHeldOutRows.
//
// Unlabelled rows are included on purpose. PCA does not use the response, so a
// row with predictors but no measured response still carries usable structure,
// and in calibration data such rows are often the majority.
func decompositionRows(train, unlabelled []int) []int {
	rows := make([]int, 0, len(train)+len(unlabelled))
	rows = append(rows, train...)
	rows = append(rows, unlabelled...)
	sort.Ints(rows)
	return rows
}

// complementRows returns every row index below n that is not in subset.
func complementRows(n int, subset []int) []int {
	in := make(map[int]struct{}, len(subset))
	for _, row := range subset {
		in[row] = struct{}{}
	}
	out := make([]int, 0, n-len(subset))
	for row := 0; row < n; row++ {
		if _, skip := in[row]; !skip {
			out = append(out, row)
		}
	}
	return out
}

// subsetRows copies the given rows into a new matrix, preserving their order.
func subsetRows(data types.Matrix, rows []int) types.Matrix {
	out := make(types.Matrix, len(rows))
	for i, row := range rows {
		out[i] = make([]float64, len(data[row]))
		copy(out[i], data[row])
	}
	return out
}

// countGroups reports how many distinct groups the given rows fall into.
func countGroups(rows []int, groups []int) int {
	if groups == nil {
		return len(rows)
	}
	seen := make(map[int]struct{}, len(rows))
	for _, row := range rows {
		if row >= 0 && row < len(groups) {
			seen[groups[row]] = struct{}{}
		}
	}
	return len(seen)
}

// schemeName normalises the recorded scheme, so that an empty configuration is
// reported as what it actually did rather than as a blank.
func schemeName(scheme string) string {
	if scheme == "" {
		return types.CVRandom
	}
	return scheme
}
