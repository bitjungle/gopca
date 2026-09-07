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

package transform

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Lambda is searched over this interval, which is the range used in practice
// and by scipy's implementations. Values outside it correspond to transforms so
// severe that the result is dominated by the transform rather than the data.
const (
	lambdaMin = -5.0
	lambdaMax = 5.0
	// lambdaTolerance is the width below which the search stops. Four decimal
	// places is finer than the parameter is ever reported to and far finer than
	// the difference makes to the transformed values.
	lambdaTolerance = 1e-6
)

// applyPower applies Box-Cox or Yeo-Johnson to each named column.
//
// # Why these and not another root or logarithm
//
// The fixed-exponent transforms in this package -- log, square root, square --
// require the user to guess how skewed the data is. Box-Cox and Yeo-Johnson fit
// their exponent to the column by maximum likelihood instead, so the strength
// of the correction comes from the data.
//
// Box-Cox needs strictly positive values. Yeo-Johnson is the extension that
// handles zero and negative ones, which is exactly where log and sqrt refuse
// (#861), and so is the answer for zero-inflated concentration and count data.
//
// # What this is for, and what it is not for
//
// PCA makes no normality assumption, so these are not applied to satisfy one.
// The reason to use them is that a strongly skewed variable exerts leverage out
// of proportion to its information: a handful of large values can dominate a
// component through their distance from the mean alone. Reducing the skew
// reduces that leverage.
//
// # References
//
//	Box, G.E.P. & Cox, D.R. (1964). An Analysis of Transformations. Journal of
//	the Royal Statistical Society, Series B 26(2), 211-252.
//	Yeo, I.-K. & Johnson, R.A. (2000). A New Family of Power Transformations to
//	Improve Normality or Symmetry. Biometrika 87(4), 954-959.
//
// Algorithm complexity: O(n log(1/ε)) per column for the golden-section search.
func applyPower(data [][]string, columnTypes map[string]string, headers []string, opts Options, result *Result) error {
	for _, colName := range opts.Columns {
		colIndex := findColumn(headers, colName)
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		if columnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf(
				"Column '%s' is not numeric, skipping", colName))
			continue
		}

		values, rows := numericColumn(data, colIndex)
		if len(values) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf(
				"Column '%s' has no values to transform", colName))
			continue
		}

		// Box-Cox is undefined at and below zero. Refuse the whole column
		// rather than transforming part of it, for the reason #861 established:
		// a column holding some transformed and some raw values carries two
		// scales in one variable and nothing downstream can detect it.
		if opts.Type == BoxCox {
			var offending []int
			for i, value := range values {
				if value <= 0 {
					offending = append(offending, rows[i]+1)
				}
			}
			if len(offending) > 0 {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"Column '%s' left unchanged: Box-Cox is undefined for the "+
						"non-positive values in %s. Yeo-Johnson handles them",
					colName, describeRows(offending)))
				continue
			}
		}

		lambda := 0.0
		estimated := false
		if opts.Lambda != nil {
			lambda = *opts.Lambda
		} else {
			lambda = estimateLambda(values, opts.Type)
			estimated = true
		}

		for i, value := range values {
			data[rows[i]][colIndex] = fmt.Sprintf("%.6g", powerTransform(value, lambda, opts.Type))
		}

		how := fmt.Sprintf("λ = %.4f, fitted by maximum likelihood", lambda)
		if !estimated {
			how = fmt.Sprintf("λ = %.4f, as supplied", lambda)
		}
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf(
			"%s applied to '%s' (%s)", transformDisplayName(opts.Type), colName, how))
	}

	return nil
}

// transformDisplayName renders a transform's name for a message.
func transformDisplayName(t Type) string {
	if t == BoxCox {
		return "Box-Cox"
	}
	return "Yeo-Johnson"
}

// numericColumn returns the parseable values of a column and the row indices
// they came from.
//
// Blanks and unparseable cells are excluded and left untouched, as they are for
// the other numeric transforms: they were never in the column's units, so
// leaving one cannot put the column into mixed units.
func numericColumn(data [][]string, colIndex int) ([]float64, []int) {
	var values []float64
	var rows []int
	for i := range data {
		if colIndex >= len(data[i]) {
			continue
		}
		text := strings.TrimSpace(data[i][colIndex])
		if text == "" {
			continue
		}
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			continue
		}
		values = append(values, number)
		rows = append(rows, i)
	}
	return values, rows
}

// powerTransform applies one transform at a given lambda.
//
// Both families are written so the lambda = 0 case joins continuously with its
// neighbours: (x^λ − 1)/λ tends to ln(x) as λ tends to zero, which is why the
// logarithm is the λ = 0 member rather than a special case bolted on.
func powerTransform(x, lambda float64, t Type) float64 {
	if t == BoxCox {
		if lambda == 0 {
			return math.Log(x)
		}
		return (math.Pow(x, lambda) - 1) / lambda
	}

	// Yeo-Johnson, which is defined on the whole real line by treating the two
	// sides separately -- and by using 2−λ on the negative side, which is what
	// makes the two halves join smoothly at zero.
	if x >= 0 {
		if lambda == 0 {
			return math.Log1p(x)
		}
		return (math.Pow(x+1, lambda) - 1) / lambda
	}
	if lambda == 2 {
		return -math.Log1p(-x)
	}
	return -(math.Pow(-x+1, 2-lambda) - 1) / (2 - lambda)
}

// logLikelihood is the profile log-likelihood of lambda, up to a constant.
//
//	ℓ(λ) = −(n/2)·ln(σ²(λ)) + (λ−1)·Σ ln|Jacobian|
//
// The first term rewards a transform that makes the values less dispersed on a
// log scale; the second is the Jacobian of the transform, without which the
// likelihood could be made arbitrarily large by shrinking everything towards a
// point.
func logLikelihood(values []float64, lambda float64, t Type) float64 {
	n := float64(len(values))

	transformed := make([]float64, len(values))
	for i, value := range values {
		transformed[i] = powerTransform(value, lambda, t)
		if math.IsNaN(transformed[i]) || math.IsInf(transformed[i], 0) {
			return math.Inf(-1)
		}
	}

	mean := 0.0
	for _, value := range transformed {
		mean += value
	}
	mean /= n

	variance := 0.0
	for _, value := range transformed {
		diff := value - mean
		variance += diff * diff
	}
	variance /= n
	if variance <= 0 {
		return math.Inf(-1)
	}

	jacobian := 0.0
	for _, value := range values {
		if t == BoxCox {
			jacobian += math.Log(value)
		} else {
			// The Yeo-Johnson Jacobian is sign(x)·ln(|x|+1), which reduces to
			// the Box-Cox form for positive x shifted by one.
			jacobian += math.Copysign(math.Log1p(math.Abs(value)), value)
		}
	}

	return -0.5*n*math.Log(variance) + (lambda-1)*jacobian
}

// estimateLambda finds the lambda maximising the profile log-likelihood.
//
// Golden-section search rather than a derivative method: the likelihood is
// smooth and unimodal in practice but its derivative is awkward to write for
// the two-sided Yeo-Johnson form, and a derivative-free search over a bounded
// interval cannot run away. The interval is the one used in practice and by
// scipy.
func estimateLambda(values []float64, t Type) float64 {
	const goldenRatio = 0.6180339887498949 // 1/φ

	lo, hi := lambdaMin, lambdaMax
	c := hi - goldenRatio*(hi-lo)
	d := lo + goldenRatio*(hi-lo)
	fc := logLikelihood(values, c, t)
	fd := logLikelihood(values, d, t)

	for hi-lo > lambdaTolerance {
		if fc > fd {
			hi, d, fd = d, c, fc
			c = hi - goldenRatio*(hi-lo)
			fc = logLikelihood(values, c, t)
		} else {
			lo, c, fc = c, d, fd
			d = lo + goldenRatio*(hi-lo)
			fd = logLikelihood(values, d, t)
		}
	}

	return (lo + hi) / 2
}
