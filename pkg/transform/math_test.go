// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

const tol = 1e-9

// tolFmt is the tolerance for values that have been formatted with %.6g and
// parsed back — six significant figures limits precision to ~1e-5 for values
// of order 1.
const tolFmt = 1e-4

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func parseResult(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		t.Fatalf("parseResult: cannot parse %q: %v", s, err)
	}
	return v
}

func makeInput(data [][]string, headers []string, colTypes map[string]string) Input {
	return Input{
		Data:               data,
		Headers:            headers,
		ColumnTypes:        colTypes,
		CategoricalColumns: map[string][]string{},
		Rows:               len(data),
		Columns:            len(headers),
	}
}

// ─── Log transform ───────────────────────────────────────────────────────────

func TestApply_Log_Basic(t *testing.T) {
	in := makeInput(
		[][]string{{"1"}, {"math.E"}, {"10"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)
	// Use plain numbers
	in.Data = [][]string{{"1"}, {"2.718281828"}, {"10"}}

	res, err := Apply(in, Options{Type: Log, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply log: %v", err)
	}

	if len(res.TransformedColumns) != 1 || res.TransformedColumns[0] != "X" {
		t.Errorf("expected TransformedColumns=[X], got %v", res.TransformedColumns)
	}

	// log(1) == 0
	v0 := parseResult(t, res.Data[0][0])
	if !almostEqual(v0, 0.0, tol) {
		t.Errorf("log(1): expected 0, got %v", v0)
	}

	// log(e) ≈ 1
	v1 := parseResult(t, res.Data[1][0])
	if !almostEqual(v1, 1.0, 1e-6) {
		t.Errorf("log(e): expected ~1, got %v", v1)
	}

	// log(10) ≈ 2.302585... (tolFmt accounts for %.6g rounding)
	v2 := parseResult(t, res.Data[2][0])
	if !almostEqual(v2, math.Log(10), tolFmt) {
		t.Errorf("log(10): expected %v, got %v", math.Log(10), v2)
	}
}

func TestApply_Log_NonPositive(t *testing.T) {
	in := makeInput(
		[][]string{{"0"}, {"-1"}, {"4"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Log, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Row 0 (value 0) and row 1 (value -1) should be skipped (messages added).
	// Row 2 should be transformed.
	v2 := parseResult(t, res.Data[2][0])
	if !almostEqual(v2, math.Log(4), tolFmt) {
		t.Errorf("log(4): expected %v, got %v", math.Log(4), v2)
	}

	warnings := 0
	for _, m := range res.Messages {
		if strings.Contains(m, "Warning") {
			warnings++
		}
	}
	if warnings < 2 {
		t.Errorf("expected at least 2 warning messages for non-positive values, got %d", warnings)
	}
}

// ─── Sqrt transform ──────────────────────────────────────────────────────────

func TestApply_Sqrt_Basic(t *testing.T) {
	in := makeInput(
		[][]string{{"4"}, {"9"}, {"0"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Sqrt, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply sqrt: %v", err)
	}

	expected := []float64{2.0, 3.0, 0.0}
	for i, exp := range expected {
		got := parseResult(t, res.Data[i][0])
		if !almostEqual(got, exp, tol) {
			t.Errorf("sqrt row %d: expected %v, got %v", i, exp, got)
		}
	}
}

func TestApply_Sqrt_Negative(t *testing.T) {
	in := makeInput(
		[][]string{{"-1"}, {"4"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Sqrt, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Row 0 skipped, row 1 transformed.
	v := parseResult(t, res.Data[1][0])
	if !almostEqual(v, 2.0, tol) {
		t.Errorf("sqrt(4): expected 2, got %v", v)
	}

	hasWarning := false
	for _, m := range res.Messages {
		if strings.Contains(m, "Warning") {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected warning for negative value")
	}
}

// ─── Square transform ────────────────────────────────────────────────────────

func TestApply_Square_Basic(t *testing.T) {
	in := makeInput(
		[][]string{{"2"}, {"3"}, {"-4"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Square, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply square: %v", err)
	}

	expected := []float64{4.0, 9.0, 16.0}
	for i, exp := range expected {
		got := parseResult(t, res.Data[i][0])
		if !almostEqual(got, exp, tol) {
			t.Errorf("square row %d: expected %v, got %v", i, exp, got)
		}
	}
}

// ─── Column not found / non-numeric ──────────────────────────────────────────

func TestApply_Math_ColumnNotFound(t *testing.T) {
	in := makeInput(
		[][]string{{"1"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Log, Columns: []string{"NoSuchColumn"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Errorf("expected no transformed columns, got %v", res.TransformedColumns)
	}
	if len(res.Messages) == 0 {
		t.Error("expected a message for missing column")
	}
}

func TestApply_Math_NonNumericColumn(t *testing.T) {
	in := makeInput(
		[][]string{{"cat"}},
		[]string{"X"},
		map[string]string{"X": "categorical"},
	)

	res, err := Apply(in, Options{Type: Sqrt, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.TransformedColumns) != 0 {
		t.Errorf("expected no transformed columns for categorical, got %v", res.TransformedColumns)
	}
}

// ─── Unsupported type ─────────────────────────────────────────────────────────

func TestApply_UnsupportedType(t *testing.T) {
	in := makeInput(
		[][]string{{"1"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	_, err := Apply(in, Options{Type: "bogus", Columns: []string{"X"}})
	if err == nil {
		t.Error("expected error for unsupported transform type")
	}
}

// ─── Empty data ───────────────────────────────────────────────────────────────

func TestApply_EmptyData(t *testing.T) {
	in := makeInput(nil, []string{"X"}, map[string]string{"X": "numeric"})
	_, err := Apply(in, Options{Type: Log, Columns: []string{"X"}})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// ─── Input immutability ───────────────────────────────────────────────────────

func TestApply_DoesNotMutateInput(t *testing.T) {
	original := [][]string{{"4"}, {"9"}}
	in := makeInput(original, []string{"X"}, map[string]string{"X": "numeric"})

	_, err := Apply(in, Options{Type: Sqrt, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if original[0][0] != "4" || original[1][0] != "9" {
		t.Error("Apply must not mutate the original input data")
	}
}
