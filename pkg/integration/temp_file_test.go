// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTempFileManager_CreateAndStop(t *testing.T) {
	m := NewTempFileManager()
	if m == nil {
		t.Fatal("NewTempFileManager returned nil")
	}
	m.Stop()
}

func TestTempFileManager_CreateTempFile(t *testing.T) {
	m := NewTempFileManager()
	defer m.Stop()

	path, err := m.CreateTempFile("test", ".csv")
	if err != nil {
		t.Fatalf("CreateTempFile: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	if !strings.HasSuffix(path, ".csv") {
		t.Errorf("expected .csv suffix, got %q", path)
	}
	if !strings.Contains(filepath.Base(path), "test_") {
		t.Errorf("expected prefix 'test_' in filename, got %q", filepath.Base(path))
	}
}

func TestTempFileManager_RegisterAndCleanupOldFiles(t *testing.T) {
	m := NewTempFileManager()
	defer m.Stop()

	// Create a real temp file on disk
	tmpFile, err := os.CreateTemp("", "gopca_test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	tmpPath := tmpFile.Name()

	// Register it with a creation time far in the past to force cleanup
	m.mu.Lock()
	m.files[tmpPath] = time.Now().Add(-48 * time.Hour) // 48 hours old
	m.maxAge = 1 * time.Hour                           // maxAge = 1 hour → file is expired
	m.mu.Unlock()

	m.CleanupOldFiles()

	// File should be removed
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		_ = os.Remove(tmpPath) // best-effort cleanup if test failed
		t.Error("expected temp file to be removed after CleanupOldFiles")
	}

	// Entry should also be removed from tracking map
	m.mu.Lock()
	_, stillTracked := m.files[tmpPath]
	m.mu.Unlock()
	if stillTracked {
		t.Error("expected file to be removed from tracking map")
	}
}

func TestTempFileManager_CleanupSkipsNewFiles(t *testing.T) {
	m := NewTempFileManager()
	defer m.Stop()

	// Create a real temp file on disk
	tmpFile, err := os.CreateTemp("", "gopca_test_new_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Register it as just created (should NOT be cleaned up)
	m.RegisterTempFile(tmpPath)

	m.CleanupOldFiles() // maxAge is 24 hours, file is < 1 second old

	// File should still exist
	if _, err := os.Stat(tmpPath); err != nil {
		t.Error("expected recently-registered file to survive CleanupOldFiles")
	}
}

func TestCleanupGoPCATempFiles_DoesNotError(t *testing.T) {
	// Just ensure the function runs without panicking or returning an error
	// even when there are no GoPCA temp files present.
	if err := CleanupGoPCATempFiles(); err != nil {
		t.Errorf("CleanupGoPCATempFiles returned unexpected error: %v", err)
	}
}
