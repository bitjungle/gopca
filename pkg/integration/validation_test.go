// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"strings"
	"testing"
)

// helpers

func makeInput(rows, cols int, headers []string, colTypes map[string]string, data [][]string) ValidationInput {
	return ValidationInput{
		Headers:     headers,
		Data:        data,
		ColumnTypes: colTypes,
		Rows:        rows,
		Columns:     cols,
	}
}

func hasPrefix(messages []string, prefix string) bool {
	for _, m := range messages {
		if strings.HasPrefix(m, prefix) {
			return true
		}
	}
	return false
}

func countPrefix(messages []string, prefix string) int {
	n := 0
	for _, m := range messages {
		if strings.HasPrefix(m, prefix) {
			n++
		}
	}
	return n
}

// ─── IsValid ──────────────────────────────────────────────────────────────────

func TestValidateForGoPCA_ValidDataset(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
	)
	res := ValidateForGoPCA(in)
	if !res.IsValid {
		t.Errorf("expected valid dataset, got messages: %v", res.Messages)
	}
	if hasPrefix(res.Messages, "ERROR:") {
		t.Error("expected no ERROR messages for clean dataset")
	}
}

// ─── Row count ────────────────────────────────────────────────────────────────

func TestValidateForGoPCA_TooFewRows(t *testing.T) {
	in := makeInput(1, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{{"1", "2", "3"}},
	)
	res := ValidateForGoPCA(in)
	if res.IsValid {
		t.Error("expected invalid for <2 rows")
	}
	if !hasPrefix(res.Messages, "ERROR:") {
		t.Error("expected ERROR message for too few rows")
	}
}

func TestValidateForGoPCA_ZeroRows(t *testing.T) {
	in := makeInput(0, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{},
	)
	res := ValidateForGoPCA(in)
	if res.IsValid {
		t.Error("expected invalid for 0 rows")
	}
}

// ─── Numeric column count ─────────────────────────────────────────────────────

func TestValidateForGoPCA_ZeroNumericColumns(t *testing.T) {
	in := makeInput(3, 2,
		[]string{"A", "B"},
		map[string]string{"A": "categorical", "B": "categorical"},
		[][]string{{"x", "y"}, {"x", "y"}, {"x", "y"}},
	)
	res := ValidateForGoPCA(in)
	if res.IsValid {
		t.Error("expected invalid for 0 numeric columns")
	}
	if !hasPrefix(res.Messages, "ERROR:") {
		t.Error("expected ERROR message for zero numeric columns")
	}
}

func TestValidateForGoPCA_OneNumericColumn(t *testing.T) {
	in := makeInput(3, 2,
		[]string{"A", "B"},
		map[string]string{"A": "numeric", "B": "categorical"},
		[][]string{{"1", "x"}, {"2", "x"}, {"3", "x"}},
	)
	res := ValidateForGoPCA(in)
	if res.IsValid {
		t.Error("expected invalid for 1 numeric column")
	}
}

func TestValidateForGoPCA_TwoNumericColumns_ValidButWarning(t *testing.T) {
	in := makeInput(3, 2,
		[]string{"A", "B"},
		map[string]string{"A": "numeric", "B": "numeric"},
		[][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}},
	)
	res := ValidateForGoPCA(in)
	if !res.IsValid {
		t.Error("expected valid for exactly 2 numeric columns")
	}
	if !hasPrefix(res.Messages, "WARNING:") {
		t.Error("expected WARNING for only 2 numeric columns")
	}
}

func TestValidateForGoPCA_ThreeNumericColumns_Info(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
	)
	res := ValidateForGoPCA(in)
	if !res.IsValid {
		t.Errorf("expected valid, got: %v", res.Messages)
	}
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "3 numeric columns") {
			found = true
		}
	}
	if !found {
		t.Error("expected INFO message about 3 numeric columns")
	}
}

// ─── Missing values ───────────────────────────────────────────────────────────

func TestValidateForGoPCA_HighMissingColumn(t *testing.T) {
	// 2 out of 3 rows missing in column A → 66.7% → WARNING
	in := makeInput(3, 2,
		[]string{"A", "B"},
		map[string]string{"A": "numeric", "B": "numeric"},
		[][]string{{"1", "1"}, {"NA", "2"}, {"", "3"}},
	)
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "WARNING:") && strings.Contains(m, "Column 'A'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected WARNING for high missing in column A, got: %v", res.Messages)
	}
}

func TestValidateForGoPCA_MissingValueTokens(t *testing.T) {
	// All recognised missing tokens should be counted.
	tokens := []string{"NA", "N/A", "nan", "NaN", "null", "NULL", ""}
	for _, tok := range tokens {
		in := makeInput(3, 2,
			[]string{"A", "B"},
			map[string]string{"A": "numeric", "B": "numeric"},
			// Only one cell in A — use the token; make rest valid.
			[][]string{{tok, "1"}, {"0", "2"}, {"0", "3"}},
		)
		res := ValidateForGoPCA(in)
		found := false
		for _, m := range res.Messages {
			if strings.Contains(m, "missing values") {
				found = true
			}
		}
		if !found {
			t.Errorf("token %q: expected missing-value message", tok)
		}
	}
}

func TestValidateForGoPCA_OverallMissingInfo(t *testing.T) {
	// One missing cell in a 2×2 dataset → INFO about overall missing%.
	in := makeInput(2, 2,
		[]string{"A", "B"},
		map[string]string{"A": "numeric", "B": "numeric"},
		[][]string{{"NA", "1"}, {"2", "3"}},
	)
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "missing values") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected INFO about overall missing%%, got: %v", res.Messages)
	}
}

func TestValidateForGoPCA_NoMissingNoOverallInfo(t *testing.T) {
	// Zero missing cells → no overall missing INFO.
	in := makeInput(2, 2,
		[]string{"A", "B"},
		map[string]string{"A": "numeric", "B": "numeric"},
		[][]string{{"1", "2"}, {"3", "4"}},
	)
	res := ValidateForGoPCA(in)
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "missing values") {
			t.Errorf("unexpected missing-value INFO for clean dataset: %q", m)
		}
	}
}

// ─── Column type summaries ────────────────────────────────────────────────────

func TestValidateForGoPCA_CategoricalColumnsInfo(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "Cat"},
		map[string]string{"A": "numeric", "B": "numeric", "Cat": "categorical"},
		[][]string{{"1", "2", "x"}, {"3", "4", "y"}, {"5", "6", "x"}},
	)
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "categorical") {
			found = true
		}
	}
	if !found {
		t.Error("expected INFO about categorical columns")
	}
}

func TestValidateForGoPCA_TargetColumnsInfo(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "T"},
		map[string]string{"A": "numeric", "B": "numeric", "T": "target"},
		[][]string{{"1", "2", "0"}, {"3", "4", "1"}, {"5", "6", "0"}},
	)
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "target") {
			found = true
		}
	}
	if !found {
		t.Error("expected INFO about target columns")
	}
}

// ─── Large dataset / row names ────────────────────────────────────────────────

func TestValidateForGoPCA_LargeDatasetInfo(t *testing.T) {
	in := ValidationInput{
		Headers:     []string{"A", "B", "C"},
		Data:        nil,
		ColumnTypes: map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		Rows:        10001,
		Columns:     3,
	}
	// Data is nil — row-count check fires first; then large-dataset INFO.
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "Large dataset") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Large dataset INFO for >10000 rows, got: %v", res.Messages)
	}
}

func TestValidateForGoPCA_RowNamesInfo(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
	)
	in.RowNames = []string{"r1", "r2", "r3"}
	res := ValidateForGoPCA(in)
	found := false
	for _, m := range res.Messages {
		if strings.HasPrefix(m, "INFO:") && strings.Contains(m, "Row names") {
			found = true
		}
	}
	if !found {
		t.Error("expected INFO about row names")
	}
}

func TestValidateForGoPCA_NoRowNamesNoInfo(t *testing.T) {
	in := makeInput(3, 3,
		[]string{"A", "B", "C"},
		map[string]string{"A": "numeric", "B": "numeric", "C": "numeric"},
		[][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
	)
	res := ValidateForGoPCA(in)
	for _, m := range res.Messages {
		if strings.Contains(m, "Row names") {
			t.Errorf("unexpected row-name message when RowNames is nil: %q", m)
		}
	}
}

// ─── Multiple errors ──────────────────────────────────────────────────────────

func TestValidateForGoPCA_MultipleErrors(t *testing.T) {
	// 1 row + 0 numeric columns → 2 ERRORs, isValid=false.
	in := makeInput(1, 2,
		[]string{"A", "B"},
		map[string]string{"A": "categorical", "B": "categorical"},
		[][]string{{"x", "y"}},
	)
	res := ValidateForGoPCA(in)
	if res.IsValid {
		t.Error("expected invalid")
	}
	if countPrefix(res.Messages, "ERROR:") < 2 {
		t.Errorf("expected at least 2 ERROR messages, got: %v", res.Messages)
	}
}
