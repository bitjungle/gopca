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

package cobra

import (
	"math"
	"reflect"
	"testing"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
)

func TestParseFolds(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"10", 10, false},
		{"2", 2, false},
		{"loo", 0, false},
		{"LOO", 0, false},
		{"leave-one-out", 0, false},
		{"  loo  ", 0, false},
		{"1", 0, true},
		{"0", 0, true},
		{"-3", 0, true},
		{"many", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseFolds(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseFolds(%q) succeeded, expected an error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFolds(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseFolds(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestEncodeGroups checks that categorical levels become stable identifiers.
// Ordering by first appearance rather than by map iteration keeps a design
// reproducible from its seed, which is what makes a recorded CVReport meaningful.
func TestEncodeGroups(t *testing.T) {
	levels := []string{"b", "a", "b", "c", "a", "a"}
	want := []int{0, 1, 0, 2, 1, 1}

	for attempt := 0; attempt < 5; attempt++ {
		got := encodeGroups(levels)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("encodeGroups = %v, want %v", got, want)
		}
	}

	if got := encodeGroups(nil); len(got) != 0 {
		t.Errorf("encodeGroups(nil) = %v, want empty", got)
	}
}

// TestApplyMissingStrategyKeepsResponseAligned is the important one in this file.
//
// Dropping rows with incomplete predictors shortens the data matrix. The response
// and every categorical column are indexed by row, so they must lose exactly the
// same rows. Filtering the matrix alone would pair each surviving sample with the
// response of a different sample, and every downstream number would be wrong while
// looking entirely ordinary.
func TestApplyMissingStrategyKeepsResponseAligned(t *testing.T) {
	data := &pkgcsv.Data{
		Matrix: types.Matrix{
			{1, 1},
			{2, math.NaN()}, // dropped
			{3, 3},
			{math.NaN(), 4}, // dropped
			{5, 5},
		},
		Headers:  []string{"a", "b"},
		RowNames: []string{"r1", "r2", "r3", "r4", "r5"},
		Rows:     5,
		Columns:  2,
		CategoricalColumns: map[string][]string{
			"batch": {"A", "B", "C", "D", "E"},
		},
	}
	targets := map[string][]float64{
		"y#target": {10, 20, 30, 40, 50},
	}

	opts := &RegressOptions{MissingStrategy: "drop", Method: "svd"}
	if err := applyMissingStrategy(opts, data, targets); err != nil {
		t.Fatalf("applyMissingStrategy: %v", err)
	}

	if data.Rows != 3 {
		t.Fatalf("kept %d rows, want 3", data.Rows)
	}

	// Rows 0, 2 and 4 survive, so the response must be 10, 30, 50 in that order.
	wantResponse := []float64{10, 30, 50}
	if !reflect.DeepEqual(targets["y#target"], wantResponse) {
		t.Errorf("response = %v, want %v: the response is no longer aligned with its rows",
			targets["y#target"], wantResponse)
	}
	wantBatch := []string{"A", "C", "E"}
	if !reflect.DeepEqual(data.CategoricalColumns["batch"], wantBatch) {
		t.Errorf("grouping column = %v, want %v", data.CategoricalColumns["batch"], wantBatch)
	}
	wantNames := []string{"r1", "r3", "r5"}
	if !reflect.DeepEqual(data.RowNames, wantNames) {
		t.Errorf("row names = %v, want %v", data.RowNames, wantNames)
	}

	// The surviving predictor rows must be the ones whose response was kept.
	wantMatrix := types.Matrix{{1, 1}, {3, 3}, {5, 5}}
	if !reflect.DeepEqual(data.Matrix, wantMatrix) {
		t.Errorf("matrix = %v, want %v", data.Matrix, wantMatrix)
	}
}

// TestApplyMissingStrategyRefusesLearnedImputation pins a correctness decision.
// Mean and median imputation estimate values from the data, so applying them
// before cross-validation lets the held-out rows influence what the model trains
// on. Accepting them here would make every reported error optimistic with nothing
// in the output to reveal it.
func TestApplyMissingStrategyRefusesLearnedImputation(t *testing.T) {
	for _, strategy := range []string{"mean", "median"} {
		t.Run(strategy, func(t *testing.T) {
			data := &pkgcsv.Data{
				Matrix:  types.Matrix{{1, 1}, {2, math.NaN()}},
				Rows:    2,
				Columns: 2,
			}
			opts := &RegressOptions{MissingStrategy: strategy, Method: "svd"}
			err := applyMissingStrategy(opts, data, map[string][]float64{})
			if err == nil {
				t.Fatalf("%s imputation should be refused before cross-validation", strategy)
			}
		})
	}
}

func TestApplyMissingStrategyValidation(t *testing.T) {
	newData := func() *pkgcsv.Data {
		return &pkgcsv.Data{
			Matrix:  types.Matrix{{1, 1}, {2, math.NaN()}},
			Rows:    2,
			Columns: 2,
		}
	}

	tests := []struct {
		name    string
		opts    *RegressOptions
		wantErr bool
	}{
		{"error strategy refuses missing data", &RegressOptions{MissingStrategy: "error", Method: "svd"}, true},
		{"native requires nipals", &RegressOptions{MissingStrategy: "native", Method: "svd"}, true},
		{"native with nipals is accepted", &RegressOptions{MissingStrategy: "native", Method: "nipals"}, false},
		{"zero is accepted", &RegressOptions{MissingStrategy: "zero", Method: "svd"}, false},
		{"unknown strategy", &RegressOptions{MissingStrategy: "guess", Method: "svd"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyMissingStrategy(tt.opts, newData(), map[string][]float64{})
			if tt.wantErr != (err != nil) {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildPCRConfig(t *testing.T) {
	data := &pkgcsv.Data{Rows: 6, Columns: 3}
	categorical := map[string][]string{"batch": {"a", "a", "b", "b", "c", "c"}}

	t.Run("fixed component count skips cross-validation", func(t *testing.T) {
		opts := &RegressOptions{Response: "y", Components: 4, Scale: "standard", Method: "svd"}
		config, err := buildPCRConfig(opts, data, categorical)
		if err != nil {
			t.Fatalf("buildPCRConfig: %v", err)
		}
		if config.Selection.Mode != "fixed" || config.Selection.Fixed != 4 {
			t.Errorf("expected a fixed selection of 4, got %+v", config.Selection)
		}
		if !config.PCA.StandardScale {
			t.Error("expected standard scaling")
		}
	})

	t.Run("grouping resolves to per-row identifiers", func(t *testing.T) {
		opts := &RegressOptions{
			Response: "y", MaxComponents: 3, CV: "3", CVScheme: "random",
			CVGroup: "batch", Scale: "none", Method: "svd", Select: types.SelectOneSE,
		}
		config, err := buildPCRConfig(opts, data, categorical)
		if err != nil {
			t.Fatalf("buildPCRConfig: %v", err)
		}
		if config.Selection.CV.GroupBy != "batch" {
			t.Errorf("GroupBy = %q, want batch", config.Selection.CV.GroupBy)
		}
		want := []int{0, 0, 1, 1, 2, 2}
		if !reflect.DeepEqual(config.Selection.CV.Groups, want) {
			t.Errorf("Groups = %v, want %v", config.Selection.CV.Groups, want)
		}
	})

	t.Run("leave-one-out is zero folds", func(t *testing.T) {
		opts := &RegressOptions{
			Response: "y", MaxComponents: 3, CV: "loo", CVScheme: "random",
			Scale: "none", Method: "svd", Select: types.SelectMin,
		}
		config, err := buildPCRConfig(opts, data, categorical)
		if err != nil {
			t.Fatalf("buildPCRConfig: %v", err)
		}
		if config.Selection.CV.Folds != 0 {
			t.Errorf("Folds = %d, want 0 to mean one fold per group",
				config.Selection.CV.Folds)
		}
	})

	errorCases := []struct {
		name string
		opts *RegressOptions
	}{
		{"invalid scale", &RegressOptions{Scale: "sideways", CV: "5", CVScheme: "random"}},
		{"invalid scheme", &RegressOptions{Scale: "none", CV: "5", CVScheme: "diagonal"}},
		{"invalid fold count", &RegressOptions{Scale: "none", CV: "one", CVScheme: "random"}},
		{"unknown grouping column", &RegressOptions{
			Scale: "none", CV: "2", CVScheme: "random", CVGroup: "absent"}},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildPCRConfig(tt.opts, data, categorical); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

// TestBuildPCRConfigRejectsMismatchedGrouping guards against a grouping column
// whose length no longer matches the data, which is what a row-dropping bug
// upstream would look like by the time it reached here.
func TestBuildPCRConfigRejectsMismatchedGrouping(t *testing.T) {
	data := &pkgcsv.Data{Rows: 6, Columns: 3}
	categorical := map[string][]string{"batch": {"a", "b"}}

	opts := &RegressOptions{
		Response: "y", MaxComponents: 2, CV: "2", CVScheme: "random",
		CVGroup: "batch", Scale: "none", Method: "svd",
	}
	if _, err := buildPCRConfig(opts, data, categorical); err == nil {
		t.Error("expected a grouping column of the wrong length to be refused")
	}
}

func TestSortedKeysAreStable(t *testing.T) {
	m := map[string][]float64{"zebra": nil, "alpha": nil, "monkey": nil}
	want := []string{"alpha", "monkey", "zebra"}
	for attempt := 0; attempt < 10; attempt++ {
		if got := sortedKeys(m); !reflect.DeepEqual(got, want) {
			t.Fatalf("sortedKeys = %v, want %v: map order must not leak into output", got, want)
		}
	}
}

func TestCountObserved(t *testing.T) {
	observed, total := countObserved([]float64{1, math.NaN(), 3, math.Inf(1), 5})
	if observed != 3 || total != 5 {
		t.Errorf("countObserved = (%d, %d), want (3, 5)", observed, total)
	}
}

func TestUnknownResponseErrorMentionsCategorical(t *testing.T) {
	categorical := map[string][]string{"species": {"a", "b"}}
	err := unknownResponseError("species", map[string][]float64{"y#target": nil}, categorical)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !contains(err.Error(), "categorical") || !contains(err.Error(), "classification") {
		t.Errorf("the message should explain why a categorical column cannot be a response: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle || len(needle) == 0 ||
			indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
