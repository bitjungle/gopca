// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package security

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AllowedCommands defines the whitelist of commands that can be executed
var AllowedCommands = map[string]bool{
	// System utilities (safe, read-only operations)
	"open":     true, // macOS
	"pgrep":    true, // macOS/Linux process checking
	"tasklist": true, // Windows process checking

	// Application binaries (full paths will be validated)
	"GoPCA":         true,
	"GoCSV":         true,
	"gopca-desktop": true,
	"gocsv":         true,
}

// AllowedArguments defines safe arguments for specific commands
var AllowedArguments = map[string]map[string]bool{
	"open": {
		"-a":     true, // Application
		"-n":     true, // New instance
		"--args": true, // Arguments
		"--open": true, // Open file (our custom arg)
	},
	"tasklist": {
		"/FI": true, // Filter
	},
	"pgrep": {
		"-x": true, // Exact match
	},
}

// ValidateCommand validates a command and its arguments for security
func ValidateCommand(cmd string, args []string) error {
	// Get the base command name
	baseCmd := filepath.Base(cmd)

	// Remove .exe suffix on Windows for comparison
	if runtime.GOOS == "windows" && strings.HasSuffix(baseCmd, ".exe") {
		baseCmd = strings.TrimSuffix(baseCmd, ".exe")
	}

	// Check if command is in whitelist
	if !AllowedCommands[baseCmd] {
		// Check if it's a full path to an allowed app
		if !isAllowedAppPath(cmd) {
			return fmt.Errorf("command not allowed: %s", baseCmd)
		}
	}

	// Validate arguments for known commands
	if allowedArgs, ok := AllowedArguments[baseCmd]; ok {
		for _, arg := range args {
			// Skip file paths and values
			if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "/") {
				// Check if it's an allowed flag
				if !allowedArgs[arg] && !isFilePath(arg) && !isAllowedValue(arg) {
					return fmt.Errorf("argument not allowed for %s: %s", baseCmd, arg)
				}
			}
		}
	}

	// Check for command injection patterns
	for _, arg := range args {
		if err := validateArgumentSafety(arg); err != nil {
			return fmt.Errorf("unsafe argument: %w", err)
		}
	}

	return nil
}

// isAllowedAppPath checks if the command is a path to an allowed application
func isAllowedAppPath(cmd string) bool {
	// Normalize the path
	cleanPath := filepath.Clean(cmd)
	baseName := filepath.Base(cleanPath)

	// Remove .exe on Windows
	if runtime.GOOS == "windows" && strings.HasSuffix(baseName, ".exe") {
		baseName = strings.TrimSuffix(baseName, ".exe")
	}

	// Check against allowed app names
	allowedApps := []string{"GoPCA", "GoCSV", "gopca-desktop", "gocsv"}
	for _, app := range allowedApps {
		if baseName == app {
			// Verify it's in a reasonable location (not in system directories)
			if err := validateAppLocation(cleanPath); err == nil {
				return true
			}
		}
	}

	return false
}

// validateAppLocation ensures the app is in a reasonable location
func validateAppLocation(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check it's not in system directories
	systemDirs := SystemDirectories
	if runtime.GOOS == "windows" {
		systemDirs = WindowsSystemDirectories
	}

	for _, sysDir := range systemDirs {
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(sysDir)) {
			return fmt.Errorf("application in system directory")
		}
	}

	return nil
}

// isFilePath checks if an argument looks like a file path
func isFilePath(arg string) bool {
	// Simple heuristic: contains path separator or has file extension
	return strings.Contains(arg, string(filepath.Separator)) ||
		strings.Contains(arg, ".") ||
		strings.HasPrefix(arg, "~") ||
		strings.HasPrefix(arg, "/") ||
		(runtime.GOOS == "windows" && len(arg) > 2 && arg[1] == ':')
}

// isAllowedValue checks if an argument is an allowed value (not a flag)
func isAllowedValue(arg string) bool {
	// Allow specific patterns
	allowedPatterns := []string{
		"IMAGENAME eq", // Windows tasklist filter
	}

	for _, pattern := range allowedPatterns {
		if strings.HasPrefix(arg, pattern) {
			return true
		}
	}

	// Allow app names as values
	allowedValues := []string{
		"GoPCA", "GoCSV", "gopca-desktop", "gocsv",
		"GoPCA.exe", "GoCSV.exe", "gopca-desktop.exe", "gocsv.exe",
		"GoPCA.app", "GoCSV.app",
	}

	for _, val := range allowedValues {
		if arg == val || strings.Contains(arg, val) {
			return true
		}
	}

	return false
}

// validateArgumentSafety checks for command injection patterns
func validateArgumentSafety(arg string) error {
	// Check for shell metacharacters that could lead to injection
	dangerousChars := []string{
		";", "|", "&", "$", "`", "(", ")", "{", "}", "[", "]",
		"<", ">", "\\n", "\\r", "\\x00",
	}

	for _, char := range dangerousChars {
		if strings.Contains(arg, char) {
			return fmt.Errorf("contains dangerous character: %s", char)
		}
	}

	// Check for command substitution patterns
	dangerousPatterns := []string{
		"$(", "${", "`", "&&", "||",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("contains dangerous pattern: %s", pattern)
		}
	}

	// Check for excessive length (potential buffer overflow)
	if len(arg) > 1024 {
		return fmt.Errorf("argument too long: %d characters", len(arg))
	}

	return nil
}

// SecureCommand creates a secure exec.Cmd with validation
func SecureCommand(name string, args ...string) (*exec.Cmd, error) {
	// Validate the command and arguments
	if err := ValidateCommand(name, args); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Create the command
	cmd := exec.Command(name, args...)

	// Set secure environment (minimal, controlled environment)
	// Don't inherit all parent environment variables
	cmd.Env = []string{
		"PATH=" + getSecurePath(),
		"HOME=" + getHomeDir(),
		"TMPDIR=" + getTempDir(),
	}

	return cmd, nil
}

// getSecurePath returns a minimal, secure PATH
func getSecurePath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32;C:\Windows`
	case "darwin":
		return "/usr/bin:/bin:/usr/sbin:/sbin"
	default: // Linux
		return "/usr/bin:/bin:/usr/sbin:/sbin"
	}
}

// getHomeDir safely gets the home directory
func getHomeDir() string {
	// Use a safe method to get home directory
	if runtime.GOOS == "windows" {
		return `C:\Users\Public`
	}
	return "/tmp" // Safe default for Unix-like systems
}

// getTempDir safely gets the temp directory
func getTempDir() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\Temp`
	}
	return "/tmp"
}
