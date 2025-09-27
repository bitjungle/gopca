// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package security

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateInputPath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid file",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			path:    filepath.Join(tmpDir, "nonexistent.csv"),
			wantErr: true,
		},
		{
			name:    "directory not file",
			path:    tmpDir,
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	tmpDir := t.TempDir()
	validOutput := filepath.Join(tmpDir, "output.csv")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid output path",
			path:    validOutput,
			wantErr: false,
		},
		{
			name:    "nonexistent parent",
			path:    filepath.Join(tmpDir, "nonexistent", "output.csv"),
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

func TestValidateOutputPath_SystemDirectories(t *testing.T) {
	systemPath := "/etc/passwd"
	if runtime.GOOS == "windows" {
		systemPath = `C:\Windows\System32\test.csv`
	}

	err := ValidateOutputPath(systemPath)
	if err == nil {
		t.Error("ValidateOutputPath() should reject system directories")
	}
}

func TestCheckSystemDirectory(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "user directory",
			path:    filepath.Join("/home/user", "data.csv"),
			wantErr: false,
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name    string
			path    string
			wantErr bool
		}{
			name:    "windows system directory",
			path:    `C:\Windows\test.csv`,
			wantErr: true,
		})
	} else {
		tests = append(tests, struct {
			name    string
			path    string
			wantErr bool
		}{
			name:    "unix system directory",
			path:    "/etc/test.csv",
			wantErr: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSystemDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkSystemDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureTempFile_Created(t *testing.T) {
	file, err := SecureTempFile("test-*.csv")
	if err != nil {
		t.Fatalf("SecureTempFile() error = %v", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	if file.Name() == "" {
		t.Error("SecureTempFile() returned empty name")
	}

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}

	mode := info.Mode().Perm()
	if runtime.GOOS != "windows" && mode&0077 != 0 {
		t.Errorf("SecureTempFile() created file with insecure permissions: %o", mode)
	}
}

func TestJailPath_Validation(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		root    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path in jail",
			root:    tmpDir,
			path:    filepath.Join(tmpDir, "test.csv"),
			wantErr: false,
		},
		{
			name:    "relative path in jail",
			root:    tmpDir,
			path:    "test.csv",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := JailPath(tt.root, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("JailPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()
	realFile := filepath.Join(tmpDir, "real.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	if err := os.WriteFile(realFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.Symlink(realFile, linkFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	resolved, err := ResolveSymlinks(linkFile)
	if err != nil {
		t.Fatalf("ResolveSymlinks() error = %v", err)
	}

	expectedReal, _ := filepath.EvalSymlinks(realFile)
	if resolved != expectedReal {
		t.Errorf("ResolveSymlinks() = %s, want %s", resolved, expectedReal)
	}
}

func TestValidateBasicPath_DirectoryTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "normal path",
			path:    "/home/user/data.csv",
			wantErr: false,
		},
		{
			name:    "null byte",
			path:    "/home/user/test\x00.csv",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBasicPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBasicPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUnixPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid unix path",
			path:    "/home/user/data.csv",
			wantErr: false,
		},
		{
			name:    "absolute unix path",
			path:    "/var/data/file.txt",
			wantErr: false,
		},
		{
			name:    "relative unix path",
			path:    "data/file.csv",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUnixPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUnixPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
