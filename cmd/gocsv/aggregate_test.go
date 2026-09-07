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

// replicateFixture is two samples measured twice each, which is the case this
// feature exists for.
func replicateFixture() *FileData {
	return &FileData{
		Headers:  []string{"SampleID", "Operator", "Abs1", "Abs2"},
		RowNames: []string{"r1", "r2", "r3", "r4"},
		Data: [][]string{
			{"S1", "AB", "10", "2"},
			{"S1", "AB", "20", "4"},
			{"S2", "CD", "30", "6"},
			{"S2", "CD", "50", "10"},
		},
		Rows:    4,
		Columns: 4,
		ColumnTypes: map[string]string{
			"SampleID": "categorical", "Operator": "categorical",
			"Abs1": "numeric", "Abs2": "numeric",
		},
		CategoricalColumns: map[string][]string{
			"SampleID": {"S1", "S1", "S2", "S2"},
			"Operator": {"AB", "AB", "CD", "CD"},
		},
	}
}

func cellsOf(t *testing.T, data *FileData, column string) []string {
	t.Helper()
	index := findColumnIndex(data.Headers, column)
	if index == -1 {
		t.Fatalf("no column %q in %v", column, data.Headers)
	}
	out := make([]string, 0, len(data.Data))
	for _, row := range data.Data {
		out = append(out, row[index])
	}
	return out
}

// TestAggregateAveragesReplicates covers the central case.
func TestAggregateAveragesReplicates(t *testing.T) {
	data := replicateFixture()
	cmd, err := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err != nil {
		t.Fatalf("NewAggregateRowsCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if data.Rows != 2 || len(data.Data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data.Data))
	}
	if got := strings.Join(cellsOf(t, data, "SampleID"), ","); got != "S1,S2" {
		t.Errorf("group column = %s, want S1,S2", got)
	}
	if got := strings.Join(cellsOf(t, data, "Abs1"), ","); got != "15,40" {
		t.Errorf("Abs1 = %s, want 15,40", got)
	}
	if got := strings.Join(cellsOf(t, data, "Abs2"), ","); got != "3,8" {
		t.Errorf("Abs2 = %s, want 3,8", got)
	}

	// Row names become the group values: the rows they named no longer exist,
	// and the group is what identifies the new row. Unique by construction,
	// which is what row names require (#859).
	if got := strings.Join(data.RowNames, ","); got != "S1,S2" {
		t.Errorf("row names = %s, want S1,S2", got)
	}
	if data.RowNamesHeader != "SampleID" {
		t.Errorf("RowNamesHeader = %q, want SampleID", data.RowNamesHeader)
	}
}

func TestAggregateFunctions(t *testing.T) {
	tests := []struct {
		fn   AggregateFunc
		want string
	}{
		{AggregateMean, "15,40"},
		{AggregateSum, "30,80"},
		{AggregateFirst, "10,30"},
		{AggregateMedian, "15,40"},
	}
	for _, tt := range tests {
		t.Run(string(tt.fn), func(t *testing.T) {
			data := replicateFixture()
			cmd, _ := NewAggregateRowsCommand(NewApp(), data,
				AggregateOptions{GroupBy: "SampleID", Func: tt.fn})
			if err := cmd.Execute(data); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if got := strings.Join(cellsOf(t, data, "Abs1"), ","); got != tt.want {
				t.Errorf("Abs1 = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestAggregateClearsDisagreeingText covers the decision not to pick a winner.
//
// Where a group disagrees, choosing one of the competing values would assert
// something about the aggregated sample that no row said. The cell is cleared
// and the count reported instead.
func TestAggregateClearsDisagreeingText(t *testing.T) {
	data := replicateFixture()
	data.Data[1][1] = "ZZ" // the two S1 replicates now disagree on Operator

	cmd, _ := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got := cellsOf(t, data, "Operator")
	if got[0] != "" {
		t.Errorf("a disagreeing group should clear the cell, got %q", got[0])
	}
	if got[1] != "CD" {
		t.Errorf("an agreeing group should keep its value, got %q", got[1])
	}
	if !strings.Contains(cmd.GetDescription(), "cleared") {
		t.Errorf("the description should report the cleared cells: %q", cmd.GetDescription())
	}
}

// TestAggregateTextAgreementIgnoresBlanks checks a gap does not count as
// disagreement: one replicate recording the operator and the others leaving it
// blank still agrees on the operator.
func TestAggregateTextAgreementIgnoresBlanks(t *testing.T) {
	data := replicateFixture()
	data.Data[1][1] = ""

	cmd, _ := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := cellsOf(t, data, "Operator")[0]; got != "AB" {
		t.Errorf("got %q, want AB — a blank is not a competing value", got)
	}
}

// TestAggregateSkipsMissingNumbers guards against treating a gap as zero.
//
// Averaging a missing value in as zero drags the result towards zero in
// proportion to how much data is absent, which is a silent bias rather than a
// visible gap.
func TestAggregateSkipsMissingNumbers(t *testing.T) {
	data := replicateFixture()
	data.Data[1][2] = "" // one of S1's two Abs1 readings is missing

	cmd, _ := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := cellsOf(t, data, "Abs1")[0]; got != "10" {
		t.Errorf("Abs1 = %q, want 10 — the mean of the one present value, not 5", got)
	}

	// A group with nothing present must stay empty rather than become zero.
	all := replicateFixture()
	all.Data[0][2], all.Data[1][2] = "", ""
	cmd2, _ := NewAggregateRowsCommand(NewApp(), all,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd2.Execute(all); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := cellsOf(t, all, "Abs1")[0]; got != "" {
		t.Errorf("an entirely missing group gave %q, want empty", got)
	}
}

// TestAggregateRefusesBlankGroupValues covers the case that has no good answer.
//
// A blank is not a group. Averaging the unlabelled rows together would invent a
// sample; dropping them silently would lose data. The operation stops and says
// which rows to deal with, and Filter Rows removes them in one step.
func TestAggregateRefusesBlankGroupValues(t *testing.T) {
	data := replicateFixture()
	data.Data[2][0] = ""

	_, err := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err == nil {
		t.Fatal("rows with no group value were accepted")
	}
	for _, want := range []string{"row 3", "Filter Rows"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should mention %q, got %v", want, err)
		}
	}
	if len(data.Data) != 4 {
		t.Errorf("a refused aggregation modified the data")
	}
}

// TestAggregateRebuildsPerRowMaps is the check for the quietest failure.
//
// CategoricalColumns and NumericTargetColumns hold one entry per row. After
// aggregation the rows have been combined, not removed, so the old entries
// describe a table that no longer exists and cannot simply be filtered.
func TestAggregateRebuildsPerRowMaps(t *testing.T) {
	data := replicateFixture()
	data.NumericTargetColumns = map[string][]types.JSONFloat64{}
	data.Headers = append(data.Headers, "Yield#target")
	data.ColumnTypes["Yield#target"] = "target"
	for i, v := range []string{"1", "3", "5", "7"} {
		data.Data[i] = append(data.Data[i], v)
	}
	data.NumericTargetColumns["Yield#target"] = []types.JSONFloat64{1, 3, 5, 7}
	data.Columns = len(data.Headers)

	cmd, _ := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for column, values := range data.CategoricalColumns {
		if len(values) != len(data.Data) {
			t.Errorf("CategoricalColumns[%q] has %d entries for %d rows",
				column, len(values), len(data.Data))
		}
	}
	if got := data.CategoricalColumns["Operator"]; strings.Join(got, ",") != "AB,CD" {
		t.Errorf("CategoricalColumns[Operator] = %v, want [AB CD]", got)
	}
	targets := data.NumericTargetColumns["Yield#target"]
	if len(targets) != 2 {
		t.Fatalf("target column has %d entries for %d rows", len(targets), len(data.Data))
	}
	if targets[0] != 2 || targets[1] != 6 {
		t.Errorf("targets = %v, want the group means [2 6]", targets)
	}
}

func TestAggregateUndo(t *testing.T) {
	data := replicateFixture()
	cmd, _ := NewAggregateRowsCommand(NewApp(), data,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if len(data.Data) != 4 {
		t.Fatalf("undo restored %d rows, want 4", len(data.Data))
	}
	if got := strings.Join(data.RowNames, ","); got != "r1,r2,r3,r4" {
		t.Errorf("row names = %s, want the originals", got)
	}
	if data.RowNamesHeader != "" {
		t.Errorf("RowNamesHeader = %q, want it restored to empty", data.RowNamesHeader)
	}
	if got := strings.Join(data.CategoricalColumns["Operator"], ","); got != "AB,AB,CD,CD" {
		t.Errorf("per-row map not restored: %v", data.CategoricalColumns["Operator"])
	}
}

func TestPreviewAggregate(t *testing.T) {
	app := NewApp()
	data := replicateFixture()

	preview := app.PreviewAggregate(data, AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if preview.Groups != 2 || preview.Rows != 4 {
		t.Errorf("preview = %+v, want 2 groups from 4 rows", preview)
	}
	if preview.LargestSize != 2 || preview.SmallestSize != 2 {
		t.Errorf("group sizes = %d..%d, want 2..2", preview.SmallestSize, preview.LargestSize)
	}

	// Grouping by a column where every value is distinct changes nothing, and
	// that is what a mistyped grouping column looks like.
	nothing := app.PreviewAggregate(data, AggregateOptions{GroupBy: "Abs1", Func: AggregateMean})
	if nothing.Groups != nothing.Rows {
		t.Errorf("grouping by a unique column should give one group per row, got %+v", nothing)
	}

	// Conflicts are counted before anything is applied.
	conflicting := replicateFixture()
	conflicting.Data[1][1] = "ZZ"
	withConflict := app.PreviewAggregate(conflicting,
		AggregateOptions{GroupBy: "SampleID", Func: AggregateMean})
	if withConflict.TextConflicts != 1 {
		t.Errorf("expected 1 text conflict, got %d", withConflict.TextConflicts)
	}

	if len(data.Data) != 4 {
		t.Error("PreviewAggregate modified the data")
	}
}

// TestDeleteRowsKeepsPerRowMapsAligned is the regression test for #871.
//
// DeleteRowsCommand updated Data and RowNames and left CategoricalColumns and
// NumericTargetColumns behind. Those hold one entry per row, so after a
// deletion they were longer than the table and every value from that row
// onwards was attached to the wrong sample.
//
// It went unnoticed because nothing analytical read those per-row values —
// which is a property of the current consumers, not of the code. The assertion
// is on the contents, not the length: a length check passes when the values are
// merely shifted.
func TestDeleteRowsKeepsPerRowMapsAligned(t *testing.T) {
	data := replicateFixture()
	data.NumericTargetColumns = map[string][]types.JSONFloat64{
		"Yield#target": {1, 2, 3, 4},
	}

	cmd := NewDeleteRowsCommand(data, []int{1}) // drop the second row
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(data.Data) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(data.Data))
	}
	if got := strings.Join(data.CategoricalColumns["Operator"], ","); got != "AB,CD,CD" {
		t.Errorf("CategoricalColumns[Operator] = %q, want \"AB,CD,CD\" — the entry for "+
			"the deleted row must go with it", got)
	}
	if got := strings.Join(data.CategoricalColumns["SampleID"], ","); got != "S1,S2,S2" {
		t.Errorf("CategoricalColumns[SampleID] = %q, want \"S1,S2,S2\"", got)
	}
	targets := data.NumericTargetColumns["Yield#target"]
	if len(targets) != 3 || targets[0] != 1 || targets[1] != 3 || targets[2] != 4 {
		t.Errorf("targets = %v, want [1 3 4] — value 2 belonged to the deleted row", targets)
	}
	if got := strings.Join(data.RowNames, ","); got != "r1,r3,r4" {
		t.Errorf("row names = %s, want r1,r3,r4", got)
	}
}
