// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"fmt"
	"strings"
)

// Apply executes the transformation described by opts against a deep copy of
// the input data and returns the modified data as a Result. The original Input
// is never modified.
//
// Unsupported transformation types return an error.
func Apply(in Input, opts Options) (*Result, error) {
	if len(in.Data) == 0 {
		return nil, fmt.Errorf("no data to transform")
	}

	// Deep-copy input data so the original is never mutated.
	headers := make([]string, len(in.Headers))
	copy(headers, in.Headers)

	data := make([][]string, len(in.Data))
	for i := range in.Data {
		data[i] = make([]string, len(in.Data[i]))
		copy(data[i], in.Data[i])
	}

	columnTypes := make(map[string]string, len(in.ColumnTypes))
	for k, v := range in.ColumnTypes {
		columnTypes[k] = v
	}

	catCols := make(map[string][]string, len(in.CategoricalColumns))
	for k, v := range in.CategoricalColumns {
		values := make([]string, len(v))
		copy(values, v)
		catCols[k] = values
	}

	result := &Result{
		TransformedColumns: []string{},
		NewColumns:         []string{},
		Messages:           []string{},
	}

	switch opts.Type {
	case Log, Sqrt, Square:
		if err := applyMath(data, columnTypes, headers, opts, result); err != nil {
			return nil, err
		}
	case Standardize:
		if err := applyStandardize(data, columnTypes, headers, opts, result); err != nil {
			return nil, err
		}
	case MinMax:
		if err := applyMinMax(data, columnTypes, headers, opts, result); err != nil {
			return nil, err
		}
	case Bin:
		if err := applyBin(data, columnTypes, catCols, headers, opts, result); err != nil {
			return nil, err
		}
	case OneHot:
		if err := applyOneHot(data, columnTypes, catCols, &headers, opts, result); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported transformation type: %s", opts.Type)
	}

	result.Headers = headers
	result.Data = data
	result.ColumnTypes = columnTypes
	result.CategoricalColumns = catCols
	result.Columns = len(headers)

	return result, nil
}

// GetTransformableColumns returns the column names from in that are eligible
// for the given transformation type.
//
// Mathematical and scaling transforms (Log, Sqrt, Square, Standardize, MinMax,
// Bin) require numeric columns. Columns with the "#target" suffix are excluded.
// OneHot requires categorical columns.
func GetTransformableColumns(in Input, transformType Type) []string {
	columns := []string{}

	for _, header := range in.Headers {
		colType := in.ColumnTypes[header]

		switch transformType {
		case Log, Sqrt, Square, Standardize, MinMax, Bin:
			if colType == "numeric" && !strings.HasSuffix(header, "#target") {
				columns = append(columns, header)
			}
		case OneHot:
			if colType == "categorical" {
				columns = append(columns, header)
			}
		}
	}

	return columns
}

// findColumn returns the index of colName in headers, or -1 if not found.
func findColumn(headers []string, colName string) int {
	for i, h := range headers {
		if h == colName {
			return i
		}
	}
	return -1
}
