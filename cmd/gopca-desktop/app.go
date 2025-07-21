package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	Headers             []string               `json:"headers"`
	RowNames            []string               `json:"rowNames"`
	Data                [][]float64            `json:"data"`
	CategoricalColumns  map[string][]string    `json:"categoricalColumns,omitempty"`
}

// NaNSentinel is a special value used to represent NaN in JSON transport
const NaNSentinel = -999999.0

// MarshalJSON implements custom JSON marshaling to handle NaN values
func (f *FileData) MarshalJSON() ([]byte, error) {
	// Create a copy of the data with NaN replaced by sentinel value
	jsonData := make([][]float64, len(f.Data))
	for i, row := range f.Data {
		jsonData[i] = make([]float64, len(row))
		for j, val := range row {
			if math.IsNaN(val) {
				jsonData[i][j] = NaNSentinel
			} else {
				jsonData[i][j] = val
			}
		}
	}
	
	// Use a type alias to avoid infinite recursion
	type Alias FileData
	return json.Marshal(&struct {
		*Alias
		Data [][]float64 `json:"data"`
	}{
		Alias: (*Alias)(f),
		Data:  jsonData,
	})
}

// PCARequest represents a PCA analysis request from the frontend
type PCARequest struct {
	Data            [][]float64 `json:"data"`
	Headers         []string    `json:"headers"`
	RowNames        []string    `json:"rowNames"`
	Components      int         `json:"components"`
	MeanCenter      bool        `json:"meanCenter"`
	StandardScale   bool        `json:"standardScale"`
	RobustScale     bool        `json:"robustScale"`
	Method          string      `json:"method"`
	ExcludedRows    []int       `json:"excludedRows,omitempty"`
	ExcludedColumns []int       `json:"excludedColumns,omitempty"`
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
}

// ParseCSV parses CSV content and returns data matrix and headers
func (a *App) ParseCSV(content string) (*FileData, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.LazyQuotes = true    // Allow lazy quotes
	reader.TrimLeadingSpace = true // Trim leading space
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header row and one data row")
	}
	
	// First row is headers
	headers := records[0]
	
	// Check if first column contains row names
	hasRowNames := false
	firstValue := records[1][0]
	if _, err := strconv.ParseFloat(firstValue, 64); err != nil {
		// First column is not numeric, likely row names
		hasRowNames = true
		headers = headers[1:] // Remove first header
	}
	
	// Detect which columns are categorical vs numeric
	startIdx := 0
	if hasRowNames {
		startIdx = 1
	}
	
	numericCols := []int{}
	categoricalCols := []int{}
	categoricalData := make(map[string][]string)
	
	// Check each column to see if it's numeric or categorical
	for j := startIdx; j < len(records[0]); j++ {
		isNumeric := true
		// Check first 10 rows or all rows if less than 10
		checkRows := 10
		if len(records)-1 < checkRows {
			checkRows = len(records) - 1
		}
		
		for i := 1; i <= checkRows; i++ {
			val := strings.TrimSpace(records[i][j])
			// Treat 'm' as a missing value (numeric)
			if val != "m" {
				if _, err := strconv.ParseFloat(val, 64); err != nil {
					isNumeric = false
					break
				}
			}
		}
		
		if isNumeric {
			numericCols = append(numericCols, j)
		} else {
			categoricalCols = append(categoricalCols, j)
			colName := headers[j-startIdx]
			categoricalData[colName] = []string{}
		}
	}
	
	// Parse data
	var data [][]float64
	var rowNames []string
	var numericHeaders []string
	
	for i := 1; i < len(records); i++ {
		record := records[i]
		
		if hasRowNames {
			rowNames = append(rowNames, record[0])
		}
		
		// Extract numeric data
		row := make([]float64, len(numericCols))
		for idx, j := range numericCols {
			valStr := strings.TrimSpace(record[j])
			// Handle missing value indicator 'm'
			if valStr == "m" {
				row[idx] = math.NaN()
			} else {
				val, err := strconv.ParseFloat(valStr, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid number at row %d, col %d: %s", i, j, record[j])
				}
				row[idx] = val
			}
		}
		data = append(data, row)
		
		// Extract categorical data
		for _, j := range categoricalCols {
			colName := headers[j-startIdx]
			categoricalData[colName] = append(categoricalData[colName], record[j])
		}
	}
	
	// Build numeric headers
	for _, j := range numericCols {
		numericHeaders = append(numericHeaders, headers[j-startIdx])
	}
	
	result := &FileData{
		Headers:  numericHeaders,
		RowNames: rowNames,
		Data:     data,
	}
	
	// Only add categorical columns if there are any
	if len(categoricalData) > 0 {
		result.CategoricalColumns = categoricalData
	}
	
	return result, nil
}

// RunPCA performs PCA analysis on the provided data
func (a *App) RunPCA(request PCARequest) PCAResponse {
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
	
	// Convert sentinel values back to NaN
	floatData := make([][]float64, len(request.Data))
	for i, row := range request.Data {
		floatData[i] = make([]float64, len(row))
		for j, val := range row {
			if val == NaNSentinel {
				floatData[i][j] = math.NaN()
			} else {
				floatData[i][j] = val
			}
		}
	}
	
	// Filter data if exclusions are provided
	dataToAnalyze := floatData
	
	if len(request.ExcludedRows) > 0 || len(request.ExcludedColumns) > 0 {
		// Filter the data matrix
		filteredData, err := utils.FilterMatrix(floatData, request.ExcludedRows, request.ExcludedColumns)
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
	
	// Create PCA configuration
	config := types.PCAConfig{
		Components:      request.Components,
		MeanCenter:      request.MeanCenter,
		StandardScale:   request.StandardScale,
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
	
	// Check for NaN values in the data to analyze
	hasNaN := false
	nanCount := 0
	for _, row := range dataToAnalyze {
		for _, val := range row {
			if math.IsNaN(val) {
				hasNaN = true
				nanCount++
			}
		}
	}
	
	if hasNaN {
		return PCAResponse{
			Success: false,
			Error:   fmt.Sprintf("Cannot perform PCA: data contains %d missing values. Please exclude columns with missing values or handle them before analysis.", nanCount),
		}
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
	
	return PCAResponse{
		Success: true,
		Result:  result,
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

// SaveFile handles saving exported plot data
func (a *App) SaveFile(fileName string, dataURL string) error {
	// Show save dialog
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: fileName,
		Title:          "Save Plot",
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
func (a *App) LoadIrisDataset() (*FileData, error) {
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

