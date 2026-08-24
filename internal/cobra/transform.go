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
	parsedDelim, delimErr := parseDelimiter(opts.Delimiter)
	if delimErr != nil {
		return fmt.Errorf("invalid delimiter: %w", delimErr)
	}
	parseOpts.Delimiter = parsedDelim
	// Use ParseMixedWithTargets to properly identify and exclude target columns
	parseOpts.ParseMode = pkgcsv.ParseMixedWithTargets

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

	// Extract feature columns that match the model's feature labels
	// This handles cases where target columns are present in the data
	modelFeatures := pcaOutputData.Model.FeatureLabels

	// Create a map for quick lookup of model feature indices
	modelFeatureMap := make(map[string]int)
	for i, label := range modelFeatures {
		modelFeatureMap[label] = i
	}

	// Find indices of data columns that match model features
	dataColumnIndices := make([]int, 0, len(modelFeatures))
	missingFeatures := make([]string, 0)

	for _, modelFeature := range modelFeatures {
		found := false
		for j, dataHeader := range data.Headers {
			if dataHeader == modelFeature {
				dataColumnIndices = append(dataColumnIndices, j)
				found = true
				break
			}
		}
		if !found {
			missingFeatures = append(missingFeatures, modelFeature)
		}
	}

	// Check if all required features are present
	if len(missingFeatures) > 0 {
		return fmt.Errorf("missing required features in data: %v", missingFeatures)
	}

	// Filter the data matrix to include only the feature columns in the correct order
	filteredMatrix := make([][]float64, len(data.Matrix))
	for i := range data.Matrix {
		filteredMatrix[i] = make([]float64, len(dataColumnIndices))
		for j, colIdx := range dataColumnIndices {
			filteredMatrix[i][j] = data.Matrix[i][colIdx]
		}
	}

	// Update data structure with filtered matrix
	data.Matrix = filteredMatrix
	data.Columns = len(modelFeatures)
	data.Headers = modelFeatures

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

	// Some methods cannot project new data from the model file alone, and the
	// shape of their stored loadings would otherwise be misread (#809).
	if err := checkTransformSupported(pcaOutputData.Metadata.Config.Method); err != nil {
		return err
	}

	// Project data using loadings
	scores, err := ProjectData(processedData, pcaOutputData.Model.Loadings)
	if err != nil {
		return fmt.Errorf("projection failed: %w", err)
	}

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
	fmt.Printf("%-15s", "Sample_ID")
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
		if err := os.MkdirAll(outputDir, 0750); err != nil {
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

// checkTransformSupported reports whether a model of the given method can
// project new data from the model file alone.
//
// Kernel PCA projects through the kernel evaluated between new samples and the
// training set, so it needs the training data, which the model file does not
// carry. Temporal PCA's loadings live over a lagged embedding, so new data would
// have to be re-embedded with the same lag structure first. Neither can be
// approximated by multiplying by the stored loadings, so both are refused rather
// than given a plausible but wrong answer.
func checkTransformSupported(method string) error {
	switch strings.ToLower(method) {
	case "kernel":
		return fmt.Errorf("this model was fitted with kernel PCA, which cannot transform new data: " +
			"projection requires the original training data and the kernel function, which the model file does not store. " +
			"Run 'pca analyze' over the combined data instead")
	case "temporal":
		return fmt.Errorf("this model was fitted with temporal PCA, which cannot transform new data: " +
			"its loadings describe a lagged embedding rather than the input variables, so new data must be re-embedded. " +
			"Run 'pca analyze' over the combined series instead")
	}
	return nil
}
