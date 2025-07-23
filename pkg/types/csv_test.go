package types

import (
	"math"
	"strings"
	"testing"
)

func TestCSVParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		format      CSVFormat
		input       string
		wantRows    int
		wantCols    int
		wantErr     bool
		checkValues func(*testing.T, *CSVData)
	}{
		{
			name: "Basic comma-separated CSV",
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       []string{"", "NA", "NaN"},
			},
			input: `Sample,Feature1,Feature2,Feature3
Row1,1.5,2.3,4.1
Row2,3.2,NA,5.6
Row3,2.1,3.4,NaN`,
			wantRows: 3,
			wantCols: 3,
			checkValues: func(t *testing.T, data *CSVData) {
				// Check headers
				expectedHeaders := []string{"Feature1", "Feature2", "Feature3"}
				for i, h := range expectedHeaders {
					if data.Headers[i] != h {
						t.Errorf("Header[%d] = %s, want %s", i, data.Headers[i], h)
					}
				}

				// Check row names
				expectedRows := []string{"Row1", "Row2", "Row3"}
				for i, r := range expectedRows {
					if data.RowNames[i] != r {
						t.Errorf("RowName[%d] = %s, want %s", i, data.RowNames[i], r)
					}
				}

				// Check values
				if data.Matrix[0][0] != 1.5 {
					t.Errorf("Matrix[0][0] = %f, want 1.5", data.Matrix[0][0])
				}

				// Check NaN handling
				if !math.IsNaN(data.Matrix[1][1]) {
					t.Errorf("Matrix[1][1] should be NaN, got %f", data.Matrix[1][1])
				}
				if !data.MissingMask[1][1] {
					t.Errorf("MissingMask[1][1] should be true")
				}

				if !math.IsNaN(data.Matrix[2][2]) {
					t.Errorf("Matrix[2][2] should be NaN, got %f", data.Matrix[2][2])
				}
			},
		},
		{
			name: "Semicolon-separated with comma decimals",
			format: CSVFormat{
				FieldDelimiter:   ';',
				DecimalSeparator: ',',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       []string{"", "NA", "m"},
			},
			input: `Sample;Var1;Var2;Var3
S1;1,5;2,3;4,1
S2;3,2;m;5,6
S3;2,1;3,4;`,
			wantRows: 3,
			wantCols: 3,
			checkValues: func(t *testing.T, data *CSVData) {
				// Check decimal conversion
				if data.Matrix[0][0] != 1.5 {
					t.Errorf("Matrix[0][0] = %f, want 1.5", data.Matrix[0][0])
				}
				if data.Matrix[0][1] != 2.3 {
					t.Errorf("Matrix[0][1] = %f, want 2.3", data.Matrix[0][1])
				}

				// Check 'm' as missing value
				if !math.IsNaN(data.Matrix[1][1]) {
					t.Errorf("Matrix[1][1] should be NaN (m), got %f", data.Matrix[1][1])
				}

				// Check empty string as missing
				if !math.IsNaN(data.Matrix[2][2]) {
					t.Errorf("Matrix[2][2] should be NaN (empty), got %f", data.Matrix[2][2])
				}
			},
		},
		{
			name: "Tab-separated values",
			format: CSVFormat{
				FieldDelimiter:   '\t',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NULL"},
			},
			input:    "X1\tX2\tX3\n1.0\t2.0\t3.0\n4.0\tNULL\t6.0",
			wantRows: 2,
			wantCols: 3,
			checkValues: func(t *testing.T, data *CSVData) {
				if len(data.RowNames) != 0 {
					t.Errorf("Expected no row names, got %d", len(data.RowNames))
				}

				if !math.IsNaN(data.Matrix[1][1]) {
					t.Errorf("Matrix[1][1] should be NaN (NULL), got %f", data.Matrix[1][1])
				}
			},
		},
		{
			name: "Special float values",
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       false,
				HasRowNames:      false,
				NullValues:       []string{},
			},
			input:    "1.0,Inf,-Inf\n2.0,+Inf,infinity",
			wantRows: 2,
			wantCols: 3,
			checkValues: func(t *testing.T, data *CSVData) {
				if !math.IsInf(data.Matrix[0][1], 1) {
					t.Errorf("Matrix[0][1] should be +Inf, got %f", data.Matrix[0][1])
				}
				if !math.IsInf(data.Matrix[0][2], -1) {
					t.Errorf("Matrix[0][2] should be -Inf, got %f", data.Matrix[0][2])
				}
				if !math.IsInf(data.Matrix[1][1], 1) {
					t.Errorf("Matrix[1][1] should be +Inf, got %f", data.Matrix[1][1])
				}
			},
		},
		{
			name:   "Error: inconsistent columns",
			format: DefaultCSVFormat(),
			input: `A,B,C
1,2,3
4,5`,
			wantErr: true,
		},
		{
			name:    "Error: no data",
			format:  DefaultCSVFormat(),
			input:   ``,
			wantErr: true,
		},
		{
			name:   "Error: invalid number",
			format: DefaultCSVFormat(),
			input: `A,B
1,abc
2,3`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser(tt.format)
			data, err := parser.Parse(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if data.Rows != tt.wantRows {
				t.Errorf("Rows = %d, want %d", data.Rows, tt.wantRows)
			}

			if data.Columns != tt.wantCols {
				t.Errorf("Columns = %d, want %d", data.Columns, tt.wantCols)
			}

			if tt.checkValues != nil {
				tt.checkValues(t, data)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name           string
		sample         string
		wantDelimiter  rune
		wantDecimalSep rune
		wantHasHeaders bool
	}{
		{
			name: "Comma-separated with headers",
			sample: `Name,Value1,Value2
John,123.45,678.90
Jane,234.56,789.01`,
			wantDelimiter:  ',',
			wantDecimalSep: '.',
			wantHasHeaders: true,
		},
		{
			name: "Semicolon with comma decimals",
			sample: `Name;Value1;Value2
John;123,45;678,90
Jane;234,56;789,01`,
			wantDelimiter:  ';',
			wantDecimalSep: ',',
			wantHasHeaders: true,
		},
		{
			name:           "Tab-separated",
			sample:         "A\tB\tC\n1\t2\t3\n4\t5\t6",
			wantDelimiter:  '\t',
			wantDecimalSep: '.',
			wantHasHeaders: true,
		},
		{
			name: "No headers (all numeric)",
			sample: `1.0,2.0,3.0
4.0,5.0,6.0
7.0,8.0,9.0`,
			wantDelimiter:  ',',
			wantDecimalSep: '.',
			wantHasHeaders: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := DetectFormat([]byte(tt.sample))
			if err != nil {
				t.Errorf("DetectFormat() error = %v", err)
				return
			}

			if format.FieldDelimiter != tt.wantDelimiter {
				t.Errorf("FieldDelimiter = %c, want %c", format.FieldDelimiter, tt.wantDelimiter)
			}

			if format.DecimalSeparator != tt.wantDecimalSep {
				t.Errorf("DecimalSeparator = %c, want %c", format.DecimalSeparator, tt.wantDecimalSep)
			}

			if format.HasHeaders != tt.wantHasHeaders {
				t.Errorf("HasHeaders = %v, want %v", format.HasHeaders, tt.wantHasHeaders)
			}
		})
	}
}

func TestGetMissingValueInfo(t *testing.T) {
	// Create test data with known missing values
	data := &CSVData{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0, 4.0},
			{5.0, 6.0, math.NaN(), 8.0},
			{9.0, math.NaN(), 11.0, 12.0},
			{13.0, 14.0, 15.0, 16.0},
		},
		MissingMask: [][]bool{
			{false, true, false, false},
			{false, false, true, false},
			{false, true, false, false},
			{false, false, false, false},
		},
		Rows:    4,
		Columns: 4,
	}

	tests := []struct {
		name                   string
		selectedColumns        []int
		wantMissing            bool
		wantTotalMissing       int
		wantRowsAffected       int
		wantColumnsWithMissing int
	}{
		{
			name:                   "All columns",
			selectedColumns:        []int{0, 1, 2, 3},
			wantMissing:            true,
			wantTotalMissing:       3,
			wantRowsAffected:       3,
			wantColumnsWithMissing: 2,
		},
		{
			name:                   "Only columns with missing",
			selectedColumns:        []int{1, 2},
			wantMissing:            true,
			wantTotalMissing:       3,
			wantRowsAffected:       3,
			wantColumnsWithMissing: 2,
		},
		{
			name:                   "Only columns without missing",
			selectedColumns:        []int{0, 3},
			wantMissing:            false,
			wantTotalMissing:       0,
			wantRowsAffected:       0,
			wantColumnsWithMissing: 0,
		},
		{
			name:                   "Single column with missing",
			selectedColumns:        []int{1},
			wantMissing:            true,
			wantTotalMissing:       2,
			wantRowsAffected:       2,
			wantColumnsWithMissing: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := data.GetMissingValueInfo(tt.selectedColumns)

			if info.HasMissing() != tt.wantMissing {
				t.Errorf("HasMissing() = %v, want %v", info.HasMissing(), tt.wantMissing)
			}

			if info.TotalMissing != tt.wantTotalMissing {
				t.Errorf("TotalMissing = %d, want %d", info.TotalMissing, tt.wantTotalMissing)
			}

			if len(info.RowsAffected) != tt.wantRowsAffected {
				t.Errorf("RowsAffected count = %d, want %d", len(info.RowsAffected), tt.wantRowsAffected)
			}

			if len(info.ColumnIndices) != tt.wantColumnsWithMissing {
				t.Errorf("ColumnIndices count = %d, want %d", len(info.ColumnIndices), tt.wantColumnsWithMissing)
			}
		})
	}
}

func TestMissingValueInfo_GetSummary(t *testing.T) {
	tests := []struct {
		name string
		info MissingValueInfo
		want string
	}{
		{
			name: "No missing values",
			info: MissingValueInfo{
				TotalMissing: 0,
			},
			want: "No missing values in selected columns",
		},
		{
			name: "Some missing values",
			info: MissingValueInfo{
				TotalMissing:  5,
				RowsAffected:  []int{1, 3, 5},
				ColumnIndices: []int{2, 4},
			},
			want: "5 missing values found, affecting 3 rows, in 2 columns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.GetSummary()
			if got != tt.want {
				t.Errorf("GetSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
