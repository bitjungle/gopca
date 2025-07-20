package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/bitjungle/complab/internal/core"
	"github.com/bitjungle/complab/internal/utils"
	"github.com/bitjungle/complab/pkg/types"
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
	Headers  []string    `json:"headers"`
	RowNames []string    `json:"rowNames"`
	Data     [][]float64 `json:"data"`
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
	
	// Parse data
	var data [][]float64
	var rowNames []string
	
	for i := 1; i < len(records); i++ {
		record := records[i]
		startIdx := 0
		
		if hasRowNames {
			rowNames = append(rowNames, record[0])
			startIdx = 1
		}
		
		row := make([]float64, len(record)-startIdx)
		for j := startIdx; j < len(record); j++ {
			val, err := strconv.ParseFloat(record[j], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number at row %d, col %d: %s", i, j, record[j])
			}
			row[j-startIdx] = val
		}
		data = append(data, row)
	}
	
	return &FileData{
		Headers:  headers,
		RowNames: rowNames,
		Data:     data,
	}, nil
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
	
	// Filter data if exclusions are provided
	dataToAnalyze := request.Data
	
	if len(request.ExcludedRows) > 0 || len(request.ExcludedColumns) > 0 {
		// Filter the data matrix
		filteredData, err := utils.FilterMatrix(request.Data, request.ExcludedRows, request.ExcludedColumns)
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
	
	// Perform PCA
	engine := core.NewPCAEngine()
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
	
	return PCAResponse{
		Success: true,
		Result:  result,
	}
}

