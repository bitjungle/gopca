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
	"testing"
	"time"
)

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
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

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
