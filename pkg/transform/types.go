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

// Type identifies a supported transformation.
type Type string

const (
	// Log applies the natural logarithm to each value. Values must be positive.
	Log Type = "log"
	// Sqrt applies the square root to each value. Values must be non-negative.
	Sqrt Type = "sqrt"
	// Square squares each value.
	Square Type = "square"
	// Standardize applies z-score standardization (mean=0, std=1).
	Standardize Type = "standardize"
	// MinMax applies min-max scaling to a configurable target range (default [0, 1]).
	MinMax Type = "minmax"
	// Bin discretizes a numeric column into equal-width bins.
	Bin Type = "bin"
	// OneHot encodes a categorical column into one binary column per unique value.
	OneHot Type = "onehot"
	// Ordinal encodes a categorical column into a single integer column, one
	// code per unique value, assigned by position in a caller-supplied order.
	Ordinal Type = "ordinal"
)

// Options configures a transformation.
type Options struct {
	// Type is the transformation to apply.
	Type Type
	// Columns lists the column names to transform.
	Columns []string
	// BinCount is the number of bins for Bin transformations (default: 5).
	BinCount int
	// MinValue is the lower bound of the target range for MinMax scaling (default: 0).
	MinValue float64
	// MaxValue is the upper bound of the target range for MinMax scaling (default: 1).
	MaxValue float64

	// RemoveOriginal drops the source column after one-hot encoding it.
	//
	// The polarity is deliberate: the zero value keeps the column, so a caller
	// that does not set the field cannot silently lose data. Discarding a column
	// is the kind of thing that should be asked for rather than defaulted into.
	//
	// One-hot encoding used to remove the source unconditionally, which made it
	// the only transformation in this package to destroy its input -- binning
	// also derives new columns and has always kept the original. It also cost
	// something specific in this suite: GoPCA colours plots by categorical
	// columns, so encoding "species" removed the ability to colour by it. That
	// is why testdata/iris/iris.csv ships both species#target and species.
	RemoveOriginal bool

	// CategoryOrder gives, per column, the category values in the order their
	// integer codes should follow: the first value encodes to 0, the second to
	// 1, and so on. Used by Ordinal encoding only.
	//
	// A column absent from the map -- or listing only some of its values -- is
	// completed alphabetically, which is what scikit-learn's LabelEncoder does
	// unconditionally (its classes_ attribute is sorted). That default is
	// deliberately not the only option: alphabetical order is wrong for exactly
	// the data ordinal encoding is appropriate for. "low", "medium", "high"
	// sorts to high=0, low=1, medium=2, scrambling the order the codes are
	// supposed to carry.
	//
	// Reference: scikit-learn documents LabelEncoder as being for target values
	// y and not input X, pointing to OrdinalEncoder for features. Everything
	// this package produces is destined to become X, hence the ordering control.
	CategoryOrder map[string][]string
}

// Input carries the tabular data and metadata that transform functions operate on.
// It mirrors the relevant fields of the application FileData type so that
// callers can pass raw slices without introducing a package dependency.
type Input struct {
	// Data is the row-major data matrix (each row is a slice of string values).
	Data [][]string
	// Headers is the ordered list of column names.
	Headers []string
	// ColumnTypes maps each column name to its detected type ("numeric" or "categorical").
	ColumnTypes map[string]string
	// CategoricalColumns maps each categorical column name to its value slice.
	// Used by binning (to register newly discretized columns) and one-hot encoding
	// (to remove the source column after expansion).
	CategoricalColumns map[string][]string
	// Rows is the number of data rows.
	Rows int
	// Columns is the number of columns (= len(Headers)).
	Columns int
}

// Result carries the output of [Apply].
// The original Input is never modified; all changes are reflected here.
type Result struct {
	// TransformedColumns lists the column names that were successfully transformed.
	TransformedColumns []string
	// NewColumns lists column names added during the transformation (e.g. one-hot columns).
	NewColumns []string
	// Messages contains informational and warning messages produced during the transformation.
	Messages []string
	// Headers is the updated ordered list of column names after the transformation.
	Headers []string
	// Data is the updated row-major data matrix after the transformation.
	Data [][]string
	// ColumnTypes is the updated column-type map after the transformation.
	ColumnTypes map[string]string
	// CategoricalColumns is the updated categorical-column value map after the transformation.
	CategoricalColumns map[string][]string
	// Columns is the updated number of columns.
	Columns int
}
