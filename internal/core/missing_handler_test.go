package core

import (
	"math"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestMissingValueHandler_HandleMissingValues(t *testing.T) {
	// Create test data with known missing values
	// Row 0: col 1 is NaN
	// Row 1: col 2 is NaN  
	// Row 2: col 1 is NaN
	// Row 3: no NaN
	data := types.Matrix{
		{1.0, math.NaN(), 3.0, 4.0},
		{5.0, 6.0, math.NaN(), 8.0},
		{9.0, math.NaN(), 11.0, 12.0},
		{13.0, 14.0, 15.0, 16.0},
	}

	// Missing info for columns 1 and 2
	missingInfo := &types.MissingValueInfo{
		ColumnIndices:   []int{1, 2},
		RowsAffected:    []int{0, 1, 2},
		TotalMissing:    3,
		MissingByColumn: map[int]int{1: 2, 2: 1},
	}

	tests := []struct {
		name         string
		strategy     types.MissingValueStrategy
		selectedCols []int
		wantRows     int
		checkResult  func(*testing.T, types.Matrix)
		wantErr      bool
	}{
		{
			name:         "Drop rows with missing values",
			strategy:     types.MissingDrop,
			selectedCols: []int{0, 1, 2, 3},
			wantRows:     1, // Only row 3 has no missing values
			checkResult: func(t *testing.T, result types.Matrix) {
				if result[0][0] != 13.0 {
					t.Errorf("Expected first row to be the original 4th row, got %f", result[0][0])
				}
			},
		},
		{
			name:         "Impute with mean",
			strategy:     types.MissingMean,
			selectedCols: []int{0, 1, 2, 3},
			wantRows:     4,
			checkResult: func(t *testing.T, result types.Matrix) {
				// Column 1 mean: (6 + 14) / 2 = 10
				if result[0][1] != 10.0 {
					t.Errorf("Expected mean imputation for [0][1], got %f", result[0][1])
				}
				if result[2][1] != 10.0 {
					t.Errorf("Expected mean imputation for [2][1], got %f", result[2][1])
				}
				
				// Column 2 mean: (3 + 11 + 15) / 3 = 9.67
				expectedMean := (3.0 + 11.0 + 15.0) / 3.0
				if math.Abs(result[1][2]-expectedMean) > 0.01 {
					t.Errorf("Expected mean imputation for [1][2], got %f, want %f", result[1][2], expectedMean)
				}
			},
		},
		{
			name:         "Impute with median",
			strategy:     types.MissingMedian,
			selectedCols: []int{0, 1, 2, 3},
			wantRows:     4,
			checkResult: func(t *testing.T, result types.Matrix) {
				// For median imputation test:
				// Original data has these values in column 1: [NaN, 6, NaN, 14]
				// Non-missing values: [6, 14], median = 10
				// So positions [0][1] and [2][1] should be 10.0
				
				// Debug output
				t.Logf("Result matrix shape: %d x %d", len(result), len(result[0]))
				for i := 0; i < len(result); i++ {
					t.Logf("Row %d col 1: %f (isNaN: %v)", i, result[i][1], math.IsNaN(result[i][1]))
				}
				
				// Check if NaN values were properly imputed
				if math.IsNaN(result[0][1]) {
					t.Errorf("Result[0][1] is still NaN, should be imputed with median 10.0")
				} else if result[0][1] != 10.0 {
					t.Errorf("Expected median imputation for [0][1], got %f, want 10.0", result[0][1])
				}
				
				if math.IsNaN(result[2][1]) {
					t.Errorf("Result[2][1] is still NaN, should be imputed with median 10.0")
				} else if result[2][1] != 10.0 {
					t.Errorf("Expected median imputation for [2][1], got %f, want 10.0", result[2][1])
				}
				
				// Original data has these values in column 2: [3, NaN, 11, 15]
				// Non-missing values: [3, 11, 15], median = 11
				if math.IsNaN(result[1][2]) {
					t.Errorf("Result[1][2] is still NaN, should be imputed with median 11.0")
				} else if result[1][2] != 11.0 {
					t.Errorf("Expected median imputation for [1][2], got %f, want 11.0", result[1][2])
				}
			},
		},
		{
			name:         "No missing values in selected columns",
			strategy:     types.MissingDrop,
			selectedCols: []int{0, 3}, // Columns without missing values
			wantRows:     4,
			checkResult: func(t *testing.T, result types.Matrix) {
				// Data should be unchanged
				if result[0][0] != 1.0 || result[3][3] != 16.0 {
					t.Errorf("Data should be unchanged when no missing values")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMissingValueHandler(tt.strategy)
			
			// Create a copy of data to avoid modifying the original
			dataCopy := make(types.Matrix, len(data))
			for i := range data {
				dataCopy[i] = make([]float64, len(data[i]))
				copy(dataCopy[i], data[i])
			}
			
			// Verify the test data is set up correctly
			if tt.name == "Impute with median" {
				if !math.IsNaN(dataCopy[0][1]) {
					t.Fatalf("Test data setup error: dataCopy[0][1] should be NaN, got %f", dataCopy[0][1])
				}
				if !math.IsNaN(dataCopy[2][1]) {
					t.Fatalf("Test data setup error: dataCopy[2][1] should be NaN, got %f", dataCopy[2][1])
				}
				if !math.IsNaN(dataCopy[1][2]) {
					t.Fatalf("Test data setup error: dataCopy[1][2] should be NaN, got %f", dataCopy[1][2])
				}
				
				// Debug: print original data
				t.Logf("Original dataCopy column 1 values:")
				for i := 0; i < len(dataCopy); i++ {
					t.Logf("  Row %d: %v (isNaN: %v)", i, dataCopy[i][1], math.IsNaN(dataCopy[i][1]))
				}
			}
			
			// For the "no missing values" test, create appropriate missing info
			actualMissingInfo := missingInfo
			if tt.name == "No missing values in selected columns" {
				actualMissingInfo = &types.MissingValueInfo{
					ColumnIndices:   []int{},
					RowsAffected:    []int{},
					TotalMissing:    0,
					MissingByColumn: map[int]int{},
				}
			}
			
			result, err := handler.HandleMissingValues(dataCopy, actualMissingInfo, tt.selectedCols)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleMissingValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != tt.wantRows {
				t.Errorf("Result rows = %d, want %d", len(result), tt.wantRows)
			}
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestValidateDataForPCA(t *testing.T) {
	tests := []struct {
		name         string
		data         types.Matrix
		selectedCols []int
		wantErr      bool
		errContains  string
	}{
		{
			name: "Valid data",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			selectedCols: []int{0, 1, 2},
			wantErr:      false,
		},
		{
			name:         "Empty data",
			data:         types.Matrix{},
			selectedCols: []int{0},
			wantErr:      true,
			errContains:  "no data rows",
		},
		{
			name: "Insufficient samples",
			data: types.Matrix{
				{1.0, 2.0, 3.0},
			},
			selectedCols: []int{0, 1, 2},
			wantErr:      true,
			errContains:  "insufficient samples",
		},
		{
			name: "Column with all NaN",
			data: types.Matrix{
				{1.0, math.NaN(), 3.0},
				{4.0, math.NaN(), 6.0},
				{7.0, math.NaN(), 9.0},
			},
			selectedCols: []int{0, 1, 2},
			wantErr:      true,
			errContains:  "column 1 contains only missing values",
		},
		{
			name: "Valid with some NaN",
			data: types.Matrix{
				{1.0, math.NaN(), 3.0},
				{4.0, 5.0, 6.0},
				{7.0, 8.0, 9.0},
			},
			selectedCols: []int{0, 1, 2},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDataForPCA(tt.data, tt.selectedCols)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataForPCA() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestCalculateColumnStatistics(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, math.NaN(), 6.0},
		{7.0, 8.0, 9.0},
		{10.0, math.NaN(), 12.0},
	}

	handler := &MissingValueHandler{}

	t.Run("Calculate mean", func(t *testing.T) {
		stats := handler.calculateColumnStatistics(data, []int{0, 1, 2}, true)
		
		// Column 0: mean of [1, 4, 7, 10] = 5.5
		if stats[0] != 5.5 {
			t.Errorf("Column 0 mean = %f, want 5.5", stats[0])
		}
		
		// Column 1: mean of [2, 8] = 5.0
		if stats[1] != 5.0 {
			t.Errorf("Column 1 mean = %f, want 5.0", stats[1])
		}
		
		// Column 2: mean of [3, 6, 9, 12] = 7.5
		if stats[2] != 7.5 {
			t.Errorf("Column 2 mean = %f, want 7.5", stats[2])
		}
	})

	t.Run("Calculate median", func(t *testing.T) {
		stats := handler.calculateColumnStatistics(data, []int{0, 1, 2}, false)
		
		// Column 0: median of [1, 4, 7, 10] = 5.5
		if stats[0] != 5.5 {
			t.Errorf("Column 0 median = %f, want 5.5", stats[0])
		}
		
		// Column 1: median of [2, 8] = 5.0
		if stats[1] != 5.0 {
			t.Errorf("Column 1 median = %f, want 5.0", stats[1])
		}
		
		// Column 2: median of [3, 6, 9, 12] = 7.5
		if stats[2] != 7.5 {
			t.Errorf("Column 2 median = %f, want 7.5", stats[2])
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}