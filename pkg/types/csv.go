package types

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// CSVFormat defines the format and parsing options for CSV files
type CSVFormat struct {
	FieldDelimiter   rune     // Field separator: ',', ';', '\t'
	DecimalSeparator rune     // Decimal separator: '.', ','
	HasHeaders       bool     // First row contains column names
	HasRowNames      bool     // First column contains row names
	NullValues       []string // Strings to treat as missing values
}

// DefaultCSVFormat returns the default CSV format options
func DefaultCSVFormat() CSVFormat {
	return CSVFormat{
		FieldDelimiter:   ',',
		DecimalSeparator: '.',
		HasHeaders:       true,
		HasRowNames:      true,
		NullValues:       []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL", "m"},
	}
}

// CSVData represents parsed CSV data with metadata
type CSVData struct {
	Matrix       Matrix   // The numerical data
	Headers      []string // Column names (if present)
	RowNames     []string // Row names (if present)
	MissingMask  [][]bool // Track NaN locations (true = missing)
	Rows         int      // Number of data rows
	Columns      int      // Number of data columns
}

// CSVParser provides methods for parsing CSV files
type CSVParser struct {
	format CSVFormat
}

// NewCSVParser creates a new CSV parser with the given format
func NewCSVParser(format CSVFormat) *CSVParser {
	return &CSVParser{format: format}
}

// Parse reads and parses CSV data from an io.Reader
func (p *CSVParser) Parse(r io.Reader) (*CSVData, error) {
	reader := csv.NewReader(r)
	reader.Comma = p.format.FieldDelimiter
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields initially

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
	if p.format.HasHeaders {
		if len(records) <= currentRow {
			return nil, fmt.Errorf("no data rows after header")
		}

		headerRow := records[currentRow]
		currentRow++

		// Extract column names, skipping first if it's a row name column
		startCol := 0
		if p.format.HasRowNames {
			startCol = 1
		}

		if startCol >= len(headerRow) {
			return nil, fmt.Errorf("no data columns found")
		}

		data.Headers = make([]string, len(headerRow)-startCol)
		copy(data.Headers, headerRow[startCol:])
	}

	// Process data rows
	dataRows := records[currentRow:]
	if len(dataRows) == 0 {
		return nil, fmt.Errorf("no data rows found")
	}

	// Determine dimensions
	startCol := 0
	if p.format.HasRowNames {
		startCol = 1
		data.RowNames = make([]string, len(dataRows))
	}

	// Validate consistent column count
	expectedCols := len(records[currentRow]) - startCol
	if expectedCols <= 0 {
		return nil, fmt.Errorf("no data columns found")
	}

	// Initialize matrix and missing mask
	data.Matrix = make(Matrix, len(dataRows))
	data.MissingMask = make([][]bool, len(dataRows))

	// Create null value map for fast lookup
	nullMap := make(map[string]bool)
	for _, nv := range p.format.NullValues {
		nullMap[nv] = true
	}

	// Parse data
	for i, row := range dataRows {
		if len(row) < startCol {
			return nil, fmt.Errorf("row %d has insufficient columns", i+1)
		}

		// Extract row name if present
		if p.format.HasRowNames {
			data.RowNames[i] = row[0]
		}

		// Validate column count
		actualCols := len(row) - startCol
		if actualCols != expectedCols {
			return nil, fmt.Errorf("row %d has %d data columns, expected %d",
				i+1, actualCols, expectedCols)
		}

		// Parse numerical data
		data.Matrix[i] = make([]float64, expectedCols)
		data.MissingMask[i] = make([]bool, expectedCols)

		for j := 0; j < expectedCols; j++ {
			value := strings.TrimSpace(row[startCol+j])

			// Check for null values
			if nullMap[value] {
				data.Matrix[i][j] = math.NaN()
				data.MissingMask[i][j] = true
				continue
			}

			// Handle decimal separator if needed
			if p.format.DecimalSeparator == ',' && p.format.FieldDelimiter == ';' {
				value = strings.ReplaceAll(value, ",", ".")
			}

			// Try to parse as float
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				// Try special cases
				switch strings.ToLower(value) {
				case "inf", "+inf", "infinity":
					val = math.Inf(1)
				case "-inf", "-infinity":
					val = math.Inf(-1)
				default:
					return nil, fmt.Errorf("cannot parse '%s' at row %d, column %d as number",
						value, i+1, j+1)
				}
			}

			data.Matrix[i][j] = val
			data.MissingMask[i][j] = false
		}
	}

	// Set dimensions
	data.Rows = len(data.Matrix)
	data.Columns = expectedCols

	// Validate column names if present
	if len(data.Headers) > 0 && len(data.Headers) != data.Columns {
		return nil, fmt.Errorf("column name count (%d) doesn't match data columns (%d)",
			len(data.Headers), data.Columns)
	}

	return data, nil
}

// DetectFormat attempts to detect the CSV format from a sample of the file
func DetectFormat(sample []byte) (*CSVFormat, error) {
	// Convert sample to string for analysis
	sampleStr := string(sample)
	lines := strings.Split(sampleStr, "\n")
	
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient sample data for format detection")
	}

	format := DefaultCSVFormat()

	// Count delimiters in first few lines
	commaCount := 0
	semicolonCount := 0
	tabCount := 0

	for i := 0; i < len(lines) && i < 5; i++ {
		line := lines[i]
		commaCount += strings.Count(line, ",")
		semicolonCount += strings.Count(line, ";")
		tabCount += strings.Count(line, "\t")
	}

	// Detect field delimiter based on counts
	if semicolonCount > commaCount && semicolonCount > tabCount {
		format.FieldDelimiter = ';'
		// Check if commas are used as decimal separators
		if strings.Contains(sampleStr, ";") && strings.Contains(sampleStr, ",") {
			// Look for patterns like "3,14" between semicolons
			format.DecimalSeparator = ','
		}
	} else if tabCount > commaCount && tabCount > semicolonCount {
		format.FieldDelimiter = '\t'
	}
	// Default is comma, already set

	// Try to detect if first row has headers by checking if values are numeric
	firstLine := strings.Split(lines[0], string(format.FieldDelimiter))
	secondLine := strings.Split(lines[1], string(format.FieldDelimiter))

	if len(firstLine) > 0 && len(secondLine) > 0 {
		// Check if first line values are non-numeric while second line has numbers
		firstLineNumeric := 0
		secondLineNumeric := 0

		for _, val := range firstLine {
			trimmedVal := strings.TrimSpace(val)
			// If using comma as decimal separator, replace comma with dot for parsing
			if format.DecimalSeparator == ',' {
				trimmedVal = strings.ReplaceAll(trimmedVal, ",", ".")
			}
			if _, err := strconv.ParseFloat(trimmedVal, 64); err == nil {
				firstLineNumeric++
			}
		}

		for _, val := range secondLine {
			trimmedVal := strings.TrimSpace(val)
			// If using comma as decimal separator, replace comma with dot for parsing
			if format.DecimalSeparator == ',' {
				trimmedVal = strings.ReplaceAll(trimmedVal, ",", ".")
			}
			if _, err := strconv.ParseFloat(trimmedVal, 64); err == nil {
				secondLineNumeric++
			}
		}

		// If first line has fewer numeric values than second, likely has headers
		format.HasHeaders = firstLineNumeric < secondLineNumeric
	}

	return &format, nil
}

// GetMissingValueInfo returns information about missing values in selected columns
func (d *CSVData) GetMissingValueInfo(selectedColumns []int) *MissingValueInfo {
	info := &MissingValueInfo{
		ColumnIndices:    []int{},
		RowsAffected:     []int{},
		TotalMissing:     0,
		MissingByColumn:  make(map[int]int),
	}

	// If no columns selected, analyze all columns
	if len(selectedColumns) == 0 {
		selectedColumns = make([]int, d.Columns)
		for i := range selectedColumns {
			selectedColumns[i] = i
		}
	}

	// Track which rows have missing values in selected columns
	rowHasMissing := make(map[int]bool)

	// Analyze each selected column
	for _, col := range selectedColumns {
		if col < 0 || col >= d.Columns {
			continue
		}

		columnMissingCount := 0
		for row := 0; row < d.Rows; row++ {
			if d.MissingMask[row][col] {
				columnMissingCount++
				rowHasMissing[row] = true
				info.TotalMissing++
			}
		}

		if columnMissingCount > 0 {
			info.ColumnIndices = append(info.ColumnIndices, col)
			info.MissingByColumn[col] = columnMissingCount
		}
	}

	// Collect affected rows
	for row := range rowHasMissing {
		info.RowsAffected = append(info.RowsAffected, row)
	}

	return info
}

// MissingValueInfo contains information about missing values in the data
type MissingValueInfo struct {
	ColumnIndices   []int         // Columns that contain missing values
	RowsAffected    []int         // Rows that contain missing values in selected columns
	TotalMissing    int           // Total number of missing values
	MissingByColumn map[int]int   // Missing count per column
}

// HasMissing returns true if there are any missing values
func (m *MissingValueInfo) HasMissing() bool {
	return m.TotalMissing > 0
}

// GetSummary returns a human-readable summary of missing values
func (m *MissingValueInfo) GetSummary() string {
	if !m.HasMissing() {
		return "No missing values in selected columns"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%d missing values found", m.TotalMissing))
	parts = append(parts, fmt.Sprintf("affecting %d rows", len(m.RowsAffected)))
	parts = append(parts, fmt.Sprintf("in %d columns", len(m.ColumnIndices)))

	return strings.Join(parts, ", ")
}