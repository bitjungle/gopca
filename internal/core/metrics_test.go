package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
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

	// Mean-centered data (original data - means)
	// Note: means are {5.5, 3.2, 2.1}
	originalData := types.Matrix{
		{1.0, 0.6, 0.9},    // {6.5-5.5, 3.8-3.2, 3.0-2.1}
		{-2.0, -0.4, -0.9}, // {3.5-5.5, 2.8-3.2, 1.2-2.1}
		{0.5, -0.7, 0.7},   // {6.0-5.5, 2.5-3.2, 2.8-2.1}
		{-0.7, -0.2, -0.3}, // {4.8-5.5, 3.0-3.2, 1.8-2.1}
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
	// Using mean-centered data as would be the case in real PCA
	scores := mat.NewDense(2, 1, []float64{
		1.0,
		-1.0,
	})

	loadings := mat.NewDense(2, 1, []float64{
		0.8,
		0.6,
	})

	// These would be the original means before centering
	mean := []float64{5.0, 3.0}
	stdDev := []float64{1.0, 1.0}

	// Mean-centered data (originalData - mean)
	centeredData := types.Matrix{
		{0.8, 0.6},   // Original was {5.8, 3.6}, centered: {5.8-5.0, 3.6-3.0}
		{-0.8, -0.6}, // Original was {4.2, 2.4}, centered: {4.2-5.0, 2.4-3.0}
	}

	calc := NewPCAMetricsCalculator(scores, loadings, mean, stdDev)

	// Calculate RSS for first sample
	rss0, err := calc.calculateRSS(0, centeredData)
	if err != nil {
		t.Fatalf("Failed to calculate RSS: %v", err)
	}

	// Calculate RSS for second sample
	rss1, err := calc.calculateRSS(1, centeredData)
	if err != nil {
		t.Fatalf("Failed to calculate RSS: %v", err)
	}

	// RSS should be non-negative
	if rss0 < 0 || rss1 < 0 {
		t.Errorf("RSS should be non-negative, got %f and %f", rss0, rss1)
	}

	// With perfect reconstruction in centered space:
	// Sample 0: reconstruction = 1.0 * [0.8, 0.6] = [0.8, 0.6]
	// Sample 1: reconstruction = -1.0 * [0.8, 0.6] = [-0.8, -0.6]
	// These match the centered data exactly, so RSS should be 0
	if math.Abs(rss0) > 1e-10 {
		t.Errorf("RSS for sample 0 should be ~0, got %f", rss0)
	}
	if math.Abs(rss1) > 1e-10 {
		t.Errorf("RSS for sample 1 should be ~0, got %f", rss1)
	}

	t.Logf("RSS values: sample 0 = %f, sample 1 = %f", rss0, rss1)
}

func TestRSSWithMeanCenteredData(t *testing.T) {
	// Test that specifically verifies the fix for issue #83
	// This simulates a scenario similar to the Iris dataset with mean centering

	// 3 samples, 4 features, 2 components
	scores := mat.NewDense(3, 2, []float64{
		2.5, 0.5,
		-1.5, 1.2,
		-1.0, -1.7,
	})

	loadings := mat.NewDense(4, 2, []float64{
		0.5, 0.3,
		0.4, -0.5,
		0.6, 0.4,
		0.3, -0.6,
	})

	// Original means (similar to Iris scale)
	mean := []float64{5.8, 3.0, 3.7, 1.2}
	stdDev := []float64{0.8, 0.4, 1.7, 0.7}

	// Mean-centered data
	centeredData := types.Matrix{
		{0.7, 0.4, 0.8, -0.2},   // Sample 1 centered
		{-1.3, -0.5, -1.2, 0.3}, // Sample 2 centered
		{0.2, 0.0, -0.7, -0.4},  // Sample 3 centered
	}

	calc := NewPCAMetricsCalculator(scores, loadings, mean, stdDev)

	// Calculate RSS for all samples
	for i := 0; i < 3; i++ {
		rss, err := calc.calculateRSS(i, centeredData)
		if err != nil {
			t.Fatalf("Failed to calculate RSS for sample %d: %v", i, err)
		}

		// With 2 components out of 4 features, there will be some reconstruction error
		// But RSS should be reasonable (not in the tens or hundreds)
		if rss > 5.0 {
			t.Errorf("RSS for sample %d is too high (%f), indicating the preprocessing space mismatch bug", i, rss)
		}

		t.Logf("Sample %d RSS: %f", i, rss)
	}
}

func TestRSSWithSNVPreprocessing(t *testing.T) {
	// Test that verifies RSS calculation works correctly with SNV preprocessing
	// This addresses the issue where SNV-only preprocessing was broken

	// 3 samples, 4 features (simulating spectral data), 2 components
	scores := mat.NewDense(3, 2, []float64{
		1.8, -0.5,
		-1.2, 0.8,
		-0.6, -0.3,
	})

	loadings := mat.NewDense(4, 2, []float64{
		0.6, 0.2,
		0.5, -0.4,
		0.4, 0.5,
		0.3, -0.3,
	})

	// No column means stored (SNV doesn't remove column means)
	mean := []float64{}
	stdDev := []float64{}

	// SNV-preprocessed data (each row normalized to mean=0, std=1)
	// This simulates spectral data after SNV
	snvData := types.Matrix{
		{1.2, -0.8, 0.3, -0.7},  // Row mean=0, std≈1
		{-0.5, 1.5, -0.6, -0.4}, // Row mean=0, std≈1
		{0.2, 0.9, -1.3, 0.2},   // Row mean=0, std≈1
	}

	calc := NewPCAMetricsCalculator(scores, loadings, mean, stdDev)

	// Calculate RSS for all samples
	for i := 0; i < 3; i++ {
		rss, err := calc.calculateRSS(i, snvData)
		if err != nil {
			t.Fatalf("Failed to calculate RSS for sample %d: %v", i, err)
		}

		// With SNV preprocessing and 2 components out of 4 features,
		// there will be reconstruction error. The key is that RSS should be reasonable,
		// not in the hundreds or thousands as it was with the bug
		if rss > 10.0 {
			t.Errorf("RSS for sample %d is too high (%f), indicating a preprocessing mismatch", i, rss)
		}

		t.Logf("Sample %d RSS with SNV: %f", i, rss)
	}
}

func TestMetricsWithReferenceValues(t *testing.T) {
	// Test case based on Python reference implementation
	// Using 2 components as in the Python example
	scores := mat.NewDense(4, 2, []float64{
		1.5, 0.8,
		-1.2, 0.5,
		0.6, -1.1,
		-0.9, -0.2,
	})

	loadings := mat.NewDense(5, 2, []float64{
		0.5, 0.3,
		0.4, -0.4,
		0.3, 0.5,
		0.6, 0.2,
		0.2, -0.6,
	})

	// No mean/stddev to match Python's centered data
	calc := NewPCAMetricsCalculator(scores, loadings, nil, nil)

	// Original preprocessed data (already centered/scaled)
	originalData := types.Matrix{
		{0.8, 0.6, 0.4, 1.0, 0.2},
		{-0.5, -0.3, -0.2, -0.8, -0.1},
		{0.3, 0.2, 0.1, 0.4, 0.0},
		{-0.6, -0.5, -0.3, -0.6, -0.1},
	}

	// Calculate metrics
	metrics, err := calc.CalculateMetrics(originalData)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Check Mahalanobis distances are in reasonable range (0-10, not millions)
	for i, m := range metrics {
		if m.Mahalanobis > 10.0 {
			t.Errorf("Mahalanobis distance for sample %d is too high: %f (expected < 10)", i, m.Mahalanobis)
		}
		if m.Mahalanobis < 0 {
			t.Errorf("Mahalanobis distance for sample %d is negative: %f", i, m.Mahalanobis)
		}
		t.Logf("Sample %d: Mahalanobis = %f", i, m.Mahalanobis)
	}

	// Check RSS values are in reasonable range (0-2, not millions)
	for i, m := range metrics {
		if m.RSS > 2.0 {
			t.Errorf("RSS for sample %d is too high: %f (expected < 2)", i, m.RSS)
		}
		if m.RSS < 0 {
			t.Errorf("RSS for sample %d is negative: %f", i, m.RSS)
		}
		t.Logf("Sample %d: RSS = %f", i, m.RSS)
	}
}

func TestHighDimensionalData(t *testing.T) {
	// Test with high-dimensional data (like NIR spectroscopy)
	// 10 samples, 3 components, 100 features
	nSamples := 10
	nComponents := 3
	nFeatures := 100

	// Create realistic scores
	scoresData := make([]float64, nSamples*nComponents)
	for i := range scoresData {
		scoresData[i] = float64(i%5-2) * 0.5 // Values between -1 and 1
	}
	scores := mat.NewDense(nSamples, nComponents, scoresData)

	// Create realistic loadings
	loadingsData := make([]float64, nFeatures*nComponents)
	for i := range loadingsData {
		loadingsData[i] = math.Sin(float64(i)) * 0.3 // Sinusoidal pattern
	}
	loadings := mat.NewDense(nFeatures, nComponents, loadingsData)

	// Create calculator with adaptive regularization
	calc := NewPCAMetricsCalculator(scores, loadings, nil, nil)

	// Create synthetic original data
	originalData := make(types.Matrix, nSamples)
	for i := 0; i < nSamples; i++ {
		originalData[i] = make([]float64, nFeatures)
		for j := 0; j < nFeatures; j++ {
			// Reconstruct approximately from scores and loadings with small noise
			val := 0.0
			for k := 0; k < nComponents; k++ {
				val += scores.At(i, k) * loadings.At(j, k)
			}
			originalData[i][j] = val + 0.01*math.Sin(float64(i*j)) // Add small noise
		}
	}

	// Calculate metrics
	metrics, err := calc.CalculateMetrics(originalData)
	if err != nil {
		t.Fatalf("Failed to calculate metrics for high-dimensional data: %v", err)
	}

	// Verify all metrics are in reasonable ranges
	for i, m := range metrics {
		if m.Mahalanobis > 100.0 || m.Mahalanobis < 0 {
			t.Errorf("Mahalanobis distance out of range for sample %d: %f", i, m.Mahalanobis)
		}
		if m.RSS > 10.0 || m.RSS < 0 {
			t.Errorf("RSS out of range for sample %d: %f", i, m.RSS)
		}
		if math.IsNaN(m.Mahalanobis) || math.IsInf(m.Mahalanobis, 0) {
			t.Errorf("Mahalanobis distance is NaN or Inf for sample %d", i)
		}
		if math.IsNaN(m.RSS) || math.IsInf(m.RSS, 0) {
			t.Errorf("RSS is NaN or Inf for sample %d", i)
		}
	}

	t.Logf("High-dimensional test passed with %d samples, %d components, %d features", nSamples, nComponents, nFeatures)
}

func TestCalculateT2Limits(t *testing.T) {
	// Test T² confidence limits calculation
	nSamples := 100
	nComponents := 3
	nFeatures := 10

	// Create sample data
	scores := mat.NewDense(nSamples, nComponents, nil)
	loadings := mat.NewDense(nFeatures, nComponents, nil)

	calc := NewPCAMetricsCalculator(scores, loadings, nil, nil)

	// Calculate T² limits
	limit95, limit99 := calc.CalculateT2Limits()

	// Basic validation
	if limit95 <= 0 {
		t.Errorf("95%% T² limit should be positive, got %f", limit95)
	}
	if limit99 <= 0 {
		t.Errorf("99%% T² limit should be positive, got %f", limit99)
	}
	if limit99 <= limit95 {
		t.Errorf("99%% limit (%f) should be greater than 95%% limit (%f)", limit99, limit95)
	}

	// Test with known values
	// For n=100, p=3, expected approximate values based on F-distribution
	expectedLimit95 := 3.0 * 99.0 / 97.0 * 2.70 // F(3,97,0.95) ≈ 2.70
	expectedLimit99 := 3.0 * 99.0 / 97.0 * 3.98 // F(3,97,0.99) ≈ 3.98

	// Allow some tolerance
	if math.Abs(limit95-expectedLimit95) > 0.5 {
		t.Logf("Warning: 95%% limit %f differs from expected %f", limit95, expectedLimit95)
	}
	if math.Abs(limit99-expectedLimit99) > 0.5 {
		t.Logf("Warning: 99%% limit %f differs from expected %f", limit99, expectedLimit99)
	}

	t.Logf("T² limits calculated: 95%%=%f, 99%%=%f", limit95, limit99)
}

func TestCalculateQLimits(t *testing.T) {
	// Test Q-residual confidence limits calculation
	nSamples := 100
	nComponents := 3
	nFeatures := 10
	totalComponents := 8 // Less than features, more than retained

	// Create sample eigenvalues
	// Typical pattern: decreasing eigenvalues
	eigenvalues := []float64{
		5.0, 3.0, 1.5, // Retained components
		0.8, 0.5, 0.3, 0.1, 0.05, // Non-retained components
	}

	scores := mat.NewDense(nSamples, nComponents, nil)
	loadings := mat.NewDense(nFeatures, nComponents, nil)

	calc := NewPCAMetricsCalculator(scores, loadings, nil, nil)

	// Calculate Q limits
	limit95, limit99 := calc.CalculateQLimits(eigenvalues, totalComponents)

	// Basic validation
	if limit95 <= 0 {
		t.Errorf("95%% Q limit should be positive, got %f", limit95)
	}
	if limit99 <= 0 {
		t.Errorf("99%% Q limit should be positive, got %f", limit99)
	}
	if limit99 <= limit95 {
		t.Errorf("99%% limit (%f) should be greater than 95%% limit (%f)", limit99, limit95)
	}

	t.Logf("Q limits calculated: 95%%=%f, 99%%=%f", limit95, limit99)

	// Test edge case: all variance in retained components
	eigenvaluesNoResidual := []float64{5.0, 3.0, 1.5, 0.0, 0.0, 0.0}
	limit95Zero, limit99Zero := calc.CalculateQLimits(eigenvaluesNoResidual, 6)

	if limit95Zero != 0 || limit99Zero != 0 {
		t.Errorf("Q limits should be 0 when all variance is retained, got 95%%=%f, 99%%=%f",
			limit95Zero, limit99Zero)
	}
}

func TestCalculateLimitsEdgeCases(t *testing.T) {
	// Test edge cases for limit calculations

	// Case 1: More components than samples
	scores1 := mat.NewDense(5, 10, nil) // 5 samples, 10 components (invalid)
	loadings1 := mat.NewDense(20, 10, nil)
	calc1 := NewPCAMetricsCalculator(scores1, loadings1, nil, nil)

	limit95, limit99 := calc1.CalculateT2Limits()
	if limit95 != 0 || limit99 != 0 {
		t.Errorf("T² limits should be 0 when n <= p, got 95%%=%f, 99%%=%f", limit95, limit99)
	}

	// Case 2: Empty eigenvalues
	calc2 := NewPCAMetricsCalculator(mat.NewDense(10, 2, nil), mat.NewDense(5, 2, nil), nil, nil)
	qLimit95, qLimit99 := calc2.CalculateQLimits([]float64{}, 0)
	if qLimit95 != 0 || qLimit99 != 0 {
		t.Errorf("Q limits should be 0 with empty eigenvalues, got 95%%=%f, 99%%=%f", qLimit95, qLimit99)
	}
}
