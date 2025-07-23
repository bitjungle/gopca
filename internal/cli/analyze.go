package cli

import (
	"fmt"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/urfave/cli/v2"
)

func analyzeCommand() *cli.Command {
	return &cli.Command{
		Name:      "analyze",
		Usage:     "Perform PCA analysis on input data",
		ArgsUsage: "<input.csv>",
		Description: `The analyze command performs Principal Component Analysis on the input CSV file.

USAGE:
  gopca-cli analyze [OPTIONS] <input.csv>

  The input CSV file should be specified as the last argument.
  All options must come BEFORE the filename.

EXAMPLES:
  # Basic analysis with default settings (2 components, table output)
  gopca-cli analyze data/iris_data.csv

  # Standard scaling with 3 components
  gopca-cli analyze --scale standard -c 3 data/iris_data.csv

  # Save results to CSV file
  gopca-cli analyze -f csv -o results.csv data/iris_data.csv

  # JSON output with all results
  gopca-cli analyze -f json --output-all data/iris_data.csv

  # Quiet mode for scripting (CSV to stdout)
  gopca-cli analyze -f csv --quiet data/iris_data.csv

  # Exclude specific rows and columns
  gopca-cli analyze --exclude-rows 1,5-10 --exclude-cols 3,4 data/iris_data.csv

  # Kernel PCA with RBF kernel
  gopca-cli analyze --method kernel --kernel-type rbf --kernel-gamma 0.5 data/iris_data.csv

  # Kernel PCA with polynomial kernel
  gopca-cli analyze --method kernel --kernel-type poly --kernel-degree 3 data/iris_data.csv

The analysis includes:
  - Data preprocessing (mean centering, scaling)
  - PCA computation using SVD, NIPALS, or Kernel methods
  - Kernel PCA for non-linear dimensionality reduction
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
				Usage: "PCA algorithm: svd, nipals, kernel",
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
				Usage: "CSV field delimiter (comma, semicolon, tab)",
				Value: ",",
			},
			&cli.StringFlag{
				Name:  "decimal-separator",
				Usage: "Decimal separator (dot, comma)",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "na-values",
				Usage: "String(s) representing missing values (comma-separated)",
				Value: ",NA,N/A,nan,NaN,null,NULL,m",
			},
			&cli.StringFlag{
				Name:  "missing-strategy",
				Usage: "How to handle missing values: error, drop, mean, median",
				Value: "error",
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
			
			// Data filtering
			&cli.StringFlag{
				Name:  "exclude-rows",
				Usage: "Exclude rows by index (1-based, e.g., '1,3,5-7')",
			},
			&cli.StringFlag{
				Name:  "exclude-cols",
				Usage: "Exclude columns by index (1-based, e.g., '2,4-6,8')",
			},
			
			// Kernel PCA parameters
			&cli.StringFlag{
				Name:  "kernel-type",
				Usage: "Kernel type for kernel PCA: rbf, linear, poly",
			},
			&cli.Float64Flag{
				Name:  "kernel-gamma",
				Usage: "Gamma parameter for RBF and polynomial kernels",
				Value: 1.0,
			},
			&cli.IntFlag{
				Name:  "kernel-degree",
				Usage: "Degree for polynomial kernel",
				Value: 3,
			},
			&cli.Float64Flag{
				Name:  "kernel-coef0",
				Usage: "Independent term for polynomial kernel",
				Value: 0.0,
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
	case "svd", "nipals", "kernel":
		// Valid methods
	default:
		return fmt.Errorf("invalid PCA method: %s (must be svd, nipals, or kernel)", method)
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
	delimiter := c.String("delimiter")
	if delimiter == "tab" {
		delimiter = "\t"
	}
	if len(delimiter) != 1 {
		return fmt.Errorf("delimiter must be a single character")
	}
	
	// Validate decimal separator
	decimalSep := c.String("decimal-separator")
	if decimalSep != "." && decimalSep != "," && decimalSep != "dot" && decimalSep != "comma" {
		return fmt.Errorf("decimal-separator must be 'dot' or 'comma'")
	}
	
	// Validate missing strategy
	missingStrategy := c.String("missing-strategy")
	if missingStrategy != "error" && missingStrategy != "drop" && missingStrategy != "mean" && missingStrategy != "median" {
		return fmt.Errorf("missing-strategy must be one of: error, drop, mean, median")
	}
	
	// Validate kernel parameters if kernel method is selected
	if method == "kernel" {
		kernelType := c.String("kernel-type")
		if kernelType == "" {
			return fmt.Errorf("kernel-type must be specified when using kernel PCA method")
		}
		
		switch kernelType {
		case "rbf", "linear", "poly":
			// Valid kernel types
		default:
			return fmt.Errorf("invalid kernel type: %s (must be rbf, linear, or poly)", kernelType)
		}
		
		// Validate kernel-specific parameters
		if kernelType == "rbf" || kernelType == "poly" {
			if c.Float64("kernel-gamma") <= 0 {
				return fmt.Errorf("kernel-gamma must be positive for %s kernel", kernelType)
			}
		}
		
		if kernelType == "poly" {
			if c.Int("kernel-degree") < 1 {
				return fmt.Errorf("kernel-degree must be at least 1 for polynomial kernel")
			}
		}
	}
	
	return nil
}

// contains checks if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func runAnalyze(c *cli.Context) error {
	inputFile := c.Args().First()
	verbose := c.Bool("verbose")
	quiet := c.Bool("quiet")
	
	// Parse CSV options
	parseOpts := NewCSVParseOptions()
	parseOpts.HasHeaders = !c.Bool("no-headers")
	parseOpts.HasIndex = !c.Bool("no-index")
	
	// Handle delimiter
	delimiter := c.String("delimiter")
	if delimiter == "tab" {
		delimiter = "\t"
	}
	parseOpts.Delimiter = rune(delimiter[0])
	
	// Handle decimal separator
	decimalSep := c.String("decimal-separator")
	if decimalSep == "dot" {
		parseOpts.DecimalSeparator = '.'
	} else if decimalSep == "comma" || decimalSep == "," {
		parseOpts.DecimalSeparator = ','
	} else {
		parseOpts.DecimalSeparator = rune(decimalSep[0])
	}
	
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
	
	// Parse exclusion flags
	var excludedRows, excludedCols []int
	
	if excludeRowsStr := c.String("exclude-rows"); excludeRowsStr != "" {
		excludedRows, err = utils.ParseRanges(excludeRowsStr)
		if err != nil {
			return fmt.Errorf("invalid exclude-rows format: %w", err)
		}
		if verbose && len(excludedRows) > 0 {
			fmt.Printf("\nExcluding %d rows: %v\n", len(excludedRows), excludedRows)
		}
	}
	
	if excludeColsStr := c.String("exclude-cols"); excludeColsStr != "" {
		excludedCols, err = utils.ParseRanges(excludeColsStr)
		if err != nil {
			return fmt.Errorf("invalid exclude-cols format: %w", err)
		}
		if verbose && len(excludedCols) > 0 {
			fmt.Printf("Excluding %d columns: %v\n", len(excludedCols), excludedCols)
		}
	}
	
	// Apply exclusions to data if needed
	if len(excludedRows) > 0 || len(excludedCols) > 0 {
		// Filter the data matrix
		filteredData, err := utils.FilterMatrix(data.Matrix, excludedRows, excludedCols)
		if err != nil {
			return fmt.Errorf("failed to filter data: %w", err)
		}
		data.Matrix = filteredData
		
		// Filter row names
		if len(excludedRows) > 0 && len(data.RowNames) > 0 {
			filteredRowNames, err := utils.FilterStringSlice(data.RowNames, excludedRows)
			if err != nil {
				return fmt.Errorf("failed to filter row names: %w", err)
			}
			data.RowNames = filteredRowNames
		}
		
		// Filter column names
		if len(excludedCols) > 0 && len(data.Headers) > 0 {
			filteredHeaders, err := utils.FilterStringSlice(data.Headers, excludedCols)
			if err != nil {
				return fmt.Errorf("failed to filter column names: %w", err)
			}
			data.Headers = filteredHeaders
		}
		
		// Update dimensions
		data.Rows = len(data.Matrix)
		if data.Rows > 0 {
			data.Columns = len(data.Matrix[0])
		} else {
			data.Columns = 0
		}
		
		if verbose {
			fmt.Printf("\nData after filtering:")
			fmt.Print(GetDataSummary(data))
		}
	}
	
	// Handle missing values after filtering
	// Get the columns that will be used for PCA (after exclusion)
	selectedCols := make([]int, 0, data.Columns)
	for i := 0; i < data.Columns; i++ {
		if !contains(excludedCols, i) {
			selectedCols = append(selectedCols, i)
		}
	}
	
	// Check for missing values in selected columns
	missingInfo := data.GetMissingValueInfo(selectedCols)
	if missingInfo.HasMissing() {
		if verbose {
			fmt.Printf("\nMissing values detected: %s\n", missingInfo.GetSummary())
		}
		
		// Handle based on strategy
		missingStrategy := c.String("missing-strategy")
		switch missingStrategy {
		case "error":
			return fmt.Errorf("missing values found in selected columns - use --missing-strategy to specify handling")
		case "drop", "mean", "median":
			handler := core.NewMissingValueHandler(types.MissingValueStrategy(missingStrategy))
			cleanData, err := handler.HandleMissingValues(data.Matrix, missingInfo, selectedCols)
			if err != nil {
				return fmt.Errorf("failed to handle missing values: %w", err)
			}
			
			// Update data matrix and affected row names
			if missingStrategy == "drop" && len(data.RowNames) > 0 {
				// Filter row names to match the cleaned data
				cleanRowNames := make([]string, 0, len(cleanData))
				droppedRows := make(map[int]bool)
				for _, row := range missingInfo.RowsAffected {
					droppedRows[row] = true
				}
				for i, name := range data.RowNames {
					if !droppedRows[i] {
						cleanRowNames = append(cleanRowNames, name)
					}
				}
				data.RowNames = cleanRowNames
			}
			
			data.Matrix = cleanData
			data.Rows = len(cleanData)
			
			if verbose {
				fmt.Printf("Applied %s strategy for missing values\n", missingStrategy)
				if missingStrategy == "drop" {
					fmt.Printf("Removed %d rows containing missing values\n", len(missingInfo.RowsAffected))
				}
			}
		}
	}
	
	// Configure PCA
	pcaConfig := types.PCAConfig{
		Components:      c.Int("components"),
		MeanCenter:      !c.Bool("no-mean-centering"),
		StandardScale:   c.String("scale") == "standard",
		RobustScale:     c.String("scale") == "robust",
		Method:          c.String("method"),
		ExcludedRows:    excludedRows,
		ExcludedColumns: excludedCols,
		MissingStrategy: types.MissingValueStrategy(c.String("missing-strategy")),
	}
	
	// Add kernel parameters if using kernel PCA
	if c.String("method") == "kernel" {
		pcaConfig.KernelType = c.String("kernel-type")
		pcaConfig.KernelGamma = c.Float64("kernel-gamma")
		pcaConfig.KernelDegree = c.Int("kernel-degree")
		pcaConfig.KernelCoef0 = c.Float64("kernel-coef0")
	}
	
	// Check if requested components exceed available dimensions
	maxComponents := min(data.Rows-1, data.Columns)
	if pcaConfig.Components > maxComponents {
		return fmt.Errorf("requested %d components but data only supports maximum %d components", 
			pcaConfig.Components, maxComponents)
	}
	
	// Apply preprocessing if needed (kernel PCA typically doesn't use standard preprocessing)
	var processedData types.Matrix
	var preprocessor *core.Preprocessor
	
	if pcaConfig.Method == "kernel" {
		// Kernel PCA handles its own centering in the kernel space
		processedData = data.Matrix
		if verbose {
			fmt.Println("\nSkipping standard preprocessing for kernel PCA")
		}
	} else {
		preprocessor = core.NewPreprocessor(
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
		processedData, err = preprocessor.FitTransform(data.Matrix)
		if err != nil {
			return fmt.Errorf("preprocessing failed: %w", err)
		}
	}
	
	// Run PCA
	if verbose {
		fmt.Printf("\nRunning PCA analysis using %s method...\n", pcaConfig.Method)
	}
	
	engine := core.NewPCAEngineForMethod(pcaConfig.Method)
	result, err := engine.Fit(processedData, pcaConfig)
	if err != nil {
		return fmt.Errorf("PCA analysis failed: %w", err)
	}
	
	// Add preprocessing statistics to the result (if preprocessing was done)
	if preprocessor != nil {
		result.Means = preprocessor.GetMeans()
		result.StdDevs = preprocessor.GetStdDevs()
	}
	
	if verbose {
		fmt.Println("\n✓ PCA analysis completed successfully")
		fmt.Printf("  - Explained variance: %.1f%% (PC1), %.1f%% (PC2)\n", 
			result.ExplainedVarRatio[0], result.ExplainedVarRatio[1])
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