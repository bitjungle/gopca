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

package transform

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// applyBin discretizes the specified numeric columns into equal-width bins.
// Each value is replaced with a label "Bin_N" (1-indexed). The column type is
// updated to "categorical" and the binned values are registered in catCols.
func applyBin(data [][]string, columnTypes map[string]string, catCols map[string][]string, headers []string, opts Options, result *Result) error {
	binCount := opts.BinCount
	if binCount <= 0 {
		binCount = 5 // default number of bins
	}

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

		// Collect numeric values and their row positions.
		values := make([]float64, 0)
		indices := make([]int, 0)
		minVal := math.Inf(1)
		maxVal := math.Inf(-1)

		for i := range data {
			if colIndex >= len(data[i]) {
				continue
			}
			v := strings.TrimSpace(data[i][colIndex])
			if v == "" {
				continue
			}
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				continue
			}
			values = append(values, num)
			indices = append(indices, i)
			if num < minVal {
				minVal = num
			}
			if num > maxVal {
				maxVal = num
			}
		}

		if len(values) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no numeric values", colName))
			continue
		}

		if maxVal <= minVal {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has constant values, cannot bin", colName))
			continue
		}

		binWidth := (maxVal - minVal) / float64(binCount)

		for i, idx := range indices {
			binIndex := int((values[i] - minVal) / binWidth)
			if binIndex >= binCount {
				binIndex = binCount - 1
			}
			data[idx][colIndex] = fmt.Sprintf("Bin_%d", binIndex+1)
		}

		// Update metadata.
		columnTypes[colName] = "categorical"

		catValues := make([]string, len(data))
		for i := range data {
			if colIndex < len(data[i]) {
				catValues[i] = data[i][colIndex]
			}
		}
		catCols[colName] = catValues

		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Binned %d values in column '%s' into %d bins", len(values), colName, binCount))
	}

	return nil
}
