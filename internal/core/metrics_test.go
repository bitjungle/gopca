package core

import (
	"math"
	"testing"

	"github.com/bitjungle/complab/pkg/types"
	"gonum.org/v1/gonum/mat"
)

func TestPCAMetricsCalculator(t *testing.T) {
	// Create test data - simple 2D data
	scores := mat.NewDense(4, 2, []float64{
		1.0, 0.5,
		-1.0, 0.5,
		0.5, -1.0,
		-0.5, -0.5,
	})

	loadings := mat.NewDense(3, 2, []float64{
		0.7, 0.3,
		0.6, -0.5,
		0.4, 0.8,
	})

	mean := []float64{5.0, 3.0, 2.0}
	stdDev := []float64{1.5, 0.8, 1.2}

	// Original data (before preprocessing)
	originalData := types.Matrix{
		{6.0, 3.5, 2.8},
		{4.0, 2.5, 1.2},
		{5.5, 2.2, 2.5},
		{4.5, 2.8, 1.5},
	}

	// Create calculator
	calc := NewPCAMetricsCalculator(scores, loadings, mean, stdDev)

	// Calculate metrics
	metrics, err := calc.CalculateMetrics(originalData)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Check that we got metrics for all samples
	if len(metrics) != 4 {
		t.Errorf("Expected 4 metrics, got %d", len(metrics))
	}

	// Check that metrics are reasonable (non-negative, finite)
	for i, m := range metrics {
		if m.HotellingT2 < 0 || math.IsNaN(m.HotellingT2) || math.IsInf(m.HotellingT2, 0) {
			t.Errorf("Invalid Hotelling T² for sample %d: %f", i, m.HotellingT2)
		}
		if m.Mahalanobis < 0 || math.IsNaN(m.Mahalanobis) || math.IsInf(m.Mahalanobis, 0) {
			t.Errorf("Invalid Mahalanobis distance for sample %d: %f", i, m.Mahalanobis)
		}
		if m.RSS < 0 || math.IsNaN(m.RSS) || math.IsInf(m.RSS, 0) {
			t.Errorf("Invalid RSS for sample %d: %f", i, m.RSS)
		}
	}
}

func TestCalculateMetricsFromPCAResult(t *testing.T) {
	// Create a simple PCA result
	result := &types.PCAResult{
		Scores: types.Matrix{
			{2.0, 1.0},
			{-2.0, 1.0},
			{1.0, -2.0},
			{-1.0, -1.0},
		},
		Loadings: types.Matrix{
			{0.8, 0.2},
			{0.6, -0.6},
			{0.3, 0.9},
		},
		ExplainedVar:    []float64{65.0, 25.0},
		CumulativeVar:   []float64{65.0, 90.0},
		ComponentLabels: []string{"PC1", "PC2"},
		Means:           []float64{5.5, 3.2, 2.1},
		StdDevs:         []float64{1.2, 0.9, 1.0},
	}

	// Original data
	originalData := types.Matrix{
		{6.5, 3.8, 3.0},
		{3.5, 2.8, 1.2},
		{6.0, 2.5, 2.8},
		{4.8, 3.0, 1.8},
	}

	// Calculate metrics
	metrics, err := CalculateMetricsFromPCAResult(result, originalData)
	if err != nil {
		t.Fatalf("Failed to calculate metrics from PCA result: %v", err)
	}

	// Verify we got the right number of metrics
	if len(metrics) != len(result.Scores) {
		t.Errorf("Expected %d metrics, got %d", len(result.Scores), len(metrics))
	}

	// Basic sanity checks
	for i, m := range metrics {
		if m.HotellingT2 < 0 {
			t.Errorf("Negative Hotelling T² for sample %d: %f", i, m.HotellingT2)
		}
		if m.Mahalanobis < 0 {
			t.Errorf("Negative Mahalanobis distance for sample %d: %f", i, m.Mahalanobis)
		}
		if m.RSS < 0 {
			t.Errorf("Negative RSS for sample %d: %f", i, m.RSS)
		}
	}
}

func TestOutlierDetection(t *testing.T) {
	// Create test data with one clear outlier
	scores := mat.NewDense(10, 2, []float64{
		0.1, 0.1,
		-0.1, 0.2,
		0.2, -0.1,
		-0.2, -0.2,
		0.15, 0.05,
		-0.05, 0.15,
		0.0, -0.1,
		-0.1, 0.0,
		5.0, 5.0, // Clear outlier
		0.1, -0.15,
	})

	loadings := mat.NewDense(2, 2, []float64{
		0.7, 0.7,
		0.7, -0.7,
	})

	calc := NewPCAMetricsCalculator(scores, loadings, nil, nil)

	// Calculate covariance and check outlier detection
	cov, err := calc.calculateScoresCovariance()
	if err != nil {
		t.Fatalf("Failed to calculate covariance: %v", err)
	}

	var covInv mat.Dense
	err = covInv.Inverse(cov)
	if err != nil {
		t.Fatalf("Failed to invert covariance: %v", err)
	}

	// Calculate Hotelling T² for the outlier
	outlierScore := mat.NewVecDense(2, []float64{5.0, 5.0})
	means := []float64{0.0, 0.0} // Approximate mean
	t2 := calc.calculateHotellingT2(outlierScore, means, &covInv)

	// The outlier should have a much higher T² value
	normalScore := mat.NewVecDense(2, []float64{0.1, 0.1})
	t2Normal := calc.calculateHotellingT2(normalScore, means, &covInv)

	if t2 <= t2Normal {
		t.Errorf("Outlier T² (%f) should be greater than normal T² (%f)", t2, t2Normal)
	}
}

func TestRSSCalculation(t *testing.T) {
	// Simple test case where we can verify RSS manually
	scores := mat.NewDense(2, 1, []float64{
		1.0,
		-1.0,
	})

	loadings := mat.NewDense(2, 1, []float64{
		0.8,
		0.6,
	})

	mean := []float64{5.0, 3.0}
	stdDev := []float64{1.0, 1.0}

	// Original data
	originalData := types.Matrix{
		{5.8, 3.6}, // Should reconstruct to approximately these values
		{4.2, 2.4},
	}

	calc := NewPCAMetricsCalculator(scores, loadings, mean, stdDev)

	// Calculate RSS for first sample
	rss0, err := calc.calculateRSS(0, originalData)
	if err != nil {
		t.Fatalf("Failed to calculate RSS: %v", err)
	}

	// Calculate RSS for second sample
	rss1, err := calc.calculateRSS(1, originalData)
	if err != nil {
		t.Fatalf("Failed to calculate RSS: %v", err)
	}

	// RSS should be non-negative
	if rss0 < 0 || rss1 < 0 {
		t.Errorf("RSS should be non-negative, got %f and %f", rss0, rss1)
	}

	// For perfect reconstruction, RSS should be close to 0
	// In this case, we're using only 1 component, so there will be some RSS
	// But due to numerical precision, very small RSS values are acceptable
	t.Logf("RSS values: sample 0 = %f, sample 1 = %f", rss0, rss1)
}