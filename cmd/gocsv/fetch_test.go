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
