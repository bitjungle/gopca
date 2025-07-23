package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseRanges parses a comma-separated string of indices and ranges into a slice of integers.
// Examples:
//   - "1,3,5" returns [1, 3, 5]
//   - "1-3,5" returns [1, 2, 3, 5]
//   - "1,3-5,7" returns [1, 3, 4, 5, 7]
//
// Note: Input indices are 1-based (human-friendly), output indices are 0-based
func ParseRanges(input string) ([]int, error) {
	if input == "" {
		return []int{}, nil
	}

	// Use a map to avoid duplicates
	indexMap := make(map[int]bool)

	// Split by comma
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a range (contains hyphen)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start index in range %s: %v", part, err)
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end index in range %s: %v", part, err)
			}

			if start < 1 || end < 1 {
				return nil, fmt.Errorf("indices must be positive (1-based), got range %d-%d", start, end)
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}

			// Add all indices in the range (convert to 0-based)
			for i := start; i <= end; i++ {
				indexMap[i-1] = true
			}
		} else {
			// Single index
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid index %s: %v", part, err)
			}

			if index < 1 {
				return nil, fmt.Errorf("indices must be positive (1-based), got %d", index)
			}

			// Convert to 0-based
			indexMap[index-1] = true
		}
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(indexMap))
	for index := range indexMap {
		result = append(result, index)
	}
	sort.Ints(result)

	return result, nil
}

// FilterMatrix removes specified rows and columns from a matrix
func FilterMatrix(data [][]float64, excludedRows, excludedColumns []int) ([][]float64, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Create maps for O(1) lookup
	excludeRowMap := make(map[int]bool)
	for _, idx := range excludedRows {
		if idx < 0 || idx >= len(data) {
			return nil, fmt.Errorf("row index %d out of bounds (0-%d)", idx, len(data)-1)
		}
		excludeRowMap[idx] = true
	}

	excludeColMap := make(map[int]bool)
	numCols := len(data[0])
	for _, idx := range excludedColumns {
		if idx < 0 || idx >= numCols {
			return nil, fmt.Errorf("column index %d out of bounds (0-%d)", idx, numCols-1)
		}
		excludeColMap[idx] = true
	}

	// Build filtered matrix
	var filtered [][]float64
	for i, row := range data {
		if excludeRowMap[i] {
			continue
		}

		var filteredRow []float64
		for j, val := range row {
			if !excludeColMap[j] {
				filteredRow = append(filteredRow, val)
			}
		}

		if len(filteredRow) > 0 {
			filtered = append(filtered, filteredRow)
		}
	}

	return filtered, nil
}

// FilterStringSlice removes elements at specified indices from a string slice
func FilterStringSlice(items []string, excludedIndices []int) ([]string, error) {
	// Create map for O(1) lookup
	excludeMap := make(map[int]bool)
	for _, idx := range excludedIndices {
		if idx < 0 || idx >= len(items) {
			return nil, fmt.Errorf("index %d out of bounds (0-%d)", idx, len(items)-1)
		}
		excludeMap[idx] = true
	}

	var filtered []string
	for i, item := range items {
		if !excludeMap[i] {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}
