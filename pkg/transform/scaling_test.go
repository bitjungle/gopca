// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"math"
	"testing"
)

// ─── Standardize ─────────────────────────────────────────────────────────────

func TestApply_Standardize_MeanAndStd(t *testing.T) {
	// Values: 1, 2, 3, 4, 5 → mean=3, std=sqrt(2) ≈ 1.414
	in := makeInput(
		[][]string{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Standardize, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply standardize: %v", err)
	}

	// Compute mean and std of outputs — should be ≈0 and ≈1.
	sum, sum2 := 0.0, 0.0
	for _, row := range res.Data {
		v := parseResult(t, row[0])
		sum += v
		sum2 += v * v
	}
	n := float64(len(res.Data))
	mean := sum / n
	stdDev := math.Sqrt(sum2/n - mean*mean)

	// tolFmt accounts for %.6g rounding when the values were stored as strings.
	if !almostEqual(mean, 0.0, 1e-4) {
		t.Errorf("standardized mean: expected 0, got %v", mean)
	}
	if !almostEqual(stdDev, 1.0, 1e-4) {
		t.Errorf("standardized std: expected 1, got %v", stdDev)
	}
}

func TestApply_Standardize_ZeroVariance(t *testing.T) {
	// All values identical → zero variance, cannot standardize.
	in := makeInput(
		[][]string{{"5"}, {"5"}, {"5"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Standardize, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns for zero-variance column")
	}
	if len(res.Messages) == 0 {
		t.Error("expected message for zero variance")
	}
}

func TestApply_Standardize_InsufficientValues(t *testing.T) {
	in := makeInput(
		[][]string{{"5"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Standardize, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns for single-value column")
	}
}

// ─── MinMax ──────────────────────────────────────────────────────────────────

func TestApply_MinMax_DefaultRange(t *testing.T) {
	// Values: 0, 5, 10 → scaled to [0, 1]: 0, 0.5, 1
	in := makeInput(
		[][]string{{"0"}, {"5"}, {"10"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: MinMax, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply minmax: %v", err)
	}

	expected := []float64{0.0, 0.5, 1.0}
	for i, exp := range expected {
		got := parseResult(t, res.Data[i][0])
		if !almostEqual(got, exp, tol) {
			t.Errorf("minmax row %d: expected %v, got %v", i, exp, got)
		}
	}
}

func TestApply_MinMax_CustomRange(t *testing.T) {
	// Values: 0, 5, 10 → scaled to [-1, 1]: -1, 0, 1
	in := makeInput(
		[][]string{{"0"}, {"5"}, {"10"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: MinMax, Columns: []string{"X"}, MinValue: -1.0, MaxValue: 1.0})
	if err != nil {
		t.Fatalf("Apply minmax custom range: %v", err)
	}

	expected := []float64{-1.0, 0.0, 1.0}
	for i, exp := range expected {
		got := parseResult(t, res.Data[i][0])
		if !almostEqual(got, exp, tol) {
			t.Errorf("minmax custom row %d: expected %v, got %v", i, exp, got)
		}
	}
}

func TestApply_MinMax_ConstantColumn(t *testing.T) {
	// All values identical → cannot scale.
	in := makeInput(
		[][]string{{"7"}, {"7"}, {"7"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: MinMax, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns for constant column")
	}
}

func TestApply_MinMax_AllInRange(t *testing.T) {
	// After scaling, all output values must be in [targetMin, targetMax].
	in := makeInput(
		[][]string{{"3"}, {"1"}, {"4"}, {"1"}, {"5"}, {"9"}, {"2"}, {"6"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: MinMax, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, row := range res.Data {
		v := parseResult(t, row[0])
		if v < -tol || v > 1+tol {
			t.Errorf("row %d: scaled value %v out of [0, 1]", i, v)
		}
	}
}

// ─── GetTransformableColumns ──────────────────────────────────────────────────

func TestGetTransformableColumns_Numeric(t *testing.T) {
	in := makeInput(
		[][]string{{"1", "a", "2"}},
		[]string{"N1", "Cat", "N2"},
		map[string]string{"N1": "numeric", "Cat": "categorical", "N2": "numeric"},
	)

	cols := GetTransformableColumns(in, Standardize)
	if len(cols) != 2 {
		t.Errorf("expected 2 numeric columns, got %v", cols)
	}
}

func TestGetTransformableColumns_ExcludesTarget(t *testing.T) {
	in := makeInput(
		[][]string{{"1", "2"}},
		[]string{"X", "Y#target"},
		map[string]string{"X": "numeric", "Y#target": "numeric"},
	)

	cols := GetTransformableColumns(in, Log)
	if len(cols) != 1 || cols[0] != "X" {
		t.Errorf("expected only [X], got %v", cols)
	}
}

func TestGetTransformableColumns_OneHot(t *testing.T) {
	in := makeInput(
		[][]string{{"1", "a"}},
		[]string{"N", "C"},
		map[string]string{"N": "numeric", "C": "categorical"},
	)

	cols := GetTransformableColumns(in, OneHot)
	if len(cols) != 1 || cols[0] != "C" {
		t.Errorf("expected only [C] for OneHot, got %v", cols)
	}
}
