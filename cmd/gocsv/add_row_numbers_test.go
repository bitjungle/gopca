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
	"strconv"
	"strings"
	"testing"
)

// Some datasets carry nothing that can identify a row. #923.

func numberlessFixture(rows int) *FileData {
	d := &FileData{
		Headers:     []string{"A", "B"},
		ColumnTypes: map[string]string{"A": "numeric", "B": "numeric"},
		Columns:     2,
		Rows:        rows,
	}
	for i := 0; i < rows; i++ {
		d.Data = append(d.Data, []string{strconv.Itoa(i), strconv.Itoa(i * 2)})
	}
	return d
}

func TestAddRowNumbersGivesEveryRowAUniqueName(t *testing.T) {
	data := numberlessFixture(5)
	cmd, err := NewAddRowNumbersCommand(&App{}, data)
	if err != nil {
		t.Fatalf("NewAddRowNumbersCommand: %v", err)
	}
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(data.RowNames) != 5 {
		t.Fatalf("got %d row names, want 5", len(data.RowNames))
	}
	if data.RowNames[0] != "1" || data.RowNames[4] != "5" {
		t.Errorf("names = %v, want 1..5", data.RowNames)
	}
	if data.RowNamesHeader != "Sample_ID" {
		t.Errorf("header = %q, want Sample_ID", data.RowNamesHeader)
	}

	// The point of the exercise: they must satisfy the rule that rejected every
	// column in the file.
	if check := checkRowNameCandidate(data.RowNames); !check.OK {
		t.Errorf("generated names do not qualify as row names: %s", check.Reason)
	}
}

func TestAddRowNumbersAddsNoColumn(t *testing.T) {
	// Row names are not a column. Inserting one would put a numeric sequence in
	// the table, where it would enter the PCA -- and dominate it.
	data := numberlessFixture(4)
	cmd, _ := NewAddRowNumbersCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(data.Headers) != 2 {
		t.Errorf("headers = %v, want the original two", data.Headers)
	}
	for _, row := range data.Data {
		if len(row) != 2 {
			t.Fatalf("row widened to %d cells", len(row))
		}
	}
}

func TestAddRowNumbersRefusesWhenRowNamesExist(t *testing.T) {
	// The control. Overwriting identifiers that mean something, in exchange for
	// ordinals that do not, is not something to do quietly.
	data := numberlessFixture(3)
	data.RowNames = []string{"alpha", "beta", "gamma"}
	data.RowNamesHeader = "SampleName"

	_, err := NewAddRowNumbersCommand(&App{}, data)
	if err == nil {
		t.Fatal("accepted a file that already has row names")
	}
	if !strings.Contains(err.Error(), "SampleName") {
		t.Errorf("error should name the existing row-name column, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Move Row Names into Table") {
		t.Errorf("error should say how to proceed, got: %v", err)
	}
}

func TestAddRowNumbersAvoidsAHeaderCollision(t *testing.T) {
	data := numberlessFixture(3)
	data.Headers = []string{"Sample_ID", "B"}
	cmd, err := NewAddRowNumbersCommand(&App{}, data)
	if err != nil {
		t.Fatalf("NewAddRowNumbersCommand: %v", err)
	}
	if cmd.header == "Sample_ID" {
		t.Error("reused a header the table already has; export would produce two Sample_ID columns")
	}
}

func TestAddRowNumbersUndo(t *testing.T) {
	data := numberlessFixture(3)
	cmd, _ := NewAddRowNumbersCommand(&App{}, data)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if len(data.RowNames) != 0 || data.RowNamesHeader != "" {
		t.Errorf("undo left names=%v header=%q", data.RowNames, data.RowNamesHeader)
	}
}

func TestAddRowNumbersRefusesAnEmptyFile(t *testing.T) {
	if _, err := NewAddRowNumbersCommand(&App{}, &FileData{}); err == nil {
		t.Error("accepted a file with no rows")
	}
}
