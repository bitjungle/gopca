// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"fmt"
	"strings"
)

// ValidationInput carries the dataset fields needed to validate compatibility
// with GoPCA. It mirrors the relevant fields of the application FileData type
// so that the function can be used without a dependency on the Wails layer.
type ValidationInput struct {
	// Headers is the ordered list of column names.
	Headers []string
	// Data is the row-major data matrix (each row is a slice of string values).
	Data [][]string
	// ColumnTypes maps each column name to its detected type: "numeric",
	// "categorical", or "target".
	ColumnTypes map[string]string
	// RowNames holds the per-row label values when a row-name column was
	// detected during import (may be nil).
	RowNames []string
	// Rows is the number of data rows.
	Rows int
	// Columns is the number of columns (= len(Headers)).
	Columns int
}

// ValidationResult reports whether a dataset is compatible with GoPCA and
// provides a list of messages categorised as ERROR, WARNING, or INFO.
type ValidationResult struct {
	// IsValid is true when no ERROR-level messages were generated.
	IsValid bool
	// Messages contains zero or more messages prefixed with "ERROR:",
	// "WARNING:", or "INFO:".
	Messages []string
}

// missingValueTokens is the set of string representations treated as missing
// data during GoPCA compatibility validation.
var missingValueTokens = map[string]bool{
	"":     true,
	"NA":   true,
	"N/A":  true,
	"nan":  true,
	"NaN":  true,
	"null": true,
	"NULL": true,
}

// ValidateForGoPCA inspects in and returns a ValidationResult describing
// whether the dataset can be used for PCA analysis in GoPCA Desktop.
//
// Checks performed:
//   - Minimum row count (≥2)
//   - Minimum numeric column count (≥2 required, ≥3 recommended)
//   - Per-column missing value percentage (warns at >50%)
//   - Overall missing value percentage
//   - Dataset size (info at >10 000 rows)
//   - Row name presence (info)
//   - Categorical and target column counts (info)
func ValidateForGoPCA(in ValidationInput) *ValidationResult {
	var messages []string
	var numericColumns, categoricalColumns, targetColumns, totalMissing int

	// Minimum row count.
	if in.Rows < 2 {
		messages = append(messages, fmt.Sprintf("ERROR: Data must have at least 2 rows (found %d)", in.Rows))
	}

	// Count column types.
	for _, colType := range in.ColumnTypes {
		switch colType {
		case "numeric":
			numericColumns++
		case "categorical":
			categoricalColumns++
		case "target":
			targetColumns++
		}
	}

	// Per-column missing value check.
	for colIdx, header := range in.Headers {
		missingInCol := 0
		for i := 0; i < in.Rows; i++ {
			if i >= len(in.Data) {
				break
			}
			if colIdx >= len(in.Data[i]) {
				continue
			}
			trimmed := strings.TrimSpace(in.Data[i][colIdx])
			if missingValueTokens[trimmed] {
				missingInCol++
				totalMissing++
			}
		}
		if in.Rows > 0 {
			pct := float64(missingInCol) / float64(in.Rows) * 100
			if pct > 50 {
				messages = append(messages, fmt.Sprintf("WARNING: Column '%s' has %.1f%% missing values", header, pct))
			}
		}
	}

	// Categorical / target column summaries.
	if categoricalColumns > 0 {
		messages = append(messages, fmt.Sprintf(
			"INFO: %d categorical column(s) detected - these will be excluded from PCA but available for visualization",
			categoricalColumns))
	}
	if targetColumns > 0 {
		messages = append(messages, fmt.Sprintf(
			"INFO: %d target column(s) detected - these will be excluded from PCA but available for visualization",
			targetColumns))
	}

	// Numeric column count.
	switch {
	case numericColumns < 2:
		messages = append(messages, fmt.Sprintf("ERROR: Need at least 2 numeric columns for PCA (found %d)", numericColumns))
	case numericColumns < 3:
		messages = append(messages, fmt.Sprintf("WARNING: Only %d numeric columns found - PCA results may be limited", numericColumns))
	default:
		messages = append(messages, fmt.Sprintf("INFO: %d numeric columns will be used for PCA analysis", numericColumns))
	}

	// Overall missing percentage.
	totalCells := in.Rows * in.Columns
	if totalCells > 0 && totalMissing > 0 {
		pct := float64(totalMissing) / float64(totalCells) * 100
		messages = append(messages, fmt.Sprintf("INFO: Dataset contains %.1f%% missing values (%d cells)", pct, totalMissing))
	}

	// Large dataset advisory.
	if in.Rows > 10000 {
		messages = append(messages, fmt.Sprintf("INFO: Large dataset detected (%d rows) - processing may take time", in.Rows))
	}

	// Row names present.
	if len(in.RowNames) > 0 {
		messages = append(messages, "INFO: Row names detected in first column")
	}

	// Validity: any ERROR message makes the result invalid.
	isValid := true
	for _, m := range messages {
		if strings.HasPrefix(m, "ERROR:") {
			isValid = false
			break
		}
	}

	return &ValidationResult{IsValid: isValid, Messages: messages}
}
