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

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// NIPALS (Nonlinear Iterative Partial Least Squares) Algorithm Helpers
//
// This file contains helper functions for the NIPALS algorithm implementations.
// The NIPALS algorithm is an iterative method for computing principal components
// that can handle missing data and large datasets efficiently.
//
// Reference: Wold, H. (1966). Estimation of principal components and related models
// by iterative least squares. In Multivariate Analysis (P.R. Krishnaiah, ed.),
// Academic Press, New York, pp. 391-420.

// findMaxVarianceColumn finds the column with maximum variance in the matrix.
// This is used to initialize the score vector for each component.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - maxVarCol: index of column with maximum variance
//   - maxVar: the maximum variance value
func findMaxVarianceColumn(X *mat.Dense) (maxVarCol int, maxVar float64) {
	_, m := X.Dims()
	n := X.RawMatrix().Rows

	for j := 0; j < m; j++ {
		col := mat.Col(nil, j, X)
		var sum, sumSq float64
		for _, v := range col {
			sum += v
			sumSq += v * v
		}
		variance := sumSq/float64(n) - (sum/float64(n))*(sum/float64(n))
		if variance > maxVar {
			maxVar = variance
			maxVarCol = j
		}
	}
	return maxVarCol, maxVar
}

// findMaxVarianceColumnWithMissing finds the column with maximum variance,
// handling missing (NaN) values by excluding them from variance calculation.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - maxVarCol: index of column with maximum variance
//   - maxVar: the maximum variance value
func findMaxVarianceColumnWithMissing(X *mat.Dense) (maxVarCol int, maxVar float64) {
	n, m := X.Dims()

	for j := 0; j < m; j++ {
		var sum, sumSq float64
		count := 0
		for i := 0; i < n; i++ {
			v := X.At(i, j)
			if !math.IsNaN(v) {
				sum += v
				sumSq += v * v
				count++
			}
		}
		if count > 0 {
			mean := sum / float64(count)
			variance := sumSq/float64(count) - mean*mean
			if variance > maxVar {
				maxVar = variance
				maxVarCol = j
			}
		}
	}
	return maxVarCol, maxVar
}

// computeColumnMeansWithMissing calculates column means, excluding NaN values.
// Used for mean-centering data when missing values are present.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func computeColumnMeansWithMissing(X *mat.Dense) []float64 {
	n, m := X.Dims()
	columnMeans := make([]float64, m)

	for j := 0; j < m; j++ {
		sum := 0.0
		count := 0
		for i := 0; i < n; i++ {
			val := X.At(i, j)
			if !math.IsNaN(val) {
				sum += val
				count++
			}
		}
		if count > 0 {
			columnMeans[j] = sum / float64(count)
		}
	}
	return columnMeans
}

// centerMatrixWithMissing centers the matrix by subtracting column means,
// only modifying non-missing (non-NaN) values.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func centerMatrixWithMissing(X *mat.Dense, means []float64) {
	n, m := X.Dims()
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			val := X.At(i, j)
			if !math.IsNaN(val) {
				X.Set(i, j, val-means[j])
			}
		}
	}
}

// initializeScoreVector initializes the score vector with the column
// having maximum variance. This provides a good starting point for iteration.
//
// Algorithm complexity: O(n) where n is the number of rows
func initializeScoreVector(X *mat.Dense, maxVarCol int, n int) *mat.VecDense {
	t := mat.NewVecDense(n, nil)
	col := mat.Col(nil, maxVarCol, X)
	for i := 0; i < n; i++ {
		t.SetVec(i, col[i])
	}
	return t
}

// initializeScoreVectorWithMissing initializes the score vector with the column
// having maximum variance, handling missing values by replacing them with
// the column mean.
//
// Algorithm complexity: O(n) where n is the number of rows
func initializeScoreVectorWithMissing(X *mat.Dense, maxVarCol int, n int) *mat.VecDense {
	t := mat.NewVecDense(n, nil)

	// First pass: calculate column mean for non-missing values
	var colSum float64
	colCount := 0
	for i := 0; i < n; i++ {
		v := X.At(i, maxVarCol)
		if !math.IsNaN(v) {
			colSum += v
			colCount++
		}
	}

	// Second pass: initialize t, using mean for missing values
	colMean := 0.0
	if colCount > 0 {
		colMean = colSum / float64(colCount)
	}

	for i := 0; i < n; i++ {
		v := X.At(i, maxVarCol)
		if !math.IsNaN(v) {
			t.SetVec(i, v)
		} else {
			t.SetVec(i, colMean)
		}
	}
	return t
}

// computeLoadingVector computes the loading vector p = X^T * t / (t^T * t)
// and normalizes it to unit length.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - p: normalized loading vector
//   - error: if vector has zero variance
func computeLoadingVector(X *mat.Dense, t *mat.VecDense, tolerance float64) (*mat.VecDense, error) {
	m := X.RawMatrix().Cols

	// p = X^T * t / (t^T * t)
	p := mat.NewVecDense(m, nil)
	p.MulVec(X.T(), t)
	tNorm := mat.Dot(t, t)
	if tNorm < tolerance {
		return nil, fmt.Errorf("score vector has zero variance")
	}
	p.ScaleVec(1.0/tNorm, p)

	// Normalize p to unit length
	pNorm := math.Sqrt(mat.Dot(p, p))
	if pNorm < tolerance {
		return nil, fmt.Errorf("loading vector has zero variance")
	}
	p.ScaleVec(1.0/pNorm, p)

	return p, nil
}

// computeLoadingVectorWithMissing computes the loading vector with missing value handling.
// For each loading p[j], it computes: p[j] = sum(X[i,j] * t[i]) / sum(t[i]^2)
// where the sum is only over non-missing values of X[i,j].
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - p: normalized loading vector
//   - error: if vector has zero variance
func computeLoadingVectorWithMissing(X *mat.Dense, t *mat.VecDense, tolerance float64) (*mat.VecDense, error) {
	n, m := X.Dims()
	p := mat.NewVecDense(m, nil)

	for j := 0; j < m; j++ {
		numerator := 0.0
		denominator := 0.0
		count := 0
		for i := 0; i < n; i++ {
			xVal := X.At(i, j)
			tVal := t.AtVec(i)
			if !math.IsNaN(xVal) && !math.IsNaN(tVal) {
				numerator += xVal * tVal
				denominator += tVal * tVal
				count++
			}
		}
		if count > 0 && denominator > tolerance {
			p.SetVec(j, numerator/denominator)
		} else {
			p.SetVec(j, 0)
		}
	}

	// Normalize p to unit length
	pNorm := 0.0
	for j := 0; j < m; j++ {
		pVal := p.AtVec(j)
		if !math.IsNaN(pVal) {
			pNorm += pVal * pVal
		}
	}
	pNorm = math.Sqrt(pNorm)
	if pNorm < tolerance {
		return nil, fmt.Errorf("loading vector has zero variance")
	}
	p.ScaleVec(1.0/pNorm, p)

	return p, nil
}

// updateScoreVector computes the updated score vector t = X * p / (p^T * p).
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - t: updated score vector
//   - error: if vector has zero variance
func updateScoreVector(X *mat.Dense, p *mat.VecDense, tolerance float64) (*mat.VecDense, error) {
	n := X.RawMatrix().Rows
	t := mat.NewVecDense(n, nil)

	// t = X * p / (p^T * p)
	t.MulVec(X, p)
	pNormSq := mat.Dot(p, p)
	if pNormSq < tolerance {
		return nil, fmt.Errorf("loading vector has zero norm")
	}
	t.ScaleVec(1.0/pNormSq, t)

	return t, nil
}

// updateScoreVectorWithMissing computes the updated score vector with missing value handling.
// For each score t[i], it computes: t[i] = sum(X[i,j] * p[j]) / sum(p[j]^2)
// where the sum is only over non-missing values of X[i,j].
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
//
// Returns:
//   - t: updated score vector (keeps old values for rows with all missing data)
func updateScoreVectorWithMissing(X *mat.Dense, p *mat.VecDense, tOld *mat.VecDense, tolerance float64) *mat.VecDense {
	n, m := X.Dims()
	t := mat.NewVecDense(n, nil)

	for i := 0; i < n; i++ {
		numerator := 0.0
		denominator := 0.0
		count := 0
		for j := 0; j < m; j++ {
			xVal := X.At(i, j)
			pVal := p.AtVec(j)
			if !math.IsNaN(xVal) && !math.IsNaN(pVal) {
				numerator += xVal * pVal
				denominator += pVal * pVal
				count++
			}
		}
		if count > 0 && denominator > tolerance {
			t.SetVec(i, numerator/denominator)
		} else {
			// If no valid data for this sample, keep previous value
			t.SetVec(i, tOld.AtVec(i))
		}
	}

	return t
}

// checkConvergence checks if the NIPALS iteration has converged by comparing
// the L2 norm of the difference between successive score vectors.
//
// Algorithm complexity: O(n) where n is the number of rows
//
// Returns:
//   - true if converged (difference < tolerance), false otherwise
func checkConvergence(t, tOld *mat.VecDense, tolerance float64) bool {
	diff := mat.NewVecDense(t.Len(), nil)
	diff.SubVec(t, tOld)
	return mat.Norm(diff, 2) < tolerance
}

// deflateMatrix performs matrix deflation: X = X - t * p^T
// This removes the variance explained by the current component.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func deflateMatrix(X *mat.Dense, t *mat.VecDense, p *mat.VecDense) {
	n, m := X.Dims()
	tData := make([]float64, n)
	pData := make([]float64, m)

	for i := 0; i < n; i++ {
		tData[i] = t.AtVec(i)
	}
	for j := 0; j < m; j++ {
		pData[j] = p.AtVec(j)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			X.Set(i, j, X.At(i, j)-tData[i]*pData[j])
		}
	}
}

// deflateMatrixWithMissing performs matrix deflation only on non-missing values.
// This ensures that NaN values remain NaN after deflation.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func deflateMatrixWithMissing(X *mat.Dense, tData, pData []float64) {
	n, m := X.Dims()
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if !math.IsNaN(X.At(i, j)) {
				X.Set(i, j, X.At(i, j)-tData[i]*pData[j])
			}
		}
	}
}

// extractVectorData extracts data from a gonum vector into a float64 slice.
// This is a utility function for storing computed components.
//
// Algorithm complexity: O(n) where n is the vector length
func extractVectorData(v *mat.VecDense) []float64 {
	n := v.Len()
	data := make([]float64, n)
	for i := 0; i < n; i++ {
		data[i] = v.AtVec(i)
	}
	return data
}

// computeColumnStdDevsWithMissing calculates per-column sample standard
// deviations, excluding NaN entries. Columns with fewer than two observed
// values, or with variance below MinVarianceThreshold, get a divisor of 1 so
// that scaling leaves them unchanged rather than dividing by ~0 — the same
// guard the complete-data Preprocessor applies.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func computeColumnStdDevsWithMissing(X *mat.Dense, means []float64) []float64 {
	n, m := X.Dims()
	stdDevs := make([]float64, m)

	for j := 0; j < m; j++ {
		sumSq := 0.0
		count := 0
		for i := 0; i < n; i++ {
			val := X.At(i, j)
			if !math.IsNaN(val) {
				d := val - means[j]
				sumSq += d * d
				count++
			}
		}
		stdDevs[j] = 1.0
		if count > 1 {
			if sd := math.Sqrt(sumSq / float64(count-1)); sd >= MinVarianceThreshold {
				stdDevs[j] = sd
			}
		}
	}
	return stdDevs
}

// computeColumnMediansWithMissing calculates per-column medians, excluding NaN
// entries. Columns with no observed value get a median of 0.
//
// Algorithm complexity: O(n*m*log n) where n is rows, m is columns
func computeColumnMediansWithMissing(X *mat.Dense) []float64 {
	n, m := X.Dims()
	medians := make([]float64, m)
	buf := make([]float64, 0, n)

	for j := 0; j < m; j++ {
		buf = buf[:0]
		for i := 0; i < n; i++ {
			if val := X.At(i, j); !math.IsNaN(val) {
				buf = append(buf, val)
			}
		}
		if len(buf) == 0 {
			continue
		}
		sort.Float64s(buf)
		medians[j] = stat.Quantile(0.5, stat.Empirical, buf, nil)
	}
	return medians
}

// computeColumnMADsWithMissing calculates per-column median absolute deviations,
// excluding NaN entries and scaled by 1.4826 so that MAD ≈ σ for normally
// distributed data — matching medianAbsoluteDeviation on complete data.
// Degenerate columns get a divisor of 1.
//
// Algorithm complexity: O(n*m*log n) where n is rows, m is columns
func computeColumnMADsWithMissing(X *mat.Dense, medians []float64) []float64 {
	n, m := X.Dims()
	mads := make([]float64, m)
	buf := make([]float64, 0, n)

	for j := 0; j < m; j++ {
		buf = buf[:0]
		for i := 0; i < n; i++ {
			if val := X.At(i, j); !math.IsNaN(val) {
				buf = append(buf, math.Abs(val-medians[j]))
			}
		}
		mads[j] = 1.0
		if len(buf) == 0 {
			continue
		}
		sort.Float64s(buf)
		if mad := stat.Quantile(0.5, stat.Empirical, buf, nil) * 1.4826; mad >= MinVarianceThreshold {
			mads[j] = mad
		}
	}
	return mads
}

// scaleMatrixWithMissing divides each column by its divisor, leaving missing
// entries missing.
//
// Algorithm complexity: O(n*m) where n is rows, m is columns
func scaleMatrixWithMissing(X *mat.Dense, divisors []float64) {
	n, m := X.Dims()
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if val := X.At(i, j); !math.IsNaN(val) {
				X.Set(i, j, val/divisors[j])
			}
		}
	}
}
