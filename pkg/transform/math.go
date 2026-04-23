// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package transform

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// applyMath applies element-wise mathematical transformations (log, sqrt, square)
// to the specified columns in data. Messages about skipped values and counts are
// appended to result.
func applyMath(data [][]string, columnTypes map[string]string, headers []string, opts Options, result *Result) error {
	for _, colName := range opts.Columns {
		colIndex := findColumn(headers, colName)
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}

		if columnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}

		transformedCount := 0
		for i := range data {
			if colIndex >= len(data[i]) {
				continue
			}

			value := strings.TrimSpace(data[i][colIndex])
			if value == "" {
				continue
			}

			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}

			var transformed float64
			switch opts.Type {
			case Log:
				if num <= 0 {
					result.Messages = append(result.Messages, fmt.Sprintf("Warning: Non-positive value in row %d, column '%s' - cannot apply log", i+1, colName))
					continue
				}
				transformed = math.Log(num)
			case Sqrt:
				if num < 0 {
					result.Messages = append(result.Messages, fmt.Sprintf("Warning: Negative value in row %d, column '%s' - cannot apply sqrt", i+1, colName))
					continue
				}
				transformed = math.Sqrt(num)
			case Square:
				transformed = num * num
			}

			data[i][colIndex] = fmt.Sprintf("%.6g", transformed)
			transformedCount++
		}

		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Transformed %d values in column '%s'", transformedCount, colName))
	}

	return nil
}
