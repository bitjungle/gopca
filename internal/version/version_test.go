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

package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	// Runtime-derived fields must always be populated.
	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, want GOOS/GOARCH form", info.Platform)
	}

	// Build-time fields fall back to their defaults when not set via ldflags.
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}
	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestInfoShort(t *testing.T) {
	info := Info{Version: "1.2.3"}
	if got := info.Short(); got != "1.2.3" {
		t.Errorf("Short() = %q, want %q", got, "1.2.3")
	}
}

func TestInfoString(t *testing.T) {
	info := Info{
		Version:   "1.2.3",
		GitCommit: "abc123",
		BuildDate: "2026-01-01T00:00:00Z",
		GoVersion: "go1.26",
		Platform:  "darwin/arm64",
	}
	got := info.String()
	for _, want := range []string{"1.2.3", "abc123", "2026-01-01T00:00:00Z", "go1.26", "darwin/arm64"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, missing %q", got, want)
		}
	}
}
