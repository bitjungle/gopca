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

// FilterOperator names a comparison a row filter can make.
type FilterOperator string

const (
	FilterEquals       FilterOperator = "equals"
	FilterNotEquals    FilterOperator = "not_equals"
	FilterContains     FilterOperator = "contains"
	FilterNotContains  FilterOperator = "not_contains"
	FilterGreater      FilterOperator = "greater"
	FilterGreaterEqual FilterOperator = "greater_equal"
	FilterLess         FilterOperator = "less"
	FilterLessEqual    FilterOperator = "less_equal"
	FilterIsEmpty      FilterOperator = "is_empty"
	FilterIsNotEmpty   FilterOperator = "is_not_empty"
)

// FilterCondition describes which rows a filter selects.
//
// Mode decides what happens to them: "keep" discards everything else, "remove"
// discards the matches. Both are offered because both are natural -- "analyse
// batch 3 only" and "drop the QC standards" are the same operation seen from
// opposite sides, and making the user invert one of them by hand invites the
// mistake of inverting it wrongly.
type FilterCondition struct {
	Column   string         `json:"column"`
	Operator FilterOperator `json:"operator"`
	Value    string         `json:"value"`
	Mode     string         `json:"mode"` // "keep" or "remove"
}

// matches reports whether one cell satisfies the condition.
//
// The rule for empty cells is that they match only is_empty, and is_not_empty
// excludes them. Every other operator treats a blank as no answer rather than
// as a value.
//
// This matters most for the negations. Under a plain string comparison an empty
// cell is "not equal to" everything, so "remove rows where Region is not Nord"
// would quietly take the unlabelled rows too -- a filter about Region deleting
// rows on the strength of having no Region at all. Missing is not a value, and
// deciding a row's fate on the absence of one should be something the user asks
// for explicitly.
func (c FilterCondition) matches(cell string) bool {
	trimmed := strings.TrimSpace(cell)

	switch c.Operator {
	case FilterIsEmpty:
		return trimmed == ""
	case FilterIsNotEmpty:
		return trimmed != ""
	}

	if trimmed == "" {
		return false
	}

	// Numeric comparison when both sides are numbers, so "5" and "5.0" compare
	// equal and the ordering operators mean what they say. Falling back to text
	// keeps the operators usable on categorical columns.
	cellNum, cellIsNum := parseNumber(trimmed)
	wantNum, wantIsNum := parseNumber(strings.TrimSpace(c.Value))
	bothNumeric := cellIsNum && wantIsNum

	switch c.Operator {
	case FilterEquals:
		if bothNumeric {
			return cellNum == wantNum
		}
		return strings.EqualFold(trimmed, strings.TrimSpace(c.Value))
	case FilterNotEquals:
		if bothNumeric {
			return cellNum != wantNum
		}
		return !strings.EqualFold(trimmed, strings.TrimSpace(c.Value))
	case FilterContains:
		return strings.Contains(strings.ToLower(trimmed), strings.ToLower(strings.TrimSpace(c.Value)))
	case FilterNotContains:
		return !strings.Contains(strings.ToLower(trimmed), strings.ToLower(strings.TrimSpace(c.Value)))
	}

	// The ordering operators are only meaningful on numbers. A cell that is not
	// a number never matches one, rather than falling back to string ordering,
	// where "10" sorts before "9".
	if !bothNumeric {
		return false
	}
	switch c.Operator {
	case FilterGreater:
		return cellNum > wantNum
	case FilterGreaterEqual:
		return cellNum >= wantNum
	case FilterLess:
		return cellNum < wantNum
	case FilterLessEqual:
		return cellNum <= wantNum
	}
	return false
}

func parseNumber(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

// validate reports why a condition cannot be applied, or nil.
func (c FilterCondition) validate(data *FileData) error {
	if data == nil || len(data.Data) == 0 {
		return fmt.Errorf("there is no data to filter")
	}
	if findColumnIndex(data.Headers, c.Column) == -1 {
		return fmt.Errorf("no column named %q", c.Column)
	}
	if c.Mode != "keep" && c.Mode != "remove" {
		return fmt.Errorf("mode must be \"keep\" or \"remove\", got %q", c.Mode)
	}

	switch c.Operator {
	case FilterEquals, FilterNotEquals, FilterContains, FilterNotContains,
		FilterGreater, FilterGreaterEqual, FilterLess, FilterLessEqual,
		FilterIsEmpty, FilterIsNotEmpty:
	default:
		return fmt.Errorf("unknown operator %q", c.Operator)
	}

	// The ordering operators need something to compare against.
	switch c.Operator {
	case FilterGreater, FilterGreaterEqual, FilterLess, FilterLessEqual:
		if _, ok := parseNumber(strings.TrimSpace(c.Value)); !ok {
			return fmt.Errorf("%q is not a number, so it cannot be compared with %s",
				c.Value, c.Operator)
		}
	}
	return nil
}

// findColumnIndex returns the index of a header, or -1.
func findColumnIndex(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}

// selectedRows returns the indices the condition matches.
func selectedRows(data *FileData, c FilterCondition) []int {
	colIndex := findColumnIndex(data.Headers, c.Column)
	if colIndex == -1 {
		return nil
	}

	var matched []int
	for i, row := range data.Data {
		cell := ""
		if colIndex < len(row) {
			cell = row[colIndex]
		}
		if c.matches(cell) {
			matched = append(matched, i)
		}
	}
	return matched
}

// FilterPreview says what a filter would do, without doing it.
type FilterPreview struct {
	Matched   int    `json:"matched"`
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	Error     string `json:"error,omitempty"`
}

// PreviewFilter reports how many rows a condition selects and how many would
// survive it.
//
// A filter that would empty the table, or one that would change nothing, is
// worth seeing before it is applied rather than after.
func (a *App) PreviewFilter(data *FileData, condition FilterCondition) FilterPreview {
	if err := condition.validate(data); err != nil {
		return FilterPreview{Error: err.Error()}
	}

	matched := len(selectedRows(data, condition))
	total := len(data.Data)
	remaining := matched
	if condition.Mode == "remove" {
		remaining = total - matched
	}
	return FilterPreview{Matched: matched, Total: total, Remaining: remaining}
}

// FilterRowsCommand keeps or removes the rows a condition selects.
type FilterRowsCommand struct {
	app       *App
	condition FilterCondition
	before    *FileData
	removed   int
}

// NewFilterRowsCommand validates the condition and captures the pre-state.
func NewFilterRowsCommand(app *App, data *FileData, condition FilterCondition) (*FilterRowsCommand, error) {
	if err := condition.validate(data); err != nil {
		return nil, err
	}
	return &FilterRowsCommand{
		app:       app,
		condition: condition,
		before:    deepCopyFileData(data),
	}, nil
}

// Execute applies the filter.
func (c *FilterRowsCommand) Execute(data *FileData) error {
	matched := map[int]bool{}
	for _, i := range selectedRows(data, c.condition) {
		matched[i] = true
	}

	keep := make([]int, 0, len(data.Data))
	for i := range data.Data {
		if matched[i] == (c.condition.Mode == "keep") {
			keep = append(keep, i)
		}
	}
	c.removed = len(data.Data) - len(keep)

	newData := make([][]string, 0, len(keep))
	for _, i := range keep {
		newData = append(newData, data.Data[i])
	}
	data.Data = newData

	if len(data.RowNames) > 0 {
		newNames := make([]string, 0, len(keep))
		for _, i := range keep {
			if i < len(data.RowNames) {
				newNames = append(newNames, data.RowNames[i])
			}
		}
		data.RowNames = newNames
	}

	// The categorical and target maps hold one entry per row, parallel to Data.
	// Dropping rows without dropping the matching entries leaves them longer
	// than the table and silently misaligned -- every value attached to the
	// wrong row from the first deletion onwards.
	for column, values := range data.CategoricalColumns {
		filtered := make([]string, 0, len(keep))
		for _, i := range keep {
			if i < len(values) {
				filtered = append(filtered, values[i])
			}
		}
		data.CategoricalColumns[column] = filtered
	}
	for column, values := range data.NumericTargetColumns {
		filtered := make([]types.JSONFloat64, 0, len(keep))
		for _, i := range keep {
			if i < len(values) {
				filtered = append(filtered, values[i])
			}
		}
		data.NumericTargetColumns[column] = filtered
	}

	data.Rows = len(data.Data)
	return nil
}

// Undo restores the rows.
func (c *FilterRowsCommand) Undo(data *FileData) error {
	restored := deepCopyFileData(c.before)
	*data = *restored
	return nil
}

// GetDescription implements Command.
func (c *FilterRowsCommand) GetDescription() string {
	verb := "Keep"
	if c.condition.Mode == "remove" {
		verb = "Remove"
	}
	return fmt.Sprintf("%s rows where %s %s %s (%d removed)",
		verb, c.condition.Column, c.condition.Operator, c.condition.Value, c.removed)
}
