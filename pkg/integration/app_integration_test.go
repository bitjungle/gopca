// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetCommonPaths(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		wantMin int // Minimum expected number of paths
	}{
		{
			name:    "GoPCA Desktop paths",
			appName: "gopca-desktop",
			wantMin: 2, // At least 2 paths per platform
		},
		{
			name:    "GoCSV paths",
			appName: "gocsv",
			wantMin: 2, // At least 2 paths per platform
		},
		{
			name:    "Unknown app",
			appName: "unknown-app",
			wantMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := GetCommonPaths(tt.appName)
			if len(paths) < tt.wantMin {
				t.Errorf("GetCommonPaths(%q) returned %d paths, want at least %d", tt.appName, len(paths), tt.wantMin)
			}

			// Check that paths are appropriate for the current OS
			for _, path := range paths {
				if runtime.GOOS == "windows" {
					if filepath.Separator != '\\' && len(path) > 0 && path[0] != 'C' {
						t.Errorf("Invalid Windows path: %s", path)
					}
				} else {
					if len(path) > 0 && path[0] != '/' && !filepath.IsAbs(path) {
						t.Errorf("Invalid Unix path: %s", path)
					}
				}
			}
		})
	}
}

func TestCheckApp(t *testing.T) {
	// Create a temporary executable for testing
	tempDir := t.TempDir()
	tempExe := filepath.Join(tempDir, "test-app")
	if runtime.GOOS == "windows" {
		tempExe += ".exe"
	}

	// Create a dummy executable file
	if err := os.WriteFile(tempExe, []byte("dummy"), 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	tests := []struct {
		name          string
		config        AppConfig
		wantInstalled bool
	}{
		{
			name: "App found in custom path",
			config: AppConfig{
				Name:        "test-app",
				CommonPaths: []string{tempExe},
				DisplayName: "Test App",
			},
			wantInstalled: true,
		},
		{
			name: "App not found",
			config: AppConfig{
				Name:        "nonexistent-app",
				CommonPaths: []string{"/path/that/does/not/exist"},
				DisplayName: "Nonexistent App",
			},
			wantInstalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := CheckApp(tt.config)
			if status.Installed != tt.wantInstalled {
				t.Errorf("CheckApp() installed = %v, want %v", status.Installed, tt.wantInstalled)
			}

			if tt.wantInstalled && status.Path == "" {
				t.Error("CheckApp() found app but Path is empty")
			}

			if !tt.wantInstalled && status.Error == "" {
				t.Error("CheckApp() app not found but Error is empty")
			}
		})
	}
}

func TestIsAppRunning_NonexistentApp(t *testing.T) {
	// An app with this name can't possibly be running.
	if IsAppRunning("definitely-not-a-real-app-xyz-gopca-test") {
		t.Error("expected false for a non-existent app name")
	}
}

func TestGetCompanionAppPaths_ReturnsSlice(t *testing.T) {
	// Use the current test binary's path as a stand-in for "current exe".
	// The function just does path arithmetic — we verify it returns a non-nil slice
	// without panicking, regardless of whether the paths exist.
	fakeExePath := filepath.Join(t.TempDir(), "GoPCA.app", "Contents", "MacOS", "GoPCA")

	for _, target := range []string{"gocsv", "gopca-desktop"} {
		paths := getCompanionAppPaths(fakeExePath, target)
		if paths == nil {
			t.Errorf("getCompanionAppPaths(%q, %q) returned nil", fakeExePath, target)
		}
	}
}

func TestGetCompanionAppPaths_UnknownTarget(t *testing.T) {
	fakeExePath := filepath.Join(t.TempDir(), "SomeApp")
	paths := getCompanionAppPaths(fakeExePath, "unknown-app")
	// Should return an empty (but non-nil) slice for an unknown target app.
	if paths == nil {
		t.Error("expected non-nil slice for unknown target app")
	}
}

func TestLaunchWithFile(t *testing.T) {
	// This test is limited because we can't easily test launching real applications
	// We'll test error cases only

	tests := []struct {
		name     string
		appPath  string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Nonexistent app",
			appPath:  "/path/to/nonexistent/app",
			filePath: "test.txt",
			wantErr:  true,
		},
		{
			name:     "Nonexistent file",
			appPath:  "/usr/bin/ls", // Assuming ls exists on Unix systems
			filePath: "/path/to/nonexistent/file.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LaunchWithFile(tt.appPath, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LaunchWithFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
