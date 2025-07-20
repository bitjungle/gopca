package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// OutputData represents the complete output data structure
type OutputData struct {
	Metadata OutputMetadata  `json:"metadata"`
	Results  []SampleResult  `json:"results"`
	Summary  OutputSummary   `json:"summary,omitempty"`
}

// OutputMetadata contains analysis metadata
type OutputMetadata struct {
	NSamples      int    `json:"n_samples"`
	NFeatures     int    `json:"n_features"`
	NComponents   int    `json:"n_components"`
	Preprocessing string `json:"preprocessing"`
}

// SampleResult contains results for a single sample
type SampleResult struct {
	ID      string             `json:"id"`
	Scores  map[string]float64 `json:"scores,omitempty"`
	Metrics *SampleMetrics     `json:"metrics,omitempty"`
}

// SampleMetrics contains advanced metrics for a sample
type SampleMetrics struct {
	HotellingT2         float64 `json:"hotelling_t2"`
	MahalanobisDistance float64 `json:"mahalanobis_distance"`
	RSS                 float64 `json:"rss"`
	IsOutlier           bool    `json:"is_outlier"`
}

// OutputSummary contains summary statistics
type OutputSummary struct {
	ExplainedVariance   []float64         `json:"explained_variance,omitempty"`
	CumulativeVariance  []float64         `json:"cumulative_variance,omitempty"`
	Loadings            map[string][]float64 `json:"loadings,omitempty"`
}

// generateOutputPaths creates output file paths based on input file and format
func generateOutputPaths(inputFile, outputDir, format string) map[string]string {
	paths := make(map[string]string)
	
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
	
	// Generate paths based on format
	switch format {
	case "csv":
		paths["samples"] = filepath.Join(dir, baseName+"_pca_samples.csv")
		paths["features"] = filepath.Join(dir, baseName+"_pca_features.csv")
	case "json":
		paths["output"] = filepath.Join(dir, baseName+"_pca.json")
	}
	
	return paths
}

// convertToPCAOutputData converts PCAResult and CSVData to PCAOutputData
func convertToPCAOutputData(result *types.PCAResult, data *CSVData, includeMetrics bool) *types.PCAOutputData {
	// Create sample data
	sampleData := types.SampleData{
		Names:  data.RowNames,
		Scores: result.Scores,
	}
	
	// Add metrics if requested
	if includeMetrics {
		// Calculate actual metrics using the metrics calculator
		metrics, err := core.CalculateMetricsFromPCAResult(result, data.Matrix)
		if err != nil {
			// If metrics calculation fails, use zero values
			fmt.Fprintf(os.Stderr, "Warning: Failed to calculate metrics: %v\n", err)
			sampleData.Metrics = make([]types.SampleMetrics, len(result.Scores))
			for i := range sampleData.Metrics {
				sampleData.Metrics[i] = types.SampleMetrics{
					HotellingT2: 0.0,
					Mahalanobis: 0.0,
					RSS:         0.0,
					IsOutlier:   false,
				}
			}
		} else {
			sampleData.Metrics = metrics
		}
	}
	
	// Create feature data
	featureData := types.FeatureData{
		Names:    data.ColumnNames,
		Loadings: result.Loadings,
		Means:    result.Means,
		StdDevs:  result.StdDevs,
	}
	
	// Determine preprocessing type
	preprocessing := "none"
	if len(result.Means) > 0 {
		preprocessing = "mean_centered"
		if len(result.StdDevs) > 0 {
			preprocessing = "standard_scaled"
		}
	}
	
	// Create metadata
	metadata := types.PCAMetadata{
		NSamples:           data.Rows,
		NFeatures:          data.Columns,
		NComponents:        len(result.ComponentLabels),
		Method:             "svd", // TODO: Get from config
		Preprocessing:      preprocessing,
		ExplainedVariance:  result.ExplainedVar,
		CumulativeVariance: result.CumulativeVar,
	}
	
	return &types.PCAOutputData{
		Samples:  sampleData,
		Features: featureData,
		Metadata: metadata,
	}
}

// outputTableFormat outputs results in table format to stdout
func outputTableFormat(result *types.PCAResult, data *CSVData, 
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
	// Calculate metrics if requested
	var metrics []types.SampleMetrics
	if includeMetrics && outputScores {
		var err error
		metrics, err = core.CalculateMetricsFromPCAResult(result, data.Matrix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to calculate metrics: %v\n", err)
			// Create placeholder metrics
			metrics = make([]types.SampleMetrics, len(result.Scores))
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
		
		// Add data rows (show first 10 and last 5 for large datasets)
		nRows := len(result.Scores)
		rowsToShow := nRows
		if nRows > 15 {
			rowsToShow = 15
		}
		
		for i := 0; i < rowsToShow; i++ {
			rowIdx := i
			if i >= 10 && nRows > 15 {
				// Skip to last 5 rows
				rowIdx = nRows - (15 - i)
				if i == 10 {
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
		
		if nRows > 15 {
			fmt.Printf("\nShowing first 10 and last 5 of %d samples\n", nRows)
		}
	}
	
	// Output loadings table
	if outputLoadings {
		fmt.Println("\nPCA Loadings:")
		fmt.Println("──────────────────────────────────────────────────────────────")
		
		// Print headers
		fmt.Printf("%-25s", "Variable")
		for i := 0; i < len(result.ComponentLabels); i++ {
			fmt.Printf("%12s", result.ComponentLabels[i])
		}
		fmt.Println()
		fmt.Println("──────────────────────────────────────────────────────────────")
		
		// Add loading rows
		for i := 0; i < len(data.ColumnNames); i++ {
			fmt.Printf("%-25s", data.ColumnNames[i])
			for j := 0; j < len(result.ComponentLabels); j++ {
				fmt.Printf("%12.4f", result.Loadings[i][j])
			}
			fmt.Println()
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
	
	return nil
}

// outputCSVFormat outputs results in CSV format
func outputCSVFormat(result *types.PCAResult, data *CSVData, inputFile, outputDir string,
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
	// Convert to PCAOutputData
	outputData := convertToPCAOutputData(result, data, includeMetrics)
	
	// Generate output paths
	paths := generateOutputPaths(inputFile, outputDir, "csv")
	
	// Create output directory if needed
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	
	// Write samples CSV
	if err := writeSamplesCSV(outputData, paths["samples"], includeMetrics); err != nil {
		return err
	}
	
	// Write features CSV
	if err := writeFeaturesCSV(outputData, paths["features"]); err != nil {
		return err
	}
	
	fmt.Printf("\nResults saved to:\n")
	fmt.Printf("  Samples: %s\n", paths["samples"])
	fmt.Printf("  Features: %s\n", paths["features"])
	
	return nil
}

// writeSamplesCSV writes the samples data to a CSV file
func writeSamplesCSV(data *types.PCAOutputData, filename string, includeMetrics bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create samples file: %w", err)
	}
	defer func() { _ = file.Close() }()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header
	headers := []string{""} // Index column
	for i := 0; i < data.Metadata.NComponents; i++ {
		headers = append(headers, fmt.Sprintf("PC%d", i+1))
	}
	if includeMetrics {
		headers = append(headers, "hotelling_t2", "mahalanobis_distances", 
			"residual_sum_of_squares", "outlier_mask")
	}
	
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	
	// Write data rows
	for i := 0; i < len(data.Samples.Scores); i++ {
		row := []string{}
		
		// Sample name
		if i < len(data.Samples.Names) && data.Samples.Names[i] != "" {
			row = append(row, data.Samples.Names[i])
		} else {
			row = append(row, fmt.Sprintf("Sample_%d", i+1))
		}
		
		// PC scores
		for j := 0; j < data.Metadata.NComponents; j++ {
			row = append(row, strconv.FormatFloat(data.Samples.Scores[i][j], 'f', -1, 64))
		}
		
		// Metrics
		if includeMetrics && data.Samples.Metrics != nil {
			metrics := data.Samples.Metrics[i]
			row = append(row, 
				strconv.FormatFloat(metrics.HotellingT2, 'f', -1, 64),
				strconv.FormatFloat(metrics.Mahalanobis, 'f', -1, 64),
				strconv.FormatFloat(metrics.RSS, 'f', -1, 64),
				strconv.FormatBool(metrics.IsOutlier))
		}
		
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	
	return nil
}

// writeFeaturesCSV writes the features data to a CSV file
func writeFeaturesCSV(data *types.PCAOutputData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create features file: %w", err)
	}
	defer func() { _ = file.Close() }()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header with feature names
	headers := []string{""} // Index column
	headers = append(headers, data.Features.Names...)
	
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	
	// Write loadings for each PC
	for i := 0; i < data.Metadata.NComponents; i++ {
		row := []string{fmt.Sprintf("PC%d", i+1)}
		
		for j := 0; j < data.Metadata.NFeatures; j++ {
			row = append(row, strconv.FormatFloat(data.Features.Loadings[j][i], 'f', -1, 64))
		}
		
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write loadings row: %w", err)
		}
	}
	
	// Write mean row
	if len(data.Features.Means) > 0 {
		row := []string{"mean"}
		for _, mean := range data.Features.Means {
			row = append(row, strconv.FormatFloat(mean, 'f', -1, 64))
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write mean row: %w", err)
		}
	}
	
	// Write stdev row
	if len(data.Features.StdDevs) > 0 {
		row := []string{"stdev"}
		for _, stdev := range data.Features.StdDevs {
			row = append(row, strconv.FormatFloat(stdev, 'f', -1, 64))
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write stdev row: %w", err)
		}
	}
	
	return nil
}

// outputJSONFormat outputs results in JSON format
func outputJSONFormat(result *types.PCAResult, data *CSVData, inputFile, outputDir string,
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
	// Convert to PCAOutputData
	outputData := convertToPCAOutputData(result, data, includeMetrics)
	
	// Generate output paths
	paths := generateOutputPaths(inputFile, outputDir, "json")
	outputFile := paths["output"]
	
	// Create output directory if needed
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
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

