package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

func TestAnalyzeMixedColumns(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Create test CSV with mixed columns
	csvContent := `sample,var1,var2,var3,group,concentration#target
A1,1.2,2.3,3.4,control,10.5
A2,1.3,2.5,3.2,control,10.2
A3,1.1,2.4,3.5,control,11.0
B1,2.2,3.3,4.4,treatment,15.5
B2,2.3,3.5,4.2,treatment,15.2
B3,2.1,3.4,4.5,treatment,16.0`

	csvFile := filepath.Join(testDir, "mixed.csv")
	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run analyze command with eigencorrelations
	app := NewApp()
	err := app.Run([]string{
		"gopca-cli",
		"analyze",
		"-c", "2",
		"--eigencorrelations",
		"--group-column", "group",
		"-f", "json",
		"-o", testDir,
		"--verbose",
		csvFile,
	})

	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	// Check output file exists
	outputFile := filepath.Join(testDir, "mixed_pca.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}

	// Read and parse output
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	var output types.PCAOutputData
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}

	// Check preserved columns
	if output.PreservedColumns == nil {
		t.Fatal("preserved columns not included in output")
	}

	// Check categorical data
	if len(output.PreservedColumns.Categorical) != 1 {
		t.Fatalf("expected 1 categorical column, got %d", len(output.PreservedColumns.Categorical))
	}

	groupData, ok := output.PreservedColumns.Categorical["group"]
	if !ok {
		t.Fatal("group column not found in categorical data")
	}

	if len(groupData) != 6 {
		t.Fatalf("expected 6 group values, got %d", len(groupData))
	}

	// Check target data
	if len(output.PreservedColumns.NumericTarget) != 1 {
		t.Fatalf("expected 1 target column, got %d", len(output.PreservedColumns.NumericTarget))
	}

	targetData, ok := output.PreservedColumns.NumericTarget["concentration#target"]
	if !ok {
		t.Fatal("concentration#target column not found in target data")
	}

	if len(targetData) != 6 {
		t.Fatalf("expected 6 target values, got %d", len(targetData))
	}

	// Check eigencorrelations
	if output.Eigencorrelations == nil {
		t.Fatal("eigencorrelations not included in output")
	}

	if len(output.Eigencorrelations.Variables) < 2 {
		t.Fatalf("expected at least 2 variables in eigencorrelations, got %d",
			len(output.Eigencorrelations.Variables))
	}

	// Check that concentration#target is correlated with PC1
	concCorr, ok := output.Eigencorrelations.Correlations["concentration#target"]
	if !ok {
		t.Fatal("concentration#target not found in correlations")
	}

	// PC1 should have high correlation with concentration
	if concCorr[0] < 0.9 {
		t.Errorf("expected high correlation between PC1 and concentration, got %f", concCorr[0])
	}
}

func TestAnalyzeTargetColumnAutoDetection(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Create test CSV with auto-detectable target column
	csvContent := `sample,var1,var2,var3,pH #target
S1,1.2,2.3,3.4,7.5
S2,1.3,2.5,3.2,7.2
S3,1.1,2.4,3.5,7.8`

	csvFile := filepath.Join(testDir, "target.csv")
	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run analyze command without specifying target columns
	app := NewApp()
	err := app.Run([]string{
		"gopca-cli",
		"analyze",
		"-c", "2",
		"-f", "json",
		"-o", testDir,
		csvFile,
	})

	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	// Read and parse output
	outputFile := filepath.Join(testDir, "target_pca.json")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	var output types.PCAOutputData
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}

	// Check that target column was auto-detected and excluded
	if output.PreservedColumns == nil {
		t.Fatal("preserved columns not included in output")
	}

	targetData, ok := output.PreservedColumns.NumericTarget["pH #target"]
	if !ok {
		t.Fatal("pH #target column not auto-detected as target column")
	}

	if len(targetData) != 3 {
		t.Fatalf("expected 3 target values, got %d", len(targetData))
	}
}
