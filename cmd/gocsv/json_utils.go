package main

import (
	"encoding/json"
	"math"
)

// JSONFloat64 is a float64 that marshals NaN and Inf values as null
type JSONFloat64 float64

// MarshalJSON implements the json.Marshaler interface
func (f JSONFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(f))
}

// FileData represents the structure of loaded file data
// This version uses JSONFloat64 to handle NaN values safely
type FileData struct {
	Headers              []string                   `json:"headers"`
	RowNames             []string                   `json:"rowNames,omitempty"`
	Data                 [][]string                 `json:"data"`
	Rows                 int                        `json:"rows"`
	Columns              int                        `json:"columns"`
	CategoricalColumns   map[string][]string        `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]JSONFloat64   `json:"numericTargetColumns,omitempty"`
	ColumnTypes          map[string]string          `json:"columnTypes,omitempty"`
}

// ConvertFloat64MapToJSON converts a map of float64 slices to JSONFloat64 slices
func ConvertFloat64MapToJSON(data map[string][]float64) map[string][]JSONFloat64 {
	if data == nil {
		return nil
	}
	
	result := make(map[string][]JSONFloat64, len(data))
	for key, values := range data {
		jsonValues := make([]JSONFloat64, len(values))
		for i, v := range values {
			jsonValues[i] = JSONFloat64(v)
		}
		result[key] = jsonValues
	}
	return result
}