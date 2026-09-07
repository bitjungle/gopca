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
	"fmt"
	"math"
	"strconv"
	"strings"
)

// closureTolerance is the relative spread of row sums below which a set of
// columns is reported as looking closed -- percentages summing to 100, or
// proportions summing to 1.
//
// Only used to describe the data back to the user. CLR does not require
// closure: a subcomposition, a subset of the parts, is still compositional and
// is analysed this way routinely.
const closureTolerance = 0.01

// applyCLR applies the centred log-ratio transform to a set of columns treated
// together as one composition.
//
//	clr(x)_i = ln( x_i / g(x) )     where g(x) is the geometric mean of the row
//
// which is computed as ln(x_i) - mean(ln(x)), since that is numerically better
// behaved than forming the product first: the geometric mean of forty parts
// each around 0.02 underflows, and the logarithms do not.
//
// # Why this exists
//
// Compositional data carries a constant-sum constraint. Percentages, mineral
// assays and food composition describe proportions of a whole, so if one part
// rises the others must fall whatever the underlying chemistry. That makes the
// covariance matrix singular and forces spurious negative correlations between
// parts, so principal components describe the closure as much as the samples.
// PCA on raw closed data is a textbook error rather than a subtle one.
//
// The log-ratio transform removes the constraint by describing each part
// relative to the geometric mean of the whole, which is what Aitchison's
// framework does. CLR is chosen over its alternatives because it keeps one
// output column per input column: loadings stay interpretable as variables,
// which ILR coordinates are not, and it needs no arbitrary choice of
// denominator, which ALR does.
//
// # References
//
//	Aitchison, J. (1986). The Statistical Analysis of Compositional Data.
//	Chapman & Hall. Chapter 4.
//	Egozcue et al. (2003). Isometric Logratio Transformations for Compositional
//	Data Analysis. Mathematical Geology 35(3), for the ILR comparison.
//
// Algorithm complexity: O(n·p) for n rows and p columns.
func applyCLR(data [][]string, columnTypes map[string]string, catCols map[string][]string, headers *[]string, opts Options, result *Result) error {
	if len(opts.Columns) < 2 {
		return fmt.Errorf("a composition needs at least two parts, got %d", len(opts.Columns))
	}

	indices := make([]int, 0, len(opts.Columns))
	seen := make(map[string]bool, len(opts.Columns))
	for _, colName := range opts.Columns {
		if seen[colName] {
			return fmt.Errorf("column %q is listed more than once; each part can appear "+
				"in a composition only once", colName)
		}
		seen[colName] = true

		colIndex := findColumn(*headers, colName)
		if colIndex == -1 {
			return fmt.Errorf("column %q not found", colName)
		}
		if columnTypes[colName] != "numeric" {
			return fmt.Errorf("column %q is not numeric; a composition is made of "+
				"measured parts", colName)
		}
		indices = append(indices, colIndex)
	}

	// Read the whole composition first. Nothing is written until every row has
	// been checked, so a refusal leaves the data exactly as it was.
	parts := make([][]float64, len(data))
	for i := range data {
		row := make([]float64, len(indices))
		for j, colIndex := range indices {
			value := ""
			if colIndex < len(data[i]) {
				value = strings.TrimSpace(data[i][colIndex])
			}
			if value == "" {
				// A missing part is not a zero part. Guessing which it is would
				// be the same fabrication the zero rule exists to prevent.
				return fmt.Errorf("row %d has no value for %q; a composition must be "+
					"complete before it can be transformed", i+1, opts.Columns[j])
			}
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("row %d, column %q: %q is not a number", i+1, opts.Columns[j], value)
			}
			if number < 0 {
				return fmt.Errorf("row %d, column %q is negative (%g). A composition "+
					"describes parts of a whole, so a negative part has no meaning and "+
					"no substitution can repair it", i+1, opts.Columns[j], number)
			}
			row[j] = number
		}
		parts[i] = row
	}

	if err := replaceZeros(parts, opts, result); err != nil {
		return err
	}

	// Describe the data back before changing it, so the user can see whether
	// what they selected behaves like a composition at all.
	result.Messages = append(result.Messages, describeClosure(parts, opts.Columns))

	transformed := make([][]float64, len(parts))
	for i, row := range parts {
		logs := make([]float64, len(row))
		mean := 0.0
		for j, value := range row {
			logs[j] = math.Log(value)
			mean += logs[j]
		}
		mean /= float64(len(row))

		out := make([]float64, len(row))
		for j := range row {
			out[j] = logs[j] - mean
		}
		transformed[i] = out
	}

	newColumns := make([]string, 0, len(indices))
	for j, colName := range opts.Columns {
		newColName := uniqueColumnName(*headers, colName+"_clr")
		*headers = append(*headers, newColName)
		columnTypes[newColName] = "numeric"
		newColumns = append(newColumns, newColName)

		for i := range data {
			data[i] = append(data[i], formatCLR(transformed[i][j]))
		}
	}

	if opts.RemoveOriginal {
		// Highest index first, so the indices still to be removed stay valid
		// as the columns beneath them disappear.
		sorted := append([]int(nil), indices...)
		for a := 0; a < len(sorted); a++ {
			for b := a + 1; b < len(sorted); b++ {
				if sorted[b] > sorted[a] {
					sorted[a], sorted[b] = sorted[b], sorted[a]
				}
			}
		}
		for _, colIndex := range sorted {
			removeColumn(data, columnTypes, catCols, headers, (*headers)[colIndex], colIndex)
		}
	}

	result.TransformedColumns = append(result.TransformedColumns, opts.Columns...)
	result.NewColumns = append(result.NewColumns, newColumns...)
	result.Messages = append(result.Messages, fmt.Sprintf(
		"Centred log-ratio applied to %d parts: %s",
		len(newColumns), strings.Join(newColumns, ", ")))

	return nil
}

// replaceZeros substitutes zeros using multiplicative replacement, or refuses.
//
// The logarithm is undefined at zero, so a composition containing zeros cannot
// be transformed as it stands. There is a standard remedy -- replace each zero
// with a small value below the detection limit and scale the remaining parts so
// the row still sums to what it did -- but it invents a measurement that was
// not made, and the difference between "absent" and "below the detection limit"
// is a scientific judgement the software has no basis for.
//
// So the default is refusal, naming the rows. The replacement happens only when
// a value is supplied, and is reported when it does.
//
// Reference: Martín-Fernández, Barceló-Vidal & Pawlowsky-Glahn (2003),
// Mathematical Geology 35(3), 253-278.
func replaceZeros(parts [][]float64, opts Options, result *Result) error {
	var affectedRows []int
	zeroCount := 0
	for i, row := range parts {
		rowHasZero := false
		for _, value := range row {
			if value == 0 {
				zeroCount++
				rowHasZero = true
			}
		}
		if rowHasZero {
			affectedRows = append(affectedRows, i+1)
		}
	}
	if zeroCount == 0 {
		return nil
	}

	if opts.ZeroReplacement <= 0 {
		return fmt.Errorf("%d zero value(s) in %s. The logarithm is undefined at zero, "+
			"so the transform cannot proceed. Supply a replacement value below your "+
			"detection limit to substitute them, which is a judgement about what a "+
			"zero means in your data and so has to be yours",
			zeroCount, describeRows(affectedRows))
	}

	for i, row := range parts {
		zeros := 0
		total := 0.0
		for _, value := range row {
			if value == 0 {
				zeros++
			}
			total += value
		}
		if zeros == 0 {
			continue
		}

		// A row that is entirely zeros has no composition to preserve, and
		// substituting every part would invent one. Worse, it would invent a
		// convincing one: replacing every part with the same value gives a
		// clr of all zeros, which is exactly what a perfectly equal
		// composition looks like. A row where nothing was measured would
		// arrive in the analysis indistinguishable from a real sample.
		if total == 0 {
			return fmt.Errorf("row %d has no measured parts at all. Multiplicative "+
				"replacement rescales the parts that were measured, and there are "+
				"none here, so there is nothing to preserve -- remove the row rather "+
				"than substituting a composition for it", i+1)
		}

		if opts.ZeroReplacement*float64(zeros) >= total {
			return fmt.Errorf("the replacement value %g is too large: %d zeros would "+
				"account for the whole of a row summing to %g",
				opts.ZeroReplacement, zeros, total)
		}

		// Multiplicative replacement: the non-zero parts are scaled down so the
		// row total is unchanged, which keeps the ratios among them exactly as
		// they were. Simply substituting would inflate the total and shift
		// every ratio in the row.
		scale := 1 - opts.ZeroReplacement*float64(zeros)/total
		for j, value := range row {
			if value == 0 {
				row[j] = opts.ZeroReplacement
			} else {
				row[j] = value * scale
			}
		}
	}

	result.Messages = append(result.Messages, fmt.Sprintf(
		"Replaced %d zero value(s) with %g in %s, scaling the other parts so each row "+
			"total is unchanged (multiplicative replacement)",
		zeroCount, opts.ZeroReplacement, describeRows(affectedRows)))
	return nil
}

// describeClosure reports whether the selected columns behave like a closed
// composition, without requiring it.
//
// CLR is valid on any positive data and a subcomposition is still
// compositional, so a varying total is not an error. It is worth saying out
// loud either way: a constant total confirms the columns are the whole of
// something, and a wildly varying one is a hint that the selection may not be
// a composition at all.
func describeClosure(parts [][]float64, names []string) string {
	if len(parts) == 0 {
		return ""
	}

	sums := make([]float64, len(parts))
	min, max := math.Inf(1), math.Inf(-1)
	mean := 0.0
	for i, row := range parts {
		total := 0.0
		for _, value := range row {
			total += value
		}
		sums[i] = total
		mean += total
		if total < min {
			min = total
		}
		if total > max {
			max = total
		}
	}
	mean /= float64(len(parts))

	if mean > 0 && (max-min)/mean < closureTolerance {
		return fmt.Sprintf(
			"The %d selected parts sum to %.4g in every row, so they look like a "+
				"closed composition", len(names), mean)
	}
	return fmt.Sprintf(
		"The %d selected parts sum to between %.4g and %.4g, so they are not closed. "+
			"That is fine for a subcomposition, but check the selection is the set of "+
			"parts you meant", len(names), min, max)
}

// describeRows renders row numbers, listing the first few.
func describeRows(rows []int) string {
	const maxListed = 5
	if len(rows) == 0 {
		return "no rows"
	}

	listed := rows
	suffix := ""
	if len(rows) > maxListed {
		listed = rows[:maxListed]
		suffix = fmt.Sprintf(" and %d more", len(rows)-maxListed)
	}
	numbers := make([]string, len(listed))
	for i, row := range listed {
		numbers[i] = strconv.Itoa(row)
	}
	noun := "row"
	if len(rows) != 1 {
		noun = "rows"
	}
	return fmt.Sprintf("%s %s%s", noun, strings.Join(numbers, ", "), suffix)
}

// clrPrecision is the number of significant figures CLR results are written
// with.
//
// Twelve rather than the six used elsewhere in this package: these are
// logarithms, typically between -5 and 5, and six figures would discard
// detail a later analysis can still resolve. Twelve is far more than any
// instrument carries and short of float64's sixteen, so a stored value
// differs from the computed one by around 1e-12 relative -- which is what the
// tests compare against rather than pretending the round trip is exact.
const clrPrecision = 12

// formatCLR renders a transformed value.
func formatCLR(value float64) string {
	return strconv.FormatFloat(value, 'g', clrPrecision, 64)
}
