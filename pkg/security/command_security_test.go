// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package security

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateCommand_AllowedCommands(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "open command",
			cmd:     "open",
			args:    []string{"-a", "GoPCA"},
			wantErr: false,
		},
		{
			name:    "pgrep command",
			cmd:     "pgrep",
			args:    []string{"-x", "gopca-desktop"},
			wantErr: false,
		},
		{
			name:    "disallowed command",
			cmd:     "rm",
			args:    []string{"-rf", "/"},
			wantErr: true,
		},
		{
			name:    "curl disallowed",
			cmd:     "curl",
			args:    []string{"http://example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommand_Arguments(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "open with allowed args",
			cmd:     "open",
			args:    []string{"-a", "GoPCA.app", "--args", "test.csv"},
			wantErr: false,
		},
		{
			name:    "open with file path",
			cmd:     "open",
			args:    []string{"-a", "GoPCA", "/path/to/file.csv"},
			wantErr: false,
		},
		{
			name:    "command injection attempt",
			cmd:     "open",
			args:    []string{"-a", "GoPCA; rm -rf /"},
			wantErr: true,
		},
		{
			name:    "pipe injection",
			cmd:     "open",
			args:    []string{"-a", "GoPCA | evil"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsAllowedAppPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "GoPCA in user dir",
			path: filepath.Join("/Users/test/Applications", "GoPCA"),
			want: true,
		},
		{
			name: "GoCSV in user dir",
			path: filepath.Join("/home/user/bin", "gocsv"),
			want: true,
		},
		{
			name: "system binary",
			path: "/bin/bash",
			want: false,
		},
		{
			name: "random app",
			path: "/Applications/Safari.app",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedAppPath(tt.path)
			if got != tt.want {
				t.Errorf("isAllowedAppPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "absolute path unix",
			arg:  "/path/to/file.csv",
			want: true,
		},
		{
			name: "relative path",
			arg:  "data/file.csv",
			want: true,
		},
		{
			name: "home path",
			arg:  "~/Documents/data.csv",
			want: true,
		},
		{
			name: "file with extension",
			arg:  "file.csv",
			want: true,
		},
		{
			name: "flag",
			arg:  "-a",
			want: false,
		},
		{
			name: "app name",
			arg:  "GoPCA",
			want: false,
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name string
			arg  string
			want bool
		}{
			name: "windows path",
			arg:  `C:\Users\test\file.csv`,
			want: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFilePath(tt.arg)
			if got != tt.want {
				t.Errorf("isFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAllowedValue(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "app name",
			arg:  "GoPCA",
			want: true,
		},
		{
			name: "app with exe",
			arg:  "GoCSV.exe",
			want: true,
		},
		{
			name: "app bundle",
			arg:  "GoPCA.app",
			want: true,
		},
		{
			name: "tasklist filter",
			arg:  "IMAGENAME eq GoPCA.exe",
			want: true,
		},
		{
			name: "random value",
			arg:  "somethingelse",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedValue(tt.arg)
			if got != tt.want {
				t.Errorf("isAllowedValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateArgumentSafety(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{
			name:    "safe argument",
			arg:     "test.csv",
			wantErr: false,
		},
		{
			name:    "safe path",
			arg:     "/home/user/data/file.csv",
			wantErr: false,
		},
		{
			name:    "semicolon injection",
			arg:     "file.csv; rm -rf /",
			wantErr: true,
		},
		{
			name:    "pipe injection",
			arg:     "file.csv | cat",
			wantErr: true,
		},
		{
			name:    "ampersand injection",
			arg:     "file.csv && evil",
			wantErr: true,
		},
		{
			name:    "command substitution",
			arg:     "$(evil)",
			wantErr: true,
		},
		{
			name:    "backtick substitution",
			arg:     "`evil`",
			wantErr: true,
		},
		{
			name:    "excessive length",
			arg:     strings.Repeat("a", 1025),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgumentSafety(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateArgumentSafety() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid command",
			cmdName: "open",
			args:    []string{"-a", "GoPCA"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			cmdName: "rm",
			args:    []string{"-rf", "/"},
			wantErr: true,
		},
		{
			name:    "command injection",
			cmdName: "open",
			args:    []string{"-a", "GoPCA; evil"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := SecureCommand(tt.cmdName, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && cmd == nil {
				t.Error("SecureCommand() returned nil cmd without error")
			}

			if cmd != nil {
				if len(cmd.Env) == 0 {
					t.Error("SecureCommand() created cmd with empty environment")
				}
			}
		})
	}
}

func TestGetSecurePath(t *testing.T) {
	path := getSecurePath()

	if path == "" {
		t.Error("getSecurePath() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(path, "System32") {
			t.Error("Windows path should contain System32")
		}
	case "darwin", "linux":
		if !strings.Contains(path, "/usr/bin") {
			t.Error("Unix path should contain /usr/bin")
		}
	}
}

func TestGetHomeDir(t *testing.T) {
	home := getHomeDir()

	if home == "" {
		t.Error("getHomeDir() returned empty string")
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(home, "Users") {
			t.Error("Windows home should contain Users")
		}
	}
}

func TestGetTempDir(t *testing.T) {
	temp := getTempDir()

	if temp == "" {
		t.Error("getTempDir() returned empty string")
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(temp, "Temp") {
			t.Error("Windows temp should contain Temp")
		}
	} else {
		if temp != "/tmp" {
			t.Errorf("Unix temp should be /tmp, got %s", temp)
		}
	}
}

func TestValidateCommand_WindowsExe(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	err := ValidateCommand("tasklist.exe", []string{"/FI", "IMAGENAME eq GoPCA.exe"})
	if err != nil {
		t.Errorf("ValidateCommand() error = %v for windows command", err)
	}
}

func TestValidateAppLocation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "user applications",
			path:    filepath.Join("/Users/test/Applications", "GoPCA"),
			wantErr: false,
		},
		{
			name:    "system directory",
			path:    "/usr/bin/gopca",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppLocation(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAppLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
