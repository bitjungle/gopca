package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
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

// FileData represents the loaded CSV data
type FileData struct {
	Headers              []string              `json:"headers"`
	RowNames             []string              `json:"rowNames,omitempty"`
	Data                 [][]string            `json:"data"`
	Rows                 int                   `json:"rows"`
	Columns              int                   `json:"columns"`
	CategoricalColumns   map[string][]string   `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]float64  `json:"numericTargetColumns,omitempty"`
	ColumnTypes          map[string]string     `json:"columnTypes,omitempty"` // "numeric", "categorical", "target"
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
		selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "Select CSV File",
			Filters: []runtime.FileFilter{
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
			runtime.LogWarning(a.ctx, fmt.Sprintf("Large file detected: %d MB", len(content)/1024/1024))
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
	runtime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))

	return fileData, nil
}

// loadExcel loads data from an Excel file
func (a *App) loadExcel(filePath string) (*FileData, error) {
	runtime.LogInfo(a.ctx, fmt.Sprintf("Loading Excel file: %s", filePath))
	
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
		runtime.LogInfo(a.ctx, fmt.Sprintf("Multiple sheets found. Using first sheet: %s", selectedSheet))
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
	runtime.LogInfo(a.ctx, fmt.Sprintf("Excel data converted to CSV, %d bytes", csvContent.Len()))
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
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to parse CSV: %v", lastErr))
			return nil, fmt.Errorf("failed to parse CSV: %w", lastErr)
		}
		runtime.LogError(a.ctx, "No data found in file")
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
		NumericTargetColumns: numericTargetData,
		ColumnTypes:          columnTypes,
	}
	
	runtime.LogInfo(a.ctx, fmt.Sprintf("Parsed data: %d rows, %d columns, %d headers", csvData.Rows, csvData.Columns, len(csvData.Headers)))

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
		NumericTargetColumns: numericTargetData,
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
	selection, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save CSV File",
		DefaultFilename: "exported_data.csv",
		Filters: []runtime.FileFilter{
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

	runtime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
	return nil
}

// SaveExcel saves data to an Excel file
func (a *App) SaveExcel(data *FileData) error {
	// Show save dialog
	selection, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save Excel File",
		DefaultFilename: "exported_data.xlsx",
		Filters: []runtime.FileFilter{
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
	
	runtime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
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

// isMissingValue checks if a value is considered missing
func isMissingValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	
	// Check common missing value representations
	lowerValue := strings.ToLower(value)
	missingIndicators := []string{"na", "n/a", "nan", "null", "none", "missing", "-", "?"}
	for _, indicator := range missingIndicators {
		if lowerValue == indicator {
			return true
		}
	}
	
	return false
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

	// Calculate mean of non-missing values
	sum := 0.0
	count := 0
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if !isMissingValue(value) {
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					sum += num
					count++
				}
			}
		}
	}

	if count == 0 {
		return // No valid values to calculate mean
	}

	mean := sum / float64(count)
	meanStr := strconv.FormatFloat(mean, 'f', -1, 64)

	// Fill missing values
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				data.Data[rowIdx][colIdx] = meanStr
			}
		}
	}
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

	// Collect non-missing values
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

	if len(values) == 0 {
		return // No valid values
	}

	// Sort values
	sort.Float64s(values)

	// Calculate median
	var median float64
	n := len(values)
	if n%2 == 0 {
		median = (values[n/2-1] + values[n/2]) / 2
	} else {
		median = values[n/2]
	}

	medianStr := strconv.FormatFloat(median, 'f', -1, 64)

	// Fill missing values
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				data.Data[rowIdx][colIdx] = medianStr
			}
		}
	}
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

	// Collect valid numeric values
	values := []float64{}
	for rowIdx := 0; rowIdx < data.Rows && rowIdx < len(data.Data); rowIdx++ {
		if colIdx < len(data.Data[rowIdx]) {
			value := strings.TrimSpace(data.Data[rowIdx][colIdx])
			if isMissingValue(value) {
				stats.Missing++
			} else if num, err := strconv.ParseFloat(value, 64); err == nil {
				values = append(values, num)
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