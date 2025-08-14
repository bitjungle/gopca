// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTempFileManager tests the temporary file manager functionality
func TestTempFileManager(t *testing.T) {
	manager := NewTempFileManager()
	defer manager.Stop()

	// Test file creation
	tempFile, err := manager.CreateTempFile("test", ".csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Check that file path is valid
	if !strings.Contains(tempFile, "test") || !strings.HasSuffix(tempFile, ".csv") {
		t.Errorf("Invalid temp file path: %s", tempFile)
	}

	// Create actual file
	if err := os.WriteFile(tempFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Verify file is tracked
	manager.mu.Lock()
	_, exists := manager.files[tempFile]
	manager.mu.Unlock()

	if !exists {
		t.Error("Temp file not tracked by manager")
	}

	// Clean up
	os.Remove(tempFile)
}

// TestJSONConsistency tests JSON field naming consistency
func TestJSONConsistency(t *testing.T) {
	type TestStruct struct {
		ID           int       `json:"id"`
		UserName     string    `json:"userName"`
		IsActive     bool      `json:"isActive"`
		CreatedAt    time.Time `json:"createdAt"`
		MissingMask  [][]bool  `json:"missingMask,omitempty"`
		ExplainedVar []float64 `json:"explainedVar"`
		SNVApplied   bool      `json:"snvApplied"`
	}

	testStruct := TestStruct{
		ID:       1,
		UserName: "test",
		IsActive: true,
	}

	results, err := CheckJSONConsistency(testStruct)
	if err != nil {
		t.Fatalf("Failed to check JSON consistency: %v", err)
	}

	// Check for inconsistencies
	for _, result := range results {
		if result.Inconsistent {
			t.Logf("Warning: Inconsistent JSON tag for field %s: got %s, expected %s",
				result.GoFieldName, result.JSONTag, result.TSFieldName)
		}
	}
}

// TestJSONMarshaling tests that JSON marshaling preserves data
func TestJSONMarshaling(t *testing.T) {
	type TestData struct {
		Values [][]float64            `json:"values"`
		Labels []string               `json:"labels"`
		Config map[string]interface{} `json:"config,omitempty"`
	}

	original := TestData{
		Values: [][]float64{{1.0, 2.0}, {3.0, 4.0}},
		Labels: []string{"A", "B"},
		Config: map[string]interface{}{
			"option1": true,
			"option2": 42,
		},
	}

	if err := ValidateJSONMarshaling(&original); err != nil {
		t.Errorf("JSON marshaling validation failed: %v", err)
	}
}

// TestAppLaunchingTimeout tests that app launching has proper timeout
func TestAppLaunchingTimeout(t *testing.T) {
	// This test verifies the timeout mechanism is in place
	// by checking the LaunchWithFile function structure

	// Create a non-existent app path
	fakePath := "/nonexistent/app"
	tempFile := filepath.Join(os.TempDir(), "test.csv")

	// Create temp file
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile)

	// Try to launch non-existent app (should fail quickly)
	start := time.Now()
	err := LaunchWithFile(fakePath, tempFile)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected error for non-existent app")
	}

	// Should fail quickly (within 3 seconds due to timeout)
	if elapsed > 3*time.Second {
		t.Errorf("Launch took too long: %v", elapsed)
	}
}

// TestCleanupGoPCATempFiles tests the cleanup of old temporary files
func TestCleanupGoPCATempFiles(t *testing.T) {
	tempDir := os.TempDir()

	// Create an old temp file
	oldFile := filepath.Join(tempDir, "gopca_test_old.csv")
	if err := os.WriteFile(oldFile, []byte("old data"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	// Modify the file time to make it old
	oldTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}

	// Create a new temp file
	newFile := filepath.Join(tempDir, "gopca_test_new.csv")
	if err := os.WriteFile(newFile, []byte("new data"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	defer os.Remove(newFile)

	// Run cleanup
	if err := CleanupGoPCATempFiles(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Old file should be removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been removed")
		os.Remove(oldFile) // Clean up if test fails
	}

	// New file should still exist
	if _, err := os.Stat(newFile); err != nil {
		t.Error("New file should still exist")
	}
}
