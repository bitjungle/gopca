// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
