// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"fmt"
	"strings"
	"testing"
)

// ─── Binning ──────────────────────────────────────────────────────────────────

func TestApply_Bin_DefaultBinCount(t *testing.T) {
	// 10 values uniformly spaced → 5 bins (default).
	data := make([][]string, 10)
	for i := range data {
		data[i] = []string{fmt.Sprintf("%d", i+1)} // 1..10
	}
	in := Input{
		Data:               data,
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               10,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply bin: %v", err)
	}

	// Count distinct bin labels produced.
	bins := map[string]int{}
	for _, row := range res.Data {
		bins[row[0]]++
	}
	if len(bins) != 5 {
		t.Errorf("expected 5 distinct bins, got %d: %v", len(bins), bins)
	}

	// Column type should now be categorical.
	if res.ColumnTypes["X"] != "categorical" {
		t.Errorf("expected column type 'categorical' after binning, got %q", res.ColumnTypes["X"])
	}

	// CategoricalColumns should contain "X".
	if _, ok := res.CategoricalColumns["X"]; !ok {
		t.Error("expected 'X' to be registered in CategoricalColumns after binning")
	}
}

func TestApply_Bin_CustomBinCount(t *testing.T) {
	// 9 values 1..9 → 3 bins.
	data := make([][]string, 9)
	for i := range data {
		data[i] = []string{fmt.Sprintf("%d", i+1)}
	}
	in := Input{
		Data:               data,
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               9,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}, BinCount: 3})
	if err != nil {
		t.Fatalf("Apply bin custom count: %v", err)
	}

	bins := map[string]int{}
	for _, row := range res.Data {
		bins[row[0]]++
	}
	if len(bins) != 3 {
		t.Errorf("expected 3 bins, got %d: %v", len(bins), bins)
	}
}

func TestApply_Bin_AllValuesAssigned(t *testing.T) {
	// Every row should receive a bin label (not remain numeric or empty).
	data := [][]string{{"1"}, {"5"}, {"10"}, {"3"}, {"7"}, {"2"}}
	in := Input{
		Data:               data,
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               len(data),
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}, BinCount: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, row := range res.Data {
		if !strings.HasPrefix(row[0], "Bin_") {
			t.Errorf("row %d: expected Bin_* label, got %q", i, row[0])
		}
	}
}

func TestApply_Bin_CorrectBinBoundaries(t *testing.T) {
	// Values: 0, 3, 6, 9 with binCount=3 → bin width=3.
	// 0 → Bin_1, 3 → Bin_2 (or Bin_1 boundary), 6 → Bin_3 (or Bin_2), 9 → Bin_3
	data := [][]string{{"0"}, {"3"}, {"6"}, {"9"}}
	in := Input{
		Data:               data,
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               4,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}, BinCount: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The last value (9) hits exactly binIndex==3 which should be clamped to Bin_3.
	lastBin := res.Data[3][0]
	if lastBin != "Bin_3" {
		t.Errorf("expected last value to land in Bin_3 (clamped), got %q", lastBin)
	}
}

func TestApply_Bin_NonNumericColumn(t *testing.T) {
	in := Input{
		Data:               [][]string{{"cat"}},
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "categorical"},
		CategoricalColumns: map[string][]string{},
		Rows:               1,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns for categorical column")
	}
}

func TestApply_Bin_ConstantColumn(t *testing.T) {
	// All values identical — binWidth would be 0 → must be handled gracefully.
	in := Input{
		Data:               [][]string{{"5"}, {"5"}, {"5"}},
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               3,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns for constant-value column")
	}
	if len(res.Messages) == 0 {
		t.Error("expected message explaining why binning was skipped")
	}
}

func TestApply_Bin_NoNumericValues(t *testing.T) {
	in := Input{
		Data:               [][]string{{""}},
		Headers:            []string{"X"},
		ColumnTypes:        map[string]string{"X": "numeric"},
		CategoricalColumns: map[string][]string{},
		Rows:               1,
		Columns:            1,
	}

	res, err := Apply(in, Options{Type: Bin, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.TransformedColumns) != 0 {
		t.Error("expected no transformed columns when no numeric values")
	}
	if len(res.Messages) == 0 {
		t.Error("expected message for empty column")
	}
}
