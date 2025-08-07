// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"math"
	"testing"
)

func TestParseNumericValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		separator rune
		want      float64
		wantErr   bool
		checkNaN  bool
		checkInf  int // 0 = not inf, 1 = +inf, -1 = -inf
	}{
		// Basic numbers with dot separator
		{"integer", "123", '.', 123, false, false, 0},
		{"decimal", "123.45", '.', 123.45, false, false, 0},
		{"negative", "-42.5", '.', -42.5, false, false, 0},
		{"scientific notation", "1.23e4", '.', 12300, false, false, 0},
		{"scientific negative exp", "1.5e-3", '.', 0.0015, false, false, 0},

		// Numbers with comma separator
		{"comma decimal", "123,45", ',', 123.45, false, false, 0},
		{"comma negative", "-42,5", ',', -42.5, false, false, 0},
		{"comma scientific", "1,23e4", ',', 12300, false, false, 0},

		// Special values
		{"infinity lowercase", "inf", '.', 0, false, false, 1},
		{"infinity uppercase", "INF", '.', 0, false, false, 1},
		{"positive infinity", "+inf", '.', 0, false, false, 1},
		{"infinity word", "infinity", '.', 0, false, false, 1},
		{"negative infinity", "-inf", '.', 0, false, false, -1},
		{"negative infinity word", "-infinity", '.', 0, false, false, -1},
		{"nan lowercase", "nan", '.', 0, false, true, 0},
		{"nan uppercase", "NaN", '.', 0, false, true, 0},
		{"nan mixed case", "NAN", '.', 0, false, true, 0},

		// Whitespace handling
		{"leading whitespace", "  123", '.', 123, false, false, 0},
		{"trailing whitespace", "123  ", '.', 123, false, false, 0},
		{"surrounded whitespace", "  123.45  ", '.', 123.45, false, false, 0},

		// Error cases
		{"empty string", "", '.', 0, true, false, 0},
		{"only whitespace", "   ", '.', 0, true, false, 0},
		{"not a number", "hello", '.', 0, true, false, 0},
		{"mixed content", "123abc", '.', 0, true, false, 0},
		{"multiple dots", "12.34.56", '.', 0, true, false, 0},
		{"multiple commas", "12,34,56", ',', 0, true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNumericValue(tt.value, tt.separator)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNumericValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.checkNaN {
				if !math.IsNaN(got) {
					t.Errorf("ParseNumericValue() = %v, want NaN", got)
				}
			} else if tt.checkInf != 0 {
				if !math.IsInf(got, tt.checkInf) {
					t.Errorf("ParseNumericValue() = %v, want Inf(%d)", got, tt.checkInf)
				}
			} else if got != tt.want {
				t.Errorf("ParseNumericValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNumericValueWithMissing(t *testing.T) {
	defaultMissing := DefaultMissingValues()

	tests := []struct {
		name        string
		value       string
		separator   rune
		missing     []string
		wantValue   float64
		wantMissing bool
		wantErr     bool
		checkNaN    bool
	}{
		// Regular numbers
		{"regular number", "123.45", '.', defaultMissing, 123.45, false, false, false},
		{"negative number", "-42", '.', defaultMissing, -42, false, false, false},

		// Missing values
		{"NA", "NA", '.', defaultMissing, 0, true, false, true},
		{"empty string", "", '.', defaultMissing, 0, true, false, true},
		{"null", "null", '.', defaultMissing, 0, true, false, true},
		{"whitespace", "  ", '.', defaultMissing, 0, true, false, true},

		// Special numeric values (not missing)
		{"infinity", "inf", '.', defaultMissing, math.Inf(1), false, false, false},
		{"nan numeric", "nan", '.', []string{"NA"}, 0, false, false, true}, // "nan" parses as NaN, not missing

		// Error cases
		{"invalid number", "abc", '.', defaultMissing, 0, false, true, false},
		{"partial number", "123abc", '.', defaultMissing, 0, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, isMissing, err := ParseNumericValueWithMissing(tt.value, tt.separator, tt.missing)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNumericValueWithMissing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if isMissing != tt.wantMissing {
				t.Errorf("ParseNumericValueWithMissing() isMissing = %v, want %v", isMissing, tt.wantMissing)
			}

			if tt.wantErr {
				return
			}

			if tt.checkNaN || (tt.wantMissing && math.IsNaN(got)) {
				if !math.IsNaN(got) {
					t.Errorf("ParseNumericValueWithMissing() = %v, want NaN", got)
				}
			} else if !tt.wantMissing && got != tt.wantValue {
				t.Errorf("ParseNumericValueWithMissing() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestIsNumericString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		separator rune
		want      bool
	}{
		{"integer", "123", '.', true},
		{"decimal", "123.45", '.', true},
		{"negative", "-42", '.', true},
		{"scientific", "1e10", '.', true},
		{"comma decimal", "123,45", ',', true},
		{"infinity", "inf", '.', true},
		{"not numeric", "hello", '.', false},
		{"empty", "", '.', false},
		{"mixed", "123abc", '.', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumericString(tt.value, tt.separator); got != tt.want {
				t.Errorf("IsNumericString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFloatSlice(t *testing.T) {
	defaultMissing := DefaultMissingValues()

	tests := []struct {
		name      string
		values    []string
		separator rune
		missing   []string
		want      []float64
		wantErr   bool
		checkNaN  []int // indices that should be NaN
	}{
		{
			name:      "all valid numbers",
			values:    []string{"1", "2.5", "-3", "4e2"},
			separator: '.',
			missing:   defaultMissing,
			want:      []float64{1, 2.5, -3, 400},
			wantErr:   false,
		},
		{
			name:      "with missing values",
			values:    []string{"1", "NA", "3", ""},
			separator: '.',
			missing:   defaultMissing,
			want:      []float64{1, math.NaN(), 3, math.NaN()},
			wantErr:   false,
			checkNaN:  []int{1, 3},
		},
		{
			name:      "comma separator",
			values:    []string{"1,5", "2,0", "3,14"},
			separator: ',',
			missing:   defaultMissing,
			want:      []float64{1.5, 2.0, 3.14},
			wantErr:   false,
		},
		{
			name:      "error on invalid",
			values:    []string{"1", "abc", "3"},
			separator: '.',
			missing:   defaultMissing,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "empty slice",
			values:    []string{},
			separator: '.',
			missing:   defaultMissing,
			want:      []float64{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloatSlice(tt.values, tt.separator, tt.missing)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFloatSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseFloatSlice() len = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				isNaN := false
				for _, idx := range tt.checkNaN {
					if i == idx {
						isNaN = true
						break
					}
				}

				if isNaN {
					if !math.IsNaN(got[i]) {
						t.Errorf("ParseFloatSlice()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if got[i] != tt.want[i] {
						t.Errorf("ParseFloatSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func BenchmarkParseNumericValue(b *testing.B) {
	testValues := []string{"123.45", "-42", "1e10", "inf", "hello"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range testValues {
			_, _ = ParseNumericValue(v, '.')
		}
	}
}
