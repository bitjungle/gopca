package cli

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/bitjungle/gopca/pkg/types"
)

// CSVData wraps the unified types.CSVData for backward compatibility
type CSVData struct {
	*types.CSVData
}

// CSVParseOptions contains options for parsing CSV files
type CSVParseOptions struct {
	Delimiter        rune     // Field delimiter
	DecimalSeparator rune     // Decimal separator (for European formats)
	HasHeaders       bool     // First row contains column names
	HasIndex         bool     // First column contains row names
	NullValues       []string // Strings to treat as missing values
}

// NewCSVParseOptions creates default parse options
func NewCSVParseOptions() CSVParseOptions {
	return CSVParseOptions{
		Delimiter:        ',',
		DecimalSeparator: '.',
		HasHeaders:       true,
		HasIndex:         true,
		NullValues:       []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL", "m"},
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
	// Convert options to unified format
	format := types.CSVFormat{
		FieldDelimiter:   options.Delimiter,
		DecimalSeparator: options.DecimalSeparator,
		HasHeaders:       options.HasHeaders,
		HasRowNames:      options.HasIndex,
		NullValues:       options.NullValues,
	}

	// Use unified parser
	parser := types.NewCSVParser(format)
	unifiedData, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	// Wrap in CSVData for backward compatibility
	return &CSVData{CSVData: unifiedData}, nil
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
			if j < len(data.Headers) {
				colName = data.Headers[j]
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

	if len(data.Headers) > 0 {
		sb.WriteString(fmt.Sprintf("Column names: %s", strings.Join(data.Headers, ", ")))
		if len(data.Headers) > 5 {
			sb.WriteString(fmt.Sprintf(" (showing first 5 of %d)\n", len(data.Headers)))
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
