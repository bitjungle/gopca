// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"reflect"
	"testing"
)

func TestDefaultMissingValues(t *testing.T) {
	defaults := DefaultMissingValues()
	expected := []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL", "m"}

	if !reflect.DeepEqual(defaults, expected) {
		t.Errorf("DefaultMissingValues() = %v, want %v", defaults, expected)
	}
}

func TestIsMissingValue(t *testing.T) {
	defaultIndicators := DefaultMissingValues()

	tests := []struct {
		name       string
		value      string
		indicators []string
		want       bool
	}{
		// Empty string cases
		{"empty string", "", defaultIndicators, true},
		{"whitespace only", "   ", defaultIndicators, true},
		{"tabs and spaces", "\t  \n", defaultIndicators, true},

		// Standard missing values
		{"NA uppercase", "NA", defaultIndicators, true},
		{"na lowercase", "na", defaultIndicators, true},
		{"N/A", "N/A", defaultIndicators, true},
		{"n/a lowercase", "n/a", defaultIndicators, true},
		{"NaN", "NaN", defaultIndicators, true},
		{"nan lowercase", "nan", defaultIndicators, true},
		{"null", "null", defaultIndicators, true},
		{"NULL uppercase", "NULL", defaultIndicators, true},
		{"m", "m", defaultIndicators, true},
		{"M uppercase", "M", defaultIndicators, true},

		// With whitespace
		{"NA with leading space", " NA", defaultIndicators, true},
		{"NA with trailing space", "NA ", defaultIndicators, true},
		{"NA with surrounding spaces", "  NA  ", defaultIndicators, true},

		// Non-missing values
		{"regular number", "123", defaultIndicators, false},
		{"regular text", "hello", defaultIndicators, false},
		{"NA as part of word", "NATIONAL", defaultIndicators, false},
		{"contains NA", "BANANA", defaultIndicators, false},

		// Custom indicators
		{"custom indicator", "missing", []string{"missing", "absent"}, true},
		{"not in custom list", "NA", []string{"missing", "absent"}, false},
		{"case insensitive custom", "MISSING", []string{"missing"}, true},

		// Edge cases
		{"single space not in list", " ", []string{"NA"}, false},
		{"empty indicators list", "NA", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMissingValue(tt.value, tt.indicators); got != tt.want {
				t.Errorf("IsMissingValue(%q, %v) = %v, want %v",
					tt.value, tt.indicators, got, tt.want)
			}
		})
	}
}

func TestContainsMissingValues(t *testing.T) {
	defaultIndicators := DefaultMissingValues()

	tests := []struct {
		name       string
		values     []string
		indicators []string
		want       bool
	}{
		{
			name:       "no missing values",
			values:     []string{"1", "2", "3", "4"},
			indicators: defaultIndicators,
			want:       false,
		},
		{
			name:       "contains NA",
			values:     []string{"1", "NA", "3", "4"},
			indicators: defaultIndicators,
			want:       true,
		},
		{
			name:       "contains empty string",
			values:     []string{"1", "2", "", "4"},
			indicators: defaultIndicators,
			want:       true,
		},
		{
			name:       "all missing",
			values:     []string{"NA", "null", "", "NaN"},
			indicators: defaultIndicators,
			want:       true,
		},
		{
			name:       "empty slice",
			values:     []string{},
			indicators: defaultIndicators,
			want:       false,
		},
		{
			name:       "whitespace counted as missing",
			values:     []string{"1", "  ", "3"},
			indicators: defaultIndicators,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsMissingValues(tt.values, tt.indicators); got != tt.want {
				t.Errorf("ContainsMissingValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountMissingValues(t *testing.T) {
	defaultIndicators := DefaultMissingValues()

	tests := []struct {
		name       string
		values     []string
		indicators []string
		want       int
	}{
		{
			name:       "no missing values",
			values:     []string{"1", "2", "3", "4"},
			indicators: defaultIndicators,
			want:       0,
		},
		{
			name:       "one missing value",
			values:     []string{"1", "NA", "3", "4"},
			indicators: defaultIndicators,
			want:       1,
		},
		{
			name:       "multiple missing values",
			values:     []string{"NA", "2", "null", "", "5"},
			indicators: defaultIndicators,
			want:       3,
		},
		{
			name:       "all missing",
			values:     []string{"NA", "null", "", "NaN"},
			indicators: defaultIndicators,
			want:       4,
		},
		{
			name:       "empty slice",
			values:     []string{},
			indicators: defaultIndicators,
			want:       0,
		},
		{
			name:       "case variations",
			values:     []string{"na", "NA", "Na", "nA"},
			indicators: defaultIndicators,
			want:       4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountMissingValues(tt.values, tt.indicators); got != tt.want {
				t.Errorf("CountMissingValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkIsMissingValue(b *testing.B) {
	indicators := DefaultMissingValues()
	testValues := []string{"NA", "123", "", "hello", "null", "3.14"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range testValues {
			_ = IsMissingValue(v, indicators)
		}
	}
}
