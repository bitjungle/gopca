// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TempFileManager manages temporary files created for app-to-app communication
type TempFileManager struct {
	mu              sync.Mutex
	files           map[string]time.Time // filepath -> creation time
	maxAge          time.Duration
	cleanupInterval time.Duration
	stopChan        chan bool
}

// NewTempFileManager creates a new temporary file manager
func NewTempFileManager() *TempFileManager {
	manager := &TempFileManager{
		files:           make(map[string]time.Time),
		maxAge:          24 * time.Hour, // Keep files for 24 hours by default
		cleanupInterval: 1 * time.Hour,  // Clean up every hour
		stopChan:        make(chan bool),
	}

	// Start cleanup goroutine
	go manager.cleanupRoutine()

	return manager
}

// RegisterTempFile registers a temporary file for tracking
func (m *TempFileManager) RegisterTempFile(filepath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[filepath] = time.Now()
}

// CleanupOldFiles removes temporary files older than maxAge
func (m *TempFileManager) CleanupOldFiles() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for filepath, createdAt := range m.files {
		if now.Sub(createdAt) > m.maxAge {
			// Try to remove the file
			if err := os.Remove(filepath); err != nil {
				// File might already be deleted or in use
				if !os.IsNotExist(err) && os.Getenv("GOPCA_DEBUG") == "1" {
					fmt.Printf("[DEBUG] Failed to clean up temp file %s: %v\n", filepath, err)
				}
			}
			// Remove from tracking regardless
			delete(m.files, filepath)
		}
	}
}

// cleanupRoutine runs periodic cleanup
func (m *TempFileManager) cleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.CleanupOldFiles()
		case <-m.stopChan:
			return
		}
	}
}

// Stop stops the cleanup routine
func (m *TempFileManager) Stop() {
	close(m.stopChan)
}

// CreateTempFile creates a temporary file with a unique name
func (m *TempFileManager) CreateTempFile(prefix, extension string) (string, error) {
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	pid := os.Getpid()

	// Create unique filename
	filename := fmt.Sprintf("%s_%s_%d%s", prefix, timestamp, pid, extension)
	filepath := filepath.Join(tempDir, filename)

	// Register for cleanup
	m.RegisterTempFile(filepath)

	return filepath, nil
}

// CleanupGoPCATempFiles removes old GoPCA temporary files from the system temp directory
func CleanupGoPCATempFiles() error {
	tempDir := os.TempDir()

	// Find all GoPCA temp files
	pattern := filepath.Join(tempDir, "gopca_*.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list temp files: %w", err)
	}

	now := time.Now()
	maxAge := 24 * time.Hour

	for _, file := range matches {
		info, err := os.Stat(file)
		if err != nil {
			continue // Skip if we can't stat the file
		}

		// Remove files older than maxAge
		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
				if os.Getenv("GOPCA_DEBUG") == "1" {
					fmt.Printf("[DEBUG] Failed to clean up old temp file %s: %v\n", file, err)
				}
			}
		}
	}

	return nil
}
