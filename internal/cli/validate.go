package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/urfave/cli/v2"
)

func validateCommand() *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "Validate input data for PCA analysis",
		ArgsUsage: "<input.csv>",
		Description: `The validate command checks the input CSV file for common issues before PCA analysis.

USAGE:
  gopca-cli validate [OPTIONS] <input.csv>

EXAMPLES:
  # Basic validation
  gopca-cli validate data/iris_data.csv

  # Show detailed summary
  gopca-cli validate --summary data/iris_data.csv

  # Strict mode (fail on warnings)
  gopca-cli validate --strict data/iris_data.csv

The validation includes:
  - File format and structure
  - Missing values detection
  - Data type consistency
  - Numerical range checks
  - Low variance detection
  - High missing value warnings`,
		Flags: []cli.Flag{
			// Data format flags (same as analyze)
			&cli.BoolFlag{
				Name:  "no-headers",
				Usage: "First row contains data, not column names",
			},
			&cli.BoolFlag{
				Name:  "no-index",
				Usage: "First column contains data, not row names",
			},
			&cli.StringFlag{
				Name:  "delimiter",
				Usage: "CSV field delimiter",
				Value: ",",
			},
			&cli.StringFlag{
				Name:  "na-values",
				Usage: "String(s) representing missing values (comma-separated)",
				Value: "NA,NaN",
			},
			
			// Validation options
			&cli.BoolFlag{
				Name:  "strict",
				Usage: "Enable strict validation (fail on warnings)",
			},
			&cli.BoolFlag{
				Name:  "summary",
				Usage: "Show data summary statistics",
			},
		},
		Action: runValidate,
		Before: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("missing required argument: input CSV file")
			}
			return nil
		},
	}
}

func runValidate(c *cli.Context) error {
	inputFile := c.Args().First()
	strict := c.Bool("strict")
	showSummary := c.Bool("summary")
	
	// Parse CSV options
	parseOpts := NewCSVParseOptions()
	parseOpts.HasHeaders = !c.Bool("no-headers")
	parseOpts.HasIndex = !c.Bool("no-index")
	parseOpts.Delimiter = rune(c.String("delimiter")[0])
	
	// Parse NA values
	if naValues := c.String("na-values"); naValues != "" {
		parseOpts.NullValues = strings.Split(naValues, ",")
		for i := range parseOpts.NullValues {
			parseOpts.NullValues[i] = strings.TrimSpace(parseOpts.NullValues[i])
		}
	}
	
	fmt.Printf("Validating file: %s\n", inputFile)
	
	// Load CSV data
	data, err := ParseCSV(inputFile, parseOpts)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}
	
	// Basic validation
	if err := ValidateCSVData(data); err != nil {
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
	
	if showSummary {
		fmt.Println("\nData summary:")
		summary := GetDataSummary(data)
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
		
		if strict {
			return fmt.Errorf("validation failed with %d warnings in strict mode", len(warnings))
		}
	} else {
		fmt.Println("\n✓ No warnings found")
	}
	
	fmt.Println("\n✓ Data is ready for PCA analysis")
	
	return nil
}