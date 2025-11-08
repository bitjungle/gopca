// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"math"
	"strings"
	"testing"
)

func TestParseNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		wantRows int
		wantCols int
		wantErr  bool
	}{
		{
			name: "simple numeric CSV",
			input: `A,B,C
1,2,3
4,5,6`,
			opts: func() Options {
				opts := DefaultOptions()
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name: "with row names",
			input: `"",A,B,C
row1,1,2,3
row2,4,5,6`,
			opts:     DefaultOptions(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name: "European format",
			input: `A;B;C
1,1;2,2;3,3
4,4;5,5;6,6`,
			opts: func() Options {
				opts := EuropeanOptions()
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name: "with missing values",
			input: `A,B,C
1,NA,3
4,5,`,
			opts: func() Options {
				opts := DefaultOptions()
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name:  "tab delimited",
			input: "A\tB\tC\n1\t2\t3\n4\t5\t6",
			opts: func() Options {
				opts := TabDelimitedOptions()
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name:     "empty file",
			input:    "",
			opts:     DefaultOptions(),
			wantRows: 0,
			wantCols: 0,
			wantErr:  true,
		},
		{
			name:     "no data rows",
			input:    `A,B,C`,
			opts:     DefaultOptions(),
			wantRows: 0,
			wantCols: 0,
			wantErr:  true,
		},
		{
			name: "skip rows",
			input: `Comment line
A,B,C
1,2,3
4,5,6`,
			opts: func() Options {
				opts := DefaultOptions()
				opts.SkipRows = 1
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name: "max rows limit",
			input: `A,B,C
1,2,3
4,5,6
7,8,9`,
			opts: func() Options {
				opts := DefaultOptions()
				opts.MaxRows = 2
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
		{
			name: "select columns",
			input: `A,B,C,D,E
1,2,3,4,5
6,7,8,9,10`,
			opts: func() Options {
				opts := DefaultOptions()
				opts.Columns = []int{0, 2, 4}
				opts.HasRowNames = false
				return opts
			}(),
			wantRows: 2,
			wantCols: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(tt.opts)
			data, err := reader.Read(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if data.Rows != tt.wantRows {
					t.Errorf("Parse() rows = %v, want %v", data.Rows, tt.wantRows)
				}
				if data.Columns != tt.wantCols {
					t.Errorf("Parse() columns = %v, want %v", data.Columns, tt.wantCols)
				}
			}
		})
	}
}

func TestParseWithRowNames(t *testing.T) {
	input := `"",A,B,C
row1,1,2,3
row2,4,5,6`

	reader := NewReader(DefaultOptions())
	data, err := reader.Read(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.RowNames) != 2 {
		t.Errorf("expected 2 row names, got %d", len(data.RowNames))
	}

	if data.RowNames[0] != "row1" || data.RowNames[1] != "row2" {
		t.Errorf("unexpected row names: %v", data.RowNames)
	}

	if len(data.Headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(data.Headers))
	}
}

func TestParseMissingValues(t *testing.T) {
	input := `A,B,C
1,NA,3
4,,6
7,null,9`

	opts := DefaultOptions()
	opts.HasRowNames = false
	reader := NewReader(opts)
	data, err := reader.Read(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that missing values are properly marked
	if !math.IsNaN(data.Matrix[0][1]) {
		t.Error("expected NaN for 'NA' value")
	}

	if !math.IsNaN(data.Matrix[1][1]) {
		t.Error("expected NaN for empty value")
	}

	if !math.IsNaN(data.Matrix[2][1]) {
		t.Error("expected NaN for 'null' value")
	}

	// Check missing mask
	if data.MissingMask != nil {
		if !data.MissingMask[0][1] {
			t.Error("expected missing mask to be true for 'NA'")
		}
		if !data.MissingMask[1][1] {
			t.Error("expected missing mask to be true for empty")
		}
		if !data.MissingMask[2][1] {
			t.Error("expected missing mask to be true for 'null'")
		}
	}
}

func TestParseEuropeanFormat(t *testing.T) {
	input := `A;B;C
1,5;2,3;3,7
4,2;5,8;6,1`

	opts := EuropeanOptions()
	opts.HasRowNames = false
	reader := NewReader(opts)
	data, err := reader.Read(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check decimal parsing
	tolerance := 0.001
	if math.Abs(data.Matrix[0][0]-1.5) > tolerance {
		t.Errorf("expected 1.5, got %f", data.Matrix[0][0])
	}
	if math.Abs(data.Matrix[0][1]-2.3) > tolerance {
		t.Errorf("expected 2.3, got %f", data.Matrix[0][1])
	}
}

func TestParseString(t *testing.T) {
	input := `Name,Age,City
Alice,30,NYC
Bob,25,LA`

	opts := DefaultOptions()
	opts.ParseMode = ParseString
	opts.HasRowNames = false

	reader := NewReader(opts)
	data, err := reader.Read(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.StringData == nil {
		t.Fatal("expected string data to be present")
	}

	if len(data.StringData) != 2 {
		t.Errorf("expected 2 rows, got %d", len(data.StringData))
	}

	if data.StringData[0][0] != "Alice" {
		t.Errorf("expected 'Alice', got %s", data.StringData[0][0])
	}
}

func TestParseWithInfinity(t *testing.T) {
	input := `A,B,C
1,inf,3
4,-inf,6`

	opts := DefaultOptions()
	opts.HasRowNames = false
	reader := NewReader(opts)
	data, err := reader.Read(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !math.IsInf(data.Matrix[0][1], 1) {
		t.Error("expected positive infinity for 'inf'")
	}

	if !math.IsInf(data.Matrix[1][1], -1) {
		t.Error("expected negative infinity for '-inf'")
	}
}

func TestParseInconsistentColumns(t *testing.T) {
	input := `A,B,C
1,2,3
4,5`

	opts := DefaultOptions()
	opts.HasRowNames = false
	reader := NewReader(opts)
	_, err := reader.Read(strings.NewReader(input))

	if err == nil {
		t.Error("expected error for inconsistent columns")
	}
}

func TestParseInvalidNumeric(t *testing.T) {
	input := `A,B,C
1,2,3
4,abc,6`

	opts := DefaultOptions()
	opts.HasRowNames = false
	reader := NewReader(opts)
	_, err := reader.Read(strings.NewReader(input))

	if err == nil {
		t.Error("expected error for invalid numeric value")
	}
}

// TestTargetColumns verifies that target columns are correctly excluded from the feature matrix
// but remain available in NumericTargetColumns for visualization purposes.
// This is a regression test for issue #598.
func TestTargetColumns(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		targetCols       []string
		wantFeatureCols  int
		wantTargetCols   int
		expectedFeatures []string
		expectedTargets  []string
	}{
		{
			name: "single target column",
			input: `"",A,B,C,Target
row1,1,2,3,100
row2,4,5,6,200`,
			targetCols:       []string{"Target"},
			wantFeatureCols:  3,
			wantTargetCols:   1,
			expectedFeatures: []string{"A", "B", "C"},
			expectedTargets:  []string{"Target"},
		},
		{
			name: "multiple target columns",
			input: `"",A,B,C,Target1,Target2
row1,1,2,3,100,101
row2,4,5,6,200,201`,
			targetCols:       []string{"Target1", "Target2"},
			wantFeatureCols:  3,
			wantTargetCols:   2,
			expectedFeatures: []string{"A", "B", "C"},
			expectedTargets:  []string{"Target1", "Target2"},
		},
		{
			name: "target columns with whitespace in input string",
			input: `"",A,B,C,Target1,Target2
row1,1,2,3,100,101
row2,4,5,6,200,201`,
			targetCols:       []string{"Target1", "Target2"}, // Already trimmed (simulating analyze.go behavior)
			wantFeatureCols:  3,
			wantTargetCols:   2,
			expectedFeatures: []string{"A", "B", "C"},
			expectedTargets:  []string{"Target1", "Target2"},
		},
		{
			name: "no target columns",
			input: `"",A,B,C
row1,1,2,3
row2,4,5,6`,
			targetCols:       nil,
			wantFeatureCols:  3,
			wantTargetCols:   0,
			expectedFeatures: []string{"A", "B", "C"},
			expectedTargets:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.TargetCols = tt.targetCols
			if len(tt.targetCols) > 0 {
				opts.ParseMode = ParseMixedWithTargets
			}
			reader := NewReader(opts)
			data, err := reader.Read(strings.NewReader(tt.input))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check feature matrix dimensions
			if len(data.Headers) != tt.wantFeatureCols {
				t.Errorf("feature columns = %d, want %d (got: %v)",
					len(data.Headers), tt.wantFeatureCols, data.Headers)
			}

			// Check target columns count
			if len(data.NumericTargetColumns) != tt.wantTargetCols {
				t.Errorf("target columns = %d, want %d (got: %v)",
					len(data.NumericTargetColumns), tt.wantTargetCols, data.NumericTargetColumns)
			}

			// Verify feature column names
			for i, expected := range tt.expectedFeatures {
				if i >= len(data.Headers) {
					t.Errorf("missing feature column at index %d: %s", i, expected)
					continue
				}
				if data.Headers[i] != expected {
					t.Errorf("feature column %d = %s, want %s",
						i, data.Headers[i], expected)
				}
			}

			// Verify target column names
			targetNames := make([]string, 0, len(data.NumericTargetColumns))
			for name := range data.NumericTargetColumns {
				targetNames = append(targetNames, name)
			}
			if len(targetNames) != len(tt.expectedTargets) {
				t.Errorf("target column count = %d, want %d",
					len(targetNames), len(tt.expectedTargets))
			}
			for _, expected := range tt.expectedTargets {
				if _, exists := data.NumericTargetColumns[expected]; !exists {
					t.Errorf("target column %s not found in NumericTargetColumns", expected)
				}
			}

			// Verify that target columns are NOT in the feature matrix
			for _, targetName := range tt.expectedTargets {
				for _, featureName := range data.Headers {
					if featureName == targetName {
						t.Errorf("target column %s found in feature matrix (should be excluded)",
							targetName)
					}
				}
			}
		})
	}
}
