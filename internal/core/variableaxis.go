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
	"sort"
	"strconv"

	"github.com/bitjungle/gopca/pkg/types"
)

// Deciding whether the variables form an axis a derivative can be taken along.
//
// Savitzky-Golay slides a window across consecutive variables and fits a
// polynomial to them. That is only meaningful if two things hold, and they are
// separate questions answered by separate evidence:
//
//  1. The ordering carries continuity -- adjacent variables measure nearly the
//     same thing, so a local polynomial has something to fit. This is a property
//     of the data.
//  2. The variables are evenly spaced -- the filter's coefficients assume it.
//     This is a property of the axis, and no amount of data smoothness can
//     establish it.
//
// Neither test subsumes the other, and each is blind to what the other catches.
// Excluding a band from the middle of a spectrum leaves the data as smooth as it
// was while making non-adjacent wavelengths neighbours; renaming columns from
// wavelengths to labels destroys the spacing evidence while the data is
// untouched. Both are checked, and reported separately.

// ContinuityThreshold is the largest von Neumann ratio still treated as a
// continuum.
//
// Measured across every dataset in testdata/: the spectral ones (all corn
// variants, bronir2) fall between 0.00032 and 0.0068, and nothing else comes
// below 1.13 -- a gap of 166 between the two groups. Corn with Gaussian noise at
// 20% of its overall standard deviation still reaches only 0.08, so this leaves
// real but noisy spectra a wide margin. The nearest non-spectral dataset sits
// more than four times above it.
const ContinuityThreshold = 0.5

// minVariablesForFilter is the smallest number of variables that can carry a
// filter at all: the window must be at least 3 and cannot exceed the data.
const minVariablesForFilter = 3

// AxisReport describes the variable axis of a dataset.
type AxisReport struct {
	// Variables is the number of variables examined.
	Variables int

	// Continuity is the von Neumann ratio: the mean square successive difference
	// along each row divided by that row's variance, taken as a median across
	// rows. Under a random ordering its expectation is 2, so it is a statistic
	// with a reference value rather than an arbitrary scale. A smooth curve
	// drives it toward zero.
	//
	// Reference: von Neumann, J. (1941). Distribution of the ratio of the mean
	// square successive difference to the variance. Annals of Mathematical
	// Statistics 12(4), 367-395.
	Continuity float64

	// SmoothnessFactor restates Continuity as "adjacent variables differ this
	// many times less than a random ordering would", which is the form worth
	// showing a user: 6093 for corn, 1.8 for wine.
	SmoothnessFactor float64

	// IsContinuous reports whether a derivative along this axis is meaningful.
	IsContinuous bool

	// NamesNumeric reports whether the column names parse as numbers, which is
	// what allows the spacing to be checked at all.
	NamesNumeric bool

	// SpacingUniform reports whether those numbers advance in a single constant
	// step. False when names are not numeric, when they are not monotonic, or
	// when a gap has been opened by excluding variables.
	SpacingUniform bool

	// DistinctSteps counts the different step sizes found between neighbouring
	// variables; 1 is uniform. Zero when the names are not numeric.
	DistinctSteps int
}

// AnalyzeVariableAxis examines the matrix that will actually be analysed.
//
// It must be given the data after any row and column exclusions, not the file as
// loaded: excluding an interior band is precisely the case the spacing check
// exists to catch, and a report computed before the exclusion describes a matrix
// that no longer exists.
//
// headers may be nil or the wrong length, in which case only the continuity part
// of the report is filled in.
//
// Complexity: O(rows * variables).
func AnalyzeVariableAxis(data types.Matrix, headers []string) AxisReport {
	report := AxisReport{}
	if len(data) == 0 {
		return report
	}
	report.Variables = len(data[0])

	report.Continuity, report.SmoothnessFactor = continuityOf(data)
	report.IsContinuous = report.Variables >= minVariablesForFilter &&
		report.Continuity > 0 &&
		report.Continuity <= ContinuityThreshold

	if len(headers) == report.Variables {
		report.NamesNumeric, report.SpacingUniform, report.DistinctSteps = spacingOf(headers)
	}
	return report
}

// continuityOf returns the median von Neumann ratio across rows, and the same
// figure expressed as a factor relative to the random-ordering expectation of 2.
func continuityOf(data types.Matrix) (ratio, factor float64) {
	ratios := make([]float64, 0, len(data))
	for _, row := range data {
		if len(row) < 2 {
			continue
		}
		// A row with a missing value has no meaningful successive differences
		// around it, and a flat row has no variance to divide by. Both are
		// skipped rather than defaulting to a value that would sway the median.
		var sumSq, sum, count float64
		finite := true
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				finite = false
				break
			}
			sum += v
			count++
		}
		if !finite || count < 2 {
			continue
		}
		mean := sum / count
		for _, v := range row {
			sumSq += (v - mean) * (v - mean)
		}
		variance := sumSq / (count - 1)
		if variance <= 0 {
			continue
		}

		var mssd float64
		for j := 1; j < len(row); j++ {
			d := row[j] - row[j-1]
			mssd += d * d
		}
		mssd /= float64(len(row) - 1)

		ratios = append(ratios, mssd/variance)
	}
	if len(ratios) == 0 {
		return 0, 0
	}

	// Median rather than mean: a handful of corrupted or unusually flat rows
	// should not decide whether the axis is a continuum.
	sort.Float64s(ratios)
	mid := len(ratios) / 2
	if len(ratios)%2 == 1 {
		ratio = ratios[mid]
	} else {
		ratio = (ratios[mid-1] + ratios[mid]) / 2
	}
	if ratio > 0 {
		factor = 2 / ratio
	}
	return ratio, factor
}

// spacingOf inspects the column names for an evenly spaced numeric axis.
func spacingOf(headers []string) (numeric, uniform bool, distinctSteps int) {
	values := make([]float64, 0, len(headers))
	for _, h := range headers {
		v, err := strconv.ParseFloat(h, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		values = append(values, v)
	}
	// Spectral files routinely carry a few non-wavelength columns alongside the
	// spectrum, so require most rather than all of the names to be numbers.
	if len(headers) == 0 || float64(len(values))/float64(len(headers)) < 0.9 || len(values) < 2 {
		return false, false, 0
	}

	steps := make([]float64, 0, len(values)-1)
	for i := 1; i < len(values); i++ {
		steps = append(steps, values[i]-values[i-1])
	}

	// Monotonic in one direction, or the ordering is not an axis at all.
	first := steps[0]
	for _, s := range steps {
		if s == 0 || (s > 0) != (first > 0) {
			return true, false, len(distinctValues(steps))
		}
	}

	distinct := distinctValues(steps)
	return true, len(distinct) == 1, len(distinct)
}

// distinctValues groups steps that agree to within a relative tolerance, so that
// floating-point noise in parsed wavelengths does not read as a gap.
func distinctValues(steps []float64) []float64 {
	sorted := append([]float64(nil), steps...)
	sort.Float64s(sorted)

	var distinct []float64
	for _, s := range sorted {
		matched := false
		for _, d := range distinct {
			scale := math.Max(math.Abs(d), math.Abs(s))
			if scale == 0 || math.Abs(d-s)/scale < 1e-6 {
				matched = true
				break
			}
		}
		if !matched {
			distinct = append(distinct, s)
		}
	}
	return distinct
}
