// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/bitjungle/gopca/pkg/validation"
	"github.com/spf13/cobra"
)

// TransformOptions holds all the options for the transform command
type TransformOptions struct {
	// Output options
	OutputFormat string
	OutputDir    string

	// Data format options
	NoHeaders bool
	NoIndex   bool
	Delimiter string
	NAValues  string
}

// NewTransformCommand creates the transform subcommand
func NewTransformCommand() *cobra.Command {
	opts := &TransformOptions{}

	cmd := &cobra.Command{
		Use:   "transform [flags] <model.json> <input.csv>",
		Short: "Transform new data using a trained PCA model",
		Long: `Transform new data using a previously trained PCA model.

The transform command applies a saved PCA model to new data, projecting
it into the principal component space. The model must be in JSON format
from a previous analyze command.

EXAMPLES:
  # Transform new data using saved model
  pca transform model.json new_data.csv

  # Transform and save to specific directory
  pca transform -o results/ model.json new_data.csv

  # Transform data with different CSV format
  pca transform --delimiter ";" model.json data.csv`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransform(opts, args[0], args[1])
		},
	}

	// Output options
	cmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "table",
		"Output format: table, json")
	cmd.Flags().StringVarP(&opts.OutputDir, "output", "o", "",
		"Output directory for results")

	// Data format options
	cmd.Flags().BoolVar(&opts.NoHeaders, "no-headers", false,
		"First row contains data, not column names")
	cmd.Flags().BoolVar(&opts.NoIndex, "no-index", false,
		"First column contains data, not row names")
	cmd.Flags().StringVar(&opts.Delimiter, "delimiter", ",",
		"CSV field delimiter")
	cmd.Flags().StringVar(&opts.NAValues, "na-values", ",NA,N/A,nan,NaN,null,NULL,m",
		"Comma-separated list of strings representing missing values")

	return cmd
}

// runTransform executes the transform command
func runTransform(opts *TransformOptions, modelFile, inputFile string) error {
	// Load the PCA model
	modelData, err := os.ReadFile(modelFile)
	if err != nil {
		return fmt.Errorf("failed to read model file: %w", err)
	}

	// Validate model against schema
	validator, err := validation.NewModelValidator("v1")
	if err != nil {
		// Schema validation not available, continue without validation
		fmt.Fprintf(os.Stderr, "Warning: Schema validation not available: %v\n", err)
	} else {
		if err := validator.ValidateModel(modelData); err != nil {
			return fmt.Errorf("model validation failed: %w", err)
		}
	}

	var pcaOutputData types.PCAOutputData
	if err := json.Unmarshal(modelData, &pcaOutputData); err != nil {
		return fmt.Errorf("failed to parse model JSON: %w", err)
	}

	// Parse CSV options
	parseOpts := pkgcsv.DefaultOptions()
	parseOpts.HasHeaders = !opts.NoHeaders
	parseOpts.HasRowNames = !opts.NoIndex
	parseOpts.Delimiter = rune(opts.Delimiter[0])
	parseOpts.ParseMode = pkgcsv.ParseMixed

	// Parse NA values
	if opts.NAValues != "" {
		parseOpts.NullValues = strings.Split(opts.NAValues, ",")
		for i := range parseOpts.NullValues {
			parseOpts.NullValues[i] = strings.TrimSpace(parseOpts.NullValues[i])
		}
	}

	// Load new data
	reader := pkgcsv.NewReader(parseOpts)
	data, err := reader.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Validate data
	if err := validateCSVData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}

	// Check that the number of features matches
	if len(data.Headers) != len(pcaOutputData.Model.FeatureLabels) {
		return fmt.Errorf("feature count mismatch: model has %d features, data has %d",
			len(pcaOutputData.Model.FeatureLabels), len(data.Headers))
	}

	// Create preprocessor from saved parameters
	preprocessor := core.NewPreprocessorWithScaleOnly(
		pcaOutputData.Preprocessing.MeanCenter,
		pcaOutputData.Preprocessing.StandardScale,
		pcaOutputData.Preprocessing.RobustScale,
		pcaOutputData.Preprocessing.ScaleOnly,
		pcaOutputData.Preprocessing.SNV,
		pcaOutputData.Preprocessing.VectorNorm,
	)

	// Restore preprocessing parameters
	if err := preprocessor.SetFittedParameters(
		pcaOutputData.Preprocessing.Parameters.FeatureMeans,
		pcaOutputData.Preprocessing.Parameters.FeatureStdDevs,
		pcaOutputData.Preprocessing.Parameters.FeatureMedians,
		pcaOutputData.Preprocessing.Parameters.FeatureMADs,
		pcaOutputData.Preprocessing.Parameters.RowMeans,
		pcaOutputData.Preprocessing.Parameters.RowStdDevs,
	); err != nil {
		return fmt.Errorf("failed to restore preprocessing parameters: %w", err)
	}

	// Apply preprocessing
	processedData, err := preprocessor.Transform(data.Matrix)
	if err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}

	// Project data using loadings
	scores := ProjectData(processedData, pcaOutputData.Model.Loadings)

	// Create result structure
	result := &types.PCAResult{
		Scores:          scores,
		Loadings:        pcaOutputData.Model.Loadings,
		ExplainedVar:    pcaOutputData.Model.ExplainedVariance,
		CumulativeVar:   pcaOutputData.Model.CumulativeVariance,
		ComponentLabels: pcaOutputData.Model.ComponentLabels,
		Method:          pcaOutputData.Metadata.Config.Method,
	}

	// Output results based on format
	switch opts.OutputFormat {
	case "json":
		return outputTransformJSON(result, data, inputFile, opts.OutputDir)
	default: // table
		return outputTransformTable(result, data)
	}
}

// Output functions for transform command
func outputTransformTable(result *types.PCAResult, data *pkgcsv.Data) error {
	fmt.Println("\nTransformed Scores:")
	fmt.Println("──────────────────────────────────────────────────────────────")

	// Print headers
	fmt.Printf("%-15s", "Sample")
	for i := 0; i < len(result.ComponentLabels); i++ {
		fmt.Printf("%12s", result.ComponentLabels[i])
	}
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────")

	// Print scores
	for i := 0; i < len(result.Scores); i++ {
		sampleID := fmt.Sprintf("Sample_%d", i+1)
		if i < len(data.RowNames) {
			sampleID = data.RowNames[i]
		}
		fmt.Printf("%-15s", sampleID)

		for j := 0; j < len(result.ComponentLabels); j++ {
			fmt.Printf("%12.4f", result.Scores[i][j])
		}
		fmt.Println()
	}

	return nil
}

func outputTransformJSON(result *types.PCAResult, data *pkgcsv.Data,
	inputFile, outputDir string) error {
	// Generate output path
	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	if outputDir != "" {
		dir = outputDir
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	outputFile := filepath.Join(dir, baseName+"_transformed.json")

	// Create output structure
	type TransformOutput struct {
		Samples []struct {
			ID     string             `json:"id"`
			Scores map[string]float64 `json:"scores"`
		} `json:"samples"`
	}

	var output TransformOutput
	for i := 0; i < len(result.Scores); i++ {
		sampleID := fmt.Sprintf("Sample_%d", i+1)
		if i < len(data.RowNames) {
			sampleID = data.RowNames[i]
		}

		scores := make(map[string]float64)
		for j := 0; j < len(result.ComponentLabels); j++ {
			scores[result.ComponentLabels[j]] = result.Scores[i][j]
		}

		output.Samples = append(output.Samples, struct {
			ID     string             `json:"id"`
			Scores map[string]float64 `json:"scores"`
		}{
			ID:     sampleID,
			Scores: scores,
		})
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
