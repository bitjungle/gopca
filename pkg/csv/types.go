// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package csv provides unified CSV parsing, writing, and validation functionality
// for the GoPCA monorepo. It consolidates previously scattered CSV operations
// into a single, well-tested package following the DRY principle.
package csv

import (
	"io"

	"github.com/bitjungle/gopca/pkg/types"
)

// ParseMode defines how CSV data should be parsed
type ParseMode int

const (
	// ParseNumeric treats all data as numeric values
	ParseNumeric ParseMode = iota
	// ParseString treats all data as strings
	ParseString
	// ParseMixed automatically detects column types
	ParseMixed
	// ParseMixedWithTargets detects columns and identifies target columns
	ParseMixedWithTargets
)

// Options provides unified configuration for CSV operations
type Options struct {
	// Parsing options
	Delimiter        rune      // Field delimiter: ',', ';', '\t'
	DecimalSeparator rune      // Decimal separator: '.', ','
	HasHeaders       bool      // First row contains column names
	HasRowNames      bool      // First column contains row names
	NullValues       []string  // Strings to treat as missing values
	ParseMode        ParseMode // How to parse the data
	TargetSuffix     string    // Suffix to identify target columns (e.g., "#target")

	// Reading options (for large files)
	SkipRows      int   // Number of rows to skip at start
	MaxRows       int   // Maximum rows to read (0 for all)
	Columns       []int // Specific columns to read (empty for all)
	StreamingMode bool  // Enable streaming for large files

	// Writing options
	FloatFormat byte // Format for float output: 'g', 'f', 'e'
	Precision   int  // Decimal precision for float output (-1 for auto)
}

// DefaultOptions returns sensible default options for CSV operations
func DefaultOptions() Options {
	return Options{
		Delimiter:        ',',
		DecimalSeparator: '.',
		HasHeaders:       true,
		HasRowNames:      true,
		NullValues:       []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL", "m"},
		ParseMode:        ParseNumeric,
		TargetSuffix:     "#target",
		SkipRows:         0,
		MaxRows:          0,
		Columns:          nil,
		StreamingMode:    false,
		FloatFormat:      'g',
		Precision:        -1,
	}
}

// EuropeanOptions returns options for European CSV format (semicolon delimiter, comma decimal)
func EuropeanOptions() Options {
	opts := DefaultOptions()
	opts.Delimiter = ';'
	opts.DecimalSeparator = ','
	return opts
}

// TabDelimitedOptions returns options for tab-delimited files
func TabDelimitedOptions() Options {
	opts := DefaultOptions()
	opts.Delimiter = '\t'
	return opts
}

// Data represents parsed CSV data with support for different data types
type Data struct {
	// Core numeric data (always present for PCA)
	Matrix      types.Matrix // Numeric data matrix
	Headers     []string     // Column names
	RowNames    []string     // Row names
	MissingMask [][]bool     // Track missing values (true = missing)
	Rows        int          // Number of data rows
	Columns     int          // Number of data columns

	// Additional data types (optional)
	StringData           [][]string           // Raw string data (for GoCSV)
	CategoricalColumns   map[string][]string  // Categorical columns by name
	NumericTargetColumns map[string][]float64 // Numeric target columns
}

// DataProvider is an interface that different data representations can implement
// to provide consistent access to CSV data regardless of internal structure
type DataProvider interface {
	GetHeaders() []string
	GetRowNames() []string
	GetDimensions() (rows, cols int)
	HasNumericData() bool
	HasStringData() bool
}

// CSVWritable is an interface for data that can be written to CSV
type CSVWritable interface {
	DataProvider
	WriteHeaders(w io.Writer, opts Options) error
	WriteRow(w io.Writer, index int, opts Options) error
}

// Reader provides unified CSV reading functionality
type Reader struct {
	opts Options
}

// Writer provides unified CSV writing functionality
type Writer struct {
	opts Options
}

// Validator provides CSV validation functionality
type Validator struct {
	opts Options
}

// ValidationResult contains the results of CSV validation
type ValidationResult struct {
	Valid       bool
	Errors      []string
	Warnings    []string
	ColumnStats []ColumnStatistics
}

// ColumnStatistics contains statistics for a single column
type ColumnStatistics struct {
	Name            string
	Index           int
	DataType        string // "numeric", "categorical", "mixed"
	NonMissing      int
	Missing         int
	MissingPercent  float64
	Mean            float64 // For numeric columns
	StdDev          float64 // For numeric columns
	Min             float64 // For numeric columns
	Max             float64 // For numeric columns
	UniqueValues    int     // For categorical columns
	HasZeroVariance bool    // Warning flag
}

// ConversionHelpers provide utilities for converting between data representations

// ToNumericMatrix converts string data to numeric matrix with missing value tracking
func ToNumericMatrix(stringData [][]string, nullValues []string) (types.Matrix, [][]bool, error) {
	// Implementation will be in reader.go
	return nil, nil, nil
}

// ToStringMatrix converts numeric matrix to string representation
func ToStringMatrix(matrix types.Matrix, precision int) [][]string {
	// Implementation will be in writer.go
	return nil
}
