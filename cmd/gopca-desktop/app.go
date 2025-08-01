package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitjungle/gopca/internal/cli"
	"github.com/bitjungle/gopca/internal/config"
	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/datasets"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/internal/version"
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

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return version.Get().Short()
}

// CalculateEllipsesRequest represents a request to calculate confidence ellipses
type CalculateEllipsesRequest struct {
	Scores      [][]float64 `json:"scores"`
	GroupLabels []string    `json:"groupLabels"`
	XComponent  int         `json:"xComponent"`
	YComponent  int         `json:"yComponent"`
}

// CalculateEllipsesResponse represents the response with calculated ellipses
type CalculateEllipsesResponse struct {
	GroupEllipses90 map[string]EllipseParams `json:"groupEllipses90"`
	GroupEllipses95 map[string]EllipseParams `json:"groupEllipses95"`
	GroupEllipses99 map[string]EllipseParams `json:"groupEllipses99"`
	Success         bool                     `json:"success"`
	Error           string                   `json:"error,omitempty"`
}

// CalculateEllipses calculates confidence ellipses for given scores and groups
func (a *App) CalculateEllipses(request CalculateEllipsesRequest) CalculateEllipsesResponse {
	// Validate input
	if len(request.Scores) == 0 || len(request.GroupLabels) == 0 {
		return CalculateEllipsesResponse{
			Success: false,
			Error:   "Invalid input: scores and group labels are required",
		}
	}

	if len(request.Scores) != len(request.GroupLabels) {
		return CalculateEllipsesResponse{
			Success: false,
			Error:   fmt.Sprintf("Scores and group labels must have the same length (scores: %d, labels: %d)", len(request.Scores), len(request.GroupLabels)),
		}
	}

	// Validate scores structure
	if len(request.Scores[0]) == 0 {
		return CalculateEllipsesResponse{
			Success: false,
			Error:   "Scores matrix has no columns",
		}
	}

	// Check component indices
	maxComponent := len(request.Scores[0]) - 1
	if request.XComponent < 0 || request.XComponent > maxComponent || request.YComponent < 0 || request.YComponent > maxComponent {
		return CalculateEllipsesResponse{
			Success: false,
			Error:   fmt.Sprintf("Component indices out of bounds (x: %d, y: %d, max: %d)", request.XComponent, request.YComponent, maxComponent),
		}
	}

	// Convert scores to matrix
	scoresMatrix := mat.NewDense(len(request.Scores), len(request.Scores[0]), nil)
	for i, row := range request.Scores {
		for j, val := range row {
			scoresMatrix.Set(i, j, val)
		}
	}

	// Calculate ellipses for all three confidence levels
	response := CalculateEllipsesResponse{
		Success:         true,
		GroupEllipses90: make(map[string]EllipseParams),
		GroupEllipses95: make(map[string]EllipseParams),
		GroupEllipses99: make(map[string]EllipseParams),
	}

	confidenceLevels := []float64{0.90, 0.95, 0.99}
	allErrors := []string{}
	for _, confidenceLevel := range confidenceLevels {
		coreEllipses, err := core.CalculateGroupEllipses(scoresMatrix, request.GroupLabels, request.XComponent, request.YComponent, confidenceLevel)
		if err != nil {
			// Log error but continue with other confidence levels
			allErrors = append(allErrors, fmt.Sprintf("%.0f%%: %v", confidenceLevel*100, err))
		}
		if err == nil && len(coreEllipses) > 0 {
			ellipses := make(map[string]EllipseParams)
			for group, ellipse := range coreEllipses {
				ellipses[group] = EllipseParams{
					CenterX:         ellipse.CenterX,
					CenterY:         ellipse.CenterY,
					MajorAxis:       ellipse.MajorAxis,
					MinorAxis:       ellipse.MinorAxis,
					Angle:           ellipse.Angle,
					ConfidenceLevel: ellipse.ConfidenceLevel,
				}
			}

			switch confidenceLevel {
			case 0.90:
				response.GroupEllipses90 = ellipses
			case 0.95:
				response.GroupEllipses95 = ellipses
			case 0.99:
				response.GroupEllipses99 = ellipses
			}
		}
	}

	// If we have some ellipses but also some errors, include a warning
	if len(allErrors) > 0 && (len(response.GroupEllipses90) > 0 || len(response.GroupEllipses95) > 0 || len(response.GroupEllipses99) > 0) {
		// Some ellipses were calculated successfully, just log warnings
		fmt.Printf("Warning: Some ellipse calculations failed: %v\n", allErrors)
	} else if len(allErrors) > 0 && len(response.GroupEllipses90) == 0 && len(response.GroupEllipses95) == 0 && len(response.GroupEllipses99) == 0 {
		// No ellipses were calculated
		response.Success = false
		response.Error = fmt.Sprintf("Failed to calculate any ellipses: %v", allErrors)
	}

	return response
}

// FileData represents the structure of CSV data for the frontend
type FileData struct {
	Headers              []string             `json:"headers"`
	RowNames             []string             `json:"rowNames"`
	Data                 [][]float64          `json:"data"`
	MissingMask          [][]bool             `json:"missingMask,omitempty"`
	CategoricalColumns   map[string][]string  `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]float64 `json:"numericTargetColumns,omitempty"`
}

// PCARequest represents a PCA analysis request from the frontend
type PCARequest struct {
	Data             [][]float64 `json:"data"`
	MissingMask      [][]bool    `json:"missingMask,omitempty"`
	Headers          []string    `json:"headers"`
	RowNames         []string    `json:"rowNames"`
	Components       int         `json:"components"`
	MeanCenter       bool        `json:"meanCenter"`
	StandardScale    bool        `json:"standardScale"`
	RobustScale      bool        `json:"robustScale"`
	ScaleOnly        bool        `json:"scaleOnly"`
	SNV              bool        `json:"snv"`
	VectorNorm       bool        `json:"vectorNorm"`
	Method           string      `json:"method"`
	ExcludedRows     []int       `json:"excludedRows,omitempty"`
	ExcludedColumns  []int       `json:"excludedColumns,omitempty"`
	MissingStrategy  string      `json:"missingStrategy,omitempty"`
	CalculateMetrics bool        `json:"calculateMetrics,omitempty"`
	// Kernel PCA parameters
	KernelType   string  `json:"kernelType,omitempty"`
	KernelGamma  float64 `json:"kernelGamma,omitempty"`
	KernelDegree int     `json:"kernelDegree,omitempty"`
	KernelCoef0  float64 `json:"kernelCoef0,omitempty"`
	// Grouping parameters for confidence ellipses
	GroupColumn string   `json:"groupColumn,omitempty"`
	GroupLabels []string `json:"groupLabels,omitempty"`
	// Metadata for eigencorrelations
	MetadataNumeric            map[string][]float64 `json:"metadataNumeric,omitempty"`
	MetadataCategorical        map[string][]string  `json:"metadataCategorical,omitempty"`
	CalculateEigencorrelations bool                 `json:"calculateEigencorrelations,omitempty"`
}

// EllipseParams represents confidence ellipse parameters for a group
type EllipseParams struct {
	CenterX         float64 `json:"centerX"`
	CenterY         float64 `json:"centerY"`
	MajorAxis       float64 `json:"majorAxis"`
	MinorAxis       float64 `json:"minorAxis"`
	Angle           float64 `json:"angle"`
	ConfidenceLevel float64 `json:"confidenceLevel"`
}

// PCAResponse represents the PCA analysis results
type PCAResponse struct {
	Success         bool                     `json:"success"`
	Error           string                   `json:"error,omitempty"`
	Result          *PCAResultJSON           `json:"result,omitempty"`
	Info            string                   `json:"info,omitempty"`
	GroupEllipses90 map[string]EllipseParams `json:"groupEllipses90,omitempty"`
	GroupEllipses95 map[string]EllipseParams `json:"groupEllipses95,omitempty"`
	GroupEllipses99 map[string]EllipseParams `json:"groupEllipses99,omitempty"`
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

		// Filter group labels if rows are excluded
		if len(request.ExcludedRows) > 0 && len(request.GroupLabels) > 0 {
			newGroupLabels := make([]string, 0)
			for i := 0; i < len(request.GroupLabels); i++ {
				if !contains(request.ExcludedRows, i) {
					newGroupLabels = append(newGroupLabels, request.GroupLabels[i])
				}
			}
			request.GroupLabels = newGroupLabels
		}

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
				Error:   fmt.Sprintf("Missing values detected (%d values). Please select a strategy to handle them: 'drop' to remove rows, 'mean' to impute with column means, or 'native' for NIPALS native handling.", missingInfo.TotalMissing),
			}
		case "native":
			// For native handling with NIPALS, we don't pre-process missing values
			// Validate that NIPALS method is selected
			if strings.ToLower(request.Method) != "nipals" {
				return PCAResponse{
					Success: false,
					Error:   "Native missing value handling is only supported with the NIPALS method",
				}
			}
			// Data remains unchanged, NIPALS will handle missing values internally
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

				// Also filter group labels if provided
				if len(request.GroupLabels) > 0 {
					newGroupLabels := []string{}
					for i := 0; i < len(request.GroupLabels); i++ {
						if keepRows[i] {
							newGroupLabels = append(newGroupLabels, request.GroupLabels[i])
						}
					}
					request.GroupLabels = newGroupLabels
				}
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
		ScaleOnly:       request.ScaleOnly,
		SNV:             request.SNV,
		VectorNorm:      request.VectorNorm,
		Method:          strings.ToLower(request.Method),
		ExcludedRows:    request.ExcludedRows,
		ExcludedColumns: request.ExcludedColumns,
		MissingStrategy: types.MissingValueStrategy(request.MissingStrategy),
	}

	// Add kernel parameters if using kernel PCA
	if strings.ToLower(request.Method) == "kernel" {
		config.KernelType = request.KernelType
		config.KernelGamma = request.KernelGamma
		config.KernelDegree = request.KernelDegree
		config.KernelCoef0 = request.KernelCoef0

		// Skip preprocessing that involves centering for kernel PCA
		// But allow scale-only, SNV, and vector normalization
		if !config.ScaleOnly {
			config.MeanCenter = false
			config.StandardScale = false
			config.RobustScale = false
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

	// Calculate diagnostic metrics if requested
	if request.CalculateMetrics && strings.ToLower(request.Method) != "kernel" {
		// For RSS calculation, we need to use data preprocessed exactly as it was for PCA fitting
		// This ensures the data and reconstruction are in the same space
		preprocessedData := dataToAnalyze

		// Apply the same preprocessing that was used for PCA
		if config.MeanCenter || config.StandardScale || config.RobustScale || config.ScaleOnly || config.SNV || config.VectorNorm {
			// Create preprocessor with the same settings used for PCA
			preprocessor := core.NewPreprocessorWithScaleOnly(config.MeanCenter, config.StandardScale, config.RobustScale, config.ScaleOnly, config.SNV, config.VectorNorm)
			var err error
			preprocessedData, err = preprocessor.FitTransform(dataToAnalyze)
			if err != nil {
				fmt.Printf("Warning: failed to preprocess data for metrics: %v\n", err)
				preprocessedData = dataToAnalyze // Fallback to original data
			}
		}

		// Calculate metrics using the appropriately preprocessed data
		metrics, err := core.CalculateMetricsFromPCAResult(result, preprocessedData)
		if err != nil {
			// Don't fail the whole PCA, just log the error
			fmt.Printf("Warning: failed to calculate diagnostic metrics: %v\n", err)
		} else {
			result.Metrics = metrics

			// Calculate confidence limits
			scores := utils.MatrixToDense(result.Scores)
			loadings := utils.MatrixToDense(result.Loadings)
			calculator := core.NewPCAMetricsCalculator(scores, loadings, result.Means, result.StdDevs)

			// Calculate T² limits
			result.T2Limit95, result.T2Limit99 = calculator.CalculateT2Limits()

			// For Q limits calculation, we need all eigenvalues including non-retained ones
			// Since we don't have them in the current implementation, we'll use a simplified approach
			result.QLimit95 = 0.0
			result.QLimit99 = 0.0
		}
	}

	// Calculate eigencorrelations if requested
	if request.CalculateEigencorrelations && (len(request.MetadataNumeric) > 0 || len(request.MetadataCategorical) > 0) {
		fmt.Println("Calculating eigencorrelations...")

		// Convert scores to mat.Matrix
		scoresMatrix := mat.NewDense(len(result.Scores), len(result.Scores[0]), nil)
		for i, row := range result.Scores {
			for j, val := range row {
				scoresMatrix.Set(i, j, val)
			}
		}

		// Create correlation request
		corrRequest := core.CorrelationRequest{
			Scores:              scoresMatrix,
			MetadataNumeric:     request.MetadataNumeric,
			MetadataCategorical: request.MetadataCategorical,
			Components:          nil,       // Use all components
			Method:              "pearson", // Default to Pearson
		}

		// Calculate correlations
		corrResult, err := core.CalculateEigencorrelations(corrRequest)
		if err != nil {
			fmt.Printf("Warning: failed to calculate eigencorrelations: %v\n", err)
		} else {
			result.Eigencorrelations = &types.EigencorrelationResult{
				Correlations: corrResult.Correlations,
				PValues:      corrResult.PValues,
				Variables:    corrResult.Variables,
				Components:   corrResult.Components,
				Method:       "pearson",
			}
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

	// Calculate confidence ellipses for all confidence levels if groups are provided
	var groupEllipses90, groupEllipses95, groupEllipses99 map[string]EllipseParams
	if len(request.GroupLabels) > 0 && len(result.Scores) > 0 {
		// Convert scores to matrix once
		scoresMatrix := mat.NewDense(len(result.Scores), len(result.Scores[0]), nil)
		for i, row := range result.Scores {
			for j, val := range row {
				scoresMatrix.Set(i, j, val)
			}
		}

		// Calculate ellipses for all three confidence levels (default to PC1 vs PC2)
		confidenceLevels := []float64{0.90, 0.95, 0.99}
		for _, confidenceLevel := range confidenceLevels {
			coreEllipses, err := core.CalculateGroupEllipses(scoresMatrix, request.GroupLabels, 0, 1, confidenceLevel)
			if err == nil && len(coreEllipses) > 0 {
				ellipses := make(map[string]EllipseParams)
				for group, ellipse := range coreEllipses {
					ellipses[group] = EllipseParams{
						CenterX:         ellipse.CenterX,
						CenterY:         ellipse.CenterY,
						MajorAxis:       ellipse.MajorAxis,
						MinorAxis:       ellipse.MinorAxis,
						Angle:           ellipse.Angle,
						ConfidenceLevel: ellipse.ConfidenceLevel,
					}
				}

				// Assign to appropriate variable
				switch confidenceLevel {
				case 0.90:
					groupEllipses90 = ellipses
				case 0.95:
					groupEllipses95 = ellipses
				case 0.99:
					groupEllipses99 = ellipses
				}
			}
		}
	}

	return PCAResponse{
		Success:         true,
		Result:          ConvertPCAResultToJSON(result),
		Info:            infoMsg,
		GroupEllipses90: groupEllipses90,
		GroupEllipses95: groupEllipses95,
		GroupEllipses99: groupEllipses99,
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

// LoadDatasetFile loads a CSV file from the embedded data
func (a *App) LoadDatasetFile(filename string) (*FileDataJSON, error) {
	// First try to get the embedded dataset
	if content, ok := datasets.GetDataset(filename); ok {
		return a.ParseCSV(content)
	}

	// If not found in embedded data, try file system as fallback
	// This is useful during development
	possiblePaths := []string{
		filepath.Join("data", filename),
		filepath.Join("..", "..", "data", filename),
		filepath.Join("../../data", filename),
	}

	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return a.ParseCSV(string(content))
		}
	}

	return nil, fmt.Errorf("dataset file not found: %s", filename)
}

// PCAConfig represents PCA configuration from the frontend
type PCAConfig struct {
	Components       int    `json:"components"`
	MeanCenter       bool   `json:"meanCenter"`
	StandardScale    bool   `json:"standardScale"`
	RobustScale      bool   `json:"robustScale"`
	ScaleOnly        bool   `json:"scaleOnly"`
	SNV              bool   `json:"snv"`
	VectorNorm       bool   `json:"vectorNorm"`
	Method           string `json:"method"`
	MissingStrategy  string `json:"missingStrategy"`
	CalculateMetrics bool   `json:"calculateMetrics"`
	// Kernel PCA parameters
	KernelType   string  `json:"kernelType,omitempty"`
	KernelGamma  float64 `json:"kernelGamma,omitempty"`
	KernelDegree int     `json:"kernelDegree,omitempty"`
	KernelCoef0  float64 `json:"kernelCoef0,omitempty"`
}

// ExportPCAModelRequest contains the data needed to export a PCA model
type ExportPCAModelRequest struct {
	Data            [][]float64      `json:"data"`
	Headers         []string         `json:"headers"`
	RowNames        []string         `json:"rowNames"`
	PCAResult       *types.PCAResult `json:"pcaResult"`
	Config          PCAConfig        `json:"config"`
	ExcludedRows    []int            `json:"excludedRows"`
	ExcludedColumns []int            `json:"excludedColumns"`
}

// ExportPCAModel exports the complete PCA model to a JSON file
func (a *App) ExportPCAModel(request ExportPCAModelRequest) error {
	// Show save dialog
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "pca_model.json",
		Title:           "Export PCA Model",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files",
				Pattern:     "*.json",
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

	// Convert PCA config to types.PCAConfig
	pcaConfig := types.PCAConfig{
		Components:      request.Config.Components,
		MeanCenter:      request.Config.MeanCenter,
		StandardScale:   request.Config.StandardScale,
		RobustScale:     request.Config.RobustScale,
		ScaleOnly:       request.Config.ScaleOnly,
		SNV:             request.Config.SNV,
		VectorNorm:      request.Config.VectorNorm,
		Method:          request.Config.Method,
		ExcludedRows:    request.ExcludedRows,
		ExcludedColumns: request.ExcludedColumns,
		MissingStrategy: types.MissingValueStrategy(request.Config.MissingStrategy),
		KernelType:      request.Config.KernelType,
		KernelGamma:     request.Config.KernelGamma,
		KernelDegree:    request.Config.KernelDegree,
		KernelCoef0:     request.Config.KernelCoef0,
	}

	// Create preprocessor to get the preprocessing parameters
	preprocessor := core.NewPreprocessorWithScaleOnly(
		pcaConfig.MeanCenter,
		pcaConfig.StandardScale,
		pcaConfig.RobustScale,
		pcaConfig.ScaleOnly,
		pcaConfig.SNV,
		pcaConfig.VectorNorm,
	)

	// We need to fit the preprocessor to get the parameters
	// But we already have them in the PCAResult
	// Create a mock CSVData structure for the output conversion
	csvData := &cli.CSVData{
		CSVData: &types.CSVData{
			Headers:  request.Headers,
			RowNames: request.RowNames,
			Matrix:   request.Data,
			Rows:     len(request.Data),
			Columns:  len(request.Headers),
		},
	}

	// Convert to PCAOutputData using the CLI function
	outputData := cli.ConvertToPCAOutputData(request.PCAResult, csvData, true, pcaConfig, preprocessor)

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal model data: %v", err)
	}

	// Write to file
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write model file: %v", err)
	}

	return nil
}

// GetGUIConfig returns the GUI configuration
func (a *App) GetGUIConfig() *config.GUIConfig {
	return config.DefaultGUIConfig()
}
