// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"github.com/bitjungle/gopca/pkg/types"
)

// FileData represents the structure of loaded file data
// This version uses JSONFloat64 to handle NaN values safely
type FileData struct {
	Headers              []string                       `json:"headers"`
	RowNames             []string                       `json:"rowNames,omitempty"`
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
