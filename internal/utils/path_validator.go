package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFilePath checks if a file path is safe to use
// It prevents directory traversal attacks and ensures the path is clean
func ValidateFilePath(path string) error {
	// Clean the path to remove any redundant elements
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: directory traversal detected")
	}

	// Check for absolute paths that might access system files
	if filepath.IsAbs(cleanPath) {
		// Allow absolute paths but warn about them
		// In a more restrictive environment, you might want to reject these
		return nil
	}

	return nil
}

// ValidateOutputPath ensures an output path is safe to write to
func ValidateOutputPath(path string) error {
	if err := ValidateFilePath(path); err != nil {
		return err
	}

	// Additional checks for output paths
	dir := filepath.Dir(path)

	// Ensure we're not writing to system directories
	systemDirs := []string{"/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin", "/sys", "/proc"}
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(filepath.Clean(dir), sysDir) {
			return fmt.Errorf("cannot write to system directory: %s", dir)
		}
	}

	return nil
}
