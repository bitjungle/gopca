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
	"sort"
	"strconv"
	"strings"
)

// AnalyzeDataQuality performs comprehensive data quality analysis on the given
// input and returns a DataQualityReport. Returns an error if the input is empty.
func AnalyzeDataQuality(in AnalysisInput) (*DataQualityReport, error) {
	if in.Rows == 0 || in.Columns == 0 || len(in.Data) == 0 {
		return nil, fmt.Errorf("no data to analyze")
	}

	report := &DataQualityReport{
		DataProfile: DataProfile{
			Rows:    in.Rows,
			Columns: in.Columns,
		},
		ColumnAnalysis:  make([]ColumnAnalysis, 0, in.Columns),
		Issues:          []QualityIssue{},
		Recommendations: []Recommendation{},
	}

	// Count column types
	for _, colType := range in.ColumnTypes {
		switch colType {
		case "numeric":
			report.DataProfile.NumericColumns++
		case "categorical":
			report.DataProfile.CategoricalColumns++
		case "target":
			report.DataProfile.TargetColumns++
		}
	}

	// Calculate missing data percentage
	missingStats := AnalyzeMissing(in.Data, in.Headers)
	report.DataProfile.MissingPercent = missingStats.MissingPercent

	// Detect duplicate rows
	report.DataProfile.DuplicateRows = countDuplicateRows(in.Data, in.Rows)

	// Estimate memory size
	report.DataProfile.MemorySize = estimateMemorySize(in.Rows, in.Columns)

	// Analyze each column
	for colIdx, header := range in.Headers {
		colAnalysis := analyzeColumn(in, colIdx, header)
		report.ColumnAnalysis = append(report.ColumnAnalysis, colAnalysis)
	}

	// Calculate correlations for numeric columns
	correlations := calculateCorrelations(in.Data, in.Headers, in.ColumnTypes, in.Rows)

	// Generate issues based on analysis
	report.Issues = generateQualityIssues(report, correlations)

	// Generate recommendations
	report.Recommendations = generateRecommendations(report)

	// Calculate overall quality score
	report.QualityScore = calculateQualityScore(report)

	return report, nil
}

// analyzeColumn performs detailed statistical analysis on a single column.
func analyzeColumn(in AnalysisInput, colIdx int, header string) ColumnAnalysis {
	analysis := ColumnAnalysis{
		Name: header,
		Type: "numeric",
	}

	if in.ColumnTypes != nil {
		if colType, exists := in.ColumnTypes[header]; exists {
			analysis.Type = colType
		}
	}

	if analysis.Type == "numeric" {
		analysis.Stats = analyzeNumericStats(in.Data, in.Rows, colIdx)
		analysis.Distribution = analyzeDistribution(in.Data, in.Rows, colIdx)
		analysis.Outliers = detectOutliers(in.Data, in.Rows, colIdx, analysis.Stats)
	} else {
		analysis.Stats = analyzeCategoricalStats(in.Data, in.Rows, colIdx)
	}

	analysis.QualityScore = calculateColumnQualityScore(analysis)

	return analysis
}

// analyzeNumericStats computes descriptive statistics for a numeric column.
func analyzeNumericStats(data [][]string, rows, colIdx int) ColumnStatistics {
	stats := ColumnStatistics{Count: rows}

	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		if colIdx < len(data[rowIdx]) {
			if isMissing(strings.TrimSpace(data[rowIdx][colIdx])) {
				stats.Missing++
			}
		}
	}

	if stats.Count > 0 {
		stats.MissingPercent = float64(stats.Missing) / float64(stats.Count) * 100
	}

	values := getNumericColumn(data, colIdx)
	if len(values) == 0 {
		return stats
	}

	sort.Float64s(values)

	stats.Unique = countUnique(values)
	mean := calculateMean(values)
	stats.Mean = &mean
	median := calculateMedian(values)
	stats.Median = &median
	stdDev := calculateStdDev(values, mean)
	stats.StdDev = &stdDev
	min := values[0]
	stats.Min = &min
	max := values[len(values)-1]
	stats.Max = &max
	q1 := calculatePercentile(values, 25)
	stats.Q1 = &q1
	q3 := calculatePercentile(values, 75)
	stats.Q3 = &q3
	iqr := q3 - q1
	stats.IQR = &iqr
	skewness := calculateSkewness(values, mean, stdDev)
	stats.Skewness = &skewness
	kurtosis := calculateKurtosis(values, mean, stdDev)
	stats.Kurtosis = &kurtosis

	return stats
}

// analyzeCategoricalStats computes frequency statistics for a categorical column.
func analyzeCategoricalStats(data [][]string, rows, colIdx int) ColumnStatistics {
	stats := ColumnStatistics{
		Count:      rows,
		Categories: make(map[string]int),
	}

	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		if colIdx < len(data[rowIdx]) {
			value := strings.TrimSpace(data[rowIdx][colIdx])
			if isMissing(value) {
				stats.Missing++
			} else {
				stats.Categories[value]++
			}
		}
	}

	if stats.Count > 0 {
		stats.MissingPercent = float64(stats.Missing) / float64(stats.Count) * 100
	}

	stats.Unique = len(stats.Categories)

	if len(stats.Categories) > 0 {
		maxCount := 0
		mode := ""
		for value, count := range stats.Categories {
			if count > maxCount {
				maxCount = count
				mode = value
			}
		}
		stats.Mode = &mode
	}

	return stats
}

// analyzeDistribution builds a histogram and classifies the distribution shape
// for a numeric column. Returns an empty DistributionInfo if fewer than 10
// non-missing values are present.
func analyzeDistribution(data [][]string, rows, colIdx int) DistributionInfo {
	dist := DistributionInfo{}

	values := make([]float64, 0, rows)
	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		if colIdx < len(data[rowIdx]) {
			value := strings.TrimSpace(data[rowIdx][colIdx])
			if !isMissing(value) {
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					values = append(values, num)
				}
			}
		}
	}

	if len(values) < 10 {
		return dist
	}

	sort.Float64s(values)
	minVal, maxVal := values[0], values[len(values)-1]
	binWidth := (maxVal - minVal) / 10

	if binWidth > 0 {
		dist.Histogram = make([]HistogramBin, 10)
		for i := 0; i < 10; i++ {
			binMin := minVal + float64(i)*binWidth
			binMax := binMin + binWidth
			if i == 9 {
				binMax = maxVal + 0.001 // Include the maximum value in the last bin
			}
			dist.Histogram[i] = HistogramBin{Min: binMin, Max: binMax}
		}

		for _, v := range values {
			idx := int((v - minVal) / binWidth)
			if idx >= 10 {
				idx = 9
			}
			dist.Histogram[idx].Count++
		}
	}

	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)
	skewness := calculateSkewness(values, mean, stdDev)
	kurtosis := calculateKurtosis(values, mean, stdDev)

	dist.IsNormal = math.Abs(skewness) < 0.5 && math.Abs(kurtosis) < 1.0

	switch {
	case dist.IsNormal:
		dist.DistType = "normal"
	case math.Abs(skewness) > 1.0:
		if skewness > 0 {
			dist.DistType = "right-skewed"
		} else {
			dist.DistType = "left-skewed"
		}
	case len(dist.Histogram) > 0:
		peaks := 0
		for i := 1; i < len(dist.Histogram)-1; i++ {
			if dist.Histogram[i].Count > dist.Histogram[i-1].Count &&
				dist.Histogram[i].Count > dist.Histogram[i+1].Count {
				peaks++
			}
		}
		if peaks >= 2 {
			dist.DistType = "bimodal"
		} else {
			dist.DistType = "unknown"
		}
	}

	return dist
}

// detectOutliers returns outliers found in a numeric column using both the IQR
// method (1.5×IQR fence) and the Z-score method (threshold = 3.0). A row is
// only reported once even if it triggers both methods.
func detectOutliers(data [][]string, rows, colIdx int, stats ColumnStatistics) []OutlierInfo {
	outliers := []OutlierInfo{}

	if stats.Q1 == nil || stats.Q3 == nil || stats.Mean == nil || stats.StdDev == nil {
		return outliers
	}

	iqrPositive := stats.IQR != nil && *stats.IQR > 0
	var lowerBound, upperBound float64
	if iqrPositive {
		lowerBound = *stats.Q1 - 1.5*(*stats.IQR)
		upperBound = *stats.Q3 + 1.5*(*stats.IQR)
	}
	zThreshold := 3.0

	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		if colIdx >= len(data[rowIdx]) {
			continue
		}
		value := strings.TrimSpace(data[rowIdx][colIdx])
		if isMissing(value) {
			continue
		}
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}

		if iqrPositive && (num < lowerBound || num > upperBound) {
			outliers = append(outliers, OutlierInfo{
				RowIndex: rowIdx,
				Value:    value,
				Method:   "iqr",
				Score:    math.Abs(num-*stats.Median) / *stats.IQR,
			})
			continue
		}

		if *stats.StdDev > 0 {
			zScore := math.Abs(num-*stats.Mean) / *stats.StdDev
			if zScore > zThreshold {
				outliers = append(outliers, OutlierInfo{
					RowIndex: rowIdx,
					Value:    value,
					Method:   "zscore",
					Score:    zScore,
				})
			}
		}
	}

	return outliers
}

// countDuplicateRows returns the number of rows that appear more than once.
func countDuplicateRows(data [][]string, rows int) int {
	rowMap := make(map[string]int)
	duplicates := 0

	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		key := strings.Join(data[rowIdx], "|")
		rowMap[key]++
		if rowMap[key] >= 2 {
			duplicates++
		}
	}

	return duplicates
}

// estimateMemorySize returns a human-readable estimate of the dataset's
// in-memory footprint assuming ~10 bytes per cell.
func estimateMemorySize(rows, cols int) string {
	bytes := rows * cols * 10
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}

// calculateQualityScore computes an overall quality score (0–100) for the
// dataset, deducting points for missing values, duplicates, outliers, and
// insufficient numeric columns.
func calculateQualityScore(report *DataQualityReport) float64 {
	score := 100.0

	score -= report.DataProfile.MissingPercent * 0.5

	if report.DataProfile.Rows > 0 {
		dupPct := float64(report.DataProfile.DuplicateRows) / float64(report.DataProfile.Rows) * 100
		score -= dupPct * 0.3
	}

	for _, col := range report.ColumnAnalysis {
		if col.Stats.MissingPercent > 50 {
			score -= 5.0
		}
	}

	totalOutliers := 0
	for _, col := range report.ColumnAnalysis {
		totalOutliers += len(col.Outliers)
	}
	if report.DataProfile.Rows > 0 && report.DataProfile.NumericColumns > 0 {
		outlierPct := float64(totalOutliers) / float64(report.DataProfile.Rows*report.DataProfile.NumericColumns) * 100
		score -= outlierPct * 0.2
	}

	if report.DataProfile.NumericColumns < 3 {
		score -= 20.0
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateColumnQualityScore computes a quality score (0–100) for a single
// column based on its missing-value rate, outlier rate, and variance.
func calculateColumnQualityScore(analysis ColumnAnalysis) float64 {
	score := 100.0

	score -= analysis.Stats.MissingPercent * 0.5

	if analysis.Stats.Count > 0 {
		outlierPct := float64(len(analysis.Outliers)) / float64(analysis.Stats.Count) * 100
		score -= outlierPct * 0.3
	}

	if analysis.Type == "numeric" && analysis.Stats.StdDev != nil && *analysis.Stats.StdDev < 0.01 {
		score -= 10.0
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
