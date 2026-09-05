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

package main

import (
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestFilterMatches(t *testing.T) {
	tests := []struct {
		name     string
		operator FilterOperator
		value    string
		cell     string
		want     bool
		why      string
	}{
		{name: "equals text", operator: FilterEquals, value: "Nord", cell: "Nord", want: true},
		{name: "equals is case-insensitive", operator: FilterEquals, value: "nord", cell: "Nord", want: true,
			why: "filtering is a search; requiring the user to match case would be a trap"},
		{name: "equals compares numbers numerically", operator: FilterEquals, value: "5.0", cell: "5", want: true,
			why: "5 and 5.0 are the same measurement"},
		{name: "not equals", operator: FilterNotEquals, value: "Nord", cell: "Sor", want: true},
		{name: "contains", operator: FilterContains, value: "or", cell: "Nord", want: true},
		{name: "not contains", operator: FilterNotContains, value: "zz", cell: "Nord", want: true},
		{name: "greater", operator: FilterGreater, value: "5", cell: "6", want: true},
		{name: "greater is strict", operator: FilterGreater, value: "5", cell: "5", want: false},
		{name: "greater or equal", operator: FilterGreaterEqual, value: "5", cell: "5", want: true},
		{name: "less", operator: FilterLess, value: "5", cell: "4", want: true},
		{name: "less or equal", operator: FilterLessEqual, value: "5", cell: "5", want: true},

		{name: "ordering does not fall back to text", operator: FilterGreater, value: "5", cell: "apple", want: false,
			why: "string ordering would put \"10\" before \"9\", which is not what > means"},

		// The empty-cell rule.
		{name: "is empty matches a blank", operator: FilterIsEmpty, value: "", cell: "", want: true},
		{name: "is empty matches whitespace", operator: FilterIsEmpty, value: "", cell: "   ", want: true},
		{name: "is empty does not match a value", operator: FilterIsEmpty, value: "", cell: "Nord", want: false},
		{name: "is not empty", operator: FilterIsNotEmpty, value: "", cell: "Nord", want: true},
		{name: "is not empty excludes a blank", operator: FilterIsNotEmpty, value: "", cell: "", want: false},
		{
			name: "a blank is not \"not equal to\" anything", operator: FilterNotEquals, value: "Nord", cell: "",
			want: false,
			why: "otherwise \"remove rows where Region is not Nord\" quietly takes the " +
				"unlabelled rows too -- deciding a row's fate on the absence of a value",
		},
		{name: "a blank does not contain anything", operator: FilterNotContains, value: "zz", cell: "", want: false},
		{name: "a blank is not greater than anything", operator: FilterGreater, value: "-99", cell: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := FilterCondition{Operator: tt.operator, Value: tt.value}
			if got := c.matches(tt.cell); got != tt.want {
				t.Errorf("matches(%q) = %v, want %v. %s", tt.cell, got, tt.want, tt.why)
			}
		})
	}
}

func filterFixture() *FileData {
	return &FileData{
		Headers:  []string{"Region", "Score"},
		RowNames: []string{"P1", "P2", "P3", "P4"},
		Data: [][]string{
			{"Nord", "10"},
			{"Sor", "20"},
			{"Nord", "30"},
			{"", "40"},
		},
		Rows:        4,
		Columns:     2,
		ColumnTypes: map[string]string{"Region": "categorical", "Score": "numeric"},
		CategoricalColumns: map[string][]string{
			"Region": {"Nord", "Sor", "Nord", ""},
		},
		NumericTargetColumns: map[string][]types.JSONFloat64{
			"Score": {10, 20, 30, 40},
		},
	}
}

func TestFilterRowsKeepAndRemove(t *testing.T) {
	tests := []struct {
		name      string
		condition FilterCondition
		wantNames string
	}{
		{
			name:      "keep matching",
			condition: FilterCondition{Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "keep"},
			wantNames: "P1,P3",
		},
		{
			name:      "remove matching",
			condition: FilterCondition{Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "remove"},
			wantNames: "P2,P4",
		},
		{
			name:      "numeric comparison",
			condition: FilterCondition{Column: "Score", Operator: FilterGreater, Value: "15", Mode: "keep"},
			wantNames: "P2,P3,P4",
		},
		{
			name:      "drop the unlabelled rows",
			condition: FilterCondition{Column: "Region", Operator: FilterIsEmpty, Value: "", Mode: "remove"},
			wantNames: "P1,P2,P3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := filterFixture()
			cmd, err := NewFilterRowsCommand(&App{}, data, tt.condition)
			if err != nil {
				t.Fatalf("NewFilterRowsCommand: %v", err)
			}
			if err := cmd.Execute(data); err != nil {
				t.Fatalf("Execute: %v", err)
			}

			if got := strings.Join(data.RowNames, ","); got != tt.wantNames {
				t.Errorf("rows kept = %s, want %s", got, tt.wantNames)
			}
			if data.Rows != len(data.Data) {
				t.Errorf("Rows = %d but there are %d data rows", data.Rows, len(data.Data))
			}
		})
	}
}

// TestFilterKeepsPerRowMapsAligned is the check that would have caught the
// quietest failure here.
//
// CategoricalColumns and NumericTargetColumns hold one entry per row, parallel
// to Data. Dropping rows without dropping the matching entries leaves them
// longer than the table and misaligned from the first deletion onwards -- every
// value attached to the wrong row, with nothing to show for it.
func TestFilterKeepsPerRowMapsAligned(t *testing.T) {
	data := filterFixture()
	cmd, _ := NewFilterRowsCommand(&App{}, data,
		FilterCondition{Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "keep"})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := len(data.CategoricalColumns["Region"]); got != len(data.Data) {
		t.Errorf("CategoricalColumns has %d entries for %d rows", got, len(data.Data))
	}
	if got := strings.Join(data.CategoricalColumns["Region"], ","); got != "Nord,Nord" {
		t.Errorf("CategoricalColumns[Region] = %q, want \"Nord,Nord\"", got)
	}
	if got := len(data.NumericTargetColumns["Score"]); got != len(data.Data) {
		t.Errorf("NumericTargetColumns has %d entries for %d rows", got, len(data.Data))
	}
	if data.NumericTargetColumns["Score"][0] != 10 || data.NumericTargetColumns["Score"][1] != 30 {
		t.Errorf("target values misaligned: %v", data.NumericTargetColumns["Score"])
	}
}

func TestFilterUndo(t *testing.T) {
	data := filterFixture()
	cmd, _ := NewFilterRowsCommand(&App{}, data,
		FilterCondition{Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "keep"})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if len(data.Data) != 4 || strings.Join(data.RowNames, ",") != "P1,P2,P3,P4" {
		t.Errorf("undo did not restore the rows: %v", data.RowNames)
	}
	if len(data.CategoricalColumns["Region"]) != 4 {
		t.Errorf("undo did not restore the per-row maps: %v", data.CategoricalColumns)
	}
	if len(data.NumericTargetColumns["Score"]) != 4 {
		t.Errorf("undo did not restore the target values: %v", data.NumericTargetColumns)
	}
}

func TestPreviewFilter(t *testing.T) {
	app := &App{}
	data := filterFixture()

	keep := app.PreviewFilter(data, FilterCondition{
		Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "keep"})
	if keep.Matched != 2 || keep.Total != 4 || keep.Remaining != 2 {
		t.Errorf("keep preview = %+v, want 2 of 4 leaving 2", keep)
	}

	remove := app.PreviewFilter(data, FilterCondition{
		Column: "Region", Operator: FilterEquals, Value: "Nord", Mode: "remove"})
	if remove.Matched != 2 || remove.Remaining != 2 {
		t.Errorf("remove preview = %+v, want 2 matched leaving 2", remove)
	}

	// A filter that would empty the table must be visible before it is applied.
	empty := app.PreviewFilter(data, FilterCondition{
		Column: "Region", Operator: FilterEquals, Value: "nowhere", Mode: "keep"})
	if empty.Remaining != 0 {
		t.Errorf("preview = %+v, want 0 remaining", empty)
	}

	// The preview must not modify anything.
	if len(data.Data) != 4 {
		t.Errorf("PreviewFilter changed the data: %d rows left", len(data.Data))
	}
}

func TestFilterValidation(t *testing.T) {
	app := &App{}
	data := filterFixture()

	tests := []struct {
		name      string
		condition FilterCondition
		wantErr   string
	}{
		{
			name:      "unknown column",
			condition: FilterCondition{Column: "Nope", Operator: FilterEquals, Value: "x", Mode: "keep"},
			wantErr:   "no column named",
		},
		{
			name:      "unknown operator",
			condition: FilterCondition{Column: "Region", Operator: "sideways", Value: "x", Mode: "keep"},
			wantErr:   "unknown operator",
		},
		{
			name:      "unknown mode",
			condition: FilterCondition{Column: "Region", Operator: FilterEquals, Value: "x", Mode: "maybe"},
			wantErr:   "mode must be",
		},
		{
			name:      "ordering against a non-number",
			condition: FilterCondition{Column: "Score", Operator: FilterGreater, Value: "many", Mode: "keep"},
			wantErr:   "is not a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.PreviewFilter(data, tt.condition); !strings.Contains(got.Error, tt.wantErr) {
				t.Errorf("preview error = %q, want it to mention %q", got.Error, tt.wantErr)
			}
			if _, err := app.ExecuteFilterRows(data, tt.condition); err == nil {
				t.Error("the command accepted a condition the preview rejected")
			}
			if len(data.Data) != 4 {
				t.Errorf("a refused filter modified the data: %d rows", len(data.Data))
			}
		})
	}
}
