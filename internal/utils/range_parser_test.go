package utils

import (
	"reflect"
	"testing"
)

func TestParseRanges(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    []int{},
			wantErr: false,
		},
		{
			name:    "single index",
			input:   "3",
			want:    []int{2}, // 0-based
			wantErr: false,
		},
		{
			name:    "multiple indices",
			input:   "1,3,5",
			want:    []int{0, 2, 4}, // 0-based
			wantErr: false,
		},
		{
			name:    "simple range",
			input:   "2-4",
			want:    []int{1, 2, 3}, // 0-based
			wantErr: false,
		},
		{
			name:    "mixed indices and ranges",
			input:   "1,3-5,7",
			want:    []int{0, 2, 3, 4, 6}, // 0-based
			wantErr: false,
		},
		{
			name:    "with spaces",
			input:   "1, 3 - 5 , 7",
			want:    []int{0, 2, 3, 4, 6}, // 0-based
			wantErr: false,
		},
		{
			name:    "duplicates removed",
			input:   "1,2,1,3,2-4",
			want:    []int{0, 1, 2, 3}, // 0-based, sorted, unique
			wantErr: false,
		},
		{
			name:    "invalid index",
			input:   "0",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative index",
			input:   "-1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid range format",
			input:   "1-2-3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "reversed range",
			input:   "5-3",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "non-numeric",
			input:   "a,b,c",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRanges(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterMatrix(t *testing.T) {
	tests := []struct {
		name            string
		data            [][]float64
		excludedRows    []int
		excludedColumns []int
		want            [][]float64
		wantErr         bool
	}{
		{
			name: "no exclusions",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			excludedRows:    []int{},
			excludedColumns: []int{},
			want: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			wantErr: false,
		},
		{
			name: "exclude rows",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			excludedRows:    []int{0, 2},
			excludedColumns: []int{},
			want: [][]float64{
				{4, 5, 6},
			},
			wantErr: false,
		},
		{
			name: "exclude columns",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			excludedRows:    []int{},
			excludedColumns: []int{0, 2},
			want: [][]float64{
				{2},
				{5},
				{8},
			},
			wantErr: false,
		},
		{
			name: "exclude both",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			excludedRows:    []int{1},
			excludedColumns: []int{1},
			want: [][]float64{
				{1, 3},
				{7, 9},
			},
			wantErr: false,
		},
		{
			name: "out of bounds row",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
			},
			excludedRows:    []int{3},
			excludedColumns: []int{},
			want:            nil,
			wantErr:         true,
		},
		{
			name: "out of bounds column",
			data: [][]float64{
				{1, 2, 3},
				{4, 5, 6},
			},
			excludedRows:    []int{},
			excludedColumns: []int{3},
			want:            nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterMatrix(tt.data, tt.excludedRows, tt.excludedColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterMatrix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterMatrix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterStringSlice(t *testing.T) {
	tests := []struct {
		name            string
		items           []string
		excludedIndices []int
		want            []string
		wantErr         bool
	}{
		{
			name:            "no exclusions",
			items:           []string{"a", "b", "c"},
			excludedIndices: []int{},
			want:            []string{"a", "b", "c"},
			wantErr:         false,
		},
		{
			name:            "exclude some",
			items:           []string{"a", "b", "c", "d"},
			excludedIndices: []int{1, 3},
			want:            []string{"a", "c"},
			wantErr:         false,
		},
		{
			name:            "out of bounds",
			items:           []string{"a", "b", "c"},
			excludedIndices: []int{3},
			want:            nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterStringSlice(tt.items, tt.excludedIndices)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
