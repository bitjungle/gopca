// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package integration provides shared functionality for integrating
// GoPCA and GoCSV applications, including app detection and launching.
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AppStatus represents the installation status of an external application
type AppStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// AppConfig contains configuration for detecting and launching an application
type AppConfig struct {
	Name        string   // Application name (e.g., "gopca-desktop", "gocsv")
	CommonPaths []string // Common installation paths to check
	DisplayName string   // User-friendly name for error messages
}

// GetCommonPaths returns platform-specific common installation paths for an application
func GetCommonPaths(appName string) []string {
	switch appName {
	case "gopca-desktop":
		return getGoPCACommonPaths()
	case "gocsv":
		return getGoCSVCommonPaths()
	default:
		return []string{}
	}
}

// getGoPCACommonPaths returns common installation paths for GoPCA Desktop
func getGoPCACommonPaths() []string {
	paths := []string{}

	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/Applications/GoPCA.app/Contents/MacOS/GoPCA",
			"/Applications/GoPCA Desktop.app/Contents/MacOS/GoPCA Desktop",
			"/usr/local/bin/gopca-desktop",
		)
	case "windows":
		paths = append(paths,
			"C:\\Program Files\\GoPCA\\GoPCA.exe",
			"C:\\Program Files\\GoPCA Desktop\\GoPCA Desktop.exe",
			"C:\\Program Files (x86)\\GoPCA\\GoPCA.exe",
		)
	case "linux":
		paths = append(paths,
			"/usr/local/bin/gopca-desktop",
			"/usr/bin/gopca-desktop",
			"/opt/gopca-desktop/gopca-desktop",
		)
	}

	// Add user home directory paths
	if home, err := os.UserHomeDir(); err == nil {
		switch runtime.GOOS {
		case "darwin":
			paths = append(paths, home+"/Applications/GoPCA.app/Contents/MacOS/GoPCA")
		case "linux":
			paths = append(paths, home+"/.local/bin/gopca-desktop")
		}
	}

	return paths
}

// getGoCSVCommonPaths returns common installation paths for GoCSV
func getGoCSVCommonPaths() []string {
	paths := []string{}

	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/Applications/GoCSV.app/Contents/MacOS/GoCSV",
			"/usr/local/bin/gocsv",
		)
	case "windows":
		paths = append(paths,
			"C:\\Program Files\\GoCSV\\GoCSV.exe",
			"C:\\Program Files (x86)\\GoCSV\\GoCSV.exe",
		)
	case "linux":
		paths = append(paths,
			"/usr/local/bin/gocsv",
			"/usr/bin/gocsv",
			"/opt/gocsv/gocsv",
		)
	}

	// Add user home directory paths
	if home, err := os.UserHomeDir(); err == nil {
		switch runtime.GOOS {
		case "darwin":
			paths = append(paths, home+"/Applications/GoCSV.app/Contents/MacOS/GoCSV")
		case "linux":
			paths = append(paths, home+"/.local/bin/gocsv")
		}
	}

	return paths
}

// getCompanionAppPaths returns possible paths for companion apps in the same directory
func getCompanionAppPaths(currentExePath string, targetApp string) []string {
	paths := []string{}
	execDir := filepath.Dir(currentExePath)

	// Special handling for macOS App Translocation
	// When an app is translocated, it runs from /private/var/folders/.../AppTranslocation/
	// but the companion app is in the original location
	if runtime.GOOS == "darwin" && strings.Contains(currentExePath, "/AppTranslocation/") {
		if os.Getenv("GOPCA_DEBUG") == "1" {
			fmt.Printf("[DEBUG] App is translocated, searching common locations for companion apps\n")
		}

		home, _ := os.UserHomeDir()
		commonLocations := []string{
			"/Applications",
			filepath.Join(home, "Applications"),
			filepath.Join(home, "Downloads"),
			filepath.Join(home, "Desktop"),
			// Common extracted paths
			filepath.Join(home, "Downloads", "gopca-macos-universal"),
			filepath.Join(home, "Downloads", "release-bundles", "gopca-macos-universal"),
		}

		// Search each location for both apps together
		for _, dir := range commonLocations {
			var targetPath string
			var companionPath string

			switch targetApp {
			case "gocsv":
				targetPath = filepath.Join(dir, "GoCSV.app", "Contents", "MacOS", "GoCSV")
				companionPath = filepath.Join(dir, "GoPCA.app", "Contents", "MacOS", "GoPCA")
			case "gopca-desktop":
				targetPath = filepath.Join(dir, "GoPCA.app", "Contents", "MacOS", "GoPCA")
				companionPath = filepath.Join(dir, "GoCSV.app", "Contents", "MacOS", "GoCSV")
			}

			if os.Getenv("GOPCA_DEBUG") == "1" {
				fmt.Printf("[DEBUG] Checking translocation fallback: %s\n", targetPath)
			}

			// Check if both apps exist in this location
			if _, err := os.Stat(targetPath); err == nil {
				if _, err := os.Stat(companionPath); err == nil {
					// Both apps found in same location
					if os.Getenv("GOPCA_DEBUG") == "1" {
						fmt.Printf("[DEBUG] Found both apps in: %s\n", dir)
					}
					// Prepend to paths so it's checked first
					paths = append([]string{targetPath}, paths...)
				}
			}
		}
	}

	switch targetApp {
	case "gocsv":
		// Check for GoCSV variants
		switch runtime.GOOS {
		case "darwin":
			// Check if we're inside an app bundle
			if strings.Contains(currentExePath, ".app/Contents/MacOS") {
				// Go up to the directory containing the .app bundles
				// From /path/to/GoPCA.app/Contents/MacOS/GoPCA, we need to go up 4 levels
				appBundleParent := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentExePath))))
				targetPath := filepath.Join(appBundleParent, "GoCSV.app", "Contents", "MacOS", "GoCSV")
				if os.Getenv("GOPCA_DEBUG") == "1" {
					fmt.Printf("[DEBUG] Looking for GoCSV at: %s\n", targetPath)
					fmt.Printf("[DEBUG] Parent directory: %s\n", appBundleParent)
				}
				paths = append(paths, targetPath)
			} else {
				// Not in an app bundle, check for apps in same directory
				paths = append(paths,
					filepath.Join(execDir, "GoCSV.app", "Contents", "MacOS", "GoCSV"),
					filepath.Join(execDir, "GoCSV"),
					filepath.Join(execDir, "gocsv"),
				)
			}
		case "windows":
			// Windows: Check same directory
			paths = append(paths,
				filepath.Join(execDir, "GoCSV.exe"),
				filepath.Join(execDir, "gocsv.exe"),
			)
		default: // Linux
			// Linux: Check same directory
			paths = append(paths,
				filepath.Join(execDir, "GoCSV"),
				filepath.Join(execDir, "gocsv"),
			)
		}
	case "gopca-desktop":
		// Check for GoPCA variants
		switch runtime.GOOS {
		case "darwin":
			// Check if we're inside an app bundle
			if strings.Contains(currentExePath, ".app/Contents/MacOS") {
				// Go up to the directory containing the .app bundles
				// From /path/to/GoCSV.app/Contents/MacOS/GoCSV, we need to go up 4 levels
				appBundleParent := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentExePath))))
				paths = append(paths,
					filepath.Join(appBundleParent, "GoPCA.app", "Contents", "MacOS", "GoPCA"),
				)
			} else {
				// Not in an app bundle, check for apps in same directory
				paths = append(paths,
					filepath.Join(execDir, "GoPCA.app", "Contents", "MacOS", "GoPCA"),
					filepath.Join(execDir, "GoPCA"),
					filepath.Join(execDir, "gopca-desktop"),
				)
			}
		case "windows":
			// Windows: Check same directory
			paths = append(paths,
				filepath.Join(execDir, "GoPCA.exe"),
				filepath.Join(execDir, "gopca-desktop.exe"),
			)
		default: // Linux
			// Linux: Check same directory
			paths = append(paths,
				filepath.Join(execDir, "GoPCA"),
				filepath.Join(execDir, "gopca-desktop"),
			)
		}
	}

	return paths
}

// CheckApp checks if an application is installed on the system
func CheckApp(config AppConfig) *AppStatus {
	status := &AppStatus{Installed: false}

	// Debug logging (can be removed after testing)
	if os.Getenv("GOPCA_DEBUG") == "1" {
		fmt.Printf("[DEBUG] CheckApp: Looking for %s\n", config.Name)
	}

	// First check PATH for multiple possible names
	possibleNames := []string{config.Name}

	// Add alternative names based on app
	switch config.Name {
	case "gocsv":
		possibleNames = append(possibleNames, "GoCSV")
	case "gopca-desktop":
		possibleNames = append(possibleNames, "GoPCA")
	}

	// Check PATH for all possible names
	for _, name := range possibleNames {
		if path, err := exec.LookPath(name); err == nil {
			status.Installed = true
			status.Path = path
			return status
		}
	}

	// Check for companion apps in same directory as current executable
	if currentExe, err := os.Executable(); err == nil {
		originalPath := currentExe
		// Resolve symlinks to get the real path
		if realPath, err := filepath.EvalSymlinks(currentExe); err == nil {
			currentExe = realPath
		}

		if os.Getenv("GOPCA_DEBUG") == "1" {
			fmt.Printf("[DEBUG] Original executable: %s\n", originalPath)
			fmt.Printf("[DEBUG] Resolved executable: %s\n", currentExe)
			fmt.Printf("[DEBUG] Looking for app: %s\n", config.Name)
		}

		companionPaths := getCompanionAppPaths(currentExe, config.Name)
		for _, p := range companionPaths {
			if os.Getenv("GOPCA_DEBUG") == "1" {
				fmt.Printf("[DEBUG] Checking companion path: %s\n", p)
			}
			if _, err := os.Stat(p); err == nil {
				if os.Getenv("GOPCA_DEBUG") == "1" {
					fmt.Printf("[DEBUG] Found at: %s\n", p)
				}
				status.Installed = true
				status.Path = p
				return status
			}
		}
	}

	// Check common installation paths
	for _, p := range config.CommonPaths {
		if _, err := os.Stat(p); err == nil {
			status.Installed = true
			status.Path = p
			return status
		}
	}

	status.Error = fmt.Sprintf("%s not found in PATH or common installation locations", config.DisplayName)
	return status
}

// LaunchWithFile launches an application with a file argument
func LaunchWithFile(appPath, filePath string) error {
	// Validate that the app exists
	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("application not found at %s: %w", appPath, err)
	}

	// Validate that the file exists
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found at %s: %w", filePath, err)
	}

	var cmd *exec.Cmd

	// Platform-specific launching
	switch runtime.GOOS {
	case "darwin":
		// On macOS, check if this is an app bundle
		if strings.Contains(appPath, ".app/Contents/MacOS/") {
			// Extract the .app bundle path and app name
			appIndex := strings.Index(appPath, ".app")
			if appIndex != -1 {
				appBundlePath := appPath[:appIndex+4] // Include ".app"
				appName := filepath.Base(appBundlePath)
				appName = strings.TrimSuffix(appName, ".app")

				// Check if the app is already running
				if IsAppRunning(appName) {
					// App is running, use 'open -n' to force a new instance
					// This ensures the file argument is properly passed
					// Wails apps don't handle AppleScript open events, so we need a new instance
					cmd = exec.Command("open", "-n", "-a", appBundlePath, "--args", "--open", filePath)

					if os.Getenv("GOPCA_DEBUG") == "1" {
						fmt.Printf("[DEBUG] App %s is running, forcing new instance with file\n", appName)
					}
				} else {
					// App not running, use the standard open command
					cmd = exec.Command("open", "-a", appBundlePath, "--args", "--open", filePath)

					if os.Getenv("GOPCA_DEBUG") == "1" {
						fmt.Printf("[DEBUG] App %s not running, using open command\n", appName)
					}
				}
			} else {
				// Fallback to direct execution
				cmd = exec.Command(appPath, "--open", filePath)
			}
		} else {
			// Regular binary
			cmd = exec.Command(appPath, "--open", filePath)
		}
	case "windows":
		// On Windows, just launch the executable with the file
		cmd = exec.Command(appPath, "--open", filePath)
	default: // Linux and others
		// On Linux, just launch the executable with the file
		cmd = exec.Command(appPath, "--open", filePath)
	}

	// Start the application (don't wait for it to finish)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch application: %w", err)
	}

	// Detach from the process so it continues running after our app exits
	// Ignore the error as it's not critical - the app will still run
	_ = cmd.Process.Release()

	return nil
}

// IsAppRunning checks if an application is currently running
func IsAppRunning(appName string) bool {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, use pgrep to check if the app is running
		// Extract just the app name from paths like "GoPCA.app"
		baseName := appName
		if strings.Contains(appName, ".app") {
			baseName = strings.TrimSuffix(filepath.Base(appName), ".app")
		}

		cmd := exec.Command("pgrep", "-x", baseName)
		err := cmd.Run()
		return err == nil // pgrep returns 0 if process found

	case "windows":
		// On Windows, use tasklist to check for running processes
		baseName := filepath.Base(appName)
		if !strings.HasSuffix(baseName, ".exe") {
			baseName += ".exe"
		}

		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", baseName))
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), baseName)

	default: // Linux
		// On Linux, use pgrep similar to macOS
		baseName := filepath.Base(appName)
		cmd := exec.Command("pgrep", "-x", baseName)
		err := cmd.Run()
		return err == nil
	}
}
