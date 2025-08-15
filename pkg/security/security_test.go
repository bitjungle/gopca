// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateNumericInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		min       float64
		max       float64
		wantValue float64
		wantErr   bool
	}{
		{"valid integer", "42", 0, 100, 42, false},
		{"valid float", "3.14", 0, 10, 3.14, false},
		{"valid negative", "-5.5", -10, 10, -5.5, false},
		{"valid scientific", "1.5e2", 0, 200, 150, false},
		{"empty input", "", 0, 100, 0, true},
		{"out of range high", "150", 0, 100, 0, true},
		{"out of range low", "-5", 0, 100, 0, true},
		{"invalid characters", "12abc", 0, 100, 0, true},
		{"SQL injection attempt", "1; DROP TABLE", 0, 100, 0, true},
		{"NaN", "NaN", 0, 100, 0, true},
		{"Infinity", "Inf", 0, 100, 0, true},
		{"multiple dots", "1.2.3", 0, 100, 0, true},
		{"gamma validation", "0.001", MinKernelGamma, MaxKernelGamma, 0.001, false},
		{"decimal less than 1", "0.5", 0, 1, 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateNumericInput(tt.input, tt.min, tt.max, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumericInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("ValidateNumericInput() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestValidateIntegerInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		min     int
		max     int
		want    int
		wantErr bool
	}{
		{"valid positive", "42", 0, 100, 42, false},
		{"valid negative", "-5", -10, 10, -5, false},
		{"valid with plus", "+25", 0, 50, 25, false},
		{"empty input", "", 0, 100, 0, true},
		{"float input", "3.14", 0, 100, 0, true},
		{"out of range", "150", 0, 100, 0, true},
		{"invalid characters", "12abc", 0, 100, 0, true},
		{"components validation", "5", MinComponents, MaxComponents, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateIntegerInput(tt.input, tt.min, tt.max, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntegerInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateIntegerInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateStringInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		maxLength    int
		allowedChars string
		wantErr      bool
	}{
		{"valid string", "hello world", 100, "", false},
		{"empty allowed", "", 100, "", false},
		{"max length ok", strings.Repeat("a", 100), 100, "", false},
		{"too long", strings.Repeat("a", 101), 100, "", true},
		{"null bytes removed", "hello\x00world", 100, "", false},
		{"control chars removed", "hello\x01\x02world", 100, "", false},
		{"allowed chars only", "abc123", 10, "abc123", false},
		{"disallowed chars", "abc$", 10, "abc123", true},
		{"unicode valid", "Hello 世界", 100, "", false},
		{"invalid UTF-8", string([]byte{0xff, 0xfe}), 100, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateStringInput(tt.input, tt.maxLength, tt.allowedChars, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateComponentCount(t *testing.T) {
	tests := []struct {
		name        string
		components  int
		maxFeatures int
		wantErr     bool
	}{
		{"valid count", 5, 10, false},
		{"equal to features", 10, 10, false},
		{"too few", 0, 10, true},
		{"too many for features", 15, 10, true},
		{"exceeds max limit", MaxComponents + 1, MaxComponents + 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateComponentCount(tt.components, tt.maxFeatures)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateComponentCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateKernelParameters(t *testing.T) {
	tests := []struct {
		name       string
		kernelType string
		gamma      float64
		degree     float64
		coef0      float64
		wantErr    bool
	}{
		{"valid rbf", "rbf", 0.01, 0, 0, false},
		{"valid polynomial", "polynomial", 0.1, 3, 1, false},
		{"valid sigmoid", "sigmoid", 0.01, 0, 0.5, false},
		{"valid linear", "linear", 0, 0, 0, false},
		{"invalid kernel", "invalid", 0, 0, 0, true},
		{"gamma too small", "rbf", 1e-7, 0, 0, true},
		{"gamma too large", "rbf", 1e7, 0, 0, true},
		{"degree out of range", "polynomial", 0.1, 11, 0, true},
		{"coef0 out of range", "sigmoid", 0.1, 0, 1001, true},
		{"decimal gamma", "rbf", 0.001, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKernelParameters(tt.kernelType, tt.gamma, tt.degree, tt.coef0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKernelParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDataDimensions(t *testing.T) {
	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"valid small", 100, 50, false},
		{"valid large", 10000, 1000, false},
		{"zero rows", 0, 10, true},
		{"zero cols", 10, 0, true},
		{"negative rows", -1, 10, true},
		{"too many rows", MaxCSVRows + 1, 10, true},
		{"too many cols", 10, MaxCSVColumns + 1, true},
		{"memory limit exceeded", 100000, 10000, true}, // Would exceed 2GB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDataDimensions(tt.rows, tt.cols)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataDimensions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal filename", "data.csv", "data.csv"},
		{"path separators", "../../etc/passwd", "____etc_passwd"},
		{"special chars", "file<>:|?.txt", "file_____.txt"},
		{"hidden file", ".hidden", "hidden"},
		{"empty after sanitize", "...", "_."},
		{"very long name", strings.Repeat("a", 300), strings.Repeat("a", 255)},
		{"shell command", "file;rm -rf /", "file_rm -rf _"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateCSVDelimiter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    rune
		wantErr bool
	}{
		{"comma", ",", ',', false},
		{"semicolon", ";", ';', false},
		{"tab", "\t", '\t', false},
		{"pipe", "|", '|', false},
		{"space", " ", ' ', false},
		{"invalid char", "#", 0, true},
		{"multiple chars", ",,", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCSVDelimiter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCSVDelimiter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateCSVDelimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"normal path", "data/file.csv", false},
		{"absolute path", "/home/user/data.csv", false},
		{"parent directory", "../data.csv", true},
		{"nested traversal", "data/../../etc/passwd", true},
		{"hidden traversal", "data/../../passwd", true},
		{"null byte", "file\x00.csv", true},
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

func TestJailPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		userPath string
		wantErr  bool
	}{
		{"normal file", "/data", "file.csv", false},
		{"subdirectory", "/data", "sub/file.csv", false},
		{"escape attempt", "/data", "../etc/passwd", true},
		{"absolute escape", "/data", "/etc/passwd", false}, // absolute paths within jail are allowed
		{"complex escape", "/data", "sub/../../etc/passwd", true},
		{"stay in jail", "/data", "sub/../file.csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := JailPath(tt.basePath, tt.userPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("JailPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWindowsPath(t *testing.T) {
	// Don't skip on non-Windows - we want to test the validation logic
	// on all platforms to ensure CI works correctly

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Valid Windows paths
		{"drive letter path", `C:\Users\data.csv`, false},
		{"different drive", `D:\temp\file.txt`, false},
		{"lowercase drive letter", `c:\windows\temp.csv`, false},
		{"multiple subdirectories", `C:\Users\test\Documents\data.csv`, false},
		{"temp directory path", `C:\Users\RUNNER~1\AppData\Local\Temp\test.csv`, false},
		{"path with numbers", `C:\folder1\folder2\file3.csv`, false},

		// Invalid paths with colons in wrong places
		{"colon in filename", `C:\test\file:name.csv`, true},
		{"multiple colons", `C:\test:data\file.csv`, true},
		{"colon at end", `C:\test\file:`, true},
		{"colon without drive letter", `:test\file.csv`, true},

		// Reserved names
		{"reserved name CON", `C:\data\CON.txt`, true},
		{"reserved name PRN", `PRN`, true},
		{"reserved COM1", `COM1.txt`, true},
		{"reserved LPT1", `C:\test\LPT1`, true},

		// Invalid characters
		{"pipe character", `C:\test|file.txt`, true},
		{"question mark", `C:\test?file.txt`, true},
		{"asterisk", `C:\test*file.txt`, true},
		{"less than", `C:\test<file.txt`, true},
		{"greater than", `C:\test>file.txt`, true},
		{"quotes", `C:\test"file.txt`, true},

		// Trailing dots and spaces
		{"trailing dot", `C:\test\file.`, true},
		{"trailing space", `C:\test\file `, true},
		{"folder trailing dot", `C:\test.\file.csv`, true},
		{"folder trailing space", `C:\test \file.csv`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWindowsPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWindowsPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureTempFile(t *testing.T) {
	f, err := SecureTempFile("test")
	if err != nil {
		t.Fatalf("SecureTempFile() error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Check file exists
	if _, err := os.Stat(f.Name()); err != nil {
		t.Errorf("temp file should exist: %v", err)
	}

	// Check file starts with expected prefix
	base := filepath.Base(f.Name())
	if !strings.HasPrefix(base, "gopca_test") {
		t.Errorf("temp file name should start with 'gopca_test', got %s", base)
	}

	// Check permissions on Unix
	if runtime.GOOS != "windows" {
		info, _ := f.Stat()
		mode := info.Mode()
		if mode.Perm() != 0600 {
			t.Errorf("temp file should have 0600 permissions, got %v", mode.Perm())
		}
	}
}
