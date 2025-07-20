package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitjungle/complab/internal/core"
	"github.com/bitjungle/complab/pkg/types"
	"github.com/spf13/cobra"
)

var (
	inputFile     string
	outputFile    string
	components    int
	meanCenter    bool
	standardScale bool
	method        string
	outputFormat  string
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Perform PCA analysis on input data",
	Long: `Analyze performs Principal Component Analysis on your input data.

Examples:
  # Basic PCA with 2 components
  complab-cli analyze -i data.csv -o results.csv

  # PCA with 3 components and standard scaling
  complab-cli analyze -i data.csv -o results.csv -c 3 --standard-scale

  # Use SVD method and output as JSON
  complab-cli analyze -i data.csv -o results.json -m svd -f json`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Required flags
	analyzeCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input CSV file (required)")
	analyzeCmd.MarkFlagRequired("input")

	// Optional flags
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	analyzeCmd.Flags().IntVarP(&components, "components", "c", 2, "Number of principal components")
	analyzeCmd.Flags().BoolVar(&meanCenter, "mean-center", true, "Apply mean centering to data")
	analyzeCmd.Flags().BoolVar(&standardScale, "standard-scale", false, "Apply standard scaling to data")
	analyzeCmd.Flags().StringVarP(&method, "method", "m", "nipals", "PCA method: nipals or svd")
	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "csv", "Output format: csv, json, or tsv")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Validate method
	if method != "nipals" && method != "svd" {
		return fmt.Errorf("invalid method: %s (must be 'nipals' or 'svd')", method)
	}

	// Validate output format
	if outputFormat != "csv" && outputFormat != "json" && outputFormat != "tsv" {
		return fmt.Errorf("invalid format: %s (must be 'csv', 'json', or 'tsv')", outputFormat)
	}

	// Log start if verbose
	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "Loading data from %s...\n", inputFile)
	}

	// Load data
	data, headers, hasRowNames, err := detectAndLoadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	
	if hasRowNames && verbose && !quiet {
		fmt.Fprintf(os.Stderr, "Detected row names in first column, skipping...\n")
	}

	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "Loaded data: %d rows × %d columns\n", len(data), len(data[0]))
	}

	// Note: Preprocessing will be handled by the PCA engine based on config

	// Create PCA config
	config := types.PCAConfig{
		Components:    components,
		MeanCenter:    meanCenter,
		StandardScale: standardScale,
		Method:        method,
	}

	// Run PCA
	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "Running PCA with %s method...\n", method)
	}

	pca := core.NewPCAEngine()
	result, err := pca.FitTransform(data, config)
	if err != nil {
		return fmt.Errorf("PCA analysis failed: %w", err)
	}

	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "PCA completed. Cumulative explained variance: %.2f%%\n", 
			result.CumulativeVar[len(result.CumulativeVar)-1])
	}

	// Set component labels if not present
	if len(result.ComponentLabels) == 0 {
		result.ComponentLabels = make([]string, components)
		for i := 0; i < components; i++ {
			result.ComponentLabels[i] = fmt.Sprintf("PC%d", i+1)
		}
	}

	// Output results
	var output *os.File
	if outputFile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	}

	switch outputFormat {
	case "json":
		err = outputJSON(output, result, headers)
	case "csv":
		err = outputCSV(output, result, headers)
	case "tsv":
		err = outputTSV(output, result, headers)
	}

	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if verbose && !quiet && outputFile != "" {
		fmt.Fprintf(os.Stderr, "Results saved to %s\n", outputFile)
	}

	return nil
}

func outputJSON(w *os.File, result *types.PCAResult, headers []string) error {
	// Create a structure that includes headers
	output := struct {
		Headers []string `json:"headers"`
		*types.PCAResult
	}{
		Headers:    headers,
		PCAResult: result,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputCSV(w *os.File, result *types.PCAResult, headers []string) error {
	return outputDelimited(w, result, headers, ",")
}

func outputTSV(w *os.File, result *types.PCAResult, headers []string) error {
	return outputDelimited(w, result, headers, "\t")
}

func outputDelimited(w *os.File, result *types.PCAResult, headers []string, delimiter string) error {
	// Write scores section
	fmt.Fprintln(w, "# SCORES")
	
	// Write header
	fmt.Fprint(w, "Sample")
	for _, label := range result.ComponentLabels {
		fmt.Fprintf(w, "%s%s", delimiter, label)
	}
	fmt.Fprintln(w)

	// Write scores
	for i, row := range result.Scores {
		fmt.Fprintf(w, "Sample_%d", i+1)
		for _, val := range row {
			fmt.Fprintf(w, "%s%.6f", delimiter, val)
		}
		fmt.Fprintln(w)
	}

	// Write loadings section
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# LOADINGS")
	
	// Write header
	fmt.Fprint(w, "Variable")
	for _, label := range result.ComponentLabels {
		fmt.Fprintf(w, "%s%s", delimiter, label)
	}
	fmt.Fprintln(w)

	// Write loadings
	for i, row := range result.Loadings {
		varName := fmt.Sprintf("Var_%d", i+1)
		if i < len(headers) {
			varName = headers[i]
		}
		fmt.Fprint(w, varName)
		for _, val := range row {
			fmt.Fprintf(w, "%s%.6f", delimiter, val)
		}
		fmt.Fprintln(w)
	}

	// Write variance explained section
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# VARIANCE EXPLAINED")
	fmt.Fprintf(w, "Component%sVariance%sCumulative\n", delimiter, delimiter)
	for i, variance := range result.ExplainedVar {
		fmt.Fprintf(w, "%s%s%.6f%s%.6f\n", 
			result.ComponentLabels[i], delimiter, 
			variance, delimiter, 
			result.CumulativeVar[i])
	}

	return nil
}