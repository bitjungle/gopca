// Package integration provides comprehensive integration tests for the GoPCA monorepo.
// These tests validate end-to-end workflows, cross-application communication,
// and ensure all components work together correctly.
package integration

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	TempDir      string
	CLIPath      string
	Timeout      time.Duration
	Verbose      bool
	SkipSlow     bool
	TestDataPath string
}

// NewTestConfig creates a test configuration
func NewTestConfig(t *testing.T) *TestConfig {
	tempDir := t.TempDir()

	// Build CLI if needed
	cliPath := filepath.Join(tempDir, "pca")
	if runtime.GOOS == "windows" {
		cliPath += ".exe"
	}

	return &TestConfig{
		TempDir:      tempDir,
		CLIPath:      cliPath,
		Timeout:      30 * time.Second,
		Verbose:      testing.Verbose(),
		SkipSlow:     testing.Short(),
		TestDataPath: filepath.Join("testdata"),
	}
}

// BuildCLI builds the pca CLI for testing
func (tc *TestConfig) BuildCLI(t *testing.T) {
	t.Helper()

	// First, check if there's a pre-built CLI (from make build)
	// This is the case in CI where 'make build' runs before tests
	prebuiltPath := filepath.Join("build", "pca")
	if runtime.GOOS == "windows" {
		prebuiltPath += ".exe"
	}

	// Try to find the prebuilt CLI by checking various possible locations
	possiblePaths := []string{
		prebuiltPath,                                        // build/pca (if in project root)
		filepath.Join("..", "..", prebuiltPath),             // ../../build/pca (if in internal/integration)
		filepath.Join("..", "..", "..", "..", prebuiltPath), // ../../../../build/pca (if deeper in test)
	}

	var sourceFile string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			sourceFile = path
			break
		}
	}

	if sourceFile != "" {
		// Found pre-built CLI, copy it to test directory
		t.Logf("Using pre-built CLI from %s", sourceFile)
		input, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read pre-built CLI: %v", err)
		}

		err = os.WriteFile(tc.CLIPath, input, 0755)
		if err != nil {
			t.Fatalf("Failed to copy CLI to test directory: %v", err)
		}
		return
	}

	// No pre-built CLI found, build it from source
	// Find project root by looking for go.mod
	// Start from current working directory, not from temp directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root (go.mod) starting from %s", cwd)
		}
		projectRoot = parent
	}

	cliSource := filepath.Join(projectRoot, "cmd", "gopca-cli")
	cmd := exec.Command("go", "build", "-o", tc.CLIPath, cliSource)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI from %s: %v\nOutput: %s", cliSource, err, output)
	}
	t.Logf("Built CLI from source at %s", cliSource)
}

// RunCLI executes the CLI with given arguments
func (tc *TestConfig) RunCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(tc.CLIPath, args...)
	cmd.Dir = tc.TempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set timeout
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("CLI error: %v\nStderr: %s", err, stderr.String())
		}
		return stdout.String(), nil
	case <-time.After(tc.Timeout):
		if err := cmd.Process.Kill(); err != nil {
			// Process may have already exited, ignore error
			_ = err
		}
		return "", fmt.Errorf("CLI timeout after %v", tc.Timeout)
	}
}

// CreateTestCSV creates a CSV file with test data
func (tc *TestConfig) CreateTestCSV(t *testing.T, name string, data [][]string) string {
	t.Helper()

	path := filepath.Join(tc.TempDir, name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Logf("Warning: failed to close file: %v", err)
		}
	}()

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(data); err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}

	return path
}

// LoadJSONResult loads and parses a JSON result file
func (tc *TestConfig) LoadJSONResult(t *testing.T, path string) map[string]interface{} {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	return result
}

// CompareMatrices compares two matrices within tolerance
func CompareMatrices(a, b [][]float64, tolerance float64) error {
	if len(a) != len(b) {
		return fmt.Errorf("matrix dimension mismatch: %d vs %d rows", len(a), len(b))
	}

	for i := range a {
		if len(a[i]) != len(b[i]) {
			return fmt.Errorf("matrix dimension mismatch at row %d: %d vs %d cols",
				i, len(a[i]), len(b[i]))
		}

		for j := range a[i] {
			diff := math.Abs(a[i][j] - b[i][j])
			if diff > tolerance {
				return fmt.Errorf("value mismatch at [%d,%d]: %f vs %f (diff: %f)",
					i, j, a[i][j], b[i][j], diff)
			}
		}
	}

	return nil
}

// CompareVectors compares two vectors within tolerance
func CompareVectors(a, b []float64, tolerance float64) error {
	if len(a) != len(b) {
		return fmt.Errorf("vector dimension mismatch: %d vs %d", len(a), len(b))
	}

	for i := range a {
		diff := math.Abs(a[i] - b[i])
		if diff > tolerance {
			return fmt.Errorf("value mismatch at [%d]: %f vs %f (diff: %f)",
				i, a[i], b[i], diff)
		}
	}

	return nil
}

// GenerateTestMatrix creates a deterministic test matrix
func GenerateTestMatrix(rows, cols int, seed float64) [][]string {
	data := make([][]string, rows+1)

	// Header row
	data[0] = make([]string, cols+1)
	data[0][0] = "Sample"
	for j := 1; j <= cols; j++ {
		data[0][j] = fmt.Sprintf("Feature%d", j)
	}

	// Data rows
	for i := 1; i <= rows; i++ {
		data[i] = make([]string, cols+1)
		data[i][0] = fmt.Sprintf("S%d", i)
		for j := 1; j <= cols; j++ {
			// Generate deterministic values
			value := seed * float64(i) * float64(j) / float64(rows*cols)
			value = math.Sin(value)*10 + float64(i+j)
			data[i][j] = fmt.Sprintf("%.6f", value)
		}
	}

	return data
}

// CheckFileExists verifies a file exists and is non-empty
func CheckFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("File does not exist: %s", path)
		return
	}

	if info.Size() == 0 {
		t.Errorf("File is empty: %s", path)
	}
}

// ParseCSVOutput parses CSV output from CLI
func ParseCSVOutput(output string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(output))
	return reader.ReadAll()
}

// ExtractJSONFromOutput extracts JSON from mixed CLI output
func ExtractJSONFromOutput(output string) (map[string]interface{}, error) {
	// Find JSON start and end
	start := strings.Index(output, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON found in output")
	}

	// Find matching closing brace
	depth := 0
	end := -1
outer:
	for i := start; i < len(output); i++ {
		switch output[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break outer
			}
		}
	}

	if end == -1 {
		return nil, fmt.Errorf("incomplete JSON in output")
	}

	jsonStr := output[start:end]
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}

// SkipIfShort skips test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// SkipOnPlatform skips test on specific platforms
func SkipOnPlatform(t *testing.T, platforms ...string) {
	for _, platform := range platforms {
		if runtime.GOOS == platform {
			t.Skipf("Skipping test on %s", platform)
		}
	}
}

// RequirePlatform skips test unless on specific platform
func RequirePlatform(t *testing.T, platform string) {
	if runtime.GOOS != platform {
		t.Skipf("Test requires %s, running on %s", platform, runtime.GOOS)
	}
}

// CreateSampleDataset creates a standard test dataset
type SampleDataset struct {
	Name       string
	Path       string
	Rows       int
	Cols       int
	HasMissing bool
	HasGroups  bool
	HasTargets bool
}

func (tc *TestConfig) CreateSampleDatasets(t *testing.T) map[string]*SampleDataset {
	t.Helper()

	datasets := make(map[string]*SampleDataset)

	// Small complete dataset
	smallData := GenerateTestMatrix(10, 5, 1.0)
	smallPath := tc.CreateTestCSV(t, "small.csv", smallData)
	datasets["small"] = &SampleDataset{
		Name: "small", Path: smallPath, Rows: 10, Cols: 5,
	}

	// Medium dataset with groups
	mediumData := GenerateTestMatrix(50, 10, 2.0)
	// Add group column
	mediumData[0] = append(mediumData[0], "Group")
	for i := 1; i < len(mediumData); i++ {
		group := "A"
		if i > len(mediumData)/2 {
			group = "B"
		}
		mediumData[i] = append(mediumData[i], group)
	}
	mediumPath := tc.CreateTestCSV(t, "medium.csv", mediumData)
	datasets["medium"] = &SampleDataset{
		Name: "medium", Path: mediumPath, Rows: 50, Cols: 10, HasGroups: true,
	}

	// Dataset with missing values
	missingData := GenerateTestMatrix(20, 8, 3.0)
	// Introduce missing values
	missingData[5][3] = ""
	missingData[10][5] = "NA"
	missingData[15][7] = "NaN"
	missingPath := tc.CreateTestCSV(t, "missing.csv", missingData)
	datasets["missing"] = &SampleDataset{
		Name: "missing", Path: missingPath, Rows: 20, Cols: 8, HasMissing: true,
	}

	return datasets
}

// AssertNoError fails test if error is not nil
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertError fails test if error is nil
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", msg)
	}
}

// AssertContains checks if string contains substring
func AssertContains(t *testing.T, str, substr, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("%s: string does not contain expected substring\nString: %s\nExpected: %s",
			msg, str, substr)
	}
}

// Benchmark helpers
type BenchmarkResult struct {
	Duration time.Duration
	MemUsed  uint64
	CPUUsage float64
}

func (tc *TestConfig) BenchmarkCLI(t *testing.T, args ...string) *BenchmarkResult {
	t.Helper()

	start := time.Now()
	_, err := tc.RunCLI(t, args...)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// Parse memory usage if available in output
	var memUsed uint64
	// Memory parsing is not yet implemented

	return &BenchmarkResult{
		Duration: duration,
		MemUsed:  memUsed,
	}
}
