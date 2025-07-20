package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bitjungle/complab/internal/core"
	internalio "github.com/bitjungle/complab/internal/io"
	"github.com/bitjungle/complab/pkg/types"
)

func main() {
	var (
		inputFile        = flag.String("input", "", "Input CSV file (required)")
		outputFile       = flag.String("output", "", "Output file for metrics")
		components       = flag.Int("components", 0, "Number of components to use (0=all)")
		significance     = flag.Float64("significance", 0.01, "Significance level for outlier detection")
		format          = flag.String("format", "json", "Output format (json, csv)")
		pcaFirst        = flag.Bool("pca", true, "Run PCA analysis first")
	)
	
	flag.Parse()
	
	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// Load data - try to detect if first column contains row names
	data, headers, err := loadDataWithRowNameDetection(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Loaded data: %d observations, %d variables\n", len(data), len(data[0]))
	
	// Run PCA if requested
	var result *types.PCAResult
	if *pcaFirst {
		fmt.Println("Running PCA analysis...")
		
		config := types.PCAConfig{
			Components:    *components,
			MeanCenter:    true,
			StandardScale: false,
			Method:        "nipals",
		}
		
		if config.Components == 0 || config.Components > len(data[0]) {
			config.Components = len(data[0])
		}
		
		pca := core.NewPCAEngine()
		result, err = pca.Fit(data, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "PCA analysis failed: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("PCA completed with %d components\n", len(result.ExplainedVar))
	} else {
		fmt.Fprintf(os.Stderr, "Error: metrics calculation requires PCA results\n")
		os.Exit(1)
	}
	
	// Calculate metrics
	fmt.Println("Calculating metrics...")
	metricsCalculator := core.NewMetricsCalculator()
	metricsConfig := types.MetricsConfig{
		NumComponents:             *components,
		SignificanceLevel:        *significance,
		CalculateContributions:   true,
		CalculateConfidenceEllipse: true,
	}
	
	metrics, err := metricsCalculator.CalculateMetrics(result, data, metricsConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to calculate metrics: %v\n", err)
		os.Exit(1)
	}
	
	// Display summary
	displaySummary(metrics)
	
	// Debug: Calculate and display threshold
	if *components > 0 && len(data) > *components {
		df1 := float64(*components)
		df2 := float64(len(data) - *components)
		scale := (df1 * float64(len(data)-1)) / df2
		fmt.Printf("\n\nDebug Info:")
		fmt.Printf("\n  Components (p): %d", *components)
		fmt.Printf("\n  Samples (n): %d", len(data))
		fmt.Printf("\n  df1: %.0f, df2: %.0f", df1, df2)
		fmt.Printf("\n  Scale factor: %.4f", scale)
		fmt.Printf("\n  Max T²: %.4f", findMax(metrics.HotellingT2))
	}
	
	// Save if output file specified
	if *outputFile != "" {
		var saveErr error
		switch *format {
		case "json":
			saveErr = saveJSON(metrics, *outputFile)
		case "csv":
			saveErr = saveCSV(metrics, *outputFile, headers)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported format: %s\n", *format)
			os.Exit(1)
		}
		
		if saveErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to save metrics: %v\n", saveErr)
			os.Exit(1)
		}
		
		fmt.Printf("\nMetrics saved to %s\n", *outputFile)
	}
}

func displaySummary(metrics *types.PCAMetrics) {
	fmt.Println("\n=== PCA Metrics Summary ===")
	
	// Count outliers
	outlierCount := 0
	outlierIndices := []int{}
	for i, isOutlier := range metrics.OutlierMask {
		if isOutlier {
			outlierCount++
			if len(outlierIndices) < 10 {
				outlierIndices = append(outlierIndices, i)
			}
		}
	}
	
	fmt.Printf("\nOutlier Detection:")
	fmt.Printf("\n  Total observations: %d", len(metrics.OutlierMask))
	fmt.Printf("\n  Outliers detected: %d (%.1f%%)", outlierCount, 
		float64(outlierCount)/float64(len(metrics.OutlierMask))*100)
	
	if len(outlierIndices) > 0 {
		fmt.Printf("\n  First outlier indices: %v", outlierIndices)
	}
	
	// Display statistics
	fmt.Printf("\n\nMahalanobis Distance Statistics:")
	fmt.Printf("\n  Min: %.4f", findMin(metrics.MahalanobisDistances))
	fmt.Printf("\n  Max: %.4f", findMax(metrics.MahalanobisDistances))
	fmt.Printf("\n  Mean: %.4f", findMean(metrics.MahalanobisDistances))
	
	fmt.Printf("\n\nHotelling's T² Statistics:")
	fmt.Printf("\n  Min: %.4f", findMin(metrics.HotellingT2))
	fmt.Printf("\n  Max: %.4f", findMax(metrics.HotellingT2))
	fmt.Printf("\n  Mean: %.4f", findMean(metrics.HotellingT2))
	
	// Display confidence ellipse info if available
	if metrics.ConfidenceEllipse.ConfidenceLevel > 0 {
		fmt.Printf("\n\nConfidence Ellipse (%.0f%%):", metrics.ConfidenceEllipse.ConfidenceLevel*100)
		fmt.Printf("\n  Center: (%.4f, %.4f)", metrics.ConfidenceEllipse.CenterX, metrics.ConfidenceEllipse.CenterY)
		fmt.Printf("\n  Major axis: %.4f", metrics.ConfidenceEllipse.MajorAxis)
		fmt.Printf("\n  Minor axis: %.4f", metrics.ConfidenceEllipse.MinorAxis)
	}
}

func findMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func findMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func findMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func saveJSON(metrics *types.PCAMetrics, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

func loadDataWithRowNameDetection(filename string) (types.Matrix, []string, error) {
	// First try to load normally
	data, headers, err := internalio.LoadCSV(filename, internalio.CSVOptions{
		Delimiter: ',',
		HasHeader: true,
	})
	
	if err == nil {
		return data, headers, nil
	}
	
	// If error, try skipping first column (potential row names)
	file, err2 := os.Open(filename)
	if err2 != nil {
		return nil, nil, err // Return original error
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.Comma = ','
	
	// Read header
	headerRow, err2 := reader.Read()
	if err2 != nil {
		return nil, nil, err
	}
	
	// Skip first column (row names) in header
	if len(headerRow) > 1 {
		headers = headerRow[1:]
	}
	
	// Read data
	var matrix types.Matrix
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		
		// Skip first column and parse rest as floats
		if len(record) > 1 {
			row := make([]float64, len(record)-1)
			for i := 1; i < len(record); i++ {
				val, err := strconv.ParseFloat(record[i], 64)
				if err != nil {
					return nil, nil, fmt.Errorf("error parsing value '%s' at column %d: %w", record[i], i, err)
				}
				row[i-1] = val
			}
			matrix = append(matrix, row)
		}
	}
	
	return matrix, headers, nil
}

func saveCSV(metrics *types.PCAMetrics, filename string, headers []string) error {
	// Prepare data for CSV
	nObs := len(metrics.MahalanobisDistances)
	data := make([][]float64, nObs)
	
	for i := 0; i < nObs; i++ {
		row := []float64{
			metrics.MahalanobisDistances[i],
			metrics.HotellingT2[i],
			metrics.RSS[i],
			metrics.QResiduals[i],
		}
		if metrics.OutlierMask[i] {
			row = append(row, 1.0)
		} else {
			row = append(row, 0.0)
		}
		data[i] = row
	}
	
	// Create headers for metrics
	metricHeaders := []string{
		"Mahalanobis_Distance",
		"Hotelling_T2",
		"RSS",
		"Q_Residuals",
		"Is_Outlier",
	}
	
	// Save using CSV writer
	config := internalio.CSVOptions{
		Delimiter: ',',
		HasHeader: true,
	}
	
	return internalio.SaveCSV(filename, data, metricHeaders, config)
}