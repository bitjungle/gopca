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

import "testing"

// The #category marker exists for the one thing a column's type cannot tell us:
// that its numbers are labels (#914).

func TestNumericCategoryColumnIsHeldOutOfTheAnalysis(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent(
		"ID,x1,x2,proc_num#category\nS1,1,2,10\nS2,3,4,11\nS3,5,6,10\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if got := data.ColumnTypes["proc_num#category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical", got)
	}
	if _, ok := data.CategoricalColumns["proc_num#category"]; !ok {
		t.Error("not offered as a categorical column, so it cannot colour a plot")
	}
}

func TestWithoutTheMarkerTheSameColumnIsAVariable(t *testing.T) {
	// The control. Without this, a parser that made everything categorical
	// would satisfy the test above.
	app := &App{}
	data, err := app.parseCSVContent(
		"ID,x1,x2,proc_num\nS1,1,2,10\nS2,3,4,11\nS3,5,6,10\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if got := data.ColumnTypes["proc_num"]; got != "numeric" {
		t.Errorf("type = %q, want numeric: an unmarked number is a measurement", got)
	}
}

func TestCategoryMarkerAcceptsTheSpacedForm(t *testing.T) {
	// swiss_roll.csv ships "color #target", so the form does occur in real
	// files. It works because the suffix test does not care what precedes the
	// marker -- this pins the accepted form rather than a separate code path.
	app := &App{}
	data, err := app.parseCSVContent(
		"ID,x1,site #category\nS1,1,3\nS2,2,4\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if got := data.ColumnTypes["site #category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical", got)
	}
}

func TestCategoryMarkerOnTextIsRedundantNotWrong(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent(
		"ID,x1,species#category\nS1,1,setosa\nS2,2,virginica\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if got := data.ColumnTypes["species#category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical", got)
	}
}

func TestToggleCategoryMarksAndUnmarks(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent("ID,x1,proc_num\nS1,1,10\nS2,2,11\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	idx := -1
	for i, h := range data.Headers {
		if h == "proc_num" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("proc_num not found in %v", data.Headers)
	}

	on := NewToggleCategoryColumnCommand(app, data, idx)
	if err := on.Execute(data); err != nil {
		t.Fatalf("toggle on: %v", err)
	}
	if data.Headers[idx] != "proc_num#category" {
		t.Fatalf("header = %q, want proc_num#category", data.Headers[idx])
	}
	if got := data.ColumnTypes["proc_num#category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical", got)
	}

	// Removing it must hand the column back to its values, not leave it
	// categorical because it once carried the marker.
	off := NewToggleCategoryColumnCommand(app, data, idx)
	if err := off.Execute(data); err != nil {
		t.Fatalf("toggle off: %v", err)
	}
	if data.Headers[idx] != "proc_num" {
		t.Fatalf("header = %q, want proc_num", data.Headers[idx])
	}
	if got := data.ColumnTypes["proc_num"]; got != "numeric" {
		t.Errorf("type = %q, want numeric: the column holds numbers", got)
	}

	// And undo restores the marked state it was in.
	if err := off.Undo(data); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if data.Headers[idx] != "proc_num#category" {
		t.Errorf("after undo header = %q, want proc_num#category", data.Headers[idx])
	}
	if got := data.ColumnTypes["proc_num#category"]; got != "categorical" {
		t.Errorf("after undo type = %q, want categorical", got)
	}
}

func TestUnmarkingATextColumnLeavesItCategorical(t *testing.T) {
	// The asymmetry worth pinning: a text column has nothing to revert to.
	app := &App{}
	data, err := app.parseCSVContent("ID,x1,species\nS1,1,setosa\nS2,2,virginica\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	idx := -1
	for i, h := range data.Headers {
		if h == "species" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("species not found in %v", data.Headers)
	}

	on := NewToggleCategoryColumnCommand(app, data, idx)
	if err := on.Execute(data); err != nil {
		t.Fatalf("toggle on: %v", err)
	}
	off := NewToggleCategoryColumnCommand(app, data, idx)
	if err := off.Execute(data); err != nil {
		t.Fatalf("toggle off: %v", err)
	}
	if got := data.ColumnTypes["species"]; got != "categorical" {
		t.Errorf("type = %q, want categorical: text is categorical whatever the name says", got)
	}
}
