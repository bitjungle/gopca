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

package dataquality

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/utils"
)

// missingIndicators is the set of strings treated as missing values by GoCSV.
// It extends the shared default set with additional common indicators.
var missingIndicators = append(utils.DefaultMissingValues(), "-", "?", "none", "missing")

// isMissing reports whether value should be treated as a missing value.
func isMissing(value string) bool {
	return utils.IsMissingValue(value, missingIndicators)
}

// AnalyzeMissing returns missing-value statistics for the given data matrix
// and column headers. An empty or nil data slice returns a zeroed stats struct.
func AnalyzeMissing(data [][]string, headers []string) *MissingValueStats {
	rows := len(data)
	cols := len(headers)

	stats := &MissingValueStats{
		TotalCells:  rows * cols,
		ColumnStats: make(map[string]*ColumnMissing, cols),
		RowStats:    make(map[int]*RowMissing),
	}

	if rows == 0 || cols == 0 {
		return stats
	}

	// Analyse by column
	for colIdx, header := range headers {
		colStats := &ColumnMissing{
			Name:        header,
			TotalValues: rows,
		}

		missingIndices := make([]int, 0)
		for rowIdx := 0; rowIdx < rows; rowIdx++ {
			if rowIdx >= len(data) || colIdx >= len(data[rowIdx]) {
				continue
			}
			if isMissing(strings.TrimSpace(data[rowIdx][colIdx])) {
				colStats.MissingValues++
				stats.MissingCells++
				missingIndices = append(missingIndices, rowIdx)
			}
		}

		if colStats.TotalValues > 0 {
			colStats.MissingPercent = float64(colStats.MissingValues) / float64(colStats.TotalValues) * 100
		}
		colStats.Pattern = detectMissingPattern(missingIndices, rows)
		stats.ColumnStats[header] = colStats
	}

	// Analyse by row
	for rowIdx := 0; rowIdx < rows; rowIdx++ {
		rowStats := &RowMissing{
			Index:       rowIdx,
			TotalValues: cols,
		}

		if rowIdx < len(data) {
			for colIdx := 0; colIdx < cols && colIdx < len(data[rowIdx]); colIdx++ {
				if isMissing(strings.TrimSpace(data[rowIdx][colIdx])) {
					rowStats.MissingValues++
				}
			}
		}

		if rowStats.TotalValues > 0 {
			rowStats.MissingPercent = float64(rowStats.MissingValues) / float64(rowStats.TotalValues) * 100
		}
		if rowStats.MissingValues > 0 {
			stats.RowStats[rowIdx] = rowStats
		}
	}

	if stats.TotalCells > 0 {
		stats.MissingPercent = float64(stats.MissingCells) / float64(stats.TotalCells) * 100
	}

	return stats
}

// detectMissingPattern classifies the spatial pattern of missing value indices
// within a column.
func detectMissingPattern(missingIndices []int, totalRows int) string {
	if len(missingIndices) == 0 {
		return "none"
	}

	// All missing at top?
	allTop := true
	for i, idx := range missingIndices {
		if idx != i {
			allTop = false
			break
		}
	}
	if allTop {
		return "top"
	}

	// All missing at bottom?
	allBottom := true
	start := totalRows - len(missingIndices)
	for i, idx := range missingIndices {
		if idx != start+i {
			allBottom = false
			break
		}
	}
	if allBottom {
		return "bottom"
	}

	// Regular interval (systematic)?
	if len(missingIndices) > 2 {
		intervals := make([]int, 0, len(missingIndices)-1)
		for i := 1; i < len(missingIndices); i++ {
			intervals = append(intervals, missingIndices[i]-missingIndices[i-1])
		}
		systematic := true
		first := intervals[0]
		for _, iv := range intervals[1:] {
			if iv != first {
				systematic = false
				break
			}
		}
		if systematic && first > 1 {
			return "systematic"
		}
	}

	return "random"
}

// Fill applies a fill strategy to a deep copy of data and returns the new
// matrix. The original data is never modified.
//
// strategy values: "mean", "median", "mode", "forward", "backward", "custom".
// If req.Column is empty, all columns are processed; otherwise only the named
// column is filled.
func Fill(data [][]string, headers []string, columnTypes map[string]string, req FillRequest) ([][]string, error) {
	rows := len(data)
	if rows == 0 {
		return nil, fmt.Errorf("no data to process")
	}

	// Deep copy
	result := make([][]string, rows)
	for i, row := range data {
		result[i] = make([]string, len(row))
		copy(result[i], row)
	}

	// Determine columns to process
	colIndices := make([]int, 0, len(headers))
	if req.Column == "" {
		for i := range headers {
			colIndices = append(colIndices, i)
		}
	} else {
		found := false
		for i, h := range headers {
			if h == req.Column {
				colIndices = append(colIndices, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %q not found", req.Column)
		}
	}

	for _, colIdx := range colIndices {
		switch req.Strategy {
		case "mean":
			fillWithMean(result, headers, columnTypes, colIdx)
		case "median":
			fillWithMedian(result, headers, columnTypes, colIdx)
		case "mode":
			fillWithMode(result, colIdx)
		case "forward":
			fillForward(result, colIdx)
		case "backward":
			fillBackward(result, colIdx)
		case "custom":
			fillWithCustomValue(result, colIdx, req.Value)
		default:
			return nil, fmt.Errorf("unknown fill strategy: %q", req.Strategy)
		}
	}

	return result, nil
}

// ─── fill strategy implementations ──────────────────────────────────────────

func fillWithMean(data [][]string, headers []string, columnTypes map[string]string, colIdx int) {
	if colIdx >= len(headers) {
		return
	}
	colType := "numeric"
	if columnTypes != nil {
		if t, ok := columnTypes[headers[colIdx]]; ok {
			colType = t
		}
	}
	if colType != "numeric" {
		fillWithMode(data, colIdx)
		return
	}
	values := getNumericColumn(data, colIdx)
	if len(values) == 0 {
		return
	}
	mean := calculateMean(values)
	fillValue := strconv.FormatFloat(mean, 'f', -1, 64)
	replaceMissing(data, colIdx, fillValue)
}

func fillWithMedian(data [][]string, headers []string, columnTypes map[string]string, colIdx int) {
	if colIdx >= len(headers) {
		return
	}
	colType := "numeric"
	if columnTypes != nil {
		if t, ok := columnTypes[headers[colIdx]]; ok {
			colType = t
		}
	}
	if colType != "numeric" {
		fillWithMode(data, colIdx)
		return
	}
	values := getNumericColumn(data, colIdx)
	if len(values) == 0 {
		return
	}
	sort.Float64s(values)
	median := calculateMedian(values)
	fillValue := strconv.FormatFloat(median, 'f', -1, 64)
	replaceMissing(data, colIdx, fillValue)
}

func fillWithMode(data [][]string, colIdx int) {
	counts := make(map[string]int)
	for _, row := range data {
		if colIdx < len(row) {
			v := strings.TrimSpace(row[colIdx])
			if !isMissing(v) {
				counts[v]++
			}
		}
	}
	if len(counts) == 0 {
		return
	}
	mode := ""
	max := 0
	for v, c := range counts {
		if c > max {
			max = c
			mode = v
		}
	}
	replaceMissing(data, colIdx, mode)
}

func fillForward(data [][]string, colIdx int) {
	last := ""
	for _, row := range data {
		if colIdx < len(row) {
			v := strings.TrimSpace(row[colIdx])
			if isMissing(v) {
				if last != "" {
					row[colIdx] = last
				}
			} else {
				last = v
			}
		}
	}
}

func fillBackward(data [][]string, colIdx int) {
	last := ""
	for i := len(data) - 1; i >= 0; i-- {
		row := data[i]
		if colIdx < len(row) {
			v := strings.TrimSpace(row[colIdx])
			if isMissing(v) {
				if last != "" {
					row[colIdx] = last
				}
			} else {
				last = v
			}
		}
	}
}

func fillWithCustomValue(data [][]string, colIdx int, customValue string) {
	replaceMissing(data, colIdx, customValue)
}

// ─── internal helpers ────────────────────────────────────────────────────────

// replaceMissing overwrites all missing cells in colIdx with fillValue.
func replaceMissing(data [][]string, colIdx int, fillValue string) {
	for _, row := range data {
		if colIdx < len(row) && isMissing(strings.TrimSpace(row[colIdx])) {
			row[colIdx] = fillValue
		}
	}
}

// getNumericColumn returns all non-missing numeric values from colIdx.
func getNumericColumn(data [][]string, colIdx int) []float64 {
	values := make([]float64, 0, len(data))
	for _, row := range data {
		if colIdx >= len(row) {
			continue
		}
		v := strings.TrimSpace(row[colIdx])
		if isMissing(v) {
			continue
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			values = append(values, f)
		}
	}
	return values
}
