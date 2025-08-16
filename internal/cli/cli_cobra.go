// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cli

import (
	"github.com/bitjungle/gopca/internal/cobra"
	"github.com/bitjungle/gopca/internal/version"
)

// RunCobra executes the Cobra-based CLI application
func RunCobra() {
	// Set version information
	info := version.Get()
	cobra.Version = info.Short()
	cobra.BuildTime = info.BuildDate
	cobra.Commit = info.GitCommit

	// Execute the CLI
	cobra.Execute()
}
