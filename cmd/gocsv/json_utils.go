package main

import (
	"encoding/json"
	"math"
)

// JSONFloat64 is a custom type that handles NaN and Inf values for JSON serialization
type JSONFloat64 float64

// MarshalJSON implements the json.Marshaler interface
func (f JSONFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return json.Marshal(nil) // Convert NaN/Inf to null
	}
	return json.Marshal(float64(f))
}

// ConvertFloat64SliceToJSON converts a slice of float64 to a slice of JSONFloat64
func ConvertFloat64SliceToJSON(values []float64) []JSONFloat64 {
	result := make([]JSONFloat64, len(values))
	for i, v := range values {
		result[i] = JSONFloat64(v)
	}
	return result
}

// ConvertNumericTargetColumns converts numeric target columns to JSON-safe format
func ConvertNumericTargetColumns(columns map[string][]float64) map[string][]JSONFloat64 {
	if columns == nil {
		return nil
	}
	
	result := make(map[string][]JSONFloat64, len(columns))
	for key, values := range columns {
		result[key] = ConvertFloat64SliceToJSON(values)
	}
	return result
}

// FileDataJSON is a JSON-safe version of FileData
type FileDataJSON struct {
	Headers              []string                   `json:"headers"`
	RowNames             []string                   `json:"rowNames,omitempty"`
	Data                 [][]string                 `json:"data"`
	Rows                 int                        `json:"rows"`
	Columns              int                        `json:"columns"`
	CategoricalColumns   map[string][]string        `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]JSONFloat64   `json:"numericTargetColumns,omitempty"`
	ColumnTypes          map[string]string          `json:"columnTypes,omitempty"`
}

// ToJSON converts FileData to a JSON-safe version
func (f *FileData) ToJSON() *FileDataJSON {
	return &FileDataJSON{
		Headers:              f.Headers,
		RowNames:             f.RowNames,
		Data:                 f.Data,
		Rows:                 f.Rows,
		Columns:              f.Columns,
		CategoricalColumns:   f.CategoricalColumns,
		NumericTargetColumns: ConvertNumericTargetColumns(f.NumericTargetColumns),
		ColumnTypes:          f.ColumnTypes,
	}
}