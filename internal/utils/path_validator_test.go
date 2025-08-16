// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	// Create a temp file for testing valid paths
	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()
	if err := tmpFile.Close(); err != nil {
		t.Logf("Warning: failed to close temp file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid existing file",
			path:    tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "/nonexistent/file.csv",
			wantErr: true,
		},
		{
			name:    "directory traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "hidden directory traversal",
			path:    "data/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "complex path with traversal",
			path:    "./data/../../../secret.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	// Create a temp directory for testing valid output paths
	tmpDir, err := os.MkdirTemp("", "test_output")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid output path in temp dir",
			path:    filepath.Join(tmpDir, "results.csv"),
			wantErr: false,
		},
		{
			name:    "system directory",
			path:    "/etc/important.conf",
			wantErr: true,
		},
		{
			name:    "bin directory",
			path:    "/bin/malicious",
			wantErr: true,
		},
		{
			name:    "proc directory",
			path:    "/proc/something",
			wantErr: true,
		},
		{
			name:    "non-existent parent directory",
			path:    "/nonexistent/dir/output.csv",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
