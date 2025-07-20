package cmd

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/bitjungle/complab/internal/io"
)

// detectAndLoadCSV automatically detects if the first column contains row names
// and loads the CSV accordingly
func detectAndLoadCSV(filename string) (data [][]float64, headers []string, hasRowNames bool, err error) {
	// Get default options
	opts := io.DefaultCSVOptions()
	
	// Quick check: try to read first line to see if first column is numeric
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, false, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.Comma = ','
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, false, err
	}
	
	// Read first data row
	firstRow, err := reader.Read()
	if err != nil {
		return nil, nil, false, err
	}
	
	// Check if first column is numeric
	_, err = strconv.ParseFloat(firstRow[0], 64)
	if err != nil {
		// First column is not numeric, skip it
		hasRowNames = true
		numCols := len(header) - 1
		opts.Columns = make([]int, numCols)
		for i := 0; i < numCols; i++ {
			opts.Columns[i] = i + 1
		}
	}
	
	// Reset file position
	file.Seek(0, 0)
	
	// Load data with determined options
	data, headers, err = io.LoadCSV(filename, opts)
	return data, headers, hasRowNames, err
}