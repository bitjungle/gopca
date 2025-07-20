package io

import (
	"bytes"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bitjungle/complab/pkg/types"
)

// Test basic CSV loading
func TestLoadCSV(t *testing.T) {
	// Create a temporary CSV file
	content := `x,y,z
1.0,2.0,3.0
4.0,5.0,6.0
7.0,8.0,9.0`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	// Load the CSV
	data, headers, err := LoadCSV(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	// Check headers
	expectedHeaders := []string{"x", "y", "z"}
	if !reflect.DeepEqual(headers, expectedHeaders) {
		t.Errorf("Headers mismatch: got %v, expected %v", headers, expectedHeaders)
	}

	// Check data dimensions
	if len(data) != 3 || len(data[0]) != 3 {
		t.Errorf("Wrong dimensions: got %dx%d, expected 3x3", len(data), len(data[0]))
	}

	// Check values
	expected := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
	}

	for i := range expected {
		for j := range expected[i] {
			if data[i][j] != expected[i][j] {
				t.Errorf("Value mismatch at [%d,%d]: got %f, expected %f",
					i, j, data[i][j], expected[i][j])
			}
		}
	}
}

// Test CSV with different delimiter
func TestLoadCSVWithDelimiter(t *testing.T) {
	content := `x;y;z
1.0;2.0;3.0
4.0;5.0;6.0`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	options := DefaultCSVOptions()
	options.Delimiter = ';'

	data, headers, err := LoadCSV(tmpfile, options)
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	if len(headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(headers))
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(data))
	}
}

// Test CSV with missing values
func TestLoadCSVWithMissing(t *testing.T) {
	content := `a,b,c
1.0,2.0,3.0
4.0,NA,6.0
7.0,8.0,`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	data, _, err := LoadCSV(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	// Check that missing values are NaN
	if !math.IsNaN(data[1][1]) {
		t.Errorf("Expected NaN at [1,1], got %f", data[1][1])
	}

	if !math.IsNaN(data[2][2]) {
		t.Errorf("Expected NaN at [2,2], got %f", data[2][2])
	}
}

// Test column selection
func TestLoadCSVWithColumnSelection(t *testing.T) {
	content := `a,b,c,d,e
1,2,3,4,5
6,7,8,9,10`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	options := DefaultCSVOptions()
	options.Columns = []int{0, 2, 4} // Select columns a, c, e

	data, headers, err := LoadCSV(tmpfile, options)
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	// Check selected headers
	expectedHeaders := []string{"a", "c", "e"}
	if !reflect.DeepEqual(headers, expectedHeaders) {
		t.Errorf("Headers mismatch: got %v, expected %v", headers, expectedHeaders)
	}

	// Check data
	if len(data[0]) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(data[0]))
	}

	// Check values
	expected := types.Matrix{
		{1, 3, 5},
		{6, 8, 10},
	}

	for i := range expected {
		for j := range expected[i] {
			if data[i][j] != expected[i][j] {
				t.Errorf("Value mismatch at [%d,%d]: got %f, expected %f",
					i, j, data[i][j], expected[i][j])
			}
		}
	}
}

// Test skip rows functionality
func TestLoadCSVSkipRows(t *testing.T) {
	content := `# Comment line
# Another comment
x,y
1,2
3,4`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	options := DefaultCSVOptions()
	options.SkipRows = 2

	data, headers, err := LoadCSV(tmpfile, options)
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	if len(headers) != 2 || headers[0] != "x" {
		t.Errorf("Headers not read correctly after skipping rows")
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 data rows, got %d", len(data))
	}
}

// Test max rows limit
func TestLoadCSVMaxRows(t *testing.T) {
	content := `x,y
1,2
3,4
5,6
7,8
9,10`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	options := DefaultCSVOptions()
	options.MaxRows = 3

	data, _, err := LoadCSV(tmpfile, options)
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("Expected 3 rows (max limit), got %d", len(data))
	}
}

// Test saving CSV
func TestSaveCSV(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, math.NaN(), 6.0},
		{7.0, 8.0, 9.0},
	}
	headers := []string{"col1", "col2", "col3"}

	tmpfile := filepath.Join(t.TempDir(), "output.csv")

	err := SaveCSV(tmpfile, data, headers, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("SaveCSV failed: %v", err)
	}

	// Read it back
	loaded, loadedHeaders, err := LoadCSV(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("Failed to load saved CSV: %v", err)
	}

	// Check headers
	if !reflect.DeepEqual(loadedHeaders, headers) {
		t.Errorf("Headers mismatch after save/load")
	}

	// Check data (including NaN)
	for i := range data {
		for j := range data[i] {
			if math.IsNaN(data[i][j]) {
				if !math.IsNaN(loaded[i][j]) {
					t.Errorf("NaN not preserved at [%d,%d]", i, j)
				}
			} else if data[i][j] != loaded[i][j] {
				t.Errorf("Value mismatch at [%d,%d]: got %f, expected %f",
					i, j, loaded[i][j], data[i][j])
			}
		}
	}
}

// Test CSV inspection
func TestInspectCSV(t *testing.T) {
	content := `a,b,c
1,2.5,hello
3,4.0,world
5,NA,test`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	info, err := InspectCSV(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("InspectCSV failed: %v", err)
	}

	if info.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", info.Rows)
	}

	if info.Columns != 3 {
		t.Errorf("Expected 3 columns, got %d", info.Columns)
	}

	if !info.HasMissing {
		t.Error("Expected HasMissing to be true")
	}

	if info.MissingCount != 1 {
		t.Errorf("Expected 1 missing value, got %d", info.MissingCount)
	}

	// Check inferred types
	expectedTypes := []string{"integer", "float", "string"}
	if !reflect.DeepEqual(info.DataTypes, expectedTypes) {
		t.Errorf("Type inference mismatch: got %v, expected %v",
			info.DataTypes, expectedTypes)
	}
}

// Test streaming reader
func TestStreamingReader(t *testing.T) {
	content := `x,y,z
1,2,3
4,5,6
7,8,9`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	reader, err := NewStreamingReader(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("NewStreamingReader failed: %v", err)
	}
	defer func() { _ = reader.Close() }()

	// Check headers
	headers := reader.Headers()
	if len(headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(headers))
	}

	// Read rows one by one
	expected := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	for i, exp := range expected {
		row, err := reader.Next()
		if err != nil {
			t.Fatalf("Failed to read row %d: %v", i, err)
		}

		if !reflect.DeepEqual(row, exp) {
			t.Errorf("Row %d mismatch: got %v, expected %v", i, row, exp)
		}
	}

	// Should get EOF on next read
	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

// Test error handling
func TestLoadCSVErrors(t *testing.T) {
	// Test non-existent file
	_, _, err := LoadCSV("non_existent_file.csv", DefaultCSVOptions())
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test malformed CSV
	content := `a,b,c
1,2,3
4,5` // Missing value in second row

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	_, _, err = LoadCSV(tmpfile, DefaultCSVOptions())
	if err == nil {
		t.Error("Expected error for malformed CSV")
	}

	// Test invalid numeric values
	content2 := `a,b
1,2
three,4`

	tmpfile2 := createTempCSV(t, content2)
	defer func() { _ = os.Remove(tmpfile2) }()

	_, _, err = LoadCSV(tmpfile2, DefaultCSVOptions())
	if err == nil {
		t.Error("Expected error for non-numeric value")
	}
}

// Test special float values
func TestSpecialFloatValues(t *testing.T) {
	content := `a,b,c
1.0,inf,-inf
2.0,Infinity,-Infinity`

	tmpfile := createTempCSV(t, content)
	defer func() { _ = os.Remove(tmpfile) }()

	data, _, err := LoadCSV(tmpfile, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("LoadCSV failed: %v", err)
	}

	// Check infinity values
	if !math.IsInf(data[0][1], 1) {
		t.Errorf("Expected +Inf at [0,1], got %f", data[0][1])
	}

	if !math.IsInf(data[0][2], -1) {
		t.Errorf("Expected -Inf at [0,2], got %f", data[0][2])
	}

	if !math.IsInf(data[1][1], 1) {
		t.Errorf("Expected +Inf at [1,1], got %f", data[1][1])
	}
}

// Test WriteCSV to buffer
func TestWriteCSV(t *testing.T) {
	data := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}
	headers := []string{"x", "y"}

	var buf bytes.Buffer
	err := WriteCSV(&buf, data, headers, DefaultCSVOptions())
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	expected := "x,y\n1,2\n3,4\n"
	if buf.String() != expected {
		t.Errorf("CSV output mismatch:\ngot:\n%s\nexpected:\n%s",
			buf.String(), expected)
	}
}

// Helper function to create temporary CSV file
func createTempCSV(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name()
}

// Benchmark CSV loading
func BenchmarkLoadCSV(b *testing.B) {
	// Create a larger CSV for benchmarking
	var buf strings.Builder
	buf.WriteString("x,y,z,w\n")
	for i := 0; i < 1000; i++ {
		buf.WriteString("1.0,2.0,3.0,4.0\n")
	}

	tmpfile := createTempCSV(&testing.T{}, buf.String())
	defer func() { _ = os.Remove(tmpfile) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := LoadCSV(tmpfile, DefaultCSVOptions())
		if err != nil {
			b.Fatal(err)
		}
	}
}
