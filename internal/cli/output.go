package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// OutputData represents the complete output data structure
type OutputData struct {
	Metadata OutputMetadata `json:"metadata"`
	Results  []SampleResult `json:"results"`
	Summary  OutputSummary  `json:"summary,omitempty"`
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
	ExplainedVariance  []float64            `json:"explained_variance,omitempty"`
	CumulativeVariance []float64            `json:"cumulative_variance,omitempty"`
	Loadings           map[string][]float64 `json:"loadings,omitempty"`
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
	const outputFileSuffix = "_pca"
	if format == "json" {
		paths["output"] = filepath.Join(dir, baseName+outputFileSuffix+".json")
	}

	return paths
}

// ConvertToPCAOutputData converts PCAResult and CSVData to PCAOutputData
func ConvertToPCAOutputData(result *types.PCAResult, data *CSVData, includeMetrics bool,
	config types.PCAConfig, preprocessor *core.Preprocessor,
	categoricalData map[string][]string, targetData map[string][]float64) *types.PCAOutputData {

	// Create timestamp
	createdAt := time.Now().Format(time.RFC3339)

	// Create model metadata
	metadata := types.ModelMetadata{
		Version:   "1.0",
		CreatedAt: createdAt,
		Software:  "gopca",
		Config: types.ModelConfig{
			Method:          config.Method,
			NComponents:     config.Components,
			MissingStrategy: config.MissingStrategy,
			ExcludedRows:    config.ExcludedRows,
			ExcludedColumns: config.ExcludedColumns,
			KernelType:      config.KernelType,
			KernelGamma:     config.KernelGamma,
			KernelDegree:    config.KernelDegree,
			KernelCoef0:     config.KernelCoef0,
		},
	}

	// Create preprocessing info
	preprocessingInfo := types.PreprocessingInfo{
		MeanCenter:    config.MeanCenter,
		StandardScale: config.StandardScale,
		RobustScale:   config.RobustScale,
		ScaleOnly:     config.ScaleOnly,
		SNV:           config.SNV,
		VectorNorm:    config.VectorNorm,
		Parameters:    types.PreprocessingParams{},
	}

	// Add preprocessing parameters if preprocessor was used
	if preprocessor != nil {
		preprocessingInfo.Parameters.FeatureMeans = preprocessor.GetMeans()
		preprocessingInfo.Parameters.FeatureStdDevs = preprocessor.GetStdDevs()
		preprocessingInfo.Parameters.FeatureMedians = preprocessor.GetMedians()
		preprocessingInfo.Parameters.FeatureMADs = preprocessor.GetMADs()
		preprocessingInfo.Parameters.RowMeans = preprocessor.GetRowMeans()
		preprocessingInfo.Parameters.RowStdDevs = preprocessor.GetRowStdDevs()
	}

	// Create model components
	modelComponents := types.ModelComponents{
		Loadings:               result.Loadings,
		ExplainedVariance:      result.ExplainedVar,
		ExplainedVarianceRatio: result.ExplainedVarRatio,
		CumulativeVariance:     result.CumulativeVar,
		ComponentLabels:        result.ComponentLabels,
		FeatureLabels:          data.Headers,
	}

	// Create results data
	resultsData := types.ResultsData{
		Samples: types.SamplesResults{
			Names:  data.RowNames,
			Scores: result.Scores,
		},
	}

	// Add metrics if requested
	if includeMetrics {
		metrics, err := core.CalculateMetricsFromPCAResult(result, data.Matrix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to calculate metrics: %v\n", err)
		} else {
			metricsData := &types.MetricsData{
				HotellingT2: make([]float64, len(metrics)),
				Mahalanobis: make([]float64, len(metrics)),
				RSS:         make([]float64, len(metrics)),
				IsOutlier:   make([]bool, len(metrics)),
			}
			for i, m := range metrics {
				metricsData.HotellingT2[i] = m.HotellingT2
				metricsData.Mahalanobis[i] = m.Mahalanobis
				metricsData.RSS[i] = m.RSS
				metricsData.IsOutlier[i] = m.IsOutlier
			}
			resultsData.Samples.Metrics = metricsData
		}
	}

	// Create diagnostic limits
	diagnostics := types.DiagnosticLimits{
		T2Limit95: result.T2Limit95,
		T2Limit99: result.T2Limit99,
		QLimit95:  result.QLimit95,
		QLimit99:  result.QLimit99,
	}

	// Add preserved columns if provided
	var preservedColumns *types.PreservedColumns
	if len(categoricalData) > 0 || len(targetData) > 0 {
		preservedColumns = &types.PreservedColumns{
			Categorical:   categoricalData,
			NumericTarget: targetData,
		}
	}

	return &types.PCAOutputData{
		Metadata:          metadata,
		Preprocessing:     preprocessingInfo,
		Model:             modelComponents,
		Results:           resultsData,
		Diagnostics:       diagnostics,
		Eigencorrelations: result.Eigencorrelations,
		PreservedColumns:  preservedColumns,
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
		for i := 0; i < len(data.Headers); i++ {
			fmt.Printf("%-25s", data.Headers[i])
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

// outputJSONFormat outputs results in JSON format
func outputJSONFormat(result *types.PCAResult, data *CSVData, inputFile, outputDir string,
	outputScores, outputLoadings, outputVariance, includeMetrics bool,
	config types.PCAConfig, preprocessor *core.Preprocessor,
	categoricalData map[string][]string, targetData map[string][]float64) error {

	// Convert to PCAOutputData
	outputData := ConvertToPCAOutputData(result, data, includeMetrics, config, preprocessor, categoricalData, targetData)

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
