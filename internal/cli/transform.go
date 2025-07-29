package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	cli "github.com/urfave/cli/v2"
)

func transformCommand() *cli.Command {
	return &cli.Command{
		Name:      "transform",
		Usage:     "Apply an existing PCA model to new data",
		ArgsUsage: "<model.json> <input.csv>",
		Description: `The transform command applies a trained PCA model to new data.

USAGE:
  gopca-cli transform [OPTIONS] <model.json> <input.csv>

  The model JSON file and input CSV file must be specified as the last two arguments.
  All options must come BEFORE the filenames.

EXAMPLES:
  # Basic transformation
  gopca-cli transform model.json new_data.csv

  # Save results to specific file
  gopca-cli transform -o transformed_scores.csv model.json new_data.csv

  # JSON output format
  gopca-cli transform -f json -o results/ model.json new_data.csv

  # Exclude specific rows from new data
  gopca-cli transform --exclude-rows 1,5-10 model.json new_data.csv

NOTES:
  - The new data must have the same number of features as the training data
  - Column names should match for proper alignment (warning issued if different)
  - Preprocessing parameters from training are automatically applied
  - Currently supports SVD and NIPALS models (kernel PCA under development)`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory for results",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: table or json",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show detailed progress",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress all output except errors",
			},
			&cli.BoolFlag{
				Name:  "no-headers",
				Usage: "Input CSV has no header row",
			},
			&cli.BoolFlag{
				Name:  "no-index",
				Usage: "Input CSV has no index column",
			},
			&cli.StringFlag{
				Name:    "delimiter",
				Aliases: []string{"d"},
				Usage:   "CSV delimiter character",
				Value:   ",",
			},
			&cli.StringFlag{
				Name:  "decimal-separator",
				Usage: "Decimal separator: 'dot' or 'comma'",
				Value: "dot",
			},
			&cli.StringFlag{
				Name:  "na-values",
				Usage: "Comma-separated list of strings to treat as NA/missing",
				Value: "NA,NaN,null,NULL,n/a,N/A",
			},
			&cli.StringFlag{
				Name:  "exclude-rows",
				Usage: "Comma-separated list of row indices to exclude (1-based)",
			},
			&cli.BoolFlag{
				Name:  "include-metrics",
				Usage: "Calculate and include diagnostic metrics in output",
			},
		},
		Action: runTransform,
		Before: validateTransformFlags,
	}
}

func validateTransformFlags(c *cli.Context) error {
	// Validate verbose and quiet flags
	if c.Bool("verbose") && c.Bool("quiet") {
		return fmt.Errorf("cannot use both --verbose and --quiet flags")
	}

	// Validate arguments
	if c.NArg() < 2 {
		return fmt.Errorf("missing required arguments: model.json and input.csv")
	}

	// Validate format
	format := c.String("format")
	switch format {
	case "table", "json":
		// Valid formats
	default:
		return fmt.Errorf("invalid output format: %s (must be table or json)", format)
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

	return nil
}

func runTransform(c *cli.Context) error {
	modelFile := c.Args().Get(0)
	inputFile := c.Args().Get(1)
	verbose := c.Bool("verbose")
	quiet := c.Bool("quiet")

	// Load the model
	if verbose {
		fmt.Printf("Loading model from %s...\n", modelFile)
	}

	modelData, err := loadPCAModel(modelFile)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	// Validate model
	if err := validateModel(modelData); err != nil {
		return fmt.Errorf("invalid model: %w", err)
	}

	if verbose {
		fmt.Printf("✓ Model loaded successfully\n")
		fmt.Printf("  - Method: %s\n", modelData.Metadata.Config.Method)
		fmt.Printf("  - Components: %d\n", modelData.Metadata.Config.NComponents)
		fmt.Printf("  - Features: %d\n", len(modelData.Model.FeatureLabels))
	}

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
	switch decimalSep {
	case "dot":
		parseOpts.DecimalSeparator = '.'
	case "comma", ",":
		parseOpts.DecimalSeparator = ','
	default:
		parseOpts.DecimalSeparator = rune(decimalSep[0])
	}

	// Parse NA values
	if naValues := c.String("na-values"); naValues != "" {
		parseOpts.NullValues = strings.Split(naValues, ",")
		for i := range parseOpts.NullValues {
			parseOpts.NullValues[i] = strings.TrimSpace(parseOpts.NullValues[i])
		}
	}

	// Load new data
	if verbose {
		fmt.Printf("\nLoading data from %s...\n", inputFile)
	}

	data, err := ParseCSV(inputFile, parseOpts)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Validate data
	if err := ValidateCSVData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}

	// Handle excluded rows
	if excludeStr := c.String("exclude-rows"); excludeStr != "" {
		excludedRows, err := utils.ParseRanges(excludeStr)
		if err != nil {
			return fmt.Errorf("invalid exclude-rows format: %w", err)
		}
		// Filter data matrix
		filteredData := [][]float64{}
		filteredRowNames := []string{}
		for i, row := range data.Matrix {
			// Convert to 1-based index for comparison
			if !contains(excludedRows, i+1) {
				filteredData = append(filteredData, row)
				if i < len(data.RowNames) {
					filteredRowNames = append(filteredRowNames, data.RowNames[i])
				}
			}
		}
		data.Matrix = filteredData
		data.RowNames = filteredRowNames

		if verbose {
			fmt.Printf("  - Excluded %d rows\n", len(excludedRows))
		}
	}

	// Validate feature count
	expectedFeatures := len(modelData.Model.FeatureLabels)
	actualFeatures := len(data.Headers)
	if actualFeatures != expectedFeatures {
		return fmt.Errorf("feature count mismatch: model expects %d features, but data has %d",
			expectedFeatures, actualFeatures)
	}

	// Check feature names match (warning only)
	if verbose && !quiet {
		mismatchedFeatures := []string{}
		for i, expected := range modelData.Model.FeatureLabels {
			if i < len(data.Headers) && data.Headers[i] != expected {
				mismatchedFeatures = append(mismatchedFeatures,
					fmt.Sprintf("%s (expected) != %s (actual)", expected, data.Headers[i]))
			}
		}
		if len(mismatchedFeatures) > 0 {
			fmt.Println("\nWarning: Feature names don't match:")
			for _, mismatch := range mismatchedFeatures {
				fmt.Printf("  - %s\n", mismatch)
			}
			fmt.Println("  Proceeding with transformation based on column order...")
		}
	}

	if verbose {
		fmt.Printf("\n✓ Data loaded successfully\n")
		fmt.Printf("  - Samples: %d\n", len(data.RowNames))
		fmt.Printf("  - Features: %d\n", len(data.Headers))
	}

	// Apply transformation
	if verbose {
		fmt.Println("\nApplying PCA transformation...")
	}

	scores, err := transformData(modelData, data.Matrix)
	if err != nil {
		return fmt.Errorf("transformation failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ Transformation completed successfully")
	}

	// Create result structure for output
	result := createTransformResult(scores, data.RowNames, modelData)

	// Calculate metrics if requested
	if c.Bool("include-metrics") && verbose {
		fmt.Println("\nCalculating diagnostic metrics...")
	}

	// Handle output
	outputFormat := c.String("format")
	outputDir := c.String("output-dir")
	includeMetrics := c.Bool("include-metrics")

	switch outputFormat {
	case "table":
		if !quiet {
			err = outputTransformTable(result, includeMetrics)
		}
	case "json":
		err = outputTransformJSON(result, inputFile, outputDir, includeMetrics, modelData)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("output failed: %w", err)
	}

	return nil
}

// loadPCAModel loads a PCA model from a JSON file
func loadPCAModel(filename string) (*types.PCAOutputData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	var model types.PCAOutputData
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse model JSON: %w", err)
	}

	return &model, nil
}

// validateModel checks if the model has all required fields
func validateModel(model *types.PCAOutputData) error {
	if len(model.Model.Loadings) == 0 {
		return fmt.Errorf("model missing loadings matrix")
	}

	if model.Metadata.Config.Method == "" {
		return fmt.Errorf("model missing method information")
	}

	// Check for kernel PCA (not yet supported)
	if model.Metadata.Config.Method == "kernel" {
		return fmt.Errorf("kernel PCA models are not yet supported for transformation")
	}

	return nil
}

// transformData applies the PCA transformation to new data
func transformData(model *types.PCAOutputData, data types.Matrix) (types.Matrix, error) {
	// Create PCA engine
	engine := core.NewPCAEngine()

	// Reconstruct the engine from the model
	if err := reconstructEngine(engine, model); err != nil {
		return nil, fmt.Errorf("failed to reconstruct PCA engine: %w", err)
	}

	// Apply transformation
	scores, err := engine.Transform(data)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	return scores, nil
}

// reconstructEngine rebuilds the PCA engine from a saved model
func reconstructEngine(engine types.PCAEngine, model *types.PCAOutputData) error {
	// We need to use the concrete type to access setter methods
	pcaImpl, ok := engine.(*core.PCAImpl)
	if !ok {
		return fmt.Errorf("engine is not a PCAImpl instance")
	}

	// Set up preprocessing if needed
	prepInfo := model.Preprocessing
	if prepInfo.MeanCenter || prepInfo.StandardScale || prepInfo.RobustScale ||
		prepInfo.ScaleOnly || prepInfo.SNV || prepInfo.VectorNorm {
		preprocessor := core.NewPreprocessorWithScaleOnly(
			prepInfo.MeanCenter,
			prepInfo.StandardScale,
			prepInfo.RobustScale,
			prepInfo.ScaleOnly,
			prepInfo.SNV,
			prepInfo.VectorNorm,
		)

		// Set the fitted parameters
		params := prepInfo.Parameters
		if err := preprocessor.SetFittedParameters(
			params.FeatureMeans,
			params.FeatureStdDevs,
			params.FeatureMedians,
			params.FeatureMADs,
			params.RowMeans,
			params.RowStdDevs,
		); err != nil {
			return fmt.Errorf("failed to set preprocessing parameters: %w", err)
		}

		pcaImpl.SetPreprocessor(preprocessor)
	}

	// Set the loadings and configuration
	if err := pcaImpl.SetLoadings(model.Model.Loadings, model.Metadata.Config.NComponents); err != nil {
		return fmt.Errorf("failed to set loadings: %w", err)
	}

	return nil
}

// createTransformResult creates a result structure for output
func createTransformResult(scores types.Matrix, sampleNames []string, model *types.PCAOutputData) *TransformResult {
	return &TransformResult{
		Scores:          scores,
		SampleNames:     sampleNames,
		ComponentLabels: model.Model.ComponentLabels,
		ModelMethod:     model.Metadata.Config.Method,
	}
}

// TransformResult holds the transformation results
type TransformResult struct {
	Scores          types.Matrix
	SampleNames     []string
	ComponentLabels []string
	ModelMethod     string
	Metrics         []types.SampleMetrics // Optional metrics
}

// outputTransformTable outputs results in table format
func outputTransformTable(result *TransformResult, includeMetrics bool) error {
	fmt.Println("\nPCA Transformation Results:")
	fmt.Println("===========================")

	// Output scores
	fmt.Printf("\nScores (first 10 samples):\n")
	fmt.Printf("%-15s", "Sample")
	for _, label := range result.ComponentLabels {
		fmt.Printf("%12s", label)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 15+12*len(result.ComponentLabels)))

	// Show preview of samples
	const previewMaxRows = 10
	maxSamples := len(result.SampleNames)
	if maxSamples > previewMaxRows {
		maxSamples = previewMaxRows
	}

	for i := 0; i < maxSamples; i++ {
		fmt.Printf("%-15s", result.SampleNames[i])
		for j := 0; j < len(result.ComponentLabels); j++ {
			fmt.Printf("%12.4f", result.Scores[i][j])
		}
		fmt.Println()
	}

	if len(result.SampleNames) > 10 {
		fmt.Printf("... (%d more samples)\n", len(result.SampleNames)-10)
	}

	return nil
}

// outputTransformJSON outputs results in JSON format
func outputTransformJSON(result *TransformResult, inputFile, outputDir string,
	includeMetrics bool, model *types.PCAOutputData) error {

	// Create output structure
	output := struct {
		Metadata struct {
			TransformedAt string `json:"transformed_at"`
			InputFile     string `json:"input_file"`
			ModelMethod   string `json:"model_method"`
			NSamples      int    `json:"n_samples"`
			NComponents   int    `json:"n_components"`
		} `json:"metadata"`
		Results []struct {
			ID     string             `json:"id"`
			Scores map[string]float64 `json:"scores"`
		} `json:"results"`
	}{}

	// Set metadata
	output.Metadata.TransformedAt = os.Getenv("TZ")
	output.Metadata.InputFile = inputFile
	output.Metadata.ModelMethod = result.ModelMethod
	output.Metadata.NSamples = len(result.SampleNames)
	output.Metadata.NComponents = len(result.ComponentLabels)

	// Convert scores to output format
	output.Results = make([]struct {
		ID     string             `json:"id"`
		Scores map[string]float64 `json:"scores"`
	}, len(result.SampleNames))

	for i, name := range result.SampleNames {
		output.Results[i].ID = name
		output.Results[i].Scores = make(map[string]float64)
		for j, label := range result.ComponentLabels {
			output.Results[i].Scores[label] = result.Scores[i][j]
		}
	}

	// Generate output path
	paths := generateOutputPaths(inputFile, outputDir, "json")
	outputFile := paths["output"]

	// Modify filename to indicate it's a transformation
	outputFile = strings.Replace(outputFile, "_pca.json", "_transformed.json", 1)

	// Create output directory if needed
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Marshal JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write output
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("\nResults saved to: %s\n", outputFile)
	return nil
}
