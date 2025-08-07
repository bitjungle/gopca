// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestKernelPCA_VarianceScaling(t *testing.T) {
	// Create test data with different scales
	data := types.Matrix{
		{1.0, 100.0, 1000.0},
		{2.0, 200.0, 2000.0},
		{3.0, 300.0, 3000.0},
		{4.0, 400.0, 4000.0},
		{5.0, 500.0, 5000.0},
	}

	// Test without variance scaling
	configNoScaling := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.1,
		ScaleOnly:   false,
	}

	engine1 := NewKernelPCAEngine()
	result1, err := engine1.Fit(data, configNoScaling)
	if err != nil {
		t.Fatalf("Fit without scaling failed: %v", err)
	}

	// Test with variance scaling
	configWithScaling := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.1,
		ScaleOnly:   true,
	}

	engine2 := NewKernelPCAEngine()
	result2, err := engine2.Fit(data, configWithScaling)
	if err != nil {
		t.Fatalf("Fit with scaling failed: %v", err)
	}

	// Compare results - they should be different
	scoresChanged := false
	for i := 0; i < len(result1.Scores); i++ {
		for j := 0; j < len(result1.Scores[i]); j++ {
			if math.Abs(result1.Scores[i][j]-result2.Scores[i][j]) > 1e-10 {
				scoresChanged = true
				break
			}
		}
		if scoresChanged {
			break
		}
	}

	if !scoresChanged {
		t.Error("Variance scaling had no effect on Kernel PCA scores")
	}

	// Verify preprocessing was applied
	if !result2.PreprocessingApplied {
		t.Error("PreprocessingApplied should be true when variance scaling is used")
	}
	if result1.PreprocessingApplied {
		t.Error("PreprocessingApplied should be false when no preprocessing is used")
	}

	// Check that explained variance is also different
	varianceChanged := false
	for i := 0; i < len(result1.ExplainedVar); i++ {
		if math.Abs(result1.ExplainedVar[i]-result2.ExplainedVar[i]) > 1e-10 {
			varianceChanged = true
			break
		}
	}

	if !varianceChanged {
		t.Error("Variance scaling had no effect on explained variance")
	}
}

func TestKernelPCA_VarianceScalingWithIris(t *testing.T) {
	// Simulate Iris-like data with different feature scales
	// Sepal length (cm), Sepal width (cm), Petal length (mm), Petal width (mm)
	irisLikeData := types.Matrix{
		{5.1, 3.5, 14.0, 2.0}, // Note: petal measurements in mm (10x scale)
		{4.9, 3.0, 14.0, 2.0},
		{4.7, 3.2, 13.0, 2.0},
		{4.6, 3.1, 15.0, 2.0},
		{5.0, 3.6, 14.0, 2.0},
		{5.4, 3.9, 17.0, 4.0},
		{4.6, 3.4, 14.0, 3.0},
		{5.0, 3.4, 15.0, 2.0},
		{7.0, 3.2, 47.0, 14.0}, // Different species
		{6.4, 3.2, 45.0, 15.0},
		{6.9, 3.1, 49.0, 15.0},
		{5.5, 2.3, 40.0, 13.0},
	}

	// Run without scaling
	configNoScaling := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
		ScaleOnly:   false,
	}

	engine1 := NewKernelPCAEngine()
	result1, err := engine1.Fit(irisLikeData, configNoScaling)
	if err != nil {
		t.Fatalf("Fit without scaling failed: %v", err)
	}

	// Run with variance scaling
	configWithScaling := types.PCAConfig{
		Components:  2,
		Method:      "kernel",
		KernelType:  "rbf",
		KernelGamma: 0.5,
		ScaleOnly:   true,
	}

	engine2 := NewKernelPCAEngine()
	result2, err := engine2.Fit(irisLikeData, configWithScaling)
	if err != nil {
		t.Fatalf("Fit with scaling failed: %v", err)
	}

	// The scores should be noticeably different
	maxDiff := 0.0
	for i := 0; i < len(result1.Scores); i++ {
		for j := 0; j < len(result1.Scores[i]); j++ {
			diff := math.Abs(result1.Scores[i][j] - result2.Scores[i][j])
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	t.Logf("Maximum difference in scores: %f", maxDiff)

	// With such different scales, we expect a significant difference
	if maxDiff < 0.01 {
		t.Error("Expected larger difference in scores when features have very different scales")
	}
}
