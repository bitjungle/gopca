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

// Package version provides version information for GoPCA.
// The version variables are populated at build time via ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of GoPCA (e.g., "0.9.0", "1.0.0")
	// Set at build time via: -ldflags "-X github.com/bitjungle/gopca/internal/version.Version=x.y.z"
	Version = "dev"

	// GitCommit is the git commit hash
	// Set at build time via: -ldflags "-X github.com/bitjungle/gopca/internal/version.GitCommit=abc123"
	GitCommit = "unknown"

	// BuildDate is the build date in RFC3339 format
	// Set at build time via: -ldflags "-X github.com/bitjungle/gopca/internal/version.BuildDate=2024-01-01T00:00:00Z"
	BuildDate = "unknown"
)

// Info contains all version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the complete version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("GoPCA %s (%s) built on %s with %s for %s",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}

// Short returns a short version string (just the version number)
func (i Info) Short() string {
	return i.Version
}
