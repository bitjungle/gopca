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
