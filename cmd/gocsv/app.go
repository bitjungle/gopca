package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
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

// LoadCSV loads a CSV file and returns its data
func (a *App) LoadCSV(filePath string) (*FileData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	headers := records[0]
	data := records[1:]

	return &FileData{
		Headers: headers,
		Data:    data,
		Rows:    len(data),
		Columns: len(headers),
	}, nil
}

// ValidateForGoPCA validates that the CSV data is compatible with GoPCA
func (a *App) ValidateForGoPCA(data *FileData) (bool, []string) {
	var errors []string

	// Check if data has at least 2 rows and 2 columns
	if data.Rows < 2 {
		errors = append(errors, "Data must have at least 2 rows")
	}
	if data.Columns < 2 {
		errors = append(errors, "Data must have at least 2 columns")
	}

	// TODO: Add more validation rules based on GoPCA requirements
	// - Check for numeric columns
	// - Check for missing values
	// - Validate column types

	return len(errors) == 0, errors
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return "1.0.0"
}