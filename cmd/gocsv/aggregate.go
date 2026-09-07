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
	"sort"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// AggregateFunc names how the rows of a group are collapsed to one value.
type AggregateFunc string

const (
	AggregateMean   AggregateFunc = "mean"
	AggregateMedian AggregateFunc = "median"
	AggregateSum    AggregateFunc = "sum"
	AggregateFirst  AggregateFunc = "first"
)

// AggregateOptions describes a group-and-collapse operation.
type AggregateOptions struct {
	// GroupBy is the column whose value identifies the group.
	GroupBy string `json:"groupBy"`
	// Func is applied to the numeric columns.
	Func AggregateFunc `json:"func"`
}

// AggregatePreview says what the operation would do, without doing it.
type AggregatePreview struct {
	Groups       int `json:"groups"`
	Rows         int `json:"rows"`
	LargestSize  int `json:"largestSize"`
	SmallestSize int `json:"smallestSize"`
	// TextConflicts counts cells where a text column disagrees within a group.
	TextConflicts int    `json:"textConflicts"`
	Error         string `json:"error,omitempty"`
}

// aggregateGroups collects the row indices belonging to each distinct value of
// the grouping column, in the order the values first appear.
//
// First-appearance order rather than sorted, so the result keeps the shape of
// the original file. Sorting would silently reorder a table that may already be
// in a meaningful sequence.
func aggregateGroups(data *FileData, colIndex int) (keys []string, groups map[string][]int, blanks []int) {
	groups = map[string][]int{}
	for i, row := range data.Data {
		value := ""
		if colIndex < len(row) {
			value = strings.TrimSpace(row[colIndex])
		}
		if value == "" {
			blanks = append(blanks, i+1)
			continue
		}
		if _, seen := groups[value]; !seen {
			keys = append(keys, value)
		}
		groups[value] = append(groups[value], i)
	}
	return keys, groups, blanks
}

// validate reports why an aggregation cannot be applied, or nil.
func (o AggregateOptions) validate(data *FileData) (int, error) {
	if data == nil || len(data.Data) == 0 {
		return -1, fmt.Errorf("there is no data to aggregate")
	}

	colIndex := findColumnIndex(data.Headers, o.GroupBy)
	if colIndex == -1 {
		return -1, fmt.Errorf("no column named %q", o.GroupBy)
	}

	switch o.Func {
	case AggregateMean, AggregateMedian, AggregateSum, AggregateFirst:
	default:
		return -1, fmt.Errorf("unknown aggregate function %q", o.Func)
	}

	// A blank is not a group. Averaging the unlabelled rows together would
	// invent a sample, and dropping them silently would lose data, so the
	// operation stops and says which rows to deal with first -- Filter Rows
	// removes them in one step.
	if _, _, blanks := aggregateGroups(data, colIndex); len(blanks) > 0 {
		return -1, fmt.Errorf("%d row(s) have no value in %q (%s). A blank is not a "+
			"group: averaging them together would invent a sample and dropping them "+
			"would lose data. Remove or label them first -- Filter Rows will do it in "+
			"one step", len(blanks), o.GroupBy, describeRowNumbers(blanks))
	}

	return colIndex, nil
}

// PreviewAggregate reports what an aggregation would produce.
//
// The group count is the number worth seeing before committing: if it equals
// the row count, every group has one member and the operation would change
// nothing, which is what a mistyped grouping column looks like.
func (a *App) PreviewAggregate(data *FileData, options AggregateOptions) AggregatePreview {
	colIndex, err := options.validate(data)
	if err != nil {
		return AggregatePreview{Error: err.Error()}
	}

	keys, groups, _ := aggregateGroups(data, colIndex)
	preview := AggregatePreview{
		Groups:       len(keys),
		Rows:         len(data.Data),
		SmallestSize: len(data.Data),
	}
	for _, key := range keys {
		size := len(groups[key])
		if size > preview.LargestSize {
			preview.LargestSize = size
		}
		if size < preview.SmallestSize {
			preview.SmallestSize = size
		}
	}

	for j, header := range data.Headers {
		if j == colIndex || data.ColumnTypes[header] == "numeric" || data.ColumnTypes[header] == "target" {
			continue
		}
		for _, key := range keys {
			if _, agreed := singleTextValue(data, groups[key], j); !agreed {
				preview.TextConflicts++
			}
		}
	}

	return preview
}

// singleTextValue returns the one value a text column takes across a group, and
// whether the group agrees on it.
//
// Blanks are ignored when deciding agreement: a group where one replicate
// recorded the operator and the others left it empty agrees on the operator.
func singleTextValue(data *FileData, rows []int, colIndex int) (string, bool) {
	found := ""
	seen := false
	for _, i := range rows {
		if colIndex >= len(data.Data[i]) {
			continue
		}
		value := strings.TrimSpace(data.Data[i][colIndex])
		if value == "" {
			continue
		}
		if !seen {
			found, seen = value, true
			continue
		}
		if value != found {
			return "", false
		}
	}
	return found, true
}

// AggregateRowsCommand collapses each group of rows to a single row.
type AggregateRowsCommand struct {
	app     *App
	options AggregateOptions
	before  *FileData

	groups    int
	conflicts int
}

// NewAggregateRowsCommand validates the request and captures the pre-state.
func NewAggregateRowsCommand(app *App, data *FileData, options AggregateOptions) (*AggregateRowsCommand, error) {
	if _, err := options.validate(data); err != nil {
		return nil, err
	}
	return &AggregateRowsCommand{
		app:     app,
		options: options,
		before:  deepCopyFileData(data),
	}, nil
}

// Execute collapses the groups.
//
// Averaging replicate measurements to one row per sample is standard practice,
// and it also removes a hazard rather than mitigating one: replicates left as
// separate rows leak between cross-validation folds, so the same sample appears
// in both training and validation unless the user remembers to set --cv-group.
func (c *AggregateRowsCommand) Execute(data *FileData) error {
	colIndex, err := c.options.validate(data)
	if err != nil {
		return err
	}

	keys, groups, _ := aggregateGroups(data, colIndex)
	c.groups = len(keys)

	newData := make([][]string, 0, len(keys))
	newRowNames := make([]string, 0, len(keys))

	for _, key := range keys {
		rows := groups[key]
		out := make([]string, len(data.Headers))

		for j, header := range data.Headers {
			switch {
			case j == colIndex:
				// Written here and removed with the column below, so the loop
				// stays a simple pass over the original headers.
				out[j] = key

			case data.ColumnTypes[header] == "numeric" || data.ColumnTypes[header] == "target":
				out[j] = aggregateNumeric(data, rows, j, c.options.Func)

			default:
				// A text column keeps the value its group agrees on. Where the
				// group disagrees the cell is left empty and counted, because
				// picking one of the competing values would fabricate a fact
				// about the aggregated sample that no row asserted.
				value, agreed := singleTextValue(data, rows, j)
				if !agreed {
					c.conflicts++
					value = ""
				}
				out[j] = value
			}
		}

		newData = append(newData, out)
		newRowNames = append(newRowNames, key)
	}

	data.Data = newData
	data.Rows = len(newData)

	// Row names become the group values. They have to change -- the rows they
	// named no longer exist -- and the group value is what identifies the new
	// row. It is unique by construction, which is what row names require (#859).
	data.RowNames = newRowNames
	data.RowNamesHeader = c.options.GroupBy

	// The grouping column is then taken out of the table, because it now holds
	// exactly the row names and under the same heading. Leaving it produced two
	// columns with identical headers and identical values, which reads as a
	// mistake rather than as thoroughness. Nothing is lost: the values are the
	// row names, and "Move Row Names into Table" puts them back as a column.
	removeColumnAt(data, colIndex)
	delete(data.ColumnTypes, c.options.GroupBy)

	// The per-row maps are rebuilt from the collapsed table rather than
	// filtered, since their rows have been combined rather than removed.
	rebuildPerRowMaps(data)

	return nil
}

// aggregateNumeric applies the chosen function to one column of one group.
//
// Non-numeric and blank cells are skipped rather than treated as zero. A gap is
// not a measurement of nothing, and averaging it in as zero would drag the
// result towards zero in proportion to how much data is missing.
func aggregateNumeric(data *FileData, rows []int, colIndex int, fn AggregateFunc) string {
	values := make([]float64, 0, len(rows))
	for _, i := range rows {
		if colIndex >= len(data.Data[i]) {
			continue
		}
		text := strings.TrimSpace(data.Data[i][colIndex])
		if text == "" {
			continue
		}
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			continue
		}
		values = append(values, number)
	}

	if len(values) == 0 {
		// Every value in the group was missing, so the aggregate is missing
		// too. Writing 0 would turn absence into a measurement.
		return ""
	}

	switch fn {
	case AggregateFirst:
		return formatAggregate(values[0])
	case AggregateSum:
		total := 0.0
		for _, v := range values {
			total += v
		}
		return formatAggregate(total)
	case AggregateMedian:
		sorted := append([]float64(nil), values...)
		sort.Float64s(sorted)
		mid := len(sorted) / 2
		if len(sorted)%2 == 1 {
			return formatAggregate(sorted[mid])
		}
		return formatAggregate((sorted[mid-1] + sorted[mid]) / 2)
	default: // mean
		total := 0.0
		for _, v := range values {
			total += v
		}
		return formatAggregate(total / float64(len(values)))
	}
}

// formatAggregate renders an aggregated value.
//
// Twelve significant figures rather than six: an average of replicates carries
// more resolution than any one of them, and rounding it here would discard
// precision the analysis can still use.
func formatAggregate(value float64) string {
	return strconv.FormatFloat(value, 'g', 12, 64)
}

// rebuildPerRowMaps regenerates the categorical and target maps from the table.
//
// Aggregation combines rows rather than removing them, so the old per-row
// entries do not correspond to anything in the new table and cannot be
// filtered. They are rebuilt from the collapsed data, which is the only source
// that now describes it.
func rebuildPerRowMaps(data *FileData) {
	for column := range data.CategoricalColumns {
		colIndex := findColumnIndex(data.Headers, column)
		if colIndex == -1 {
			delete(data.CategoricalColumns, column)
			continue
		}
		values := make([]string, len(data.Data))
		for i, row := range data.Data {
			if colIndex < len(row) {
				values[i] = row[colIndex]
			}
		}
		data.CategoricalColumns[column] = values
	}

	for column := range data.NumericTargetColumns {
		colIndex := findColumnIndex(data.Headers, column)
		if colIndex == -1 {
			delete(data.NumericTargetColumns, column)
			continue
		}
		values := make([]types.JSONFloat64, len(data.Data))
		for i, row := range data.Data {
			if colIndex >= len(row) {
				continue
			}
			number, err := strconv.ParseFloat(strings.TrimSpace(row[colIndex]), 64)
			if err != nil {
				continue
			}
			values[i] = types.JSONFloat64(number)
		}
		data.NumericTargetColumns[column] = values
	}
}

// Undo restores the rows as they were.
func (c *AggregateRowsCommand) Undo(data *FileData) error {
	restored := deepCopyFileData(c.before)
	*data = *restored
	return nil
}

// GetDescription implements Command.
func (c *AggregateRowsCommand) GetDescription() string {
	noun := "groups"
	if c.groups == 1 {
		noun = "group"
	}
	description := fmt.Sprintf("Aggregated %d rows into %d %s by %s (%s)",
		len(c.before.Data), c.groups, noun, c.options.GroupBy, c.options.Func)
	if c.conflicts > 0 {
		description += fmt.Sprintf(", %d text cell(s) cleared where a group disagreed",
			c.conflicts)
	}
	return description
}

// describeRowNumbers renders a list of row numbers, listing the first few.
func describeRowNumbers(rows []int) string {
	const maxListed = 5
	listed := rows
	suffix := ""
	if len(rows) > maxListed {
		listed = rows[:maxListed]
		suffix = fmt.Sprintf(" and %d more", len(rows)-maxListed)
	}
	numbers := make([]string, len(listed))
	for i, row := range listed {
		numbers[i] = strconv.Itoa(row)
	}
	return fmt.Sprintf("row %s%s", strings.Join(numbers, ", "), suffix)
}
