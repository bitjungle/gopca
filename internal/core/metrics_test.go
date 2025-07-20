package core

import (
	"math"
	"testing"

	"github.com/bitjungle/complab/pkg/types"
)

func TestCalculateMahalanobisDistances(t *testing.T) {
	tests := []struct {
		name    string
		data    [][]float64
		wantErr bool
	}{
		{
			name: "2D data",
			data: [][]float64{
				{1.0, 2.5},
				{2.0, 3.2},
				{3.0, 3.8},
				{4.0, 5.1},
			},
			wantErr: false,
		},
		{
			name: "1D data",
			data: [][]float64{
				{1.0},
				{2.0},
				{3.0},
				{4.0},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := extractScoresMatrix(tt.data, len(tt.data[0]))
			distances, err := calculateMahalanobisDistances(scores)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateMahalanobisDistances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Check that distances are non-negative
				for i, d := range distances {
					if d < 0 || math.IsNaN(d) || math.IsInf(d, 0) {
						t.Errorf("Invalid distance at index %d: %v", i, d)
					}
				}
				
				// Check correct number of distances
				if len(distances) != len(tt.data) {
					t.Errorf("Expected %d distances, got %d", len(tt.data), len(distances))
				}
			}
		})
	}
}

func TestCalculateHotellingT2(t *testing.T) {
	tests := []struct {
		name     string
		data     [][]float64
		nSamples int
		wantErr  bool
	}{
		{
			name: "2D data",
			data: [][]float64{
				{1.0, 2.5},
				{2.0, 3.2},
				{3.0, 3.8},
				{4.0, 5.1},
			},
			nSamples: 4,
			wantErr:  false,
		},
		{
			name: "1D data",
			data: [][]float64{
				{1.0},
				{2.0},
				{3.0},
				{4.0},
			},
			nSamples: 4,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := extractScoresMatrix(tt.data, len(tt.data[0]))
			t2Stats, err := calculateHotellingT2(scores, tt.nSamples)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateHotellingT2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Check that T² statistics are non-negative
				for i, t2 := range t2Stats {
					if t2 < 0 || math.IsNaN(t2) || math.IsInf(t2, 0) {
						t.Errorf("Invalid T² statistic at index %d: %v", i, t2)
					}
				}
				
				// Check correct number of statistics
				if len(t2Stats) != len(tt.data) {
					t.Errorf("Expected %d T² statistics, got %d", len(tt.data), len(t2Stats))
				}
			}
		})
	}
}

func TestDetectOutliersFromT2(t *testing.T) {
	tests := []struct {
		name         string
		t2Stats      []float64
		nSamples     int
		nComponents  int
		significance float64
		expectedSum  int // Expected number of outliers (approximate)
	}{
		{
			name:         "No outliers",
			t2Stats:      []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			nSamples:     5,
			nComponents:  2,
			significance: 0.01,
			expectedSum:  0,
		},
		{
			name:         "Some outliers",
			t2Stats:      []float64{0.1, 0.2, 50.0, 0.4, 100.0},
			nSamples:     20,
			nComponents:  2,
			significance: 0.01,
			expectedSum:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outliers := detectOutliersFromT2(tt.t2Stats, tt.nSamples, tt.nComponents, tt.significance)
			
			// Check correct length
			if len(outliers) != len(tt.t2Stats) {
				t.Errorf("Expected %d outlier flags, got %d", len(tt.t2Stats), len(outliers))
			}
			
			// Count outliers
			sum := 0
			for _, o := range outliers {
				if o {
					sum++
				}
			}
			
			// For "Some outliers" test, check that we detected at least some
			if tt.expectedSum > 0 && sum == 0 {
				t.Errorf("Expected to detect outliers but found none")
			}
		})
	}
}

func TestNewMetricsCalculator(t *testing.T) {
	mc := NewMetricsCalculator()
	if mc == nil {
		t.Error("NewMetricsCalculator() returned nil")
	}
}

func TestCalculateMetrics(t *testing.T) {
	mc := NewMetricsCalculator()
	
	// Create a simple PCA result
	result := &types.PCAResult{
		Scores: [][]float64{
			{-2.0, 0.5},
			{-1.0, -0.5},
			{1.0, -0.5},
			{2.0, 0.5},
		},
		Loadings: [][]float64{
			{0.7, 0.7},
			{-0.7, 0.7},
		},
		ExplainedVar: []float64{0.8, 0.2},
	}
	
	// Create simple data
	data := types.Matrix{
		{4.0, 4.0},
		{4.5, 5.5},
		{5.5, 4.5},
		{6.0, 6.0},
	}
	
	config := types.MetricsConfig{
		NumComponents:             2,
		SignificanceLevel:        0.01,
		CalculateContributions:   true,
		CalculateConfidenceEllipse: true,
	}
	
	metrics, err := mc.CalculateMetrics(result, data, config)
	if err != nil {
		t.Fatalf("CalculateMetrics() error = %v", err)
	}
	
	// Verify all fields are populated
	if len(metrics.MahalanobisDistances) != len(data) {
		t.Errorf("Expected %d Mahalanobis distances, got %d", len(data), len(metrics.MahalanobisDistances))
	}
	
	if len(metrics.HotellingT2) != len(data) {
		t.Errorf("Expected %d Hotelling T² values, got %d", len(data), len(metrics.HotellingT2))
	}
	
	if len(metrics.RSS) != len(data) {
		t.Errorf("Expected %d RSS values, got %d", len(data), len(metrics.RSS))
	}
	
	if len(metrics.OutlierMask) != len(data) {
		t.Errorf("Expected %d outlier flags, got %d", len(data), len(metrics.OutlierMask))
	}
	
	if metrics.ContributionScores == nil {
		t.Error("Expected contribution scores to be calculated")
	}
}

func TestCalculateContributions(t *testing.T) {
	mc := NewMetricsCalculator()
	
	result := &types.PCAResult{
		Loadings: [][]float64{
			{0.6, 0.8},  // First PC
			{-0.8, 0.6}, // Second PC
		},
	}
	
	data := types.Matrix{{1, 2}, {3, 4}} // Data not used in contribution calculation
	
	contributions := mc.CalculateContributions(result, data)
	
	if contributions == nil {
		t.Fatal("CalculateContributions() returned nil")
	}
	
	// Check dimensions
	if len(contributions) != 2 { // 2 variables
		t.Errorf("Expected 2 variables, got %d", len(contributions))
	}
	
	for i, contrib := range contributions {
		if len(contrib) != 2 { // 2 components
			t.Errorf("Expected 2 components for variable %d, got %d", i, len(contrib))
		}
		
		// Check that contributions sum to 1 for each component
		for j := range contrib {
			sum := 0.0
			for k := range contributions {
				sum += contributions[k][j]
			}
			if math.Abs(sum-1.0) > 1e-10 {
				t.Errorf("Contributions for component %d don't sum to 1: %v", j, sum)
			}
		}
	}
}