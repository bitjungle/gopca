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
	"math"
	"strconv"
	"strings"
)

// calculateCorrelations returns a map of Pearson correlation coefficients for
// all pairs of numeric columns. Self-correlations are set to 1.0.
func calculateCorrelations(data [][]string, headers []string, columnTypes map[string]string, rows int) map[string]map[string]float64 {
	correlations := make(map[string]map[string]float64)

	numericCols := make([]int, 0)
	numericHeaders := make([]string, 0)
	for i, header := range headers {
		if colType, exists := columnTypes[header]; exists && colType == "numeric" {
			numericCols = append(numericCols, i)
			numericHeaders = append(numericHeaders, header)
		}
	}

	for i, col1 := range numericCols {
		h1 := numericHeaders[i]
		if _, exists := correlations[h1]; !exists {
			correlations[h1] = make(map[string]float64)
		}

		for j, col2 := range numericCols {
			h2 := numericHeaders[j]
			if i == j {
				correlations[h1][h2] = 1.0
			} else {
				correlations[h1][h2] = calculatePearsonCorrelation(data, rows, col1, col2)
			}
		}
	}

	return correlations
}

// calculatePearsonCorrelation returns the Pearson r between two columns, using
// only rows where both values are non-missing and numeric.
// Returns 0 when fewer than two paired observations are available.
func calculatePearsonCorrelation(data [][]string, rows, col1, col2 int) float64 {
	type pair struct{ x, y float64 }
	pairs := make([]pair, 0, rows)

	for rowIdx := 0; rowIdx < rows && rowIdx < len(data); rowIdx++ {
		row := data[rowIdx]
		if col1 >= len(row) || col2 >= len(row) {
			continue
		}
		v1 := strings.TrimSpace(row[col1])
		v2 := strings.TrimSpace(row[col2])
		if isMissing(v1) || isMissing(v2) {
			continue
		}
		n1, err1 := strconv.ParseFloat(v1, 64)
		n2, err2 := strconv.ParseFloat(v2, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		pairs = append(pairs, pair{n1, n2})
	}

	if len(pairs) < 2 {
		return 0
	}

	mean1, mean2 := 0.0, 0.0
	for _, p := range pairs {
		mean1 += p.x
		mean2 += p.y
	}
	n := float64(len(pairs))
	mean1 /= n
	mean2 /= n

	num, den1, den2 := 0.0, 0.0, 0.0
	for _, p := range pairs {
		d1, d2 := p.x-mean1, p.y-mean2
		num += d1 * d2
		den1 += d1 * d1
		den2 += d2 * d2
	}

	if den1 == 0 || den2 == 0 {
		return 0
	}

	return num / math.Sqrt(den1*den2)
}
