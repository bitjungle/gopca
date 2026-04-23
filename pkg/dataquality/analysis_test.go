// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

import (
	"math"
	"strings"
	"testing"
)

// ─── AnalyzeDataQuality ──────────────────────────────────────────────────────

func TestAnalyzeDataQuality_EmptyData(t *testing.T) {
	_, err := AnalyzeDataQuality(AnalysisInput{})
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestAnalyzeDataQuality_BasicReport(t *testing.T) {
	in := AnalysisInput{
		Data:        [][]string{{"1.0", "a"}, {"2.0", "b"}, {"3.0", "c"}},
		Headers:     []string{"X", "Y"},
		ColumnTypes: map[string]string{"X": "numeric", "Y": "categorical"},
		Rows:        3,
		Columns:     2,
	}
	report, err := AnalyzeDataQuality(in)
	if err != nil {
		t.Fatalf("AnalyzeDataQuality: %v", err)
	}
	if report.DataProfile.Rows != 3 {
		t.Errorf("expected 3 rows, got %d", report.DataProfile.Rows)
	}
	if report.DataProfile.NumericColumns != 1 {
		t.Errorf("expected 1 numeric column, got %d", report.DataProfile.NumericColumns)
	}
	if report.DataProfile.CategoricalColumns != 1 {
		t.Errorf("expected 1 categorical column, got %d", report.DataProfile.CategoricalColumns)
	}
	if report.DataProfile.MissingPercent != 0 {
		t.Errorf("expected 0%% missing, got %v", report.DataProfile.MissingPercent)
	}
	if len(report.ColumnAnalysis) != 2 {
		t.Errorf("expected 2 column analyses, got %d", len(report.ColumnAnalysis))
	}
	if report.QualityScore <= 0 || report.QualityScore > 100 {
		t.Errorf("quality score out of range: %v", report.QualityScore)
	}
}

// ─── analyzeNumericStats ─────────────────────────────────────────────────────

func TestAnalyzeNumericStats_Basic(t *testing.T) {
	data := [][]string{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}}
	stats := analyzeNumericStats(data, 5, 0)

	if stats.Count != 5 {
		t.Errorf("count: expected 5, got %d", stats.Count)
	}
	if stats.Missing != 0 {
		t.Errorf("missing: expected 0, got %d", stats.Missing)
	}
	if stats.Mean == nil || !almostEqual(*stats.Mean, 3.0, tol) {
		t.Errorf("mean: expected 3.0, got %v", stats.Mean)
	}
	if stats.Median == nil || !almostEqual(*stats.Median, 3.0, tol) {
		t.Errorf("median: expected 3.0, got %v", stats.Median)
	}
	if stats.Min == nil || *stats.Min != 1.0 {
		t.Errorf("min: expected 1.0, got %v", stats.Min)
	}
	if stats.Max == nil || *stats.Max != 5.0 {
		t.Errorf("max: expected 5.0, got %v", stats.Max)
	}
}

func TestAnalyzeNumericStats_WithMissing(t *testing.T) {
	data := [][]string{{"1"}, {""}, {"3"}}
	stats := analyzeNumericStats(data, 3, 0)

	if stats.Missing != 1 {
		t.Errorf("missing: expected 1, got %d", stats.Missing)
	}
	if !almostEqual(stats.MissingPercent, 100.0/3.0, 0.01) {
		t.Errorf("missing percent: expected ~33.3, got %v", stats.MissingPercent)
	}
	if stats.Mean == nil || !almostEqual(*stats.Mean, 2.0, tol) {
		t.Errorf("mean: expected 2.0 (ignores missing), got %v", stats.Mean)
	}
}

func TestAnalyzeNumericStats_AllMissing(t *testing.T) {
	data := [][]string{{""}, {"NA"}}
	stats := analyzeNumericStats(data, 2, 0)

	if stats.Missing != 2 {
		t.Errorf("expected 2 missing, got %d", stats.Missing)
	}
	if stats.Mean != nil {
		t.Error("mean should be nil when all values are missing")
	}
}

// ─── analyzeCategoricalStats ─────────────────────────────────────────────────

func TestAnalyzeCategoricalStats_Basic(t *testing.T) {
	data := [][]string{{"a"}, {"b"}, {"a"}, {"a"}, {""}}
	stats := analyzeCategoricalStats(data, 5, 0)

	if stats.Count != 5 {
		t.Errorf("count: expected 5, got %d", stats.Count)
	}
	if stats.Missing != 1 {
		t.Errorf("missing: expected 1, got %d", stats.Missing)
	}
	if stats.Unique != 2 {
		t.Errorf("unique: expected 2, got %d", stats.Unique)
	}
	if stats.Mode == nil || *stats.Mode != "a" {
		t.Errorf("mode: expected 'a', got %v", stats.Mode)
	}
}

// ─── analyzeDistribution ─────────────────────────────────────────────────────

func TestAnalyzeDistribution_InsufficientData(t *testing.T) {
	data := [][]string{{"1"}, {"2"}, {"3"}} // fewer than 10
	dist := analyzeDistribution(data, 3, 0)
	if dist.Histogram != nil {
		t.Error("expected nil histogram for <10 values")
	}
}

func TestAnalyzeDistribution_Normal(t *testing.T) {
	// Symmetric values → skewness ≈ 0 → should classify as normal
	vals := [][]string{
		{"1"}, {"2"}, {"3"}, {"4"}, {"5"},
		{"5"}, {"4"}, {"3"}, {"2"}, {"1"},
	}
	dist := analyzeDistribution(vals, len(vals), 0)
	if !dist.IsNormal {
		t.Error("symmetric data should be classified as normal")
	}
	if dist.DistType != "normal" {
		t.Errorf("expected 'normal', got %q", dist.DistType)
	}
}

// ─── detectOutliers ──────────────────────────────────────────────────────────

func TestDetectOutliers_Clear(t *testing.T) {
	// 9 values near 5, one extreme outlier
	data := [][]string{
		{"5"}, {"5"}, {"5"}, {"5"}, {"5"},
		{"5"}, {"5"}, {"5"}, {"5"}, {"1000"},
	}
	stats := analyzeNumericStats(data, len(data), 0)
	outliers := detectOutliers(data, len(data), 0, stats)
	if len(outliers) == 0 {
		t.Error("expected at least one outlier detected")
	}
	found := false
	for _, o := range outliers {
		if o.Value == "1000" {
			found = true
		}
	}
	if !found {
		t.Error("expected 1000 to be detected as outlier")
	}
}

func TestDetectOutliers_NoStats(t *testing.T) {
	// stats with nil Q1/Q3 should return empty slice
	data := [][]string{{"1"}}
	outliers := detectOutliers(data, 1, 0, ColumnStatistics{})
	if len(outliers) != 0 {
		t.Error("expected no outliers when stats are nil")
	}
}

// ─── countDuplicateRows ──────────────────────────────────────────────────────

func TestCountDuplicateRows(t *testing.T) {
	data := [][]string{
		{"a", "b"},
		{"c", "d"},
		{"a", "b"}, // duplicate of row 0
		{"a", "b"}, // another duplicate
	}
	got := countDuplicateRows(data, len(data))
	// rows 2 and 3 are duplicates of row 0 → 2 counted
	if got != 2 {
		t.Errorf("expected 2 duplicate rows, got %d", got)
	}
}

func TestCountDuplicateRows_NoDuplicates(t *testing.T) {
	data := [][]string{{"1"}, {"2"}, {"3"}}
	if got := countDuplicateRows(data, len(data)); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// ─── estimateMemorySize ──────────────────────────────────────────────────────

func TestEstimateMemorySize(t *testing.T) {
	tests := []struct {
		rows, cols int
		wantSuffix string
	}{
		{1, 1, "B"},
		{100, 10, "KB"},
		{10000, 100, "MB"},
	}
	for _, tt := range tests {
		got := estimateMemorySize(tt.rows, tt.cols)
		if len(got) == 0 {
			t.Errorf("estimateMemorySize(%d,%d) returned empty", tt.rows, tt.cols)
		}
		// Just check it ends with the expected unit
		if !strings.HasSuffix(got, tt.wantSuffix) {
			t.Errorf("estimateMemorySize(%d,%d) = %q, want suffix %q", tt.rows, tt.cols, got, tt.wantSuffix)
		}
	}
}

// ─── calculateQualityScore ───────────────────────────────────────────────────

func TestCalculateQualityScore_Perfect(t *testing.T) {
	report := &DataQualityReport{
		DataProfile: DataProfile{
			Rows:           10,
			NumericColumns: 5,
		},
	}
	score := calculateQualityScore(report)
	if score != 100.0 {
		t.Errorf("perfect dataset: expected 100, got %v", score)
	}
}

func TestCalculateQualityScore_Clamped(t *testing.T) {
	// Simulate terrible data
	report := &DataQualityReport{
		DataProfile: DataProfile{
			Rows:           10,
			NumericColumns: 0,
			MissingPercent: 90,
		},
	}
	score := calculateQualityScore(report)
	if score < 0 || score > 100 {
		t.Errorf("score must be 0-100, got %v", score)
	}
}

// ─── calculateCorrelations ───────────────────────────────────────────────────

func TestCalculateCorrelations_Perfect(t *testing.T) {
	data := [][]string{
		{"1", "1"},
		{"2", "2"},
		{"3", "3"},
	}
	headers := []string{"A", "B"}
	types := map[string]string{"A": "numeric", "B": "numeric"}
	corr := calculateCorrelations(data, headers, types, 3)

	if math.Abs(corr["A"]["B"]-1.0) > 1e-9 {
		t.Errorf("expected perfect correlation (1.0), got %v", corr["A"]["B"])
	}
}

func TestCalculatePearsonCorrelation_InsufficientData(t *testing.T) {
	data := [][]string{{"1", "1"}}
	r := calculatePearsonCorrelation(data, 1, 0, 1)
	if r != 0 {
		t.Errorf("expected 0 for <2 pairs, got %v", r)
	}
}

// ─── generateQualityIssues ───────────────────────────────────────────────────

func TestGenerateQualityIssues_HighMissing(t *testing.T) {
	report := &DataQualityReport{
		DataProfile: DataProfile{MissingPercent: 25},
	}
	issues := generateQualityIssues(report, nil)
	found := false
	for _, iss := range issues {
		if iss.Category == "missing" && iss.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Error("expected error-level missing issue for 25% missing data")
	}
}

func TestGenerateQualityIssues_Duplicates(t *testing.T) {
	report := &DataQualityReport{
		DataProfile: DataProfile{DuplicateRows: 5},
	}
	issues := generateQualityIssues(report, nil)
	found := false
	for _, iss := range issues {
		if iss.Category == "duplicate" {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate issue")
	}
}

// ─── generateRecommendations ─────────────────────────────────────────────────

func TestGenerateRecommendations_FewNumericCols(t *testing.T) {
	report := &DataQualityReport{
		DataProfile: DataProfile{NumericColumns: 1},
	}
	recs := generateRecommendations(report)
	found := false
	for _, r := range recs {
		if r.Category == "columns" {
			found = true
		}
	}
	if !found {
		t.Error("expected recommendation to add more numeric columns")
	}
}
