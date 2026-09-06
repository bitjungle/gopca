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

package transform

import (
	"fmt"
	"sort"
	"strings"
)

// applyOneHot performs one-hot encoding on the specified categorical columns.
// For each column, K new binary columns are added, one per unique value. The
// source column is kept unless opts.RemoveOriginal is set, in which case it is
// removed and the column order in headers and data is updated accordingly.
//
// Columns with more than 20 unique non-missing values are skipped to avoid
// combinatorial explosion.
func applyOneHot(data [][]string, columnTypes map[string]string, catCols map[string][]string, headers *[]string, opts Options, result *Result) error {
	for _, colName := range opts.Columns {
		colIndex := findColumn(*headers, colName)
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}

		if columnTypes[colName] != "categorical" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not categorical, skipping", colName))
			continue
		}

		// Gather unique non-empty values.
		uniqueSet := make(map[string]bool)
		for i := range data {
			if colIndex >= len(data[i]) {
				continue
			}
			v := strings.TrimSpace(data[i][colIndex])
			if v != "" {
				uniqueSet[v] = true
			}
		}

		if len(uniqueSet) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no values", colName))
			continue
		}

		if len(uniqueSet) > 20 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has too many unique values (%d), skipping one-hot encoding", colName, len(uniqueSet)))
			continue
		}

		sortedValues := make([]string, 0, len(uniqueSet))
		for v := range uniqueSet {
			sortedValues = append(sortedValues, v)
		}
		sort.Strings(sortedValues)

		// Append one new column per unique value.
		newColumns := make([]string, 0, len(sortedValues))
		for _, val := range sortedValues {
			newColName := fmt.Sprintf("%s_%s", colName, val)
			*headers = append(*headers, newColName)
			columnTypes[newColName] = "numeric"
			newColumns = append(newColumns, newColName)

			for i := range data {
				if colIndex < len(data[i]) && strings.TrimSpace(data[i][colIndex]) == val {
					data[i] = append(data[i], "1")
				} else {
					data[i] = append(data[i], "0")
				}
			}
		}

		// The source column stays unless removal was asked for. Keeping it is
		// what lets GoPCA still colour plots by the category after encoding it.
		//
		// Guarded rather than skipped with continue: the bookkeeping below
		// records the transformation either way, and jumping past it would make
		// the encoding invisible to the caller in exactly the case that is now
		// the default.
		if opts.RemoveOriginal {
			// Remove the original column from headers and each row.
			*headers = append((*headers)[:colIndex], (*headers)[colIndex+1:]...)
			delete(columnTypes, colName)
			delete(catCols, colName)

			for i := range data {
				if colIndex < len(data[i]) {
					data[i] = append(data[i][:colIndex], data[i][colIndex+1:]...)
				}
			}
		}

		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.NewColumns = append(result.NewColumns, newColumns...)
		result.Messages = append(result.Messages, fmt.Sprintf("One-hot encoded column '%s' into %d new columns", colName, len(newColumns)))
	}

	return nil
}
