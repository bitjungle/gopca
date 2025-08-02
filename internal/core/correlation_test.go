package core

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// TestPearsonCorrelation tests the Pearson correlation calculation
func TestPearsonCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		wantR     float64
		wantP     float64
		tolerance float64
		wantErr   bool
	}{
		{
			name:      "Perfect positive correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{2, 4, 6, 8, 10},
			wantR:     1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Perfect negative correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{10, 8, 6, 4, 2},
			wantR:     -1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "No correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{3, 1, 4, 1, 5},
			wantR:     0.0,
			wantP:     1.0,
			tolerance: 0.4, // More tolerance for no correlation
			wantErr:   false,
		},
		{
			name:      "Strong positive correlation",
			x:         []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			y:         []float64{2.1, 3.9, 6.1, 7.8, 10.2, 11.9, 14.1, 15.8, 18.2, 19.9},
			wantR:     0.999,
			wantP:     0.001,
			tolerance: 0.01,
			wantErr:   false,
		},
		{
			name:      "With missing values",
			x:         []float64{1, 2, math.NaN(), 4, 5},
			y:         []float64{2, 4, 6, math.NaN(), 10},
			wantR:     1.0,
			wantP:     0.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:    "Too few observations",
			x:       []float64{1, 2},
			y:       []float64{2, 4},
			wantErr: true,
		},
		{
			name:    "All missing values",
			x:       []float64{math.NaN(), math.NaN(), math.NaN()},
			y:       []float64{math.NaN(), math.NaN(), math.NaN()},
			wantErr: true,
		},
		{
			name:    "Different lengths",
			x:       []float64{1, 2, 3},
			y:       []float64{1, 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, p, err := pearsonCorrelation(tt.x, tt.y)

			if tt.wantErr {
				if err == nil {
					t.Errorf("pearsonCorrelation() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("pearsonCorrelation() unexpected error = %v", err)
				return
			}

			if math.Abs(r-tt.wantR) > tt.tolerance {
				t.Errorf("pearsonCorrelation() r = %v, want %v (tolerance %v)", r, tt.wantR, tt.tolerance)
			}

			// For p-values, check significance thresholds
			// Strong correlations should have low p-values
			if math.Abs(r) > 0.9 && p > 0.05 {
				t.Errorf("pearsonCorrelation() p = %v, expected significant (< 0.05) for r = %v", p, r)
			}
			// Weak correlations should have high p-values
			if math.Abs(r) < 0.3 && p < 0.5 {
				t.Errorf("pearsonCorrelation() p = %v, expected non-significant (> 0.5) for r = %v", p, r)
			}
		})
	}
}

// TestSpearmanCorrelation tests the Spearman correlation calculation
func TestSpearmanCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		wantR     float64
		tolerance float64
		wantErr   bool
	}{
		{
			name:      "Perfect monotonic positive",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{1, 4, 9, 16, 25}, // y = x^2, perfect monotonic
			wantR:     1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Perfect monotonic negative",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{25, 16, 9, 4, 1},
			wantR:     -1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "Non-linear but monotonic",
			x:         []float64{1, 2, 3, 4, 5, 6, 7, 8},
			y:         []float64{1, 8, 27, 64, 125, 216, 343, 512}, // y = x^3
			wantR:     1.0,
			tolerance: 1e-10,
			wantErr:   false,
		},
		{
			name:      "With ties",
			x:         []float64{1, 2, 2, 3, 4, 4, 5},
			y:         []float64{1, 2, 3, 4, 5, 5, 6},
			wantR:     0.95,
			tolerance: 0.1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _, err := spearmanCorrelation(tt.x, tt.y)

			if tt.wantErr {
				if err == nil {
					t.Errorf("spearmanCorrelation() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("spearmanCorrelation() unexpected error = %v", err)
				return
			}

			if math.Abs(r-tt.wantR) > tt.tolerance {
				t.Errorf("spearmanCorrelation() r = %v, want %v (tolerance %v)", r, tt.wantR, tt.tolerance)
			}
		})
	}
}

// TestRank tests the ranking function
func TestRank(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want []float64
	}{
		{
			name: "No ties",
			x:    []float64{3, 1, 4, 1.5, 9},
			want: []float64{3, 1, 4, 2, 5},
		},
		{
			name: "With ties",
			x:    []float64{1, 2, 2, 3, 3, 3, 4},
			want: []float64{1, 2.5, 2.5, 5, 5, 5, 7},
		},
		{
			name: "All same",
			x:    []float64{5, 5, 5, 5},
			want: []float64{2.5, 2.5, 2.5, 2.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rank(tt.x)
			for i := range got {
				if math.Abs(got[i]-tt.want[i]) > 1e-10 {
					t.Errorf("rank() at index %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestOneHotEncode tests the one-hot encoding function
func TestOneHotEncode(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		wantKeys   []string
		wantValues map[string][]float64
	}{
		{
			name:       "Basic categories",
			categories: []string{"A", "B", "A", "C", "B"},
			wantKeys:   []string{"A", "B", "C"},
			wantValues: map[string][]float64{
				"A": {1, 0, 1, 0, 0},
				"B": {0, 1, 0, 0, 1},
				"C": {0, 0, 0, 1, 0},
			},
		},
		{
			name:       "With empty categories",
			categories: []string{"A", "", "B", "", "A"},
			wantKeys:   []string{"A", "B"},
			wantValues: map[string][]float64{
				"A": {1, 0, 0, 0, 1},
				"B": {0, 0, 1, 0, 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := oneHotEncode(tt.categories)

			// Check number of encoded variables
			if len(got) != len(tt.wantKeys) {
				t.Errorf("oneHotEncode() returned %d variables, want %d", len(got), len(tt.wantKeys))
			}

			// Check each expected key exists with correct values
			for _, key := range tt.wantKeys {
				values, exists := got[key]
				if !exists {
					t.Errorf("oneHotEncode() missing key %s", key)
					continue
				}

				wantValues := tt.wantValues[key]
				if len(values) != len(wantValues) {
					t.Errorf("oneHotEncode() key %s has %d values, want %d", key, len(values), len(wantValues))
					continue
				}

				for i := range values {
					if values[i] != wantValues[i] {
						t.Errorf("oneHotEncode() key %s at index %d = %v, want %v", key, i, values[i], wantValues[i])
					}
				}
			}
		})
	}
}

// TestCalculateEigencorrelations tests the main correlation calculation function
func TestCalculateEigencorrelations(t *testing.T) {
	// Create test PC scores
	scores := mat.NewDense(10, 3, []float64{
		// PC1  PC2  PC3
		1, 0, 0,
		2, 0, 0,
		3, 0, 0,
		4, 0, 0,
		5, 0, 0,
		6, 0, 0,
		7, 0, 0,
		8, 0, 0,
		9, 0, 0,
		10, 0, 0,
	})

	tests := []struct {
		name             string
		request          CorrelationRequest
		wantNumVars      int
		checkCorrelation func(t *testing.T, result *CorrelationResult)
		wantErr          bool
	}{
		{
			name: "Numeric variables with Pearson",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // Perfect correlation with PC1
					"var2": {10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, // Perfect negative correlation with PC1
				},
				Components: []int{0, 1, 2},
				Method:     "pearson",
			},
			wantNumVars: 2,
			checkCorrelation: func(t *testing.T, result *CorrelationResult) {
				// Check var1 correlation with PC1
				if math.Abs(result.Correlations["var1"][0]-1.0) > 0.001 {
					t.Errorf("var1 PC1 correlation = %v, want ~1.0", result.Correlations["var1"][0])
				}
				// Check var2 correlation with PC1
				if math.Abs(result.Correlations["var2"][0]+1.0) > 0.001 {
					t.Errorf("var2 PC1 correlation = %v, want ~-1.0", result.Correlations["var2"][0])
				}
			},
			wantErr: false,
		},
		{
			name: "Categorical variables",
			request: CorrelationRequest{
				Scores: scores,
				MetadataCategorical: map[string][]string{
					"group": {"A", "A", "A", "B", "B", "B", "C", "C", "C", "C"},
				},
				Components: []int{0},
				Method:     "pearson",
			},
			wantNumVars: 3, // One-hot encoded: group_A, group_B, group_C
			wantErr:     false,
		},
		{
			name: "Spearman correlation",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 4, 9, 16, 25, 36, 49, 64, 81, 100}, // Non-linear but monotonic
				},
				Components: []int{0},
				Method:     "spearman",
			},
			wantNumVars: 1,
			checkCorrelation: func(t *testing.T, result *CorrelationResult) {
				// Should have perfect Spearman correlation
				if math.Abs(result.Correlations["var1"][0]-1.0) > 0.001 {
					t.Errorf("var1 PC1 Spearman correlation = %v, want ~1.0", result.Correlations["var1"][0])
				}
			},
			wantErr: false,
		},
		{
			name: "Invalid method",
			request: CorrelationRequest{
				Scores:     scores,
				Components: []int{0},
				Method:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "Nil scores",
			request: CorrelationRequest{
				Scores: nil,
				Method: "pearson",
			},
			wantErr: true,
		},
		{
			name: "Mismatched lengths",
			request: CorrelationRequest{
				Scores: scores,
				MetadataNumeric: map[string][]float64{
					"var1": {1, 2, 3}, // Wrong length
				},
				Method: "pearson",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateEigencorrelations(tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateEigencorrelations() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateEigencorrelations() unexpected error = %v", err)
				return
			}

			// Check number of variables
			if len(result.Variables) != tt.wantNumVars {
				t.Errorf("CalculateEigencorrelations() returned %d variables, want %d", len(result.Variables), tt.wantNumVars)
			}

			// Run specific correlation checks if provided
			if tt.checkCorrelation != nil {
				tt.checkCorrelation(t, result)
			}
		})
	}
}

// TestStatisticalFunctions tests the statistical helper functions
func TestStatisticalFunctions(t *testing.T) {
	t.Run("normalCDF", func(t *testing.T) {
		tests := []struct {
			z    float64
			want float64
			tol  float64
		}{
			{0, 0.5, 0.001},
			{1, 0.8413, 0.001},
			{-1, 0.1587, 0.001},
			{2, 0.9772, 0.001},
			{-2, 0.0228, 0.001},
		}

		for _, tt := range tests {
			got := normalCDF(tt.z)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("normalCDF(%v) = %v, want %v ± %v", tt.z, got, tt.want, tt.tol)
			}
		}
	})

	t.Run("studentTCDF", func(t *testing.T) {
		// Just test that it returns reasonable values
		tests := []struct {
			t  float64
			df float64
		}{
			{0, 10},
			{1, 10},
			{2, 10},
			{0, 50}, // Should use normal approximation
		}

		for _, tt := range tests {
			got := studentTCDF(tt.t, tt.df)
			if got < 0 || got > 1 {
				t.Errorf("studentTCDF(%v, %v) = %v, want value in [0,1]", tt.t, tt.df, got)
			}
		}
	})
}
