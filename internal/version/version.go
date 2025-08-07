// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
