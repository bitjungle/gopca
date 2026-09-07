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

func clrFixture(rows ...[]string) Input {
	names := make([]string, len(rows[0]))
	types := map[string]string{}
	for j := range names {
		names[j] = string(rune('A' + j))
		types[names[j]] = "numeric"
	}
	data := make([][]string, len(rows))
	copy(data, rows)
	return Input{
		Data: data, Headers: names, ColumnTypes: types,
		Rows: len(rows), Columns: len(names),
	}
}

func clrColumn(t *testing.T, res *Result, name string) []float64 {
	t.Helper()
	index := findColumn(res.Headers, name)
	if index == -1 {
		t.Fatalf("no column %q in %v", name, res.Headers)
	}
	out := make([]float64, 0, len(res.Data))
	for _, row := range res.Data {
		value, err := strconv.ParseFloat(row[index], 64)
		if err != nil {
			t.Fatalf("parsing %q: %v", row[index], err)
		}
		out = append(out, value)
	}
	return out
}

// TestCLRRefusesZerosByDefault covers the decision at the heart of this
// transform.
//
// The logarithm is undefined at zero, and zeros are routine in trace-element
// data. There is a standard remedy, but it invents a measurement that was not
// made: the difference between "absent" and "below the detection limit" is a
// scientific judgement the software has no basis for. So the default is
// refusal, and the substitution happens only when asked for.
func TestCLRRefusesZerosByDefault(t *testing.T) {
	in := clrFixture([]string{"50", "30", "20"}, []string{"60", "0", "40"})

	_, err := Apply(in, Options{Type: CLR, Columns: []string{"A", "B", "C"}})
	if err == nil {
		t.Fatal("a composition containing a zero was transformed; ln(0) is undefined")
	}
	for _, want := range []string{"zero", "row 2", "detection limit"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should mention %q, got %v", want, err)
		}
	}

	// A refusal must leave the data untouched.
	if len(in.Headers) != 3 {
		t.Errorf("a refused transform modified the input: %v", in.Headers)
	}
}

// TestCLRZeroReplacementIsMultiplicative checks the substitution preserves the
// row total, which is what keeps the ratios among the observed parts intact.
//
// Simply writing the replacement into the cell would inflate the total and
// shift every other ratio in the row -- a silent change to parts that were
// measured, caused by one that was not.
func TestCLRZeroReplacementIsMultiplicative(t *testing.T) {
	in := clrFixture([]string{"60", "0", "40"})

	res, err := Apply(in, Options{
		Type: CLR, Columns: []string{"A", "B", "C"}, ZeroReplacement: 0.01})
	if err != nil {
		t.Fatalf("Apply CLR: %v", err)
	}

	// With a total of 100 and one zero replaced by 0.01, the other parts are
	// scaled by (1 - 0.01/100), so the row still totals 100.
	scale := 1 - 0.01/100.0
	want := []float64{60 * scale, 0.01, 40 * scale}
	logs := make([]float64, 3)
	mean := 0.0
	for i, v := range want {
		logs[i] = math.Log(v)
		mean += logs[i]
	}
	mean /= 3

	// Tolerance reflects the stored precision, not the arithmetic. Results are
	// written with twelve significant figures (clrPrecision), so a value near 3
	// round-trips to within about 3e-12; asserting tighter than that would be
	// testing the formatter rather than the transform.
	for j, colName := range []string{"A", "B", "C"} {
		got := clrColumn(t, res, colName+"_clr")[0]
		if diff := math.Abs(got - (logs[j] - mean)); diff > 1e-10 {
			t.Errorf("part %s: got %.12g, want %.12g", colName, got, logs[j]-mean)
		}
	}

	joined := strings.Join(res.Messages, " ")
	for _, want := range []string{"Replaced 1 zero", "multiplicative"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the substitution must be reported; %q missing from %v", want, res.Messages)
		}
	}
}

func TestCLRRefusals(t *testing.T) {
	tests := []struct {
		name    string
		rows    [][]string
		columns []string
		opts    func(*Options)
		wantErr string
	}{
		{
			name:    "negative part",
			rows:    [][]string{{"50", "-30", "80"}},
			columns: []string{"A", "B", "C"},
			wantErr: "negative",
		},
		{
			name:    "missing part",
			rows:    [][]string{{"50", "", "50"}},
			columns: []string{"A", "B", "C"},
			wantErr: "no value",
		},
		{
			name:    "non-numeric part",
			rows:    [][]string{{"50", "trace", "50"}},
			columns: []string{"A", "B", "C"},
			wantErr: "is not a number",
		},
		{
			name:    "one part is not a composition",
			rows:    [][]string{{"50", "30", "20"}},
			columns: []string{"A"},
			wantErr: "at least two parts",
		},
		{
			name:    "a part listed twice",
			rows:    [][]string{{"50", "30", "20"}},
			columns: []string{"A", "A", "B"},
			wantErr: "more than once",
		},
		{
			name:    "replacement larger than the row",
			rows:    [][]string{{"0", "0", "0.001"}},
			columns: []string{"A", "B", "C"},
			opts:    func(o *Options) { o.ZeroReplacement = 5 },
			wantErr: "too large",
		},
		{
			// A row of nothing but zeros has no composition to preserve, and
			// substituting every part invents a convincing one: every part
			// equal gives a clr of all zeros, which is exactly what a real
			// sample with equal parts looks like. Such a row would reach the
			// analysis indistinguishable from a measured one.
			name:    "a row with no measured parts at all",
			rows:    [][]string{{"50", "30", "20"}, {"0", "0", "0"}},
			columns: []string{"A", "B", "C"},
			opts:    func(o *Options) { o.ZeroReplacement = 0.01 },
			wantErr: "no measured parts at all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := clrFixture(tt.rows...)
			opts := Options{Type: CLR, Columns: tt.columns}
			if tt.opts != nil {
				tt.opts(&opts)
			}
			_, err := Apply(in, opts)
			if err == nil {
				t.Fatalf("expected a refusal mentioning %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q should mention %q", err, tt.wantErr)
			}
		})
	}
}

// TestCLRReportsClosure covers what the transform says about the selection.
//
// CLR does not require closure -- a subcomposition is still compositional --
// so a varying total is reported rather than refused. Saying which it is helps
// the user notice a selection that is not a composition at all.
func TestCLRReportsClosure(t *testing.T) {
	closed := clrFixture([]string{"50", "30", "20"}, []string{"10", "10", "80"})
	res, err := Apply(closed, Options{Type: CLR, Columns: []string{"A", "B", "C"}})
	if err != nil {
		t.Fatalf("Apply CLR: %v", err)
	}
	if !strings.Contains(strings.Join(res.Messages, " "), "closed composition") {
		t.Errorf("rows summing to 100 should be reported as closed: %v", res.Messages)
	}

	open := clrFixture([]string{"1", "2", "4"}, []string{"10", "5", "1"})
	res, err = Apply(open, Options{Type: CLR, Columns: []string{"A", "B", "C"}})
	if err != nil {
		t.Fatalf("Apply CLR on a subcomposition: %v", err)
	}
	if !strings.Contains(strings.Join(res.Messages, " "), "not closed") {
		t.Errorf("varying totals should be reported as not closed: %v", res.Messages)
	}
	if len(res.NewColumns) != 3 {
		t.Errorf("a subcomposition must still transform, got %v", res.NewColumns)
	}
}

func TestCLRKeepsAndRemovesSources(t *testing.T) {
	t.Run("kept by default", func(t *testing.T) {
		res, err := Apply(clrFixture([]string{"50", "30", "20"}),
			Options{Type: CLR, Columns: []string{"A", "B", "C"}})
		if err != nil {
			t.Fatalf("Apply CLR: %v", err)
		}
		for _, name := range []string{"A", "B", "C"} {
			if findColumn(res.Headers, name) == -1 {
				t.Errorf("source %q should be kept, headers are %v", name, res.Headers)
			}
		}
	})

	t.Run("removed on request", func(t *testing.T) {
		in := clrFixture([]string{"50", "30", "20", "9"})
		in.Headers = []string{"A", "B", "C", "Other"}
		in.ColumnTypes["Other"] = "numeric"
		in.Columns = 4

		res, err := Apply(in, Options{
			Type: CLR, Columns: []string{"A", "B", "C"}, RemoveOriginal: true})
		if err != nil {
			t.Fatalf("Apply CLR: %v", err)
		}
		for _, name := range []string{"A", "B", "C"} {
			if findColumn(res.Headers, name) != -1 {
				t.Errorf("source %q should have been removed: %v", name, res.Headers)
			}
		}
		// The untouched column and the results must survive intact.
		if got := clrColumn(t, res, "Other")[0]; got != 9 {
			t.Errorf("an unrelated column was damaged: %v", got)
		}
		for i, row := range res.Data {
			if len(row) != len(res.Headers) {
				t.Errorf("row %d has %d cells but %d headers", i, len(row), len(res.Headers))
			}
		}
	})
}
