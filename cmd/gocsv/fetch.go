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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	"application/vnd.ms-excel":          ".xls",
	"application/zip":                   ".zip",
	"application/x-zip-compressed":      ".zip",
	"application/x-zip":                 ".zip",
	"multipart/x-zip":                   ".zip",
}

// fetchRemoteFile downloads the file at url to a secure temporary file and
// returns its path. The caller MUST remove the temp file when done.
//
// Extension is determined by (in order): URL path suffix, Content-Type header,
// magic bytes from the response body. Returns an error if the type cannot be
// determined or the download fails.
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

	// Detect extension: URL path suffix → Content-Type → magic bytes.
	// resp.Request.URL is the final (post-redirect) URL.
	ext := strings.ToLower(path.Ext(resp.Request.URL.Path))
	if ext == "" {
		ct := resp.Header.Get("Content-Type")
		if i := strings.Index(ct, ";"); i != -1 {
			ct = strings.TrimSpace(ct[:i])
		}
		// MIME types are case-insensitive (RFC 2045); normalise before lookup.
		ext = contentTypeToExt[strings.ToLower(ct)]
	}

	// Last resort: peek at the first 8 bytes of the response body.
	// We buffer them so the full body can still be written to the temp file.
	var peeked []byte
	if ext == "" {
		buf := make([]byte, 8)
		n, _ := io.ReadFull(resp.Body, buf)
		peeked = buf[:n]
		switch {
		case string(peeked[:min(4, n)]) == "PAR1":
			ext = ".parquet"
		case n >= 4 && peeked[0] == 'P' && peeked[1] == 'K' && peeked[2] == 0x03 && peeked[3] == 0x04:
			ext = ".zip"
		case n >= 2 && (string(peeked[:2]) == "<!" || strings.ToLower(string(peeked[:2])) == "<h"):
			return "", fmt.Errorf("URL points to a webpage, not a downloadable file")
		default:
			if n > 0 && peeked[0] >= 0x20 && peeked[0] < 0x7F {
				ext = ".csv"
			}
		}
	}
	if ext == "" {
		return "", fmt.Errorf("could not determine file type from URL or Content-Type header")
	}

	// Reconstruct the full body: prepend any peeked bytes.
	body := io.Reader(resp.Body)
	if len(peeked) > 0 {
		body = io.MultiReader(bytes.NewReader(peeked), resp.Body)
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
	written, err := io.Copy(tmpFile, io.LimitReader(body, fetchMaxBytes+1))
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

const peekTimeout = 10 * time.Second

// URLPeekResult holds the result of a HEAD-based URL inspection.
// Error is a user-friendly string rather than a Go error so that failures
// can be displayed in the UI without the frontend catching a rejected promise.
type URLPeekResult struct {
	URL           string `json:"url"`           // final URL after any rewriting
	FileFormat    string `json:"fileFormat"`    // "csv","tsv","xlsx","xls","parquet", or ""
	FileSizeBytes int64  `json:"fileSizeBytes"` // -1 when Content-Length is absent
	Accessible    bool   `json:"accessible"`
	Error         string `json:"error,omitempty"`
}

// rewriteGitHubURL converts a GitHub blob URL to the equivalent raw content URL.
// All other URLs are returned unchanged.
//
//	github.com/user/repo/blob/branch/path → raw.githubusercontent.com/user/repo/branch/path
func rewriteGitHubURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.Host != "github.com" {
		return rawURL
	}
	// Path must contain /blob/ to be a viewable file URL.
	parts := strings.SplitN(u.Path, "/blob/", 2)
	if len(parts) != 2 {
		return rawURL
	}
	u.Host = "raw.githubusercontent.com"
	u.Path = parts[0] + "/" + parts[1]
	return u.String()
}

// detectFormatFromExt returns a format name for known file extensions.
func detectFormatFromExt(urlPath string) string {
	switch strings.ToLower(path.Ext(urlPath)) {
	case ".parquet":
		return "parquet"
	case ".csv":
		return "csv"
	case ".tsv":
		return "tsv"
	case ".xlsx":
		return "xlsx"
	case ".xls":
		return "xls"
	case ".zip":
		return "zip"
	}
	return ""
}

// detectFormatFromMIME maps a Content-Type value to a format name.
func detectFormatFromMIME(ct string) string {
	if i := strings.Index(ct, ";"); i != -1 {
		ct = strings.TrimSpace(ct[:i])
	}
	ct = strings.ToLower(ct)
	switch ct {
	case "application/vnd.apache.parquet":
		return "parquet"
	case "text/csv":
		return "csv"
	case "text/tab-separated-values":
		return "tsv"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "application/vnd.ms-excel":
		return "xls"
	case "text/html", "application/xhtml+xml":
		return "html"
	case "application/zip", "application/x-zip-compressed",
		"application/x-zip", "multipart/x-zip":
		return "zip"
	}
	return ""
}

// detectFormatFromMagicBytes reads the first 8 bytes via a Range request and
// identifies the file type by magic number. Returns "" if unknown.
func detectFormatFromMagicBytes(rawURL string) string {
	ctx, cancel := context.WithTimeout(context.Background(), peekTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Range", "bytes=0-7")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	buf := make([]byte, 8)
	n, _ := io.ReadFull(resp.Body, buf)
	if n < 4 {
		return ""
	}
	switch {
	case string(buf[:4]) == "PAR1":
		return "parquet"
	case buf[0] == 'P' && buf[1] == 'K' && buf[2] == 0x03 && buf[3] == 0x04:
		return "zip"
	case string(buf[:2]) == "<!" || string(buf[:2]) == "<h" || string(buf[:2]) == "<H":
		return "html"
	default:
		// If the first bytes are printable ASCII, treat as CSV.
		if buf[0] >= 0x20 && buf[0] < 0x7F {
			return "csv"
		}
	}
	return ""
}

// PeekRemoteURL inspects a URL with a HEAD request to determine file format
// and size without downloading the full file. It rewrites GitHub blob URLs
// to raw content URLs automatically.
func (a *App) PeekRemoteURL(rawURL string) *URLPeekResult {
	finalURL := rewriteGitHubURL(strings.TrimSpace(rawURL))

	ctx, cancel := context.WithTimeout(context.Background(), peekTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, finalURL, nil)
	if err != nil {
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: "Invalid URL: " + err.Error()}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: "Could not connect to server: " + err.Error()}
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusPartialContent:
		// continue
	case http.StatusNotFound:
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: "File not found (HTTP 404)."}
	case http.StatusForbidden, http.StatusUnauthorized:
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: "Access denied — this file may require authentication."}
	default:
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: fmt.Sprintf("Server returned HTTP %d.", resp.StatusCode)}
	}

	// Detect format: URL extension → Content-Type → magic bytes.
	format := detectFormatFromExt(resp.Request.URL.Path)
	if format == "" {
		format = detectFormatFromMIME(resp.Header.Get("Content-Type"))
	}
	if format == "html" {
		return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
			Error: "This URL points to a webpage, not a downloadable file. Try right-clicking the download button on the data portal and copying the direct link address."}
	}
	if format == "" {
		format = detectFormatFromMagicBytes(finalURL)
		if format == "html" {
			return &URLPeekResult{URL: finalURL, FileSizeBytes: -1,
				Error: "This URL points to a webpage, not a downloadable file. Try right-clicking the download button on the data portal and copying the direct link address."}
		}
	}

	size := resp.ContentLength // -1 when absent
	if format == "" {
		return &URLPeekResult{URL: finalURL, Accessible: false, FileSizeBytes: size,
			Error: "Could not determine file type. Only CSV, TSV, Excel, Parquet, and ZIP files are supported."}
	}

	return &URLPeekResult{
		URL:           finalURL,
		FileFormat:    format,
		FileSizeBytes: size,
		Accessible:    true,
	}
}
