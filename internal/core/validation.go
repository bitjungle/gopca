// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"fmt"
	"math"

	"github.com/bitjungle/gopca/pkg/types"
)

// ValidateDataMatrix validates the basic structure and content of a data matrix
func ValidateDataMatrix(data types.Matrix) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	n := len(data)
	m := len(data[0])

	// Check rectangular matrix
	for i, row := range data {
		if len(row) != m {
			return fmt.Errorf("inconsistent row length at index %d: expected %d, got %d", i, m, len(row))
		}
	}

	// Check dimensions
	if n < 2 {
		return fmt.Errorf("insufficient samples: need at least 2, got %d", n)
	}

	if m < 1 {
		return fmt.Errorf("insufficient features: need at least 1, got %d", m)
	}

	return nil
}

// ValidateNaNValues checks for NaN values in the data matrix
func ValidateNaNValues(data types.Matrix, allowNaN bool) error {
	if allowNaN {
		return nil
	}

	n := len(data)
	m := len(data[0])

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if math.IsNaN(data[i][j]) {
				return fmt.Errorf("NaN value found at row %d, column %d - use missing value handling before PCA", i+1, j+1)
			}
		}
	}

	return nil
}

// ValidateComponentCount validates the number of components requested
func ValidateComponentCount(components, maxComponents int) error {
	if components <= 0 {
		return fmt.Errorf("number of components must be positive, got %d", components)
	}

	if components > maxComponents {
		return fmt.Errorf("too many components requested: maximum %d, got %d", maxComponents, components)
	}

	return nil
}

// CalculateMaxComponents calculates the maximum number of components for a data matrix
func CalculateMaxComponents(rows, cols int) int {
	if cols < rows {
		return cols
	}
	return rows
}

// ValidatePCAInput performs complete validation for PCA input
func ValidatePCAInput(data types.Matrix, config types.PCAConfig) error {
	// Basic matrix validation
	if err := ValidateDataMatrix(data); err != nil {
		return err
	}

	// Check for NaN values (unless using NIPALS with native missing value handling)
	allowNaN := config.Method == "nipals" && config.MissingStrategy == types.MissingNative
	if err := ValidateNaNValues(data, allowNaN); err != nil {
		return err
	}

	// Validate component count
	n := len(data)
	m := len(data[0])
	maxComponents := CalculateMaxComponents(n, m)
	if err := ValidateComponentCount(config.Components, maxComponents); err != nil {
		return err
	}

	return nil
}

// ValidateKernelConfig validates kernel-specific configuration
func ValidateKernelConfig(config types.PCAConfig) error {
	if config.KernelType == "" {
		return fmt.Errorf("kernel type must be specified for kernel PCA")
	}

	switch config.KernelType {
	case "rbf":
		if config.KernelGamma < 0 {
			return fmt.Errorf("gamma must be non-negative for RBF kernel")
		}
	case "poly", "polynomial":
		if config.KernelGamma < 0 {
			return fmt.Errorf("gamma must be non-negative for polynomial kernel")
		}
		if config.KernelDegree < 1 {
			return fmt.Errorf("degree must be at least 1 for polynomial kernel")
		}
	case "linear":
		// No specific validation needed
	default:
		return fmt.Errorf("unsupported kernel type: %s", config.KernelType)
	}

	return nil
}

// ValidateVectorPair validates that two vectors have the same length
func ValidateVectorPair(x, y []float64) error {
	if len(x) != len(y) {
		return fmt.Errorf("vectors must have the same length: got %d and %d", len(x), len(y))
	}
	return nil
}

// CheckForConstantColumns checks if any columns have zero or near-zero variance
func CheckForConstantColumns(data types.Matrix) ([]int, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}

	n := len(data)
	m := len(data[0])
	constantCols := []int{}

	for j := 0; j < m; j++ {
		// Calculate variance for column j
		var sum, sumSq float64
		for i := 0; i < n; i++ {
			val := data[i][j]
			sum += val
			sumSq += val * val
		}
		mean := sum / float64(n)
		variance := sumSq/float64(n) - mean*mean

		// Check if variance is near zero
		if math.Abs(variance) < MinVarianceThreshold {
			constantCols = append(constantCols, j)
		}
	}

	return constantCols, nil
}
