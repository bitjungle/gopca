package main

import (
	"encoding/json"
	"math"
)

// JSONFloat64 is a float64 that marshals NaN and Inf values as null
type JSONFloat64 float64

func (f JSONFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(f))
}

// FileDataJSON is a JSON-safe version of FileData
type FileDataJSON struct {
	Headers            []string            `json:"headers"`
	RowNames           []string            `json:"rowNames"`
	Data               [][]JSONFloat64     `json:"data"`
	MissingMask        [][]bool            `json:"missingMask,omitempty"`
	CategoricalColumns map[string][]string `json:"categoricalColumns,omitempty"`
}

// ToJSONSafe converts FileData to a JSON-safe version
func (fd *FileData) ToJSONSafe() *FileDataJSON {
	if fd == nil {
		return nil
	}

	// Convert float64 data to JSONFloat64 and build missing mask
	jsonData := make([][]JSONFloat64, len(fd.Data))
	missingMask := make([][]bool, len(fd.Data))
	hasMissing := false

	for i, row := range fd.Data {
		jsonData[i] = make([]JSONFloat64, len(row))
		missingMask[i] = make([]bool, len(row))
		for j, val := range row {
			jsonData[i][j] = JSONFloat64(val)
			if math.IsNaN(val) {
				missingMask[i][j] = true
				hasMissing = true
			}
		}
	}

	result := &FileDataJSON{
		Headers:            fd.Headers,
		RowNames:           fd.RowNames,
		Data:               jsonData,
		CategoricalColumns: fd.CategoricalColumns,
	}

	// Only include missing mask if there are missing values
	if hasMissing {
		result.MissingMask = missingMask
	}

	return result
}
