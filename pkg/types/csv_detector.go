package types

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"
	"strings"
)

// columnTypeDetectionSampleSize defines how many rows to check when detecting column types
const columnTypeDetectionSampleSize = 10

// isNumericValue checks if a string value represents a numeric value
func isNumericValue(value string, format CSVFormat) (bool, float64) {
	// Check if it's in null values list
	for _, nv := range format.NullValues {
		if value == nv {
			return true, math.NaN() // Null values are considered numeric (as NaN)
		}
	}

	// Handle decimal separator if needed
	testValue := value
	if format.DecimalSeparator == ',' {
		testValue = strings.ReplaceAll(value, ",", ".")
	}

	// Try to parse as float
	val, err := strconv.ParseFloat(testValue, 64)
	if err == nil {
		return true, val
	}

	// Check special cases
	switch strings.ToLower(value) {
	case "inf", "+inf", "infinity":
		return true, math.Inf(1)
	case "-inf", "-infinity":
		return true, math.Inf(-1)
	}

	return false, 0
}

// DetectColumnTypes reads a CSV file and determines which columns are numeric vs categorical
func DetectColumnTypes(r io.Reader, format CSVFormat) (numericCols []int, categoricalCols []int, headers []string, err error) {
	// Initialize return slices
	numericCols = make([]int, 0)
	categoricalCols = make([]int, 0)

	csvReader := csv.NewReader(r)
	csvReader.Comma = format.FieldDelimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, nil, err
	}

	if len(records) < 2 {
		return nil, nil, nil, nil
	}

	// Handle headers
	startRow := 0
	if format.HasHeaders {
		headers = records[0]
		if format.HasRowNames && len(headers) > 0 {
			headers = headers[1:] // Skip row name header
		}
		startRow = 1
	}

	// Determine column count
	startCol := 0
	if format.HasRowNames {
		startCol = 1
	}

	numCols := len(records[0]) - startCol
	if numCols <= 0 {
		return nil, nil, headers, nil
	}

	// For each column, check if it's numeric or categorical
	for j := 0; j < numCols; j++ {
		isNumeric := true
		hasAnyValue := false

		// Check first N rows to determine type
		for i := startRow; i < len(records) && i < startRow+columnTypeDetectionSampleSize; i++ {
			if j+startCol >= len(records[i]) {
				continue
			}

			value := strings.TrimSpace(records[i][j+startCol])
			if value == "" {
				continue
			}

			hasAnyValue = true

			// Check if the value is numeric
			isNum, _ := isNumericValue(value, format)
			if !isNum {
				// Not a number - this is categorical
				isNumeric = false
				break
			}
		}

		if !hasAnyValue {
			// Empty column - treat as numeric
			numericCols = append(numericCols, j)
		} else if isNumeric {
			numericCols = append(numericCols, j)
		} else {
			categoricalCols = append(categoricalCols, j)
		}
	}

	return numericCols, categoricalCols, headers, nil
}
