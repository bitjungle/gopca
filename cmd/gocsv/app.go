package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"github.com/bitjungle/gopca/pkg/types"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx     context.Context
	history *CommandHistory
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		history: NewCommandHistory(100), // Keep last 100 commands
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}


// ValidationResult represents the result of GoPCA validation
type ValidationResult struct {
	IsValid  bool     `json:"isValid"`
	Messages []string `json:"messages"`
}

// LoadCSV loads a CSV file and returns its data
func (a *App) LoadCSV(filePath string) (*FileData, error) {
	// If no filepath provided, show file dialog
	if filePath == "" {
		selection, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
			Title: "Select CSV File",
			Filters: []wailsruntime.FileFilter{
				{
					DisplayName: "Supported Files (*.csv,*.xlsx,*.xls,*.tsv)",
					Pattern:     "*.csv;*.xlsx;*.xls;*.tsv",
				},
				{
					DisplayName: "CSV Files (*.csv)",
					Pattern:     "*.csv",
				},
				{
					DisplayName: "Excel Files (*.xlsx,*.xls)",
					Pattern:     "*.xlsx;*.xls",
				},
				{
					DisplayName: "TSV Files (*.tsv)",
					Pattern:     "*.tsv",
				},
				{
					DisplayName: "All Files (*.*)",
					Pattern:     "*.*",
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error showing file dialog: %w", err)
		}
		if selection == "" {
			return nil, fmt.Errorf("no file selected")
		}
		filePath = selection
	}

	// Check file extension
	ext := filepath.Ext(filePath)
	var fileData *FileData
	
	switch ext {
	case ".xlsx", ".xls":
		// Handle Excel files
		var err error
		fileData, err = a.loadExcel(filePath)
		if err != nil {
			return nil, fmt.Errorf("error loading Excel file: %w", err)
		}
	case ".tsv", ".csv", "":
		// Handle CSV/TSV files
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
		
		// Check file size
		if len(content) > 100*1024*1024 { // 100MB
			wailsruntime.LogWarning(a.ctx, fmt.Sprintf("Large file detected: %d MB", len(content)/1024/1024))
		}
		
		// Parse using GoPCA's parser with format detection
		fileData, err = a.parseCSVContent(string(content), ext)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Store the filename for display
	wailsruntime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))

	return fileData, nil
}

// loadExcel loads data from an Excel file
func (a *App) loadExcel(filePath string) (*FileData, error) {
	wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Loading Excel file: %s", filePath))
	
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()
	
	// Get list of sheets
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}
	
	// For now, use the first sheet. TODO: Add sheet selection dialog
	selectedSheet := sheets[0]
	if len(sheets) > 1 {
		wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Multiple sheets found. Using first sheet: %s", selectedSheet))
	}
	
	// Get all rows from the selected sheet
	rows, err := f.GetRows(selectedSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %w", selectedSheet, err)
	}
	
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in sheet %s", selectedSheet)
	}
	
	// Convert Excel data to CSV format for parsing
	var csvContent strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				csvContent.WriteString(",")
			}
			// Quote cells that contain commas or quotes
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") {
				csvContent.WriteString("\"")
				csvContent.WriteString(strings.ReplaceAll(cell, "\"", "\"\""))
				csvContent.WriteString("\"")
			} else {
				csvContent.WriteString(cell)
			}
		}
		csvContent.WriteString("\n")
	}
	
	// Parse the CSV content using GoPCA's parser
	wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Excel data converted to CSV, %d bytes", csvContent.Len()))
	return a.parseCSVContent(csvContent.String(), ".csv")
}

// parseCSVContent parses CSV content using GoPCA's parser
func (a *App) parseCSVContent(content string, ext string) (*FileData, error) {
	// Configure format based on file extension
	defaultFormat := types.DefaultCSVFormat()
	formats := []types.CSVFormat{
		defaultFormat, // Standard CSV: comma with dot decimal
	}
	
	// Add TSV format if TSV file
	if ext == ".tsv" {
		formats = []types.CSVFormat{
			{
				FieldDelimiter:   '\t',
				DecimalSeparator: '.',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       defaultFormat.NullValues,
			},
		}
	} else {
		// Try multiple CSV formats
		formats = append(formats, 
			types.CSVFormat{
				FieldDelimiter:   ';',
				DecimalSeparator: ',',
				HasHeaders:       true,
				HasRowNames:      true,
				NullValues:       defaultFormat.NullValues,
			},
		)
	}

	var csvData *types.CSVData
	var categoricalData map[string][]string
	var numericTargetData map[string][]float64
	var lastErr error

	// Try each format until one works
	for _, format := range formats {
		reader := strings.NewReader(content)
		data, catData, targetData, err := types.ParseCSVMixedWithTargets(reader, format, nil)
		if err == nil && data != nil && data.Columns > 0 {
			csvData = data
			categoricalData = catData
			numericTargetData = targetData
			break
		}
		if err != nil {
			lastErr = err
		}
	}

	if csvData == nil {
		if lastErr != nil {
			wailsruntime.LogError(a.ctx, fmt.Sprintf("Failed to parse CSV: %v", lastErr))
			return nil, fmt.Errorf("failed to parse CSV: %w", lastErr)
		}
		wailsruntime.LogError(a.ctx, "No data found in file")
		return nil, fmt.Errorf("no data found in file")
	}

	// Convert numeric matrix to string matrix for display
	stringData := make([][]string, len(csvData.Matrix))
	for i, row := range csvData.Matrix {
		stringData[i] = make([]string, len(row))
		for j, val := range row {
			if csvData.MissingMask != nil && csvData.MissingMask[i][j] {
				stringData[i][j] = ""
			} else {
				stringData[i][j] = strconv.FormatFloat(val, 'g', -1, 64)
			}
		}
	}

	// Build column types map
	columnTypes := make(map[string]string)
	
	// Mark numeric columns
	for _, header := range csvData.Headers {
		columnTypes[header] = "numeric"
	}
	
	// Mark categorical columns
	for colName := range categoricalData {
		columnTypes[colName] = "categorical"
	}
	
	// Mark target columns
	for colName := range numericTargetData {
		columnTypes[colName] = "target"
	}

	// Create FileData with all information
	fileData := &FileData{
		Headers:              csvData.Headers,
		RowNames:             csvData.RowNames,
		Data:                 stringData,
		Rows:                 csvData.Rows,
		Columns:              csvData.Columns,
		CategoricalColumns:   categoricalData,
		NumericTargetColumns: ConvertFloat64MapToJSON(numericTargetData),
		ColumnTypes:          columnTypes,
	}
	
	wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Parsed data: %d rows, %d columns, %d headers", csvData.Rows, csvData.Columns, len(csvData.Headers)))

	// If we have categorical or target columns, we need to combine them with numeric data
	// for the full data display
	if len(categoricalData) > 0 || len(numericTargetData) > 0 {
		fileData = a.combineAllColumns(csvData, categoricalData, numericTargetData)
	}

	return fileData, nil
}

// combineAllColumns combines numeric, categorical, and target columns for display
func (a *App) combineAllColumns(csvData *types.CSVData, categoricalData map[string][]string, numericTargetData map[string][]float64) *FileData {
	// Start with numeric columns from csvData
	allHeaders := make([]string, 0)
	allData := make([][]string, csvData.Rows)
	columnTypes := make(map[string]string)
	
	// Initialize rows
	for i := range allData {
		allData[i] = make([]string, 0)
	}
	
	// Add numeric columns
	for colIdx, header := range csvData.Headers {
		allHeaders = append(allHeaders, header)
		columnTypes[header] = "numeric"
		
		for rowIdx := 0; rowIdx < csvData.Rows; rowIdx++ {
			if csvData.MissingMask != nil && csvData.MissingMask[rowIdx][colIdx] {
				allData[rowIdx] = append(allData[rowIdx], "")
			} else {
				allData[rowIdx] = append(allData[rowIdx], strconv.FormatFloat(csvData.Matrix[rowIdx][colIdx], 'g', -1, 64))
			}
		}
	}
	
	// Add categorical columns
	for colName, values := range categoricalData {
		allHeaders = append(allHeaders, colName)
		columnTypes[colName] = "categorical"
		
		for rowIdx, value := range values {
			if rowIdx < len(allData) {
				allData[rowIdx] = append(allData[rowIdx], value)
			}
		}
	}
	
	// Add numeric target columns
	for colName, values := range numericTargetData {
		allHeaders = append(allHeaders, colName)
		columnTypes[colName] = "target"
		
		for rowIdx, value := range values {
			if rowIdx < len(allData) {
				allData[rowIdx] = append(allData[rowIdx], strconv.FormatFloat(value, 'g', -1, 64))
			}
		}
	}
	
	return &FileData{
		Headers:              allHeaders,
		RowNames:             csvData.RowNames,
		Data:                 allData,
		Rows:                 csvData.Rows,
		Columns:              len(allHeaders),
		CategoricalColumns:   categoricalData,
		NumericTargetColumns: ConvertFloat64MapToJSON(numericTargetData),
		ColumnTypes:          columnTypes,
	}
}

// ValidateForGoPCA validates that the CSV data is compatible with GoPCA
func (a *App) ValidateForGoPCA(data *FileData) *ValidationResult {
	var warnings []string
	var numericColumns int
	var categoricalColumns int
	var targetColumns int
	var totalMissing int

	// Check minimum data requirements
	if data.Rows < 2 {
		warnings = append(warnings, "ERROR: Data must have at least 2 rows (found "+fmt.Sprintf("%d", data.Rows)+")")
	}

	// Count column types using our pre-detected types
	for _, colType := range data.ColumnTypes {
		switch colType {
		case "numeric":
			numericColumns++
		case "categorical":
			categoricalColumns++
		case "target":
			targetColumns++
		}
	}

	// Check for missing values in the data
	for colIdx, header := range data.Headers {
		missingInCol := 0
		
		for i := 0; i < data.Rows; i++ {
			if i >= len(data.Data) {
				break
			}
			value := data.Data[i][colIdx]
			
			// Check for missing values
			trimmed := strings.TrimSpace(value)
			if trimmed == "" || trimmed == "NA" || trimmed == "N/A" || 
			   trimmed == "nan" || trimmed == "NaN" || trimmed == "null" || trimmed == "NULL" {
				missingInCol++
				totalMissing++
			}
		}
		
		// Report high missing value percentage
		if data.Rows > 0 {
			missingPercent := float64(missingInCol) / float64(data.Rows) * 100
			if missingPercent > 50 {
				warnings = append(warnings, fmt.Sprintf("WARNING: Column '%s' has %.1f%% missing values", header, missingPercent))
			}
		}
	}

	// Report column type summary
	if categoricalColumns > 0 {
		warnings = append(warnings, fmt.Sprintf("INFO: %d categorical column(s) detected - these will be excluded from PCA but available for visualization", categoricalColumns))
	}
	
	if targetColumns > 0 {
		warnings = append(warnings, fmt.Sprintf("INFO: %d target column(s) detected - these will be excluded from PCA but available for visualization", targetColumns))
	}

	// Check if we have enough numeric columns for PCA
	if numericColumns < 2 {
		warnings = append(warnings, fmt.Sprintf("ERROR: Need at least 2 numeric columns for PCA (found %d)", numericColumns))
	} else if numericColumns < 3 {
		warnings = append(warnings, fmt.Sprintf("WARNING: Only %d numeric columns found - PCA results may be limited", numericColumns))
	} else {
		warnings = append(warnings, fmt.Sprintf("INFO: %d numeric columns will be used for PCA analysis", numericColumns))
	}

	// Report overall missing data
	totalCells := data.Rows * data.Columns
	if totalCells > 0 {
		missingPercent := float64(totalMissing) / float64(totalCells) * 100
		if missingPercent > 0 {
			warnings = append(warnings, fmt.Sprintf("INFO: Dataset contains %.1f%% missing values (%d cells)", missingPercent, totalMissing))
		}
	}

	// Check for reasonable data size
	if data.Rows > 10000 {
		warnings = append(warnings, fmt.Sprintf("INFO: Large dataset detected (%d rows) - processing may take time", data.Rows))
	}

	// Check if row names were detected
	if len(data.RowNames) > 0 {
		warnings = append(warnings, "INFO: Row names detected in first column")
	}
	
	// Determine if data is valid (no ERRORs)
	isValid := true
	for _, warning := range warnings {
		if strings.HasPrefix(warning, "ERROR:") {
			isValid = false
			break
		}
	}

	return &ValidationResult{
		IsValid:  isValid,
		Messages: warnings,
	}
}

// SaveCSV saves the data to a CSV file
func (a *App) SaveCSV(data *FileData) error {
	// Show save dialog
	selection, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title: "Save CSV File",
		DefaultFilename: "exported_data.csv",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "CSV Files (*.csv)",
				Pattern:     "*.csv",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error showing save dialog: %w", err)
	}
	if selection == "" {
		return fmt.Errorf("no file selected")
	}

	// Create the file
	file, err := os.Create(selection)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers with row name column if present
	headers := data.Headers
	if len(data.RowNames) > 0 {
		// Add empty header for row names column
		headers = append([]string{""}, headers...)
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	// Write data with row names
	for i, row := range data.Data {
		outputRow := row
		if len(data.RowNames) > 0 && i < len(data.RowNames) {
			// Prepend row name to the row
			outputRow = append([]string{data.RowNames[i]}, row...)
		}
		if err := writer.Write(outputRow); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	wailsruntime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
	return nil
}

// SaveExcel saves data to an Excel file
func (a *App) SaveExcel(data *FileData) error {
	// Show save dialog
	selection, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title: "Save Excel File",
		DefaultFilename: "exported_data.xlsx",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Excel Files (*.xlsx)",
				Pattern:     "*.xlsx",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error showing save dialog: %w", err)
	}
	if selection == "" {
		return fmt.Errorf("no file selected")
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()
	
	// Create a new sheet
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}
	
	// Write headers with row names if present
	headers := data.Headers
	if len(data.RowNames) > 0 {
		// Add row name header
		headers = append([]string{"RowName"}, headers...)
	}
	
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return fmt.Errorf("failed to get cell coordinate: %w", err)
		}
		f.SetCellValue(sheetName, cell, header)
		
		// Style headers
		style, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Pattern: 1,
				Color:   []string{"#E0E0E0"},
			},
		})
		if err == nil {
			f.SetCellStyle(sheetName, cell, cell, style)
		}
	}
	
	// Write data rows
	for rowIdx, row := range data.Data {
		excelRow := rowIdx + 2 // Excel rows are 1-indexed, plus header row
		
		// Write row name if present
		colOffset := 0
		if len(data.RowNames) > 0 && rowIdx < len(data.RowNames) {
			cell, err := excelize.CoordinatesToCellName(1, excelRow)
			if err == nil {
				f.SetCellValue(sheetName, cell, data.RowNames[rowIdx])
			}
			colOffset = 1
		}
		
		// Write data cells
		for colIdx, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1+colOffset, excelRow)
			if err != nil {
				continue
			}
			
			// Try to convert to number if possible
			if num, err := strconv.ParseFloat(value, 64); err == nil && value != "" {
				f.SetCellValue(sheetName, cell, num)
			} else {
				f.SetCellValue(sheetName, cell, value)
			}
			
			// Apply column type styling
			if data.ColumnTypes != nil {
				header := data.Headers[colIdx]
				if colType, exists := data.ColumnTypes[header]; exists {
					var style int
					switch colType {
					case "target":
						// Light yellow background for target columns
						style, _ = f.NewStyle(&excelize.Style{
							Fill: excelize.Fill{
								Type:    "pattern",
								Pattern: 1,
								Color:   []string{"#FFFFCC"},
							},
						})
					case "categorical":
						// Light blue background for categorical columns
						style, _ = f.NewStyle(&excelize.Style{
							Fill: excelize.Fill{
								Type:    "pattern",
								Pattern: 1,
								Color:   []string{"#E6F3FF"},
							},
						})
					}
					if style > 0 {
						f.SetCellStyle(sheetName, cell, cell, style)
					}
				}
			}
		}
	}
	
	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, 12)
	}
	
	// Set active sheet
	f.SetActiveSheet(index)
	
	// Save file
	if err := f.SaveAs(selection); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}
	
	wailsruntime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
	return nil
}

// MissingValueStats represents statistics about missing values in the data
type MissingValueStats struct {
	TotalCells     int                       `json:"totalCells"`
	MissingCells   int                       `json:"missingCells"`
	MissingPercent float64                   `json:"missingPercent"`
	ColumnStats    map[string]*ColumnMissing `json:"columnStats"`
	RowStats       map[int]*RowMissing       `json:"rowStats"`
}

// ColumnMissing represents missing value stats for a column
type ColumnMissing struct {
	Name           string  `json:"name"`
	TotalValues    int     `json:"totalValues"`
	MissingValues  int     `json:"missingValues"`
	MissingPercent float64 `json:"missingPercent"`
	Pattern        string  `json:"pattern"` // "random", "systematic", "top", "bottom"
}

// RowMissing represents missing value stats for a row
type RowMissing struct {
	Index          int     `json:"index"`
	TotalValues    int     `json:"totalValues"`
	MissingValues  int     `json:"missingValues"`
	MissingPercent float64 `json:"missingPercent"`
}

// AnalyzeMissingValues analyzes missing value patterns in the data
func (a *App) AnalyzeMissingValues(data *FileData) *MissingValueStats {
	if data == nil || len(data.Data) == 0 {
		return &MissingValueStats{
			ColumnStats: make(map[string]*ColumnMissing),
			RowStats:    make(map[int]*RowMissing),
		}
	}

	stats := &MissingValueStats{
		TotalCells:  data.Rows * data.Columns,
		ColumnStats: make(map[string]*ColumnMissing),
		RowStats:    make(map[int]*RowMissing),
	}

	// Analyze by column
	for colIdx, header := range data.Headers {
		colStats := &ColumnMissing{
			Name:        header,
			TotalValues: data.Rows,
		}

		missingIndices := []int{}
		for rowIdx := 0; rowIdx < data.Rows; rowIdx++ {
			if rowIdx >= len(data.Data) || colIdx >= len(data.Data[rowIdx]) {
				continue
			}
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				colStats.MissingValues++
				stats.MissingCells++
				missingIndices = append(missingIndices, rowIdx)
			}
		}

		if colStats.TotalValues > 0 {
			colStats.MissingPercent = float64(colStats.MissingValues) / float64(colStats.TotalValues) * 100
		}

		// Detect pattern
		colStats.Pattern = detectMissingPattern(missingIndices, data.Rows)
		stats.ColumnStats[header] = colStats
	}

	// Analyze by row
	for rowIdx := 0; rowIdx < data.Rows; rowIdx++ {
		rowStats := &RowMissing{
			Index:       rowIdx,
			TotalValues: data.Columns,
		}

		if rowIdx < len(data.Data) {
			for colIdx := 0; colIdx < data.Columns && colIdx < len(data.Data[rowIdx]); colIdx++ {
				value := strings.TrimSpace(data.Data[rowIdx][colIdx])
				if isMissingValue(value) {
					rowStats.MissingValues++
				}
			}
		}

		if rowStats.TotalValues > 0 {
			rowStats.MissingPercent = float64(rowStats.MissingValues) / float64(rowStats.TotalValues) * 100
		}

		// Only include rows with missing values
		if rowStats.MissingValues > 0 {
			stats.RowStats[rowIdx] = rowStats
		}
	}

	// Calculate overall percentage
	if stats.TotalCells > 0 {
		stats.MissingPercent = float64(stats.MissingCells) / float64(stats.TotalCells) * 100
	}

	return stats
}


// detectMissingPattern analyzes the pattern of missing values
func detectMissingPattern(missingIndices []int, totalRows int) string {
	if len(missingIndices) == 0 {
		return "none"
	}

	// Check if all missing are at the top
	allTop := true
	for i, idx := range missingIndices {
		if idx != i {
			allTop = false
			break
		}
	}
	if allTop {
		return "top"
	}

	// Check if all missing are at the bottom
	allBottom := true
	startIdx := totalRows - len(missingIndices)
	for i, idx := range missingIndices {
		if idx != startIdx+i {
			allBottom = false
			break
		}
	}
	if allBottom {
		return "bottom"
	}

	// Check for systematic pattern (regular intervals)
	if len(missingIndices) > 2 {
		intervals := []int{}
		for i := 1; i < len(missingIndices); i++ {
			intervals = append(intervals, missingIndices[i]-missingIndices[i-1])
		}
		
		// Check if all intervals are the same
		systematic := true
		if len(intervals) > 0 {
			firstInterval := intervals[0]
			for _, interval := range intervals[1:] {
				if interval != firstInterval {
					systematic = false
					break
				}
			}
			if systematic && firstInterval > 1 {
				return "systematic"
			}
		}
	}

	return "random"
}

// FillMissingValuesRequest represents a request to fill missing values
type FillMissingValuesRequest struct {
	Strategy string `json:"strategy"` // "mean", "median", "mode", "forward", "backward", "custom"
	Column   string `json:"column"`   // Column name, or empty for all columns
	Value    string `json:"value"`    // Custom value for "custom" strategy
}

// FillMissingValues fills missing values in the data according to the specified strategy
func (a *App) FillMissingValues(data *FileData, request FillMissingValuesRequest) (*FileData, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to process")
	}

	// Clone the data to avoid modifying the original
	result := &FileData{
		Headers:              data.Headers,
		RowNames:             data.RowNames,
		Data:                 make([][]string, len(data.Data)),
		Rows:                 data.Rows,
		Columns:              data.Columns,
		CategoricalColumns:   data.CategoricalColumns,
		NumericTargetColumns: data.NumericTargetColumns,
		ColumnTypes:          data.ColumnTypes,
	}

	// Deep copy the data
	for i := range data.Data {
		result.Data[i] = make([]string, len(data.Data[i]))
		copy(result.Data[i], data.Data[i])
	}

	// Determine which columns to process
	columnsToProcess := []int{}
	if request.Column == "" {
		// Process all columns
		for i := 0; i < data.Columns; i++ {
			columnsToProcess = append(columnsToProcess, i)
		}
	} else {
		// Find the specific column
		found := false
		for i, header := range data.Headers {
			if header == request.Column {
				columnsToProcess = append(columnsToProcess, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column '%s' not found", request.Column)
		}
	}

	// Apply the fill strategy
	for _, colIdx := range columnsToProcess {
		switch request.Strategy {
		case "mean":
			fillWithMean(result, colIdx)
		case "median":
			fillWithMedian(result, colIdx)
		case "mode":
			fillWithMode(result, colIdx)
		case "forward":
			fillForward(result, colIdx)
		case "backward":
			fillBackward(result, colIdx)
		case "custom":
			fillWithCustomValue(result, colIdx, request.Value)
		default:
			return nil, fmt.Errorf("unknown fill strategy: %s", request.Strategy)
		}
	}

	return result, nil
}

// fillWithMean fills missing values with the column mean (numeric columns only)
func fillWithMean(data *FileData, colIdx int) {
	if colIdx >= len(data.Headers) {
		return
	}

	// Check if column is numeric
	colType := "numeric"
	if data.ColumnTypes != nil {
		if t, exists := data.ColumnTypes[data.Headers[colIdx]]; exists {
			colType = t
		}
	}

	if colType != "numeric" {
		// For non-numeric columns, use mode instead
		fillWithMode(data, colIdx)
		return
	}

	// Use utility function to fill with mean
	fillMissingWithMean(data.Data, colIdx)
}

// fillWithMedian fills missing values with the column median (numeric columns only)
func fillWithMedian(data *FileData, colIdx int) {
	if colIdx >= len(data.Headers) {
		return
	}

	// Check if column is numeric
	colType := "numeric"
	if data.ColumnTypes != nil {
		if t, exists := data.ColumnTypes[data.Headers[colIdx]]; exists {
			colType = t
		}
	}

	if colType != "numeric" {
		// For non-numeric columns, use mode instead
		fillWithMode(data, colIdx)
		return
	}

	// Use utility function to fill with median
	fillMissingWithMedian(data.Data, colIdx)
}

// fillWithMode fills missing values with the most frequent value
func fillWithMode(data *FileData, colIdx int) {
	if colIdx >= len(data.Headers) {
		return
	}

	// Count occurrences of each value
	valueCounts := make(map[string]int)
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if !isMissingValue(value) {
				valueCounts[value]++
			}
		}
	}

	if len(valueCounts) == 0 {
		return // No valid values
	}

	// Find mode (most frequent value)
	mode := ""
	maxCount := 0
	for value, count := range valueCounts {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	// Fill missing values
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				data.Data[rowIdx][colIdx] = mode
			}
		}
	}
}

// fillForward fills missing values with the previous non-missing value
func fillForward(data *FileData, colIdx int) {
	if colIdx >= len(data.Headers) {
		return
	}

	lastValidValue := ""
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				if lastValidValue != "" {
					data.Data[rowIdx][colIdx] = lastValidValue
				}
			} else {
				lastValidValue = value
			}
		}
	}
}

// fillBackward fills missing values with the next non-missing value
func fillBackward(data *FileData, colIdx int) {
	if colIdx >= len(data.Headers) {
		return
	}

	lastValidValue := ""
	for rowIdx := data.Rows - 1; rowIdx >= 0 && rowIdx < len(data.Data); rowIdx-- {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				if lastValidValue != "" {
					data.Data[rowIdx][colIdx] = lastValidValue
				}
			} else {
				lastValidValue = value
			}
		}
	}
}

// fillWithCustomValue fills missing values with a custom value
func fillWithCustomValue(data *FileData, colIdx int, customValue string) {
	if colIdx >= len(data.Headers) {
		return
	}

	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				data.Data[rowIdx][colIdx] = customValue
			}
		}
	}
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return "1.0.0"
}

// DataQualityReport represents a comprehensive data quality analysis
type DataQualityReport struct {
	DataProfile     DataProfile      `json:"dataProfile"`
	ColumnAnalysis  []ColumnAnalysis `json:"columnAnalysis"`
	QualityScore    float64          `json:"qualityScore"`
	Issues          []QualityIssue   `json:"issues"`
	Recommendations []Recommendation `json:"recommendations"`
}

// DataProfile contains overall dataset statistics
type DataProfile struct {
	Rows              int     `json:"rows"`
	Columns           int     `json:"columns"`
	NumericColumns    int     `json:"numericColumns"`
	CategoricalColumns int    `json:"categoricalColumns"`
	TargetColumns     int     `json:"targetColumns"`
	MissingPercent    float64 `json:"missingPercent"`
	DuplicateRows     int     `json:"duplicateRows"`
	MemorySize        string  `json:"memorySize"` // Estimated memory usage
}

// ColumnAnalysis contains detailed analysis for each column
type ColumnAnalysis struct {
	Name         string           `json:"name"`
	Type         string           `json:"type"` // "numeric", "categorical", "target"
	Stats        ColumnStatistics `json:"stats"`
	Distribution DistributionInfo `json:"distribution"`
	Outliers     []OutlierInfo    `json:"outliers"`
	QualityScore float64          `json:"qualityScore"`
}

// ColumnStatistics contains statistical measures for a column
type ColumnStatistics struct {
	Count          int              `json:"count"`
	Missing        int              `json:"missing"`
	MissingPercent float64          `json:"missingPercent"`
	Unique         int              `json:"unique"`
	Mean           *float64         `json:"mean,omitempty"`
	Median         *float64         `json:"median,omitempty"`
	Mode           *string          `json:"mode,omitempty"`
	StdDev         *float64         `json:"stdDev,omitempty"`
	Min            *float64         `json:"min,omitempty"`
	Max            *float64         `json:"max,omitempty"`
	Q1             *float64         `json:"q1,omitempty"`
	Q3             *float64         `json:"q3,omitempty"`
	IQR            *float64         `json:"iqr,omitempty"`
	Skewness       *float64         `json:"skewness,omitempty"`
	Kurtosis       *float64         `json:"kurtosis,omitempty"`
	Categories     map[string]int   `json:"categories,omitempty"` // For categorical columns
}

// DistributionInfo contains information about data distribution
type DistributionInfo struct {
	Histogram      []HistogramBin `json:"histogram,omitempty"`
	IsNormal       bool           `json:"isNormal"`
	NormalityPValue float64        `json:"normalityPValue,omitempty"`
	DistType       string         `json:"distType"` // "normal", "skewed", "bimodal", "uniform", etc.
}

// HistogramBin represents a bin in a histogram
type HistogramBin struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int     `json:"count"`
}

// OutlierInfo contains information about detected outliers
type OutlierInfo struct {
	RowIndex int     `json:"rowIndex"`
	Value    string  `json:"value"`
	Method   string  `json:"method"` // "iqr", "zscore"
	Score    float64 `json:"score"`  // Z-score or IQR multiplier
}

// QualityIssue represents a data quality problem
type QualityIssue struct {
	Severity    string   `json:"severity"` // "error", "warning", "info"
	Category    string   `json:"category"` // "missing", "outlier", "duplicate", "type", "correlation"
	Description string   `json:"description"`
	Affected    []string `json:"affected"` // Column names or row indices
	Impact      string   `json:"impact"` // Impact on PCA analysis
}

// Recommendation represents a suggestion for improving data quality
type Recommendation struct {
	Priority    string   `json:"priority"` // "high", "medium", "low"
	Category    string   `json:"category"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Columns     []string `json:"columns,omitempty"`
}

// AnalyzeDataQuality performs comprehensive data quality analysis
func (a *App) AnalyzeDataQuality(data *FileData) (*DataQualityReport, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to analyze")
	}

	// Initialize report
	report := &DataQualityReport{
		DataProfile: DataProfile{
			Rows:    data.Rows,
			Columns: data.Columns,
		},
		ColumnAnalysis:  make([]ColumnAnalysis, 0, data.Columns),
		Issues:          []QualityIssue{},
		Recommendations: []Recommendation{},
	}

	// Count column types
	for _, colType := range data.ColumnTypes {
		switch colType {
		case "numeric":
			report.DataProfile.NumericColumns++
		case "categorical":
			report.DataProfile.CategoricalColumns++
		case "target":
			report.DataProfile.TargetColumns++
		}
	}

	// Calculate missing data percentage
	missingStats := a.AnalyzeMissingValues(data)
	report.DataProfile.MissingPercent = missingStats.MissingPercent

	// Detect duplicate rows
	report.DataProfile.DuplicateRows = countDuplicateRows(data)

	// Estimate memory size
	report.DataProfile.MemorySize = estimateMemorySize(data)

	// Analyze each column
	for colIdx, header := range data.Headers {
		colAnalysis := analyzeColumn(data, colIdx, header)
		report.ColumnAnalysis = append(report.ColumnAnalysis, colAnalysis)
	}

	// Calculate correlations for numeric columns
	correlations := calculateCorrelations(data)

	// Generate issues based on analysis
	report.Issues = generateQualityIssues(report, correlations)

	// Generate recommendations
	report.Recommendations = generateRecommendations(report)

	// Calculate overall quality score
	report.QualityScore = calculateQualityScore(report)

	return report, nil
}

// analyzeColumn performs detailed analysis on a single column
func analyzeColumn(data *FileData, colIdx int, header string) ColumnAnalysis {
	analysis := ColumnAnalysis{
		Name: header,
		Type: "numeric", // Default
	}

	// Get column type
	if data.ColumnTypes != nil {
		if colType, exists := data.ColumnTypes[header]; exists {
			analysis.Type = colType
		}
	}

	// Calculate statistics based on column type
	if analysis.Type == "numeric" {
		analysis.Stats = calculateNumericStats(data, colIdx)
		analysis.Distribution = analyzeDistribution(data, colIdx)
		analysis.Outliers = detectOutliers(data, colIdx, analysis.Stats)
	} else {
		analysis.Stats = calculateCategoricalStats(data, colIdx)
	}

	// Calculate column quality score
	analysis.QualityScore = calculateColumnQualityScore(analysis)

	return analysis
}

// calculateNumericStats calculates statistics for numeric columns
func calculateNumericStats(data *FileData, colIdx int) ColumnStatistics {
	stats := ColumnStatistics{
		Count: data.Rows,
	}

	// Use utility function to get numeric values
	values := getNumericValues(data.Data, colIdx)
	
	// Count missing values
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			if isMissingValue(data.Data[rowIdx][colIdx]) {
				stats.Missing++
			}
		}
	}

	if stats.Count > 0 {
		stats.MissingPercent = float64(stats.Missing) / float64(stats.Count) * 100
	}

	if len(values) == 0 {
		return stats
	}

	// Sort values for percentile calculations
	sort.Float64s(values)

	// Basic statistics
	stats.Unique = countUnique(values)
	mean := calculateMean(values)
	stats.Mean = &mean
	median := calculateMedian(values)
	stats.Median = &median
	stdDev := calculateStdDev(values, mean)
	stats.StdDev = &stdDev
	min := values[0]
	stats.Min = &min
	max := values[len(values)-1]
	stats.Max = &max

	// Quartiles
	q1 := calculatePercentile(values, 25)
	stats.Q1 = &q1
	q3 := calculatePercentile(values, 75)
	stats.Q3 = &q3
	iqr := q3 - q1
	stats.IQR = &iqr

	// Higher moments
	skewness := calculateSkewness(values, mean, stdDev)
	stats.Skewness = &skewness
	kurtosis := calculateKurtosis(values, mean, stdDev)
	stats.Kurtosis = &kurtosis

	return stats
}

// calculateCategoricalStats calculates statistics for categorical columns
func calculateCategoricalStats(data *FileData, colIdx int) ColumnStatistics {
	stats := ColumnStatistics{
		Count:      data.Rows,
		Categories: make(map[string]int),
	}

	// Count occurrences
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				stats.Missing++
			} else {
				stats.Categories[value]++
			}
		}
	}

	if stats.Count > 0 {
		stats.MissingPercent = float64(stats.Missing) / float64(stats.Count) * 100
	}

	stats.Unique = len(stats.Categories)

	// Find mode
	if len(stats.Categories) > 0 {
		maxCount := 0
		mode := ""
		for value, count := range stats.Categories {
			if count > maxCount {
				maxCount = count
				mode = value
			}
		}
		stats.Mode = &mode
	}

	return stats
}

// Helper functions for statistical calculations

func countUnique(values []float64) int {
	unique := make(map[float64]bool)
	for _, v := range values {
		unique[v] = true
	}
	return len(unique)
}

func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMedian(values []float64) float64 {
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2
	}
	return values[n/2]
}

func calculateStdDev(values []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)))
}

func calculatePercentile(values []float64, percentile float64) float64 {
	index := (percentile / 100) * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	weight := index - float64(lower)
	
	if lower == upper {
		return values[lower]
	}
	
	return values[lower]*(1-weight) + values[upper]*weight
}

func calculateSkewness(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		z := (v - mean) / stdDev
		sum += z * z * z
	}
	
	return (n / ((n - 1) * (n - 2))) * sum
}

func calculateKurtosis(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		z := (v - mean) / stdDev
		sum += z * z * z * z
	}
	
	return (n*(n+1)/((n-1)*(n-2)*(n-3)))*sum - 3*(n-1)*(n-1)/((n-2)*(n-3))
}

// analyzeDistribution analyzes the distribution of numeric data
func analyzeDistribution(data *FileData, colIdx int) DistributionInfo {
	dist := DistributionInfo{}
	
	// Collect valid numeric values
	values := []float64{}
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if !isMissingValue(value) {
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					values = append(values, num)
				}
			}
		}
	}
	
	if len(values) < 10 {
		return dist
	}
	
	// Create histogram with 10 bins
	sort.Float64s(values)
	min, max := values[0], values[len(values)-1]
	binWidth := (max - min) / 10
	
	if binWidth > 0 {
		dist.Histogram = make([]HistogramBin, 10)
		for i := 0; i < 10; i++ {
			binMin := min + float64(i)*binWidth
			binMax := binMin + binWidth
			if i == 9 {
				binMax = max + 0.001 // Include max value in last bin
			}
			
			dist.Histogram[i] = HistogramBin{
				Min: binMin,
				Max: binMax,
			}
		}
		
		// Count values in each bin
		for _, v := range values {
			binIndex := int((v - min) / binWidth)
			if binIndex >= 10 {
				binIndex = 9
			}
			dist.Histogram[binIndex].Count++
		}
	}
	
	// Simple normality test based on skewness and kurtosis
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)
	skewness := calculateSkewness(values, mean, stdDev)
	kurtosis := calculateKurtosis(values, mean, stdDev)
	
	// Very simple normality check
	dist.IsNormal = math.Abs(skewness) < 0.5 && math.Abs(kurtosis) < 1.0
	
	// Determine distribution type
	if dist.IsNormal {
		dist.DistType = "normal"
	} else if math.Abs(skewness) > 1.0 {
		if skewness > 0 {
			dist.DistType = "right-skewed"
		} else {
			dist.DistType = "left-skewed"
		}
	} else if len(dist.Histogram) > 0 {
		// Check for bimodal distribution
		peaks := 0
		for i := 1; i < len(dist.Histogram)-1; i++ {
			if dist.Histogram[i].Count > dist.Histogram[i-1].Count &&
			   dist.Histogram[i].Count > dist.Histogram[i+1].Count {
				peaks++
			}
		}
		if peaks >= 2 {
			dist.DistType = "bimodal"
		} else {
			dist.DistType = "unknown"
		}
	}
	
	return dist
}

// detectOutliers detects outliers using IQR and Z-score methods
func detectOutliers(data *FileData, colIdx int, stats ColumnStatistics) []OutlierInfo {
	outliers := []OutlierInfo{}
	
	if stats.Q1 == nil || stats.Q3 == nil || stats.Mean == nil || stats.StdDev == nil {
		return outliers
	}
	
	// IQR method
	lowerBound := *stats.Q1 - 1.5**stats.IQR
	upperBound := *stats.Q3 + 1.5**stats.IQR
	
	// Z-score threshold
	zThreshold := 3.0
	
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if !isMissingValue(value) {
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					// Check IQR method
					if num < lowerBound || num > upperBound {
						outliers = append(outliers, OutlierInfo{
							RowIndex: rowIdx,
							Value:    value,
							Method:   "iqr",
							Score:    math.Abs(num-*stats.Median) / *stats.IQR,
						})
					}
					
					// Check Z-score method
					if *stats.StdDev > 0 {
						zScore := math.Abs(num-*stats.Mean) / *stats.StdDev
						if zScore > zThreshold {
							// Only add if not already detected by IQR
							alreadyDetected := false
							for _, o := range outliers {
								if o.RowIndex == rowIdx {
									alreadyDetected = true
									break
								}
							}
							if !alreadyDetected {
								outliers = append(outliers, OutlierInfo{
									RowIndex: rowIdx,
									Value:    value,
									Method:   "zscore",
									Score:    zScore,
								})
							}
						}
					}
				}
			}
		}
	}
	
	return outliers
}

// countDuplicateRows counts the number of duplicate rows in the dataset
func countDuplicateRows(data *FileData) int {
	rowMap := make(map[string]int)
	duplicates := 0
	
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		// Create a key from the row data
		rowKey := strings.Join(data.Data[rowIdx], "|")
		rowMap[rowKey]++
		if rowMap[rowKey] == 2 {
			duplicates++ // Count the first duplicate
		} else if rowMap[rowKey] > 2 {
			duplicates++ // Count subsequent duplicates
		}
	}
	
	return duplicates
}

// estimateMemorySize estimates the memory usage of the dataset
func estimateMemorySize(data *FileData) string {
	// Rough estimation: average 10 bytes per cell
	bytes := data.Rows * data.Columns * 10
	
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

// calculateCorrelations calculates correlations between numeric columns
func calculateCorrelations(data *FileData) map[string]map[string]float64 {
	correlations := make(map[string]map[string]float64)
	
	// Get numeric columns
	numericCols := []int{}
	numericHeaders := []string{}
	for i, header := range data.Headers {
		if colType, exists := data.ColumnTypes[header]; exists && colType == "numeric" {
			numericCols = append(numericCols, i)
			numericHeaders = append(numericHeaders, header)
		}
	}
	
	// Calculate pairwise correlations
	for i, col1 := range numericCols {
		header1 := numericHeaders[i]
		if _, exists := correlations[header1]; !exists {
			correlations[header1] = make(map[string]float64)
		}
		
		for j, col2 := range numericCols {
			header2 := numericHeaders[j]
			
			if i == j {
				correlations[header1][header2] = 1.0
			} else {
				corr := calculatePearsonCorrelation(data, col1, col2)
				correlations[header1][header2] = corr
			}
		}
	}
	
	return correlations
}

// calculatePearsonCorrelation calculates Pearson correlation between two columns
func calculatePearsonCorrelation(data *FileData, col1, col2 int) float64 {
	// Collect paired values
	pairs := [][2]float64{}
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if col1 < len(data.Data[rowIdx]) && col2 < len(data.Data[rowIdx]) {
			val1 := strings.TrimSpace(data.Data[rowIdx][col1])
			val2 := strings.TrimSpace(data.Data[rowIdx][col2])
			
			if !isMissingValue(val1) && !isMissingValue(val2) {
				if num1, err1 := strconv.ParseFloat(val1, 64); err1 == nil {
					if num2, err2 := strconv.ParseFloat(val2, 64); err2 == nil {
						pairs = append(pairs, [2]float64{num1, num2})
					}
				}
			}
		}
	}
	
	if len(pairs) < 2 {
		return 0
	}
	
	// Calculate means
	mean1, mean2 := 0.0, 0.0
	for _, pair := range pairs {
		mean1 += pair[0]
		mean2 += pair[1]
	}
	mean1 /= float64(len(pairs))
	mean2 /= float64(len(pairs))
	
	// Calculate correlation
	num, den1, den2 := 0.0, 0.0, 0.0
	for _, pair := range pairs {
		diff1 := pair[0] - mean1
		diff2 := pair[1] - mean2
		num += diff1 * diff2
		den1 += diff1 * diff1
		den2 += diff2 * diff2
	}
	
	if den1 == 0 || den2 == 0 {
		return 0
	}
	
	return num / math.Sqrt(den1*den2)
}

// generateQualityIssues generates quality issues based on the analysis
func generateQualityIssues(report *DataQualityReport, correlations map[string]map[string]float64) []QualityIssue {
	issues := []QualityIssue{}
	
	// Check for high missing data
	if report.DataProfile.MissingPercent > 20 {
		issues = append(issues, QualityIssue{
			Severity:    "error",
			Category:    "missing",
			Description: fmt.Sprintf("Dataset has %.1f%% missing values", report.DataProfile.MissingPercent),
			Impact:      "High missing data can significantly affect PCA results",
		})
	} else if report.DataProfile.MissingPercent > 10 {
		issues = append(issues, QualityIssue{
			Severity:    "warning",
			Category:    "missing",
			Description: fmt.Sprintf("Dataset has %.1f%% missing values", report.DataProfile.MissingPercent),
			Impact:      "Missing data may affect PCA results",
		})
	}
	
	// Check for columns with high missing data
	for _, col := range report.ColumnAnalysis {
		if col.Stats.MissingPercent > 50 {
			issues = append(issues, QualityIssue{
				Severity:    "error",
				Category:    "missing",
				Description: fmt.Sprintf("Column '%s' has %.1f%% missing values", col.Name, col.Stats.MissingPercent),
				Affected:    []string{col.Name},
				Impact:      "Columns with >50% missing data should be removed",
			})
		}
	}
	
	// Check for duplicate rows
	if report.DataProfile.DuplicateRows > 0 {
		issues = append(issues, QualityIssue{
			Severity:    "warning",
			Category:    "duplicate",
			Description: fmt.Sprintf("Found %d duplicate rows", report.DataProfile.DuplicateRows),
			Impact:      "Duplicate rows can bias PCA results",
		})
	}
	
	// Check for outliers
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && len(col.Outliers) > 0 {
			outlierPercent := float64(len(col.Outliers)) / float64(col.Stats.Count) * 100
			if outlierPercent > 10 {
				issues = append(issues, QualityIssue{
					Severity:    "warning",
					Category:    "outlier",
					Description: fmt.Sprintf("Column '%s' has %d outliers (%.1f%%)", col.Name, len(col.Outliers), outlierPercent),
					Affected:    []string{col.Name},
					Impact:      "Outliers can disproportionately influence PCA components",
				})
			}
		}
	}
	
	// Check for highly correlated variables
	for col1, corrMap := range correlations {
		for col2, corr := range corrMap {
			if col1 < col2 && math.Abs(corr) > 0.95 {
				issues = append(issues, QualityIssue{
					Severity:    "warning",
					Category:    "correlation",
					Description: fmt.Sprintf("Columns '%s' and '%s' are highly correlated (r=%.3f)", col1, col2, corr),
					Affected:    []string{col1, col2},
					Impact:      "Highly correlated variables provide redundant information in PCA",
				})
			}
		}
	}
	
	// Check for low variance columns
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.StdDev != nil && *col.Stats.StdDev < 0.01 {
			issues = append(issues, QualityIssue{
				Severity:    "info",
				Category:    "variance",
				Description: fmt.Sprintf("Column '%s' has very low variance (σ=%.4f)", col.Name, *col.Stats.StdDev),
				Affected:    []string{col.Name},
				Impact:      "Low variance columns contribute little to PCA",
			})
		}
	}
	
	// Check for non-normal distributions
	nonNormalCount := 0
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && !col.Distribution.IsNormal {
			nonNormalCount++
		}
	}
	if nonNormalCount > 0 {
		issues = append(issues, QualityIssue{
			Severity:    "info",
			Category:    "distribution",
			Description: fmt.Sprintf("%d numeric columns have non-normal distributions", nonNormalCount),
			Impact:      "PCA assumes normality; consider data transformations",
		})
	}
	
	return issues
}

// generateRecommendations generates recommendations based on the quality analysis
func generateRecommendations(report *DataQualityReport) []Recommendation {
	recs := []Recommendation{}
	
	// Missing data recommendations
	if report.DataProfile.MissingPercent > 10 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "missing",
			Action:      "Handle missing values",
			Description: "Use appropriate fill strategies (mean/median for numeric, mode for categorical) or remove rows/columns with excessive missing data",
		})
	}
	
	// Duplicate rows recommendation
	if report.DataProfile.DuplicateRows > 0 {
		recs = append(recs, Recommendation{
			Priority:    "medium",
			Category:    "duplicate",
			Action:      "Remove duplicate rows",
			Description: fmt.Sprintf("Remove %d duplicate rows to avoid biasing the analysis", report.DataProfile.DuplicateRows),
		})
	}
	
	// Outlier recommendations
	colsWithOutliers := []string{}
	for _, col := range report.ColumnAnalysis {
		if len(col.Outliers) > 5 {
			colsWithOutliers = append(colsWithOutliers, col.Name)
		}
	}
	if len(colsWithOutliers) > 0 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "outlier",
			Action:      "Handle outliers",
			Description: "Consider removing or transforming outliers, or use robust scaling",
			Columns:     colsWithOutliers,
		})
	}
	
	// Scaling recommendation
	varyingScales := false
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.Min != nil && col.Stats.Max != nil {
			range_ := *col.Stats.Max - *col.Stats.Min
			if range_ > 1000 || range_ < 0.01 {
				varyingScales = true
				break
			}
		}
	}
	if varyingScales {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "scaling",
			Action:      "Scale numeric columns",
			Description: "Columns have varying scales; consider standardization or normalization before PCA",
		})
	}
	
	// Distribution recommendations
	skewedCols := []string{}
	for _, col := range report.ColumnAnalysis {
		if col.Type == "numeric" && col.Stats.Skewness != nil && math.Abs(*col.Stats.Skewness) > 1.0 {
			skewedCols = append(skewedCols, col.Name)
		}
	}
	if len(skewedCols) > 0 {
		recs = append(recs, Recommendation{
			Priority:    "medium",
			Category:    "distribution",
			Action:      "Transform skewed distributions",
			Description: "Consider log or square root transformations for highly skewed columns",
			Columns:     skewedCols,
		})
	}
	
	// Correlation recommendations
	if report.DataProfile.NumericColumns < 3 {
		recs = append(recs, Recommendation{
			Priority:    "high",
			Category:    "columns",
			Action:      "Add more numeric columns",
			Description: fmt.Sprintf("Only %d numeric columns available; PCA requires multiple numeric features", report.DataProfile.NumericColumns),
		})
	}
	
	return recs
}

// calculateQualityScore calculates an overall quality score for the dataset
func calculateQualityScore(report *DataQualityReport) float64 {
	score := 100.0
	
	// Deduct for missing data
	score -= report.DataProfile.MissingPercent * 0.5
	
	// Deduct for duplicate rows
	if report.DataProfile.Rows > 0 {
		duplicatePercent := float64(report.DataProfile.DuplicateRows) / float64(report.DataProfile.Rows) * 100
		score -= duplicatePercent * 0.3
	}
	
	// Deduct for columns with excessive missing data
	for _, col := range report.ColumnAnalysis {
		if col.Stats.MissingPercent > 50 {
			score -= 5.0
		}
	}
	
	// Deduct for outliers
	totalOutliers := 0
	for _, col := range report.ColumnAnalysis {
		totalOutliers += len(col.Outliers)
	}
	if report.DataProfile.Rows > 0 {
		outlierPercent := float64(totalOutliers) / float64(report.DataProfile.Rows*report.DataProfile.NumericColumns) * 100
		score -= outlierPercent * 0.2
	}
	
	// Deduct for insufficient numeric columns
	if report.DataProfile.NumericColumns < 3 {
		score -= 20.0
	}
	
	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// calculateColumnQualityScore calculates quality score for a single column
func calculateColumnQualityScore(analysis ColumnAnalysis) float64 {
	score := 100.0
	
	// Deduct for missing data
	score -= analysis.Stats.MissingPercent * 0.5
	
	// Deduct for outliers
	if analysis.Stats.Count > 0 {
		outlierPercent := float64(len(analysis.Outliers)) / float64(analysis.Stats.Count) * 100
		score -= outlierPercent * 0.3
	}
	
	// Deduct for low variance (numeric columns)
	if analysis.Type == "numeric" && analysis.Stats.StdDev != nil {
		if *analysis.Stats.StdDev < 0.01 {
			score -= 10.0
		}
	}
	
	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// GoPCAStatus represents the installation status of GoPCA
type GoPCAStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	Error     string `json:"error,omitempty"`
}

// CheckGoPCAStatus checks if GoPCA Desktop is installed and available
func (a *App) CheckGoPCAStatus() *GoPCAStatus {
	status := &GoPCAStatus{
		Installed: false,
	}

	// Check for gopca-desktop in PATH
	path, err := exec.LookPath("gopca-desktop")
	if err == nil {
		status.Installed = true
		status.Path = path
		
		// Try to get version
		cmd := exec.Command(path, "--version")
		output, err := cmd.Output()
		if err == nil {
			status.Version = strings.TrimSpace(string(output))
		}
		return status
	}

	// Check common installation locations based on OS
	var possiblePaths []string
	switch runtime.GOOS {
	case "darwin":
		possiblePaths = []string{
			"/Applications/GoPCA Desktop.app/Contents/MacOS/gopca-desktop",
			filepath.Join(os.Getenv("HOME"), "Applications/GoPCA Desktop.app/Contents/MacOS/gopca-desktop"),
			"/usr/local/bin/gopca-desktop",
			filepath.Join(os.Getenv("HOME"), "go/bin/gopca-desktop"),
		}
	case "windows":
		possiblePaths = []string{
			"C:\\Program Files\\GoPCA Desktop\\gopca-desktop.exe",
			filepath.Join(os.Getenv("APPDATA"), "GoPCA Desktop\\gopca-desktop.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "GoPCA Desktop\\gopca-desktop.exe"),
			filepath.Join(os.Getenv("USERPROFILE"), "go\\bin\\gopca-desktop.exe"),
		}
	default: // Linux and others
		possiblePaths = []string{
			"/usr/local/bin/gopca-desktop",
			"/usr/bin/gopca-desktop",
			filepath.Join(os.Getenv("HOME"), ".local/bin/gopca-desktop"),
			filepath.Join(os.Getenv("HOME"), "go/bin/gopca-desktop"),
		}
	}

	// Check each possible path
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			status.Installed = true
			status.Path = p
			
			// Try to get version
			cmd := exec.Command(p, "--version")
			output, err := cmd.Output()
			if err == nil {
				status.Version = strings.TrimSpace(string(output))
			}
			return status
		}
	}

	status.Error = "GoPCA Desktop not found. Please ensure it is installed and in your PATH."
	return status
}

// OpenInGoPCA saves the current data to a temporary file and opens it in GoPCA Desktop
func (a *App) OpenInGoPCA(data *FileData) error {
	if data == nil || len(data.Data) == 0 {
		return fmt.Errorf("no data to export")
	}

	// Check if GoPCA is installed
	status := a.CheckGoPCAStatus()
	if !status.Installed {
		return fmt.Errorf("GoPCA Desktop not found: %s", status.Error)
	}

	// Create a temporary file
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("gocsv_export_%s.csv", timestamp))

	// Write data to temp file
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers with row name column if present
	headers := data.Headers
	if len(data.RowNames) > 0 {
		headers = append([]string{"Row"}, headers...)
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write data rows
	for i, row := range data.Data {
		rowData := row
		if len(data.RowNames) > 0 && i < len(data.RowNames) {
			rowData = append([]string{data.RowNames[i]}, row...)
		}
		if err := writer.Write(rowData); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+1, err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	// Launch GoPCA with the file
	var cmd *exec.Cmd
	
	// Use different launch methods based on OS
	switch runtime.GOOS {
	case "darwin":
		if strings.Contains(status.Path, ".app") {
			// Launch macOS app bundle with open command
			cmd = exec.Command("open", "-a", filepath.Dir(filepath.Dir(status.Path)), "--args", "--open", tempFile)
		} else {
			// Direct binary execution
			cmd = exec.Command(status.Path, "--open", tempFile)
		}
	default:
		// Windows and Linux
		cmd = exec.Command(status.Path, "--open", tempFile)
	}

	// Start the process without waiting
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch GoPCA Desktop: %w", err)
	}

	// Log the temporary file location
	wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Exported data to: %s", tempFile))
	wailsruntime.LogInfo(a.ctx, fmt.Sprintf("Launched GoPCA Desktop: %s", status.Path))

	// Schedule cleanup of temp file after a delay
	go func() {
		time.Sleep(10 * time.Second) // Give GoPCA time to load the file
		os.Remove(tempFile)
	}()

	return nil
}

// DownloadGoPCA opens the GoPCA download page in the default browser
func (a *App) DownloadGoPCA() error {
	url := "https://github.com/bitjungle/gopca/releases"
	wailsruntime.BrowserOpenURL(a.ctx, url)
	return nil
}

// UndoRedoState represents the current state of undo/redo
type UndoRedoState struct {
	CanUndo     bool     `json:"canUndo"`
	CanRedo     bool     `json:"canRedo"`
	History     []string `json:"history"`
	CurrentPos  int      `json:"currentPos"`
}

// GetUndoRedoState returns the current undo/redo state
func (a *App) GetUndoRedoState() *UndoRedoState {
	history, current := a.history.GetHistory()
	return &UndoRedoState{
		CanUndo:    a.history.CanUndo(),
		CanRedo:    a.history.CanRedo(),
		History:    history,
		CurrentPos: current,
	}
}

// Undo performs an undo operation
func (a *App) Undo() error {
	if err := a.history.Undo(); err != nil {
		return err
	}
	// Emit event to update UI
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	return nil
}

// Redo performs a redo operation
func (a *App) Redo() error {
	if err := a.history.Redo(); err != nil {
		return err
	}
	// Emit event to update UI
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	return nil
}

// ClearHistory clears the command history
func (a *App) ClearHistory() {
	a.history.Clear()
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
}

// ExecuteCellEdit executes a cell edit command
func (a *App) ExecuteCellEdit(data *FileData, row, col int, oldValue, newValue string) error {
	cmd := NewCellEditCommand(a, data, row, col, oldValue, newValue)
	if err := a.history.Execute(cmd); err != nil {
		return err
	}
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	return nil
}

// ExecuteHeaderEdit executes a header edit command
func (a *App) ExecuteHeaderEdit(data *FileData, col int, oldValue, newValue string) error {
	cmd := NewHeaderEditCommand(a, data, col, oldValue, newValue)
	if err := a.history.Execute(cmd); err != nil {
		return err
	}
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	return nil
}

// ExecuteFillMissingValues executes a fill missing values command
func (a *App) ExecuteFillMissingValues(data *FileData, strategy, column, customValue string) (*FileData, error) {
	cmd := NewFillMissingValuesCommand(a, data, strategy, column, customValue)
	if err := a.history.Execute(cmd); err != nil {
		return nil, err
	}
	wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	return data, nil
}

// ImportFileInfo represents information about a file to be imported
type ImportFileInfo struct {
	FileName    string   `json:"fileName"`
	FilePath    string   `json:"filePath"`
	FileSize    int64    `json:"fileSize"`
	FileFormat  string   `json:"fileFormat"` // "csv", "tsv", "excel", "json"
	Encoding    string   `json:"encoding"`
	Sheets      []string `json:"sheets,omitempty"` // For Excel files
	Error       string   `json:"error,omitempty"`
}

// ImportOptions represents options for importing a file
type ImportOptions struct {
	Format         string `json:"format"`
	Delimiter      string `json:"delimiter,omitempty"`      // For CSV/TSV
	HasHeaders     bool   `json:"hasHeaders"`
	HeaderRow      int    `json:"headerRow"`                // 0-based
	Sheet          string `json:"sheet,omitempty"`          // For Excel
	Range          string `json:"range,omitempty"`          // For Excel (e.g., "A1:Z100")
	RowNameColumn  int    `json:"rowNameColumn"`            // -1 if none, 0-based
	SkipRows       int    `json:"skipRows"`                 // Number of rows to skip from top
	MaxRows        int    `json:"maxRows"`                  // 0 for all rows
	SelectedColumns []int  `json:"selectedColumns,omitempty"` // Indices of columns to import
}

// FilePreview represents a preview of file contents
type FilePreview struct {
	Headers     []string          `json:"headers"`
	Data        [][]string        `json:"data"`        // First N rows
	ColumnTypes []string          `json:"columnTypes"` // Detected types
	Delimiter   string            `json:"delimiter"`   // Detected delimiter
	TotalRows   int               `json:"totalRows"`
	TotalCols   int               `json:"totalCols"`
	Issues      []string          `json:"issues,omitempty"`
}

// GetFileInfo gets information about a file for the import wizard
func (a *App) GetFileInfo(filePath string) (*ImportFileInfo, error) {
	info := &ImportFileInfo{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}
	
	// Get file size
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	info.FileSize = stat.Size()
	
	// Detect file format
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		info.FileFormat = "csv"
		info.Encoding = "UTF-8" // TODO: Detect encoding
	case ".tsv":
		info.FileFormat = "tsv"
		info.Encoding = "UTF-8"
	case ".xlsx", ".xls":
		info.FileFormat = "excel"
		// Get sheet names
		sheets, err := a.getExcelSheets(filePath)
		if err != nil {
			info.Error = fmt.Sprintf("Failed to read Excel sheets: %v", err)
		} else {
			info.Sheets = sheets
		}
	case ".json":
		info.FileFormat = "json"
		info.Encoding = "UTF-8"
	default:
		// Try to detect format by content
		info.FileFormat = a.detectFileFormat(filePath)
		info.Encoding = "UTF-8"
	}
	
	return info, nil
}

// PreviewFile generates a preview of the file with the given options
func (a *App) PreviewFile(filePath string, options ImportOptions) (*FilePreview, error) {
	preview := &FilePreview{
		Issues: []string{},
	}
	
	switch options.Format {
	case "csv", "tsv":
		return a.previewCSV(filePath, options, preview)
	case "excel":
		return a.previewExcel(filePath, options, preview)
	case "json":
		return a.previewJSON(filePath, options, preview)
	default:
		return nil, fmt.Errorf("unsupported format: %s", options.Format)
	}
}

// ImportFile imports a file with the given options
func (a *App) ImportFile(filePath string, options ImportOptions) (*FileData, error) {
	switch options.Format {
	case "csv", "tsv":
		return a.importCSVWithOptions(filePath, options)
	case "excel":
		return a.importExcelWithOptions(filePath, options)
	case "json":
		return a.importJSONWithOptions(filePath, options)
	default:
		return nil, fmt.Errorf("unsupported format: %s", options.Format)
	}
}

// getExcelSheets returns the sheet names in an Excel file
func (a *App) getExcelSheets(filePath string) ([]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	return f.GetSheetList(), nil
}

// detectFileFormat tries to detect the file format by content
func (a *App) detectFileFormat(filePath string) string {
	// Read first few bytes
	file, err := os.Open(filePath)
	if err != nil {
		return "unknown"
	}
	defer file.Close()
	
	// Read first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	buf = buf[:n]
	
	// Check for Excel magic bytes
	if len(buf) >= 8 {
		if buf[0] == 0xD0 && buf[1] == 0xCF && buf[2] == 0x11 && buf[3] == 0xE0 {
			return "excel" // Old Excel format
		}
		if buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04 {
			return "excel" // New Excel format (ZIP)
		}
	}
	
	// Check for JSON
	content := string(buf)
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return "json"
	}
	
	// Check for TSV (more tabs than commas)
	tabCount := strings.Count(content, "\t")
	commaCount := strings.Count(content, ",")
	if tabCount > commaCount*2 {
		return "tsv"
	}
	
	// Default to CSV
	return "csv"
}

// previewCSV generates a preview of a CSV/TSV file
func (a *App) previewCSV(filePath string, options ImportOptions, preview *FilePreview) (*FilePreview, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Set delimiter
	if options.Format == "tsv" || options.Delimiter == "\t" {
		reader.Comma = '\t'
	} else if options.Delimiter != "" && len(options.Delimiter) == 1 {
		reader.Comma = rune(options.Delimiter[0])
	}
	
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	// Skip rows if specified
	for i := 0; i < options.SkipRows; i++ {
		_, err := reader.Read()
		if err != nil {
			preview.Issues = append(preview.Issues, fmt.Sprintf("Failed to skip row %d: %v", i+1, err))
		}
	}
	
	// Read all data for analysis
	allData, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(allData) == 0 {
		return nil, fmt.Errorf("no data found in file")
	}
	
	preview.TotalRows = len(allData)
	preview.TotalCols = len(allData[0])
	
	// Extract headers
	if options.HasHeaders && options.HeaderRow < len(allData) {
		preview.Headers = allData[options.HeaderRow]
		// Remove header row from data
		allData = append(allData[:options.HeaderRow], allData[options.HeaderRow+1:]...)
	} else {
		// Generate default headers
		preview.Headers = make([]string, preview.TotalCols)
		for i := 0; i < preview.TotalCols; i++ {
			preview.Headers[i] = fmt.Sprintf("Column_%d", i+1)
		}
	}
	
	// Get preview data (first 100 rows or less)
	previewRows := 100
	if options.MaxRows > 0 && options.MaxRows < previewRows {
		previewRows = options.MaxRows
	}
	if len(allData) < previewRows {
		previewRows = len(allData)
	}
	
	preview.Data = allData[:previewRows]
	
	// Detect column types
	preview.ColumnTypes = make([]string, preview.TotalCols)
	for i := 0; i < preview.TotalCols; i++ {
		preview.ColumnTypes[i] = a.detectColumnType(allData, i)
	}
	
	// Detect delimiter if not specified
	if options.Delimiter == "" {
		if options.Format == "tsv" {
			preview.Delimiter = "\\t"
		} else {
			preview.Delimiter = ","
		}
	} else {
		preview.Delimiter = options.Delimiter
	}
	
	return preview, nil
}

// previewExcel generates a preview of an Excel file
func (a *App) previewExcel(filePath string, options ImportOptions, preview *FilePreview) (*FilePreview, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	// Use specified sheet or first sheet
	sheet := options.Sheet
	if sheet == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in Excel file")
		}
		sheet = sheets[0]
	}
	
	// Get all rows
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %w", sheet, err)
	}
	
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in sheet %s", sheet)
	}
	
	// Apply range if specified
	if options.Range != "" {
		// TODO: Implement range parsing
		preview.Issues = append(preview.Issues, "Range selection not yet implemented")
	}
	
	// Skip rows if specified
	if options.SkipRows > 0 && options.SkipRows < len(rows) {
		rows = rows[options.SkipRows:]
	}
	
	preview.TotalRows = len(rows)
	if len(rows) > 0 {
		preview.TotalCols = len(rows[0])
	}
	
	// Extract headers
	if options.HasHeaders && options.HeaderRow < len(rows) {
		preview.Headers = rows[options.HeaderRow]
		// Remove header row from data
		rows = append(rows[:options.HeaderRow], rows[options.HeaderRow+1:]...)
	} else {
		// Generate default headers
		preview.Headers = make([]string, preview.TotalCols)
		for i := 0; i < preview.TotalCols; i++ {
			preview.Headers[i] = fmt.Sprintf("Column_%d", i+1)
		}
	}
	
	// Get preview data
	previewRows := 100
	if options.MaxRows > 0 && options.MaxRows < previewRows {
		previewRows = options.MaxRows
	}
	if len(rows) < previewRows {
		previewRows = len(rows)
	}
	
	preview.Data = rows[:previewRows]
	
	// Detect column types
	preview.ColumnTypes = make([]string, preview.TotalCols)
	for i := 0; i < preview.TotalCols; i++ {
		preview.ColumnTypes[i] = a.detectColumnType(rows, i)
	}
	
	return preview, nil
}

// previewJSON generates a preview of a JSON file
func (a *App) previewJSON(filePath string, options ImportOptions, preview *FilePreview) (*FilePreview, error) {
	// TODO: Implement JSON preview
	return nil, fmt.Errorf("JSON import not yet implemented")
}

// importCSVWithOptions imports a CSV file with specific options
func (a *App) importCSVWithOptions(filePath string, options ImportOptions) (*FileData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Set delimiter
	if options.Format == "tsv" || options.Delimiter == "\t" {
		reader.Comma = '\t'
	} else if options.Delimiter != "" && len(options.Delimiter) == 1 {
		reader.Comma = rune(options.Delimiter[0])
	}
	
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	// Skip rows if specified
	for i := 0; i < options.SkipRows; i++ {
		_, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to skip row %d: %w", i+1, err)
		}
	}
	
	// Read all data
	allData, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(allData) == 0 {
		return nil, fmt.Errorf("no data found in file")
	}
	
	fileData := &FileData{
		CategoricalColumns:   make(map[string][]string),
		NumericTargetColumns: make(map[string][]types.JSONFloat64),
		ColumnTypes:          make(map[string]string),
	}
	
	// Extract headers
	if options.HasHeaders && options.HeaderRow < len(allData) {
		fileData.Headers = allData[options.HeaderRow]
		// Remove header row from data
		allData = append(allData[:options.HeaderRow], allData[options.HeaderRow+1:]...)
	} else {
		// Generate default headers
		fileData.Headers = make([]string, len(allData[0]))
		for i := 0; i < len(allData[0]); i++ {
			fileData.Headers[i] = fmt.Sprintf("Column_%d", i+1)
		}
	}
	
	// Extract row names if specified
	if options.RowNameColumn >= 0 && options.RowNameColumn < len(allData[0]) {
		fileData.RowNames = make([]string, len(allData))
		for i, row := range allData {
			if options.RowNameColumn < len(row) {
				fileData.RowNames[i] = row[options.RowNameColumn]
			}
		}
		
		// Remove row name column from headers and data
		fileData.Headers = append(fileData.Headers[:options.RowNameColumn], fileData.Headers[options.RowNameColumn+1:]...)
		for i := range allData {
			if options.RowNameColumn < len(allData[i]) {
				allData[i] = append(allData[i][:options.RowNameColumn], allData[i][options.RowNameColumn+1:]...)
			}
		}
	}
	
	// Apply column selection if specified
	if len(options.SelectedColumns) > 0 {
		// Filter headers
		newHeaders := make([]string, len(options.SelectedColumns))
		for i, colIdx := range options.SelectedColumns {
			if colIdx < len(fileData.Headers) {
				newHeaders[i] = fileData.Headers[colIdx]
			}
		}
		fileData.Headers = newHeaders
		
		// Filter data
		newData := make([][]string, len(allData))
		for i, row := range allData {
			newRow := make([]string, len(options.SelectedColumns))
			for j, colIdx := range options.SelectedColumns {
				if colIdx < len(row) {
					newRow[j] = row[colIdx]
				}
			}
			newData[i] = newRow
		}
		allData = newData
	}
	
	// Apply max rows if specified
	if options.MaxRows > 0 && len(allData) > options.MaxRows {
		allData = allData[:options.MaxRows]
		if fileData.RowNames != nil && len(fileData.RowNames) > options.MaxRows {
			fileData.RowNames = fileData.RowNames[:options.MaxRows]
		}
	}
	
	fileData.Data = allData
	fileData.Rows = len(allData)
	fileData.Columns = len(fileData.Headers)
	
	// Detect column types and process data
	for i, header := range fileData.Headers {
		colType := a.detectColumnType(allData, i)
		fileData.ColumnTypes[header] = colType
		
		if strings.HasSuffix(header, "#target") {
			// Skip numeric target columns for now to avoid NaN JSON serialization issues
			// These columns are stored in the regular Data array and can be used for visualization
			continue
		} else if colType == "categorical" {
			// Categorical column
			values := make([]string, len(allData))
			for j, row := range allData {
				if i < len(row) {
					values[j] = row[i]
				}
			}
			fileData.CategoricalColumns[header] = values
		}
	}
	
	// Emit file loaded event
	wailsruntime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))
	
	// Clear command history for new file
	a.ClearHistory()
	
	return fileData, nil
}

// importExcelWithOptions imports an Excel file with specific options
func (a *App) importExcelWithOptions(filePath string, options ImportOptions) (*FileData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	// Use specified sheet or first sheet
	sheet := options.Sheet
	if sheet == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in Excel file")
		}
		sheet = sheets[0]
	}
	
	// Get all rows
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %w", sheet, err)
	}
	
	// Process similar to CSV
	// Skip rows if specified
	if options.SkipRows > 0 && options.SkipRows < len(rows) {
		rows = rows[options.SkipRows:]
	}
	
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in sheet %s", sheet)
	}
	
	fileData := &FileData{
		CategoricalColumns:   make(map[string][]string),
		NumericTargetColumns: make(map[string][]types.JSONFloat64),
		ColumnTypes:          make(map[string]string),
	}
	
	// Extract headers
	if options.HasHeaders && options.HeaderRow < len(rows) {
		fileData.Headers = rows[options.HeaderRow]
		// Remove header row from data
		rows = append(rows[:options.HeaderRow], rows[options.HeaderRow+1:]...)
	} else {
		// Generate default headers
		if len(rows) > 0 {
			fileData.Headers = make([]string, len(rows[0]))
			for i := 0; i < len(rows[0]); i++ {
				fileData.Headers[i] = fmt.Sprintf("Column_%d", i+1)
			}
		}
	}
	
	// Extract row names if specified
	if options.RowNameColumn >= 0 && len(rows) > 0 && options.RowNameColumn < len(rows[0]) {
		fileData.RowNames = make([]string, len(rows))
		for i, row := range rows {
			if options.RowNameColumn < len(row) {
				fileData.RowNames[i] = row[options.RowNameColumn]
			}
		}
		
		// Remove row name column
		fileData.Headers = append(fileData.Headers[:options.RowNameColumn], fileData.Headers[options.RowNameColumn+1:]...)
		for i := range rows {
			if options.RowNameColumn < len(rows[i]) {
				rows[i] = append(rows[i][:options.RowNameColumn], rows[i][options.RowNameColumn+1:]...)
			}
		}
	}
	
	// Apply column selection if specified
	if len(options.SelectedColumns) > 0 {
		// Filter headers
		newHeaders := make([]string, len(options.SelectedColumns))
		for i, colIdx := range options.SelectedColumns {
			if colIdx < len(fileData.Headers) {
				newHeaders[i] = fileData.Headers[colIdx]
			}
		}
		fileData.Headers = newHeaders
		
		// Filter data
		newRows := make([][]string, len(rows))
		for i, row := range rows {
			newRow := make([]string, len(options.SelectedColumns))
			for j, colIdx := range options.SelectedColumns {
				if colIdx < len(row) {
					newRow[j] = row[colIdx]
				}
			}
			newRows[i] = newRow
		}
		rows = newRows
	}
	
	// Apply max rows if specified
	if options.MaxRows > 0 && len(rows) > options.MaxRows {
		rows = rows[:options.MaxRows]
		if fileData.RowNames != nil && len(fileData.RowNames) > options.MaxRows {
			fileData.RowNames = fileData.RowNames[:options.MaxRows]
		}
	}
	
	fileData.Data = rows
	fileData.Rows = len(rows)
	fileData.Columns = len(fileData.Headers)
	
	// Detect column types
	for i, header := range fileData.Headers {
		colType := a.detectColumnType(rows, i)
		fileData.ColumnTypes[header] = colType
		
		if strings.HasSuffix(header, "#target") {
			// Skip numeric target columns for now to avoid NaN JSON serialization issues
			// These columns are stored in the regular Data array and can be used for visualization
			continue
		} else if colType == "categorical" {
			// Categorical column
			values := make([]string, len(rows))
			for j, row := range rows {
				if i < len(row) {
					values[j] = row[i]
				}
			}
			fileData.CategoricalColumns[header] = values
		}
	}
	
	// Emit file loaded event
	wailsruntime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))
	
	// Clear command history for new file
	a.ClearHistory()
	
	return fileData, nil
}

// importJSONWithOptions imports a JSON file with specific options
func (a *App) importJSONWithOptions(filePath string, options ImportOptions) (*FileData, error) {
	// TODO: Implement JSON import
	return nil, fmt.Errorf("JSON import not yet implemented")
}

// SelectFileForImport opens a file dialog and returns the selected file path
func (a *App) SelectFileForImport() (string, error) {
	dialogOptions := wailsruntime.OpenDialogOptions{
		Title: "Select file to import",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "All Supported Files",
				Pattern:     "*.csv;*.tsv;*.xlsx;*.xls;*.json",
			},
			{
				DisplayName: "CSV Files",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "TSV Files",
				Pattern:     "*.tsv",
			},
			{
				DisplayName: "Excel Files",
				Pattern:     "*.xlsx;*.xls",
			},
			{
				DisplayName: "JSON Files",
				Pattern:     "*.json",
			},
		},
	}
	
	filePath, err := wailsruntime.OpenFileDialog(a.ctx, dialogOptions)
	if err != nil {
		return "", err
	}
	
	if filePath == "" {
		return "", fmt.Errorf("no file selected")
	}
	
	return filePath, nil
}

// detectColumnType detects the type of a column based on its values
func (a *App) detectColumnType(data [][]string, colIndex int) string {
	if len(data) == 0 || colIndex < 0 {
		return "unknown"
	}
	
	// Count different types
	numericCount := 0
	totalCount := 0
	uniqueValues := make(map[string]bool)
	
	for _, row := range data {
		if colIndex >= len(row) {
			continue
		}
		
		value := strings.TrimSpace(row[colIndex])
		if value == "" {
			continue
		}
		
		totalCount++
		uniqueValues[value] = true
		
		// Try to parse as float
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			numericCount++
		}
	}
	
	if totalCount == 0 {
		return "empty"
	}
	
	// If more than 90% of non-empty values are numeric, consider it numeric
	if float64(numericCount)/float64(totalCount) > 0.9 {
		return "numeric"
	}
	
	// If unique values are less than 20% of total values or less than 20, consider it categorical
	if float64(len(uniqueValues))/float64(totalCount) < 0.2 || len(uniqueValues) < 20 {
		return "categorical"
	}
	
	return "text"
}

// TransformationType represents the type of transformation
type TransformationType string

const (
	TransformLog          TransformationType = "log"
	TransformSqrt         TransformationType = "sqrt"
	TransformSquare       TransformationType = "square"
	TransformStandardize  TransformationType = "standardize"
	TransformMinMax       TransformationType = "minmax"
	TransformBin          TransformationType = "bin"
	TransformOneHot       TransformationType = "onehot"
)

// TransformOptions represents options for data transformation
type TransformOptions struct {
	Type      TransformationType `json:"type"`
	Columns   []string          `json:"columns"`
	BinCount  int               `json:"binCount,omitempty"`  // For binning
	MinValue  float64           `json:"minValue,omitempty"`  // For min-max scaling
	MaxValue  float64           `json:"maxValue,omitempty"`  // For min-max scaling
}

// TransformationResult represents the result of a transformation
type TransformationResult struct {
	Success       bool               `json:"success"`
	TransformedColumns []string      `json:"transformedColumns"`
	NewColumns    []string          `json:"newColumns,omitempty"`
	Messages      []string          `json:"messages"`
	Data          *FileData         `json:"data"`
}

// ApplyTransformation applies a transformation to the data
func (a *App) ApplyTransformation(data *FileData, options TransformOptions) (*TransformationResult, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to transform")
	}
	
	result := &TransformationResult{
		Success:            true,
		TransformedColumns: []string{},
		Messages:          []string{},
	}
	
	// Create a copy of the data
	newData := &FileData{
		Headers:              make([]string, len(data.Headers)),
		Data:                 make([][]string, len(data.Data)),
		Rows:                 data.Rows,
		Columns:              data.Columns,
		CategoricalColumns:   make(map[string][]string),
		NumericTargetColumns: make(map[string][]types.JSONFloat64),
		ColumnTypes:          make(map[string]string),
	}
	
	// Copy headers
	copy(newData.Headers, data.Headers)
	
	// Copy data
	for i := range data.Data {
		newData.Data[i] = make([]string, len(data.Data[i]))
		copy(newData.Data[i], data.Data[i])
	}
	
	// Copy row names if present
	if data.RowNames != nil {
		newData.RowNames = make([]string, len(data.RowNames))
		copy(newData.RowNames, data.RowNames)
	}
	
	// Copy column types
	for k, v := range data.ColumnTypes {
		newData.ColumnTypes[k] = v
	}
	
	// Apply transformation based on type
	switch options.Type {
	case TransformLog, TransformSqrt, TransformSquare:
		err := a.applyMathTransformation(newData, options, result)
		if err != nil {
			return nil, err
		}
	case TransformStandardize:
		err := a.applyStandardization(newData, options, result)
		if err != nil {
			return nil, err
		}
	case TransformMinMax:
		err := a.applyMinMaxScaling(newData, options, result)
		if err != nil {
			return nil, err
		}
	case TransformBin:
		err := a.applyBinning(newData, options, result)
		if err != nil {
			return nil, err
		}
	case TransformOneHot:
		err := a.applyOneHotEncoding(newData, options, result)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported transformation type: %s", options.Type)
	}
	
	result.Data = newData
	return result, nil
}

// applyMathTransformation applies mathematical transformations (log, sqrt, square)
func (a *App) applyMathTransformation(data *FileData, options TransformOptions, result *TransformationResult) error {
	for _, colName := range options.Columns {
		// Find column index
		colIndex := -1
		for i, header := range data.Headers {
			if header == colName {
				colIndex = i
				break
			}
		}
		
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		
		// Check if column is numeric
		if data.ColumnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}
		
		// Apply transformation
		transformedCount := 0
		for i := range data.Data {
			if colIndex >= len(data.Data[i]) {
				continue
			}
			
			value := strings.TrimSpace(data.Data[i][colIndex])
			if value == "" {
				continue
			}
			
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			
			var transformed float64
			switch options.Type {
			case TransformLog:
				if num <= 0 {
					result.Messages = append(result.Messages, fmt.Sprintf("Warning: Non-positive value in row %d, column '%s' - cannot apply log", i+1, colName))
					continue
				}
				transformed = math.Log(num)
			case TransformSqrt:
				if num < 0 {
					result.Messages = append(result.Messages, fmt.Sprintf("Warning: Negative value in row %d, column '%s' - cannot apply sqrt", i+1, colName))
					continue
				}
				transformed = math.Sqrt(num)
			case TransformSquare:
				transformed = num * num
			}
			
			data.Data[i][colIndex] = fmt.Sprintf("%.6g", transformed)
			transformedCount++
		}
		
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Transformed %d values in column '%s'", transformedCount, colName))
	}
	
	return nil
}

// applyStandardization applies z-score standardization
func (a *App) applyStandardization(data *FileData, options TransformOptions, result *TransformationResult) error {
	for _, colName := range options.Columns {
		// Find column index
		colIndex := -1
		for i, header := range data.Headers {
			if header == colName {
				colIndex = i
				break
			}
		}
		
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		
		// Check if column is numeric
		if data.ColumnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}
		
		// Collect values
		values := []float64{}
		indices := []int{}
		for i := range data.Data {
			if colIndex >= len(data.Data[i]) {
				continue
			}
			
			value := strings.TrimSpace(data.Data[i][colIndex])
			if value == "" {
				continue
			}
			
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			
			values = append(values, num)
			indices = append(indices, i)
		}
		
		if len(values) < 2 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has insufficient numeric values for standardization", colName))
			continue
		}
		
		// Calculate mean and std dev
		mean := 0.0
		for _, v := range values {
			mean += v
		}
		mean /= float64(len(values))
		
		variance := 0.0
		for _, v := range values {
			variance += (v - mean) * (v - mean)
		}
		variance /= float64(len(values))
		stdDev := math.Sqrt(variance)
		
		if stdDev < 1e-10 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has zero variance, cannot standardize", colName))
			continue
		}
		
		// Apply standardization
		for i, idx := range indices {
			standardized := (values[i] - mean) / stdDev
			data.Data[idx][colIndex] = fmt.Sprintf("%.6g", standardized)
		}
		
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Standardized %d values in column '%s' (mean=%.3f, std=%.3f)", len(values), colName, mean, stdDev))
	}
	
	return nil
}

// applyMinMaxScaling applies min-max scaling
func (a *App) applyMinMaxScaling(data *FileData, options TransformOptions, result *TransformationResult) error {
	targetMin := options.MinValue
	targetMax := options.MaxValue
	if targetMax <= targetMin {
		targetMin = 0.0
		targetMax = 1.0
	}
	
	for _, colName := range options.Columns {
		// Find column index
		colIndex := -1
		for i, header := range data.Headers {
			if header == colName {
				colIndex = i
				break
			}
		}
		
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		
		// Check if column is numeric
		if data.ColumnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}
		
		// Collect values and find min/max
		values := []float64{}
		indices := []int{}
		minVal := math.Inf(1)
		maxVal := math.Inf(-1)
		
		for i := range data.Data {
			if colIndex >= len(data.Data[i]) {
				continue
			}
			
			value := strings.TrimSpace(data.Data[i][colIndex])
			if value == "" {
				continue
			}
			
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			
			values = append(values, num)
			indices = append(indices, i)
			
			if num < minVal {
				minVal = num
			}
			if num > maxVal {
				maxVal = num
			}
		}
		
		if len(values) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no numeric values", colName))
			continue
		}
		
		if maxVal <= minVal {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has constant values, cannot scale", colName))
			continue
		}
		
		// Apply min-max scaling
		for i, idx := range indices {
			scaled := (values[i] - minVal) / (maxVal - minVal) * (targetMax - targetMin) + targetMin
			data.Data[idx][colIndex] = fmt.Sprintf("%.6g", scaled)
		}
		
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Scaled %d values in column '%s' to range [%.2f, %.2f]", len(values), colName, targetMin, targetMax))
	}
	
	return nil
}

// applyBinning applies binning to numeric columns
func (a *App) applyBinning(data *FileData, options TransformOptions, result *TransformationResult) error {
	binCount := options.BinCount
	if binCount <= 0 {
		binCount = 5 // Default to 5 bins
	}
	
	for _, colName := range options.Columns {
		// Find column index
		colIndex := -1
		for i, header := range data.Headers {
			if header == colName {
				colIndex = i
				break
			}
		}
		
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		
		// Check if column is numeric
		if data.ColumnTypes[colName] != "numeric" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not numeric, skipping", colName))
			continue
		}
		
		// Collect values and find min/max
		values := []float64{}
		indices := []int{}
		minVal := math.Inf(1)
		maxVal := math.Inf(-1)
		
		for i := range data.Data {
			if colIndex >= len(data.Data[i]) {
				continue
			}
			
			value := strings.TrimSpace(data.Data[i][colIndex])
			if value == "" {
				continue
			}
			
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			
			values = append(values, num)
			indices = append(indices, i)
			
			if num < minVal {
				minVal = num
			}
			if num > maxVal {
				maxVal = num
			}
		}
		
		if len(values) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no numeric values", colName))
			continue
		}
		
		// Create bins
		binWidth := (maxVal - minVal) / float64(binCount)
		
		// Apply binning
		for i, idx := range indices {
			binIndex := int((values[i] - minVal) / binWidth)
			if binIndex >= binCount {
				binIndex = binCount - 1
			}
			
			binLabel := fmt.Sprintf("Bin_%d", binIndex+1)
			data.Data[idx][colIndex] = binLabel
		}
		
		// Update column type to categorical
		data.ColumnTypes[colName] = "categorical"
		
		// Update categorical columns
		catValues := make([]string, len(data.Data))
		for i := range data.Data {
			if colIndex < len(data.Data[i]) {
				catValues[i] = data.Data[i][colIndex]
			}
		}
		data.CategoricalColumns[colName] = catValues
		
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.Messages = append(result.Messages, fmt.Sprintf("Binned %d values in column '%s' into %d bins", len(values), colName, binCount))
	}
	
	return nil
}

// applyOneHotEncoding applies one-hot encoding to categorical columns
func (a *App) applyOneHotEncoding(data *FileData, options TransformOptions, result *TransformationResult) error {
	for _, colName := range options.Columns {
		// Find column index
		colIndex := -1
		for i, header := range data.Headers {
			if header == colName {
				colIndex = i
				break
			}
		}
		
		if colIndex == -1 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' not found", colName))
			continue
		}
		
		// Check if column is categorical
		if data.ColumnTypes[colName] != "categorical" {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' is not categorical, skipping", colName))
			continue
		}
		
		// Get unique values
		uniqueValues := make(map[string]bool)
		for i := range data.Data {
			if colIndex >= len(data.Data[i]) {
				continue
			}
			value := strings.TrimSpace(data.Data[i][colIndex])
			if value != "" {
				uniqueValues[value] = true
			}
		}
		
		if len(uniqueValues) == 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has no values", colName))
			continue
		}
		
		if len(uniqueValues) > 20 {
			result.Messages = append(result.Messages, fmt.Sprintf("Column '%s' has too many unique values (%d), skipping one-hot encoding", colName, len(uniqueValues)))
			continue
		}
		
		// Create sorted list of unique values
		sortedValues := make([]string, 0, len(uniqueValues))
		for val := range uniqueValues {
			sortedValues = append(sortedValues, val)
		}
		sort.Strings(sortedValues)
		
		// Add new columns for each unique value
		newColumns := []string{}
		for _, val := range sortedValues {
			newColName := fmt.Sprintf("%s_%s", colName, val)
			data.Headers = append(data.Headers, newColName)
			data.ColumnTypes[newColName] = "numeric"
			newColumns = append(newColumns, newColName)
			
			// Add the encoded values
			for i := range data.Data {
				if colIndex < len(data.Data[i]) && strings.TrimSpace(data.Data[i][colIndex]) == val {
					data.Data[i] = append(data.Data[i], "1")
				} else {
					data.Data[i] = append(data.Data[i], "0")
				}
			}
		}
		
		// Remove original column
		data.Headers = append(data.Headers[:colIndex], data.Headers[colIndex+1:]...)
		delete(data.ColumnTypes, colName)
		delete(data.CategoricalColumns, colName)
		
		for i := range data.Data {
			if colIndex < len(data.Data[i]) {
				data.Data[i] = append(data.Data[i][:colIndex], data.Data[i][colIndex+1:]...)
			}
		}
		
		data.Columns = len(data.Headers)
		
		result.TransformedColumns = append(result.TransformedColumns, colName)
		result.NewColumns = append(result.NewColumns, newColumns...)
		result.Messages = append(result.Messages, fmt.Sprintf("One-hot encoded column '%s' into %d new columns", colName, len(newColumns)))
		
		// Adjust indices for remaining columns
		for j := range options.Columns {
			if j > 0 && options.Columns[j] != colName {
				// Need to find and update column index if it was after the removed column
				for _, h := range data.Headers {
					if h == options.Columns[j] {
						// Update for next iteration
						break
					}
				}
			}
		}
	}
	
	return nil
}

// GetTransformableColumns returns columns that can be transformed
func (a *App) GetTransformableColumns(data *FileData, transformType TransformationType) []string {
	columns := []string{}
	
	for _, header := range data.Headers {
		colType := data.ColumnTypes[header]
		
		switch transformType {
		case TransformLog, TransformSqrt, TransformSquare, TransformStandardize, TransformMinMax, TransformBin:
			// These transformations require numeric columns
			if colType == "numeric" && !strings.HasSuffix(header, "#target") {
				columns = append(columns, header)
			}
		case TransformOneHot:
			// One-hot encoding requires categorical columns
			if colType == "categorical" {
				columns = append(columns, header)
			}
		}
	}
	
	return columns
}