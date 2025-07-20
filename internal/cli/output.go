package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bitjungle/complab/pkg/types"
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

// outputTableFormat outputs results in table format to stdout
func outputTableFormat(result *types.PCAResult, data *CSVData, 
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
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
			
			// Metrics (placeholder for now)
			if includeMetrics {
				// TODO: Calculate actual metrics
				fmt.Printf("%15s%18s%10s%10s", "N/A", "N/A", "N/A", "False")
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
func outputCSVFormat(result *types.PCAResult, data *CSVData, outputFile string,
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
	// Skip if nothing to output
	if !outputScores && !outputLoadings && !outputVariance {
		return fmt.Errorf("no output requested: use --output-scores, --output-loadings, or --output-variance")
	}
	
	// For now, only handle scores output
	if !outputScores {
		return fmt.Errorf("CSV format currently only supports scores output")
	}
	
	// Determine output writer
	var w io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		w = file
	}
	
	writer := csv.NewWriter(w)
	defer writer.Flush()
	
	// Write header
	headers := []string{""}  // Index column
	for i := 0; i < len(result.ComponentLabels); i++ {
		headers = append(headers, result.ComponentLabels[i])
	}
	if includeMetrics {
		headers = append(headers, "hotelling_t2", "mahalanobis_distances", 
			"residual_sum_of_squares", "outlier_mask")
	}
	
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	
	// Write data rows
	for i := 0; i < len(result.Scores); i++ {
		row := []string{}
		
		// Row name/index
		if i < len(data.RowNames) {
			row = append(row, data.RowNames[i])
		} else {
			row = append(row, strconv.Itoa(i))
		}
		
		// PC scores
		for j := 0; j < len(result.ComponentLabels); j++ {
			row = append(row, strconv.FormatFloat(result.Scores[i][j], 'f', -1, 64))
		}
		
		// Metrics (placeholder for now)
		if includeMetrics {
			// TODO: Calculate actual metrics
			row = append(row, "0.0", "0.0", "0.0", "False")
		}
		
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	
	if outputFile != "" {
		fmt.Printf("\nResults saved to: %s\n", outputFile)
	}
	
	return nil
}

// outputJSONFormat outputs results in JSON format
func outputJSONFormat(result *types.PCAResult, data *CSVData, outputFile string,
	outputScores, outputLoadings, outputVariance, includeMetrics bool) error {
	
	// Prepare output data
	output := OutputData{
		Metadata: OutputMetadata{
			NSamples:      data.Rows,
			NFeatures:     data.Columns,
			NComponents:   len(result.ComponentLabels),
			Preprocessing: "standard", // TODO: Get from actual config
		},
		Results: []SampleResult{},
		Summary: OutputSummary{},
	}
	
	// Add sample results
	if outputScores {
		for i := 0; i < len(result.Scores); i++ {
			sampleID := fmt.Sprintf("Sample_%d", i+1)
			if i < len(data.RowNames) {
				sampleID = data.RowNames[i]
			}
			
			scores := make(map[string]float64)
			for j := 0; j < len(result.ComponentLabels); j++ {
				scores[result.ComponentLabels[j]] = result.Scores[i][j]
			}
			
			sample := SampleResult{
				ID:     sampleID,
				Scores: scores,
			}
			
			if includeMetrics {
				// TODO: Calculate actual metrics
				sample.Metrics = &SampleMetrics{
					HotellingT2:         0.0,
					MahalanobisDistance: 0.0,
					RSS:                 0.0,
					IsOutlier:           false,
				}
			}
			
			output.Results = append(output.Results, sample)
		}
	}
	
	// Add summary data
	if outputVariance {
		output.Summary.ExplainedVariance = result.ExplainedVar
		output.Summary.CumulativeVariance = result.CumulativeVar
	}
	
	if outputLoadings {
		output.Summary.Loadings = make(map[string][]float64)
		for i, varName := range data.ColumnNames {
			loadings := make([]float64, len(result.ComponentLabels))
			for j := range result.ComponentLabels {
				loadings[j] = result.Loadings[i][j]
			}
			output.Summary.Loadings[varName] = loadings
		}
	}
	
	// Marshal JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write JSON file: %w", err)
		}
		fmt.Printf("Results saved to: %s\n", outputFile)
	} else {
		fmt.Println(string(jsonData))
	}
	
	return nil
}

