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
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

)

const (
	fetchTimeout  = 60 * time.Second
	fetchMaxBytes = 500 * 1024 * 1024 // 500 MB — matches security.MaxFileSize
)

// contentTypeToExt maps MIME types to file extensions for format detection.
var contentTypeToExt = map[string]string{
	"application/vnd.apache.parquet": ".parquet",
	"text/csv":                       ".csv",
	"text/tab-separated-values":      ".tsv",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
	"application/vnd.ms-excel": ".xls",
}

// fetchRemoteFile downloads the file at url to a secure temporary file and
// returns its path. The caller MUST remove the temp file when done.
//
// Extension is determined by (in order): URL path suffix, Content-Type header.
// Returns an error if the extension cannot be determined or the download fails.
func fetchRemoteFile(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	// Detect extension from the final (post-redirect) URL path first, then
	// Content-Type as fallback. resp.Request.URL is the URL that actually served
	// the body, which differs from req.URL when the server issued a redirect.
	ext := strings.ToLower(path.Ext(resp.Request.URL.Path))
	if ext == "" {
		ct := resp.Header.Get("Content-Type")
		if i := strings.Index(ct, ";"); i != -1 {
			ct = strings.TrimSpace(ct[:i])
		}
		// MIME types are case-insensitive (RFC 2045); normalise before lookup.
		ext = contentTypeToExt[strings.ToLower(ct)]
	}
	if ext == "" {
		return "", fmt.Errorf("could not determine file type from URL or Content-Type header")
	}

	// Create a temp file with the detected extension so the caller can route it
	// through the existing extension switch. os.CreateTemp uses O_EXCL and
	// creates the file with mode 0600, so no explicit chmod is needed.
	tmpFile, err := os.CreateTemp("", "gocsv-fetch-*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Stream body to temp file with size limit.
	written, err := io.Copy(tmpFile, io.LimitReader(resp.Body, fetchMaxBytes+1))
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("download failed: %w", err)
	}
	if written > fetchMaxBytes {
		os.Remove(tmpPath)
		return "", fmt.Errorf("remote file exceeds %d MB limit", fetchMaxBytes/(1024*1024))
	}

	return tmpPath, nil
}
