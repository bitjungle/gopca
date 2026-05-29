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
