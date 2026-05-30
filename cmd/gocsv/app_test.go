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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppMultiStepUndoRedo(t *testing.T) {
	// Create app
	app := NewApp()

	// Create test data
	data := &FileData{
		Headers: []string{"A", "B", "C"},
		Data: [][]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
			{"7", "8", "9"},
		},
		Rows:    3,
		Columns: 3,
	}

	// Set initial current data (simulating LoadCSV)
	app.currentData = data

	// Execute multiple cell edits
	// Edit 1: Change cell [0,0] from "1" to "10"
	data1, err := app.ExecuteCellEdit(data, 0, 0, "1", "10")
	if err != nil {
		t.Fatalf("Edit 1 failed: %v", err)
	}
	if data1.Data[0][0] != "10" {
		t.Errorf("Edit 1: expected '10', got '%s'", data1.Data[0][0])
	}

	// Edit 2: Change cell [1,1] from "5" to "50"
	data2, err := app.ExecuteCellEdit(data1, 1, 1, "5", "50")
	if err != nil {
		t.Fatalf("Edit 2 failed: %v", err)
	}
	if data2.Data[1][1] != "50" {
		t.Errorf("Edit 2: expected '50', got '%s'", data2.Data[1][1])
	}

	// Edit 3: Change cell [2,2] from "9" to "90"
	data3, err := app.ExecuteCellEdit(data2, 2, 2, "9", "90")
	if err != nil {
		t.Fatalf("Edit 3 failed: %v", err)
	}
	if data3.Data[2][2] != "90" {
		t.Errorf("Edit 3: expected '90', got '%s'", data3.Data[2][2])
	}

	// Test undo operations
	// Undo 1: Should revert cell [2,2] from "90" to "9"
	dataUndo1, err := app.Undo(data3)
	if err != nil {
		t.Fatalf("Undo 1 failed: %v", err)
	}
	if dataUndo1.Data[2][2] != "9" {
		t.Errorf("Undo 1: expected '9', got '%s'", dataUndo1.Data[2][2])
	}
	if dataUndo1.Data[1][1] != "50" {
		t.Errorf("Undo 1: cell [1,1] should still be '50', got '%s'", dataUndo1.Data[1][1])
	}

	// Undo 2: Should revert cell [1,1] from "50" to "5"
	dataUndo2, err := app.Undo(dataUndo1)
	if err != nil {
		t.Fatalf("Undo 2 failed: %v", err)
	}
	if dataUndo2.Data[1][1] != "5" {
		t.Errorf("Undo 2: expected '5', got '%s'", dataUndo2.Data[1][1])
	}
	if dataUndo2.Data[0][0] != "10" {
		t.Errorf("Undo 2: cell [0,0] should still be '10', got '%s'", dataUndo2.Data[0][0])
	}

	// Undo 3: Should revert cell [0,0] from "10" to "1"
	dataUndo3, err := app.Undo(dataUndo2)
	if err != nil {
		t.Fatalf("Undo 3 failed: %v", err)
	}
	if dataUndo3.Data[0][0] != "1" {
		t.Errorf("Undo 3: expected '1', got '%s'", dataUndo3.Data[0][0])
	}

	// Test redo operations
	// Redo 1: Should restore cell [0,0] from "1" to "10"
	dataRedo1, err := app.Redo(dataUndo3)
	if err != nil {
		t.Fatalf("Redo 1 failed: %v", err)
	}
	if dataRedo1.Data[0][0] != "10" {
		t.Errorf("Redo 1: expected '10', got '%s'", dataRedo1.Data[0][0])
	}

	// Redo 2: Should restore cell [1,1] from "5" to "50"
	dataRedo2, err := app.Redo(dataRedo1)
	if err != nil {
		t.Fatalf("Redo 2 failed: %v", err)
	}
	if dataRedo2.Data[1][1] != "50" {
		t.Errorf("Redo 2: expected '50', got '%s'", dataRedo2.Data[1][1])
	}

	// Redo 3: Should restore cell [2,2] from "9" to "90"
	dataRedo3, err := app.Redo(dataRedo2)
	if err != nil {
		t.Fatalf("Redo 3 failed: %v", err)
	}
	if dataRedo3.Data[2][2] != "90" {
		t.Errorf("Redo 3: expected '90', got '%s'", dataRedo3.Data[2][2])
	}

	// Verify command history state
	state := app.GetUndoRedoState()
	if state.CanUndo != true {
		t.Error("Should be able to undo after 3 edits")
	}
	if state.CanRedo != false {
		t.Error("Should not be able to redo when at the end of history")
	}
	if len(state.History) != 3 {
		t.Errorf("Expected 3 commands in history, got %d", len(state.History))
	}
	if state.CurrentPos != 2 {
		t.Errorf("Expected current position to be 2, got %d", state.CurrentPos)
	}
}

// TestUnsavedChanges verifies that HasUnsavedChanges, markDirty, and markClean
// track state transitions correctly without requiring a Wails context.
// Note: ctx is nil here so EventsEmit is not called; this is intentional.
func TestUnsavedChanges(t *testing.T) {
	t.Run("InitialStateIsClean", func(t *testing.T) {
		app := NewApp()
		if app.HasUnsavedChanges() {
			t.Error("expected clean on fresh app")
		}
	})

	t.Run("MarkDirtySetsDirty", func(t *testing.T) {
		app := NewApp()
		app.markDirty()
		if !app.HasUnsavedChanges() {
			t.Error("expected dirty after markDirty")
		}
	})

	t.Run("MarkCleanClearsDirty", func(t *testing.T) {
		app := NewApp()
		app.markDirty()
		app.markClean()
		if app.HasUnsavedChanges() {
			t.Error("expected clean after markClean")
		}
	})

	t.Run("ExecuteCellEditMarksDirty", func(t *testing.T) {
		app := NewApp()
		data := &FileData{
			Headers: []string{"A"},
			Data:    [][]string{{"1"}},
			Rows:    1, Columns: 1,
		}
		app.currentData = data
		_, err := app.executeCommand(NewCellEditCommand(0, 0, "1", "2"), data, "edit")
		if err != nil {
			t.Fatalf("executeCommand failed: %v", err)
		}
		if !app.HasUnsavedChanges() {
			t.Error("expected dirty after executeCommand")
		}
	})

	t.Run("MarkCleanAfterEditResetsFlag", func(t *testing.T) {
		app := NewApp()
		data := &FileData{
			Headers: []string{"A"},
			Data:    [][]string{{"1"}},
			Rows:    1, Columns: 1,
		}
		app.currentData = data
		_, _ = app.executeCommand(NewCellEditCommand(0, 0, "1", "2"), data, "edit")
		app.markClean()
		if app.HasUnsavedChanges() {
			t.Error("expected clean after markClean following edit")
		}
	})

	t.Run("SimulatedLoadResetsFlag", func(t *testing.T) {
		app := NewApp()
		app.markDirty()

		// Simulate what LoadCSV does internally
		data := &FileData{Headers: []string{"X"}, Data: [][]string{{"1"}}, Rows: 1, Columns: 1}
		app.currentData = data
		app.history.Clear()
		app.markClean()

		if app.HasUnsavedChanges() {
			t.Error("expected clean after simulated LoadCSV")
		}
	})
}

func TestCommandHistoryWithMultipleOperations(t *testing.T) {
	// Create app
	app := NewApp()

	// Create test data
	data := &FileData{
		Headers: []string{"Name", "Age", "City"},
		Data: [][]string{
			{"Alice", "25", "NYC"},
			{"Bob", "30", "LA"},
			{"Charlie", "35", "Chicago"},
		},
		Rows:    3,
		Columns: 3,
		ColumnTypes: map[string]string{
			"Name": "text",
			"Age":  "numeric",
			"City": "text",
		},
	}

	// Set initial current data
	app.currentData = data

	// Mix of different operations
	// 1. Edit header
	data1, err := app.ExecuteHeaderEdit(data, 1, "Age", "Years")
	if err != nil {
		t.Fatalf("Header edit failed: %v", err)
	}

	// 2. Edit cell
	data2, err := app.ExecuteCellEdit(data1, 0, 1, "25", "26")
	if err != nil {
		t.Fatalf("Cell edit failed: %v", err)
	}

	// 3. Delete rows
	_, err = app.ExecuteDeleteRows(data2, []int{2})
	if err != nil {
		t.Fatalf("Delete rows failed: %v", err)
	}

	// Save the current data state for undo
	currentData := data

	// Test multi-step undo
	// Undo delete rows
	dataUndo1, err := app.Undo(currentData)
	if err != nil {
		t.Fatalf("Undo delete rows failed: %v", err)
	}
	if len(dataUndo1.Data) != 3 {
		t.Errorf("Expected 3 rows after undo delete, got %d", len(dataUndo1.Data))
	}

	// Undo cell edit
	dataUndo2, err := app.Undo(dataUndo1)
	if err != nil {
		t.Fatalf("Undo cell edit failed: %v", err)
	}
	if dataUndo2.Data[0][1] != "25" {
		t.Errorf("Expected '25' after undo cell edit, got '%s'", dataUndo2.Data[0][1])
	}

	// Undo header edit
	dataUndo3, err := app.Undo(dataUndo2)
	if err != nil {
		t.Fatalf("Undo header edit failed: %v", err)
	}
	if dataUndo3.Headers[1] != "Age" {
		t.Errorf("Expected 'Age' after undo header edit, got '%s'", dataUndo3.Headers[1])
	}

	// Redo all operations
	dataRedo1, _ := app.Redo(dataUndo3)
	dataRedo2, _ := app.Redo(dataRedo1)
	dataRedo3, err := app.Redo(dataRedo2)
	if err != nil {
		t.Fatalf("Redo failed: %v", err)
	}

	// Verify final state
	if dataRedo3.Headers[1] != "Years" {
		t.Errorf("Expected 'Years' header, got '%s'", dataRedo3.Headers[1])
	}
	if dataRedo3.Data[0][1] != "26" {
		t.Errorf("Expected '26' in cell, got '%s'", dataRedo3.Data[0][1])
	}
	if len(dataRedo3.Data) != 2 {
		t.Errorf("Expected 2 rows after redo, got %d", len(dataRedo3.Data))
	}
}

func TestLoadParquet(t *testing.T) {
	app := &App{}
	fileData, err := app.loadParquet("../../testdata/energy_mix/energy_mix.parquet")
	require.NoError(t, err)
	require.NotNil(t, fileData)

	// 7314 rows (138 countries × 59 years)
	// Sample_ID (prepended integers "1"…"7314") becomes the unique row identifier
	// country#target is a categorical column; year + 104 energy columns are numeric
	assert.Equal(t, 7314, fileData.Rows)
	assert.Equal(t, 106, fileData.Columns) // country#target + year + 104 energy cols
	assert.Equal(t, "country#target", fileData.Headers[0])
	assert.Equal(t, "year", fileData.Headers[1])
	assert.Equal(t, 7314, len(fileData.RowNames))
	assert.Equal(t, "1", fileData.RowNames[0])
	assert.Equal(t, "categorical", fileData.ColumnTypes["country#target"])
	_, hasCountry := fileData.CategoricalColumns["country#target"]
	assert.True(t, hasCountry, "country#target should be in CategoricalColumns")
	assert.NotEmpty(t, fileData.Data)
}
