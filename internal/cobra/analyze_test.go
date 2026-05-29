// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package cobra

import (
	"reflect"
	"testing"
)

func TestParseExcludeIndices(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []int{},
		},
		{
			name:     "single index",
			input:    "5",
			expected: []int{4}, // 0-based
		},
		{
			name:     "multiple indices",
			input:    "1,3,5",
			expected: []int{0, 2, 4}, // 0-based
		},
		{
			name:     "single range",
			input:    "1-5",
			expected: []int{0, 1, 2, 3, 4}, // 0-based
		},
		{
			name:     "multiple ranges",
			input:    "1-3,8-10",
			expected: []int{0, 1, 2, 7, 8, 9}, // 0-based
		},
		{
			name:     "mixed indices and ranges",
			input:    "1,3-5,8,10-12",
			expected: []int{0, 2, 3, 4, 7, 9, 10, 11}, // 0-based
		},
		{
			name:     "overlapping ranges",
			input:    "1-5,3-7",
			expected: []int{0, 1, 2, 3, 4, 5, 6}, // 0-based, deduplicated
		},
		{
			name:     "duplicate indices",
			input:    "1,2,2,3,1",
			expected: []int{0, 1, 2}, // 0-based, deduplicated
		},
		{
			name:     "with spaces",
			input:    " 1 - 3 , 5 , 7 - 9 ",
			expected: []int{0, 1, 2, 4, 6, 7, 8}, // 0-based
		},
		{
			name:     "single element range",
			input:    "5-5",
			expected: []int{4}, // 0-based
		},
		{
			name:     "unsorted input",
			input:    "10,5,1-3",
			expected: []int{0, 1, 2, 4, 9}, // 0-based, sorted
		},
		{
			name:     "invalid indices ignored",
			input:    "0,1,2,-5",
			expected: []int{0, 1}, // 0 and negative indices ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExcludeIndices(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseExcludeIndices(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestDataFiltering tests that the actual data filtering works correctly
// This is a regression test for issue #418
func TestDataFiltering(t *testing.T) {
	// Create mock CSV data
	mockData := &CSVData{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0, 4.0},
			{5.0, 6.0, 7.0, 8.0},
			{9.0, 10.0, 11.0, 12.0},
			{13.0, 14.0, 15.0, 16.0},
			{17.0, 18.0, 19.0, 20.0},
		},
		Headers:  []string{"col1", "col2", "col3", "col4"},
		RowNames: []string{"row1", "row2", "row3", "row4", "row5"},
		Rows:     5,
		Columns:  4,
	}

	testCases := []struct {
		name             string
		excludeRows      []int
		excludeCols      []int
		expectedRows     int
		expectedCols     int
		expectedFirstVal float64
		expectedRowNames []string
		expectedHeaders  []string
	}{
		{
			name:             "No exclusions",
			excludeRows:      []int{},
			excludeCols:      []int{},
			expectedRows:     5,
			expectedCols:     4,
			expectedFirstVal: 1.0,
			expectedRowNames: []string{"row1", "row2", "row3", "row4", "row5"},
			expectedHeaders:  []string{"col1", "col2", "col3", "col4"},
		},
		{
			name:             "Exclude first row",
			excludeRows:      []int{0},
			excludeCols:      []int{},
			expectedRows:     4,
			expectedCols:     4,
			expectedFirstVal: 5.0,
			expectedRowNames: []string{"row2", "row3", "row4", "row5"},
			expectedHeaders:  []string{"col1", "col2", "col3", "col4"},
		},
		{
			name:             "Exclude first column",
			excludeRows:      []int{},
			excludeCols:      []int{0},
			expectedRows:     5,
			expectedCols:     3,
			expectedFirstVal: 2.0,
			expectedRowNames: []string{"row1", "row2", "row3", "row4", "row5"},
			expectedHeaders:  []string{"col2", "col3", "col4"},
		},
		{
			name:             "Exclude multiple rows and columns",
			excludeRows:      []int{0, 2, 4}, // rows 1, 3, 5
			excludeCols:      []int{1, 3},    // cols 2, 4
			expectedRows:     2,
			expectedCols:     2,
			expectedFirstVal: 5.0, // row2, col1
			expectedRowNames: []string{"row2", "row4"},
			expectedHeaders:  []string{"col1", "col3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy of the data
			data := &CSVData{
				Matrix:   make([][]float64, len(mockData.Matrix)),
				Headers:  append([]string{}, mockData.Headers...),
				RowNames: append([]string{}, mockData.RowNames...),
				Rows:     mockData.Rows,
				Columns:  mockData.Columns,
			}
			for i := range mockData.Matrix {
				data.Matrix[i] = append([]float64{}, mockData.Matrix[i]...)
			}

			// Apply the filtering logic (extracted from runAnalyze)
			if len(tc.excludeRows) > 0 || len(tc.excludeCols) > 0 {
				excludedRowMap := make(map[int]bool)
				for _, row := range tc.excludeRows {
					excludedRowMap[row] = true
				}

				excludedColMap := make(map[int]bool)
				for _, col := range tc.excludeCols {
					excludedColMap[col] = true
				}

				filteredMatrix := make([][]float64, 0)
				filteredRowNames := make([]string, 0)
				for i, row := range data.Matrix {
					if !excludedRowMap[i] {
						filteredRow := make([]float64, 0)
						for j, val := range row {
							if !excludedColMap[j] {
								filteredRow = append(filteredRow, val)
							}
						}
						filteredMatrix = append(filteredMatrix, filteredRow)
						if len(data.RowNames) > i {
							filteredRowNames = append(filteredRowNames, data.RowNames[i])
						}
					}
				}

				filteredHeaders := make([]string, 0)
				for i, header := range data.Headers {
					if !excludedColMap[i] {
						filteredHeaders = append(filteredHeaders, header)
					}
				}

				data.Matrix = filteredMatrix
				data.Rows = len(filteredMatrix)
				data.Columns = len(filteredHeaders)
				data.Headers = filteredHeaders
				data.RowNames = filteredRowNames
			}

			// Validate results
			if data.Rows != tc.expectedRows {
				t.Errorf("Expected %d rows, got %d", tc.expectedRows, data.Rows)
			}
			if data.Columns != tc.expectedCols {
				t.Errorf("Expected %d columns, got %d", tc.expectedCols, data.Columns)
			}
			if len(data.Matrix) > 0 && len(data.Matrix[0]) > 0 {
				if data.Matrix[0][0] != tc.expectedFirstVal {
					t.Errorf("Expected first value %f, got %f", tc.expectedFirstVal, data.Matrix[0][0])
				}
			}
			if !reflect.DeepEqual(data.RowNames, tc.expectedRowNames) {
				t.Errorf("Expected row names %v, got %v", tc.expectedRowNames, data.RowNames)
			}
			if !reflect.DeepEqual(data.Headers, tc.expectedHeaders) {
				t.Errorf("Expected headers %v, got %v", tc.expectedHeaders, data.Headers)
			}
		})
	}
}

// CSVData is a mock structure for testing
type CSVData struct {
	Matrix   [][]float64
	Headers  []string
	RowNames []string
	Rows     int
	Columns  int
}

func TestParseExcludeColumns(t *testing.T) {
	headers := []string{"col1", "col2", "col3", "col4", "col5"}

	tests := []struct {
		name     string
		input    string
		headers  []string
		expected []int
	}{
		{
			name:     "empty string",
			input:    "",
			headers:  headers,
			expected: []int{},
		},
		{
			name:     "single index",
			input:    "2",
			headers:  headers,
			expected: []int{1}, // 0-based
		},
		{
			name:     "range of indices",
			input:    "1-3",
			headers:  headers,
			expected: []int{0, 1, 2}, // 0-based
		},
		{
			name:     "single column name",
			input:    "col3",
			headers:  headers,
			expected: []int{2}, // 0-based
		},
		{
			name:     "mixed names and indices",
			input:    "1,col3,5",
			headers:  headers,
			expected: []int{0, 2, 4}, // 0-based
		},
		{
			name:     "mixed names and ranges",
			input:    "col1,2-3,col5",
			headers:  headers,
			expected: []int{0, 1, 2, 4}, // 0-based
		},
		{
			name:     "out of range indices ignored",
			input:    "1,10,3",
			headers:  headers,
			expected: []int{0, 2}, // 10 is out of range
		},
		{
			name:     "invalid column names ignored",
			input:    "col1,invalid,col3",
			headers:  headers,
			expected: []int{0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExcludeColumns(tt.input, tt.headers)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseExcludeColumns(%q, %v) = %v, want %v", tt.input, tt.headers, result, tt.expected)
			}
		})
	}
}

// TestParseDelimiter verifies delimiter parsing with validation and escape sequences.
// This is a regression test for issue #600.
func TestParseDelimiter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    rune
		expectError bool
	}{
		{
			name:        "comma delimiter",
			input:       ",",
			expected:    ',',
			expectError: false,
		},
		{
			name:        "semicolon delimiter",
			input:       ";",
			expected:    ';',
			expectError: false,
		},
		{
			name:        "pipe delimiter",
			input:       "|",
			expected:    '|',
			expectError: false,
		},
		{
			name:        "tab escape sequence",
			input:       "\\t",
			expected:    '\t',
			expectError: false,
		},
		{
			name:        "newline escape sequence",
			input:       "\\n",
			expected:    '\n',
			expectError: false,
		},
		{
			name:        "carriage return escape sequence",
			input:       "\\r",
			expected:    '\r',
			expectError: false,
		},
		{
			name:        "empty delimiter - should error",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "multi-character delimiter - should error",
			input:       "ab",
			expected:    0,
			expectError: true,
		},
		{
			name:        "multi-character delimiter - should error",
			input:       "|||",
			expected:    0,
			expectError: true,
		},
		{
			name:        "unicode character",
			input:       "→",
			expected:    '→',
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDelimiter(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseDelimiter(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseDelimiter(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseDelimiter(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}
