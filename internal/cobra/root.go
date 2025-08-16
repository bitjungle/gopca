// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
