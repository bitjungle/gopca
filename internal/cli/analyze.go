package cli

import (
	"fmt"
	"strings"

	"github.com/bitjungle/complab/internal/core"
	"github.com/bitjungle/complab/pkg/types"
	"github.com/urfave/cli/v2"
)

func analyzeCommand() *cli.Command {
	return &cli.Command{
		Name:      "analyze",
		Usage:     "Perform PCA analysis on input data",
		ArgsUsage: "<input.csv>",
		Description: `The analyze command performs Principal Component Analysis on the input CSV file.

USAGE:
  complab-cli analyze [OPTIONS] <input.csv>

  The input CSV file should be specified as the last argument.
  All options must come BEFORE the filename.

EXAMPLES:
  # Basic analysis with default settings (2 components, table output)
  complab-cli analyze data/iris_data.csv

  # Standard scaling with 3 components
  complab-cli analyze --scale standard -c 3 data/iris_data.csv

  # Save results to CSV file
  complab-cli analyze -f csv -o results.csv data/iris_data.csv

  # JSON output with all results
  complab-cli analyze -f json --output-all data/iris_data.csv

  # Quiet mode for scripting (CSV to stdout)
  complab-cli analyze -f csv --quiet data/iris_data.csv

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
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory (default: same as input file)",
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
	
	// Configure PCA
	pcaConfig := types.PCAConfig{
		Components:    c.Int("components"),
		MeanCenter:    !c.Bool("no-mean-centering"),
		StandardScale: c.String("scale") == "standard",
		Method:        c.String("method"),
	}
	
	// Check if requested components exceed available dimensions
	maxComponents := min(data.Rows-1, data.Columns)
	if pcaConfig.Components > maxComponents {
		return fmt.Errorf("requested %d components but data only supports maximum %d components", 
			pcaConfig.Components, maxComponents)
	}
	
	// Apply preprocessing if needed
	preprocessor := core.NewPreprocessor(
		pcaConfig.MeanCenter,
		pcaConfig.StandardScale,
		c.String("scale") == "robust",
	)
	
	if verbose {
		fmt.Println("\nPreprocessing data...")
		if pcaConfig.MeanCenter {
			fmt.Println("  - Mean centering")
		}
		if c.String("scale") != "none" {
			fmt.Printf("  - Applying %s scaling\n", c.String("scale"))
		}
	}
	
	// Preprocess data
	processedData, err := preprocessor.FitTransform(data.Matrix)
	if err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}
	
	// Run PCA
	if verbose {
		fmt.Printf("\nRunning PCA analysis using %s method...\n", pcaConfig.Method)
	}
	
	engine := core.NewPCAEngine()
	result, err := engine.Fit(processedData, pcaConfig)
	if err != nil {
		return fmt.Errorf("PCA analysis failed: %w", err)
	}
	
	// Add preprocessing statistics to the result
	result.Means = preprocessor.GetMeans()
	result.StdDevs = preprocessor.GetStdDevs()
	
	if verbose {
		fmt.Println("\n✓ PCA analysis completed successfully")
		fmt.Printf("  - Explained variance: %.1f%% (PC1), %.1f%% (PC2)\n", 
			result.ExplainedVar[0], result.ExplainedVar[1])
		fmt.Printf("  - Cumulative variance: %.1f%%\n", 
			result.CumulativeVar[len(result.CumulativeVar)-1])
	}
	
	// Prepare output
	outputFormat := c.String("format")
	outputDir := c.String("output-dir")
	
	if verbose {
		fmt.Printf("\nOutput configuration:\n")
		fmt.Printf("  Format: %s\n", outputFormat)
		fmt.Printf("  Output dir: %s\n", outputDir)
	}
	
	// Handle output options
	outputScores := c.Bool("output-scores") || c.Bool("output-all")
	outputLoadings := c.Bool("output-loadings") || c.Bool("output-all")
	outputVariance := c.Bool("output-variance") || c.Bool("output-all")
	includeMetrics := c.Bool("include-metrics")
	
	// Format and output results
	switch outputFormat {
	case "table":
		if !quiet {
			err = outputTableFormat(result, data, outputScores, outputLoadings, 
				outputVariance, includeMetrics)
		}
	case "csv":
		// CSV output is different - don't show table
		err = outputCSVFormat(result, data, inputFile, outputDir, outputScores, outputLoadings, 
			outputVariance, includeMetrics)
	case "json":
		// JSON output is different - don't show table
		err = outputJSONFormat(result, data, inputFile, outputDir, outputScores, outputLoadings, 
			outputVariance, includeMetrics)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	
	if err != nil {
		return fmt.Errorf("output failed: %w", err)
	}
	
	return nil
}