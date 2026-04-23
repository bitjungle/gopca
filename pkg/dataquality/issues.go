// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package dataquality

import (
	"fmt"
	"math"
)

// generateQualityIssues inspects the analysis report and correlation map and
// returns a list of detected data quality issues.
func generateQualityIssues(report *DataQualityReport, correlations map[string]map[string]float64) []QualityIssue {
	issues := []QualityIssue{}

	// Dataset-level missing data
	switch {
	case report.DataProfile.MissingPercent > 20:
		issues = append(issues, QualityIssue{
			Severity:    "error",
			Category:    "missing",
			Description: fmt.Sprintf("Dataset has %.1f%% missing values", report.DataProfile.MissingPercent),
			Impact:      "High missing data can significantly affect PCA results",
		})
	case report.DataProfile.MissingPercent > 10:
		issues = append(issues, QualityIssue{
			Severity:    "warning",
			Category:    "missing",
			Description: fmt.Sprintf("Dataset has %.1f%% missing values", report.DataProfile.MissingPercent),
			Impact:      "Missing data may affect PCA results",
		})
	}

	// Column-level missing data
	for _, col := range report.ColumnAnalysis {
		if col.Stats.MissingPercent > 50 {
			issues = append(issues, QualityIssue{
				Severity:    "error",
				Category:    "missing",
				Description: fmt.Sprintf("Column '%s' has %.1f%% missing values", col.Name, col.Stats.MissingPercent),
				Affected:    []string{col.Name},
				Impact:      "Columns with >50% missing data should be removed",
			})
		}
	}

	// Duplicate rows
	if report.DataProfile.DuplicateRows > 0 {
		issues = append(issues, QualityIssue{
			Severity:    "warning",
			Category:    "duplicate",
			Description: fmt.Sprintf("Found %d duplicate rows", report.DataProfile.DuplicateRows),
			Impact:      "Duplicate rows can bias PCA results",
		})
	}

	// Outliers
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && len(col.Outliers) > 0 {
			outlierPct := float64(len(col.Outliers)) / float64(col.Stats.Count) * 100
			if outlierPct > 10 {
				issues = append(issues, QualityIssue{
					Severity:    "warning",
					Category:    "outlier",
					Description: fmt.Sprintf("Column '%s' has %d outliers (%.1f%%)", col.Name, len(col.Outliers), outlierPct),
					Affected:    []string{col.Name},
					Impact:      "Outliers can disproportionately influence PCA components",
				})
			}
		}
	}

	// High pairwise correlations
	for col1, corrMap := range correlations {
		for col2, corr := range corrMap {
			if col1 < col2 && math.Abs(corr) > 0.95 {
				issues = append(issues, QualityIssue{
					Severity:    "warning",
					Category:    "correlation",
					Description: fmt.Sprintf("Columns '%s' and '%s' are highly correlated (r=%.3f)", col1, col2, corr),
					Affected:    []string{col1, col2},
					Impact:      "Highly correlated variables provide redundant information in PCA",
				})
			}
		}
	}

	// Low variance
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.StdDev != nil && *col.Stats.StdDev < 0.01 {
			issues = append(issues, QualityIssue{
				Severity:    "info",
				Category:    "variance",
				Description: fmt.Sprintf("Column '%s' has very low variance (σ=%.4f)", col.Name, *col.Stats.StdDev),
				Affected:    []string{col.Name},
				Impact:      "Low variance columns contribute little to PCA",
			})
		}
	}

	// Non-normal distributions
	nonNormalCount := 0
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && !col.Distribution.IsNormal {
			nonNormalCount++
		}
	}
	if nonNormalCount > 0 {
		issues = append(issues, QualityIssue{
			Severity:    "info",
			Category:    "distribution",
			Description: fmt.Sprintf("%d numeric columns have non-normal distributions", nonNormalCount),
			Impact:      "PCA assumes normality; consider data transformations",
		})
	}

	return issues
}

// generateRecommendations returns prioritised, actionable recommendations
// derived from the quality report.
func generateRecommendations(report *DataQualityReport) []Recommendation {
	recs := []Recommendation{}

	if report.DataProfile.MissingPercent > 10 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "missing",
			Action:      "Handle missing values",
			Description: "Use appropriate fill strategies (mean/median for numeric, mode for categorical) or remove rows/columns with excessive missing data",
		})
	}

	if report.DataProfile.DuplicateRows > 0 {
		recs = append(recs, Recommendation{
			Priority:    "medium",
			Category:    "duplicate",
			Action:      "Remove duplicate rows",
			Description: fmt.Sprintf("Remove %d duplicate rows to avoid biasing the analysis", report.DataProfile.DuplicateRows),
		})
	}

	colsWithOutliers := []string{}
	for _, col := range report.ColumnAnalysis {
		if len(col.Outliers) > 5 {
			colsWithOutliers = append(colsWithOutliers, col.Name)
		}
	}
	if len(colsWithOutliers) > 0 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "outlier",
			Action:      "Handle outliers",
			Description: "Consider removing or transforming outliers, or use robust scaling",
			Columns:     colsWithOutliers,
		})
	}

	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.Min != nil && col.Stats.Max != nil {
			rangeVal := *col.Stats.Max - *col.Stats.Min
			if rangeVal > 1000 || rangeVal < 0.01 {
				recs = append(recs, Recommendation{
					Priority:    "high",
					Category:    "scaling",
					Action:      "Scale numeric columns",
					Description: "Columns have varying scales; consider standardization or normalization before PCA",
				})
				break
			}
		}
	}

	skewedCols := []string{}
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.Skewness != nil && math.Abs(*col.Stats.Skewness) > 1.0 {
			skewedCols = append(skewedCols, col.Name)
		}
	}
	if len(skewedCols) > 0 {
		recs = append(recs, Recommendation{
			Priority:    "medium",
			Category:    "distribution",
			Action:      "Transform skewed distributions",
			Description: "Consider log or square root transformations for highly skewed columns",
			Columns:     skewedCols,
		})
	}

	if report.DataProfile.NumericColumns < 3 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "columns",
			Action:      "Add more numeric columns",
			Description: fmt.Sprintf("Only %d numeric columns available; PCA requires multiple numeric features", report.DataProfile.NumericColumns),
		})
	}

	return recs
}
