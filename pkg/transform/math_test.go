// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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
	// This test used to assert that rows 0 and 1 were skipped and row 2 was
	// transformed. That was the defect (#861): the skip left the original
	// number in the cell, so the column came out with some values logged and
	// some raw -- one variable carrying two different units into PCA.
	//
	// The contract now is all-or-nothing per column.
	in := makeInput(
		[][]string{{"0"}, {"-1"}, {"4"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Log, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, want := range []string{"0", "-1", "4"} {
		if res.Data[i][0] != want {
			t.Errorf("row %d = %q, want %q: the column must be left untouched",
				i, res.Data[i][0], want)
		}
	}

	if len(res.TransformedColumns) != 0 {
		t.Errorf("a refused column must not be reported as transformed, got %v",
			res.TransformedColumns)
	}

	// The message has to say what is wrong, where, and that nothing changed.
	joined := strings.Join(res.Messages, " ")
	for _, want := range []string{"left unchanged", "non-positive", "rows 1, 2"} {
		if !strings.Contains(joined, want) {
			t.Errorf("message should mention %q, got %v", want, res.Messages)
		}
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
	// As with log above, this asserted the mixed-column behaviour that #861
	// was filed about.
	in := makeInput(
		[][]string{{"-1"}, {"4"}},
		[]string{"X"},
		map[string]string{"X": "numeric"},
	)

	res, err := Apply(in, Options{Type: Sqrt, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, want := range []string{"-1", "4"} {
		if res.Data[i][0] != want {
			t.Errorf("row %d = %q, want %q: the column must be left untouched",
				i, res.Data[i][0], want)
		}
	}
	if len(res.TransformedColumns) != 0 {
		t.Errorf("a refused column must not be reported as transformed, got %v",
			res.TransformedColumns)
	}
	joined := strings.Join(res.Messages, " ")
	if !strings.Contains(joined, "negative") || !strings.Contains(joined, "left unchanged") {
		t.Errorf("message should name the problem, got %v", res.Messages)
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

// TestApply_Math_AllOrNothing states the contract #861 established: a column is
// either fully transformed or not touched at all, never half.
//
// The failure this prevents is specific. Skipping an impossible value left the
// original number in place, so a concentration column containing zeros came out
// with some cells in log units and some in the original units. Nothing
// downstream can detect that -- it is a plausible-looking column of numbers --
// and it goes into PCA as one variable.
func TestApply_Math_AllOrNothing(t *testing.T) {
	tests := []struct {
		name      string
		transform Type
		values    []string
		wantTouch bool
	}{
		{
			name:      "log refuses a column containing a zero",
			transform: Log, values: []string{"1", "0", "3"}, wantTouch: false,
		},
		{
			name:      "log refuses a column containing a negative",
			transform: Log, values: []string{"1", "-2", "3"}, wantTouch: false,
		},
		{
			name:      "log transforms a wholly positive column",
			transform: Log, values: []string{"1", "2", "3"}, wantTouch: true,
		},
		{
			name:      "sqrt accepts zero",
			transform: Sqrt, values: []string{"0", "4", "9"}, wantTouch: true,
		},
		{
			name:      "sqrt refuses a negative",
			transform: Sqrt, values: []string{"1", "-4"}, wantTouch: false,
		},
		{
			name:      "square accepts anything",
			transform: Square, values: []string{"-2", "0", "3"}, wantTouch: true,
		},
		{
			// Blanks are not values in the column's units, so they neither
			// block the transform nor get one.
			name:      "blanks do not block the transform",
			transform: Log, values: []string{"1", "", "3"}, wantTouch: true,
		},
		{
			// Same reasoning: a cell that was never a number cannot be put into
			// the wrong units by being left alone.
			name:      "unparseable cells do not block the transform",
			transform: Log, values: []string{"1", "N/A", "3"}, wantTouch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := make([][]string, len(tt.values))
			for i, v := range tt.values {
				rows[i] = []string{v}
			}
			in := makeInput(rows, []string{"X"}, map[string]string{"X": "numeric"})

			res, err := Apply(in, Options{Type: tt.transform, Columns: []string{"X"}})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			changed := false
			for i, original := range tt.values {
				if res.Data[i][0] != original {
					changed = true
				}
			}

			if changed != tt.wantTouch {
				t.Errorf("column changed = %v, want %v; result was %v",
					changed, tt.wantTouch, res.Data)
			}

			// Whatever the outcome, no cell may keep its original value while a
			// sibling was transformed. That is the mixed state itself.
			if !tt.wantTouch {
				for i, original := range tt.values {
					if res.Data[i][0] != original {
						t.Errorf("row %d changed to %q in a refused column",
							i, res.Data[i][0])
					}
				}
			}
		})
	}
}

// TestApply_Math_RefusalLeavesOtherColumnsAlone checks the refusal is scoped to
// the offending column, not the whole operation.
func TestApply_Math_RefusalLeavesOtherColumnsAlone(t *testing.T) {
	in := makeInput(
		[][]string{{"1", "1"}, {"0", "2"}, {"3", "3"}},
		[]string{"Bad", "Good"},
		map[string]string{"Bad": "numeric", "Good": "numeric"},
	)

	res, err := Apply(in, Options{Type: Log, Columns: []string{"Bad", "Good"}})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	for i, want := range []string{"1", "0", "3"} {
		if res.Data[i][0] != want {
			t.Errorf("'Bad' row %d = %q, want %q", i, res.Data[i][0], want)
		}
	}
	if res.Data[0][1] == "1" {
		t.Error("'Good' should have been transformed despite 'Bad' being refused")
	}
	if len(res.TransformedColumns) != 1 || res.TransformedColumns[0] != "Good" {
		t.Errorf("TransformedColumns = %v, want [Good]", res.TransformedColumns)
	}
}
