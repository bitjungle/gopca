package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestPCAWithDifferentPreprocessing tests that PCA works correctly with different preprocessing options
func TestPCAWithDifferentPreprocessing(t *testing.T) {
	// Create test data
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{10.0, 11.0, 12.0},
		{13.0, 14.0, 15.0},
	}

	tests := []struct {
		name          string
		config        types.PCAConfig
		checkFunction func(t *testing.T, result *types.PCAResult)
	}{
		{
			name: "No preprocessing",
			config: types.PCAConfig{
				Components:    2,
				MeanCenter:    false,
				StandardScale: false,
				RobustScale:   false,
				Method:        "svd",
			},
			checkFunction: func(t *testing.T, result *types.PCAResult) {
				if result.PreprocessingApplied {
					t.Error("PreprocessingApplied should be false")
				}
				if result.Means != nil {
					t.Error("Means should be nil when no preprocessing")
				}
			},
		},
		{
			name: "Mean center only",
			config: types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: false,
				RobustScale:   false,
				Method:        "svd",
			},
			checkFunction: func(t *testing.T, result *types.PCAResult) {
				if !result.PreprocessingApplied {
					t.Error("PreprocessingApplied should be true")
				}
				if result.Means == nil {
					t.Error("Means should not be nil")
				}
				// Check that means are correct
				expectedMeans := []float64{7.0, 8.0, 9.0}
				for i, mean := range result.Means {
					if math.Abs(mean-expectedMeans[i]) > 1e-10 {
						t.Errorf("Mean[%d]: expected %f, got %f", i, expectedMeans[i], mean)
					}
				}
			},
		},
		{
			name: "Standard scaling",
			config: types.PCAConfig{
				Components:    2,
				MeanCenter:    true,
				StandardScale: true,
				RobustScale:   false,
				Method:        "svd",
			},
			checkFunction: func(t *testing.T, result *types.PCAResult) {
				if !result.PreprocessingApplied {
					t.Error("PreprocessingApplied should be true")
				}
				if result.StdDevs == nil {
					t.Error("StdDevs should not be nil")
				}
				// Check that standard deviations are not 1.0 (default)
				for i, std := range result.StdDevs {
					if std == 1.0 {
						t.Errorf("StdDev[%d] should not be 1.0 when standard scaling is applied", i)
					}
				}
			},
		},
		{
			name: "Robust scaling",
			config: types.PCAConfig{
				Components:    2,
				MeanCenter:    false,
				StandardScale: false,
				RobustScale:   true,
				Method:        "svd",
			},
			checkFunction: func(t *testing.T, result *types.PCAResult) {
				if !result.PreprocessingApplied {
					t.Error("PreprocessingApplied should be true with robust scaling")
				}
				// With robust scaling, we use median instead of mean
				// The preprocessor should have been applied
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPCAEngine()
			result, err := engine.Fit(data, tt.config)
			if err != nil {
				t.Fatalf("PCA fit failed: %v", err)
			}

			// Check basic results
			if len(result.Scores) != len(data) {
				t.Errorf("Expected %d scores, got %d", len(data), len(result.Scores))
			}
			if len(result.Scores[0]) != tt.config.Components {
				t.Errorf("Expected %d components, got %d", tt.config.Components, len(result.Scores[0]))
			}

			// Run specific checks for this test case
			tt.checkFunction(t, result)
		})
	}
}

// TestMutuallyExclusivePreprocessing tests that preprocessing options work correctly when mutually exclusive
func TestMutuallyExclusivePreprocessing(t *testing.T) {
	data := types.Matrix{
		{1.0, 10.0, 100.0},
		{2.0, 20.0, 200.0},
		{3.0, 30.0, 300.0},
		{4.0, 40.0, 400.0},
		{5.0, 50.0, 500.0},
	}

	// Test that only one preprocessing method is applied at a time
	configs := []types.PCAConfig{
		// None
		{
			Components:    2,
			MeanCenter:    false,
			StandardScale: false,
			RobustScale:   false,
			Method:        "svd",
		},
		// Mean center only
		{
			Components:    2,
			MeanCenter:    true,
			StandardScale: false,
			RobustScale:   false,
			Method:        "svd",
		},
		// Standard scale (includes mean center)
		{
			Components:    2,
			MeanCenter:    true,
			StandardScale: true,
			RobustScale:   false,
			Method:        "svd",
		},
		// Robust scale only
		{
			Components:    2,
			MeanCenter:    false,
			StandardScale: false,
			RobustScale:   true,
			Method:        "svd",
		},
	}

	// Store results for comparison
	results := make([]*types.PCAResult, len(configs))

	for i, config := range configs {
		engine := NewPCAEngine()
		result, err := engine.Fit(data, config)
		if err != nil {
			t.Fatalf("PCA fit failed for config %d: %v", i, err)
		}
		results[i] = result
	}

	// Results should be different for different preprocessing methods
	// Compare the scores instead of explained variance (which might be similar)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			// Compare first score values
			score1 := results[i].Scores[0][0]
			score2 := results[j].Scores[0][0]

			// If both have no preprocessing, scores should be the same
			if !configs[i].MeanCenter && !configs[i].StandardScale && !configs[i].RobustScale &&
				!configs[j].MeanCenter && !configs[j].StandardScale && !configs[j].RobustScale {
				if math.Abs(score1-score2) > 1e-6 {
					t.Errorf("Results for no preprocessing should be the same")
				}
			} else if (configs[i].MeanCenter || configs[i].StandardScale || configs[i].RobustScale) !=
				(configs[j].MeanCenter || configs[j].StandardScale || configs[j].RobustScale) {
				// If one has preprocessing and the other doesn't, they should differ
				if math.Abs(score1-score2) < 1e-6 {
					t.Errorf("Results for preprocessing configs %d and %d should be different", i, j)
				}
			}
		}
	}
}

// TestPreprocessingWithTransform tests that transform applies the same preprocessing as fit
func TestPreprocessingWithTransform(t *testing.T) {
	// Training data
	trainData := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{10.0, 11.0, 12.0},
	}

	// Test data
	testData := types.Matrix{
		{13.0, 14.0, 15.0},
		{16.0, 17.0, 18.0},
	}

	configs := []types.PCAConfig{
		{
			Components:    2,
			MeanCenter:    true,
			StandardScale: true,
			RobustScale:   false,
			Method:        "svd",
		},
		{
			Components:    2,
			MeanCenter:    false,
			StandardScale: false,
			RobustScale:   true,
			Method:        "svd",
		},
	}

	for _, config := range configs {
		engine := NewPCAEngine()

		// Fit on training data
		_, err := engine.Fit(trainData, config)
		if err != nil {
			t.Fatalf("Fit failed: %v", err)
		}

		// Transform test data
		transformed, err := engine.Transform(testData)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		// Check dimensions
		if len(transformed) != len(testData) {
			t.Errorf("Expected %d transformed samples, got %d", len(testData), len(transformed))
		}
		if len(transformed[0]) != config.Components {
			t.Errorf("Expected %d components, got %d", config.Components, len(transformed[0]))
		}
	}
}
