// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

import "math"

// countUnique returns the number of distinct values in a sorted or unsorted
// float64 slice.
func countUnique(values []float64) int {
	unique := make(map[float64]bool, len(values))
	for _, v := range values {
		unique[v] = true
	}
	return len(unique)
}

// calculateMean returns the arithmetic mean of values.
// Callers must ensure len(values) > 0.
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateMedian returns the median of a pre-sorted float64 slice.
// Callers must sort values before calling and ensure len(values) > 0.
func calculateMedian(values []float64) float64 {
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2
	}
	return values[n/2]
}

// calculateStdDev returns the population standard deviation of values
// given a pre-computed mean.
// Callers must ensure len(values) > 0.
func calculateStdDev(values []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)))
}

// calculatePercentile returns the value at the given percentile (0–100) in a
// pre-sorted slice using linear interpolation.
// Callers must sort values before calling and ensure len(values) > 0.
func calculatePercentile(values []float64, percentile float64) float64 {
	index := (percentile / 100) * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	weight := index - float64(lower)

	if lower == upper {
		return values[lower]
	}
	return values[lower]*(1-weight) + values[upper]*weight
}

// calculateSkewness returns the sample skewness (adjusted Fisher–Pearson
// coefficient) given pre-computed mean and standard deviation.
// Returns 0 when stdDev == 0 or len(values) < 3.
func calculateSkewness(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 3 {
		return 0
	}

	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		z := (v - mean) / stdDev
		sum += z * z * z
	}
	return (n / ((n - 1) * (n - 2))) * sum
}

// calculateKurtosis returns the excess kurtosis (Fisher definition, normal = 0)
// given pre-computed mean and standard deviation.
// Returns 0 when stdDev == 0 or len(values) < 4.
func calculateKurtosis(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 4 {
		return 0
	}

	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		z := (v - mean) / stdDev
		sum += z * z * z * z
	}
	return (n*(n+1)/((n-1)*(n-2)*(n-3)))*sum - 3*(n-1)*(n-1)/((n-2)*(n-3))
}
