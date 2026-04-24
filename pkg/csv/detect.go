// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"strconv"
	"strings"
)

// DetectColumnType inspects the values of a single column in a row-major data
// matrix and returns one of the following type strings:
//
//   - "numeric"     — >90% of non-empty values parse as float64
//   - "categorical" — unique-value ratio <20% of total, or fewer than 20 unique values
//   - "text"        — everything else
//   - "empty"       — no non-empty values found
//   - "unknown"     — data is nil/empty or colIndex is negative
//
// The function never modifies data.
func DetectColumnType(data [][]string, colIndex int) string {
	if len(data) == 0 || colIndex < 0 {
		return "unknown"
	}

	numericCount := 0
	totalCount := 0
	uniqueValues := make(map[string]bool)

	for _, row := range data {
		if colIndex >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[colIndex])
		if value == "" {
			continue
		}
		totalCount++
		uniqueValues[value] = true
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			numericCount++
		}
	}

	if totalCount == 0 {
		return "empty"
	}

	// More than 90% numeric → numeric column.
	if float64(numericCount)/float64(totalCount) > 0.9 {
		return "numeric"
	}

	// Low cardinality → categorical.
	if float64(len(uniqueValues))/float64(totalCount) < 0.2 || len(uniqueValues) < 20 {
		return "categorical"
	}

	return "text"
}
