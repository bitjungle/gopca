// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseNumericValue attempts to parse a string as a float64.
// It handles different decimal separators and special float values.
//
// Parameters:
//   - value: The string value to parse
//   - decimalSeparator: The decimal separator character ('.' or ',')
//
// Returns the parsed float64 and an error if parsing fails.
// Special cases handled:
//   - "inf", "+inf", "infinity" -> +Inf
//   - "-inf", "-infinity" -> -Inf
//   - Empty string returns an error
func ParseNumericValue(value string, decimalSeparator rune) (float64, error) {
	trimmedValue := strings.TrimSpace(value)

	if trimmedValue == "" {
		return 0, fmt.Errorf("cannot parse empty string as number")
	}

	// Handle decimal separator conversion
	parseValue := trimmedValue
	if decimalSeparator == ',' {
		// Replace comma with dot for standard parsing
		parseValue = strings.ReplaceAll(trimmedValue, ",", ".")
	}

	// Try standard float parsing first
	if val, err := strconv.ParseFloat(parseValue, 64); err == nil {
		return val, nil
	}

	// Check for special float values (case-insensitive)
	lowerValue := strings.ToLower(trimmedValue)
	switch lowerValue {
	case "inf", "+inf", "infinity", "+infinity":
		return math.Inf(1), nil
	case "-inf", "-infinity":
		return math.Inf(-1), nil
	case "nan":
		return math.NaN(), nil
	}

	return 0, fmt.Errorf("cannot parse '%s' as number", trimmedValue)
}

// ParseNumericValueWithMissing combines numeric parsing with missing value detection.
// If the value is identified as missing, it returns NaN without an error.
//
// Parameters:
//   - value: The string value to parse
//   - decimalSeparator: The decimal separator character
//   - missingIndicators: List of strings that represent missing values
//
// Returns:
//   - float64: The parsed value (NaN if missing)
//   - bool: true if the value was identified as missing
//   - error: parsing error if the value is not missing and cannot be parsed
func ParseNumericValueWithMissing(value string, decimalSeparator rune, missingIndicators []string) (float64, bool, error) {
	// Check if it's a missing value first
	if IsMissingValue(value, missingIndicators) {
		return math.NaN(), true, nil
	}

	// Try to parse as numeric
	val, err := ParseNumericValue(value, decimalSeparator)
	if err != nil {
		return 0, false, err
	}

	return val, false, nil
}

// IsNumericString checks if a string can be parsed as a number.
// This is useful for determining column types in CSV files.
func IsNumericString(value string, decimalSeparator rune) bool {
	_, err := ParseNumericValue(value, decimalSeparator)
	return err == nil
}

// ParseFloatSlice parses a slice of strings into a slice of float64 values.
// Missing values are represented as NaN in the output.
func ParseFloatSlice(values []string, decimalSeparator rune, missingIndicators []string) ([]float64, error) {
	result := make([]float64, len(values))

	for i, value := range values {
		val, _, err := ParseNumericValueWithMissing(value, decimalSeparator, missingIndicators)
		if err != nil {
			return nil, fmt.Errorf("error parsing value at index %d: %w", i, err)
		}
		result[i] = val
	}

	return result, nil
}
