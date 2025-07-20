package cli

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

func analyzeCommand() *cli.Command {
	return &cli.Command{
		Name:      "analyze",
		Usage:     "Perform PCA analysis on input data",
		ArgsUsage: "<input.csv>",
		Description: `Analyze performs Principal Component Analysis on the input CSV file.
		
The analysis includes:
  - Data preprocessing (mean centering, scaling)
  - PCA computation using SVD or NIPALS algorithm
  - Optional statistical metrics (Hotelling's T², Mahalanobis distances, RSS)
  - Multiple output formats (table, CSV, JSON)`,
		Flags: []cli.Flag{
			// General flags
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Minimal output (for scripting)",
			},
			
			// Output flags
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (default: stdout)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: table, csv, json",
				Value:   "table",
			},
			
			// PCA parameters
			&cli.IntFlag{
				Name:    "components",
				Aliases: []string{"c"},
				Usage:   "Number of principal components to compute",
				Value:   2,
			},
			&cli.StringFlag{
				Name:  "method",
				Usage: "PCA algorithm: svd, nipals",
				Value: "svd",
			},
			&cli.BoolFlag{
				Name:  "no-mean-centering",
				Usage: "Disable mean centering",
			},
			
			// Preprocessing
			&cli.StringFlag{
				Name:  "scale",
				Usage: "Scaling method: none, standard, robust",
				Value: "none",
			},
			
			// Data format
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
			
			// Output options
			&cli.BoolFlag{
				Name:  "output-scores",
				Usage: "Include PC scores in output",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "output-loadings",
				Usage: "Include loadings in output",
			},
			&cli.BoolFlag{
				Name:  "output-variance",
				Usage: "Include explained variance in output",
			},
			&cli.BoolFlag{
				Name:  "output-all",
				Usage: "Output all results",
			},
			&cli.BoolFlag{
				Name:  "include-metrics",
				Usage: "Include advanced metrics (Hotelling's T², Mahalanobis, RSS)",
			},
		},
		Action: runAnalyze,
		Before: validateAnalyzeFlags,
	}
}

func validateAnalyzeFlags(c *cli.Context) error {
	// Validate verbose and quiet flags
	if c.Bool("verbose") && c.Bool("quiet") {
		return fmt.Errorf("cannot use both --verbose and --quiet flags")
	}
	
	// Validate arguments
	if c.NArg() < 1 {
		return fmt.Errorf("missing required argument: input CSV file")
	}
	
	// Validate format
	format := c.String("format")
	switch format {
	case "table", "csv", "json":
		// Valid formats
	default:
		return fmt.Errorf("invalid output format: %s (must be table, csv, or json)", format)
	}
	
	// Validate method
	method := c.String("method")
	switch method {
	case "svd", "nipals":
		// Valid methods
	default:
		return fmt.Errorf("invalid PCA method: %s (must be svd or nipals)", method)
	}
	
	// Validate scale
	scale := c.String("scale")
	switch scale {
	case "none", "standard", "robust":
		// Valid scaling methods
	default:
		return fmt.Errorf("invalid scaling method: %s (must be none, standard, or robust)", scale)
	}
	
	// Validate components
	if c.Int("components") < 1 {
		return fmt.Errorf("number of components must be at least 1")
	}
	
	// Validate delimiter
	if len(c.String("delimiter")) != 1 {
		return fmt.Errorf("delimiter must be a single character")
	}
	
	return nil
}

func runAnalyze(c *cli.Context) error {
	inputFile := c.Args().First()
	verbose := c.Bool("verbose")
	quiet := c.Bool("quiet")
	
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
	
	// Load CSV data
	if verbose {
		fmt.Printf("Loading data from %s...\n", inputFile)
	}
	
	data, err := ParseCSV(inputFile, parseOpts)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}
	
	// Validate data
	if err := ValidateCSVData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}
	
	if verbose {
		fmt.Println("\nData summary:")
		fmt.Print(GetDataSummary(data))
	}
	
	// TODO: Implement PCA analysis
	
	// For now, show what we loaded
	if !quiet {
		fmt.Printf("\nSuccessfully loaded data: %d rows × %d columns\n", data.Rows, data.Columns)
		fmt.Printf("Analysis configuration:\n")
		fmt.Printf("  Components: %d\n", c.Int("components"))
		fmt.Printf("  Method: %s\n", c.String("method"))
		fmt.Printf("  Scaling: %s\n", c.String("scale"))
		fmt.Printf("  Mean centering: %v\n", !c.Bool("no-mean-centering"))
	}
	
	return nil
}