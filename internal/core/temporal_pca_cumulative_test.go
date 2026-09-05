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
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemporalPCACumulativeVariance(t *testing.T) {
	// Create test data - simple time series
	data := types.Matrix{
		{1, 2, 3},
		{2, 3, 4},
		{3, 4, 5},
		{4, 5, 6},
		{5, 6, 7},
		{6, 7, 8},
		{7, 8, 9},
		{8, 9, 10},
	}

	// Create temporal PCA config
	config := types.PCAConfig{
		Components:    3,
		TemporalLags:  2,
		MeanCenter:    true,
		StandardScale: false,
	}

	// Create and fit temporal PCA
	tpca := NewTemporalPCAEngine()
	result, err := tpca.Fit(data, config)
	require.NoError(t, err, "Temporal PCA fit should not fail")

	// Check that we have cumulative variance
	require.NotNil(t, result.CumulativeVar, "CumulativeVar should not be nil")
	require.Equal(t, len(result.ExplainedVarRatio), len(result.CumulativeVar),
		"CumulativeVar should have same length as ExplainedVarRatio")

	// Verify cumulative variance is calculated correctly
	// It should be the cumulative sum of explained variance ratios (percentages)
	expectedCumSum := 0.0
	for i := range result.ExplainedVarRatio {
		expectedCumSum += result.ExplainedVarRatio[i]
		assert.InDelta(t, expectedCumSum, result.CumulativeVar[i], 0.001,
			"Cumulative variance at component %d should be cumulative sum of percentages", i+1)
	}

	// The cumulative variance should be monotonically increasing
	for i := 1; i < len(result.CumulativeVar); i++ {
		assert.GreaterOrEqual(t, result.CumulativeVar[i], result.CumulativeVar[i-1],
			"Cumulative variance should be monotonically increasing")
	}

	// The last cumulative value should account for all the variance.
	lastCumVar := result.CumulativeVar[len(result.CumulativeVar)-1]
	assert.InDelta(t, 1.0, lastCumVar, 0.01,
		"Last cumulative variance should be close to 1.0 (got %.4f)", lastCumVar)

	// And it must be the normalised quantity, not a running sum of raw
	// eigenvalues. This guard predates V2 and its polarity has flipped: it used
	// to assert the value was large, because cumulative variance was a
	// percentage and raw eigenvalues summed to about 1. Now cumulative variance
	// is a fraction summing to 1 and the eigenvalues are the large quantity, so
	// the same confusion shows up the other way round.
	//
	// The intent is unchanged: catch the two being swapped. Keeping it requires
	// the eigenvalue sum to be far from 1, which on this fixture it is.
	rawEigenvalueSum := 0.0
	for _, v := range result.ExplainedVar {
		rawEigenvalueSum += v
	}
	assert.Greater(t, rawEigenvalueSum, 10.0,
		"this fixture's eigenvalues should sum well above 1, or the check below "+
			"cannot tell a fraction from an eigenvalue sum")
	assert.Less(t, lastCumVar, 2.0,
		"Cumulative variance should be a fraction of 1, not a sum of raw eigenvalues")
}

func TestTemporalPCACumulativeVarianceWithStockData(t *testing.T) {
	// Simulate stock-like data with strong trend
	rows := 100
	data := make(types.Matrix, rows)
	for i := 0; i < rows; i++ {
		// Create correlated stock-like data (open, high, low, close, volume)
		base := float64(i)*0.5 + 100
		noise := math.Sin(float64(i) * 0.1)
		data[i] = []float64{
			base + noise,                 // open
			base + noise + 1.0,           // high
			base + noise - 0.5,           // low
			base + noise + 0.2,           // close
			1000000.0 + float64(i)*10000, // volume
		}
	}

	config := types.PCAConfig{
		Components:    5,
		TemporalLags:  10,
		StandardScale: true,
	}

	tpca := NewTemporalPCAEngine()
	result, err := tpca.Fit(data, config)
	require.NoError(t, err)

	// For stock data with strong trend, first component should explain majority
	assert.Greater(t, result.ExplainedVarRatio[0], 0.5,
		"First component should explain more than half the variance for trending stock data")

	// Verify cumulative variance calculation
	for i := range result.CumulativeVar {
		if i == 0 {
			assert.Equal(t, result.ExplainedVarRatio[0], result.CumulativeVar[0],
				"First cumulative variance should equal first explained variance ratio")
		} else {
			expectedCum := result.CumulativeVar[i-1] + result.ExplainedVarRatio[i]
			assert.InDelta(t, expectedCum, result.CumulativeVar[i], 0.001,
				"Cumulative variance should be correct at component %d", i+1)
		}
	}
}
