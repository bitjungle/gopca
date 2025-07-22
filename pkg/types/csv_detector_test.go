package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestDetectColumnTypes(t *testing.T) {
	tests := []struct {
		name            string
		csvContent      string
		format          CSVFormat
		wantNumeric     []int
		wantCategorical []int
		wantHeaders     []string
		wantErr         bool
	}{
		{
			name: "all numeric columns",
			csvContent: `a,b,c
1,2,3
4,5,6`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1, 2},
			wantCategorical: []int{},
			wantHeaders:     []string{"a", "b", "c"},
		},
		{
			name: "mixed columns",
			csvContent: `name,age,score,grade
John,25,85.5,A
Jane,30,92.0,B
Bob,35,78.5,A`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1}, // age, score
			wantCategorical: []int{2},    // grade
			wantHeaders:     []string{"age", "score", "grade"},
		},
		{
			name: "with null values",
			csvContent: `x,y,z
1,NA,apple
2,3,banana
NA,4,cherry`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1}, // x and y (NA is treated as numeric)
			wantCategorical: []int{2},    // z
			wantHeaders:     []string{"x", "y", "z"},
		},
		{
			name: "empty columns",
			csvContent: `a,b,c
1,,3
4,,6`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1, 2}, // empty columns default to numeric
			wantCategorical: []int{},
			wantHeaders:     []string{"a", "b", "c"},
		},
		{
			name: "special numeric values",
			csvContent: `a,b,c
1,inf,hello
2,-inf,world`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1}, // a and b (inf is numeric)
			wantCategorical: []int{2},    // c
			wantHeaders:     []string{"a", "b", "c"},
		},
		{
			name: "no headers",
			csvContent: `1,2,A
3,4,B`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       false,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0, 1},
			wantCategorical: []int{2},
			wantHeaders:     nil,
		},
		{
			name:       "single row",
			csvContent: `a,b,c`,
			format: CSVFormat{
				FieldDelimiter: ',',
				HasHeaders:     true,
			},
			wantNumeric:     nil,
			wantCategorical: nil,
			wantHeaders:     nil,
			wantErr:         false, // Should return nil without error
		},
		{
			name: "type detection with more than 10 rows",
			csvContent: `num,cat
1,A
2,B
3,C
4,D
5,E
6,F
7,G
8,H
9,I
10,J
11,K
12,L`,
			format: CSVFormat{
				FieldDelimiter:   ',',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      false,
				NullValues:       []string{"NA"},
			},
			wantNumeric:     []int{0},
			wantCategorical: []int{1},
			wantHeaders:     []string{"num", "cat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvContent)
			numericCols, categoricalCols, headers, err := DetectColumnTypes(reader, tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectColumnTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(numericCols, tt.wantNumeric) {
				t.Errorf("numericCols = %v, want %v", numericCols, tt.wantNumeric)
			}

			// Handle nil vs empty slice comparison
			if tt.wantCategorical == nil {
				if categoricalCols != nil {
					t.Errorf("categoricalCols = %v, want nil", categoricalCols)
				}
			} else if categoricalCols == nil {
				t.Errorf("categoricalCols = nil, want %v", tt.wantCategorical)
			} else if !reflect.DeepEqual(categoricalCols, tt.wantCategorical) {
				t.Errorf("categoricalCols = %v, want %v", categoricalCols, tt.wantCategorical)
			}

			if !reflect.DeepEqual(headers, tt.wantHeaders) {
				t.Errorf("headers = %v, want %v", headers, tt.wantHeaders)
			}
		})
	}
}
