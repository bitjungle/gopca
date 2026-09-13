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
)

// #target and #category describe mutually exclusive roles: a column is an
// outcome you might predict, or a label you group by, never both (#922).

func markerFixture(t *testing.T, header, value string) (*App, *FileData, int) {
	t.Helper()
	app := &App{}
	data, err := app.parseCSVContent("ID,x1,"+header+"\nS1,1,"+value+"\nS2,2,"+value+"\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	for i, h := range data.Headers {
		if strings.HasPrefix(h, strings.Split(header, "#")[0]) {
			return app, data, i
		}
	}
	t.Fatalf("column %q not found in %v", header, data.Headers)
	return nil, nil, 0
}

func TestMarkingACategoryColumnAsTargetReplacesTheMarker(t *testing.T) {
	app, data, idx := markerFixture(t, "code#category", "10")
	NewToggleTargetColumnCommand(app, data, idx).Execute(data)

	if got := data.Headers[idx]; got != "code#target" {
		t.Fatalf("header = %q, want code#target -- the markers must not accumulate", got)
	}
	if got := data.ColumnTypes["code#target"]; got != "target" {
		t.Errorf("type = %q, want target", got)
	}
	if _, ok := data.CategoricalColumns["code#target"]; ok {
		t.Error("still registered as categorical; the frontend and pkg/transform would disagree with ColumnTypes")
	}
}

func TestMarkingATargetColumnAsCategoryReplacesTheMarker(t *testing.T) {
	app, data, idx := markerFixture(t, "code#target", "10")
	NewToggleCategoryColumnCommand(app, data, idx).Execute(data)

	if got := data.Headers[idx]; got != "code#category" {
		t.Fatalf("header = %q, want code#category", got)
	}
	if got := data.ColumnTypes["code#category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical", got)
	}
	if _, ok := data.CategoricalColumns["code#category"]; !ok {
		t.Error("not registered as categorical, so nothing downstream can colour by it")
	}
}

func TestTogglingCleansUpAColumnThatCarriesBothMarkers(t *testing.T) {
	// Files written before #922 can hold these, and they still open.
	app, data, idx := markerFixture(t, "code#target#category", "10")
	NewToggleTargetColumnCommand(app, data, idx).Execute(data)

	if got := data.Headers[idx]; got != "code#target" {
		t.Errorf("header = %q, want code#target -- toggling should reduce to one marker, not add a third", got)
	}
}

func TestRemovingTheTargetFlagRestoresWhatTheColumnHolds(t *testing.T) {
	// The pre-existing bug this change had to fix: removing the flag set the
	// type to "numeric" unconditionally, which is wrong for a column of text.
	app, data, idx := markerFixture(t, "species", "setosa")
	cmd := NewToggleTargetColumnCommand(app, data, idx)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	off := NewToggleTargetColumnCommand(app, data, idx)
	if err := off.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := data.ColumnTypes["species"]; got != "categorical" {
		t.Errorf("type = %q, want categorical: the column holds text", got)
	}
	if _, ok := data.CategoricalColumns["species"]; !ok {
		t.Error("not returned to CategoricalColumns")
	}
}

func TestRemovingTheTargetFlagFromANumericColumnGivesNumeric(t *testing.T) {
	// The companion control. Without it, a rule that always said "categorical"
	// would satisfy the test above.
	app, data, idx := markerFixture(t, "code", "10")
	NewToggleTargetColumnCommand(app, data, idx).Execute(data)
	NewToggleTargetColumnCommand(app, data, idx).Execute(data)

	if got := data.ColumnTypes["code"]; got != "numeric" {
		t.Errorf("type = %q, want numeric", got)
	}
	if _, ok := data.CategoricalColumns["code"]; ok {
		t.Error("a numeric column must not be left in CategoricalColumns")
	}
}

func TestUndoRestoresTheMarkerThatWasReplaced(t *testing.T) {
	app, data, idx := markerFixture(t, "code#category", "10")
	cmd := NewToggleTargetColumnCommand(app, data, idx)
	if err := cmd.Execute(data); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := cmd.Undo(data); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if got := data.Headers[idx]; got != "code#category" {
		t.Fatalf("header = %q, want code#category back", got)
	}
	if got := data.ColumnTypes["code#category"]; got != "categorical" {
		t.Errorf("type = %q, want categorical restored", got)
	}
	if _, ok := data.CategoricalColumns["code#category"]; !ok {
		t.Error("CategoricalColumns not restored by undo")
	}
}

func TestMarkersSurviveACsvRoundTrip(t *testing.T) {
	// The name is what persists, so the classification after reload has to agree
	// with what the toggle produced in the grid.
	app, data, idx := markerFixture(t, "code#category", "10")
	NewToggleTargetColumnCommand(app, data, idx).Execute(data)
	header := data.Headers[idx]

	reloaded, err := app.parseCSVContent("ID,x1,"+header+"\nS1,1,10\nS2,2,11\n", ".csv")
	if err != nil {
		t.Fatalf("parseCSVContent: %v", err)
	}
	if got := reloaded.ColumnTypes[header]; got != "target" {
		t.Errorf("after reload type = %q, want target", got)
	}
}

func TestStripMarkersHandlesSpacingAndCase(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"code", "code"},
		{"code#target", "code"},
		{"code #category", "code"},
		{"code# target", "code"},
		{"code#TARGET", "code"},
		{"code#target#category", "code"},
		{"code#category#target#category", "code"},
	} {
		if got := stripMarkers(tc.in); got != tc.want {
			t.Errorf("stripMarkers(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
