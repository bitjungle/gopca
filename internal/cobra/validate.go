// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"fmt"
	"math"
	"strings"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/spf13/cobra"
)

// ValidateOptions holds all the options for the validate command
type ValidateOptions struct {
	// Data format options
	NoHeaders bool
	NoIndex   bool
	Delimiter string
	NAValues  string

	// Validation options
	Strict  bool
	Summary bool
}

// NewValidateCommand creates the validate subcommand
func NewValidateCommand() *cobra.Command {
	opts := &ValidateOptions{}

	cmd := &cobra.Command{
		Use:   "validate [flags] <input.csv>",
		Short: "Validate input data for PCA analysis",
		Long: `Validate CSV data before PCA analysis.

The validate command checks your data for common issues that might
affect PCA analysis, including missing values, low variance columns,
and data format issues.

EXAMPLES:
  # Basic validation
  pca validate data.csv

  # Show detailed summary
  pca validate --summary data.csv

  # Strict mode (fail on warnings)
  pca validate --strict data.csv

The validation includes:
  • File format and structure
  • Missing values detection
  • Data type consistency
  • Numerical range checks
  • Low variance detection
  • High missing value warnings`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(opts, args[0])
		},
	}

	// Data format options
	cmd.Flags().BoolVar(&opts.NoHeaders, "no-headers", false,
		"First row contains data, not column names")
	cmd.Flags().BoolVar(&opts.NoIndex, "no-index", false,
		"First column contains data, not row names")
	cmd.Flags().StringVar(&opts.Delimiter, "delimiter", ",",
		"CSV field delimiter")
	cmd.Flags().StringVar(&opts.NAValues, "na-values", ",NA,N/A,nan,NaN,null,NULL,m",
		"Comma-separated list of strings representing missing values")

	// Validation options
	cmd.Flags().BoolVar(&opts.Strict, "strict", false,
		"Enable strict validation (fail on warnings)")
	cmd.Flags().BoolVar(&opts.Summary, "summary", false,
		"Show data summary statistics")

	return cmd
}

// runValidate executes the validate command
func runValidate(opts *ValidateOptions, inputFile string) error {
	// Parse CSV options
	parseOpts := pkgcsv.DefaultOptions()
	parseOpts.HasHeaders = !opts.NoHeaders
	parseOpts.HasRowNames = !opts.NoIndex
	parseOpts.Delimiter = rune(opts.Delimiter[0])
	parseOpts.ParseMode = pkgcsv.ParseMixedWithTargets

	// Parse NA values
	if opts.NAValues != "" {
		parseOpts.NullValues = strings.Split(opts.NAValues, ",")
		for i := range parseOpts.NullValues {
			parseOpts.NullValues[i] = strings.TrimSpace(parseOpts.NullValues[i])
		}
	}

	fmt.Printf("Validating file: %s\n", inputFile)

	// Load CSV data with target column detection
	reader := pkgcsv.NewReader(parseOpts)
	data, err := reader.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Basic validation
	if err := validateCSVData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}

	// Perform additional validation checks
	warnings := []string{}

	// Check for low variance columns
	for j := 0; j < data.Columns; j++ {
		var values []float64
		for i := 0; i < data.Rows; i++ {
			if !math.IsNaN(data.Matrix[i][j]) {
				values = append(values, data.Matrix[i][j])
			}
		}

		if len(values) > 1 {
			// Calculate variance
			mean := 0.0
			for _, v := range values {
				mean += v
			}
			mean /= float64(len(values))

			variance := 0.0
			for _, v := range values {
				variance += (v - mean) * (v - mean)
			}
			variance /= float64(len(values) - 1)

			if variance < 1e-10 {
				colName := fmt.Sprintf("column %d", j+1)
				if j < len(data.Headers) {
					colName = data.Headers[j]
				}
				warnings = append(warnings, fmt.Sprintf("%s has near-zero variance", colName))
			}
		}
	}

	// Check for high missing value percentage per column
	for j := 0; j < data.Columns; j++ {
		missingCount := 0
		for i := 0; i < data.Rows; i++ {
			if math.IsNaN(data.Matrix[i][j]) {
				missingCount++
			}
		}

		missingPercent := float64(missingCount) / float64(data.Rows) * 100
		if missingPercent > 50 {
			colName := fmt.Sprintf("column %d", j+1)
			if j < len(data.Headers) {
				colName = data.Headers[j]
			}
			warnings = append(warnings, fmt.Sprintf("%s has %.1f%% missing values", colName, missingPercent))
		}
	}

	// Display results
	fmt.Println("\n✓ Data format validation passed")
	fmt.Printf("  - Dimensions: %d rows × %d columns\n", data.Rows, data.Columns)

	// Show categorical columns if any
	if len(data.CategoricalColumns) > 0 {
		fmt.Printf("  - Categorical columns: %d\n", len(data.CategoricalColumns))
		for colName := range data.CategoricalColumns {
			fmt.Printf("    • %s\n", colName)
		}
	}

	// Show numeric target columns if any
	if len(data.NumericTargetColumns) > 0 {
		fmt.Printf("  - Numeric target columns: %d (excluded from PCA)\n", len(data.NumericTargetColumns))
		for colName := range data.NumericTargetColumns {
			fmt.Printf("    • %s\n", colName)
		}
	}

	if opts.Summary {
		fmt.Println("\nData summary:")
		summary := getDataSummary(data)
		// Add indentation to summary
		lines := strings.Split(summary, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("  %s\n", line)
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Println("\n⚠ Warnings:")
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}

		if opts.Strict {
			return fmt.Errorf("validation failed with %d warnings in strict mode", len(warnings))
		}
	} else {
		fmt.Println("\n✓ No warnings found")
	}

	fmt.Println("\n✓ Data is ready for PCA analysis")

	return nil
}
