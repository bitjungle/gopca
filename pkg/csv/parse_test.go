// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// ─── GetOriginalHeaders ───────────────────────────────────────────────────────

var commaFormat = types.CSVFormat{
	FieldDelimiter: ',',
	HasHeaders:     true,
	HasRowNames:    false,
}

var commaFormatWithRowNames = types.CSVFormat{
	FieldDelimiter: ',',
	HasHeaders:     true,
	HasRowNames:    true,
}

var tsvFormat = types.CSVFormat{
	FieldDelimiter: '\t',
	HasHeaders:     true,
	HasRowNames:    false,
}

func TestGetOriginalHeaders_Standard(t *testing.T) {
	content := "a,b,c\n1,2,3\n4,5,6\n"
	got := GetOriginalHeaders(content, commaFormat)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, h := range want {
		if got[i] != h {
			t.Errorf("[%d] got %q, want %q", i, got[i], h)
		}
	}
}

func TestGetOriginalHeaders_SkipsRowNameColumn(t *testing.T) {
	content := "rowname,x,y\nR1,1,2\nR2,3,4\n"
	got := GetOriginalHeaders(content, commaFormatWithRowNames)
	want := []string{"x", "y"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, h := range want {
		if got[i] != h {
			t.Errorf("[%d] got %q, want %q", i, got[i], h)
		}
	}
}

func TestGetOriginalHeaders_TSV(t *testing.T) {
	content := "col1\tcol2\tcol3\n1\t2\t3\n"
	got := GetOriginalHeaders(content, tsvFormat)
	want := []string{"col1", "col2", "col3"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, h := range want {
		if got[i] != h {
			t.Errorf("[%d] got %q, want %q", i, got[i], h)
		}
	}
}

func TestGetOriginalHeaders_Empty(t *testing.T) {
	got := GetOriginalHeaders("", commaFormat)
	if len(got) != 0 {
		t.Errorf("expected empty slice for empty content, got %v", got)
	}
}

func TestGetOriginalHeaders_SingleRow(t *testing.T) {
	content := "alpha,beta\n"
	got := GetOriginalHeaders(content, commaFormat)
	want := []string{"alpha", "beta"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, h := range want {
		if got[i] != h {
			t.Errorf("[%d] got %q, want %q", i, got[i], h)
		}
	}
}

// ─── CombineColumns ───────────────────────────────────────────────────────────

// makeCSVData is a helper to build a minimal *types.CSVData for tests.
func makeCSVData(headers []string, matrix [][]float64, rowNames []string, missingMask [][]bool) *types.CSVData {
	rows := len(matrix)
	cols := 0
	if rows > 0 {
		cols = len(matrix[0])
	}
	return &types.CSVData{
		Matrix:      matrix,
		Headers:     headers,
		RowNames:    rowNames,
		MissingMask: missingMask,
		Rows:        rows,
		Columns:     cols,
	}
}

func TestCombineColumns_NumericOnly(t *testing.T) {
	csvData := makeCSVData(
		[]string{"x", "y"},
		[][]float64{{1, 2}, {3, 4}},
		nil, nil,
	)
	res := CombineColumns(csvData, nil, nil, []string{"x", "y"})

	if res.Rows != 2 {
		t.Errorf("Rows: got %d, want 2", res.Rows)
	}
	if res.Columns != 2 {
		t.Errorf("Columns: got %d, want 2", res.Columns)
	}
	if res.ColumnTypes["x"] != "numeric" || res.ColumnTypes["y"] != "numeric" {
		t.Errorf("ColumnTypes: got %v", res.ColumnTypes)
	}
	// First row values
	if res.Data[0][0] != "1" || res.Data[0][1] != "2" {
		t.Errorf("Data[0]: got %v, want [1 2]", res.Data[0])
	}
}

func TestCombineColumns_ColumnOrderPreserved(t *testing.T) {
	csvData := makeCSVData(
		[]string{"n1", "n2"},
		[][]float64{{10, 20}},
		nil, nil,
	)
	catData := map[string][]string{"cat": {"A"}}
	targetData := map[string][]float64{"tgt": {99.0}}

	// Force a specific order: cat, n1, tgt, n2
	headers := []string{"cat", "n1", "tgt", "n2"}
	res := CombineColumns(csvData, catData, targetData, headers)

	if len(res.Headers) != 4 {
		t.Fatalf("Headers: got %v", res.Headers)
	}
	for i, want := range headers {
		if res.Headers[i] != want {
			t.Errorf("Headers[%d]: got %q, want %q", i, res.Headers[i], want)
		}
	}
	// Verify column types
	if res.ColumnTypes["cat"] != "categorical" {
		t.Errorf("cat: expected categorical, got %q", res.ColumnTypes["cat"])
	}
	if res.ColumnTypes["n1"] != "numeric" {
		t.Errorf("n1: expected numeric, got %q", res.ColumnTypes["n1"])
	}
	if res.ColumnTypes["tgt"] != "target" {
		t.Errorf("tgt: expected target, got %q", res.ColumnTypes["tgt"])
	}
}

func TestCombineColumns_MissingValuesAsEmptyStrings(t *testing.T) {
	mask := [][]bool{{false, true}, {true, false}}
	csvData := makeCSVData(
		[]string{"a", "b"},
		[][]float64{{1, 0}, {0, 5}},
		nil,
		mask,
	)
	res := CombineColumns(csvData, nil, nil, []string{"a", "b"})

	// row 0 col 1 is missing
	if res.Data[0][1] != "" {
		t.Errorf("expected empty for missing, got %q", res.Data[0][1])
	}
	// row 1 col 0 is missing
	if res.Data[1][0] != "" {
		t.Errorf("expected empty for missing, got %q", res.Data[1][0])
	}
	// row 0 col 0 is present
	if res.Data[0][0] != "1" {
		t.Errorf("expected \"1\", got %q", res.Data[0][0])
	}
}

func TestCombineColumns_RowNamesPreserved(t *testing.T) {
	csvData := makeCSVData(
		[]string{"v"},
		[][]float64{{7}, {8}},
		[]string{"row1", "row2"},
		nil,
	)
	res := CombineColumns(csvData, nil, nil, []string{"v"})

	if len(res.RowNames) != 2 || res.RowNames[0] != "row1" || res.RowNames[1] != "row2" {
		t.Errorf("RowNames: got %v", res.RowNames)
	}
}

func TestCombineColumns_FallbackOrderingWhenNoOriginalHeaders(t *testing.T) {
	csvData := makeCSVData(
		[]string{"n"},
		[][]float64{{42}},
		nil, nil,
	)
	catData := map[string][]string{"c": {"X"}}
	targetData := map[string][]float64{"t": {3.14}}

	// Pass nil originalHeaders → fallback ordering: numeric → categorical → target
	res := CombineColumns(csvData, catData, targetData, nil)

	if res.Columns != 3 {
		t.Errorf("Columns: got %d, want 3", res.Columns)
	}
	// All three columns should appear
	typesSeen := map[string]bool{}
	for _, h := range res.Headers {
		typesSeen[res.ColumnTypes[h]] = true
	}
	if !typesSeen["numeric"] || !typesSeen["categorical"] || !typesSeen["target"] {
		t.Errorf("missing column types: %v", typesSeen)
	}
}

func TestCombineColumns_UnknownHeadersSkipped(t *testing.T) {
	csvData := makeCSVData(
		[]string{"known"},
		[][]float64{{1}},
		nil, nil,
	)
	// "ghost" is not in any source
	res := CombineColumns(csvData, nil, nil, []string{"ghost", "known"})

	if res.Columns != 1 {
		t.Errorf("Columns: got %d, want 1 (ghost should be skipped)", res.Columns)
	}
	if res.Headers[0] != "known" {
		t.Errorf("Headers[0]: got %q, want \"known\"", res.Headers[0])
	}
}

func TestCombineColumns_CategoricalValuesCorrect(t *testing.T) {
	csvData := makeCSVData([]string{}, nil, nil, nil)
	catData := map[string][]string{"label": {"foo", "bar", "baz"}}
	csvData.Rows = 3

	res := CombineColumns(csvData, catData, nil, []string{"label"})

	if res.Data[0][0] != "foo" || res.Data[1][0] != "bar" || res.Data[2][0] != "baz" {
		t.Errorf("categorical values: got %v", []string{res.Data[0][0], res.Data[1][0], res.Data[2][0]})
	}
}

func TestCombineColumns_TargetValuesFormatted(t *testing.T) {
	csvData := makeCSVData([]string{}, nil, nil, nil)
	targetData := map[string][]float64{"score": {1.5, 2.5}}
	csvData.Rows = 2

	res := CombineColumns(csvData, nil, targetData, []string{"score"})

	if res.Data[0][0] != "1.5" || res.Data[1][0] != "2.5" {
		t.Errorf("target values: got %v", []string{res.Data[0][0], res.Data[1][0]})
	}
}

func TestCombineColumns_NumericTargetDataInResult(t *testing.T) {
	csvData := makeCSVData([]string{}, nil, nil, nil)
	targetData := map[string][]float64{"y": {10.0, 20.0}}
	csvData.Rows = 2

	res := CombineColumns(csvData, nil, targetData, []string{"y"})

	vals, ok := res.NumericTargetData["y"]
	if !ok {
		t.Fatal("NumericTargetData missing key 'y'")
	}
	if vals[0] != 10.0 || vals[1] != 20.0 {
		t.Errorf("NumericTargetData[y]: got %v", vals)
	}
}
