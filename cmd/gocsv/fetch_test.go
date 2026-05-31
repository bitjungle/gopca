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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parquetTestServer returns a test server that serves testdata/energy_mix/.
func parquetTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.FileServer(http.Dir("../../testdata/energy_mix")))
}

func TestFetchRemoteFile_Parquet(t *testing.T) {
	ts := parquetTestServer(t)
	defer ts.Close()

	tmpPath, err := fetchRemoteFile(ts.URL + "/energy_mix.parquet")
	require.NoError(t, err)
	defer os.Remove(tmpPath)

	assert.True(t, strings.HasSuffix(tmpPath, ".parquet"), "temp file should have .parquet extension")

	info, err := os.Stat(tmpPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "temp file should not be empty")
}

func TestFetchRemoteFile_ContentTypeFallback(t *testing.T) {
	// Serve a CSV with explicit Content-Type and no recognisable URL extension
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("a,b,c\n1,2,3\n4,5,6\n"))
	}))
	defer ts.Close()

	tmpPath, err := fetchRemoteFile(ts.URL + "/data")
	require.NoError(t, err)
	defer os.Remove(tmpPath)

	assert.True(t, strings.HasSuffix(tmpPath, ".csv"), "should fall back to .csv from Content-Type")
}

func TestFetchRemoteFile_404(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	_, err := fetchRemoteFile(ts.URL + "/nonexistent.parquet")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestFetchRemoteFile_UnknownType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("binary data"))
	}))
	defer ts.Close()

	// URL has no extension, Content-Type not in our map
	_, err := fetchRemoteFile(ts.URL + "/file")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not determine file type")
}

func TestFetchRemoteFile_Redirect(t *testing.T) {
	// Serve a CSV at /final.csv; /redirect redirects there.
	// Extension detection must use the final URL, not the original.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final.csv", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("a,b\n1,2\n"))
	}))
	defer ts.Close()

	tmpPath, err := fetchRemoteFile(ts.URL + "/redirect")
	require.NoError(t, err)
	defer os.Remove(tmpPath)

	assert.True(t, strings.HasSuffix(tmpPath, ".csv"), "should detect .csv from redirected URL")
}

func TestFetchRemoteFile_CaseInsensitiveMIME(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Text/CSV; charset=utf-8")
		_, _ = w.Write([]byte("x,y\n1,2\n"))
	}))
	defer ts.Close()

	tmpPath, err := fetchRemoteFile(ts.URL + "/data")
	require.NoError(t, err)
	defer os.Remove(tmpPath)

	assert.True(t, strings.HasSuffix(tmpPath, ".csv"), "mixed-case MIME type should map to .csv")
}

func TestLoadCSVFromURL(t *testing.T) {
	ts := parquetTestServer(t)
	defer ts.Close()

	app := NewApp()
	fd, err := app.LoadCSV(ts.URL + "/energy_mix.parquet")
	require.NoError(t, err)
	require.NotNil(t, fd)

	assert.Equal(t, 7314, fd.Rows)
	assert.Equal(t, 106, fd.Columns)
	assert.Equal(t, "country#target", fd.Headers[0])
	assert.Equal(t, 7314, len(fd.RowNames))
	assert.Equal(t, "1", fd.RowNames[0])
}

// ── rewriteGitHubURL ────────────────────────────────────────────────────────

func TestRewriteGitHubURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "blob URL rewritten to raw",
			input: "https://github.com/user/repo/blob/main/data.csv",
			want:  "https://raw.githubusercontent.com/user/repo/main/data.csv",
		},
		{
			name:  "blob URL with subdirectory",
			input: "https://github.com/user/repo/blob/feature-branch/path/to/data.parquet",
			want:  "https://raw.githubusercontent.com/user/repo/feature-branch/path/to/data.parquet",
		},
		{
			name:  "non-GitHub URL unchanged",
			input: "https://example.com/data.csv",
			want:  "https://example.com/data.csv",
		},
		{
			name:  "already raw URL unchanged",
			input: "https://raw.githubusercontent.com/user/repo/main/data.csv",
			want:  "https://raw.githubusercontent.com/user/repo/main/data.csv",
		},
		{
			name:  "GitHub repo root (no blob) unchanged",
			input: "https://github.com/user/repo",
			want:  "https://github.com/user/repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, rewriteGitHubURL(tt.input))
		})
	}
}

// ── PeekRemoteURL ────────────────────────────────────────────────────────────

func TestPeekRemoteURL_CSV(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Length", "18")
		_, _ = w.Write([]byte("a,b,c\n1,2,3\n4,5,6\n"))
	}))
	defer ts.Close()

	app := &App{}
	result := app.PeekRemoteURL(ts.URL + "/data.csv")

	assert.True(t, result.Accessible)
	assert.Equal(t, "csv", result.FileFormat)
	assert.Equal(t, int64(18), result.FileSizeBytes)
	assert.Empty(t, result.Error)
}

func TestPeekRemoteURL_ParquetByExtension(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte("PAR1"))
	}))
	defer ts.Close()

	app := &App{}
	// Extension in URL path takes priority over Content-Type
	result := app.PeekRemoteURL(ts.URL + "/energy_mix.parquet")

	assert.True(t, result.Accessible)
	assert.Equal(t, "parquet", result.FileFormat)
	assert.Empty(t, result.Error)
}

func TestPeekRemoteURL_HTML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!DOCTYPE html><html>"))
	}))
	defer ts.Close()

	app := &App{}
	result := app.PeekRemoteURL(ts.URL + "/page")

	assert.False(t, result.Accessible)
	assert.Empty(t, result.FileFormat)
	assert.Contains(t, result.Error, "webpage")
}

func TestPeekRemoteURL_404(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	app := &App{}
	result := app.PeekRemoteURL(ts.URL + "/missing.csv")

	assert.False(t, result.Accessible)
	assert.Contains(t, result.Error, "404")
}

func TestPeekRemoteURL_NoContentLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		// Force chunked transfer encoding so Go's HTTP server omits Content-Length.
		w.Header().Set("Transfer-Encoding", "chunked")
		rc := http.NewResponseController(w)
		_ = rc.EnableFullDuplex()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("x,y\n"))
		_, _ = w.Write([]byte("1,2\n"))
		_ = rc.Flush()
	}))
	defer ts.Close()

	app := &App{}
	result := app.PeekRemoteURL(ts.URL + "/data.csv")

	assert.True(t, result.Accessible)
	assert.Equal(t, "csv", result.FileFormat)
	assert.Equal(t, int64(-1), result.FileSizeBytes)
	assert.Empty(t, result.Error)
}

func TestPeekRemoteURL_GitHubRewrite(t *testing.T) {
	// Serve a CSV; verify that the GitHub blob URL is rewritten and the
	// request reaches the server (which stands in for raw.githubusercontent.com).
	var receivedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("a,b\n1,2\n"))
	}))
	defer ts.Close()

	// Construct a fake "GitHub blob" URL pointing at the test server.
	// We can't truly test the host rewrite without DNS tricks, so instead
	// we verify that rewriteGitHubURL produces the correct transformation
	// and that PeekRemoteURL uses the rewritten URL (result.URL field).
	app := &App{}
	ghURL := "https://github.com/user/repo/blob/main/data.csv"
	result := app.PeekRemoteURL(ghURL)
	// The rewritten URL should be raw.githubusercontent.com — it will fail
	// to connect in tests (no live network), but result.URL must be rewritten.
	assert.Equal(t, "https://raw.githubusercontent.com/user/repo/main/data.csv", result.URL)
	_ = receivedPath
}
