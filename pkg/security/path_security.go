// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SystemDirectories that should never be written to
var SystemDirectories = []string{
	"/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin",
	"/sys", "/proc", "/dev", "/boot", "/lib", "/lib64",
	"/usr/lib", "/usr/local/bin", "/usr/local/sbin",
	"/var/log", "/root", "/home/root",
}

// WindowsSystemDirectories that should never be written to
var WindowsSystemDirectories = []string{
	`C:\Windows`, `C:\Program Files`, `C:\Program Files (x86)`,
	`C:\ProgramData`, `C:\System32`, `C:\SysWOW64`,
}

// ValidateInputPath validates a path for reading operations
func ValidateInputPath(path string) error {
	// Resolve to absolute path first to handle relative paths correctly
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Now validate the resolved absolute path
	if err := validateBasicPath(absPath); err != nil {
		return fmt.Errorf("input path validation failed: %w", err)
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", absPath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Ensure it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", absPath)
	}

	// Check file size
	if info.Size() > MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), MaxFileSize)
	}

	return nil
}

// ValidateOutputPath validates a path for writing operations
func ValidateOutputPath(path string) error {
	// Resolve to absolute path first to handle relative paths correctly
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Now validate the resolved absolute path
	if err := validateBasicPath(absPath); err != nil {
		return fmt.Errorf("output path validation failed: %w", err)
	}

	// Check for system directories
	if err := checkSystemDirectory(absPath); err != nil {
		return err
	}

	// Check parent directory exists and is writable
	dir := filepath.Dir(absPath)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("parent directory does not exist: %s", dir)
		}
		return fmt.Errorf("cannot access parent directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("parent path is not a directory: %s", dir)
	}

	// Check if we can write to the directory
	testFile := filepath.Join(dir, ".gopca_write_test_"+fmt.Sprintf("%d", os.Getpid()))
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %s", dir)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// validateBasicPath performs basic path validation
func validateBasicPath(path string) error {
	// Check for empty path
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Check length
	if len(path) > MaxPathLength {
		return fmt.Errorf("path too long: %d characters (max %d)", len(path), MaxPathLength)
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check for directory traversal
	if containsTraversal(cleanPath) {
		return fmt.Errorf("directory traversal detected")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null byte in path")
	}

	// Platform-specific validation
	if runtime.GOOS == "windows" {
		if err := validateWindowsPath(cleanPath); err != nil {
			return err
		}
	} else {
		if err := validateUnixPath(cleanPath); err != nil {
			return err
		}
	}

	return nil
}

// containsTraversal checks for directory traversal patterns
func containsTraversal(path string) bool {
	// For absolute paths, we don't need to check for traversal
	// since they've already been resolved
	if filepath.IsAbs(path) {
		return false
	}

	// Check for .. patterns in relative paths
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}

	return false
}

// checkSystemDirectory ensures we're not writing to system directories
func checkSystemDirectory(absPath string) error {
	// Normalize path for comparison
	normalizedPath := filepath.Clean(strings.ToLower(absPath))

	// Check against system directories based on OS
	var systemDirs []string
	if runtime.GOOS == "windows" {
		systemDirs = WindowsSystemDirectories
	} else {
		systemDirs = SystemDirectories
	}

	for _, sysDir := range systemDirs {
		normalizedSysDir := filepath.Clean(strings.ToLower(sysDir))
		if strings.HasPrefix(normalizedPath, normalizedSysDir) {
			return fmt.Errorf("cannot write to system directory: %s", sysDir)
		}
	}

	return nil
}

// validateWindowsPath performs Windows-specific path validation
func validateWindowsPath(path string) error {
	// Check for reserved names
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	// Check each path component for reserved names
	// Use backslash as separator for Windows paths
	parts := strings.Split(path, `\`)
	for _, part := range parts {
		// Skip empty parts and drive letters
		if part == "" || (len(part) == 2 && part[1] == ':') {
			continue
		}

		// Get the base name without extension
		base := strings.ToUpper(part)
		// Remove extension if present
		if dotPos := strings.LastIndex(base, "."); dotPos != -1 {
			base = base[:dotPos]
		}

		// Check against reserved names
		for _, reserved := range reservedNames {
			if base == reserved {
				return fmt.Errorf("reserved filename: %s", reserved)
			}
		}
	}

	// Check for invalid characters (excluding colon which needs special handling)
	// Colon is valid only in specific positions:
	// - Position 1 for drive letters (C:)
	// - After \\ for UNC paths (\\server\share)
	invalidChars := `<>"|?*`
	for _, char := range invalidChars {
		if strings.ContainsRune(path, char) {
			return fmt.Errorf("invalid character in path: %c", char)
		}
	}

	// Special handling for colon - only allowed after drive letter or in UNC paths
	colonPositions := []int{}
	for i, char := range path {
		if char == ':' {
			colonPositions = append(colonPositions, i)
		}
	}

	for _, pos := range colonPositions {
		// Allow colon at position 1 for drive letters (e.g., "C:")
		if pos == 1 && len(path) > 1 {
			// Check if it's a valid drive letter
			if path[0] >= 'A' && path[0] <= 'Z' || path[0] >= 'a' && path[0] <= 'z' {
				continue // This colon is valid
			}
		}
		// All other colons are invalid
		return fmt.Errorf("invalid character in path: colon not allowed except after drive letter")
	}

	// Check for trailing dots or spaces (Windows doesn't like these)
	// Note: We need to use backslash as separator for Windows paths
	// even when running on Unix systems for cross-platform validation
	// Re-split the path to check each component
	pathComponents := strings.Split(path, `\`)
	for _, part := range pathComponents {
		// Skip empty parts and drive letters (e.g., "C:")
		if part == "" || (len(part) == 2 && part[1] == ':') {
			continue
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("invalid path component: '%s' (trailing dot or space)", part)
		}
	}

	return nil
}

// validateUnixPath performs Unix-specific path validation
func validateUnixPath(path string) error {
	// Check for null character (already done in basic validation)

	// Check each path component
	parts := strings.Split(path, "/")
	for _, part := range parts {
		// Skip empty parts (from leading or trailing slashes)
		if part == "" {
			continue
		}

		// Component length check (most filesystems limit to 255)
		if len(part) > 255 {
			return fmt.Errorf("path component too long: %s", part)
		}
	}

	return nil
}

// SecureTempFile creates a secure temporary file
func SecureTempFile(pattern string) (*os.File, error) {
	// Sanitize the pattern
	pattern = SanitizeFilename(pattern)

	// Use the system temp directory
	tempDir := os.TempDir()

	// Create temp file with secure permissions
	f, err := os.CreateTemp(tempDir, "gopca_"+pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %w", err)
	}

	// Set restrictive permissions (owner read/write only)
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0600); err != nil {
			f.Close()
			os.Remove(f.Name())
			return nil, fmt.Errorf("cannot set file permissions: %w", err)
		}
	}

	return f, nil
}

// ResolveSymlinks safely resolves symbolic links
func ResolveSymlinks(path string) (string, error) {
	// First validate the basic path
	if err := validateBasicPath(path); err != nil {
		return "", err
	}

	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, just clean the path
			return filepath.Clean(path), nil
		}
		return "", fmt.Errorf("cannot resolve symlinks: %w", err)
	}

	// Validate the resolved path doesn't escape to system directories
	if err := validateBasicPath(resolved); err != nil {
		return "", fmt.Errorf("resolved path validation failed: %w", err)
	}

	return resolved, nil
}

// JailPath ensures a path stays within a jail directory
func JailPath(basePath, userPath string) (string, error) {
	// Clean both paths
	cleanBase := filepath.Clean(basePath)
	cleanUser := filepath.Clean(userPath)

	// Make paths absolute
	absBase, err := filepath.Abs(cleanBase)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base path: %w", err)
	}

	// Join paths and clean again
	joined := filepath.Join(absBase, cleanUser)
	final := filepath.Clean(joined)

	// Ensure the final path is within the base
	if !strings.HasPrefix(final, absBase) {
		return "", fmt.Errorf("path escapes jail: %s", userPath)
	}

	return final, nil
}
