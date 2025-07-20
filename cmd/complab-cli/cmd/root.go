package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "complab-cli",
	Short: "CompLab PCA Toolkit - Professional-grade Principal Component Analysis",
	Long: `CompLab PCA Toolkit is a comprehensive tool for performing Principal Component Analysis.

It provides multiple algorithms (NIPALS and SVD), preprocessing options, and flexible output formats.
Perfect for data scientists, researchers, and engineers working with multivariate data.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
}