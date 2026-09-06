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
)

// RowNameCheck reports whether a column can serve as the row-name column, and
// says why not when it cannot.
//
// The frontend uses this to disable the menu item with a reason, so the user
// learns why before clicking rather than after. The check is also enforced in
// ExecuteSetRowNames, because a rule that lives only in the UI is not a rule --
// any other caller would bypass it.
type RowNameCheck struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

// checkRowNameCandidate applies the requirement that row names identify rows.
//
// Row names are read in one place that matters: they label the points in a
// GoPCA scores plot. Two rows sharing a name are indistinguishable there, and a
// row with no name is unlabelled, so the test is "every entry non-empty and
// distinct" rather than merely "distinct".
//
// Comparison is on the trimmed value, so "P1" and "P1 " collide -- they render
// identically as a label, and calling them distinct would defeat the purpose.
// Case is significant: "P1" and "p1" are different identifiers, which is the
// normal convention for IDs and the safer assumption, since treating them as
// equal would reject data that is in fact well-formed.
func checkRowNameCandidate(values []string) RowNameCheck {
	if len(values) == 0 {
		return RowNameCheck{Reason: "the column has no rows"}
	}

	seen := make(map[string]int, len(values))
	blanks := 0
	duplicates := 0
	var firstDuplicate string

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			blanks++
			continue
		}
		seen[trimmed]++
		if seen[trimmed] == 2 {
			duplicates++
			if firstDuplicate == "" {
				firstDuplicate = trimmed
			}
		}
	}

	switch {
	case blanks > 0 && duplicates > 0:
		return RowNameCheck{Reason: fmt.Sprintf(
			"row names must be unique and complete: %s and %s",
			pluralBlanks(blanks), pluralDuplicates(duplicates, firstDuplicate))}
	case blanks > 0:
		return RowNameCheck{Reason: fmt.Sprintf(
			"row names must be complete: %s", pluralBlanks(blanks))}
	case duplicates > 0:
		return RowNameCheck{Reason: fmt.Sprintf(
			"row names must be unique: %s", pluralDuplicates(duplicates, firstDuplicate))}
	}

	return RowNameCheck{OK: true}
}

func pluralBlanks(n int) string {
	if n == 1 {
		return "1 cell is empty"
	}
	return fmt.Sprintf("%d cells are empty", n)
}

func pluralDuplicates(n int, example string) string {
	if n == 1 {
		return fmt.Sprintf("%q appears more than once", example)
	}
	return fmt.Sprintf("%d values repeat, including %q", n, example)
}

// columnValues returns the cells of one column, padding short rows with "".
//
// Short rows are padded rather than skipped: a missing cell is an empty row
// name, which is exactly the thing checkRowNameCandidate must reject. Skipping
// them would let a ragged column pass by not looking at its gaps.
func columnValues(data *FileData, colIndex int) []string {
	values := make([]string, len(data.Data))
	for i, row := range data.Data {
		if colIndex < len(row) {
			values[i] = row[colIndex]
		}
	}
	return values
}

// CanUseAsRowNames reports whether the given column could become the row-name
// column. Bound for the frontend so the menu can explain itself.
func (a *App) CanUseAsRowNames(data *FileData, colIndex int) RowNameCheck {
	if data == nil || colIndex < 0 || colIndex >= len(data.Headers) {
		return RowNameCheck{Reason: "no such column"}
	}
	return checkRowNameCandidate(columnValues(data, colIndex))
}
