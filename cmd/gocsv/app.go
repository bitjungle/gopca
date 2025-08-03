package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	
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

// FileData represents the loaded CSV data
type FileData struct {
	Headers []string   `json:"headers"`
	Data    [][]string `json:"data"`
	Rows    int        `json:"rows"`
	Columns int        `json:"columns"`
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
	switch ext {
	case ".xlsx", ".xls":
		return nil, fmt.Errorf("Excel files are not yet supported. Please export to CSV format first")
	case ".tsv":
		// TSV files can be handled by setting the delimiter
		// Continue with CSV reader below
	case ".csv", "":
		// Default CSV handling
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file info for size checking
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	// Configure CSV reader
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	// Handle TSV files
	if ext == ".tsv" {
		reader.Comma = '\t'
	}
	
	// For very large files, we might want to implement streaming
	// For now, we'll read all at once but with a size check
	if fileInfo.Size() > 100*1024*1024 { // 100MB
		runtime.LogWarning(a.ctx, fmt.Sprintf("Large file detected: %d MB", fileInfo.Size()/1024/1024))
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	headers := records[0]
	data := records[1:]

	// Store the filename for display
	runtime.EventsEmit(a.ctx, "file-loaded", filepath.Base(filePath))

	return &FileData{
		Headers: headers,
		Data:    data,
		Rows:    len(data),
		Columns: len(headers),
	}, nil
}

// ValidateForGoPCA validates that the CSV data is compatible with GoPCA
func (a *App) ValidateForGoPCA(data *FileData) *ValidationResult {
	var warnings []string
	var numericColumns int
	var totalMissing int
	var hasTargetColumn bool

	// Check minimum data requirements
	if data.Rows < 2 {
		warnings = append(warnings, "ERROR: Data must have at least 2 rows (found "+fmt.Sprintf("%d", data.Rows)+")")
	}
	if data.Columns < 2 {
		warnings = append(warnings, "ERROR: Data must have at least 2 columns (found "+fmt.Sprintf("%d", data.Columns)+")")
	}

	// Check column types and missing values
	for colIdx, header := range data.Headers {
		// Check if it's a target column
		headerLower := strings.ToLower(header)
		if strings.HasSuffix(headerLower, "#target") || strings.HasSuffix(headerLower, "# target") {
			hasTargetColumn = true
			continue // Target columns are excluded from PCA
		}

		// Analyze column data
		hasNumeric := false
		hasText := false
		missingInCol := 0
		
		// Sample up to 100 rows for type detection
		sampleSize := data.Rows
		if sampleSize > 100 {
			sampleSize = 100
		}
		
		for i := 0; i < sampleSize; i++ {
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
				continue
			}
			
			// Check if numeric
			if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
				hasNumeric = true
			} else {
				hasText = true
			}
		}
		
		// Count numeric columns (excluding mixed type columns)
		if hasNumeric && !hasText {
			numericColumns++
		} else if hasText && !hasNumeric {
			warnings = append(warnings, fmt.Sprintf("WARNING: Column '%s' contains only text values", header))
		} else if hasText && hasNumeric {
			warnings = append(warnings, fmt.Sprintf("WARNING: Column '%s' contains mixed numeric and text values", header))
		}
		
		// Report high missing value percentage
		missingPercent := float64(missingInCol) / float64(sampleSize) * 100
		if missingPercent > 50 {
			warnings = append(warnings, fmt.Sprintf("WARNING: Column '%s' has %.1f%% missing values", header, missingPercent))
		}
	}

	// Check if we have enough numeric columns for PCA
	effectiveColumns := numericColumns
	if hasTargetColumn {
		warnings = append(warnings, "INFO: Target column(s) detected - these will be excluded from PCA analysis")
	}
	
	if effectiveColumns < 2 {
		warnings = append(warnings, fmt.Sprintf("ERROR: Need at least 2 numeric columns for PCA (found %d)", effectiveColumns))
	} else if effectiveColumns < 3 {
		warnings = append(warnings, fmt.Sprintf("WARNING: Only %d numeric columns found - PCA results may be limited", effectiveColumns))
	}

	// Report overall missing data
	totalCells := data.Rows * data.Columns
	missingPercent := float64(totalMissing) / float64(totalCells) * 100
	if missingPercent > 0 {
		warnings = append(warnings, fmt.Sprintf("INFO: Dataset contains %.1f%% missing values (%d cells)", missingPercent, totalMissing))
	}

	// Check for reasonable data size
	if data.Rows > 10000 {
		warnings = append(warnings, fmt.Sprintf("INFO: Large dataset detected (%d rows) - processing may take time", data.Rows))
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

	// Write headers
	if err := writer.Write(data.Headers); err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	// Write data
	for _, row := range data.Data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	runtime.EventsEmit(a.ctx, "file-saved", filepath.Base(selection))
	return nil
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return "1.0.0"
}