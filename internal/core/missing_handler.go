package core

import (
	"fmt"
	"math"
	"sort"

	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/stat"
)

// MissingValueHandler handles missing values in data matrices
type MissingValueHandler struct {
	strategy types.MissingValueStrategy
}

// NewMissingValueHandler creates a new missing value handler
func NewMissingValueHandler(strategy types.MissingValueStrategy) *MissingValueHandler {
	return &MissingValueHandler{strategy: strategy}
}

// HandleMissingValues processes missing values according to the specified strategy
// It only considers missing values in the selected columns
func (h *MissingValueHandler) HandleMissingValues(data types.Matrix, missingInfo *types.MissingValueInfo, selectedCols []int) (types.Matrix, error) {
	if !missingInfo.HasMissing() {
		// No missing values, return data as-is
		return data, nil
	}

	switch h.strategy {
	case types.MissingDrop:
		return h.dropRows(data, missingInfo.RowsAffected)

	case types.MissingMean:
		return h.imputeWithMean(data, missingInfo, selectedCols)

	case types.MissingMedian:
		return h.imputeWithMedian(data, missingInfo, selectedCols)

	default:
		return nil, fmt.Errorf("unsupported missing value strategy: %s", h.strategy)
	}
}

// dropRows removes rows that contain missing values in selected columns
func (h *MissingValueHandler) dropRows(data types.Matrix, rowsToRemove []int) (types.Matrix, error) {
	if len(rowsToRemove) == 0 {
		return data, nil
	}

	// Create a set for fast lookup
	removeSet := make(map[int]bool)
	for _, row := range rowsToRemove {
		removeSet[row] = true
	}

	// Create new matrix without the rows containing missing values
	cleanData := make(types.Matrix, 0, len(data)-len(rowsToRemove))
	for i, row := range data {
		if !removeSet[i] {
			// Copy the row
			newRow := make([]float64, len(row))
			copy(newRow, row)
			cleanData = append(cleanData, newRow)
		}
	}

	if len(cleanData) == 0 {
		return nil, fmt.Errorf("all rows contain missing values in selected columns")
	}

	return cleanData, nil
}

// imputeWithMean replaces missing values with column means
func (h *MissingValueHandler) imputeWithMean(data types.Matrix, missingInfo *types.MissingValueInfo, selectedCols []int) (types.Matrix, error) {
	// Calculate means for columns with missing values
	colMeans := h.calculateColumnStatistics(data, missingInfo.ColumnIndices, true)

	// Create a copy of the data
	imputedData := make(types.Matrix, len(data))
	for i := range data {
		imputedData[i] = make([]float64, len(data[i]))
		copy(imputedData[i], data[i])
	}

	// Impute missing values
	for _, col := range missingInfo.ColumnIndices {
		mean := colMeans[col]
		for row := 0; row < len(data); row++ {
			if math.IsNaN(imputedData[row][col]) {
				imputedData[row][col] = mean
			}
		}
	}

	return imputedData, nil
}

// imputeWithMedian replaces missing values with column medians
func (h *MissingValueHandler) imputeWithMedian(data types.Matrix, missingInfo *types.MissingValueInfo, selectedCols []int) (types.Matrix, error) {
	// Calculate medians for columns with missing values
	colMedians := h.calculateColumnStatistics(data, missingInfo.ColumnIndices, false)

	// Create a copy of the data
	imputedData := make(types.Matrix, len(data))
	for i := range data {
		imputedData[i] = make([]float64, len(data[i]))
		copy(imputedData[i], data[i])
	}

	// Impute missing values
	for _, col := range missingInfo.ColumnIndices {
		median, exists := colMedians[col]
		if !exists {
			return nil, fmt.Errorf("no median calculated for column %d", col)
		}
		for row := 0; row < len(data); row++ {
			if math.IsNaN(imputedData[row][col]) {
				imputedData[row][col] = median
			}
		}
	}

	return imputedData, nil
}

// calculateColumnStatistics calculates mean or median for specified columns
func (h *MissingValueHandler) calculateColumnStatistics(data types.Matrix, columns []int, calculateMean bool) map[int]float64 {
	stats := make(map[int]float64)

	for _, col := range columns {
		// Collect non-missing values
		validValues := []float64{}
		for row := 0; row < len(data); row++ {
			if !math.IsNaN(data[row][col]) {
				validValues = append(validValues, data[row][col])
			}
		}

		if len(validValues) == 0 {
			// Column has all missing values, use 0
			stats[col] = 0.0
			continue
		}

		if calculateMean {
			stats[col] = stat.Mean(validValues, nil)
		} else {
			// Calculate median manually for better control
			sort.Float64s(validValues)
			n := len(validValues)
			if n%2 == 0 {
				// Even number of values, take average of middle two
				stats[col] = (validValues[n/2-1] + validValues[n/2]) / 2.0
			} else {
				// Odd number of values, take middle value
				stats[col] = validValues[n/2]
			}
		}
	}

	return stats
}

// ValidateDataForPCA checks if data is suitable for PCA after handling missing values
func ValidateDataForPCA(data types.Matrix, selectedCols []int) error {
	if len(data) == 0 {
		return fmt.Errorf("no data rows available after missing value handling")
	}

	if len(data) < 2 {
		return fmt.Errorf("insufficient samples for PCA: need at least 2, got %d", len(data))
	}

	// Check each selected column for remaining NaN values
	for _, col := range selectedCols {
		hasValidValue := false
		allNaN := true

		for row := 0; row < len(data); row++ {
			if !math.IsNaN(data[row][col]) {
				hasValidValue = true
				allNaN = false
			}
		}

		if allNaN {
			return fmt.Errorf("column %d contains only missing values", col)
		}

		if !hasValidValue {
			return fmt.Errorf("column %d has no valid values after missing value handling", col)
		}
	}

	return nil
}
