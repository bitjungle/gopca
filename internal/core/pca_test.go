package core

import (
	"math"
	"testing"
	"time"

	"github.com/bitjungle/complab/pkg/types"
)

// Helper function to create test data
func createTestMatrix() types.Matrix {
	// Simple test data with known properties
	return types.Matrix{
		{2.5, 2.4},
		{0.5, 0.7},
		{2.2, 2.9},
		{1.9, 2.2},
		{3.1, 3.0},
		{2.3, 2.7},
		{2.0, 1.6},
		{1.0, 1.1},
		{1.5, 1.6},
		{1.1, 0.9},
	}
}

// Test basic PCA functionality
func TestPCABasic(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	// Check dimensions
	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d score rows, got %d", len(data), len(result.Scores))
	}
	
	if len(result.Scores[0]) != config.Components {
		t.Errorf("Expected %d score columns, got %d", config.Components, len(result.Scores[0]))
	}
	
	if len(result.Loadings) != len(data[0]) {
		t.Errorf("Expected %d loading rows, got %d", len(data[0]), len(result.Loadings))
	}
	
	if len(result.Loadings[0]) != config.Components {
		t.Errorf("Expected %d loading columns, got %d", config.Components, len(result.Loadings[0]))
	}
}

// Test explained variance adds up to ~100%
func TestExplainedVariance(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	// Total explained variance should be close to 100%
	totalVar := result.CumulativeVar[len(result.CumulativeVar)-1]
	if math.Abs(totalVar-100.0) > 5.0 { // Allow 5% tolerance
		t.Errorf("Total explained variance should be ~100%%, got %.2f%%", totalVar)
	}
	
	// Cumulative variance should be monotonically increasing
	for i := 1; i < len(result.CumulativeVar); i++ {
		if result.CumulativeVar[i] < result.CumulativeVar[i-1] {
			t.Errorf("Cumulative variance not monotonically increasing at index %d", i)
		}
	}
}

// Test orthogonality of loadings
func TestLoadingsOrthogonality(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	// Check that loading vectors are orthogonal
	// Compute dot product of loading columns
	dotProduct := 0.0
	for i := 0; i < len(result.Loadings); i++ {
		dotProduct += result.Loadings[i][0] * result.Loadings[i][1]
	}
	
	if math.Abs(dotProduct) > 1e-6 {
		t.Errorf("Loading vectors not orthogonal, dot product = %f", dotProduct)
	}
}

// Test with standardization
func TestPCAWithStandardization(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: true,
		Method:        "nipals",
	}
	
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	// Scores should have zero mean (approximately)
	for j := 0; j < config.Components; j++ {
		mean := 0.0
		for i := 0; i < len(result.Scores); i++ {
			mean += result.Scores[i][j]
		}
		mean /= float64(len(result.Scores))
		
		if math.Abs(mean) > 1e-6 {
			t.Errorf("Score component %d has non-zero mean: %f", j, mean)
		}
	}
}

// Test transform functionality
func TestTransform(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	// Fit the model
	_, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	// Transform same data
	transformed, err := engine.Transform(data)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	
	// Check dimensions
	if len(transformed) != len(data) {
		t.Errorf("Expected %d transformed rows, got %d", len(data), len(transformed))
	}
	
	if len(transformed[0]) != config.Components {
		t.Errorf("Expected %d transformed columns, got %d", config.Components, len(transformed[0]))
	}
}

// Test SVD-based PCA
func TestPCASVD(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}
	
	result, err := engine.Fit(data, config)
	if err != nil {
		t.Fatalf("PCA SVD fit failed: %v", err)
	}
	
	// Check dimensions
	if len(result.Scores) != len(data) {
		t.Errorf("Expected %d score rows, got %d", len(data), len(result.Scores))
	}
	
	if len(result.Scores[0]) != config.Components {
		t.Errorf("Expected %d score columns, got %d", config.Components, len(result.Scores[0]))
	}
	
	// Check that explained variance sums to ~100%
	totalVar := result.CumulativeVar[len(result.CumulativeVar)-1]
	if math.Abs(totalVar-100.0) > 5.0 {
		t.Errorf("Total explained variance should be ~100%%, got %.2f%%", totalVar)
	}
}

// Compare NIPALS and SVD results
func TestNIPALSvsSVD(t *testing.T) {
	data := createTestMatrix()
	engine := NewPCAEngine()
	
	// Run NIPALS
	configNIPALS := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	resultNIPALS, err := engine.Fit(data, configNIPALS)
	if err != nil {
		t.Fatalf("NIPALS fit failed: %v", err)
	}
	
	// Run SVD
	configSVD := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}
	resultSVD, err := engine.Fit(data, configSVD)
	if err != nil {
		t.Fatalf("SVD fit failed: %v", err)
	}
	
	// Compare explained variance (should be very close)
	for i := 0; i < len(resultNIPALS.ExplainedVar); i++ {
		diff := math.Abs(resultNIPALS.ExplainedVar[i] - resultSVD.ExplainedVar[i])
		if diff > 0.1 { // Allow 0.1% difference
			t.Errorf("Explained variance differs at component %d: NIPALS=%.2f%%, SVD=%.2f%%", 
				i, resultNIPALS.ExplainedVar[i], resultSVD.ExplainedVar[i])
		}
	}
	
	// Check that absolute values of scores are similar (sign may differ)
	for i := 0; i < len(data); i++ {
		for j := 0; j < 2; j++ {
			nipalsAbs := math.Abs(resultNIPALS.Scores[i][j])
			svdAbs := math.Abs(resultSVD.Scores[i][j])
			diff := math.Abs(nipalsAbs - svdAbs)
			if diff > 0.01 {
				t.Errorf("Score magnitudes differ at [%d,%d]: NIPALS=%.4f, SVD=%.4f", 
					i, j, nipalsAbs, svdAbs)
			}
		}
	}
}

// Test error cases
func TestPCAErrors(t *testing.T) {
	engine := NewPCAEngine()
	
	// Test empty data
	_, err := engine.Fit(types.Matrix{}, types.PCAConfig{Components: 1})
	if err == nil {
		t.Error("Expected error for empty data")
	}
	
	// Test too many components
	data := createTestMatrix()
	_, err = engine.Fit(data, types.PCAConfig{
		Components: 20, // More than dimensions
		MeanCenter: true,
	})
	if err == nil {
		t.Error("Expected error for too many components")
	}
	
	// Test transform before fit
	engine2 := NewPCAEngine()
	_, err = engine2.Transform(data)
	if err == nil {
		t.Error("Expected error for transform before fit")
	}
	
	// Test invalid method
	_, err = engine.Fit(data, types.PCAConfig{
		Components: 2,
		Method:     "invalid",
	})
	if err == nil {
		t.Error("Expected error for invalid method")
	}
}

// Test with known performance target size
func TestPerformanceTarget(t *testing.T) {
	// Test with 10,000 × 100 matrix (should complete in <1 second)
	n, m := 10000, 100
	data := make(types.Matrix, n)
	
	// Generate data with structure (to ensure convergence)
	for i := 0; i < n; i++ {
		data[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			// Create data with decreasing variance in each dimension
			data[i][j] = float64(i%100)/100.0*float64(m-j) + float64(j)/100.0
		}
	}
	
	engine := NewPCAEngine()
	config := types.PCAConfig{
		Components:    10,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	start := time.Now()
	_, err := engine.Fit(data, config)
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("PCA fit failed: %v", err)
	}
	
	if duration > time.Second {
		t.Errorf("Performance target not met: 10,000×100 matrix took %v (target: <1s)", duration)
	}
}

// Benchmark NIPALS performance
func BenchmarkNIPALS(b *testing.B) {
	// Create smaller test matrix for benchmarking
	n, m := 100, 20
	data := make(types.Matrix, n)
	
	// Generate structured data
	for i := 0; i < n; i++ {
		data[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			data[i][j] = math.Sin(float64(i)/10.0) * math.Cos(float64(j)/5.0) + 
				0.1*float64(i*j)/float64(n*m)
		}
	}
	
	engine := NewPCAEngine()
	config := types.PCAConfig{
		Components:    5,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "nipals",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Fit(data, config)
		if err != nil {
			b.Fatalf("PCA fit failed: %v", err)
		}
	}
}

// Benchmark SVD performance
func BenchmarkSVD(b *testing.B) {
	// Create smaller test matrix for benchmarking
	n, m := 100, 20
	data := make(types.Matrix, n)
	
	// Generate structured data
	for i := 0; i < n; i++ {
		data[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			data[i][j] = math.Sin(float64(i)/10.0) * math.Cos(float64(j)/5.0) + 
				0.1*float64(i*j)/float64(n*m)
		}
	}
	
	engine := NewPCAEngine()
	config := types.PCAConfig{
		Components:    5,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Fit(data, config)
		if err != nil {
			b.Fatalf("PCA fit failed: %v", err)
		}
	}
}