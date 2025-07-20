package core

import (
	"math"
	"testing"

	"github.com/bitjungle/complab/pkg/types"
)

// TestMetricsWithIrisData tests metrics calculation with a subset of iris data
func TestMetricsWithIrisData(t *testing.T) {
	// Small subset of iris data for testing
	irisData := types.Matrix{
		{5.1, 3.5, 1.4, 0.2}, // setosa
		{4.9, 3.0, 1.4, 0.2}, // setosa
		{7.0, 3.2, 4.7, 1.4}, // versicolor
		{6.4, 3.2, 4.5, 1.5}, // versicolor
		{6.3, 3.3, 6.0, 2.5}, // virginica
		{5.8, 2.7, 5.1, 1.9}, // virginica
	}

	// Run PCA
	pca := NewPCAEngine()
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}

	result, err := pca.Fit(irisData, config)
	if err != nil {
		t.Fatalf("PCA failed: %v", err)
	}
	
	t.Logf("PCA result - Scores: %d×%d, Loadings: %d×%d", 
		len(result.Scores), len(result.Scores[0]),
		len(result.Loadings), len(result.Loadings[0]))

	// Calculate metrics
	mc := NewMetricsCalculator()
	metricsConfig := types.MetricsConfig{
		NumComponents:             2,
		SignificanceLevel:        0.05,
		CalculateContributions:   true,
		CalculateConfidenceEllipse: true,
	}

	metrics, err := mc.CalculateMetrics(result, irisData, metricsConfig)
	if err != nil {
		t.Fatalf("Metrics calculation failed: %v", err)
	}

	// Verify basic properties
	if len(metrics.MahalanobisDistances) != 6 {
		t.Errorf("Expected 6 Mahalanobis distances, got %d", len(metrics.MahalanobisDistances))
	}

	if len(metrics.HotellingT2) != 6 {
		t.Errorf("Expected 6 Hotelling T² values, got %d", len(metrics.HotellingT2))
	}

	// Check that all metrics are non-negative
	for i, d := range metrics.MahalanobisDistances {
		if d < 0 || math.IsNaN(d) || math.IsInf(d, 0) {
			t.Errorf("Invalid Mahalanobis distance at index %d: %v", i, d)
		}
	}

	for i, t2 := range metrics.HotellingT2 {
		if t2 < 0 || math.IsNaN(t2) || math.IsInf(t2, 0) {
			t.Errorf("Invalid Hotelling T² at index %d: %v", i, t2)
		}
	}

	// Check residuals
	for i, rss := range metrics.RSS {
		if rss < 0 || math.IsNaN(rss) || math.IsInf(rss, 0) {
			t.Errorf("Invalid RSS at index %d: %v", i, rss)
		}
	}

	// Check contributions
	if metrics.ContributionScores == nil {
		t.Error("Expected contribution scores to be calculated")
	} else {
		t.Logf("Contributions structure: %d x %d", len(metrics.ContributionScores), 
			len(metrics.ContributionScores[0]))
		t.Logf("Result loadings: %d components x %d variables", len(result.Loadings), len(result.Loadings[0]))
		
		// The contributions should be variables × components
		// With 4 variables and 2 components, we expect 4×2
		if len(metrics.ContributionScores) != 4 {
			t.Errorf("Expected 4 rows (variables) in contributions, got %d", len(metrics.ContributionScores))
			
			// Log actual structure to debug
			for i, row := range metrics.ContributionScores {
				t.Logf("  Row %d has %d values", i, len(row))
			}
		}
	}

	// Check confidence ellipse
	if metrics.ConfidenceEllipse.ConfidenceLevel != 0.95 {
		t.Errorf("Expected confidence level 0.95, got %v", metrics.ConfidenceEllipse.ConfidenceLevel)
	}

	if metrics.ConfidenceEllipse.MajorAxis <= 0 || metrics.ConfidenceEllipse.MinorAxis <= 0 {
		t.Error("Confidence ellipse axes should be positive")
	}
}

// TestMetricsEdgeCases tests edge cases and error conditions
func TestMetricsEdgeCases(t *testing.T) {
	mc := NewMetricsCalculator()

	t.Run("nil result", func(t *testing.T) {
		_, err := mc.CalculateMetrics(nil, types.Matrix{{1, 2}}, types.MetricsConfig{})
		if err == nil {
			t.Error("Expected error for nil result")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		result := &types.PCAResult{
			Scores:   [][]float64{{1}},
			Loadings: [][]float64{{1}},
		}
		_, err := mc.CalculateMetrics(result, types.Matrix{}, types.MetricsConfig{})
		if err == nil {
			t.Error("Expected error for empty data")
		}
	})

	t.Run("single observation", func(t *testing.T) {
		data := types.Matrix{{1.0, 2.0}}
		pca := NewPCAEngine()
		result, err := pca.Fit(data, types.PCAConfig{Components: 1, MeanCenter: true})
		if err == nil {
			// Single observation should work but produce specific results
			metrics, err := mc.CalculateMetrics(result, data, types.MetricsConfig{})
			if err != nil {
				t.Errorf("Unexpected error for single observation: %v", err)
			}
			if metrics != nil && len(metrics.MahalanobisDistances) != 1 {
				t.Error("Expected one distance for single observation")
			}
		}
	})
}

// TestOutlierDetectionThreshold tests the outlier detection threshold calculation
func TestOutlierDetectionThreshold(t *testing.T) {
	// Test data with known outliers
	// Create data where some points are far from the center
	data := types.Matrix{
		{0.0, 0.0},
		{0.1, 0.1},
		{-0.1, 0.1},
		{0.1, -0.1},
		{5.0, 5.0}, // Clear outlier
		{-5.0, 5.0}, // Clear outlier
	}

	pca := NewPCAEngine()
	result, err := pca.Fit(data, types.PCAConfig{
		Components: 2,
		MeanCenter: true,
	})
	if err != nil {
		t.Fatalf("PCA failed: %v", err)
	}

	mc := NewMetricsCalculator()
	metrics, err := mc.CalculateMetrics(result, data, types.MetricsConfig{
		NumComponents:     2,
		SignificanceLevel: 0.1, // Use higher significance to catch outliers
	})
	if err != nil {
		t.Fatalf("Metrics calculation failed: %v", err)
	}

	// Count outliers
	outlierCount := 0
	for _, isOutlier := range metrics.OutlierMask {
		if isOutlier {
			outlierCount++
		}
	}

	// We expect at least the two clear outliers to be detected
	if outlierCount < 2 {
		t.Errorf("Expected at least 2 outliers, detected %d", outlierCount)
		t.Logf("T² values: %v", metrics.HotellingT2)
		t.Logf("Outlier mask: %v", metrics.OutlierMask)
	}

	// The outliers should have the highest T² values
	maxT2Idx := 0
	maxT2 := metrics.HotellingT2[0]
	for i, t2 := range metrics.HotellingT2 {
		if t2 > maxT2 {
			maxT2 = t2
			maxT2Idx = i
		}
	}

	if maxT2Idx != 4 && maxT2Idx != 5 {
		t.Errorf("Expected outlier indices 4 or 5 to have max T², got index %d", maxT2Idx)
	}
}

// BenchmarkMetricsCalculation benchmarks metrics calculation performance
func BenchmarkMetricsCalculation(b *testing.B) {
	// Create a larger dataset
	n := 1000
	p := 10
	data := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		data[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			data[i][j] = float64(i+j) / float64(n) // Simple pattern
		}
	}

	// Run PCA once
	pca := NewPCAEngine()
	result, err := pca.Fit(data, types.PCAConfig{
		Components: 5,
		MeanCenter: true,
	})
	if err != nil {
		b.Fatalf("PCA failed: %v", err)
	}

	mc := NewMetricsCalculator()
	config := types.MetricsConfig{
		NumComponents:     5,
		SignificanceLevel: 0.01,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mc.CalculateMetrics(result, data, config)
		if err != nil {
			b.Fatalf("Metrics calculation failed: %v", err)
		}
	}
}