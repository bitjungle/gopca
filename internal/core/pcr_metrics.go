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

	"github.com/bitjungle/gopca/pkg/types"
)

// RegressionMetrics summarises how far predictions fall from measured values,
// all in the units of the response.
//
// RMSE alone cannot distinguish a model that is scattered from one that is
// systematically offset, and the two call for different remedies: an offset is
// often repairable by a slope-and-bias correction, imprecision is not. The three
// are related exactly, which is the property TestRegressionMetricsDecomposition
// asserts:
//
//	RMSE² = Bias² + (n-1)/n · SEP²
//
// Reference: Naes et al. (2002), A User-Friendly Guide to Multivariate
// Calibration and Classification, Ch. 13, on RMSEP, bias and SEP.
type RegressionMetrics struct {
	// RMSE is the root mean square of the residuals.
	RMSE float64

	// Bias is the mean signed residual (predicted minus measured). A large bias
	// with a small SEP is a precise model with the wrong offset, the signature of
	// instrument drift or of transferring a calibration between instruments.
	Bias float64

	// SEP is the standard error of prediction: the bias-corrected precision.
	SEP float64

	// MAE is the mean absolute residual. It is driven by the typical error where
	// RMSE is driven by the largest, so the two failing differently is what makes
	// their disagreement informative.
	MAE float64

	// R2 is the coefficient of determination against the mean of the measured
	// values. Computed from held-out predictions it is the quantity usually
	// written Q².
	R2 float64

	// N is the number of prediction/measurement pairs scored.
	N int
}

// ComputeRegressionMetrics scores predictions against measured values.
//
// Both slices must be the same length and hold only finite values; callers are
// expected to have excluded rows without an observed response already.
//
// Algorithm complexity: O(n).
func ComputeRegressionMetrics(predicted, measured []float64) (RegressionMetrics, error) {
	n := len(measured)
	if n == 0 {
		return RegressionMetrics{}, fmt.Errorf("no observations to score")
	}
	if len(predicted) != n {
		return RegressionMetrics{}, fmt.Errorf(
			"predicted has %d values but measured has %d", len(predicted), n)
	}

	var sumSq, sumErr, sumAbs, sumY float64
	for i := 0; i < n; i++ {
		if !isFinite(predicted[i]) || !isFinite(measured[i]) {
			return RegressionMetrics{}, fmt.Errorf(
				"non-finite value at position %d: predicted %v, measured %v",
				i, predicted[i], measured[i])
		}
		e := predicted[i] - measured[i]
		sumSq += e * e
		sumErr += e
		sumAbs += math.Abs(e)
		sumY += measured[i]
	}

	m := RegressionMetrics{N: n}
	m.RMSE = math.Sqrt(sumSq / float64(n))
	m.Bias = sumErr / float64(n)
	m.MAE = sumAbs / float64(n)

	// SEP is the sample standard deviation of the residuals about their mean. It
	// needs at least two observations; with one there is no spread to estimate.
	if n > 1 {
		var sumCentred float64
		for i := 0; i < n; i++ {
			d := (predicted[i] - measured[i]) - m.Bias
			sumCentred += d * d
		}
		m.SEP = math.Sqrt(sumCentred / float64(n-1))
	}

	// R² against the mean of the measured values. Undefined when the response is
	// constant, since there is no variance to explain; reported as zero rather
	// than as an infinity that would propagate into a selection rule.
	meanY := sumY / float64(n)
	var totalSS float64
	for i := 0; i < n; i++ {
		d := measured[i] - meanY
		totalSS += d * d
	}
	if totalSS > 0 {
		m.R2 = 1 - sumSq/totalSS
	}

	return m, nil
}

// isFinite reports whether v is a usable number.
func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// SelectComponents turns a cross-validated error curve into a choice of
// component count.
//
// curve and standardError are parallel to candidates. standardError may be nil
// for rules that do not need it.
//
// The rules differ in how much evidence they demand before accepting a more
// complex model. SelectMin accepts any improvement, however small and however
// likely to be noise; the others require the improvement to be worth the
// components it costs. See the Rule constants in pkg/types for the trade-off.
func SelectComponents(candidates []int, curve, standardError []float64,
	rule string, tolerance, woldR float64) (int, error) {

	if len(candidates) == 0 {
		return 0, fmt.Errorf("no candidate component counts to choose from")
	}
	if len(curve) != len(candidates) {
		return 0, fmt.Errorf("error curve has %d points but there are %d candidates",
			len(curve), len(candidates))
	}

	best := 0
	for i, v := range curve {
		if v < curve[best] {
			best = i
		}
	}

	switch rule {
	case types.SelectMin, "":
		return candidates[best], nil

	case types.SelectOneSE:
		if standardError == nil || len(standardError) != len(candidates) {
			return 0, fmt.Errorf("the one-standard-error rule needs a standard error per candidate")
		}
		threshold := curve[best] + standardError[best]
		for i := range candidates {
			if curve[i] <= threshold {
				return candidates[i], nil
			}
		}
		return candidates[best], nil

	case types.SelectTolerance:
		if tolerance < 0 {
			return 0, fmt.Errorf("tolerance must not be negative, got %v", tolerance)
		}
		threshold := curve[best] + tolerance
		for i := range candidates {
			if curve[i] <= threshold {
				return candidates[i], nil
			}
		}
		return candidates[best], nil

	case types.SelectFirstMin:
		// Walk forward and stop where the curve first turns upward by more than
		// the noise around it. The margin matters: a curve that dips, flattens
		// within a hair, then continues down would otherwise stop on the first
		// insignificant wiggle rather than at the shoulder a reader would pick.
		// Where no standard error is available, a relative margin stands in.
		for i := 0; i+1 < len(curve); i++ {
			margin := 0.0
			if standardError != nil && i < len(standardError) {
				margin = standardError[i]
			} else {
				margin = 0.01 * math.Abs(curve[i])
			}
			if curve[i+1] > curve[i]+margin {
				return candidates[i], nil
			}
		}
		// The curve never turned, so it was still descending at the ceiling.
		return candidates[len(candidates)-1], nil

	case types.SelectWold:
		r := woldR
		if r <= 0 {
			r = 1.0
		}
		// Wold's R is defined on PRESS, the sum of squared held-out residuals, not
		// on its square root. Since curve holds RMSE, the PRESS ratio is the square
		// of the RMSE ratio. Applying the threshold to the RMSE ratio directly
		// would silently make it far stricter: a nominal 0.95 would demand a 10%
		// improvement in PRESS rather than 5%.
		//
		// Note the sharp edge in this rule. It stops at the *first* candidate whose
		// successor fails to improve enough, so a curve that plateaus before
		// falling again will stop on the plateau. That is not hypothetical: on
		// testdata/corn/corn.csv predicting Moisture, RMSECV runs
		// 0.397, 0.327, 0.337, 0.254, ... and only collapses to 0.060 at seven
		// components, so this rule stops at one and reports a model explaining a
		// quarter of the variance where 99% was available. The behaviour matches
		// the published criterion; it is offered because practitioners expect it,
		// and it is not the default for exactly this reason.
		for i := 0; i+1 < len(curve); i++ {
			if curve[i] <= 0 {
				return candidates[i], nil
			}
			ratio := curve[i+1] / curve[i]
			if ratio*ratio > r {
				return candidates[i], nil
			}
		}
		return candidates[len(candidates)-1], nil

	default:
		return 0, fmt.Errorf("unknown selection rule %q: expected one of %q, %q, %q, %q, %q",
			rule, types.SelectMin, types.SelectOneSE, types.SelectTolerance,
			types.SelectWold, types.SelectFirstMin)
	}
}
