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
)

// outputTableFormat outputs PCA results in table format
func outputTableFormat(result *types.PCAResult, data *pkgcsv.Data,
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {

	// Calculate metrics if requested (skip for kernel PCA as it doesn't have loadings)
	var metrics []types.SampleMetrics
	if includeMetrics && outputScores {
		if result.Method != "kernel" {
			var err error
			metrics, err = core.CalculateMetricsFromPCAResult(result, data.Matrix)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to calculate metrics: %v\n", err)
				// Create placeholder metrics
				metrics = make([]types.SampleMetrics, len(result.Scores))
			}
		} else {
			// For kernel PCA, use metrics from result if available
			if len(result.Metrics) > 0 {
				metrics = result.Metrics
			} else {
				// Create empty metrics for kernel PCA
				metrics = make([]types.SampleMetrics, len(result.Scores))
			}
		}
	}

	// Output scores table
	if outputScores {
		fmt.Println("\nPCA Scores:")
		fmt.Println("──────────────────────────────────────────────────────────────")

		// Print headers
		fmt.Printf("%-15s", "Sample")
		for i := 0; i < len(result.ComponentLabels); i++ {
			fmt.Printf("%12s", result.ComponentLabels[i])
		}
		if includeMetrics {
			fmt.Printf("%15s%18s%10s%10s", "Hotelling T²", "Mahalanobis Dist", "RSS", "Outlier")
		}
		fmt.Println()
		fmt.Println("──────────────────────────────────────────────────────────────")

		// Add data rows (show first 20 and last 5 for large datasets)
		nRows := len(result.Scores)
		rowsToShow := nRows
		if nRows > 25 {
			rowsToShow = 25
		}

		for i := 0; i < rowsToShow; i++ {
			rowIdx := i
			if i >= 20 && nRows > 25 {
				// Skip to last 5 rows
				rowIdx = nRows - (25 - i)
				if i == 20 {
					// Add ellipsis row
					fmt.Printf("%-15s", "...")
					for j := 0; j < len(result.ComponentLabels); j++ {
						fmt.Printf("%12s", "...")
					}
					if includeMetrics {
						fmt.Printf("%15s%18s%10s%10s", "...", "...", "...", "...")
					}
					fmt.Println()
				}
			}

			// Sample ID
			sampleID := fmt.Sprintf("Sample_%d", rowIdx+1)
			if rowIdx < len(data.RowNames) {
				sampleID = data.RowNames[rowIdx]
			}
			fmt.Printf("%-15s", sampleID)

			// PC scores
			for j := 0; j < len(result.ComponentLabels); j++ {
				fmt.Printf("%12.4f", result.Scores[rowIdx][j])
			}

			// Metrics
			if includeMetrics && metrics != nil {
				metric := metrics[rowIdx]
				outlierStr := "False"
				if metric.IsOutlier {
					outlierStr = "True"
				}
				fmt.Printf("%15.4f%18.4f%10.4f%10s",
					metric.HotellingT2, metric.Mahalanobis, metric.RSS, outlierStr)
			}

			fmt.Println()
		}

		if nRows > 25 {
			fmt.Printf("\nShowing first 20 and last 5 of %d samples\n", nRows)
		}
	}

	// Output loadings table (skip for kernel PCA which doesn't have loadings)
	if outputLoadings {
		if result.Method != "kernel" {
			fmt.Println("\nPCA Loadings:")
			fmt.Println("──────────────────────────────────────────────────────────────")

			// Print headers
			fmt.Printf("%-25s", "Variable")
			for i := 0; i < len(result.ComponentLabels); i++ {
				fmt.Printf("%12s", result.ComponentLabels[i])
			}
			fmt.Println()
			fmt.Println("──────────────────────────────────────────────────────────────")

			// Add loading rows (show first 20 and last 5 for large datasets)
			nFeatures := len(data.Headers)
			featuresToShow := nFeatures
			if nFeatures > 25 {
				featuresToShow = 25
			}

			for i := 0; i < featuresToShow; i++ {
				featureIdx := i
				if i >= 20 && nFeatures > 25 {
					// Skip to last 5 features
					featureIdx = nFeatures - (25 - i)
					if i == 20 {
						// Add ellipsis row
						fmt.Printf("%-25s", "...")
						for j := 0; j < len(result.ComponentLabels); j++ {
							fmt.Printf("%12s", "...")
						}
						fmt.Println()
					}
				}

				fmt.Printf("%-25s", data.Headers[featureIdx])
				for j := 0; j < len(result.ComponentLabels); j++ {
					fmt.Printf("%12.4f", result.Loadings[featureIdx][j])
				}
				fmt.Println()
			}

			if nFeatures > 25 {
				fmt.Printf("\nShowing first 20 and last 5 of %d features\n", nFeatures)
			}
		} else {
			fmt.Println("\nNote: Loadings are not available for Kernel PCA")
		}
	}

	// Output variance table
	if outputVariance {
		fmt.Println("\nExplained Variance:")
		fmt.Println("──────────────────────────────────────────────────────────────")
		fmt.Printf("%-15s%15s%15s\n", "Component", "Variance", "Cumulative")
		fmt.Println("──────────────────────────────────────────────────────────────")

		for i := 0; i < len(result.ComponentLabels); i++ {
			fmt.Printf("%-15s%14.1f%%%14.1f%%\n",
				result.ComponentLabels[i],
				result.ExplainedVar[i],
				result.CumulativeVar[i])
		}
	}

	// Output diagnostic limits if available
	if includeMetrics && (result.T2Limit95 > 0 || result.QLimit95 > 0) {
		fmt.Println("\nDiagnostic Confidence Limits:")
		fmt.Println("──────────────────────────────────────────────────────────────")
		fmt.Printf("%-30s%20s%20s\n", "Metric", "95% Limit", "99% Limit")
		fmt.Println("──────────────────────────────────────────────────────────────")

		if result.T2Limit95 > 0 {
			fmt.Printf("%-30s%20.4f%20.4f\n", "Hotelling's T²", result.T2Limit95, result.T2Limit99)
		}
		if result.QLimit95 > 0 {
			fmt.Printf("%-30s%20.4f%20.4f\n", "Q-residuals (SPE)", result.QLimit95, result.QLimit99)
		}
	}

	return nil
}

// outputJSONFormat outputs PCA results in JSON format
func outputJSONFormat(result *types.PCAResult, data *pkgcsv.Data, inputFile string,
	opts *AnalyzeOptions, config types.PCAConfig, preprocessor *core.Preprocessor,
	categoricalData map[string][]string, targetData map[string][]float64) error {

	// Convert to PCAOutputData
	outputData := pkgcsv.ConvertToPCAOutputData(result, data, opts.IncludeMetrics,
		config, preprocessor, categoricalData, targetData)

	// Generate output paths
	outputFile := generateOutputPath(inputFile, opts.OutputDir, "_pca.json")

	// Create output directory if needed
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Marshal JSON
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
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

// generateOutputPath creates an output file path based on input file and format
func generateOutputPath(inputFile, outputDir, suffix string) string {
	// Get the directory and base name of the input file
	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)

	// Remove extension to get the base name
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	// Use output directory if specified, otherwise use input directory
	if outputDir != "" {
		dir = outputDir
	}

	return filepath.Join(dir, baseName+suffix)
}
