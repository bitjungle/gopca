// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"strings"
)

// DefaultMissingValues returns the default set of missing value indicators.
// These are commonly used representations of missing data across different
// data sources and statistical software packages.
func DefaultMissingValues() []string {
	return []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL", "m"}
}

// IsMissingValue checks if a string value represents missing data.
// It performs a case-insensitive comparison against the provided list
// of missing value indicators.
//
// Parameters:
//   - value: The string value to check
//   - missingIndicators: List of strings that represent missing values
//
// Returns true if the value (after trimming whitespace) matches any
// of the missing indicators (case-insensitive).
func IsMissingValue(value string, missingIndicators []string) bool {
	trimmedValue := strings.TrimSpace(value)

	// Convert to lowercase for case-insensitive comparison
	lowerValue := strings.ToLower(trimmedValue)

	// Check against each missing value indicator
	for _, indicator := range missingIndicators {
		// Trim the indicator as well for consistent comparison
		trimmedIndicator := strings.TrimSpace(indicator)
		if lowerValue == strings.ToLower(trimmedIndicator) {
			return true
		}
	}

	return false
}

// ContainsMissingValues checks if a slice of strings contains any missing values.
// Useful for quick validation of data rows or columns.
func ContainsMissingValues(values []string, missingIndicators []string) bool {
	for _, value := range values {
		if IsMissingValue(value, missingIndicators) {
			return true
		}
	}
	return false
}

// CountMissingValues returns the number of missing values in a slice of strings.
func CountMissingValues(values []string, missingIndicators []string) int {
	count := 0
	for _, value := range values {
		if IsMissingValue(value, missingIndicators) {
			count++
		}
	}
	return count
}
