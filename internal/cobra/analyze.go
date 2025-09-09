// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"fmt"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/spf13/cobra"
)

// AnalyzeOptions holds all the options for the analyze command
type AnalyzeOptions struct {
	// PCA parameters
	Components int
	Method     string

	// Kernel PCA parameters
	KernelType   string
	KernelGamma  float64
	KernelDegree int
	KernelCoef0  float64

	// Temporal PCA parameters
	TemporalLags      int
	VarianceExplained float64
	ImputeMethod      string

	// Preprocessing options
	MeanCenter      bool
	Scale           string // "none", "standard", "robust"
	ScaleOnly       bool
	SNV             bool
	VectorNorm      bool
	NoMeanCentering bool

	// Data format options
	NoHeaders  bool
	NoIndex    bool
	Delimiter  string
	NAValues   string
	TargetCols string

	// Missing data handling
	MissingStrategy string
	MissingPercent  float64

	// Output options
	OutputFormat   string
	OutputDir      string
	OutputScores   bool
	OutputLoadings bool
	OutputVariance bool
	OutputAll      bool
	IncludeMetrics bool

	// Exclude options
	ExcludeRows    string
	ExcludeColumns string

	// Verbose output
	Verbose bool
}

// NewAnalyzeCommand creates the analyze subcommand
func NewAnalyzeCommand() *cobra.Command {
	opts := &AnalyzeOptions{}

	cmd := &cobra.Command{
		Use:   "analyze [flags] <input.csv>",
		Short: "Perform PCA analysis on input data",
		Long: `Perform Principal Component Analysis on CSV data.

The analyze command performs PCA on your data using various algorithms
and preprocessing options. It supports multiple output formats and
advanced diagnostics.

EXAMPLES:
  # Basic PCA with 2 components
  pca analyze data.csv --components 2

  # PCA with standardization and metrics
  pca analyze --standard-scale --include-metrics data.csv

  # Kernel PCA with RBF kernel
  pca analyze --method kernel --kernel-type rbf data.csv

  # Handle missing data by dropping rows
  pca analyze --missing-strategy drop data.csv

  # NIPALS with native missing value handling
  pca analyze --method nipals --missing-strategy native data.csv

  # Temporal PCA with 24 lags (SSA-style)
  pca analyze --method temporal --temporal-lags 24 --components 5 data.csv

  # Temporal PCA with variance explained criterion
  pca analyze --method temporal --temporal-lags 12 --var-explained 0.95 data.csv

  # Output to JSON with full results
  pca analyze -f json --output-dir results/ data.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(opts, args[0])
		},
	}

	// PCA parameters
	cmd.Flags().IntVarP(&opts.Components, "components", "c", 2,
		"Number of principal components")
	cmd.Flags().StringVarP(&opts.Method, "method", "m", "svd",
		"PCA method: svd, nipals, kernel, or temporal")

	// Kernel PCA parameters
	cmd.Flags().StringVar(&opts.KernelType, "kernel-type", "rbf",
		"Kernel type for kernel PCA: linear, poly, rbf")
	cmd.Flags().Float64Var(&opts.KernelGamma, "kernel-gamma", 0.01,
		"Gamma parameter for RBF/poly kernels")
	cmd.Flags().IntVar(&opts.KernelDegree, "kernel-degree", 3,
		"Degree for polynomial kernel")
	cmd.Flags().Float64Var(&opts.KernelCoef0, "kernel-coef0", 0.0,
		"Coef0 for polynomial kernel")

	// Temporal PCA parameters
	cmd.Flags().IntVar(&opts.TemporalLags, "temporal-lags", 0,
		"Number of time lags for temporal PCA (SSA-style)")
	cmd.Flags().Float64Var(&opts.VarianceExplained, "var-explained", 0.0,
		"Target explained variance (0.0-1.0), alternative to --components")
	cmd.Flags().StringVar(&opts.ImputeMethod, "impute-method", "none",
		"Imputation method for temporal PCA: forward, backward, linear, none")

	// Preprocessing options
	cmd.Flags().BoolVar(&opts.NoMeanCentering, "no-mean-centering", false,
		"Disable mean centering")
	cmd.Flags().StringVar(&opts.Scale, "scale", "none",
		"Scaling method: none, standard, robust")
	cmd.Flags().BoolVar(&opts.ScaleOnly, "scale-only", false,
		"Scale without centering")
	cmd.Flags().BoolVar(&opts.SNV, "snv", false,
		"Apply Standard Normal Variate transformation")
	cmd.Flags().BoolVar(&opts.VectorNorm, "vector-norm", false,
		"Apply L2 vector normalization (row-wise)")

	// Data format options
	cmd.Flags().BoolVar(&opts.NoHeaders, "no-headers", false,
		"First row contains data, not column names")
	cmd.Flags().BoolVar(&opts.NoIndex, "no-index", false,
		"First column contains data, not row names")
	cmd.Flags().StringVar(&opts.Delimiter, "delimiter", ",",
		"CSV field delimiter")
	cmd.Flags().StringVar(&opts.NAValues, "na-values", ",NA,N/A,nan,NaN,null,NULL,m",
		"Comma-separated list of strings representing missing values")
	cmd.Flags().StringVar(&opts.TargetCols, "target-columns", "",
		"Comma-separated list of target columns to exclude")

	// Missing data handling
	cmd.Flags().StringVar(&opts.MissingStrategy, "missing-strategy", "error",
		"Strategy for missing values: error (default), mean, median, zero, drop, native (NIPALS only)")
	cmd.Flags().Float64Var(&opts.MissingPercent, "missing-percent", 50.0,
		"Maximum missing percentage before dropping")

	// Output options
	cmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "table",
		"Output format: table, json")
	cmd.Flags().StringVarP(&opts.OutputDir, "output-dir", "o", "",
		"Output directory for results")
	cmd.Flags().BoolVar(&opts.OutputScores, "output-scores", true,
		"Include PC scores in output")
	cmd.Flags().BoolVar(&opts.OutputLoadings, "output-loadings", true,
		"Include loadings in output")
	cmd.Flags().BoolVar(&opts.OutputVariance, "output-variance", true,
		"Include explained variance in output")
	cmd.Flags().BoolVar(&opts.OutputAll, "output-all", false,
		"Output all results")
	cmd.Flags().BoolVar(&opts.IncludeMetrics, "include-metrics", false,
		"Calculate and include advanced metrics")

	// Exclude options
	cmd.Flags().StringVar(&opts.ExcludeRows, "exclude-rows", "",
		"Row indices to exclude (1-based): e.g., '1,3,5' or '1-5,8-10'")
	cmd.Flags().StringVar(&opts.ExcludeColumns, "exclude-columns", "",
		"Column indices/names to exclude: e.g., '1,3' or '1-5' or 'col1,col2'")

	// Verbose output
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false,
		"Enable verbose output")

	return cmd
}

// runAnalyze executes the analyze command
func runAnalyze(opts *AnalyzeOptions, inputFile string) error {
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

	// Parse target columns
	if opts.TargetCols != "" {
		parseOpts.TargetSuffix = "#target"
	}

	// Load CSV data with target column detection
	reader := pkgcsv.NewReader(parseOpts)
	data, err := reader.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Validate data
	if err := validateCSVData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}

	// Early detection and reporting of missing values
	selectedCols := make([]int, 0, data.Columns)
	for i := 0; i < data.Columns; i++ {
		selectedCols = append(selectedCols, i)
	}
	missingInfo := data.GetMissingValueInfo(selectedCols)

	if missingInfo.HasMissing() {
		totalValues := data.Rows * data.Columns
		missingPercent := float64(missingInfo.TotalMissing) * 100.0 / float64(totalValues)
		rowsWithMissing := len(missingInfo.RowsAffected)
		rowsPercent := float64(rowsWithMissing) * 100.0 / float64(data.Rows)

		// Report missing values to user
		if opts.Verbose {
			fmt.Printf("Missing values detected:\n")
			fmt.Printf("  Total missing: %d of %d values (%.1f%%)\n",
				missingInfo.TotalMissing, totalValues, missingPercent)
			fmt.Printf("  Rows affected: %d of %d rows (%.1f%%)\n",
				rowsWithMissing, data.Rows, rowsPercent)

			// Report by column
			if len(missingInfo.MissingByColumn) > 0 {
				fmt.Printf("  Missing by column:\n")
				for colIdx, count := range missingInfo.MissingByColumn {
					colName := fmt.Sprintf("Column %d", colIdx+1)
					if colIdx < len(data.Headers) && data.Headers[colIdx] != "" {
						colName = data.Headers[colIdx]
					}
					colPercent := float64(count) * 100.0 / float64(data.Rows)
					fmt.Printf("    %s: %d (%.1f%%)\n", colName, count, colPercent)
				}
			}
		}

		// Validate method compatibility with native strategy
		if opts.MissingStrategy == "native" {
			if strings.ToLower(opts.Method) != "nipals" {
				return fmt.Errorf("native missing value handling is only supported with the NIPALS method, not %s", opts.Method)
			}
		}

		// Check if using SVD with missing values without proper strategy
		if strings.ToLower(opts.Method) == "svd" && opts.MissingStrategy == "error" {
			return fmt.Errorf("missing values detected (%d values, %.1f%%). SVD requires complete data. "+
				"Use --missing-strategy with one of: drop, mean, median, zero. "+
				"Or use --method nipals with --missing-strategy native for native handling",
				missingInfo.TotalMissing, missingPercent)
		}
	}

	// Handle missing values based on strategy
	if missingInfo.HasMissing() && opts.MissingStrategy != "error" && opts.MissingStrategy != "native" {
		// Handle missing values based on strategy
		if opts.MissingStrategy != "drop" && opts.MissingStrategy != "mean" &&
			opts.MissingStrategy != "median" && opts.MissingStrategy != "zero" {
			return fmt.Errorf("invalid missing value strategy: %s. Valid options are: error, drop, mean, median, zero, native (NIPALS only)", opts.MissingStrategy)
		}

		if opts.Verbose {
			fmt.Printf("Applying missing value strategy: %s\n", opts.MissingStrategy)
		}

		if missingInfo.HasMissing() {
			// Handle missing values using the specified strategy
			handler := core.NewMissingValueHandler(types.MissingValueStrategy(opts.MissingStrategy))
			cleanData, err := handler.HandleMissingValues(data.Matrix, missingInfo, selectedCols)
			if err != nil {
				return fmt.Errorf("failed to handle missing values: %w", err)
			}

			// Update data matrix and affected row names for drop strategy
			if opts.MissingStrategy == "drop" && len(data.RowNames) > 0 {
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

			if opts.Verbose {
				if opts.MissingStrategy == "drop" {
					fmt.Printf("Dropped %d rows with missing values. Data now has %d rows.\n",
						len(missingInfo.RowsAffected), data.Rows)
				} else {
					fmt.Printf("Imputed %d missing values using %s strategy.\n",
						missingInfo.TotalMissing, opts.MissingStrategy)
				}
			}
		}
	} else if opts.MissingStrategy == "native" && missingInfo.HasMissing() {
		// NIPALS will handle missing values internally
		if opts.Verbose {
			fmt.Printf("NIPALS will handle %d missing values natively.\n", missingInfo.TotalMissing)
		}
	}

	// Create PCA configuration
	meanCenter := !opts.NoMeanCentering
	standardScale := opts.Scale == "standard"
	robustScale := opts.Scale == "robust"

	config := types.PCAConfig{
		Components:      opts.Components,
		Method:          opts.Method,
		MeanCenter:      meanCenter,
		StandardScale:   standardScale,
		RobustScale:     robustScale,
		ScaleOnly:       opts.ScaleOnly,
		SNV:             opts.SNV,
		VectorNorm:      opts.VectorNorm,
		MissingStrategy: types.MissingValueStrategy(opts.MissingStrategy),
	}

	// Add kernel parameters if using kernel PCA
	if opts.Method == "kernel" {
		config.KernelType = opts.KernelType
		config.KernelGamma = opts.KernelGamma
		config.KernelDegree = opts.KernelDegree
		config.KernelCoef0 = opts.KernelCoef0
	}

	// Add temporal parameters if using temporal PCA
	if opts.Method == "temporal" {
		config.TemporalLags = opts.TemporalLags
		config.VarianceExplained = opts.VarianceExplained
		config.ImputeMethod = opts.ImputeMethod

		// Validate temporal lags
		if config.TemporalLags <= 0 {
			return fmt.Errorf("temporal PCA requires --temporal-lags to be specified and positive")
		}

		// If using variance explained, don't require components
		if config.VarianceExplained > 0 {
			config.Components = 0 // Will be determined by variance explained
		}
	}

	// Parse exclude options
	if opts.ExcludeRows != "" {
		config.ExcludedRows = parseExcludeIndices(opts.ExcludeRows)
	}
	if opts.ExcludeColumns != "" {
		config.ExcludedColumns = parseExcludeColumns(opts.ExcludeColumns, data.Headers)
	}

	// Apply row and column exclusions to the data
	if len(config.ExcludedRows) > 0 || len(config.ExcludedColumns) > 0 {
		// Create a map for quick lookup of excluded rows
		excludedRowMap := make(map[int]bool)
		for _, row := range config.ExcludedRows {
			excludedRowMap[row] = true
		}

		// Create a map for quick lookup of excluded columns
		excludedColMap := make(map[int]bool)
		for _, col := range config.ExcludedColumns {
			excludedColMap[col] = true
		}

		// Filter rows
		filteredMatrix := make([][]float64, 0)
		filteredRowNames := make([]string, 0)
		for i, row := range data.Matrix {
			if !excludedRowMap[i] {
				// Filter columns from this row
				filteredRow := make([]float64, 0)
				for j, val := range row {
					if !excludedColMap[j] {
						filteredRow = append(filteredRow, val)
					}
				}
				filteredMatrix = append(filteredMatrix, filteredRow)
				if len(data.RowNames) > i {
					filteredRowNames = append(filteredRowNames, data.RowNames[i])
				}
			}
		}

		// Filter column headers
		filteredHeaders := make([]string, 0)
		for i, header := range data.Headers {
			if !excludedColMap[i] {
				filteredHeaders = append(filteredHeaders, header)
			}
		}

		// Update data with filtered results
		originalRows := data.Rows
		originalCols := data.Columns
		data.Matrix = filteredMatrix
		data.Rows = len(filteredMatrix)
		data.Columns = len(filteredHeaders)
		data.Headers = filteredHeaders
		data.RowNames = filteredRowNames

		if opts.Verbose {
			if len(config.ExcludedRows) > 0 {
				fmt.Printf("Excluded %d rows (keeping %d of %d rows)\n",
					len(config.ExcludedRows), data.Rows, originalRows)
			}
			if len(config.ExcludedColumns) > 0 {
				fmt.Printf("Excluded %d columns (keeping %d of %d columns)\n",
					len(config.ExcludedColumns), data.Columns, originalCols)
			}
		}
	}

	// Create preprocessor
	preprocessor := core.NewPreprocessorWithScaleOnly(
		config.MeanCenter,
		config.StandardScale,
		config.RobustScale,
		config.ScaleOnly,
		config.SNV,
		config.VectorNorm,
	)

	// Apply preprocessing
	processedData, err := preprocessor.FitTransform(data.Matrix)
	if err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}

	// Create and run PCA
	pca := core.NewPCAEngineForMethod(config.Method)
	result, err := pca.Fit(processedData, config)
	if err != nil {
		return fmt.Errorf("PCA analysis failed: %w", err)
	}

	// Output results based on format
	switch opts.OutputFormat {
	case "json":
		return outputJSONFormat(result, data, inputFile, opts, config, preprocessor,
			data.CategoricalColumns, data.NumericTargetColumns)
	default: // table
		outputScores := opts.OutputScores || opts.OutputAll
		outputLoadings := opts.OutputLoadings || opts.OutputAll
		outputVariance := opts.OutputVariance || opts.OutputAll
		return outputTableFormat(result, data,
			outputScores, outputLoadings, outputVariance, opts.IncludeMetrics, opts.VarianceExplained)
	}
}

// Helper functions for parsing exclude options
func parseExcludeIndices(excludeStr string) []int {
	indexSet := make(map[int]bool)
	parts := strings.Split(excludeStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			// Parse range format "start-end"
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) == 2 {
				var start, end int
				if _, err := fmt.Sscanf(rangeParts[0], "%d", &start); err == nil {
					if _, err := fmt.Sscanf(rangeParts[1], "%d", &end); err == nil {
						// Add all indices in range (inclusive)
						for i := start; i <= end; i++ {
							if i > 0 { // Ensure positive indices
								indexSet[i-1] = true // Convert to 0-based
							}
						}
					}
				}
			}
		} else {
			// Parse individual number
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx > 0 {
				indexSet[idx-1] = true // Convert to 0-based
			}
		}
	}

	// Convert map to sorted slice
	indices := make([]int, 0, len(indexSet))
	for idx := range indexSet {
		indices = append(indices, idx)
	}

	// Sort the indices
	if len(indices) > 1 {
		// Simple bubble sort for small arrays
		for i := 0; i < len(indices)-1; i++ {
			for j := 0; j < len(indices)-i-1; j++ {
				if indices[j] > indices[j+1] {
					indices[j], indices[j+1] = indices[j+1], indices[j]
				}
			}
		}
	}

	return indices
}

func parseExcludeColumns(excludeStr string, headers []string) []int {
	indexSet := make(map[int]bool)
	parts := strings.Split(excludeStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			// Parse range format "start-end" (only for numeric indices)
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) == 2 {
				var start, end int
				if _, err := fmt.Sscanf(rangeParts[0], "%d", &start); err == nil {
					if _, err := fmt.Sscanf(rangeParts[1], "%d", &end); err == nil {
						// Add all indices in range (inclusive)
						for i := start; i <= end; i++ {
							if i > 0 && i <= len(headers) { // Ensure valid range
								indexSet[i-1] = true // Convert to 0-based
							}
						}
					}
				}
			}
		} else {
			// Try to parse as index first
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx > 0 && idx <= len(headers) {
				indexSet[idx-1] = true // Convert to 0-based
			} else {
				// Try to match by name
				for i, header := range headers {
					if header == part {
						indexSet[i] = true
						break
					}
				}
			}
		}
	}

	// Convert map to sorted slice
	indices := make([]int, 0, len(indexSet))
	for idx := range indexSet {
		indices = append(indices, idx)
	}

	// Sort the indices
	if len(indices) > 1 {
		for i := 0; i < len(indices)-1; i++ {
			for j := 0; j < len(indices)-i-1; j++ {
				if indices[j] > indices[j+1] {
					indices[j], indices[j+1] = indices[j+1], indices[j]
				}
			}
		}
	}

	return indices
}
