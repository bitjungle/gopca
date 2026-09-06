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
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// TransposeCommand exchanges the rows and columns of the whole dataset.
//
// Instrument software commonly exports with samples across the top and
// variables down the side -- one column per spectrum, one row per wavelength --
// which is the opposite of what PCA needs. Without this the user has to fix the
// file elsewhere before GoCSV can be useful, which is the step GoCSV exists to
// remove (#862).
//
// The corner cell stays the corner cell. Laid out, the mapping is:
//
//	RowNamesHeader | Headers[0]  | Headers[1]        RowNamesHeader | RowNames[0] | RowNames[1]
//	RowNames[0]    | Data[0][0]  | Data[0][1]   ->   Headers[0]     | Data[0][0]  | Data[1][0]
//	RowNames[1]    | Data[1][0]  | Data[1][1]        Headers[1]     | Data[0][1]  | Data[1][1]
//
// so RowNamesHeader is unchanged, the headers become the row names, and the row
// names become the headers. That symmetry is why transposing twice returns the
// original, and it only works because the row-name column has a header to
// occupy the corner (#859).
type TransposeCommand struct {
	app *App

	// The whole pre-state. Transposition rewrites every field, and a dataset
	// that has been transposed cannot be reconstructed field by field: column
	// types are recomputed rather than carried, so the originals have to be
	// kept to undo.
	before *FileData
}

// NewTransposeCommand captures the pre-state.
func NewTransposeCommand(app *App, data *FileData) (*TransposeCommand, error) {
	if data == nil || len(data.Data) == 0 || len(data.Headers) == 0 {
		return nil, fmt.Errorf("there is no data to transpose")
	}
	return &TransposeCommand{app: app, before: cloneFileData(data)}, nil
}

// Execute transposes the dataset in place.
func (c *TransposeCommand) Execute(data *FileData) error {
	source := c.before

	// The new headers come from the old row names. A file with none gets
	// generated names rather than an empty header row, because a blank column
	// header cannot be selected in a dialog or named in a message.
	// One map for the whole loop. uniqueHeader rebuilds its lookup from the
	// slice it is given, which is quadratic when called once per column -- and
	// the files this feature exists for are the wide ones, where a 2000-column
	// spectrum becomes 2000 rows.
	newHeaders := make([]string, 0, len(source.Data))
	taken := make(map[string]bool, len(source.Data))
	nextSuffix := make(map[string]int, len(source.Data))
	for i := range source.Data {
		name := ""
		if i < len(source.RowNames) {
			name = strings.TrimSpace(source.RowNames[i])
		}
		if name == "" {
			name = fmt.Sprintf("Row_%d", i+1)
		}
		// Row names are not guaranteed unique -- the load path assigns the
		// first column without checking (#859) -- but headers addressed by name
		// must be. Suffix collisions rather than refusing: the user asked to
		// transpose, and a duplicate label is not a reason to decline.
		// nextSuffix remembers where the last search for this name ended, so a
		// column of identical row names does not rescan the suffixes it has
		// already used. The membership test stays, because a generated name can
		// still collide with a literal one somewhere else in the file.
		unique := name
		for taken[unique] {
			nextSuffix[name]++
			unique = fmt.Sprintf("%s_%d", name, nextSuffix[name]+1)
		}
		taken[unique] = true
		newHeaders = append(newHeaders, unique)
	}

	// The new row names come from the old headers, which were already unique.
	newRowNames := make([]string, len(source.Headers))
	copy(newRowNames, source.Headers)

	// Cell (i,j) becomes cell (j,i). Short rows read as empty rather than
	// shifting the remaining values up a column.
	newData := make([][]string, len(source.Headers))
	for i := range newData {
		row := make([]string, len(source.Data))
		for j := range source.Data {
			if i < len(source.Data[j]) {
				row[j] = source.Data[j][i]
			}
		}
		newData[i] = row
	}

	data.Headers = newHeaders
	data.RowNames = newRowNames
	data.RowNamesHeader = source.RowNamesHeader
	data.Data = newData
	data.Rows = len(newData)
	data.Columns = len(newHeaders)

	// Types are recomputed, never carried across. A transposed dataset has
	// entirely different columns: a row that mixed text and numbers becomes a
	// column that does, and inherits nothing meaningful from the column it
	// used to sit in.
	data.ColumnTypes = map[string]string{}
	data.CategoricalColumns = map[string][]string{}
	data.NumericTargetColumns = nil
	for j, header := range newHeaders {
		values := make([]string, len(newData))
		for i := range newData {
			values[i] = newData[i][j]
		}
		classifyColumn(data, header, values)
	}

	return nil
}

// Undo restores the dataset captured before the transposition.
func (c *TransposeCommand) Undo(data *FileData) error {
	restored := cloneFileData(c.before)
	*data = *restored
	return nil
}

// GetDescription implements Command.
func (c *TransposeCommand) GetDescription() string {
	return "Transpose rows and columns"
}

// cloneFileData deep-copies the fields transposition rewrites.
//
// A shallow copy would leave the undo state sharing slices with the live data,
// so the "before" would be edited alongside the "after" and undo would restore
// the transposed values.
func cloneFileData(data *FileData) *FileData {
	clone := &FileData{
		Rows:           data.Rows,
		Columns:        data.Columns,
		RowNamesHeader: data.RowNamesHeader,
	}

	clone.Headers = append([]string(nil), data.Headers...)
	clone.RowNames = append([]string(nil), data.RowNames...)

	clone.Data = make([][]string, len(data.Data))
	for i, row := range data.Data {
		clone.Data[i] = append([]string(nil), row...)
	}

	if data.ColumnTypes != nil {
		clone.ColumnTypes = make(map[string]string, len(data.ColumnTypes))
		for k, v := range data.ColumnTypes {
			clone.ColumnTypes[k] = v
		}
	}
	if data.CategoricalColumns != nil {
		clone.CategoricalColumns = make(map[string][]string, len(data.CategoricalColumns))
		for k, v := range data.CategoricalColumns {
			clone.CategoricalColumns[k] = append([]string(nil), v...)
		}
	}
	if data.NumericTargetColumns != nil {
		clone.NumericTargetColumns = make(map[string][]types.JSONFloat64, len(data.NumericTargetColumns))
		for k, v := range data.NumericTargetColumns {
			clone.NumericTargetColumns[k] = append([]types.JSONFloat64(nil), v...)
		}
	}
	return clone
}

// TransposeWarnings reports what transposing this dataset would cost, so the
// user can decide before it happens rather than discover it afterwards.
//
// Bound for the frontend. It reports rather than refuses: transposing is a
// legitimate thing to do to any table, and the consequences below are
// consequences, not errors.
func (a *App) TransposeWarnings(data *FileData) []string {
	if data == nil {
		return []string{}
	}

	warnings := []string{}

	// A #target column marks a variable as reference information. After
	// transposition that variable is a row, and the suffix ends up in a row
	// name where nothing reads it -- the column is no longer a column, so it
	// can be neither excluded from PCA nor used as a regression response.
	targets := []string{}
	for _, header := range data.Headers {
		lower := strings.ToLower(header)
		if strings.HasSuffix(lower, "#target") || strings.HasSuffix(lower, " #target") {
			targets = append(targets, header)
		}
	}
	if len(targets) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%s will become row names, and the #target marking will no longer apply "+
				"to anything: a target has to be a column.", strings.Join(targets, ", ")))
	}

	// Duplicate row names become duplicate headers, which have to be suffixed.
	seen := map[string]int{}
	duplicates := 0
	for _, name := range data.RowNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		seen[trimmed]++
		if seen[trimmed] == 2 {
			duplicates++
		}
	}
	if duplicates > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%d row name(s) repeat and will become column names, so the duplicates "+
				"will be suffixed to keep them distinct.", duplicates))
	}

	blanks := 0
	for i := range data.Data {
		if i >= len(data.RowNames) || strings.TrimSpace(data.RowNames[i]) == "" {
			blanks++
		}
	}
	if blanks > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%d row(s) have no name, so the resulting columns will be called Row_1, "+
				"Row_2 and so on.", blanks))
	}

	// The shape swap is worth stating plainly for wide files.
	warnings = append(warnings, fmt.Sprintf(
		"The result will have %d rows and %d columns, from %d and %d.",
		len(data.Headers), len(data.Data), len(data.Data), len(data.Headers)))

	return warnings
}
