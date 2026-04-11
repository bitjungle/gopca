// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestNewValidator(t *testing.T) {
	opts := DefaultOptions()
	validator := NewValidator(opts)

	if validator == nil {
		t.Fatal("expected non-nil validator")
		return
	}

	if validator.opts.Delimiter != opts.Delimiter {
		t.Errorf("expected delimiter %v, got %v", opts.Delimiter, validator.opts.Delimiter)
	}
}

func TestValidator_Validate_NilData(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	result := validator.Validate(nil)

	if result.Valid {
		t.Error("expected invalid result for nil data")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for nil data")
	}
}

func TestValidator_Validate_NoData(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix:     nil,
		StringData: nil,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for data with no matrix or string data")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for data with no content")
	}
}

func TestValidator_Validate_EmptyMatrix(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix:  [][]float64{},
		Rows:    0,
		Columns: 0,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for empty matrix")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for empty matrix")
	}
}

func TestValidator_Validate_ValidNumericData(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
			{7.0, 8.0, 9.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    3,
		Columns: 3,
	}
	result := validator.Validate(data)

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}

	if len(result.ColumnStats) != 3 {
		t.Errorf("expected 3 column stats, got %d", len(result.ColumnStats))
	}

	if result.ColumnStats[0].Mean != 4.0 {
		t.Errorf("expected mean 4.0 for column 0, got %f", result.ColumnStats[0].Mean)
	}
}

func TestValidator_Validate_InconsistentDimensions(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
		Rows:    3,
		Columns: 3,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for dimension mismatch")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for dimension mismatch")
	}
}

func TestValidator_Validate_InconsistentColumns(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0},
		},
		Rows:    2,
		Columns: 3,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for inconsistent columns")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for inconsistent columns")
	}
}

func TestValidator_Validate_AllMissingColumn(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0},
			{4.0, math.NaN(), 6.0},
		},
		Headers:     []string{"A", "B", "C"},
		Rows:        2,
		Columns:     3,
		MissingMask: [][]bool{{false, true, false}, {false, true, false}},
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for all-missing column")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for all-missing column")
	}
}

func TestValidator_Validate_ZeroVarianceWarning(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, 5.0, 3.0},
			{1.0, 5.0, 6.0},
			{1.0, 5.0, 9.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    3,
		Columns: 3,
	}
	result := validator.Validate(data)

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warnings for zero variance columns")
	}

	if !result.ColumnStats[0].HasZeroVariance {
		t.Error("expected column 0 to have zero variance")
	}
}

func TestValidator_Validate_HighMissingPercentage(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0},
			{4.0, math.NaN(), 6.0},
			{7.0, 8.0, 9.0},
		},
		Headers:     []string{"A", "B", "C"},
		Rows:        3,
		Columns:     3,
		MissingMask: [][]bool{{false, true, false}, {false, true, false}, {false, false, false}},
	}
	result := validator.Validate(data)

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warnings for high missing percentage")
	}

	if result.ColumnStats[1].MissingPercent < 50.0 {
		t.Errorf("expected >50%% missing, got %.1f%%", result.ColumnStats[1].MissingPercent)
	}
}

func TestValidator_Validate_HeaderMismatch(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix:  [][]float64{{1.0, 2.0, 3.0}},
		Headers: []string{"A", "B"},
		Rows:    1,
		Columns: 3,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for header count mismatch")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for header count mismatch")
	}
}

func TestValidator_Validate_RowNamesMismatch(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		Matrix:   [][]float64{{1.0, 2.0}, {3.0, 4.0}},
		RowNames: []string{"row1"},
		Rows:     2,
		Columns:  2,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for row names count mismatch")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for row names count mismatch")
	}
}

func TestValidator_Validate_StringData(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		StringData: [][]string{
			{"a", "b", "c"},
			{"d", "e", "f"},
		},
		Rows:    2,
		Columns: 3,
	}
	result := validator.Validate(data)

	if !result.Valid {
		t.Errorf("expected valid result for string data, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_EmptyStringData(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		StringData: [][]string{},
		Rows:       0,
		Columns:    0,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for empty string data")
	}
}

func TestValidator_Validate_StringDataInconsistentColumns(t *testing.T) {
	validator := NewValidator(DefaultOptions())
	data := &Data{
		StringData: [][]string{
			{"a", "b", "c"},
			{"d", "e"},
		},
		Rows:    2,
		Columns: 3,
	}
	result := validator.Validate(data)

	if result.Valid {
		t.Error("expected invalid result for inconsistent string columns")
	}
}

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	content := `A,B,C
1,2,3
4,5,6`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := DefaultOptions()
	opts.HasRowNames = false
	result, err := ValidateFile(testFile, opts)

	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateFile_NonExistent(t *testing.T) {
	_, err := ValidateFile("/nonexistent/file.csv", DefaultOptions())

	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidateStructure(t *testing.T) {
	tests := []struct {
		name    string
		data    *Data
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "no data present",
			data: &Data{
				Matrix:     nil,
				StringData: nil,
			},
			wantErr: true,
		},
		{
			name: "empty matrix",
			data: &Data{
				Matrix: [][]float64{},
			},
			wantErr: true,
		},
		{
			name: "valid data",
			data: &Data{
				Matrix: [][]float64{{1, 2}, {3, 4}},
			},
			wantErr: false,
		},
		{
			name: "inconsistent columns",
			data: &Data{
				Matrix: [][]float64{{1, 2, 3}, {4, 5}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStructure(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnalyzeMissingValues(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0},
			{4.0, 5.0, math.NaN()},
			{7.0, 8.0, 9.0},
		},
		Rows:        3,
		Columns:     3,
		MissingMask: [][]bool{{false, true, false}, {false, false, true}, {false, false, false}},
	}

	analysis := AnalyzeMissingValues(data)

	totalCells := analysis["total_cells"].(int)
	if totalCells != 9 {
		t.Errorf("expected 9 total cells, got %d", totalCells)
	}

	missingCells := analysis["missing_cells"].(int)
	if missingCells != 2 {
		t.Errorf("expected 2 missing cells, got %d", missingCells)
	}

	missingPct := analysis["missing_percentage"].(float64)
	expectedPct := 2.0 / 9.0 * 100
	if math.Abs(missingPct-expectedPct) > 0.01 {
		t.Errorf("expected %.2f%% missing, got %.2f%%", expectedPct, missingPct)
	}

	columnMissing := analysis["missing_by_column"].([]int)
	if len(columnMissing) != 3 {
		t.Errorf("expected 3 columns, got %d", len(columnMissing))
	}

	if columnMissing[1] != 1 {
		t.Errorf("expected 1 missing in column 1, got %d", columnMissing[1])
	}

	rowMissing := analysis["missing_by_row"].([]int)
	if len(rowMissing) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rowMissing))
	}

	if rowMissing[0] != 1 {
		t.Errorf("expected 1 missing in row 0, got %d", rowMissing[0])
	}
}

func TestAnalyzeMissingValues_NilData(t *testing.T) {
	analysis := AnalyzeMissingValues(nil)

	if len(analysis) != 0 {
		t.Error("expected empty analysis for nil data")
	}
}

func TestData_GetMissingValueInfo(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0, 4.0},
			{5.0, 6.0, math.NaN(), 8.0},
			{9.0, math.NaN(), 11.0, 12.0},
		},
		Rows:        3,
		Columns:     4,
		MissingMask: [][]bool{{false, true, false, false}, {false, false, true, false}, {false, true, false, false}},
	}

	info := data.GetMissingValueInfo([]int{1, 2})

	if info.TotalMissing != 3 {
		t.Errorf("expected 3 total missing, got %d", info.TotalMissing)
	}

	if len(info.ColumnIndices) != 2 {
		t.Errorf("expected 2 columns with missing values, got %d", len(info.ColumnIndices))
	}

	if info.MissingByColumn[1] != 2 {
		t.Errorf("expected 2 missing in column 1, got %d", info.MissingByColumn[1])
	}

	if info.MissingByColumn[2] != 1 {
		t.Errorf("expected 1 missing in column 2, got %d", info.MissingByColumn[2])
	}

	if len(info.RowsAffected) != 3 {
		t.Errorf("expected 3 rows affected, got %d", len(info.RowsAffected))
	}
}

func TestData_GetMissingValueInfo_AllColumns(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN()},
			{3.0, 4.0},
		},
		Rows:        2,
		Columns:     2,
		MissingMask: [][]bool{{false, true}, {false, false}},
	}

	info := data.GetMissingValueInfo([]int{})

	if info.TotalMissing != 1 {
		t.Errorf("expected 1 total missing, got %d", info.TotalMissing)
	}

	if len(info.ColumnIndices) != 1 {
		t.Errorf("expected 1 column with missing values, got %d", len(info.ColumnIndices))
	}
}

func TestData_GetMissingValueInfo_InvalidColumnIndex(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
		Rows:        2,
		Columns:     2,
		MissingMask: [][]bool{{false, false}, {false, false}},
	}

	info := data.GetMissingValueInfo([]int{-1, 5})

	if info.TotalMissing != 0 {
		t.Errorf("expected 0 total missing for invalid columns, got %d", info.TotalMissing)
	}
}
