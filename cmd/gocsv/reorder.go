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
	"fmt"
)

// ReorderColumnsCommand rearranges the columns of the table.
//
// Column order is a property of the data, not of the view. It decides the order
// of an export, the order GoPCA receives, and the order a loadings plot lists
// variables in -- so a reordering the user performs has to reach the data or it
// is a lie told by the grid.
//
// Before this, dragging a column header moved it on screen and nothing else:
// AG Grid columns are movable by default and nothing was listening, so the
// change was discarded at the next render and absent from every export. The
// documentation recommended the gesture (#878).
type ReorderColumnsCommand struct {
	app *App

	// order[i] is the index, in the columns as they were, of the column that
	// should end up at position i.
	order  []int
	before *FileData
}

// NewReorderColumnsCommand validates the permutation and captures the state.
//
// The order is required to be a permutation of exactly the existing columns.
// Anything else -- a repeat, a gap, a wrong length -- would silently duplicate
// or drop a column, and a reorder that loses a variable is far worse than one
// that fails.
func NewReorderColumnsCommand(app *App, data *FileData, order []int) (*ReorderColumnsCommand, error) {
	if data == nil {
		return nil, fmt.Errorf("there is no data to reorder")
	}
	if len(order) != len(data.Headers) {
		return nil, fmt.Errorf("the new order lists %d columns but the table has %d",
			len(order), len(data.Headers))
	}

	seen := make([]bool, len(data.Headers))
	for _, index := range order {
		if index < 0 || index >= len(data.Headers) {
			return nil, fmt.Errorf("column index %d is outside the table", index)
		}
		if seen[index] {
			return nil, fmt.Errorf("column %q appears twice in the new order",
				data.Headers[index])
		}
		seen[index] = true
	}

	return &ReorderColumnsCommand{
		app:    app,
		order:  append([]int(nil), order...),
		before: deepCopyFileData(data),
	}, nil
}

// Execute rearranges the headers and every row.
//
// The per-row maps are keyed by column name rather than position, so they need
// no attention here -- unlike the row-removing operations, where forgetting
// them left the table misaligned (#871).
func (c *ReorderColumnsCommand) Execute(data *FileData) error {
	if len(c.order) != len(data.Headers) {
		return fmt.Errorf("the table has changed shape since this reorder was prepared")
	}

	headers := make([]string, len(c.order))
	for i, from := range c.order {
		headers[i] = data.Headers[from]
	}

	for rowIndex, row := range data.Data {
		reordered := make([]string, len(c.order))
		for i, from := range c.order {
			if from < len(row) {
				reordered[i] = row[from]
			}
		}
		data.Data[rowIndex] = reordered
	}

	data.Headers = headers
	data.Columns = len(headers)
	return nil
}

// Undo restores the previous arrangement.
func (c *ReorderColumnsCommand) Undo(data *FileData) error {
	restored := deepCopyFileData(c.before)
	*data = *restored
	return nil
}

// GetDescription implements Command.
func (c *ReorderColumnsCommand) GetDescription() string {
	// Name the column that actually moved, rather than reciting a permutation.
	// A drag moves one column; saying which one is what makes the history
	// entry readable.
	moved := ""
	for i, from := range c.order {
		if i != from {
			moved = c.before.Headers[from]
			break
		}
	}
	if moved == "" {
		return "Reordered columns"
	}
	return fmt.Sprintf("Moved column '%s'", moved)
}
