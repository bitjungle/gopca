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

package dataquality

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// generateQualityIssues inspects the analysis report and correlation map and
// returns a list of detected data quality issues.
func generateQualityIssues(report *DataQualityReport, correlations map[string]map[string]float64, sparseRows []int) []QualityIssue {
	issues := []QualityIssue{}

	// A row holding almost nothing is usually a stray line rather than a record:
	// a note, or a continuation of the row above that an export stranded on its
	// own line. It is reported and never removed, because whether it belongs to
	// the row above, should be deleted, or is a real observation with missing
	// values depends on knowing what the file was meant to say.
	if len(sparseRows) > 0 {
		issues = append(issues, QualityIssue{
			Severity:    "warning",
			Category:    "structure",
			Description: fmt.Sprintf("%d row(s) hold far fewer values than the rest of the dataset: %s", len(sparseRows), describeRowNumbers(sparseRows)),
			Affected:    []string{},
			Impact:      "A stray row is analysed as though it were an observation; check whether it belongs to the row above or should be removed",
		})
	}

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

	issues = append(issues, varianceIssues(report)...)

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

// dominantVarianceColumn returns the numeric column accounting for more than
// half the total variance, with its share, or an empty name if none does.
//
// This is the quantity that decides whether one column will run away with the
// analysis, and it needs no threshold argument beyond "more than everything else
// combined". PCA works on variance, so a column holding more than half of it
// becomes the first component almost by itself.
//
// The absolute range test elsewhere in this file cannot find such a column,
// because dominance is relative and that test is not. A category code numbered
// 1..11 has a range of 10, which is neither large nor small, yet can carry
// almost all the variance in a table of fractions; meanwhile the same test fires
// on a trace measurement with a range of 0.0001 that affects nothing.
//
// The same shape appears without any category code involved: a sample ID, a year,
// a timestamp in seconds, or one column recorded in grams beside others in
// kilograms. What they share is a spread unrelated to the measurements around
// them, and that is what this measures.
//
// Target and category columns are excluded by their type: they are already held
// out of the analysis, so their variance cannot dominate it.
func dominantVarianceColumn(report *DataQualityReport) (string, float64) {
	const dominantShare = 0.5

	total := 0.0
	type entry struct {
		name     string
		variance float64
	}
	var columns []entry
	for _, col := range report.ColumnAnalysis {
		if col.Type != "numeric" || col.Stats.StdDev == nil {
			continue
		}
		variance := *col.Stats.StdDev * *col.Stats.StdDev
		columns = append(columns, entry{col.Name, variance})
		total += variance
	}
	if total <= 0 || len(columns) < 2 {
		return "", 0
	}

	for _, col := range columns {
		if col.variance/total > dominantShare {
			return col.name, col.variance / total * 100
		}
	}
	return "", 0
}

// describeRowNumbers lists row numbers, abbreviating a long run so the message
// stays readable when a whole block of the file is affected.
func describeRowNumbers(rows []int) string {
	const maxListed = 5

	listed := rows
	suffix := ""
	if len(rows) > maxListed {
		listed = rows[:maxListed]
		suffix = fmt.Sprintf(" and %d more", len(rows)-maxListed)
	}
	parts := make([]string, len(listed))
	for i, row := range listed {
		parts[i] = strconv.Itoa(row)
	}
	return strings.Join(parts, ", ") + suffix
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

	// One column carrying most of the variance is worth naming on its own,
	// whatever the reason -- an identifier, a timestamp, a column in the wrong
	// unit, a category code stored as an integer. The absolute test below cannot
	// find any of them, because dominance is relative and that test is not (#909).
	if name, share := dominantVarianceColumn(report); name != "" {
		recs = append(recs, Recommendation{
			Priority: "high",
			Category: "scaling",
			Action:   fmt.Sprintf("Check whether '%s' belongs in the analysis", name),
			Description: fmt.Sprintf(
				"'%s' accounts for %.2f%% of the total variance across all numeric columns. "+
					"Without scaling it will dominate every component. If it is a category code "+
					"rather than a measurement, mark it #category to hold it out.", name, share),
			Columns: []string{name},
		})
	}

	// Columns whose absolute range is extreme. A different situation from the
	// above -- this catches a table mixing millimetres with kilometres, where no
	// single column dominates but the units disagree.
	var oddScale []string
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.Min != nil && col.Stats.Max != nil {
			rangeVal := *col.Stats.Max - *col.Stats.Min
			if rangeVal > 1000 || rangeVal < 0.01 {
				oddScale = append(oddScale, col.Name)
			}
		}
	}
	if len(oddScale) > 0 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "scaling",
			Action:      "Scale numeric columns",
			Description: "Columns have varying scales; consider standardization or normalization before PCA",
			Columns:     oddScale,
		})
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

// nearConstantCV is the coefficient of variation below which a column is
// reported as carrying almost no information: a standard deviation smaller
// than a thousandth of the column's own level.
//
// Stated as a constant and named in the message, because a threshold the user
// cannot see is one they cannot argue with.
const nearConstantCV = 0.001

// varianceIssues reports columns that carry little or no variation.
//
// This replaces an absolute test, StdDev < 0.01, which was scale-dependent:
// the same measurements expressed in kilometres and in metres gave different
// answers, so it flagged perfectly good data recorded in large units and
// stayed quiet about degenerate data recorded in small ones. A threshold on a
// dimensional quantity cannot mean anything without knowing the unit, and the
// software does not.
//
// Two separate things are reported, because they are known with different
// confidence:
//
//	constant       every value identical. No threshold is involved and no
//	               judgement: the column contributes exactly nothing to any
//	               component.
//	near-constant  the spread is a vanishing fraction of the column's own
//	               level, judged by the coefficient of variation, which is
//	               dimensionless and so says the same thing whatever the unit.
//
// Neither is removed. Whether a low-variance variable matters is the user's
// judgement, and silently dropping columns is the failure #801 was about.
func varianceIssues(report *DataQualityReport) []QualityIssue {
	var constant []string
	var issues []QualityIssue

	for _, col := range report.ColumnAnalysis {
		if isConstantColumn(col) {
			constant = append(constant, col.Name)
			continue
		}

		// The coefficient of variation is undefined at a mean of zero and
		// unstable near it. A column centred on zero is not judged rather than
		// judged badly -- saying nothing is better than a number that means
		// nothing.
		if col.Type != "numeric" || col.Stats.StdDev == nil || col.Stats.Mean == nil {
			continue
		}
		mean := math.Abs(*col.Stats.Mean)
		if mean == 0 || math.IsNaN(mean) || math.IsInf(mean, 0) {
			continue
		}
		cv := *col.Stats.StdDev / mean
		if cv < nearConstantCV {
			issues = append(issues, QualityIssue{
				Severity: "info",
				Category: "variance",
				Description: fmt.Sprintf(
					"Column '%s' varies by %.4f%% of its own level (σ=%g, mean=%g)",
					col.Name, cv*100, *col.Stats.StdDev, *col.Stats.Mean),
				Affected: []string{col.Name},
				Impact: fmt.Sprintf(
					"Below %.1f%% this is close to constant. Standardization scales it to "+
						"unit variance regardless, which can turn measurement noise into an "+
						"apparent component", nearConstantCV*100),
			})
		}
	}

	if len(constant) > 0 {
		issues = append(issues, QualityIssue{
			Severity: "warning",
			Category: "variance",
			Description: fmt.Sprintf("%s no variation at all",
				columnsPhrase(constant, "has", "have")),
			Affected: constant,
			Impact: "A constant column contributes nothing to any component. It sits at " +
				"the origin of every loadings plot, where its position can be read as " +
				"meaningful rather than as an artefact of having no variance",
		})
	}

	return issues
}

// isConstantColumn reports whether every present value in a column is the same.
//
// Numeric columns are judged on min == max rather than on a standard deviation
// compared against zero, which avoids asking whether a floating-point result is
// exactly zero. Categorical columns are judged on the distinct-value count.
func isConstantColumn(col ColumnAnalysis) bool {
	if col.Stats.Count == 0 {
		// A column with nothing in it is empty, not constant, and the missing
		// data checks already have something to say about it.
		return false
	}
	if col.Type == "numeric" {
		if col.Stats.Min != nil && col.Stats.Max != nil {
			return *col.Stats.Min == *col.Stats.Max
		}
		return false
	}
	return col.Stats.Unique == 1
}

// columnsPhrase renders a list of column names with an agreeing verb.
func columnsPhrase(names []string, singular, plural string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = fmt.Sprintf("'%s'", name)
	}
	if len(names) == 1 {
		return fmt.Sprintf("Column %s %s", quoted[0], singular)
	}
	return fmt.Sprintf("Columns %s %s", strings.Join(quoted, ", "), plural)
}
