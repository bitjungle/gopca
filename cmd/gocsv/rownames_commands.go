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
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// SetRowNamesCommand promotes a column to be the row-name column.
//
// The operation is a swap, not a move: whatever was serving as row names comes
// back as column 0, carrying the header it was read with. Nothing is discarded,
// which is what makes the menu item safe to try -- correcting a bad guess by
// the loader is the reason it exists, and an irreversible correction is not
// much of one. It is the same principle applied to one-hot encoding in #854.
type SetRowNamesCommand struct {
	app      *App
	colIndex int

	// Everything needed to put the world back.
	prevRowNames       []string
	prevRowNamesHeader string
	promotedHeader     string
	promotedValues     []string
	promotedType       string
	promotedCategories []string
	promotedTargets    []types.JSONFloat64
	demotedHeader      string // "" when there were no previous row names
}

// NewSetRowNamesCommand captures the pre-state. It returns an error rather than
// a nil command so the caller can report why.
func NewSetRowNamesCommand(app *App, data *FileData, colIndex int) (*SetRowNamesCommand, error) {
	if data == nil || colIndex < 0 || colIndex >= len(data.Headers) {
		return nil, fmt.Errorf("invalid column index: %d", colIndex)
	}

	values := columnValues(data, colIndex)
	if check := checkRowNameCandidate(values); !check.OK {
		return nil, fmt.Errorf("column %q cannot be used as row names: %s",
			data.Headers[colIndex], check.Reason)
	}

	header := data.Headers[colIndex]
	cmd := &SetRowNamesCommand{
		app:                app,
		colIndex:           colIndex,
		prevRowNames:       append([]string(nil), data.RowNames...),
		prevRowNamesHeader: data.RowNamesHeader,
		promotedHeader:     header,
		promotedValues:     values,
		promotedType:       data.ColumnTypes[header],
	}
	if categories, ok := data.CategoricalColumns[header]; ok {
		cmd.promotedCategories = append([]string(nil), categories...)
	}
	if targets, ok := data.NumericTargetColumns[header]; ok {
		cmd.promotedTargets = append([]types.JSONFloat64(nil), targets...)
	}
	return cmd, nil
}

// Execute promotes the column and demotes the previous row names.
func (c *SetRowNamesCommand) Execute(data *FileData) error {
	if c.colIndex >= len(data.Headers) {
		return fmt.Errorf("invalid column index: %d", c.colIndex)
	}

	// Take the column out first, so the index still refers to what it did when
	// the command was built. Inserting the demoted column beforehand would
	// shift everything right by one.
	removeColumnAt(data, c.colIndex)
	delete(data.ColumnTypes, c.promotedHeader)
	delete(data.CategoricalColumns, c.promotedHeader)
	delete(data.NumericTargetColumns, c.promotedHeader)

	data.RowNames = append([]string(nil), c.promotedValues...)
	data.RowNamesHeader = c.promotedHeader

	// The previous row names become column 0 -- where they came from, and where
	// they would be written on export.
	if len(c.prevRowNames) > 0 {
		c.demotedHeader = uniqueHeader(data.Headers, defaultRowNameHeader(c.prevRowNamesHeader))
		insertColumnAt(data, 0, c.demotedHeader, c.prevRowNames)
		classifyColumn(data, c.demotedHeader, c.prevRowNames)
	} else {
		c.demotedHeader = ""
	}

	data.Columns = len(data.Headers)
	return nil
}

// Undo restores the column and the previous row names.
func (c *SetRowNamesCommand) Undo(data *FileData) error {
	if c.demotedHeader != "" {
		removeColumnAt(data, 0)
		delete(data.ColumnTypes, c.demotedHeader)
		delete(data.CategoricalColumns, c.demotedHeader)
	}

	insertColumnAt(data, c.colIndex, c.promotedHeader, c.promotedValues)
	if c.promotedType != "" {
		data.ColumnTypes[c.promotedHeader] = c.promotedType
	}
	if c.promotedCategories != nil {
		if data.CategoricalColumns == nil {
			data.CategoricalColumns = map[string][]string{}
		}
		data.CategoricalColumns[c.promotedHeader] = c.promotedCategories
	}
	if c.promotedTargets != nil {
		if data.NumericTargetColumns == nil {
			data.NumericTargetColumns = map[string][]types.JSONFloat64{}
		}
		data.NumericTargetColumns[c.promotedHeader] = c.promotedTargets
	}

	data.RowNames = append([]string(nil), c.prevRowNames...)
	data.RowNamesHeader = c.prevRowNamesHeader
	data.Columns = len(data.Headers)
	return nil
}

// GetDescription implements Command.
func (c *SetRowNamesCommand) GetDescription() string {
	return fmt.Sprintf("Use '%s' as row names", c.promotedHeader)
}

// MoveRowNamesIntoTableCommand turns the row-name column back into an ordinary
// column, leaving the table without row names.
type MoveRowNamesIntoTableCommand struct {
	app *App

	prevRowNames       []string
	prevRowNamesHeader string
	insertedHeader     string
}

// NewMoveRowNamesIntoTableCommand captures the pre-state.
func NewMoveRowNamesIntoTableCommand(app *App, data *FileData) (*MoveRowNamesIntoTableCommand, error) {
	if data == nil || len(data.RowNames) == 0 {
		return nil, fmt.Errorf("this file has no row names")
	}
	return &MoveRowNamesIntoTableCommand{
		app:                app,
		prevRowNames:       append([]string(nil), data.RowNames...),
		prevRowNamesHeader: data.RowNamesHeader,
	}, nil
}

// Execute moves the row names into column 0.
func (c *MoveRowNamesIntoTableCommand) Execute(data *FileData) error {
	c.insertedHeader = uniqueHeader(data.Headers, defaultRowNameHeader(c.prevRowNamesHeader))
	insertColumnAt(data, 0, c.insertedHeader, c.prevRowNames)
	classifyColumn(data, c.insertedHeader, c.prevRowNames)

	data.RowNames = nil
	data.RowNamesHeader = ""
	data.Columns = len(data.Headers)
	return nil
}

// Undo puts the row names back.
func (c *MoveRowNamesIntoTableCommand) Undo(data *FileData) error {
	removeColumnAt(data, 0)
	delete(data.ColumnTypes, c.insertedHeader)
	delete(data.CategoricalColumns, c.insertedHeader)

	data.RowNames = append([]string(nil), c.prevRowNames...)
	data.RowNamesHeader = c.prevRowNamesHeader
	data.Columns = len(data.Headers)
	return nil
}

// GetDescription implements Command.
func (c *MoveRowNamesIntoTableCommand) GetDescription() string {
	return "Move row names into the table"
}

// defaultRowNameHeader supplies a name for a row-name column that never had
// one. A blank header is the common CSV convention, but a blank *column* header
// in the grid is not addressable -- it cannot be referred to in a transform
// dialog or a validation message.
func defaultRowNameHeader(header string) string {
	if strings.TrimSpace(header) == "" {
		return "RowName"
	}
	return header
}

// uniqueHeader returns name, suffixed until it collides with nothing in taken.
func uniqueHeader(taken []string, name string) string {
	inUse := make(map[string]bool, len(taken))
	for _, header := range taken {
		inUse[header] = true
	}
	candidate := name
	for i := 2; inUse[candidate]; i++ {
		candidate = fmt.Sprintf("%s_%d", name, i)
	}
	return candidate
}

// insertColumnAt inserts a column with the given header and values at index.
func insertColumnAt(data *FileData, index int, header string, values []string) {
	if index < 0 {
		index = 0
	}
	if index > len(data.Headers) {
		index = len(data.Headers)
	}

	headers := make([]string, 0, len(data.Headers)+1)
	headers = append(headers, data.Headers[:index]...)
	headers = append(headers, header)
	headers = append(headers, data.Headers[index:]...)
	data.Headers = headers

	for i := range data.Data {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		at := index
		if at > len(data.Data[i]) {
			at = len(data.Data[i])
		}
		row := make([]string, 0, len(data.Data[i])+1)
		row = append(row, data.Data[i][:at]...)
		row = append(row, value)
		row = append(row, data.Data[i][at:]...)
		data.Data[i] = row
	}
}

// removeColumnAt drops the column at index from the headers and every row.
func removeColumnAt(data *FileData, index int) {
	if index < 0 || index >= len(data.Headers) {
		return
	}
	data.Headers = append(data.Headers[:index:index], data.Headers[index+1:]...)
	for i := range data.Data {
		if index < len(data.Data[i]) {
			data.Data[i] = append(data.Data[i][:index:index], data.Data[i][index+1:]...)
		}
	}
}

// classifyColumn records the type of a column newly added to the table.
//
// Values that all parse as numbers are numeric; anything else is categorical,
// and categorical columns additionally live in their own map. This mirrors what
// the loader decides for the same values, so a column demoted out of the
// row-name slot is typed as it would have been had it been read as data.
func classifyColumn(data *FileData, header string, values []string) {
	if data.ColumnTypes == nil {
		data.ColumnTypes = map[string]string{}
	}

	numeric := len(values) > 0
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
			numeric = false
			break
		}
	}

	if numeric {
		data.ColumnTypes[header] = "numeric"
		return
	}

	data.ColumnTypes[header] = "categorical"
	if data.CategoricalColumns == nil {
		data.CategoricalColumns = map[string][]string{}
	}
	data.CategoricalColumns[header] = append([]string(nil), values...)
}
