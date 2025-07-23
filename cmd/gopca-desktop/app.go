package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gonum.org/v1/gonum/mat"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// FileData represents the structure of CSV data for the frontend
type FileData struct {
	Headers            []string            `json:"headers"`
	RowNames           []string            `json:"rowNames"`
	Data               [][]float64         `json:"data"`
	MissingMask        [][]bool            `json:"missingMask,omitempty"`
	CategoricalColumns map[string][]string `json:"categoricalColumns,omitempty"`
}

// PCARequest represents a PCA analysis request from the frontend
type PCARequest struct {
	Data            [][]float64 `json:"data"`
	MissingMask     [][]bool    `json:"missingMask,omitempty"`
	Headers         []string    `json:"headers"`
	RowNames        []string    `json:"rowNames"`
	Components      int         `json:"components"`
	MeanCenter      bool        `json:"meanCenter"`
	StandardScale   bool        `json:"standardScale"`
	RobustScale     bool        `json:"robustScale"`
	SNV             bool        `json:"snv"`
	VectorNorm      bool        `json:"vectorNorm"`
	Method          string      `json:"method"`
	ExcludedRows    []int       `json:"excludedRows,omitempty"`
	ExcludedColumns []int       `json:"excludedColumns,omitempty"`
	MissingStrategy string      `json:"missingStrategy,omitempty"`
	CalculateMetrics bool       `json:"calculateMetrics,omitempty"`
	// Kernel PCA parameters
	KernelType   string  `json:"kernelType,omitempty"`
	KernelGamma  float64 `json:"kernelGamma,omitempty"`
	KernelDegree int     `json:"kernelDegree,omitempty"`
	KernelCoef0  float64 `json:"kernelCoef0,omitempty"`
}

// PCAResponse represents the PCA analysis results
type PCAResponse struct {
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	Result  *types.PCAResult `json:"result,omitempty"`
	Info    string           `json:"info,omitempty"`
}

// RunPCA performs PCA analysis on the provided data
func (a *App) RunPCA(request PCARequest) (response PCAResponse) {
	// Recover from any panic to prevent app crash
	defer func() {
		if r := recover(); r != nil {
			response = PCAResponse{
				Success: false,
				Error:   fmt.Sprintf("Unexpected error during PCA analysis: %v", r),
			}
		}
	}()

	// Validate request
	if len(request.Data) == 0 {
		return PCAResponse{
			Success: false,
			Error:   "No data provided",
		}
	}

	if request.Components <= 0 {
		request.Components = 2 // Default to 2 components
	}

	// Restore NaN values from missing mask
	dataToAnalyze := make([][]float64, len(request.Data))
	for i := range request.Data {
		dataToAnalyze[i] = make([]float64, len(request.Data[i]))
		for j := range request.Data[i] {
			if request.MissingMask != nil && i < len(request.MissingMask) && j < len(request.MissingMask[i]) && request.MissingMask[i][j] {
				dataToAnalyze[i][j] = math.NaN()
			} else {
				dataToAnalyze[i][j] = request.Data[i][j]
			}
		}
	}

	// Filter data if exclusions are provided
	if len(request.ExcludedRows) > 0 || len(request.ExcludedColumns) > 0 {
		// Filter the data matrix
		filteredData, err := utils.FilterMatrix(dataToAnalyze, request.ExcludedRows, request.ExcludedColumns)
		if err != nil {
			return PCAResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to filter data: %v", err),
			}
		}
		dataToAnalyze = filteredData

		// Note: We don't need to filter headers and row names for PCA computation
		// The frontend handles the display of selected data
	}

	// Check for missing values and handle based on strategy
	hasMissing := false
	missingInfo := &types.MissingValueInfo{}
	rowsDropped := 0

	// Count missing values
	for i := 0; i < len(dataToAnalyze); i++ {
		for j := 0; j < len(dataToAnalyze[i]); j++ {
			if math.IsNaN(dataToAnalyze[i][j]) {
				hasMissing = true
				missingInfo.TotalMissing++
			}
		}
	}

	// Handle missing values based on strategy
	if hasMissing {
		// Default strategy if not specified
		if request.MissingStrategy == "" {
			request.MissingStrategy = "error"
		}

		switch request.MissingStrategy {
		case "error":
			return PCAResponse{
				Success: false,
				Error:   fmt.Sprintf("Missing values detected (%d values). Please select a strategy to handle them: 'drop' to remove rows, or 'mean' to impute with column means.", missingInfo.TotalMissing),
			}
		case "drop", "mean", "median":
			// Create missing value handler
			strategy := types.MissingValueStrategy(request.MissingStrategy)
			handler := core.NewMissingValueHandler(strategy)

			// Convert to types.Matrix for handler
			matrix := types.Matrix(dataToAnalyze)

			// Create missing info manually since we don't have CSVData here
			selectedCols := make([]int, len(dataToAnalyze[0]))
			for i := range selectedCols {
				selectedCols[i] = i
			}

			// Build missing info
			actualMissingInfo := &types.MissingValueInfo{
				ColumnIndices:   selectedCols,
				RowsAffected:    []int{},
				MissingByColumn: make(map[int]int),
			}

			// Find rows with missing values
			rowsWithMissing := make(map[int]bool)
			for i := 0; i < len(dataToAnalyze); i++ {
				for j := 0; j < len(dataToAnalyze[i]); j++ {
					if math.IsNaN(dataToAnalyze[i][j]) {
						rowsWithMissing[i] = true
						actualMissingInfo.MissingByColumn[j]++
						actualMissingInfo.TotalMissing++
					}
				}
			}

			// Convert map to slice
			for row := range rowsWithMissing {
				actualMissingInfo.RowsAffected = append(actualMissingInfo.RowsAffected, row)
			}

			// Apply missing value strategy
			cleanedData, err := handler.HandleMissingValues(matrix, actualMissingInfo, selectedCols)
			if err != nil {
				return PCAResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to handle missing values: %v", err),
				}
			}

			dataToAnalyze = cleanedData
			rowsDropped = len(actualMissingInfo.RowsAffected)

			// Update row names if rows were dropped
			if request.MissingStrategy == "drop" && rowsDropped > 0 {
				// Create a map of rows to keep
				keepRows := make(map[int]bool)
				for i := 0; i < len(matrix); i++ {
					keepRows[i] = true
				}
				for _, rowIdx := range actualMissingInfo.RowsAffected {
					delete(keepRows, rowIdx)
				}

				// Filter row names
				newRowNames := []string{}
				for i := 0; i < len(request.RowNames); i++ {
					if keepRows[i] {
						newRowNames = append(newRowNames, request.RowNames[i])
					}
				}
				request.RowNames = newRowNames
			}
		default:
			return PCAResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid missing value strategy: %s", request.MissingStrategy),
			}
		}
	}

	// Create PCA configuration
	config := types.PCAConfig{
		Components:      request.Components,
		MeanCenter:      request.MeanCenter,
		StandardScale:   request.StandardScale,
		RobustScale:     request.RobustScale,
		SNV:             request.SNV,
		VectorNorm:      request.VectorNorm,
		Method:          strings.ToLower(request.Method),
		ExcludedRows:    request.ExcludedRows,
		ExcludedColumns: request.ExcludedColumns,
	}

	// Add kernel parameters if using kernel PCA
	if strings.ToLower(request.Method) == "kernel" {
		config.KernelType = request.KernelType
		config.KernelGamma = request.KernelGamma
		config.KernelDegree = request.KernelDegree
		config.KernelCoef0 = request.KernelCoef0

		// Skip standard preprocessing for kernel PCA
		config.MeanCenter = false
		config.StandardScale = false
	}

	// Perform PCA
	engine := core.NewPCAEngineForMethod(config.Method)
	result, err := engine.Fit(dataToAnalyze, config)
	if err != nil {
		return PCAResponse{
			Success: false,
			Error:   fmt.Sprintf("PCA fit failed: %v", err),
		}
	}

	// Update component labels to use filtered headers if needed
	if len(result.ComponentLabels) == 0 {
		result.ComponentLabels = make([]string, request.Components)
		for i := 0; i < request.Components; i++ {
			result.ComponentLabels[i] = fmt.Sprintf("PC%d", i+1)
		}
	}

	// Add variable labels from headers (excluding the ones that were filtered out)
	filteredHeaders := make([]string, 0)
	for j, header := range request.Headers {
		if !contains(request.ExcludedColumns, j) {
			filteredHeaders = append(filteredHeaders, header)
		}
	}
	result.VariableLabels = filteredHeaders

	// Calculate diagnostic metrics if requested
	if request.CalculateMetrics && strings.ToLower(request.Method) != "kernel" {
		// Keep a copy of the original data (before preprocessing) for residual calculations
		originalData := make(types.Matrix, len(dataToAnalyze))
		for i := range dataToAnalyze {
			originalData[i] = make([]float64, len(dataToAnalyze[i]))
			copy(originalData[i], dataToAnalyze[i])
		}

		// Create metrics calculator
		scoresDense := matrixToDense(result.Scores)
		loadingsDense := matrixToDense(result.Loadings)
		metricsCalc := core.NewPCAMetricsCalculator(scoresDense, loadingsDense, result.Means, result.StdDevs)

		// Calculate metrics
		metrics, err := metricsCalc.CalculateMetrics(originalData)
		if err != nil {
			// Don't fail the whole PCA, just log the error
			fmt.Printf("Warning: Failed to calculate diagnostic metrics: %v\n", err)
		} else {
			result.Metrics = metrics
			// TODO: Add confidence limit calculations
			// For now, set placeholder values
			result.T2Limit95 = 0.0
			result.T2Limit99 = 0.0
			result.QLimit95 = 0.0
			result.QLimit99 = 0.0
		}
	}

	// Build info message about missing value handling
	infoMsg := ""
	if hasMissing && request.MissingStrategy != "error" {
		switch request.MissingStrategy {
		case "drop":
			infoMsg = fmt.Sprintf("Dropped %d rows containing missing values.", rowsDropped)
		case "mean":
			infoMsg = fmt.Sprintf("Imputed %d missing values with column means.", missingInfo.TotalMissing)
		case "median":
			infoMsg = fmt.Sprintf("Imputed %d missing values with column medians.", missingInfo.TotalMissing)
		}
	}

	return PCAResponse{
		Success: true,
		Result:  result,
		Info:    infoMsg,
	}
}

// contains checks if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// matrixToDense converts types.Matrix to *mat.Dense
func matrixToDense(m types.Matrix) *mat.Dense {
	if len(m) == 0 || len(m[0]) == 0 {
		return mat.NewDense(0, 0, nil)
	}

	rows, cols := len(m), len(m[0])
	data := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			data[i*cols+j] = m[i][j]
		}
	}
	return mat.NewDense(rows, cols, data)
}

// SaveFile handles saving exported plot data
func (a *App) SaveFile(fileName string, dataURL string) error {
	// Show save dialog
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: fileName,
		Title:           "Save Plot",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PNG Image",
				Pattern:     "*.png",
			},
			{
				DisplayName: "SVG Image",
				Pattern:     "*.svg",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to open save dialog: %v", err)
	}

	// User cancelled
	if filePath == "" {
		return nil
	}

	// Parse the data URL
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid data URL format")
	}

	// Decode based on format
	var data []byte
	if strings.Contains(parts[0], "base64") {
		// PNG format (base64 encoded)
		data, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode base64 data: %v", err)
		}
	} else {
		// SVG format (URL encoded)
		decodedSVG, err := url.QueryUnescape(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode SVG data: %v", err)
		}
		data = []byte(decodedSVG)
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// LoadIrisDataset loads the built-in iris dataset with species column
func (a *App) LoadIrisDataset() (*FileDataJSON, error) {
	// Hardcoded iris dataset with species column
	csvContent := `sepal length (cm),sepal width (cm),petal length (cm),petal width (cm),species
5.1,3.5,1.4,0.2,setosa
4.9,3.0,1.4,0.2,setosa
4.7,3.2,1.3,0.2,setosa
4.6,3.1,1.5,0.2,setosa
5.0,3.6,1.4,0.2,setosa
5.4,3.9,1.7,0.4,setosa
4.6,3.4,1.4,0.3,setosa
5.0,3.4,1.5,0.2,setosa
4.4,2.9,1.4,0.2,setosa
4.9,3.1,1.5,0.1,setosa
5.4,3.7,1.5,0.2,setosa
4.8,3.4,1.6,0.2,setosa
4.8,3.0,1.4,0.1,setosa
4.3,3.0,1.1,0.1,setosa
5.8,4.0,1.2,0.2,setosa
5.7,4.4,1.5,0.4,setosa
5.4,3.9,1.3,0.4,setosa
5.1,3.5,1.4,0.3,setosa
5.7,3.8,1.7,0.3,setosa
5.1,3.8,1.5,0.3,setosa
5.4,3.4,1.7,0.2,setosa
5.1,3.7,1.5,0.4,setosa
4.6,3.6,1.0,0.2,setosa
5.1,3.3,1.7,0.5,setosa
4.8,3.4,1.9,0.2,setosa
5.0,3.0,1.6,0.2,setosa
5.0,3.4,1.6,0.4,setosa
5.2,3.5,1.5,0.2,setosa
5.2,3.4,1.4,0.2,setosa
4.7,3.2,1.6,0.2,setosa
4.8,3.1,1.6,0.2,setosa
5.4,3.4,1.5,0.4,setosa
5.2,4.1,1.5,0.1,setosa
5.5,4.2,1.4,0.2,setosa
4.9,3.1,1.5,0.2,setosa
5.0,3.2,1.2,0.2,setosa
5.5,3.5,1.3,0.2,setosa
4.9,3.6,1.4,0.1,setosa
4.4,3.0,1.3,0.2,setosa
5.1,3.4,1.5,0.2,setosa
5.0,3.5,1.3,0.3,setosa
4.5,2.3,1.3,0.3,setosa
4.4,3.2,1.3,0.2,setosa
5.0,3.5,1.6,0.6,setosa
5.1,3.8,1.9,0.4,setosa
4.8,3.0,1.4,0.3,setosa
5.1,3.8,1.6,0.2,setosa
4.6,3.2,1.4,0.2,setosa
5.3,3.7,1.5,0.2,setosa
5.0,3.3,1.4,0.2,setosa
7.0,3.2,4.7,1.4,versicolor
6.4,3.2,4.5,1.5,versicolor
6.9,3.1,4.9,1.5,versicolor
5.5,2.3,4.0,1.3,versicolor
6.5,2.8,4.6,1.5,versicolor
5.7,2.8,4.5,1.3,versicolor
6.3,3.3,4.7,1.6,versicolor
4.9,2.4,3.3,1.0,versicolor
6.6,2.9,4.6,1.3,versicolor
5.2,2.7,3.9,1.4,versicolor
5.0,2.0,3.5,1.0,versicolor
5.9,3.0,4.2,1.5,versicolor
6.0,2.2,4.0,1.0,versicolor
6.1,2.9,4.7,1.4,versicolor
5.6,2.9,3.6,1.3,versicolor
6.7,3.1,4.4,1.4,versicolor
5.6,3.0,4.5,1.5,versicolor
5.8,2.7,4.1,1.0,versicolor
6.2,2.2,4.5,1.5,versicolor
5.6,2.5,3.9,1.1,versicolor
5.9,3.2,4.8,1.8,versicolor
6.1,2.8,4.0,1.3,versicolor
6.3,2.5,4.9,1.5,versicolor
6.1,2.8,4.7,1.2,versicolor
6.4,2.9,4.3,1.3,versicolor
6.6,3.0,4.4,1.4,versicolor
6.8,2.8,4.8,1.4,versicolor
6.7,3.0,5.0,1.7,versicolor
6.0,2.9,4.5,1.5,versicolor
5.7,2.6,3.5,1.0,versicolor
5.5,2.4,3.8,1.1,versicolor
5.5,2.4,3.7,1.0,versicolor
5.8,2.7,3.9,1.2,versicolor
6.0,2.7,5.1,1.6,versicolor
5.4,3.0,4.5,1.5,versicolor
6.0,3.4,4.5,1.6,versicolor
6.7,3.1,4.7,1.5,versicolor
6.3,2.3,4.4,1.3,versicolor
5.6,3.0,4.1,1.3,versicolor
5.5,2.5,4.0,1.3,versicolor
5.5,2.6,4.4,1.2,versicolor
6.1,3.0,4.6,1.4,versicolor
5.8,2.6,4.0,1.2,versicolor
5.0,2.3,3.3,1.0,versicolor
5.6,2.7,4.2,1.3,versicolor
5.7,3.0,4.2,1.2,versicolor
5.7,2.9,4.2,1.3,versicolor
6.2,2.9,4.3,1.3,versicolor
5.1,2.5,3.0,1.1,versicolor
5.7,2.8,4.1,1.3,versicolor
6.3,3.3,6.0,2.5,virginica
5.8,2.7,5.1,1.9,virginica
7.1,3.0,5.9,2.1,virginica
6.3,2.9,5.6,1.8,virginica
6.5,3.0,5.8,2.2,virginica
7.6,3.0,6.6,2.1,virginica
4.9,2.5,4.5,1.7,virginica
7.3,2.9,6.3,1.8,virginica
6.7,2.5,5.8,1.8,virginica
7.2,3.6,6.1,2.5,virginica
6.5,3.2,5.1,2.0,virginica
6.4,2.7,5.3,1.9,virginica
6.8,3.0,5.5,2.1,virginica
5.7,2.5,5.0,2.0,virginica
5.8,2.8,5.1,2.4,virginica
6.4,3.2,5.3,2.3,virginica
6.5,3.0,5.5,1.8,virginica
7.7,3.8,6.7,2.2,virginica
7.7,2.6,6.9,2.3,virginica
6.0,2.2,5.0,1.5,virginica
6.9,3.2,5.7,2.3,virginica
5.6,2.8,4.9,2.0,virginica
7.7,2.8,6.7,2.0,virginica
6.3,2.7,4.9,1.8,virginica
6.7,3.3,5.7,2.1,virginica
7.2,3.2,6.0,1.8,virginica
6.2,2.8,4.8,1.8,virginica
6.1,3.0,4.9,1.8,virginica
6.4,2.8,5.6,2.1,virginica
7.2,3.0,5.8,1.6,virginica
7.4,2.8,6.1,1.9,virginica
7.9,3.8,6.4,2.0,virginica
6.4,2.8,5.6,2.2,virginica
6.3,2.8,5.1,1.5,virginica
6.1,2.6,5.6,1.4,virginica
7.7,3.0,6.1,2.3,virginica
6.3,3.4,5.6,2.4,virginica
6.4,3.1,5.5,1.8,virginica
6.0,3.0,4.8,1.8,virginica
6.9,3.1,5.4,2.1,virginica
6.7,3.1,5.6,2.4,virginica
6.9,3.1,5.1,2.3,virginica
5.8,2.7,5.1,1.9,virginica
6.8,3.2,5.9,2.3,virginica
6.7,3.3,5.7,2.5,virginica
6.7,3.0,5.2,2.3,virginica
6.3,2.5,5.0,1.9,virginica
6.5,3.0,5.2,2.0,virginica
6.2,3.4,5.4,2.3,virginica
5.9,3.0,5.1,1.8,virginica`

	// Add row names to the CSV content
	lines := strings.Split(csvContent, "\n")
	var newLines []string
	newLines = append(newLines, ","+lines[0]) // Add empty header for row names column

	speciesCount := map[string]int{"setosa": 0, "versicolor": 0, "virginica": 0}

	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], ",")
		if len(parts) >= 5 {
			species := parts[4]
			speciesCount[species]++
			rowName := fmt.Sprintf("%s_%02d", species, speciesCount[species])
			newLines = append(newLines, rowName+","+lines[i])
		}
	}

	modifiedContent := strings.Join(newLines, "\n")

	return a.ParseCSV(modifiedContent)
}
