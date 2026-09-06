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

	// Constant columns.
	//
	// A column whose every value is identical contributes nothing to any
	// component. It is not an error -- PCA handles it cleanly, scaling it to
	// zeros rather than producing NaN -- but it is silent dead weight: it
	// inflates the variable count and sits at the origin of every loadings
	// plot, where a reader may take its position as meaningful rather than as
	// an artefact of having no variance at all (#867).
	//
	// Reported, never removed. Whether it matters is the user's judgement.
	if constant := constantColumns(in); len(constant) > 0 {
		noun, verb := "Column", "has"
		if len(constant) > 1 {
			noun, verb = "Columns", "have"
		}
		messages = append(messages, fmt.Sprintf(
			"WARNING: %s %s %s no variation and will contribute nothing to any component",
			noun, strings.Join(constant, ", "), verb))
	}

	// Row names present.
	if len(in.RowNames) > 0 {
		messages = append(messages, "INFO: Row names detected in first column")

		// Row names identify rows: they label the points in a scores plot, so
		// two rows sharing a name are indistinguishable there and a row with no
		// name is unlabelled. GoCSV enforces this when a column is promoted to
		// row names, but the load path assigns the first column without
		// checking, so a file can arrive in this state (#859).
		//
		// A warning, not an error. Refusing to open or analyse a file over
		// duplicate labels would be the mistake #801 was filed about; the
		// numbers are still perfectly analysable, only the labelling is
		// ambiguous, and the user is the one who can say whether that matters.
		if duplicates, blanks := countRowNameProblems(in.RowNames); duplicates > 0 || blanks > 0 {
			switch {
			case duplicates > 0 && blanks > 0:
				messages = append(messages, fmt.Sprintf(
					"WARNING: Row names are not unique identifiers (%d repeated, %d empty) - "+
						"points sharing a label cannot be told apart in plots", duplicates, blanks))
			case duplicates > 0:
				messages = append(messages, fmt.Sprintf(
					"WARNING: %d row name(s) are repeated - points sharing a label cannot be "+
						"told apart in plots", duplicates))
			default:
				messages = append(messages, fmt.Sprintf(
					"WARNING: %d row name(s) are empty - those points will be unlabelled in plots",
					blanks))
			}
		}
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

// countRowNameProblems counts repeated and empty row names.
//
// Repeats are counted as the number of values appearing more than once, not the
// number of excess rows, so three rows called "P1" report one repeated name.
// Comparison is on the trimmed value, matching checkRowNameCandidate in GoCSV:
// two labels that render identically are the same label.
func countRowNameProblems(rowNames []string) (duplicates, blanks int) {
	seen := make(map[string]int, len(rowNames))
	for _, name := range rowNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			blanks++
			continue
		}
		seen[trimmed]++
		if seen[trimmed] == 2 {
			duplicates++
		}
	}
	return duplicates, blanks
}

// constantColumns returns the names of columns whose every present value is
// identical, in the order the columns appear.
//
// Comparison is on the trimmed string rather than a parsed number, so it works
// for categorical and numeric columns alike and needs no tolerance: "1.0" and
// "1.00" are different strings but would be the same number, and reporting a
// column that merely looks constant is worse than staying quiet. Blanks are
// skipped -- a column of one value and some gaps is still constant in the sense
// that matters, since the gaps carry no variation either.
func constantColumns(in ValidationInput) []string {
	var constant []string

	for colIndex, header := range in.Headers {
		seen := ""
		found := false
		varies := false

		for _, row := range in.Data {
			if colIndex >= len(row) {
				continue
			}
			value := strings.TrimSpace(row[colIndex])
			if value == "" {
				continue
			}
			if !found {
				seen, found = value, true
				continue
			}
			if value != seen {
				varies = true
				break
			}
		}

		// A column with nothing in it is empty, not constant; the missing-data
		// checks above already cover that case.
		if found && !varies {
			constant = append(constant, header)
		}
	}

	return constant
}
