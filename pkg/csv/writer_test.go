// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestNewWriter(t *testing.T) {
	opts := DefaultOptions()
	writer := NewWriter(opts)

	if writer == nil {
		t.Fatal("expected non-nil writer")
		return
	}

	if writer.opts.Delimiter != opts.Delimiter {
		t.Errorf("expected delimiter %v, got %v", opts.Delimiter, writer.opts.Delimiter)
	}
}

func TestWriter_Write_NumericData(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    2,
		Columns: 3,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A,B,C") {
		t.Error("expected headers in output")
	}

	if !strings.Contains(output, "1,2,3") {
		t.Error("expected first row in output")
	}

	if !strings.Contains(output, "4,5,6") {
		t.Error("expected second row in output")
	}
}

func TestWriter_Write_WithRowNames(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
		Headers:  []string{"A", "B", "C"},
		RowNames: []string{"row1", "row2"},
		Rows:     2,
		Columns:  3,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "row1") {
		t.Error("expected row1 in output")
	}

	if !strings.Contains(output, "row2") {
		t.Error("expected row2 in output")
	}
}

func TestWriter_Write_MissingValues(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0},
			{4.0, 5.0, math.NaN()},
		},
		Headers:     []string{"A", "B", "C"},
		Rows:        2,
		Columns:     3,
		MissingMask: [][]bool{{false, true, false}, {false, false, true}},
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	opts.NullValues = []string{"NA"}
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NA") && !strings.Contains(output, "NaN") {
		t.Errorf("expected NA or NaN for missing values, got: %s", output)
	}
}

func TestWriter_Write_CustomNullValue(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, math.NaN(), 3.0},
		},
		Headers:     []string{"A", "B", "C"},
		Rows:        1,
		Columns:     3,
		MissingMask: [][]bool{{false, true, false}},
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	opts.NullValues = []string{"NULL"}
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NULL") {
		t.Error("expected NULL for missing values")
	}
}

func TestWriter_Write_Infinity(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{math.Inf(1), math.Inf(-1), 3.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    1,
		Columns: 3,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Inf") {
		t.Error("expected Inf in output")
	}

	if !strings.Contains(output, "-Inf") {
		t.Error("expected -Inf in output")
	}
}

func TestWriter_Write_EuropeanFormat(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.5, 2.3, 3.7},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    1,
		Columns: 3,
	}

	var buf bytes.Buffer
	opts := EuropeanOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1,5") {
		t.Error("expected comma decimal separator")
	}

	if !strings.Contains(output, ";") {
		t.Error("expected semicolon delimiter")
	}
}

func TestWriter_Write_Precision(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.23456789, 2.0},
		},
		Headers: []string{"A", "B"},
		Rows:    1,
		Columns: 2,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	opts.Precision = 2
	opts.FloatFormat = 'f'
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "1.23456789") {
		t.Error("expected value to be rounded to 2 decimals")
	}

	if !strings.Contains(output, "1.23") {
		t.Error("expected 1.23 in output")
	}
}

func TestWriter_Write_StringData(t *testing.T) {
	data := &Data{
		StringData: [][]string{
			{"Alice", "30", "NYC"},
			{"Bob", "25", "LA"},
		},
		Headers: []string{"Name", "Age", "City"},
		Rows:    2,
		Columns: 3,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Error("expected Alice in output")
	}

	if !strings.Contains(output, "Bob") {
		t.Error("expected Bob in output")
	}
}

func TestWriter_Write_StringDataWithRowNames(t *testing.T) {
	data := &Data{
		StringData: [][]string{
			{"a", "b"},
			{"c", "d"},
		},
		Headers:  []string{"Col1", "Col2"},
		RowNames: []string{"R1", "R2"},
		Rows:     2,
		Columns:  2,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "R1") {
		t.Error("expected R1 in output")
	}

	if !strings.Contains(output, "R2") {
		t.Error("expected R2 in output")
	}
}

func TestWriter_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_output.csv")

	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    2,
		Columns: 3,
	}

	opts := DefaultOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.WriteFile(testFile, data)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "A,B,C") {
		t.Error("expected headers in file")
	}

	if !strings.Contains(output, "1,2,3") {
		t.Error("expected data in file")
	}
}

func TestWriter_WriteMatrix(t *testing.T) {
	matrix := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}
	headers := []string{"X", "Y"}
	rowNames := []string{"A", "B"}

	var buf bytes.Buffer
	opts := DefaultOptions()
	writer := NewWriter(opts)

	err := writer.WriteMatrix(&buf, matrix, headers, rowNames)
	if err != nil {
		t.Fatalf("WriteMatrix() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X,Y") {
		t.Error("expected headers in output")
	}

	if !strings.Contains(output, "A") && !strings.Contains(output, "B") {
		t.Error("expected row names in output")
	}
}

func TestWriter_WriteMatrixFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "matrix_output.csv")

	matrix := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}
	headers := []string{"X", "Y"}
	rowNames := []string{"A", "B"}

	opts := DefaultOptions()
	writer := NewWriter(opts)

	err := writer.WriteMatrixFile(testFile, matrix, headers, rowNames)
	if err != nil {
		t.Fatalf("WriteMatrixFile() error = %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "X,Y") {
		t.Error("expected headers in file")
	}
}

func TestSaveFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "save_output.csv")

	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0},
		},
		Headers: []string{"A", "B"},
		Rows:    1,
		Columns: 2,
	}

	opts := DefaultOptions()
	opts.HasRowNames = false

	err := SaveFile(testFile, data, opts)
	if err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

func TestSave(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0},
		},
		Headers: []string{"A", "B"},
		Rows:    1,
		Columns: 2,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasRowNames = false

	err := Save(&buf, data, opts)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A,B") {
		t.Error("expected headers in output")
	}
}

func TestSaveMatrix(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "matrix_save.csv")

	matrix := types.Matrix{
		{1.0, 2.0},
		{3.0, 4.0},
	}
	headers := []string{"X", "Y"}
	rowNames := []string{"A", "B"}

	opts := DefaultOptions()

	err := SaveMatrix(testFile, matrix, headers, rowNames, opts)
	if err != nil {
		t.Fatalf("SaveMatrix() error = %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "X,Y") {
		t.Error("expected headers in file")
	}
}

func TestWriter_Write_TabDelimited(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0, 3.0},
		},
		Headers: []string{"A", "B", "C"},
		Rows:    1,
		Columns: 3,
	}

	var buf bytes.Buffer
	opts := TabDelimitedOptions()
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\t") {
		t.Error("expected tab delimiter in output")
	}
}

func TestWriter_Write_NoHeaders(t *testing.T) {
	data := &Data{
		Matrix: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
		Rows:    2,
		Columns: 2,
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.HasHeaders = false
	opts.HasRowNames = false
	writer := NewWriter(opts)

	err := writer.Write(&buf, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines without headers, got %d", len(lines))
	}
}
