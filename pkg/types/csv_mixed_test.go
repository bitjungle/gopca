// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package types

import (
	"math"
	"strings"
	"testing"
)

func TestParseCSVMixed(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		format       CSVFormat
		wantRows     int
		wantCols     int
		wantCatCols  int
		wantHeaders  []string
		wantRowNames []string
		wantFirstRow []float64
		wantCatData  map[string][]string
		wantErr      bool
	}{
		{
			name: "numeric only CSV",
			csvContent: `a,b,c
1.5,2.5,3.5
4.5,5.5,6.5`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA", "N/A", "null"},
			},
			wantRows:     2,
			wantCols:     3,
			wantCatCols:  0,
			wantHeaders:  []string{"a", "b", "c"},
			wantRowNames: []string{},
			wantFirstRow: []float64{1.5, 2.5, 3.5},
			wantCatData:  map[string][]string{},
		},
		{
			name: "mixed numeric and categorical",
			csvContent: `name,age,score,category
John,25,85.5,A
Jane,30,92.0,B
Bob,35,78.5,A`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       []string{"NA", "N/A", "null"},
			},
			wantRows:     3,
			wantCols:     2, // age and score
			wantCatCols:  1, // category
			wantHeaders:  []string{"age", "score"},
			wantRowNames: []string{"John", "Jane", "Bob"},
			wantFirstRow: []float64{25, 85.5},
			wantCatData: map[string][]string{
				"category": {"A", "B", "A"},
			},
		},
		{
			name: "with missing values",
			csvContent: `x,y,z
1.0,NA,3.0
4.0,5.0,N/A
null,8.0,9.0`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA", "N/A", "null"},
			},
			wantRows:     3,
			wantCols:     3,
			wantCatCols:  0,
			wantHeaders:  []string{"x", "y", "z"},
			wantFirstRow: []float64{1.0, math.NaN(), 3.0},
		},
		{
			name: "with special float values",
			csvContent: `a,b,c
1.0,inf,-inf
2.0,3.0,4.0`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantRows:     2,
			wantCols:     3,
			wantHeaders:  []string{"a", "b", "c"},
			wantFirstRow: []float64{1.0, math.Inf(1), math.Inf(-1)},
		},
		{
			name:       "empty CSV",
			csvContent: "",
			format:     DefaultCSVFormat(),
			wantErr:    true,
		},
		{
			name:       "headers only",
			csvContent: "a,b,c",
			format: CSVFormat{
				HasHeaders: true,
			},
			wantErr: true,
		},
		{
			name: "European format with comma decimal",
			csvContent: `val1;val2;category
1,5;2,5;A
3,0;4,0;B`,
			format: CSVFormat{
				FieldDelimiter:   ';',
				DecimalSeparator: ',',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantRows:     2,
			wantCols:     2,
			wantCatCols:  1,
			wantHeaders:  []string{"val1", "val2"},
			wantFirstRow: []float64{1.5, 2.5},
			wantCatData: map[string][]string{
				"category": {"A", "B"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvContent)
			data, catData, err := ParseCSVMixed(reader, tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCSVMixed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if data.Rows != tt.wantRows {
				t.Errorf("rows = %d, want %d", data.Rows, tt.wantRows)
			}

			if data.Columns != tt.wantCols {
				t.Errorf("columns = %d, want %d", data.Columns, tt.wantCols)
			}

			if len(catData) != tt.wantCatCols {
				t.Errorf("categorical columns = %d, want %d", len(catData), tt.wantCatCols)
			}

			// Check headers
			if len(data.Headers) != len(tt.wantHeaders) {
				t.Errorf("headers length = %d, want %d", len(data.Headers), len(tt.wantHeaders))
			} else {
				for i, h := range data.Headers {
					if h != tt.wantHeaders[i] {
						t.Errorf("header[%d] = %s, want %s", i, h, tt.wantHeaders[i])
					}
				}
			}

			// Check row names
			if len(data.RowNames) != len(tt.wantRowNames) {
				t.Errorf("row names length = %d, want %d", len(data.RowNames), len(tt.wantRowNames))
			} else {
				for i, rn := range data.RowNames {
					if rn != tt.wantRowNames[i] {
						t.Errorf("rowName[%d] = %s, want %s", i, rn, tt.wantRowNames[i])
					}
				}
			}

			// Check first row values if expected
			if tt.wantFirstRow != nil && len(data.Matrix) > 0 {
				for i, val := range data.Matrix[0] {
					if i < len(tt.wantFirstRow) {
						expected := tt.wantFirstRow[i]
						if math.IsNaN(expected) {
							if !math.IsNaN(val) {
								t.Errorf("firstRow[%d] = %f, want NaN", i, val)
							}
						} else if val != expected {
							t.Errorf("firstRow[%d] = %f, want %f", i, val, expected)
						}
					}
				}
			}

			// Check categorical data
			if tt.wantCatData != nil {
				for colName, expectedVals := range tt.wantCatData {
					actualVals, exists := catData[colName]
					if !exists {
						t.Errorf("missing categorical column %s", colName)
						continue
					}
					if len(actualVals) != len(expectedVals) {
						t.Errorf("categorical column %s has %d values, want %d",
							colName, len(actualVals), len(expectedVals))
						continue
					}
					for i, val := range actualVals {
						if val != expectedVals[i] {
							t.Errorf("categorical[%s][%d] = %s, want %s",
								colName, i, val, expectedVals[i])
						}
					}
				}
			}
		})
	}
}

func TestIsNumericValue(t *testing.T) {
	format := CSVFormat{
		DecimalSeparator: '.',
		NullValues:       []string{"NA", "N/A", "null"},
	}

	tests := []struct {
		value    string
		wantNum  bool
		wantVal  float64
		checkNaN bool
		checkInf int // 0 = not inf, 1 = +inf, -1 = -inf
	}{
		{"123.45", true, 123.45, false, 0},
		{"NA", true, 0, true, 0},
		{"null", true, 0, true, 0},
		{"inf", true, 0, false, 1},
		{"-inf", true, 0, false, -1},
		{"abc", false, 0, false, 0},
		{"", false, 0, false, 0},
		{"3.14e10", true, 3.14e10, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			isNum, val := isNumericValue(tt.value, format)

			if isNum != tt.wantNum {
				t.Errorf("isNumericValue(%q) isNum = %v, want %v", tt.value, isNum, tt.wantNum)
			}

			if !isNum {
				return
			}

			if tt.checkNaN {
				if !math.IsNaN(val) {
					t.Errorf("isNumericValue(%q) val = %f, want NaN", tt.value, val)
				}
			} else if tt.checkInf != 0 {
				if tt.checkInf > 0 && !math.IsInf(val, 1) {
					t.Errorf("isNumericValue(%q) val = %f, want +Inf", tt.value, val)
				} else if tt.checkInf < 0 && !math.IsInf(val, -1) {
					t.Errorf("isNumericValue(%q) val = %f, want -Inf", tt.value, val)
				}
			} else if val != tt.wantVal {
				t.Errorf("isNumericValue(%q) val = %f, want %f", tt.value, val, tt.wantVal)
			}
		})
	}

	// Test with comma decimal separator
	formatComma := CSVFormat{
		DecimalSeparator: ',',
		NullValues:       []string{"NA"},
	}

	isNum, val := isNumericValue("123,45", formatComma)
	if !isNum || val != 123.45 {
		t.Errorf("isNumericValue(\"123,45\") with comma separator = (%v, %f), want (true, 123.45)", isNum, val)
	}
}
