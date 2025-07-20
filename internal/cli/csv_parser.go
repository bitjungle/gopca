package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// CSVData represents parsed CSV data with metadata
type CSVData struct {
	Matrix      types.Matrix // The numerical data
	RowNames    []string     // Row index names (if present)
	ColumnNames []string     // Column header names (if present)
	Rows        int          // Number of data rows
	Columns     int          // Number of data columns
}

// CSVParseOptions contains options for parsing CSV files
type CSVParseOptions struct {
	Delimiter  rune     // Field delimiter
	HasHeaders bool     // First row contains column names
	HasIndex   bool     // First column contains row names
	NullValues []string // Strings to treat as missing values
}

// NewCSVParseOptions creates default parse options
func NewCSVParseOptions() CSVParseOptions {
	return CSVParseOptions{
		Delimiter:  ',',
		HasHeaders: true,
		HasIndex:   true,
		NullValues: []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL"},
	}
}

// ParseCSV reads and parses a CSV file according to the given options
func ParseCSV(filename string, options CSVParseOptions) (*CSVData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return ParseCSVReader(file, options)
}

// ParseCSVReader parses CSV data from an io.Reader
func ParseCSVReader(r io.Reader, options CSVParseOptions) (*CSVData, error) {
	reader := csv.NewReader(r)
	reader.Comma = options.Delimiter
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	data := &CSVData{}
	currentRow := 0

	// Handle headers
	if options.HasHeaders {
		if len(records) <= currentRow {
			return nil, fmt.Errorf("no data rows after header")
		}
		
		headerRow := records[currentRow]
		currentRow++
		
		// Extract column names, skipping first if it's an index column
		startCol := 0
		if options.HasIndex {
			startCol = 1
		}
		
		if startCol >= len(headerRow) {
			return nil, fmt.Errorf("no data columns found")
		}
		
		data.ColumnNames = make([]string, len(headerRow)-startCol)
		copy(data.ColumnNames, headerRow[startCol:])
	}

	// Process data rows
	dataRows := records[currentRow:]
	if len(dataRows) == 0 {
		return nil, fmt.Errorf("no data rows found")
	}

	// Determine dimensions
	startCol := 0
	if options.HasIndex {
		startCol = 1
		data.RowNames = make([]string, len(dataRows))
	}

	// Validate consistent column count
	expectedCols := len(records[currentRow]) - startCol
	if expectedCols <= 0 {
		return nil, fmt.Errorf("no data columns found")
	}

	// Initialize matrix
	data.Matrix = make(types.Matrix, len(dataRows))
	
	// Parse data
	nullMap := makeNullMap(options.NullValues)
	
	for i, row := range dataRows {
		if len(row) < startCol {
			return nil, fmt.Errorf("row %d has insufficient columns", i+1)
		}

		// Extract row name if present
		if options.HasIndex {
			data.RowNames[i] = row[0]
		}

		// Parse numerical data
		rowData := make([]float64, expectedCols)
		actualCols := len(row) - startCol
		
		if actualCols != expectedCols {
			return nil, fmt.Errorf("row %d has %d data columns, expected %d", 
				i+1, actualCols, expectedCols)
		}

		for j := 0; j < expectedCols; j++ {
			value := strings.TrimSpace(row[startCol+j])
			
			if nullMap[value] {
				rowData[j] = math.NaN()
			} else {
				val, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse value '%s' at row %d, column %d: %w",
						value, i+1, j+1, err)
				}
				
				// Check for special float values
				switch value {
				case "Inf", "+Inf":
					val = math.Inf(1)
				case "-Inf":
					val = math.Inf(-1)
				}
				
				rowData[j] = val
			}
		}
		
		data.Matrix[i] = rowData
	}

	// Set dimensions
	data.Rows = len(data.Matrix)
	data.Columns = expectedCols

	// Validate column names if present
	if len(data.ColumnNames) > 0 && len(data.ColumnNames) != data.Columns {
		return nil, fmt.Errorf("column name count (%d) doesn't match data columns (%d)",
			len(data.ColumnNames), data.Columns)
	}

	return data, nil
}

// makeNullMap creates a map for fast null value lookup
func makeNullMap(nullValues []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range nullValues {
		m[v] = true
	}
	return m
}

// ValidateCSVData performs basic validation on parsed CSV data
func ValidateCSVData(data *CSVData) error {
	if data == nil {
		return fmt.Errorf("nil CSV data")
	}

	if len(data.Matrix) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	if data.Rows != len(data.Matrix) {
		return fmt.Errorf("row count mismatch")
	}

	// Check for consistent column count
	for i, row := range data.Matrix {
		if len(row) != data.Columns {
			return fmt.Errorf("row %d has %d columns, expected %d", 
				i+1, len(row), data.Columns)
		}
	}

	// Check for all NaN columns
	for j := 0; j < data.Columns; j++ {
		allNaN := true
		for i := 0; i < data.Rows; i++ {
			if !math.IsNaN(data.Matrix[i][j]) {
				allNaN = false
				break
			}
		}
		if allNaN {
			colName := fmt.Sprintf("%d", j+1)
			if j < len(data.ColumnNames) {
				colName = data.ColumnNames[j]
			}
			return fmt.Errorf("column '%s' contains only missing values", colName)
		}
	}

	return nil
}

// GetDataSummary returns a summary of the CSV data
func GetDataSummary(data *CSVData) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Data dimensions: %d rows × %d columns\n", data.Rows, data.Columns))
	
	if len(data.ColumnNames) > 0 {
		sb.WriteString(fmt.Sprintf("Column names: %s", strings.Join(data.ColumnNames, ", ")))
		if len(data.ColumnNames) > 5 {
			sb.WriteString(fmt.Sprintf(" (showing first 5 of %d)\n", len(data.ColumnNames)))
		} else {
			sb.WriteString("\n")
		}
	}
	
	if len(data.RowNames) > 0 {
		sb.WriteString(fmt.Sprintf("Row names: %s", strings.Join(data.RowNames[:min(5, len(data.RowNames))], ", ")))
		if len(data.RowNames) > 5 {
			sb.WriteString(fmt.Sprintf(" ... (showing first 5 of %d)\n", len(data.RowNames)))
		} else {
			sb.WriteString("\n")
		}
	}
	
	// Count missing values
	missingCount := 0
	for i := 0; i < data.Rows; i++ {
		for j := 0; j < data.Columns; j++ {
			if math.IsNaN(data.Matrix[i][j]) {
				missingCount++
			}
		}
	}
	
	totalValues := data.Rows * data.Columns
	missingPercent := float64(missingCount) / float64(totalValues) * 100
	sb.WriteString(fmt.Sprintf("Missing values: %d (%.1f%%)\n", missingCount, missingPercent))
	
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}