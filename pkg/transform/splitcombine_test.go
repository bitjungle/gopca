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

package transform

import (
	"strings"
	"testing"
)

func splitInput(values ...string) Input {
	data := make([][]string, len(values))
	for i, v := range values {
		data[i] = []string{v, "1"}
	}
	return Input{
		Data:               data,
		Headers:            []string{"ID", "Score"},
		ColumnTypes:        map[string]string{"ID": "categorical", "Score": "numeric"},
		CategoricalColumns: map[string][]string{"ID": append([]string(nil), values...)},
		Rows:               len(values),
		Columns:            2,
	}
}

func columnValues(t *testing.T, res *Result, name string) []string {
	t.Helper()
	index := findColumn(res.Headers, name)
	if index == -1 {
		t.Fatalf("no column %q in %v", name, res.Headers)
	}
	out := make([]string, 0, len(res.Data))
	for _, row := range res.Data {
		out = append(out, row[index])
	}
	return out
}

// TestApply_Split covers the case the feature exists for: a grouping key
// buried in a structured sample identifier.
//
// PCR's --cv-group needs a column naming the group, and grouping is usually
// implied by an ID like "B3_S12_r1". Without this there is no way to produce
// that column, so the flag exists with no way to feed it (#866).
func TestApply_Split(t *testing.T) {
	in := splitInput("B3_S12_r1", "B3_S12_r2", "B4_S07_r1")

	res, err := Apply(in, Options{Type: Split, Columns: []string{"ID"}, Delimiter: "_"})
	if err != nil {
		t.Fatalf("Apply split: %v", err)
	}

	if got := strings.Join(res.NewColumns, ","); got != "ID_1,ID_2,ID_3" {
		t.Fatalf("new columns = %s, want ID_1,ID_2,ID_3", got)
	}
	if got := strings.Join(columnValues(t, res, "ID_1"), ","); got != "B3,B3,B4" {
		t.Errorf("ID_1 = %s, want B3,B3,B4 -- this is the group key", got)
	}
	if got := strings.Join(columnValues(t, res, "ID_3"), ","); got != "r1,r2,r1" {
		t.Errorf("ID_3 = %s, want r1,r2,r1", got)
	}

	// The source is kept, as with the encoders (#854).
	if findColumn(res.Headers, "ID") == -1 {
		t.Errorf("the source column should be kept, headers are %v", res.Headers)
	}
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers", i, len(row), len(res.Headers))
		}
	}
}

// TestApply_Split_RaggedRows is the assertion that catches the quiet failure.
//
// A row with fewer parts must leave its trailing columns empty. Shifting the
// values left instead would put a replicate number under a batch heading --
// plausible-looking data in the wrong column, which nothing downstream detects.
func TestApply_Split_RaggedRows(t *testing.T) {
	in := splitInput("B3_S12_r1", "B4", "B5_S01")

	res, err := Apply(in, Options{Type: Split, Columns: []string{"ID"}, Delimiter: "_"})
	if err != nil {
		t.Fatalf("Apply split: %v", err)
	}

	for name, want := range map[string]string{
		"ID_1": "B3,B4,B5",
		"ID_2": "S12,,S01",
		"ID_3": "r1,,",
	} {
		if got := strings.Join(columnValues(t, res, name), ","); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestApply_Split_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		values    []string
		delimiter string
		wantNew   int
		wantMsg   string
	}{
		{
			name:      "delimiter absent from every value",
			values:    []string{"abc", "def"},
			delimiter: "_",
			wantNew:   0,
			wantMsg:   "no value contains",
		},
		{
			name:      "blank cells produce no parts",
			values:    []string{"a_b", "", "c_d"},
			delimiter: "_",
			wantNew:   2,
		},
		{
			name:      "too many parts is refused",
			values:    []string{strings.Repeat("x_", 25) + "x"},
			delimiter: "_",
			wantNew:   0,
			wantMsg:   "more than the 20 allowed",
		},
		{
			name:      "multi-character delimiter",
			values:    []string{"a::b", "c::d"},
			delimiter: "::",
			wantNew:   2,
		},
		{
			name:      "delimiter is literal, not a pattern",
			values:    []string{"a.b", "c.d"},
			delimiter: ".",
			wantNew:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Apply(splitInput(tt.values...),
				Options{Type: Split, Columns: []string{"ID"}, Delimiter: tt.delimiter})
			if err != nil {
				t.Fatalf("Apply split: %v", err)
			}
			if len(res.NewColumns) != tt.wantNew {
				t.Errorf("new columns = %v, want %d", res.NewColumns, tt.wantNew)
			}
			if tt.wantMsg != "" && !strings.Contains(strings.Join(res.Messages, " "), tt.wantMsg) {
				t.Errorf("messages should mention %q, got %v", tt.wantMsg, res.Messages)
			}
		})
	}
}

// TestApply_Split_BlankStaysBlank guards the empty-cell handling.
//
// Splitting "" yields one empty string, so a naive implementation would count
// that as a part and make a blank look like a value.
func TestApply_Split_BlankStaysBlank(t *testing.T) {
	res, err := Apply(splitInput("a_b", "  ", "c_d"),
		Options{Type: Split, Columns: []string{"ID"}, Delimiter: "_"})
	if err != nil {
		t.Fatalf("Apply split: %v", err)
	}
	if got := columnValues(t, res, "ID_1"); got[1] != "" {
		t.Errorf("a blank source produced %q in ID_1, want empty", got[1])
	}
}

func TestApply_Split_RequiresDelimiter(t *testing.T) {
	if _, err := Apply(splitInput("a_b"), Options{Type: Split, Columns: []string{"ID"}}); err == nil {
		t.Error("splitting without a delimiter should fail")
	}
}

func TestApply_Split_TypesNewColumns(t *testing.T) {
	res, err := Apply(splitInput("B3_12", "B4_07"),
		Options{Type: Split, Columns: []string{"ID"}, Delimiter: "_"})
	if err != nil {
		t.Fatalf("Apply split: %v", err)
	}
	if res.ColumnTypes["ID_1"] != "categorical" {
		t.Errorf("ID_1 holds B3/B4 and should be categorical, got %q", res.ColumnTypes["ID_1"])
	}
	if res.ColumnTypes["ID_2"] != "numeric" {
		t.Errorf("ID_2 holds 12/07 and should be numeric, got %q", res.ColumnTypes["ID_2"])
	}
	if _, ok := res.CategoricalColumns["ID_1"]; !ok {
		t.Error("a categorical column must be registered in CategoricalColumns")
	}
	if _, ok := res.CategoricalColumns["ID_2"]; ok {
		t.Error("a numeric column must not be in CategoricalColumns")
	}
}

func TestApply_Split_RemoveOriginal(t *testing.T) {
	res, err := Apply(splitInput("a_b", "c_d"),
		Options{Type: Split, Columns: []string{"ID"}, Delimiter: "_", RemoveOriginal: true})
	if err != nil {
		t.Fatalf("Apply split: %v", err)
	}
	if findColumn(res.Headers, "ID") != -1 {
		t.Errorf("'ID' should have been removed, headers are %v", res.Headers)
	}
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers: %v", i, len(row), len(res.Headers), row)
		}
	}
	if got := strings.Join(columnValues(t, res, "ID_1"), ","); got != "a,c" {
		t.Errorf("ID_1 = %s after removing the source", got)
	}
}

// ─── Combine ─────────────────────────────────────────────────────────────────

func combineInput() Input {
	return Input{
		Data:               [][]string{{"Oslo", "2024", "5"}, {"Bergen", "2025", "6"}},
		Headers:            []string{"Site", "Year", "Score"},
		ColumnTypes:        map[string]string{"Site": "categorical", "Year": "numeric", "Score": "numeric"},
		CategoricalColumns: map[string][]string{"Site": {"Oslo", "Bergen"}},
		Rows:               2,
		Columns:            3,
	}
}

func TestApply_Combine(t *testing.T) {
	res, err := Apply(combineInput(), Options{
		Type: Combine, Columns: []string{"Site", "Year"}, Separator: "_"})
	if err != nil {
		t.Fatalf("Apply combine: %v", err)
	}

	if len(res.NewColumns) != 1 || res.NewColumns[0] != "Site_Year" {
		t.Fatalf("new columns = %v, want [Site_Year]", res.NewColumns)
	}
	if got := strings.Join(columnValues(t, res, "Site_Year"), ","); got != "Oslo_2024,Bergen_2025" {
		t.Errorf("Site_Year = %s", got)
	}
	// Sources kept.
	for _, name := range []string{"Site", "Year"} {
		if findColumn(res.Headers, name) == -1 {
			t.Errorf("source %q should be kept, headers are %v", name, res.Headers)
		}
	}
}

// TestApply_Combine_OrderFollowsTheRequest pins that Site+Year and Year+Site
// are different keys, and the caller picked one.
func TestApply_Combine_OrderFollowsTheRequest(t *testing.T) {
	res, err := Apply(combineInput(), Options{
		Type: Combine, Columns: []string{"Year", "Site"}, Separator: "-"})
	if err != nil {
		t.Fatalf("Apply combine: %v", err)
	}
	if got := columnValues(t, res, "Year_Site")[0]; got != "2024-Oslo" {
		t.Errorf("got %q, want 2024-Oslo -- the requested order, not the table order", got)
	}
}

func TestApply_Combine_Options(t *testing.T) {
	t.Run("custom name", func(t *testing.T) {
		res, err := Apply(combineInput(), Options{
			Type: Combine, Columns: []string{"Site", "Year"}, Separator: "_",
			NewColumnName: "BatchKey"})
		if err != nil {
			t.Fatalf("Apply combine: %v", err)
		}
		if res.NewColumns[0] != "BatchKey" {
			t.Errorf("new column = %q, want BatchKey", res.NewColumns[0])
		}
	})

	t.Run("empty separator is allowed", func(t *testing.T) {
		res, err := Apply(combineInput(), Options{
			Type: Combine, Columns: []string{"Site", "Year"}})
		if err != nil {
			t.Fatalf("Apply combine: %v", err)
		}
		if got := columnValues(t, res, "Site_Year")[0]; got != "Oslo2024" {
			t.Errorf("got %q, want Oslo2024", got)
		}
	})

	t.Run("name collision is suffixed", func(t *testing.T) {
		in := combineInput()
		in.Headers = append(in.Headers, "Site_Year")
		in.ColumnTypes["Site_Year"] = "categorical"
		for i := range in.Data {
			in.Data[i] = append(in.Data[i], "taken")
		}
		in.Columns = len(in.Headers)

		res, err := Apply(in, Options{Type: Combine, Columns: []string{"Site", "Year"}, Separator: "_"})
		if err != nil {
			t.Fatalf("Apply combine: %v", err)
		}
		if res.NewColumns[0] == "Site_Year" {
			t.Errorf("the new column overwrote an existing one: %v", res.Headers)
		}
		if got := columnValues(t, res, "Site_Year")[0]; got != "taken" {
			t.Errorf("the existing column was overwritten: %q", got)
		}
	})

	t.Run("fewer than two columns is refused", func(t *testing.T) {
		if _, err := Apply(combineInput(), Options{Type: Combine, Columns: []string{"Site"}}); err == nil {
			t.Error("combining one column should fail")
		}
	})
}

// TestApply_Combine_RemoveOriginal checks the sources are removed without
// disturbing the result.
//
// Removal happens from the highest index down; going upward would invalidate
// the remaining indices as the columns beneath them shifted.
func TestApply_Combine_RemoveOriginal(t *testing.T) {
	res, err := Apply(combineInput(), Options{
		Type: Combine, Columns: []string{"Site", "Year"}, Separator: "_", RemoveOriginal: true})
	if err != nil {
		t.Fatalf("Apply combine: %v", err)
	}

	for _, name := range []string{"Site", "Year"} {
		if findColumn(res.Headers, name) != -1 {
			t.Errorf("source %q should have been removed, headers are %v", name, res.Headers)
		}
	}
	if got := strings.Join(columnValues(t, res, "Site_Year"), ","); got != "Oslo_2024,Bergen_2025" {
		t.Errorf("the combined column was damaged by the removal: %s", got)
	}
	if got := strings.Join(columnValues(t, res, "Score"), ","); got != "5,6" {
		t.Errorf("an unrelated column was damaged by the removal: %s", got)
	}
	for i, row := range res.Data {
		if len(row) != len(res.Headers) {
			t.Errorf("row %d has %d cells but there are %d headers: %v", i, len(row), len(res.Headers), row)
		}
	}
}

// TestApply_Combine_RejectsDuplicateColumns guards against silent data loss.
//
// Joining a column to itself produces nothing a user wants, and with
// RemoveOriginal it destroyed data: the same index was removed twice, and the
// second removal took whichever column had shifted into its place. Asking for
// ["A", "A"] deleted both A and its neighbour, leaving headers [C A_A] from
// [A B C] with no indication anything had gone.
func TestApply_Combine_RejectsDuplicateColumns(t *testing.T) {
	in := Input{
		Data:        [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
		Headers:     []string{"A", "B", "C"},
		ColumnTypes: map[string]string{"A": "categorical", "B": "categorical", "C": "categorical"},
		Rows:        2,
		Columns:     3,
	}

	_, err := Apply(in, Options{
		Type: Combine, Columns: []string{"A", "A"}, Separator: "_", RemoveOriginal: true})
	if err == nil {
		t.Fatal("combining a column with itself was accepted; with RemoveOriginal this " +
			"deletes a second, unrelated column")
	}
	if !strings.Contains(err.Error(), "more than once") {
		t.Errorf("the error should say what is wrong, got %v", err)
	}

	// The input must be untouched by the refusal.
	if len(in.Headers) != 3 || in.Data[0][1] != "b" {
		t.Errorf("a refused combination modified the input: %v / %v", in.Headers, in.Data)
	}
}

// TestGetTransformableColumns_SplitCombine checks which columns are offered.
//
// Targets are excluded for the same reason the numeric transforms exclude them:
// "#target" marks a column as reference information rather than a measurement,
// and restructuring one silently breaks that role -- with RemoveOriginal it
// would delete the target outright.
func TestGetTransformableColumns_SplitCombine(t *testing.T) {
	in := Input{
		Headers: []string{"Cat", "Num", "Yield#target"},
		ColumnTypes: map[string]string{
			"Cat": "categorical", "Num": "numeric", "Yield#target": "numeric",
		},
	}
	for _, transformType := range []Type{Split, Combine} {
		got := GetTransformableColumns(in, transformType)
		if len(got) != 2 {
			t.Errorf("%s offered %v, want the two non-target columns", transformType, got)
		}
		for _, name := range got {
			if strings.HasSuffix(name, "#target") {
				t.Errorf("%s offered the target column %q", transformType, name)
			}
		}
	}
}
