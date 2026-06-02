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

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// maxZipTotalUncompressed caps the total uncompressed size across all entries.
	maxZipTotalUncompressed = 500 * 1024 * 1024 // 500 MB

	// maxZipCompressionRatio rejects entries whose compression ratio exceeds this
	// value, guarding against zip bombs. A ratio of 100 means 1 KB compressed
	// could expand to at most 100 KB.
	maxZipCompressionRatio = 100
)

// zipDataExts lists file extensions inside a ZIP that are treated as importable
// data files. Everything else is silently skipped.
var zipDataExts = map[string]bool{
	".csv":     true,
	".tsv":     true,
	".data":    true, // UCI Machine Learning Repository convention
	".xlsx":    true,
	".xls":     true,
	".parquet": true,
}

// ZipEntry describes a single importable file found inside a ZIP archive.
type ZipEntry struct {
	Name             string `json:"name"`             // entry path as stored in the ZIP
	UncompressedSize uint64 `json:"uncompressedSize"` // bytes
	Format           string `json:"format"`           // "csv", "parquet", etc.
}

// ZipInspectResult is returned by DownloadAndInspectZip.
// Error is a user-friendly string rather than a Go error so that failures
// display in the UI rather than rejecting the Wails promise.
type ZipInspectResult struct {
	Entries []ZipEntry `json:"entries"`
	Error   string     `json:"error,omitempty"`
}

// isSafeZipName returns false if the entry name could escape the extraction
// directory via path traversal.
func isSafeZipName(name string) bool {
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return false
	}
	cleaned := filepath.ToSlash(filepath.Clean(name))
	if strings.Contains(cleaned, "..") {
		return false
	}
	return true
}

// DownloadAndInspectZip downloads the ZIP at url, validates it for safety, and
// returns the list of importable data-file entries. The ZIP temp file is kept
// on disk at a.pendingZipPath until LoadZipEntry or CancelZipImport is called.
func (a *App) DownloadAndInspectZip(url string) *ZipInspectResult {
	// Download to a temp file.
	tmpPath, err := fetchRemoteFile(url)
	if err != nil {
		return &ZipInspectResult{Error: "Download failed: " + err.Error()}
	}

	// Open the ZIP archive.
	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return &ZipInspectResult{Error: "Could not open ZIP archive: " + err.Error()}
	}
	defer zr.Close()

	var entries []ZipEntry
	var totalUncompressed uint64

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Path traversal check.
		if !isSafeZipName(f.Name) {
			os.Remove(tmpPath)
			return &ZipInspectResult{
				Error: fmt.Sprintf("ZIP contains unsafe entry name %q — archive rejected.", f.Name),
			}
		}

		// Accumulate uncompressed size for zip bomb detection.
		totalUncompressed += f.UncompressedSize64
		if totalUncompressed > maxZipTotalUncompressed {
			os.Remove(tmpPath)
			return &ZipInspectResult{
				Error: fmt.Sprintf("ZIP exceeds %d MB uncompressed size limit.", maxZipTotalUncompressed/(1024*1024)),
			}
		}

		// Per-entry compression ratio check.
		if f.CompressedSize64 > 0 &&
			f.UncompressedSize64/f.CompressedSize64 > maxZipCompressionRatio {
			os.Remove(tmpPath)
			return &ZipInspectResult{Error: "ZIP bomb detected — archive rejected."}
		}

		// Only include data files.
		ext := strings.ToLower(filepath.Ext(f.Name))
		if !zipDataExts[ext] {
			continue
		}

		format := detectFormatFromExt(f.Name)
		if format == "" {
			format = "csv" // .data and other unrecognised data exts default to csv
		}

		entries = append(entries, ZipEntry{
			Name:             f.Name,
			UncompressedSize: f.UncompressedSize64,
			Format:           format,
		})
	}

	if len(entries) == 0 {
		os.Remove(tmpPath)
		return &ZipInspectResult{
			Error: "No supported data files found in this ZIP (expected .csv, .tsv, .data, .xlsx, or .parquet).",
		}
	}

	// Keep the ZIP on disk; LoadZipEntry or CancelZipImport will clean it up.
	a.pendingZipPath = tmpPath
	return &ZipInspectResult{Entries: entries}
}

// LoadZipEntry extracts the named entry from the pending ZIP and loads it as
// FileData. It cleans up the ZIP temp file regardless of success or failure.
func (a *App) LoadZipEntry(entryName string) (*FileData, error) {
	if a.pendingZipPath == "" {
		return nil, fmt.Errorf("no ZIP file is pending")
	}
	zipPath := a.pendingZipPath
	a.pendingZipPath = ""
	defer os.Remove(zipPath)

	// Re-validate the entry name to prevent any path traversal.
	if !isSafeZipName(entryName) {
		return nil, fmt.Errorf("invalid entry name")
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("could not open ZIP: %w", err)
	}
	defer zr.Close()

	// Find the matching entry.
	var found *zip.File
	for _, f := range zr.File {
		if f.Name == entryName {
			found = f
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("entry %q not found in ZIP", entryName)
	}

	// Extract to a temp file with the correct extension.
	ext := strings.ToLower(filepath.Ext(entryName))
	tmpFile, err := os.CreateTemp("", "gocsv-zip-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("could not create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	rc, err := found.Open()
	if err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("could not open ZIP entry: %w", err)
	}

	// Honour the size limit while extracting.
	written, err := io.Copy(tmpFile, io.LimitReader(rc, maxZipTotalUncompressed+1))
	rc.Close()
	tmpFile.Close()
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}
	if uint64(written) > maxZipTotalUncompressed {
		return nil, fmt.Errorf("extracted file exceeds %d MB limit", maxZipTotalUncompressed/(1024*1024))
	}

	// Load via the appropriate handler.
	switch ext {
	case ".parquet":
		return a.loadParquet(tmpPath)
	case ".xlsx", ".xls":
		return a.loadExcel(tmpPath)
	case ".tsv":
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("could not read extracted file: %w", err)
		}
		return a.parseCSVContent(string(content), ".tsv")
	default:
		// .csv, .data and unrecognised text extensions — CSV parser with auto-detection.
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("could not read extracted file: %w", err)
		}
		return a.parseCSVContent(string(content), ".csv")
	}
}

// CancelZipImport removes the pending ZIP temp file if one exists.
// Called when the user dismisses the ZIP file picker without importing.
func (a *App) CancelZipImport() {
	if a.pendingZipPath != "" {
		os.Remove(a.pendingZipPath)
		a.pendingZipPath = ""
	}
}
