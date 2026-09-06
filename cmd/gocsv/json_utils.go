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
	"github.com/bitjungle/gopca/pkg/types"
)

// FileData represents the structure of loaded file data
// This version uses JSONFloat64 to handle NaN values safely
type FileData struct {
	Headers  []string `json:"headers"`
	RowNames []string `json:"rowNames,omitempty"`
	// RowNamesHeader is the header of the column the row names came from.
	// Carried so the file writes back with that column named as it was read,
	// and so promoting a column to row names can be undone (#859).
	RowNamesHeader       string                         `json:"rowNamesHeader,omitempty"`
	Data                 [][]string                     `json:"data"`
	Rows                 int                            `json:"rows"`
	Columns              int                            `json:"columns"`
	CategoricalColumns   map[string][]string            `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]types.JSONFloat64 `json:"numericTargetColumns,omitempty"`
	ColumnTypes          map[string]string              `json:"columnTypes,omitempty"`
}

// ConvertFloat64MapToJSON converts a map of float64 slices to JSONFloat64 slices
// This is now a wrapper around the shared function in pkg/types
func ConvertFloat64MapToJSON(data map[string][]float64) map[string][]types.JSONFloat64 {
	return types.ConvertFloat64MapToJSON(data)
}
