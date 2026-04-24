// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

import (
	"strconv"
	"testing"
)

// ─── AnalyzeMissing ──────────────────────────────────────────────────────────

func TestAnalyzeMissing_NoMissing(t *testing.T) {
	data := [][]string{
		{"1.0", "a"},
		{"2.0", "b"},
		{"3.0", "c"},
	}
	headers := []string{"X", "Y"}
	stats := AnalyzeMissing(data, headers)

	if stats.MissingCells != 0 {
		t.Errorf("expected 0 missing cells, got %d", stats.MissingCells)
	}
	if stats.MissingPercent != 0 {
		t.Errorf("expected 0%% missing, got %v", stats.MissingPercent)
	}
	for _, h := range headers {
		if stats.ColumnStats[h].Pattern != "none" {
			t.Errorf("column %q: expected pattern 'none', got %q", h, stats.ColumnStats[h].Pattern)
		}
	}
}

func TestAnalyzeMissing_AllMissing(t *testing.T) {
	data := [][]string{
		{"", ""},
		{"NA", "N/A"},
	}
	headers := []string{"A", "B"}
	stats := AnalyzeMissing(data, headers)

	if stats.MissingCells != 4 {
		t.Errorf("expected 4 missing cells, got %d", stats.MissingCells)
	}
	if stats.MissingPercent != 100.0 {
		t.Errorf("expected 100%% missing, got %v", stats.MissingPercent)
	}
}

func TestAnalyzeMissing_PartialMissing(t *testing.T) {
	data := [][]string{
		{"1", ""},
		{"2", "b"},
	}
	headers := []string{"X", "Y"}
	stats := AnalyzeMissing(data, headers)

	if stats.MissingCells != 1 {
		t.Errorf("expected 1 missing cell, got %d", stats.MissingCells)
	}
	if stats.ColumnStats["Y"].MissingValues != 1 {
		t.Errorf("expected 1 missing in column Y, got %d", stats.ColumnStats["Y"].MissingValues)
	}
}

func TestAnalyzeMissing_EmptyData(t *testing.T) {
	stats := AnalyzeMissing(nil, []string{"A"})
	if stats.MissingCells != 0 || stats.TotalCells != 0 {
		t.Error("empty data should yield all-zero stats")
	}
}

func TestAnalyzeMissing_PatternTop(t *testing.T) {
	data := [][]string{
		{""},
		{""},
		{"3"},
		{"4"},
	}
	stats := AnalyzeMissing(data, []string{"X"})
	if stats.ColumnStats["X"].Pattern != "top" {
		t.Errorf("expected pattern 'top', got %q", stats.ColumnStats["X"].Pattern)
	}
}

func TestAnalyzeMissing_PatternBottom(t *testing.T) {
	data := [][]string{
		{"1"},
		{"2"},
		{""},
		{""},
	}
	stats := AnalyzeMissing(data, []string{"X"})
	if stats.ColumnStats["X"].Pattern != "bottom" {
		t.Errorf("expected pattern 'bottom', got %q", stats.ColumnStats["X"].Pattern)
	}
}

// ─── Fill strategies ─────────────────────────────────────────────────────────

func makeData(rows [][]string) [][]string {
	result := make([][]string, len(rows))
	for i, row := range rows {
		result[i] = make([]string, len(row))
		copy(result[i], row)
	}
	return result
}

func TestFill_Mean(t *testing.T) {
	data := [][]string{
		{"1"},
		{""},
		{"3"},
	}
	result, err := Fill(makeData(data), []string{"X"}, map[string]string{"X": "numeric"}, FillRequest{Strategy: "mean"})
	if err != nil {
		t.Fatalf("Fill mean: %v", err)
	}
	// mean of [1, 3] = 2
	got := result[1][0]
	v, _ := strconv.ParseFloat(got, 64)
	if !almostEqual(v, 2.0, 1e-9) {
		t.Errorf("mean fill: expected 2.0, got %s", got)
	}
	// Non-missing cells must be unchanged
	if result[0][0] != "1" || result[2][0] != "3" {
		t.Error("mean fill: existing values were modified")
	}
}

func TestFill_Median(t *testing.T) {
	data := [][]string{
		{"1"},
		{""},
		{"3"},
		{"5"},
	}
	result, err := Fill(makeData(data), []string{"X"}, map[string]string{"X": "numeric"}, FillRequest{Strategy: "median"})
	if err != nil {
		t.Fatalf("Fill median: %v", err)
	}
	// sorted [1,3,5], median = 3
	v, _ := strconv.ParseFloat(result[1][0], 64)
	if !almostEqual(v, 3.0, 1e-9) {
		t.Errorf("median fill: expected 3.0, got %s", result[1][0])
	}
}

func TestFill_Mode(t *testing.T) {
	data := [][]string{
		{"a"},
		{"b"},
		{"a"},
		{""},
	}
	result, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "mode"})
	if err != nil {
		t.Fatalf("Fill mode: %v", err)
	}
	// mode = "a" (appears twice)
	if result[3][0] != "a" {
		t.Errorf("mode fill: expected 'a', got %q", result[3][0])
	}
}

func TestFill_Forward(t *testing.T) {
	data := [][]string{
		{"10"},
		{""},
		{""},
		{"20"},
	}
	result, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "forward"})
	if err != nil {
		t.Fatalf("Fill forward: %v", err)
	}
	if result[1][0] != "10" {
		t.Errorf("forward fill row 1: expected '10', got %q", result[1][0])
	}
	if result[2][0] != "10" {
		t.Errorf("forward fill row 2: expected '10', got %q", result[2][0])
	}
}

func TestFill_Forward_FirstRowMissing(t *testing.T) {
	data := [][]string{
		{""},
		{"5"},
	}
	result, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "forward"})
	if err != nil {
		t.Fatalf("Fill forward: %v", err)
	}
	// No prior value — first row must remain missing
	if !isMissing(result[0][0]) {
		t.Errorf("forward fill: first-row missing should stay missing, got %q", result[0][0])
	}
}

func TestFill_Backward(t *testing.T) {
	data := [][]string{
		{""},
		{""},
		{"30"},
	}
	result, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "backward"})
	if err != nil {
		t.Fatalf("Fill backward: %v", err)
	}
	if result[0][0] != "30" {
		t.Errorf("backward fill row 0: expected '30', got %q", result[0][0])
	}
	if result[1][0] != "30" {
		t.Errorf("backward fill row 1: expected '30', got %q", result[1][0])
	}
}

func TestFill_Custom(t *testing.T) {
	data := [][]string{
		{""},
		{"b"},
		{""},
	}
	result, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "custom", Value: "FILLED"})
	if err != nil {
		t.Fatalf("Fill custom: %v", err)
	}
	if result[0][0] != "FILLED" || result[2][0] != "FILLED" {
		t.Errorf("custom fill: expected 'FILLED', got %q / %q", result[0][0], result[2][0])
	}
	if result[1][0] != "b" {
		t.Error("custom fill: non-missing value was overwritten")
	}
}

func TestFill_AllMissing_Mean(t *testing.T) {
	// When every value is missing, mean fill has nothing to compute;
	// cells should remain missing (no panic).
	data := [][]string{{"", ""}, {"", ""}}
	_, err := Fill(makeData(data), []string{"A", "B"}, map[string]string{"A": "numeric", "B": "numeric"}, FillRequest{Strategy: "mean"})
	if err != nil {
		t.Fatalf("Fill mean all-missing: unexpected error: %v", err)
	}
}

func TestFill_SingleColumn(t *testing.T) {
	data := [][]string{{"1", "x"}, {"", "y"}, {"3", "z"}}
	headers := []string{"Num", "Cat"}
	result, err := Fill(makeData(data), headers, map[string]string{"Num": "numeric"}, FillRequest{Strategy: "mean", Column: "Num"})
	if err != nil {
		t.Fatalf("Fill single column: %v", err)
	}
	v, _ := strconv.ParseFloat(result[1][0], 64)
	if !almostEqual(v, 2.0, 1e-9) {
		t.Errorf("single column fill: expected 2.0, got %s", result[1][0])
	}
	// Cat column must be unchanged
	if result[0][1] != "x" || result[1][1] != "y" || result[2][1] != "z" {
		t.Error("single column fill: untargeted column was modified")
	}
}

func TestFill_UnknownColumn(t *testing.T) {
	data := [][]string{{"1"}}
	_, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "mean", Column: "NoSuchColumn"})
	if err == nil {
		t.Error("expected error for unknown column name, got nil")
	}
}

func TestFill_UnknownStrategy(t *testing.T) {
	data := [][]string{{"1"}}
	_, err := Fill(makeData(data), []string{"X"}, nil, FillRequest{Strategy: "interpolate"})
	if err == nil {
		t.Error("expected error for unknown strategy, got nil")
	}
}

func TestFill_OriginalDataUnmodified(t *testing.T) {
	original := [][]string{{"1"}, {""}, {"3"}}
	snapshot := [][]string{{"1"}, {""}, {"3"}}
	_, err := Fill(original, []string{"X"}, map[string]string{"X": "numeric"}, FillRequest{Strategy: "mean"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, row := range original {
		for j, v := range row {
			if v != snapshot[i][j] {
				t.Errorf("original data[%d][%d] was modified: expected %q, got %q", i, j, snapshot[i][j], v)
			}
		}
	}
}
