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

func reorderFixture() *FileData {
	return &FileData{
		Headers:  []string{"A", "B", "C"},
		RowNames: []string{"r1", "r2"},
		Data:     [][]string{{"a1", "b1", "c1"}, {"a2", "b2", "c2"}},
		Rows:     2,
		Columns:  3,
		ColumnTypes: map[string]string{
			"A": "numeric", "B": "categorical", "C": "numeric",
		},
		CategoricalColumns:   map[string][]string{"B": {"b1", "b2"}},
		NumericTargetColumns: map[string][]types.JSONFloat64{},
	}
}

// TestReorderMovesHeadersAndCells is the assertion that matters: a reorder that
// moves the headers but not the cells silently relabels every column.
func TestReorderMovesHeadersAndCells(t *testing.T) {
	data := reorderFixture()

	// Move C to the front: [C A B]
	cmd, err := NewReorderColumnsCommand(NewApp(), data, []int{2, 0, 1})
	if err != nil {
		t.Fatalf("NewReorderColumnsCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := strings.Join(data.Headers, ","); got != "C,A,B" {
		t.Errorf("headers = %s, want C,A,B", got)
	}
	if got := strings.Join(data.Data[0], ","); got != "c1,a1,b1" {
		t.Errorf("row 0 = %s, want c1,a1,b1 — the cells must travel with their header", got)
	}
	if got := strings.Join(data.Data[1], ","); got != "c2,a2,b2" {
		t.Errorf("row 1 = %s, want c2,a2,b2", got)
	}
	if data.Columns != 3 {
		t.Errorf("Columns = %d, want 3", data.Columns)
	}

	// The per-row maps are keyed by name, so they neither move nor break.
	if got := strings.Join(data.CategoricalColumns["B"], ","); got != "b1,b2" {
		t.Errorf("CategoricalColumns[B] = %q, should be untouched by a reorder", got)
	}
	// Row names describe rows, not columns.
	if got := strings.Join(data.RowNames, ","); got != "r1,r2" {
		t.Errorf("row names changed during a column reorder: %v", data.RowNames)
	}
}

// TestReorderRejectsBadPermutations covers the failures that would lose data.
//
// A repeat duplicates a column and drops another; a short list drops the rest.
// A reorder that loses a variable is far worse than one that refuses.
func TestReorderRejectsBadPermutations(t *testing.T) {
	tests := []struct {
		name    string
		order   []int
		wantErr string
	}{
		{name: "a column listed twice", order: []int{0, 0, 1}, wantErr: "twice"},
		{name: "too few columns", order: []int{0, 1}, wantErr: "lists 2 columns"},
		{name: "too many columns", order: []int{0, 1, 2, 2}, wantErr: "lists 4 columns"},
		{name: "index outside the table", order: []int{0, 1, 9}, wantErr: "outside the table"},
		{name: "negative index", order: []int{0, 1, -1}, wantErr: "outside the table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := reorderFixture()
			_, err := NewReorderColumnsCommand(NewApp(), data, tt.order)
			if err == nil {
				t.Fatalf("expected a refusal mentioning %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q should mention %q", err, tt.wantErr)
			}
			if strings.Join(data.Headers, ",") != "A,B,C" {
				t.Errorf("a refused reorder modified the data: %v", data.Headers)
			}
		})
	}
}

func TestReorderUndo(t *testing.T) {
	data := reorderFixture()
	cmd, _ := NewReorderColumnsCommand(NewApp(), data, []int{2, 0, 1})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if got := strings.Join(data.Headers, ","); got != "A,B,C" {
		t.Errorf("headers = %s after undo, want A,B,C", got)
	}
	if got := strings.Join(data.Data[0], ","); got != "a1,b1,c1" {
		t.Errorf("row 0 = %s after undo", got)
	}
}

// TestReorderDescriptionNamesTheMovedColumn keeps the undo history readable.
func TestReorderDescriptionNamesTheMovedColumn(t *testing.T) {
	data := reorderFixture()
	cmd, _ := NewReorderColumnsCommand(NewApp(), data, []int{2, 0, 1})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := cmd.GetDescription(); !strings.Contains(got, "'C'") {
		t.Errorf("description = %q, should name the column that moved", got)
	}

	// A no-op reorder should not claim a column moved.
	same := reorderFixture()
	noop, _ := NewReorderColumnsCommand(NewApp(), same, []int{0, 1, 2})
	if err := noop.Execute(same); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := noop.GetDescription(); strings.Contains(got, "'") {
		t.Errorf("an identity reorder should not name a column: %q", got)
	}
}

// TestReorderRaggedRows checks a short row does not gain or lose cells.
func TestReorderRaggedRows(t *testing.T) {
	data := reorderFixture()
	data.Data[1] = []string{"a2"} // a short row

	cmd, _ := NewReorderColumnsCommand(NewApp(), data, []int{2, 0, 1})
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := strings.Join(data.Data[1], ","); got != ",a2," {
		t.Errorf("short row = %q, want \",a2,\" — a2 moves to where A now is", got)
	}
	for i, row := range data.Data {
		if len(row) != len(data.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers", i, len(row), len(data.Headers))
		}
	}
}
