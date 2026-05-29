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

package cobra

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information (set at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pca",
		Short: "GoPCA - Principal Component Analysis CLI",
		Long: `GoPCA is the definitive Principal Component Analysis (PCA) application.

A focused, professional-grade tool that excels at one thing: Principal Component Analysis.
Designed for data scientists, researchers, and engineers who need robust,
mathematically correct PCA with a modern command-line interface.

Features:
  • Multiple PCA algorithms (SVD, NIPALS, Kernel)
  • Comprehensive preprocessing options
  • Advanced diagnostics and metrics
  • Multiple output formats (JSON, CSV, Table)
  • Integration with data pipelines`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add subcommands
	rootCmd.AddCommand(
		NewAnalyzeCommand(),
		NewTransformCommand(),
		NewValidateCommand(),
		NewVersionCommand(),
		NewCompletionCommand(rootCmd),
	)

	return rootCmd
}

// Execute runs the CLI application
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
