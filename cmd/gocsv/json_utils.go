package main

import (
	"github.com/bitjungle/gopca/pkg/types"
)

// FileData represents the structure of loaded file data
// This version uses JSONFloat64 to handle NaN values safely
type FileData struct {
	Headers              []string                   `json:"headers"`
	RowNames             []string                   `json:"rowNames,omitempty"`
	Data                 [][]string                 `json:"data"`
	Rows                 int                        `json:"rows"`
	Columns              int                        `json:"columns"`
	CategoricalColumns   map[string][]string        `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]types.JSONFloat64   `json:"numericTargetColumns,omitempty"`
	ColumnTypes          map[string]string          `json:"columnTypes,omitempty"`
}

// ConvertFloat64MapToJSON converts a map of float64 slices to JSONFloat64 slices
func ConvertFloat64MapToJSON(data map[string][]float64) map[string][]types.JSONFloat64 {
	if data == nil {
		return nil
	}
	
	result := make(map[string][]types.JSONFloat64, len(data))
	for key, values := range data {
		jsonValues := make([]types.JSONFloat64, len(values))
		for i, v := range values {
			jsonValues[i] = types.JSONFloat64(v)
		}
		result[key] = jsonValues
	}
	return result
}