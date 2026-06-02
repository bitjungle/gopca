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
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeZip builds an in-memory ZIP containing the given files (name → content).
func makeZip(files map[string]string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, _ := w.Create(name)
		_, _ = f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

// makeZipServer serves the given ZIP bytes from a httptest server.
func makeZipServer(t *testing.T, data []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(data)
	}))
}

// ── DownloadAndInspectZip ────────────────────────────────────────────────────

func TestDownloadAndInspectZip_SingleCSV(t *testing.T) {
	z := makeZip(map[string]string{
		"data.csv": "a,b,c\n1,2,3\n4,5,6\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	require.Empty(t, result.Error)
	require.Len(t, result.Entries, 1)
	assert.Equal(t, "data.csv", result.Entries[0].Name)
	assert.Equal(t, "csv", result.Entries[0].Format)
	assert.NotEmpty(t, app.pendingZipPath)

	// Clean up
	app.CancelZipImport()
	assert.Empty(t, app.pendingZipPath)
}

func TestDownloadAndInspectZip_MultipleEntries(t *testing.T) {
	z := makeZip(map[string]string{
		"wdbc.data":  "id,radius,texture\n1,17.99,10.38\n",
		"wdbc.names": "This file describes the dataset.\n",
		"README.md":  "# Dataset\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	require.Empty(t, result.Error)
	// Only wdbc.data should appear; .names and .md are not data extensions
	require.Len(t, result.Entries, 1)
	assert.Equal(t, "wdbc.data", result.Entries[0].Name)
	assert.Equal(t, "csv", result.Entries[0].Format)

	app.CancelZipImport()
}

func TestDownloadAndInspectZip_TwoDataFiles(t *testing.T) {
	z := makeZip(map[string]string{
		"train.csv": "a,b\n1,2\n",
		"test.csv":  "a,b\n3,4\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	require.Empty(t, result.Error)
	assert.Len(t, result.Entries, 2)

	app.CancelZipImport()
}

func TestDownloadAndInspectZip_NoDataFiles(t *testing.T) {
	z := makeZip(map[string]string{
		"README.md":  "# Info\n",
		"info.names": "Description\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "No supported data files")
	assert.Empty(t, app.pendingZipPath)
}

func TestDownloadAndInspectZip_PathTraversal(t *testing.T) {
	// Manually craft a ZIP with a dangerous entry name.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.CreateHeader(&zip.FileHeader{Name: "../../etc/passwd"})
	_, _ = f.Write([]byte("root:x:0:0\n"))
	w.Close()

	ts := makeZipServer(t, buf.Bytes())
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	assert.NotEmpty(t, result.Error)
	assert.Contains(t, strings.ToLower(result.Error), "unsafe")
	assert.Empty(t, app.pendingZipPath)
}

func TestDownloadAndInspectZip_ZipBomb(t *testing.T) {
	// 1 MB of identical bytes compresses to ~1 KB with Deflate (ratio ~1000:1),
	// well above the maxZipCompressionRatio = 100 limit.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: "bomb.csv", Method: zip.Deflate}
	f, _ := w.CreateHeader(h)
	_, _ = f.Write(bytes.Repeat([]byte("a"), 1024*1024))
	w.Close()

	ts := makeZipServer(t, buf.Bytes())
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")

	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "bomb")
	assert.Empty(t, app.pendingZipPath)
}

// ── LoadZipEntry ─────────────────────────────────────────────────────────────

func TestLoadZipEntry_CSV(t *testing.T) {
	z := makeZip(map[string]string{
		"data.csv": "x,y\n1,2\n3,4\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	inspectResult := app.DownloadAndInspectZip(ts.URL + "/archive.zip")
	require.Empty(t, inspectResult.Error)
	require.Len(t, inspectResult.Entries, 1)

	fd, err := app.LoadZipEntry("data.csv")
	require.NoError(t, err)
	require.NotNil(t, fd)
	assert.Greater(t, fd.Rows+fd.Columns, 0) // verify data was parsed
	assert.Empty(t, app.pendingZipPath)        // cleaned up after load
}

func TestLoadZipEntry_DataExtension(t *testing.T) {
	// UCI-style .data file: treated as CSV by the loader.
	z := makeZip(map[string]string{
		"wdbc.data": "id,label,radius,texture\n1,M,17.99,10.38\n2,B,20.57,17.77\n",
	})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	result := app.DownloadAndInspectZip(ts.URL + "/archive.zip")
	require.Empty(t, result.Error)

	fd, err := app.LoadZipEntry("wdbc.data")
	require.NoError(t, err)
	require.NotNil(t, fd)
	assert.Greater(t, fd.Rows+fd.Columns, 0) // verify .data parsed as CSV
	assert.Empty(t, app.pendingZipPath)
}

func TestLoadZipEntry_InvalidName(t *testing.T) {
	z := makeZip(map[string]string{"data.csv": "a,b\n1,2\n"})
	ts := makeZipServer(t, z)
	defer ts.Close()

	app := NewApp()
	app.DownloadAndInspectZip(ts.URL + "/archive.zip")
	// Attempt to load a different (non-existent) entry
	_, err := app.LoadZipEntry("../../etc/passwd")
	assert.Error(t, err)

	// Clean up the ZIP that is still pending
	if _, statErr := os.Stat(app.pendingZipPath); statErr == nil {
		app.CancelZipImport()
	}
}

// ── isSafeZipName ────────────────────────────────────────────────────────────

func TestIsSafeZipName(t *testing.T) {
	assert.True(t, isSafeZipName("data.csv"))
	assert.True(t, isSafeZipName("subdir/data.csv"))
	assert.False(t, isSafeZipName("../../etc/passwd"))
	assert.False(t, isSafeZipName("/absolute/path.csv"))
	assert.False(t, isSafeZipName("../sibling.csv"))
}
