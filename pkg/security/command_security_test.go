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
			path:    userAppPath(),
			wantErr: false,
		},
		{
			// The path has to be one the running platform actually treats as a
			// system directory. This case used to hardcode /usr/bin/gopca, which
			// is right on Unix and meaningless on Windows: filepath.Abs turns it
			// into something like D:\usr\bin\gopca, which matches none of
			// WindowsSystemDirectories, so validateAppLocation correctly returned
			// nil and the test failed.
			//
			// It went unnoticed because Windows CI did not run pkg/security at
			// all until the package list was derived rather than hand-written.
			name:    "system directory",
			path:    systemAppPath(),
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

// userAppPath returns a location the platform does not consider a system
// directory, for the case that must be accepted.
func userAppPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(`C:\Users\test\AppData\Local`, "GoPCA")
	}
	return filepath.Join("/Users/test/Applications", "GoPCA")
}

// systemAppPath returns a location the platform does consider a system
// directory, for the case that must be rejected. Both values come from the
// lists validateAppLocation itself consults, so the test cannot drift from the
// implementation by naming a directory that has since been removed.
func systemAppPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(WindowsSystemDirectories[0], "gopca.exe")
	}
	return filepath.Join(SystemDirectories[0], "gopca")
}
