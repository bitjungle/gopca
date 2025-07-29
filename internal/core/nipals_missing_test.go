package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// TestNIPALSWithMissingValues tests NIPALS algorithm with missing data
func TestNIPALSWithMissingValues(t *testing.T) {
	tests := []struct {
		name        string
		data        types.Matrix
		components  int
		wantErr     bool
		checkResult func(*testing.T, *types.PCAResult)
	}{
		{
			name: "simple data with one missing value",
			data: types.Matrix{
				{1.0, 2.5, 3.2},
				{4.1, math.NaN(), 6.0},
				{7.2, 8.1, 9.3},
				{10.0, 11.2, 11.8},
			},
			components: 2,
			wantErr:    false,
			checkResult: func(t *testing.T, result *types.PCAResult) {
				// Check that we got at least 1 component
				if result.ComponentsComputed < 1 {
					t.Errorf("Expected at least 1 component, got %d", result.ComponentsComputed)
				}
				// Check that scores and loadings have correct dimensions
				if len(result.Scores) != 4 {
					t.Errorf("Expected 4 scores rows, got %d", len(result.Scores))
				}
				if len(result.Loadings) != 3 {
					t.Errorf("Expected 3 loadings rows, got %d", len(result.Loadings))
				}
			},
		},
		{
			name: "data with multiple missing values",
			data: types.Matrix{
				{1.0, math.NaN(), 3.0, 4.0},
				{5.0, 6.0, math.NaN(), 8.0},
				{math.NaN(), 10.0, 11.0, 12.0},
				{13.0, 14.0, 15.0, math.NaN()},
				{17.0, 18.0, 19.0, 20.0},
			},
			components: 3,
			wantErr:    false,
			checkResult: func(t *testing.T, result *types.PCAResult) {
				// Check that we got at most 3 components
				if result.ComponentsComputed > 3 {
					t.Errorf("Expected at most 3 components, got %d", result.ComponentsComputed)
				}
				// Verify no NaN in results
				for i := range result.Scores {
					for j := range result.Scores[i] {
						if math.IsNaN(result.Scores[i][j]) {
							t.Errorf("Found NaN in scores at [%d][%d]", i, j)
						}
					}
				}
				for i := range result.Loadings {
					for j := range result.Loadings[i] {
						if math.IsNaN(result.Loadings[i][j]) {
							t.Errorf("Found NaN in loadings at [%d][%d]", i, j)
						}
					}
				}
			},
		},
		{
			name: "data with entire column missing",
			data: types.Matrix{
				{1.0, math.NaN(), 3.0},
				{4.0, math.NaN(), 6.0},
				{7.0, math.NaN(), 9.0},
			},
			components: 2,
			wantErr:    false,
			checkResult: func(t *testing.T, result *types.PCAResult) {
				// Should still work but with reduced effective dimensions
				if result.ComponentsComputed > 2 {
					t.Errorf("Expected at most 2 components due to missing column, got %d", result.ComponentsComputed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPCAEngine()
			config := types.PCAConfig{
				Components:      tt.components,
				Method:          "nipals",
				MissingStrategy: types.MissingNative,
				MeanCenter:      false, // Preprocessing is skipped with native missing handling
			}

			result, err := engine.Fit(tt.data, config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Fit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestNIPALSMissingVsComplete compares NIPALS results with and without missing values
func TestNIPALSMissingVsComplete(t *testing.T) {
	// Create a simple dataset
	completeData := types.Matrix{
		{2.5, 2.4, 3.5, 3.0},
		{0.5, 0.7, 1.5, 1.0},
		{2.2, 2.9, 3.2, 3.5},
		{1.9, 2.2, 2.9, 3.0},
		{3.1, 3.0, 4.1, 4.0},
		{2.3, 2.7, 3.3, 3.5},
	}

	// Create version with missing values (20% missing)
	missingData := make(types.Matrix, len(completeData))
	for i := range completeData {
		missingData[i] = make([]float64, len(completeData[i]))
		copy(missingData[i], completeData[i])
	}
	// Introduce missing values
	missingData[0][1] = math.NaN()
	missingData[2][0] = math.NaN()
	missingData[3][3] = math.NaN()
	missingData[5][2] = math.NaN()

	engine := NewPCAEngine()

	// Run PCA on complete data
	configComplete := types.PCAConfig{
		Components: 2,
		Method:     "nipals",
		MeanCenter: false, // Keep consistent with missing data test
	}
	resultComplete, err := engine.Fit(completeData, configComplete)
	if err != nil {
		t.Fatalf("Failed to fit complete data: %v", err)
	}

	// Run PCA on missing data with native handling
	configMissing := types.PCAConfig{
		Components:      2,
		Method:          "nipals",
		MissingStrategy: types.MissingNative,
		MeanCenter:      false, // Don't mean center as it might interfere with missing values
	}
	resultMissing, err := engine.Fit(missingData, configMissing)
	if err != nil {
		t.Fatalf("Failed to fit missing data: %v", err)
	}

	// Compare results - they should be similar but not identical
	// First check that both computed at least one component
	if len(resultComplete.ExplainedVar) == 0 || len(resultMissing.ExplainedVar) == 0 {
		t.Fatalf("No components computed: complete=%d, missing=%d",
			len(resultComplete.ExplainedVar), len(resultMissing.ExplainedVar))
	}

	// For NIPALS with missing values, we get eigenvalues not percentages
	// So we can't directly compare the explained variance values
	// Instead, check that both methods found meaningful components
	if resultComplete.ExplainedVar[0] <= 0 || resultMissing.ExplainedVar[0] <= 0 {
		t.Errorf("Invalid variance values: complete=%.2f, missing=%.2f",
			resultComplete.ExplainedVar[0], resultMissing.ExplainedVar[0])
	}

	// Check that loadings have similar patterns (sign may differ)
	minLoadings := len(resultComplete.Loadings)
	if len(resultMissing.Loadings) < minLoadings {
		minLoadings = len(resultMissing.Loadings)
	}
	for j := 0; j < minLoadings; j++ {
		if len(resultComplete.Loadings[j]) > 0 && len(resultMissing.Loadings[j]) > 0 {
			loading1 := resultComplete.Loadings[j][0]
			loading2 := resultMissing.Loadings[j][0]
			// Check if they have similar magnitude (allowing for sign flip)
			if math.Abs(math.Abs(loading1)-math.Abs(loading2)) > 0.3 {
				t.Errorf("Loading[%d][0] differs too much: complete=%.3f, missing=%.3f",
					j, loading1, loading2)
			}
		}
	}
}

// TestNIPALSConvergenceWithMissing tests that NIPALS converges with missing data
func TestNIPALSConvergenceWithMissing(t *testing.T) {
	// Create a dataset that might challenge convergence
	data := types.Matrix{
		{1.0, 2.0, math.NaN(), 4.0, 5.0},
		{2.0, math.NaN(), 6.0, 8.0, 10.0},
		{3.0, 6.0, 9.0, math.NaN(), 15.0},
		{4.0, 8.0, math.NaN(), 16.0, 20.0},
		{5.0, 10.0, 15.0, 20.0, math.NaN()},
		{6.0, math.NaN(), 18.0, 24.0, 30.0},
	}

	impl := &PCAImpl{}
	X := utils.MatrixToDense(data)

	scores, loadings, err := impl.nipalsAlgorithmWithMissing(X, 3)
	if err != nil {
		t.Fatalf("NIPALS failed to converge: %v", err)
	}

	// Verify orthogonality of loadings
	P := loadings
	_, k := P.Dims()
	for i := 0; i < k; i++ {
		for j := i + 1; j < k; j++ {
			pi := mat.Col(nil, i, P)
			pj := mat.Col(nil, j, P)
			dot := floats.Dot(pi, pj)
			if math.Abs(dot) > 1e-6 {
				t.Errorf("Loadings %d and %d not orthogonal: dot product = %g", i, j, dot)
			}
		}
	}

	// Verify unit norm of loadings
	for i := 0; i < k; i++ {
		pi := mat.Col(nil, i, P)
		norm := floats.Norm(pi, 2)
		if math.Abs(norm-1.0) > 1e-6 {
			t.Errorf("Loading %d not unit norm: norm = %g", i, norm)
		}
	}

	// Check that scores don't contain NaN
	n, _ := scores.Dims()
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			if math.IsNaN(scores.At(i, j)) {
				t.Errorf("Score[%d][%d] is NaN", i, j)
			}
		}
	}
}

// TestNIPALSEdgeCases tests edge cases for NIPALS with missing values
func TestNIPALSEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		data    types.Matrix
		wantErr bool
	}{
		{
			name: "sparse data with 50% missing",
			data: types.Matrix{
				{1.0, math.NaN(), 3.0, math.NaN()},
				{math.NaN(), 6.0, math.NaN(), 8.0},
				{9.0, math.NaN(), 11.0, math.NaN()},
				{math.NaN(), 14.0, math.NaN(), 16.0},
			},
			wantErr: false,
		},
		{
			name: "data with only one complete column",
			data: types.Matrix{
				{1.0, math.NaN(), math.NaN()},
				{2.0, math.NaN(), math.NaN()},
				{3.0, math.NaN(), math.NaN()},
			},
			wantErr: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPCAEngine()
			config := types.PCAConfig{
				Components:      2,
				Method:          "nipals",
				MissingStrategy: types.MissingNative,
				MeanCenter:      false, // Preprocessing is skipped with native missing handling
			}

			_, err := engine.Fit(tt.data, config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
