// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"

	"github.com/bitjungle/gopca/internal/version"
	parquet "github.com/parquet-go/parquet-go"
	"github.com/bitjungle/gopca/pkg/dataquality"
	"github.com/bitjungle/gopca/pkg/integration"
	"github.com/bitjungle/gopca/pkg/transform"
	"github.com/bitjungle/gopca/pkg/types"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx               context.Context
	history           *CommandHistory
	currentData       *FileData
	hasUnsavedChanges bool
	pendingZipPath    string // path to a downloaded ZIP awaiting entry selection
}

func (a *App) logInfo(msg string) {
	if a.ctx != nil {
		wailsruntime.LogInfo(a.ctx, msg)
	}
}

func (a *App) logWarning(msg string) {
	if a.ctx != nil {
		wailsruntime.LogWarning(a.ctx, msg)
	}
}

func (a *App) logError(msg string) {
	if a.ctx != nil {
		wailsruntime.LogError(a.ctx, msg)
	}
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

// markDirty marks the app as having unsaved changes and notifies the frontend.
// Only emits an event on the first transition to dirty to avoid noise.
func (a *App) markDirty() {
	if !a.hasUnsavedChanges {
		a.hasUnsavedChanges = true
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "unsaved-state-changed", true)
		}
	}
}

// markClean marks the app as having no unsaved changes and notifies the frontend.
// Called after a successful save or when a new file is loaded.
func (a *App) markClean() {
	if a.hasUnsavedChanges {
		a.hasUnsavedChanges = false
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "unsaved-state-changed", false)
		}
	}
}

// HasUnsavedChanges returns true if there are unsaved data changes.
// Used by OnBeforeClose to determine whether to prompt the user.
func (a *App) HasUnsavedChanges() bool {
	return a.hasUnsavedChanges
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
					DisplayName: "Supported Files (*.csv,*.xlsx,*.xls,*.tsv,*.parquet)",
					Pattern:     "*.csv;*.xlsx;*.xls;*.tsv;*.parquet",
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
					DisplayName: "Parquet Files (*.parquet)",
					Pattern:     "*.parquet",
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

	// If filePath is a remote URL, fetch it to a temp file first.
	// The temp file is named with the detected extension so the switch below routes correctly.
	if strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "http://") {
		a.logInfo(fmt.Sprintf("Fetching remote file: %s", filePath))
		tmpPath, err := fetchRemoteFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote file: %w", err)
		}
		defer os.Remove(tmpPath)
		filePath = tmpPath
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
	case ".parquet":
		// Handle Parquet files
		var err error
		fileData, err = a.loadParquet(filePath)
		if err != nil {
			return nil, fmt.Errorf("error loading Parquet file: %w", err)
		}
	case ".tsv", ".csv", "":
		// Handle CSV/TSV files
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// Check file size
		if len(content) > 100*1024*1024 { // 100MB
			a.logWarning(fmt.Sprintf("Large file detected: %d MB", len(content)/1024/1024))
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
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))
	}

	// Store the current data reference for undo/redo operations
	a.currentData = fileData
	// Clear history when loading new file
	a.history.Clear()
	a.markClean()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}

	return fileData, nil
}

// writeCSVRow writes a single row of cells to b in RFC 4180 CSV format.
func writeCSVRow(b *strings.Builder, cells []string) {
	for i, cell := range cells {
		if i > 0 {
			b.WriteString(",")
		}
		if strings.ContainsAny(cell, ",\"\n\r") {
			b.WriteString("\"")
			b.WriteString(strings.ReplaceAll(cell, "\"", "\"\""))
			b.WriteString("\"")
		} else {
			b.WriteString(cell)
		}
	}
	b.WriteString("\n")
}

// loadExcel loads data from an Excel file
func (a *App) loadExcel(filePath string) (*FileData, error) {
	a.logInfo(fmt.Sprintf("Loading Excel file: %s", filePath))

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
		a.logInfo(fmt.Sprintf("Multiple sheets found. Using first sheet: %s", selectedSheet))
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
		writeCSVRow(&csvContent, row)
	}

	// Parse the CSV content using GoPCA's parser
	a.logInfo(fmt.Sprintf("Excel data converted to CSV, %d bytes", csvContent.Len()))
	return a.parseCSVContent(csvContent.String(), ".csv")
}

// loadParquet loads data from a Parquet file and converts it to a FileData struct.
func (a *App) loadParquet(filePath string) (*FileData, error) {
	a.logInfo(fmt.Sprintf("Loading Parquet file: %s", filePath))

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Parquet file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat Parquet file: %w", err)
	}

	pf, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Parquet file: %w", err)
	}

	// Build headers: prepend Sample_ID, mark string columns as #target.
	// String columns (ByteArray kind) are categorical identifiers — marking them
	// as #target makes GoPCA treat them as group labels for coloring the scores plot.
	// Sample_ID provides a unique integer row identifier since string columns like
	// "country" are not unique (the same country appears once per year).
	fields := pf.Schema().Fields()
	headers := make([]string, 0, len(fields)+1)
	headers = append(headers, "Sample_ID")
	for _, field := range fields {
		name := field.Name()
		if field.Type().Kind() == parquet.ByteArray || field.Type().Kind() == parquet.FixedLenByteArray {
			name = name + "#target"
		}
		headers = append(headers, name)
	}

	// Build CSV: header row followed by data rows
	var csvContent strings.Builder
	writeCSVRow(&csvContent, headers)

	rowNum := 0
	buf := make([]parquet.Row, 128)
	for _, rg := range pf.RowGroups() {
		rows := rg.Rows()
		for {
			n, readErr := rows.ReadRows(buf)
			for i := 0; i < n; i++ {
				rowNum++
				row := buf[i]
				cells := make([]string, len(fields)+1)
				cells[0] = strconv.Itoa(rowNum)
				for j, val := range row {
					if j < len(fields) {
						cells[j+1] = parquetValueToString(val)
					}
				}
				writeCSVRow(&csvContent, cells)
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				rows.Close()
				return nil, fmt.Errorf("error reading Parquet rows: %w", readErr)
			}
		}
		rows.Close()
	}

	a.logInfo(fmt.Sprintf("Parquet data converted to CSV, %d bytes", csvContent.Len()))
	return a.parseCSVContent(csvContent.String(), ".csv")
}

// parquetValueToString converts a parquet.Value to its string representation.
// Null values become empty strings (GoCSV's convention for missing values).
func parquetValueToString(v parquet.Value) string {
	if v.IsNull() {
		return ""
	}
	switch v.Kind() {
	case parquet.Boolean:
		if v.Boolean() {
			return "true"
		}
		return "false"
	case parquet.Int32:
		return strconv.FormatInt(int64(v.Int32()), 10)
	case parquet.Int64:
		return strconv.FormatInt(v.Int64(), 10)
	case parquet.Float:
		return strconv.FormatFloat(float64(v.Float()), 'f', -1, 32)
	case parquet.Double:
		return strconv.FormatFloat(v.Double(), 'f', -1, 64)
	case parquet.ByteArray, parquet.FixedLenByteArray:
		return string(v.ByteArray())
	default:
		return fmt.Sprintf("%v", v)
	}
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
	var successfulFormat types.CSVFormat

	// Try each format until one works
	for _, format := range formats {
		reader := strings.NewReader(content)
		data, catData, targetData, err := types.ParseCSVMixedWithTargets(reader, format, nil)
		if err == nil && data != nil && data.Columns > 0 {
			csvData = data
			categoricalData = catData
			numericTargetData = targetData
			successfulFormat = format
			break
		}
		if err != nil {
			lastErr = err
		}
	}

	if csvData == nil {
		if lastErr != nil {
			a.logError(fmt.Sprintf("Failed to parse CSV: %v", lastErr))
			return nil, fmt.Errorf("failed to parse CSV: %w", lastErr)
		}
		a.logError("No data found in file")
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

	a.logInfo(fmt.Sprintf("Parsed data: %d rows, %d columns, %d headers", csvData.Rows, csvData.Columns, len(csvData.Headers)))

	// If we have categorical or target columns, we need to combine them with numeric data
	// for the full data display
	if len(categoricalData) > 0 || len(numericTargetData) > 0 {
		// Get all original headers to preserve column order, then merge all column types.
		allOriginalHeaders := pkgcsv.GetOriginalHeaders(content, successfulFormat)
		combined := pkgcsv.CombineColumns(csvData, categoricalData, numericTargetData, allOriginalHeaders)
		fileData = &FileData{
			Headers:              combined.Headers,
			RowNames:             combined.RowNames,
			Data:                 combined.Data,
			Rows:                 combined.Rows,
			Columns:              combined.Columns,
			CategoricalColumns:   combined.CategoricalColumns,
			NumericTargetColumns: ConvertFloat64MapToJSON(combined.NumericTargetData),
			ColumnTypes:          combined.ColumnTypes,
		}
	}

	return fileData, nil
}


// ValidateForGoPCA validates that the CSV data is compatible with GoPCA.
func (a *App) ValidateForGoPCA(data *FileData) *ValidationResult {
	in := integration.ValidationInput{
		Headers:     data.Headers,
		Data:        data.Data,
		ColumnTypes: data.ColumnTypes,
		RowNames:    data.RowNames,
		Rows:        data.Rows,
		Columns:     data.Columns,
	}
	res := integration.ValidateForGoPCA(in)
	return &ValidationResult{IsValid: res.IsValid, Messages: res.Messages}
}

// SaveCSV saves the data to a CSV file
func (a *App) SaveCSV(data *FileData) error {
	// Show save dialog
	selection, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Save CSV File",
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

	// Convert FileData to pkg/csv.Data
	csvData := &pkgcsv.Data{
		Headers:    data.Headers,
		RowNames:   data.RowNames,
		StringData: data.Data,
		Rows:       data.Rows,
		Columns:    data.Columns,
	}

	// Use pkg/csv writer with appropriate options
	opts := pkgcsv.DefaultOptions()
	opts.HasHeaders = true
	opts.HasRowNames = len(data.RowNames) > 0
	
	// Write using the unified CSV writer
	if err := pkgcsv.SaveFile(selection, csvData, opts); err != nil {
		return fmt.Errorf("error writing CSV file: %w", err)
	}

	wailsruntime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
	a.markClean()
	return nil
}

// SaveExcel saves data to an Excel file
func (a *App) SaveExcel(data *FileData) error {
	// Show save dialog
	selection, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Save Excel File",
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
	a.markClean()
	return nil
}

// FillMissingValuesRequest represents a request to fill missing values.
type FillMissingValuesRequest struct {
	Strategy string `json:"strategy"` // "mean", "median", "mode", "forward", "backward", "custom"
	Column   string `json:"column"`   // Column name, or empty for all columns
	Value    string `json:"value"`    // Custom value for "custom" strategy
}

// AnalyzeMissingValues analyzes missing value patterns in the data.
// Delegates to the dataquality package.
func (a *App) AnalyzeMissingValues(data *FileData) *dataquality.MissingValueStats {
	if data == nil || len(data.Data) == 0 {
		return &dataquality.MissingValueStats{
			ColumnStats: make(map[string]*dataquality.ColumnMissing),
			RowStats:    make(map[int]*dataquality.RowMissing),
		}
	}
	return dataquality.AnalyzeMissing(data.Data, data.Headers)
}

// FillMissingValues fills missing values in the data according to the specified strategy.
// Delegates to the dataquality package.
func (a *App) FillMissingValues(data *FileData, request FillMissingValuesRequest) (*FileData, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to process")
	}

	req := dataquality.FillRequest{
		Strategy: request.Strategy,
		Column:   request.Column,
		Value:    request.Value,
	}

	newData, err := dataquality.Fill(data.Data, data.Headers, data.ColumnTypes, req)
	if err != nil {
		return nil, err
	}

	result := &FileData{
		Headers:              data.Headers,
		RowNames:             data.RowNames,
		Data:                 newData,
		Rows:                 data.Rows,
		Columns:              data.Columns,
		CategoricalColumns:   data.CategoricalColumns,
		NumericTargetColumns: data.NumericTargetColumns,
		ColumnTypes:          data.ColumnTypes,
	}

	a.markDirty()
	return result, nil
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return version.Get().Short()
}

// AnalyzeDataQuality performs comprehensive data quality analysis on the given
// file data. Delegates to the dataquality package.
func (a *App) AnalyzeDataQuality(data *FileData) (*dataquality.DataQualityReport, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to analyze")
	}

	in := dataquality.AnalysisInput{
		Data:        data.Data,
		Headers:     data.Headers,
		ColumnTypes: data.ColumnTypes,
		RowNames:    data.RowNames,
		Rows:        data.Rows,
		Columns:     data.Columns,
	}

	return dataquality.AnalyzeDataQuality(in)
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
	// Use shared integration package to detect GoPCA
	integrationConfig := integration.AppConfig{
		Name:        "gopca-desktop",
		CommonPaths: getGoPCAPathsWithDev(),
		DisplayName: "GoPCA Desktop",
	}

	appStatus := integration.CheckApp(integrationConfig)

	status := &GoPCAStatus{
		Installed: appStatus.Installed,
		Path:      appStatus.Path,
		Error:     appStatus.Error,
	}

	// If found, try to get version
	if status.Installed && status.Path != "" {
		cmd := exec.Command(status.Path, "--version")
		output, err := cmd.Output()
		if err == nil {
			status.Version = strings.TrimSpace(string(output))
		}
	}

	return status
}

// getGoPCAPathsWithDev returns common paths plus development paths for GoPCA
func getGoPCAPathsWithDev() []string {
	// Start with common installation paths
	paths := integration.GetCommonPaths("gopca-desktop")

	// Add development-specific paths
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	cwd, _ := os.Getwd()

	devPaths := []string{
		// When running from cmd/gocsv with make csv-dev
		filepath.Join(execDir, "../../build/bin/gopca-desktop"),
		filepath.Join(execDir, "../../build/bin/gopca-desktop.exe"),
		// When running from build output directory
		filepath.Join(execDir, "../gopca-desktop"),
		filepath.Join(execDir, "../gopca-desktop.exe"),
		// macOS app bundle in development
		filepath.Join(execDir, "../../build/bin/GoPCA Desktop.app/Contents/MacOS/gopca-desktop"),
		// When running with wails dev from cmd/gocsv directory
		filepath.Join(cwd, "../../build/bin/gopca-desktop"),
		filepath.Join(cwd, "../../build/bin/gopca-desktop.exe"),
		filepath.Join(cwd, "../gopca-desktop/build/bin/gopca-desktop.app/Contents/MacOS/gopca-desktop"),
		// Direct path to GoPCA Desktop build output
		filepath.Join(cwd, "../../cmd/gopca-desktop/build/bin/gopca-desktop.app/Contents/MacOS/gopca-desktop"),
	}

	paths = append(paths, devPaths...)
	return paths
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

	// Launch GoPCA with the file using the shared integration package
	if err := integration.LaunchWithFile(status.Path, tempFile); err != nil {
		return fmt.Errorf("failed to launch GoPCA Desktop: %w", err)
	}

	// Log the temporary file location
	a.logInfo(fmt.Sprintf("Exported data to: %s", tempFile))
	a.logInfo(fmt.Sprintf("Launched GoPCA Desktop: %s", status.Path))

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
	CanUndo    bool     `json:"canUndo"`
	CanRedo    bool     `json:"canRedo"`
	History    []string `json:"history"`
	CurrentPos int      `json:"currentPos"`
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
func (a *App) Undo(data *FileData) (*FileData, error) {
	if err := a.history.Undo(data); err != nil {
		return nil, err
	}
	a.markDirty()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}
	return data, nil
}

// Redo performs a redo operation
func (a *App) Redo(data *FileData) (*FileData, error) {
	if err := a.history.Redo(data); err != nil {
		return nil, err
	}
	a.markDirty()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}
	return data, nil
}

// ClearHistory clears the command history
func (a *App) ClearHistory() {
	a.history.Clear()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}
}

// Command Execution Methods
// ========================
// The following Execute* methods provide undo/redo support for all data operations.
// They follow a consistent pattern:
// 1. Create a command object with the current state
// 2. Execute the command through the history manager
// 3. Update the current data reference
// 4. Emit state change event for UI updates
//
// All methods return the updated FileData to ensure the frontend stays synchronized.

// executeCommand is a helper that executes a command and handles common operations
func (a *App) executeCommand(cmd Command, data *FileData, operation string) (*FileData, error) {
	if err := a.history.Execute(cmd, data); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	a.currentData = data // Store the current data
	a.markDirty()

	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}
	return data, nil
}

// ExecuteCellEdit executes a cell edit command
func (a *App) ExecuteCellEdit(data *FileData, row, col int, oldValue, newValue string) (*FileData, error) {
	cmd := NewCellEditCommand(row, col, oldValue, newValue)
	return a.executeCommand(cmd, data, "edit cell")
}

// ExecuteHeaderEdit executes a header edit command
func (a *App) ExecuteHeaderEdit(data *FileData, col int, oldValue, newValue string) (*FileData, error) {
	cmd := NewHeaderEditCommand(col, oldValue, newValue)
	return a.executeCommand(cmd, data, "edit header")
}

// ExecuteFillMissingValues executes a fill missing values command
func (a *App) ExecuteFillMissingValues(data *FileData, strategy, column, customValue string) (*FileData, error) {
	cmd := NewFillMissingValuesCommand(a, data, strategy, column, customValue)
	return a.executeCommand(cmd, data, "fill missing values")
}

// ImportFileInfo represents information about a file to be imported
type ImportFileInfo struct {
	FileName   string   `json:"fileName"`
	FilePath   string   `json:"filePath"`
	FileSize   int64    `json:"fileSize"`
	FileFormat string   `json:"fileFormat"` // "csv", "tsv", "excel"
	Encoding   string   `json:"encoding"`
	Sheets     []string `json:"sheets,omitempty"` // For Excel files
	Error      string   `json:"error,omitempty"`
}

// ImportOptions represents options for importing a file
type ImportOptions struct {
	Format          string `json:"format"`
	Delimiter       string `json:"delimiter,omitempty"` // For CSV/TSV
	HasHeaders      bool   `json:"hasHeaders"`
	HeaderRow       int    `json:"headerRow"`                 // 0-based
	Sheet           string `json:"sheet,omitempty"`           // For Excel
	Range           string `json:"range,omitempty"`           // For Excel (e.g., "A1:Z100")
	RowNameColumn   int    `json:"rowNameColumn"`             // -1 if none, 0-based
	SkipRows        int    `json:"skipRows"`                  // Number of rows to skip from top
	MaxRows         int    `json:"maxRows"`                   // 0 for all rows
	SelectedColumns []int  `json:"selectedColumns,omitempty"` // Indices of columns to import
}

// FilePreview represents a preview of file contents
type FilePreview struct {
	Headers     []string   `json:"headers"`
	Data        [][]string `json:"data"`        // First N rows
	ColumnTypes []string   `json:"columnTypes"` // Detected types
	Delimiter   string     `json:"delimiter"`   // Detected delimiter
	TotalRows   int        `json:"totalRows"`
	TotalCols   int        `json:"totalCols"`
	Issues      []string   `json:"issues,omitempty"`
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
		// JSON import is not supported. Reject explicitly rather than letting a
		// .json file fall through to content-based detection (where it would be
		// misclassified as CSV). See #719.
		return nil, fmt.Errorf("JSON files are not supported by GoCSV")
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

	content := string(buf)
	content = strings.TrimSpace(content)

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
		preview.ColumnTypes[i] = pkgcsv.DetectColumnType(allData, i)
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
		preview.ColumnTypes[i] = pkgcsv.DetectColumnType(rows, i)
	}

	return preview, nil
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
		colType := pkgcsv.DetectColumnType(allData, i)
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
		colType := pkgcsv.DetectColumnType(rows, i)
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

// SelectFileForImport opens a file dialog and returns the selected file path
func (a *App) SelectFileForImport() (string, error) {
	dialogOptions := wailsruntime.OpenDialogOptions{
		Title: "Select file to import",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "All Supported Files",
				Pattern:     "*.csv;*.tsv;*.xlsx;*.xls",
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

// TransformationType represents the type of transformation
type TransformationType string

const (
	TransformLog         TransformationType = "log"
	TransformSqrt        TransformationType = "sqrt"
	TransformSquare      TransformationType = "square"
	TransformStandardize TransformationType = "standardize"
	TransformMinMax      TransformationType = "minmax"
	TransformBin         TransformationType = "bin"
	TransformOneHot      TransformationType = "onehot"
)

// TransformOptions represents options for data transformation
type TransformOptions struct {
	Type     TransformationType `json:"type"`
	Columns  []string           `json:"columns"`
	BinCount int                `json:"binCount,omitempty"` // For binning
	MinValue float64            `json:"minValue,omitempty"` // For min-max scaling
	MaxValue float64            `json:"maxValue,omitempty"` // For min-max scaling
}

// TransformationResult represents the result of a transformation
type TransformationResult struct {
	Success            bool      `json:"success"`
	TransformedColumns []string  `json:"transformedColumns"`
	NewColumns         []string  `json:"newColumns,omitempty"`
	Messages           []string  `json:"messages"`
	Data               *FileData `json:"data"`
}

// ApplyTransformation applies a transformation to the data with undo support
func (a *App) ApplyTransformation(data *FileData, options TransformOptions) (*TransformationResult, error) {
	// Use the command pattern for undo support
	cmd := NewTransformCommand(a, data, options)
	if err := a.history.Execute(cmd, data); err != nil {
		return nil, fmt.Errorf("apply transformation: %w", err)
	}
	a.currentData = data // Store the current data
	a.markDirty()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}

	// Return the transformation result
	return cmd.result, nil
}

// applyTransformationInternal delegates to pkg/transform and maps the result
// back into the FileData-based TransformationResult expected by the Wails layer.
func (a *App) applyTransformationInternal(data *FileData, options TransformOptions) (*TransformationResult, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to transform")
	}

	in := transform.Input{
		Data:               data.Data,
		Headers:            data.Headers,
		ColumnTypes:        data.ColumnTypes,
		CategoricalColumns: data.CategoricalColumns,
		Rows:               data.Rows,
		Columns:            data.Columns,
	}
	opts := transform.Options{
		Type:     transform.Type(options.Type),
		Columns:  options.Columns,
		BinCount: options.BinCount,
		MinValue: options.MinValue,
		MaxValue: options.MaxValue,
	}

	res, err := transform.Apply(in, opts)
	if err != nil {
		return nil, err
	}

	newData := &FileData{
		Headers:              res.Headers,
		Data:                 res.Data,
		Rows:                 data.Rows,
		Columns:              res.Columns,
		CategoricalColumns:   res.CategoricalColumns,
		NumericTargetColumns: make(map[string][]types.JSONFloat64),
		ColumnTypes:          res.ColumnTypes,
	}
	if data.RowNames != nil {
		newData.RowNames = make([]string, len(data.RowNames))
		copy(newData.RowNames, data.RowNames)
	}
	// Carry over any numeric target columns untouched.
	for k, v := range data.NumericTargetColumns {
		newData.NumericTargetColumns[k] = v
	}

	return &TransformationResult{
		Success:            true,
		TransformedColumns: res.TransformedColumns,
		NewColumns:         res.NewColumns,
		Messages:           res.Messages,
		Data:               newData,
	}, nil
}

// GetTransformableColumns returns columns eligible for the given transformation type.
func (a *App) GetTransformableColumns(data *FileData, transformType TransformationType) []string {
	in := transform.Input{
		Data:        data.Data,
		Headers:     data.Headers,
		ColumnTypes: data.ColumnTypes,
		Rows:        data.Rows,
		Columns:     data.Columns,
	}
	return transform.GetTransformableColumns(in, transform.Type(transformType))
}

// ExecuteDeleteRows deletes multiple rows with undo support
func (a *App) ExecuteDeleteRows(data *FileData, rowIndices []int) (*FileData, error) {
	cmd := NewDeleteRowsCommand(data, rowIndices)
	return a.executeCommand(cmd, data, "delete rows")
}

// ExecuteDeleteColumns deletes multiple columns with undo support
func (a *App) ExecuteDeleteColumns(data *FileData, colIndices []int) (*FileData, error) {
	cmd := NewDeleteColumnsCommand(a, data, colIndices)
	return a.executeCommand(cmd, data, "delete columns")
}

// ExecuteInsertRow inserts a new row with undo support
func (a *App) ExecuteInsertRow(data *FileData, index int) (*FileData, error) {
	cmd := NewInsertRowCommand(a, data, index)
	return a.executeCommand(cmd, data, "insert row")
}

// ExecuteInsertColumn inserts a new column with undo support
func (a *App) ExecuteInsertColumn(data *FileData, index int, name string) (*FileData, error) {
	cmd := NewInsertColumnCommand(a, data, index, name)
	return a.executeCommand(cmd, data, "insert column")
}

// ExecuteToggleTargetColumn toggles the #target suffix on a column with undo support
func (a *App) ExecuteToggleTargetColumn(data *FileData, colIndex int) (*FileData, error) {
	// Validate bounds
	if colIndex < 0 || colIndex >= len(data.Headers) {
		return nil, fmt.Errorf("toggle target column: invalid column index: %d", colIndex)
	}

	cmd := NewToggleTargetColumnCommand(a, data, colIndex)
	if cmd == nil {
		return nil, fmt.Errorf("toggle target column: invalid column index: %d", colIndex)
	}

	return a.executeCommand(cmd, data, "toggle target column")
}

// ExecuteDuplicateRows duplicates selected rows with undo support
func (a *App) ExecuteDuplicateRows(data *FileData, rowIndices []int) (*FileData, error) {
	if len(rowIndices) == 0 {
		return nil, fmt.Errorf("duplicate rows: no rows selected for duplication")
	}

	cmd := NewDuplicateRowCommand(a, data, rowIndices)
	return a.executeCommand(cmd, data, "duplicate rows")
}
