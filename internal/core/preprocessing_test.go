package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// Test basic preprocessing
func TestPreprocessorFitTransform(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
	}

	// Test mean centering
	prep := NewPreprocessor(true, false, false)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that each column has zero mean
	for j := 0; j < len(transformed[0]); j++ {
		sum := 0.0
		for i := 0; i < len(transformed); i++ {
			sum += transformed[i][j]
		}
		mean := sum / float64(len(transformed))
		if math.Abs(mean) > 1e-10 {
			t.Errorf("Column %d has non-zero mean: %f", j, mean)
		}
	}
}

// Test standard scaling
func TestStandardScaling(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0},
		{2.0, 20.0},
		{3.0, 30.0},
		{4.0, 40.0},
	}

	prep := NewPreprocessor(true, true, false)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that each column has unit variance
	for j := 0; j < len(transformed[0]); j++ {
		sumSq := 0.0
		for i := 0; i < len(transformed); i++ {
			sumSq += transformed[i][j] * transformed[i][j]
		}
		variance := sumSq / float64(len(transformed)-1)
		if math.Abs(variance-1.0) > 1e-10 {
			t.Errorf("Column %d does not have unit variance: %f", j, variance)
		}
	}
}

// Test robust scaling
func TestRobustScaling(t *testing.T) {
	// Data with outliers
	data := types.Matrix{
		{1.0, 2.0},
		{2.0, 3.0},
		{3.0, 4.0},
		{100.0, 5.0}, // Outlier in first column
		{4.0, 6.0},
	}

	prep := NewPreprocessor(false, false, true)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that outlier is scaled but not dominating
	if math.Abs(transformed[3][0]) < 2.0 {
		t.Error("Outlier not properly scaled")
	}
}

// Test SNV preprocessing
func TestSNVPreprocessing(t *testing.T) {
	// Test data where each row has different mean and variance
	data := types.Matrix{
		{1.0, 2.0, 3.0, 4.0},     // mean=2.5, std≈1.29
		{10.0, 20.0, 30.0, 40.0}, // mean=25, std≈12.9 (10x the first row)
		{5.0, 5.0, 5.0, 5.0},     // mean=5, std=0 (constant row)
	}

	prep := NewPreprocessorWithSNV(false, false, false, true)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that each row (except constant) has zero mean and unit variance
	for i := 0; i < 2; i++ { // First two rows
		rowSum := 0.0
		for j := 0; j < len(transformed[i]); j++ {
			rowSum += transformed[i][j]
		}
		rowMean := rowSum / float64(len(transformed[i]))
		if math.Abs(rowMean) > 1e-10 {
			t.Errorf("Row %d has non-zero mean after SNV: %f", i, rowMean)
		}

		// Check variance
		rowSumSq := 0.0
		for j := 0; j < len(transformed[i]); j++ {
			diff := transformed[i][j] - rowMean
			rowSumSq += diff * diff
		}
		rowVar := rowSumSq / float64(len(transformed[i])-1)
		if math.Abs(rowVar-1.0) > 1e-10 {
			t.Errorf("Row %d does not have unit variance after SNV: %f", i, rowVar)
		}
	}

	// Check that constant row is centered but not scaled
	constantRow := transformed[2]
	for j := 0; j < len(constantRow); j++ {
		if math.Abs(constantRow[j]) > 1e-10 {
			t.Errorf("Constant row value %d is not zero after SNV: %f", j, constantRow[j])
		}
	}
}

// Test SNV combined with column preprocessing
func TestSNVWithColumnPreprocessing(t *testing.T) {
	// Use diverse data to ensure variance after SNV
	data := types.Matrix{
		{1.0, 4.0, 7.0},
		{2.0, 5.0, 8.0},
		{3.0, 6.0, 9.0},
		{10.0, 5.0, 2.0}, // Different pattern
	}

	// Test 1: Just SNV
	prep1 := NewPreprocessorWithSNV(false, false, false, true)
	snvOnly, err := prep1.FitTransform(data)
	if err != nil {
		t.Fatalf("SNV only FitTransform failed: %v", err)
	}

	// Verify each row has zero mean after SNV
	for i := 0; i < len(snvOnly); i++ {
		rowSum := 0.0
		for j := 0; j < len(snvOnly[i]); j++ {
			rowSum += snvOnly[i][j]
		}
		rowMean := rowSum / float64(len(snvOnly[i]))
		if math.Abs(rowMean) > 1e-10 {
			t.Errorf("Row %d has non-zero mean after SNV: %f", i, rowMean)
		}
	}

	// Test 2: SNV + Mean centering (should center columns after SNV)
	prep2 := NewPreprocessorWithSNV(true, false, false, true)
	transformed, err := prep2.FitTransform(data)
	if err != nil {
		t.Fatalf("SNV + mean center FitTransform failed: %v", err)
	}

	// After SNV + mean centering, columns should have zero mean
	for j := 0; j < len(transformed[0]); j++ {
		colSum := 0.0
		for i := 0; i < len(transformed); i++ {
			colSum += transformed[i][j]
		}
		colMean := colSum / float64(len(transformed))
		if math.Abs(colMean) > 1e-10 {
			t.Errorf("Column %d has non-zero mean after SNV + mean centering: %f", j, colMean)
		}
	}
}

// Test inverse transform
func TestInverseTransform(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
	}

	prep := NewPreprocessor(true, true, false)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Inverse transform
	inversed, err := prep.InverseTransform(transformed)
	if err != nil {
		t.Fatalf("InverseTransform failed: %v", err)
	}

	// Check that we get back original data
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[0]); j++ {
			if math.Abs(inversed[i][j]-data[i][j]) > 1e-10 {
				t.Errorf("Inverse transform failed at [%d,%d]: expected %f, got %f",
					i, j, data[i][j], inversed[i][j])
			}
		}
	}
}

// Test SNV against reference implementation
func TestSNVReferenceImplementation(t *testing.T) {
	// Test with known input/output from Python reference
	data := types.Matrix{
		{1.0, 2.0, 3.0, 4.0, 5.0},
		{2.0, 4.0, 6.0, 8.0, 10.0},
	}

	prep := NewPreprocessorWithSNV(false, false, false, true)
	transformed, err := prep.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Expected values for first row after SNV
	// mean = 3, std = sqrt(2.5) ≈ 1.58114
	// Expected: [-1.26491, -0.63246, 0, 0.63246, 1.26491]
	expected := []float64{-1.26491, -0.63246, 0, 0.63246, 1.26491}

	for j, exp := range expected {
		if math.Abs(transformed[0][j]-exp) > 0.0001 {
			t.Errorf("Row 0, Col %d: expected %f, got %f", j, exp, transformed[0][j])
		}
	}

	// Second row should have same normalized values (scaled version of first)
	for j := 0; j < len(expected); j++ {
		if math.Abs(transformed[1][j]-expected[j]) > 0.0001 {
			t.Errorf("Row 1, Col %d: expected %f, got %f", j, expected[j], transformed[1][j])
		}
	}
}

// Test missing value imputation
func TestImputeMissing(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{math.NaN(), 3.0},
		{3.0, 4.0},
		{4.0, math.NaN()},
	}

	// Test mean imputation
	imputed, err := ImputeMissing(data, MissingMean)
	if err != nil {
		t.Fatalf("ImputeMissing failed: %v", err)
	}

	// Check that NaN values are replaced
	if math.IsNaN(imputed[1][0]) {
		t.Error("NaN not replaced at [1,0]")
	}
	if math.IsNaN(imputed[3][1]) {
		t.Error("NaN not replaced at [3,1]")
	}

	// Check imputed values
	expectedMean0 := (1.0 + 3.0 + 4.0) / 3.0
	if math.Abs(imputed[1][0]-expectedMean0) > 1e-10 {
		t.Errorf("Wrong imputed value at [1,0]: expected %f, got %f", expectedMean0, imputed[1][0])
	}
}

// Test row/column selection
func TestSelectRowsColumns(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{10.0, 11.0, 12.0},
	}

	subset, err := SelectRowsColumns(data, []int{0, 2}, []int{1, 2})
	if err != nil {
		t.Fatalf("SelectRowsColumns failed: %v", err)
	}

	expected := types.Matrix{
		{2.0, 3.0},
		{8.0, 9.0},
	}

	if len(subset) != len(expected) || len(subset[0]) != len(expected[0]) {
		t.Errorf("Wrong subset dimensions: got %dx%d, expected %dx%d",
			len(subset), len(subset[0]), len(expected), len(expected[0]))
	}

	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(expected[0]); j++ {
			if subset[i][j] != expected[i][j] {
				t.Errorf("Wrong value at [%d,%d]: got %f, expected %f",
					i, j, subset[i][j], expected[i][j])
			}
		}
	}
}

// Test outlier removal
func TestRemoveOutliers(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{2.0, 3.0},
		{3.0, 4.0},
		{100.0, 5.0}, // Outlier in first column
		{4.0, 6.0},
	}

	// Use a lower threshold to ensure outlier detection
	cleaned, keepRows, err := RemoveOutliers(data, 1.5) // 1.5 standard deviations
	if err != nil {
		t.Fatalf("RemoveOutliers failed: %v", err)
	}

	// Should remove the outlier row
	if len(cleaned) >= len(data) {
		t.Errorf("Outlier not removed: original %d rows, cleaned %d rows", len(data), len(cleaned))

		// Debug: print z-scores
		n, m := len(data), len(data[0])
		for j := 0; j < m; j++ {
			col := make([]float64, n)
			for i := 0; i < n; i++ {
				col[i] = data[i][j]
			}
			mean := 0.0
			for _, v := range col {
				mean += v
			}
			mean /= float64(n)

			sumSq := 0.0
			for _, v := range col {
				sumSq += (v - mean) * (v - mean)
			}
			stdDev := math.Sqrt(sumSq / float64(n-1))

			t.Logf("Column %d: mean=%.2f, stdDev=%.2f", j, mean, stdDev)
			for i := 0; i < n; i++ {
				zScore := math.Abs((data[i][j] - mean) / stdDev)
				t.Logf("  Row %d: value=%.2f, z-score=%.2f", i, data[i][j], zScore)
			}
		}
	}

	// Check that outlier row (index 3) is not in keepRows
	for _, idx := range keepRows {
		if idx == 3 {
			t.Error("Outlier row still in keepRows")
		}
	}
}

// Test variable transformations
func TestApplyTransform(t *testing.T) {
	data := types.Matrix{
		{1.0, 4.0},
		{2.0, 9.0},
		{3.0, 16.0},
	}

	// Test square root transformation
	transformed, err := ApplyTransform(data, []int{1}, TransformSqrt)
	if err != nil {
		t.Fatalf("ApplyTransform failed: %v", err)
	}

	// Check transformed values
	expected := []float64{2.0, 3.0, 4.0}
	for i := 0; i < len(transformed); i++ {
		if math.Abs(transformed[i][1]-expected[i]) > 1e-10 {
			t.Errorf("Wrong transformed value at [%d,1]: got %f, expected %f",
				i, transformed[i][1], expected[i])
		}
	}

	// First column should be unchanged
	for i := 0; i < len(data); i++ {
		if transformed[i][0] != data[i][0] {
			t.Errorf("Column 0 should be unchanged at row %d", i)
		}
	}
}

// Test variance calculation
func TestGetVarianceByColumn(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0},
		{2.0, 10.0},
		{3.0, 10.0},
		{4.0, 10.0},
	}

	variances, err := GetVarianceByColumn(data)
	if err != nil {
		t.Fatalf("GetVarianceByColumn failed: %v", err)
	}

	// First column should have non-zero variance
	if variances[0] <= 0 {
		t.Error("First column should have positive variance")
	}

	// Second column should have zero variance
	if variances[1] != 0 {
		t.Errorf("Second column should have zero variance, got %f", variances[1])
	}
}

// Test column ranking by variance
func TestGetColumnRanks(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0, 5.0},
		{1.1, 20.0, 5.1},
		{0.9, 30.0, 4.9},
		{1.0, 40.0, 5.0},
	}

	ranks, err := GetColumnRanks(data)
	if err != nil {
		t.Fatalf("GetColumnRanks failed: %v", err)
	}

	// Column 1 should have highest variance, then 2, then 0
	if ranks[0] != 1 {
		t.Errorf("Expected column 1 to rank first, got column %d", ranks[0])
	}
}
