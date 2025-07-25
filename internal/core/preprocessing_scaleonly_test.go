package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestVarianceScalingPreprocessing(t *testing.T) {
	tests := []struct {
		name       string
		data       types.Matrix
		wantMeans  []float64
		wantStdDev []float64
	}{
		{
			name: "simple 2x2 matrix",
			data: types.Matrix{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			wantMeans:  []float64{2.0, 3.0},
			wantStdDev: []float64{math.Sqrt(2), math.Sqrt(2)},
		},
		{
			name: "3x3 matrix with different scales",
			data: types.Matrix{
				{1.0, 10.0, 100.0},
				{2.0, 20.0, 200.0},
				{3.0, 30.0, 300.0},
			},
			wantMeans:  []float64{2.0, 20.0, 200.0},
			wantStdDev: []float64{1.0, 10.0, 100.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create variance scaling preprocessor (scale-only = true)
			preprocessor := NewPreprocessorWithScaleOnly(false, false, false, true, false, false)

			// Fit and transform
			transformed, err := preprocessor.FitTransform(tt.data)
			if err != nil {
				t.Fatalf("FitTransform failed: %v", err)
			}

			// Check that means are preserved (not centered)
			means := preprocessor.GetMeans()
			for i, mean := range means {
				if math.Abs(mean-tt.wantMeans[i]) > 1e-10 {
					t.Errorf("Mean[%d] = %v, want %v", i, mean, tt.wantMeans[i])
				}
			}

			// Check that standard deviations are correct
			stdDevs := preprocessor.GetStdDevs()
			for i, std := range stdDevs {
				if math.Abs(std-tt.wantStdDev[i]) > 1e-10 {
					t.Errorf("StdDev[%d] = %v, want %v", i, std, tt.wantStdDev[i])
				}
			}

			// Check that data is scaled but not centered
			for i := range tt.data {
				for j := range tt.data[i] {
					expected := tt.data[i][j] / tt.wantStdDev[j]
					if math.Abs(transformed[i][j]-expected) > 1e-10 {
						t.Errorf("Transformed[%d][%d] = %v, want %v", i, j, transformed[i][j], expected)
					}
				}
			}
		})
	}
}

func TestVarianceScalingVsStandardScale(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0},
		{2.0, 20.0},
		{3.0, 30.0},
		{4.0, 40.0},
	}

	// Standard scale (with mean centering)
	standardPreprocessor := NewPreprocessor(true, true, false)
	standardTransformed, err := standardPreprocessor.FitTransform(data)
	if err != nil {
		t.Fatalf("Standard scale FitTransform failed: %v", err)
	}

	// Scale-only (without mean centering)
	scaleOnlyPreprocessor := NewPreprocessorWithScaleOnly(false, false, false, true, false, false)
	scaleOnlyTransformed, err := scaleOnlyPreprocessor.FitTransform(data)
	if err != nil {
		t.Fatalf("Scale-only FitTransform failed: %v", err)
	}

	// Check that standard scale centers data (mean should be ~0)
	for j := 0; j < len(data[0]); j++ {
		sum := 0.0
		for i := 0; i < len(data); i++ {
			sum += standardTransformed[i][j]
		}
		mean := sum / float64(len(data))
		if math.Abs(mean) > 1e-10 {
			t.Errorf("Standard scale column %d mean = %v, want ~0", j, mean)
		}
	}

	// Check that scale-only does NOT center data (mean should be preserved)
	means := scaleOnlyPreprocessor.GetMeans()
	stdDevs := scaleOnlyPreprocessor.GetStdDevs()
	for j := 0; j < len(data[0]); j++ {
		sum := 0.0
		for i := 0; i < len(data); i++ {
			sum += scaleOnlyTransformed[i][j]
		}
		scaledMean := sum / float64(len(data))
		expectedMean := means[j] / stdDevs[j]
		if math.Abs(scaledMean-expectedMean) > 1e-10 {
			t.Errorf("Scale-only column %d mean = %v, want %v", j, scaledMean, expectedMean)
		}
	}
}

func TestVarianceScalingWithZeroVariance(t *testing.T) {
	// Test with a column that has zero variance
	data := types.Matrix{
		{1.0, 5.0, 10.0},
		{2.0, 5.0, 20.0},
		{3.0, 5.0, 30.0},
	}

	preprocessor := NewPreprocessorWithScaleOnly(false, false, false, true, false, false)
	transformed, err := preprocessor.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that zero-variance column is unchanged (scale = 1.0)
	for i := range data {
		if transformed[i][1] != data[i][1] {
			t.Errorf("Zero variance column changed: row %d, got %v want %v",
				i, transformed[i][1], data[i][1])
		}
	}
}

func TestVarianceScalingInverseTransform(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0, 100.0},
		{2.0, 20.0, 200.0},
		{3.0, 30.0, 300.0},
		{4.0, 40.0, 400.0},
	}

	preprocessor := NewPreprocessorWithScaleOnly(false, false, false, true, false, false)

	// Transform
	transformed, err := preprocessor.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Inverse transform
	inversed, err := preprocessor.InverseTransform(transformed)
	if err != nil {
		t.Fatalf("InverseTransform failed: %v", err)
	}

	// Check that we get back the original data
	for i := range data {
		for j := range data[i] {
			if math.Abs(inversed[i][j]-data[i][j]) > 1e-10 {
				t.Errorf("Inverse[%d][%d] = %v, want %v", i, j, inversed[i][j], data[i][j])
			}
		}
	}
}

func TestVarianceScalingMutuallyExclusive(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}

	// Test that scale-only is applied when both scale-only and standard scale are true
	// (scale-only should take precedence based on the logic in Transform)
	preprocessor := NewPreprocessorWithScaleOnly(true, true, false, true, false, false)
	transformed, err := preprocessor.FitTransform(data)
	if err != nil {
		t.Fatalf("FitTransform failed: %v", err)
	}

	// Check that data is scaled but not centered (scale-only behavior)
	stdDevs := preprocessor.GetStdDevs()

	for i := range data {
		for j := range data[i] {
			expected := data[i][j] / stdDevs[j]
			if math.Abs(transformed[i][j]-expected) > 1e-10 {
				t.Errorf("Transformed[%d][%d] = %v, want %v (scale-only should take precedence)",
					i, j, transformed[i][j], expected)
			}
		}
	}
}
