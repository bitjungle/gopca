package io

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/bitjungle/complab/pkg/types"
)

// CSVOptions contains configuration for CSV reading/writing
type CSVOptions struct {
	Delimiter      rune
	HasHeader      bool
	SkipRows       int
	MaxRows        int // 0 for unlimited
	Columns        []int // Specific columns to read, empty for all
	NullValues     []string // Strings to treat as NaN
	StreamingMode  bool // For large files
}

// DefaultCSVOptions returns default CSV options
func DefaultCSVOptions() CSVOptions {
	return CSVOptions{
		Delimiter:  ',',
		HasHeader:  true,
		SkipRows:   0,
		MaxRows:    0,
		Columns:    nil,
		NullValues: []string{"", "NA", "N/A", "nan", "NaN", "null", "NULL"},
	}
}

// LoadCSV reads a CSV file into a Matrix
func LoadCSV(filename string, options CSVOptions) (types.Matrix, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ReadCSV(file, options)
}

// ReadCSV reads CSV data from an io.Reader
func ReadCSV(r io.Reader, options CSVOptions) (types.Matrix, []string, error) {
	reader := csv.NewReader(r)
	reader.Comma = options.Delimiter
	reader.FieldsPerRecord = -1 // Variable number of fields allowed initially
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = options.StreamingMode

	var headers []string
	var data [][]float64
	rowCount := 0
	skipCount := options.SkipRows

	// Create a map for quick null value lookup
	nullMap := make(map[string]bool)
	for _, nv := range options.NullValues {
		nullMap[nv] = true
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading CSV at row %d: %w", rowCount+1, err)
		}

		// Skip rows if needed
		if skipCount > 0 {
			skipCount--
			continue
		}

		// Handle header
		if options.HasHeader && len(headers) == 0 {
			headers = make([]string, len(record))
			copy(headers, record)
			continue
		}

		// Check max rows
		if options.MaxRows > 0 && len(data) >= options.MaxRows {
			break
		}

		// Parse data row
		row, err := parseRow(record, options.Columns, nullMap)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing row %d: %w", rowCount+1, err)
		}

		data = append(data, row)
		rowCount++
	}

	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no data rows found")
	}

	// Validate rectangular matrix
	cols := len(data[0])
	for i, row := range data {
		if len(row) != cols {
			return nil, nil, fmt.Errorf("inconsistent columns at row %d: expected %d, got %d", 
				i+1, cols, len(row))
		}
	}

	// Extract selected column headers if specified
	if len(options.Columns) > 0 && len(headers) > 0 {
		selectedHeaders := make([]string, len(options.Columns))
		for i, col := range options.Columns {
			if col < len(headers) {
				selectedHeaders[i] = headers[col]
			} else {
				selectedHeaders[i] = fmt.Sprintf("Column_%d", col)
			}
		}
		headers = selectedHeaders
	}

	return data, headers, nil
}

// parseRow converts string values to float64
func parseRow(record []string, columns []int, nullMap map[string]bool) ([]float64, error) {
	// Determine which columns to parse
	var colIndices []int
	if len(columns) > 0 {
		colIndices = columns
	} else {
		colIndices = make([]int, len(record))
		for i := range record {
			colIndices[i] = i
		}
	}

	row := make([]float64, len(colIndices))
	for i, col := range colIndices {
		if col >= len(record) {
			return nil, fmt.Errorf("column index %d out of bounds", col)
		}

		val := strings.TrimSpace(record[col])
		
		// Check for null values
		if nullMap[val] {
			row[i] = math.NaN()
			continue
		}

		// Try to parse as float
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			// Try to handle special cases
			switch strings.ToLower(val) {
			case "inf", "+inf", "infinity":
				row[i] = math.Inf(1)
			case "-inf", "-infinity":
				row[i] = math.Inf(-1)
			default:
				return nil, fmt.Errorf("cannot parse '%s' as float in column %d", val, col)
			}
		} else {
			row[i] = f
		}
	}

	return row, nil
}

// SaveCSV writes a Matrix to a CSV file
func SaveCSV(filename string, data types.Matrix, headers []string, options CSVOptions) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return WriteCSV(file, data, headers, options)
}

// WriteCSV writes Matrix data to an io.Writer
func WriteCSV(w io.Writer, data types.Matrix, headers []string, options CSVOptions) error {
	writer := csv.NewWriter(w)
	writer.Comma = options.Delimiter
	defer writer.Flush()

	// Write headers if provided
	if len(headers) > 0 && options.HasHeader {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Write data rows
	for i, row := range data {
		record := make([]string, len(row))
		for j, val := range row {
			if math.IsNaN(val) {
				record[j] = "NaN"
			} else if math.IsInf(val, 1) {
				record[j] = "Inf"
			} else if math.IsInf(val, -1) {
				record[j] = "-Inf"
			} else {
				record[j] = strconv.FormatFloat(val, 'g', -1, 64)
			}
		}
		
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+1, err)
		}
	}

	return nil
}

// DataInfo provides information about a dataset
type DataInfo struct {
	Rows        int
	Columns     int
	Headers     []string
	HasMissing  bool
	MissingCount int
	DataTypes   []string // inferred types per column
}

// InspectCSV provides information about a CSV file without loading all data
func InspectCSV(filename string, options CSVOptions) (*DataInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = options.Delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	info := &DataInfo{
		Rows:     0,
		Columns:  0,
		Headers:  nil,
		DataTypes: nil,
	}

	skipCount := options.SkipRows
	nullMap := make(map[string]bool)
	for _, nv := range options.NullValues {
		nullMap[nv] = true
	}

	// Read a sample of rows to infer types
	sampleSize := 100
	sampleRows := [][]string{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}

		if skipCount > 0 {
			skipCount--
			continue
		}

		// Handle header
		if options.HasHeader && info.Headers == nil {
			info.Headers = make([]string, len(record))
			copy(info.Headers, record)
			info.Columns = len(record)
			continue
		}

		// Set columns if not set
		if info.Columns == 0 {
			info.Columns = len(record)
		}

		info.Rows++
		
		// Collect sample for type inference
		if len(sampleRows) < sampleSize {
			sampleCopy := make([]string, len(record))
			copy(sampleCopy, record)
			sampleRows = append(sampleRows, sampleCopy)
		}

		// Count missing values
		for _, val := range record {
			if nullMap[strings.TrimSpace(val)] {
				info.MissingCount++
				info.HasMissing = true
			}
		}
	}

	// Infer data types from sample
	if len(sampleRows) > 0 && info.Columns > 0 {
		info.DataTypes = inferDataTypes(sampleRows, nullMap)
	}

	return info, nil
}

// inferDataTypes attempts to determine the data type of each column
func inferDataTypes(samples [][]string, nullMap map[string]bool) []string {
	if len(samples) == 0 || len(samples[0]) == 0 {
		return nil
	}

	cols := len(samples[0])
	types := make([]string, cols)

	for col := 0; col < cols; col++ {
		isNumeric := true
		isInteger := true
		hasData := false

		for _, row := range samples {
			if col >= len(row) {
				continue
			}

			val := strings.TrimSpace(row[col])
			if nullMap[val] {
				continue
			}

			hasData = true

			// Try parsing as float
			if f, err := strconv.ParseFloat(val, 64); err != nil {
				isNumeric = false
				isInteger = false
				break
			} else {
				// Check if it's an integer
				if f != float64(int64(f)) {
					isInteger = false
				}
			}
		}

		if !hasData {
			types[col] = "empty"
		} else if isInteger && isNumeric {
			types[col] = "integer"
		} else if isNumeric {
			types[col] = "float"
		} else {
			types[col] = "string"
		}
	}

	return types
}

// StreamingReader provides memory-efficient reading of large CSV files
type StreamingReader struct {
	reader     *csv.Reader
	file       *os.File
	options    CSVOptions
	headers    []string
	nullMap    map[string]bool
	currentRow int
}

// NewStreamingReader creates a new streaming CSV reader
func NewStreamingReader(filename string, options CSVOptions) (*StreamingReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = options.Delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	sr := &StreamingReader{
		reader:  reader,
		file:    file,
		options: options,
		nullMap: make(map[string]bool),
	}

	for _, nv := range options.NullValues {
		sr.nullMap[nv] = true
	}

	// Skip initial rows and read header if needed
	for i := 0; i < options.SkipRows; i++ {
		if _, err := reader.Read(); err != nil {
			file.Close()
			return nil, fmt.Errorf("error skipping rows: %w", err)
		}
	}

	if options.HasHeader {
		record, err := reader.Read()
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("error reading header: %w", err)
		}
		sr.headers = make([]string, len(record))
		copy(sr.headers, record)
	}

	return sr, nil
}

// Next reads the next row from the CSV file
func (sr *StreamingReader) Next() ([]float64, error) {
	record, err := sr.reader.Read()
	if err != nil {
		return nil, err
	}

	sr.currentRow++
	return parseRow(record, sr.options.Columns, sr.nullMap)
}

// Headers returns the column headers if available
func (sr *StreamingReader) Headers() []string {
	return sr.headers
}

// Close closes the underlying file
func (sr *StreamingReader) Close() error {
	return sr.file.Close()
}