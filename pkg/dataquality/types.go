// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

// AnalysisInput carries all data needed for quality analysis. Callers
// populate this from their own domain type (e.g. FileData) so that this
// package remains independent of UI/Wails types.
type AnalysisInput struct {
	Data        [][]string
	Headers     []string
	ColumnTypes map[string]string // "numeric", "categorical", "target"
	RowNames    []string
	Rows        int
	Columns     int
}

// FillRequest describes a missing-value fill operation.
type FillRequest struct {
	Strategy string // "mean", "median", "mode", "forward", "backward", "custom"
	Column   string // Column name, or empty string to process all columns
	Value    string // Custom fill value (used when Strategy == "custom")
}

// ─── Missing value types ────────────────────────────────────────────────────

// MissingValueStats contains missing-value statistics for an entire dataset.
type MissingValueStats struct {
	TotalCells     int                       `json:"totalCells"`
	MissingCells   int                       `json:"missingCells"`
	MissingPercent float64                   `json:"missingPercent"`
	ColumnStats    map[string]*ColumnMissing `json:"columnStats"`
	RowStats       map[int]*RowMissing       `json:"rowStats"`
}

// ColumnMissing contains missing-value statistics for one column.
type ColumnMissing struct {
	Name           string  `json:"name"`
	TotalValues    int     `json:"totalValues"`
	MissingValues  int     `json:"missingValues"`
	MissingPercent float64 `json:"missingPercent"`
	Pattern        string  `json:"pattern"` // "none", "random", "systematic", "top", "bottom"
}

// RowMissing contains missing-value statistics for one row.
type RowMissing struct {
	Index          int     `json:"index"`
	TotalValues    int     `json:"totalValues"`
	MissingValues  int     `json:"missingValues"`
	MissingPercent float64 `json:"missingPercent"`
}

// ─── Data quality types ─────────────────────────────────────────────────────

// DataQualityReport is the top-level result of a full data quality analysis.
type DataQualityReport struct {
	DataProfile     DataProfile      `json:"dataProfile"`
	ColumnAnalysis  []ColumnAnalysis `json:"columnAnalysis"`
	QualityScore    float64          `json:"qualityScore"`
	Issues          []QualityIssue   `json:"issues"`
	Recommendations []Recommendation `json:"recommendations"`
}

// DataProfile contains overall dataset-level statistics.
type DataProfile struct {
	Rows               int     `json:"rows"`
	Columns            int     `json:"columns"`
	NumericColumns     int     `json:"numericColumns"`
	CategoricalColumns int     `json:"categoricalColumns"`
	TargetColumns      int     `json:"targetColumns"`
	MissingPercent     float64 `json:"missingPercent"`
	DuplicateRows      int     `json:"duplicateRows"`
	MemorySize         string  `json:"memorySize"`
}

// ColumnAnalysis contains detailed analysis for a single column.
type ColumnAnalysis struct {
	Name         string           `json:"name"`
	Type         string           `json:"type"` // "numeric", "categorical", "target"
	Stats        ColumnStatistics `json:"stats"`
	Distribution DistributionInfo `json:"distribution"`
	Outliers     []OutlierInfo    `json:"outliers"`
	QualityScore float64          `json:"qualityScore"`
}

// ColumnStatistics contains statistical measures for a column.
type ColumnStatistics struct {
	Count          int            `json:"count"`
	Missing        int            `json:"missing"`
	MissingPercent float64        `json:"missingPercent"`
	Unique         int            `json:"unique"`
	Mean           *float64       `json:"mean,omitempty"`
	Median         *float64       `json:"median,omitempty"`
	Mode           *string        `json:"mode,omitempty"`
	StdDev         *float64       `json:"stdDev,omitempty"`
	Min            *float64       `json:"min,omitempty"`
	Max            *float64       `json:"max,omitempty"`
	Q1             *float64       `json:"q1,omitempty"`
	Q3             *float64       `json:"q3,omitempty"`
	IQR            *float64       `json:"iqr,omitempty"`
	Skewness       *float64       `json:"skewness,omitempty"`
	Kurtosis       *float64       `json:"kurtosis,omitempty"`
	Categories     map[string]int `json:"categories,omitempty"`
}

// DistributionInfo describes the distribution shape of a numeric column.
type DistributionInfo struct {
	Histogram       []HistogramBin `json:"histogram,omitempty"`
	IsNormal        bool           `json:"isNormal"`
	NormalityPValue float64        `json:"normalityPValue,omitempty"`
	DistType        string         `json:"distType"` // "normal", "right-skewed", "left-skewed", "bimodal", "unknown"
}

// HistogramBin represents one bin in a histogram.
type HistogramBin struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int     `json:"count"`
}

// OutlierInfo describes one detected outlier value.
type OutlierInfo struct {
	RowIndex int     `json:"rowIndex"`
	Value    string  `json:"value"`
	Method   string  `json:"method"` // "iqr" or "zscore"
	Score    float64 `json:"score"`
}

// QualityIssue describes a detected data quality problem.
type QualityIssue struct {
	Severity    string   `json:"severity"` // "error", "warning", "info"
	Category    string   `json:"category"` // "missing", "outlier", "duplicate", "correlation", "variance", "distribution"
	Description string   `json:"description"`
	Affected    []string `json:"affected"`
	Impact      string   `json:"impact"`
}

// Recommendation is an actionable suggestion derived from the quality analysis.
type Recommendation struct {
	Priority    string   `json:"priority"` // "high", "medium", "low"
	Category    string   `json:"category"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Columns     []string `json:"columns,omitempty"`
}
