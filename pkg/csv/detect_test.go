// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"fmt"
	"testing"
)

// ─── DetectColumnType ─────────────────────────────────────────────────────────

func TestDetectColumnType_Numeric(t *testing.T) {
	data := [][]string{{"1.0"}, {"2.5"}, {"3"}, {"-4.1"}, {"0"}}
	if got := DetectColumnType(data, 0); got != "numeric" {
		t.Errorf("expected numeric, got %q", got)
	}
}

func TestDetectColumnType_NumericWithOneMissing(t *testing.T) {
	// 4 numeric + 1 empty → 4/4 non-empty = 100% numeric → "numeric"
	data := [][]string{{"1"}, {"2"}, {"3"}, {"4"}, {""}}
	if got := DetectColumnType(data, 0); got != "numeric" {
		t.Errorf("expected numeric, got %q", got)
	}
}

func TestDetectColumnType_NumericWith10PctNonNumeric(t *testing.T) {
	// 9 numeric + 1 text → 90% numeric → exactly NOT >0.9 → categorical
	rows := make([][]string, 10)
	for i := 0; i < 9; i++ {
		rows[i] = []string{fmt.Sprintf("%d", i+1)}
	}
	rows[9] = []string{"text"}
	got := DetectColumnType(rows, 0)
	// 9/10 = 0.9, which is NOT > 0.9, so falls through to categorical check.
	// 10 unique values out of 10 → ratio=1.0 ≥ 0.2 and count=10 < 20 → categorical.
	if got != "categorical" {
		t.Errorf("expected categorical for 90%% numeric, got %q", got)
	}
}

func TestDetectColumnType_Categorical_LowCardinality(t *testing.T) {
	// 20 rows, 3 unique values → ratio = 3/20 = 0.15 < 0.2 → categorical
	data := make([][]string, 20)
	for i := range data {
		data[i] = []string{[]string{"a", "b", "c"}[i%3]}
	}
	if got := DetectColumnType(data, 0); got != "categorical" {
		t.Errorf("expected categorical, got %q", got)
	}
}

func TestDetectColumnType_Categorical_FewUniqueValues(t *testing.T) {
	// 100 rows, 15 unique values → count < 20 → categorical
	data := make([][]string, 100)
	for i := range data {
		data[i] = []string{fmt.Sprintf("val%d", i%15)}
	}
	if got := DetectColumnType(data, 0); got != "categorical" {
		t.Errorf("expected categorical for <20 unique values, got %q", got)
	}
}

func TestDetectColumnType_Text(t *testing.T) {
	// 100 rows, 50 unique non-numeric values → ratio=0.5 ≥ 0.2, count=50 ≥ 20 → text
	data := make([][]string, 100)
	for i := range data {
		data[i] = []string{fmt.Sprintf("token_%d", i%50)}
	}
	if got := DetectColumnType(data, 0); got != "text" {
		t.Errorf("expected text, got %q", got)
	}
}

func TestDetectColumnType_AllEmpty(t *testing.T) {
	data := [][]string{{""}, {""}, {""}}
	if got := DetectColumnType(data, 0); got != "empty" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestDetectColumnType_EmptyData(t *testing.T) {
	if got := DetectColumnType(nil, 0); got != "unknown" {
		t.Errorf("expected unknown for nil data, got %q", got)
	}
	if got := DetectColumnType([][]string{}, 0); got != "unknown" {
		t.Errorf("expected unknown for empty data, got %q", got)
	}
}

func TestDetectColumnType_NegativeIndex(t *testing.T) {
	data := [][]string{{"1"}, {"2"}}
	if got := DetectColumnType(data, -1); got != "unknown" {
		t.Errorf("expected unknown for negative index, got %q", got)
	}
}

func TestDetectColumnType_IndexOutOfRange(t *testing.T) {
	// colIndex beyond all rows — all skipped → empty
	data := [][]string{{"1"}, {"2"}}
	if got := DetectColumnType(data, 5); got != "empty" {
		t.Errorf("expected empty for out-of-range index, got %q", got)
	}
}

func TestDetectColumnType_MultiColumn_SelectsCorrectColumn(t *testing.T) {
	// data has two columns: col0 is numeric, col1 is categorical
	data := [][]string{
		{"1.0", "red"},
		{"2.0", "blue"},
		{"3.0", "red"},
		{"4.0", "green"},
		{"5.0", "blue"},
	}
	if got := DetectColumnType(data, 0); got != "numeric" {
		t.Errorf("col0: expected numeric, got %q", got)
	}
	if got := DetectColumnType(data, 1); got != "categorical" {
		t.Errorf("col1: expected categorical, got %q", got)
	}
}

func TestDetectColumnType_NegativeNumbers(t *testing.T) {
	data := [][]string{{"-1"}, {"-2.5"}, {"-0.001"}}
	if got := DetectColumnType(data, 0); got != "numeric" {
		t.Errorf("expected numeric for negative numbers, got %q", got)
	}
}

func TestDetectColumnType_ScientificNotation(t *testing.T) {
	data := [][]string{{"1e5"}, {"2.3e-4"}, {"1.0E10"}}
	if got := DetectColumnType(data, 0); got != "numeric" {
		t.Errorf("expected numeric for scientific notation, got %q", got)
	}
}
