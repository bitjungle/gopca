package main

import (
	"testing"
)

// TestMultiStepUndoRedo tests that multiple commands can be undone and redone
func TestMultiStepUndoRedo(t *testing.T) {
	// Create test data
	data := &FileData{
		Headers:  []string{"Col1", "Col2", "Col3"},
		RowNames: []string{"Row1", "Row2"},
		Data: [][]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
		},
		Rows:         2,
		Columns:      3,
		ColumnTypes:  map[string]string{"Col1": "numeric", "Col2": "numeric", "Col3": "numeric"},
	}

	// Create command history
	history := NewCommandHistory(10)

	// Test Case 1: Multiple cell edits
	t.Run("MultipleCellEdits", func(t *testing.T) {
		// Make a copy of the data
		testData := deepCopyFileData(data)
		
		// Execute multiple cell edits
		cmd1 := NewCellEditCommand(0, 0, "1", "10")
		if err := history.Execute(cmd1, testData); err != nil {
			t.Fatalf("Failed to execute first command: %v", err)
		}
		if testData.Data[0][0] != "10" {
			t.Errorf("First edit failed: expected '10', got '%s'", testData.Data[0][0])
		}

		cmd2 := NewCellEditCommand(1, 1, "5", "50")
		if err := history.Execute(cmd2, testData); err != nil {
			t.Fatalf("Failed to execute second command: %v", err)
		}
		if testData.Data[1][1] != "50" {
			t.Errorf("Second edit failed: expected '50', got '%s'", testData.Data[1][1])
		}

		cmd3 := NewCellEditCommand(0, 2, "3", "30")
		if err := history.Execute(cmd3, testData); err != nil {
			t.Fatalf("Failed to execute third command: %v", err)
		}
		if testData.Data[0][2] != "30" {
			t.Errorf("Third edit failed: expected '30', got '%s'", testData.Data[0][2])
		}

		// Test undo all three operations
		if err := history.Undo(testData); err != nil {
			t.Fatalf("Failed to undo third command: %v", err)
		}
		if testData.Data[0][2] != "3" {
			t.Errorf("Third undo failed: expected '3', got '%s'", testData.Data[0][2])
		}

		if err := history.Undo(testData); err != nil {
			t.Fatalf("Failed to undo second command: %v", err)
		}
		if testData.Data[1][1] != "5" {
			t.Errorf("Second undo failed: expected '5', got '%s'", testData.Data[1][1])
		}

		if err := history.Undo(testData); err != nil {
			t.Fatalf("Failed to undo first command: %v", err)
		}
		if testData.Data[0][0] != "1" {
			t.Errorf("First undo failed: expected '1', got '%s'", testData.Data[0][0])
		}

		// Test redo all three operations
		if err := history.Redo(testData); err != nil {
			t.Fatalf("Failed to redo first command: %v", err)
		}
		if testData.Data[0][0] != "10" {
			t.Errorf("First redo failed: expected '10', got '%s'", testData.Data[0][0])
		}

		if err := history.Redo(testData); err != nil {
			t.Fatalf("Failed to redo second command: %v", err)
		}
		if testData.Data[1][1] != "50" {
			t.Errorf("Second redo failed: expected '50', got '%s'", testData.Data[1][1])
		}

		if err := history.Redo(testData); err != nil {
			t.Fatalf("Failed to redo third command: %v", err)
		}
		if testData.Data[0][2] != "30" {
			t.Errorf("Third redo failed: expected '30', got '%s'", testData.Data[0][2])
		}
	})

	// Test Case 2: Mixed operations
	t.Run("MixedOperations", func(t *testing.T) {
		// Reset history
		history = NewCommandHistory(10)
		testData := deepCopyFileData(data)

		// Edit a cell
		cmd1 := NewCellEditCommand(0, 0, "1", "100")
		if err := history.Execute(cmd1, testData); err != nil {
			t.Fatalf("Failed to execute cell edit: %v", err)
		}

		// Edit a header
		cmd2 := NewHeaderEditCommand(1, "Col2", "NewCol2")
		if err := history.Execute(cmd2, testData); err != nil {
			t.Fatalf("Failed to execute header edit: %v", err)
		}

		// Insert a row
		app := &App{} // Mock app for commands that need it
		cmd3 := NewInsertRowCommand(app, testData, 1)
		if err := history.Execute(cmd3, testData); err != nil {
			t.Fatalf("Failed to execute insert row: %v", err)
		}

		// Verify all changes applied
		if testData.Data[0][0] != "100" {
			t.Errorf("Cell edit not applied: expected '100', got '%s'", testData.Data[0][0])
		}
		if testData.Headers[1] != "NewCol2" {
			t.Errorf("Header edit not applied: expected 'NewCol2', got '%s'", testData.Headers[1])
		}
		if testData.Rows != 3 {
			t.Errorf("Row insert not applied: expected 3 rows, got %d", testData.Rows)
		}

		// Undo all operations
		for i := 0; i < 3; i++ {
			if err := history.Undo(testData); err != nil {
				t.Fatalf("Failed to undo operation %d: %v", i+1, err)
			}
		}

		// Verify all changes reverted
		if testData.Data[0][0] != "1" {
			t.Errorf("Cell edit not reverted: expected '1', got '%s'", testData.Data[0][0])
		}
		if testData.Headers[1] != "Col2" {
			t.Errorf("Header edit not reverted: expected 'Col2', got '%s'", testData.Headers[1])
		}
		if testData.Rows != 2 {
			t.Errorf("Row insert not reverted: expected 2 rows, got %d", testData.Rows)
		}
	})

	// Test Case 3: History limit
	t.Run("HistoryLimit", func(t *testing.T) {
		// Create history with limit of 3
		limitedHistory := NewCommandHistory(3)
		testData := deepCopyFileData(data)

		// Execute 5 commands (should only keep last 3)
		for i := 0; i < 5; i++ {
			cmd := NewCellEditCommand(0, 0, testData.Data[0][0], string(rune('A'+i)))
			if err := limitedHistory.Execute(cmd, testData); err != nil {
				t.Fatalf("Failed to execute command %d: %v", i+1, err)
			}
		}

		// Should be able to undo 3 times
		for i := 0; i < 3; i++ {
			if !limitedHistory.CanUndo() {
				t.Errorf("Should be able to undo after %d undos", i)
			}
			if err := limitedHistory.Undo(testData); err != nil {
				t.Fatalf("Failed to undo %d: %v", i+1, err)
			}
		}

		// Should not be able to undo more
		if limitedHistory.CanUndo() {
			t.Error("Should not be able to undo after 3 undos with history limit of 3")
		}

		// The value should be 'B' (the second command, as first two were dropped)
		if testData.Data[0][0] != "B" {
			t.Errorf("After max undos, expected 'B', got '%s'", testData.Data[0][0])
		}
	})
}

// TestUndoRedoState tests the undo/redo state reporting
func TestUndoRedoState(t *testing.T) {
	history := NewCommandHistory(10)
	data := &FileData{
		Headers: []string{"Col1"},
		Data:    [][]string{{"1"}},
		Rows:    1,
		Columns: 1,
	}

	// Initially, can't undo or redo
	if history.CanUndo() {
		t.Error("Should not be able to undo initially")
	}
	if history.CanRedo() {
		t.Error("Should not be able to redo initially")
	}

	// Execute a command
	cmd := NewCellEditCommand(0, 0, "1", "2")
	history.Execute(cmd, data)

	// Now can undo but not redo
	if !history.CanUndo() {
		t.Error("Should be able to undo after executing command")
	}
	if history.CanRedo() {
		t.Error("Should not be able to redo before undoing")
	}

	// Undo
	history.Undo(data)

	// Now can't undo but can redo
	if history.CanUndo() {
		t.Error("Should not be able to undo after undoing all commands")
	}
	if !history.CanRedo() {
		t.Error("Should be able to redo after undoing")
	}

	// Redo
	history.Redo(data)

	// Back to can undo but not redo
	if !history.CanUndo() {
		t.Error("Should be able to undo after redoing")
	}
	if history.CanRedo() {
		t.Error("Should not be able to redo after redoing all")
	}
}

