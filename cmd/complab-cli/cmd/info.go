package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about input data file",
	Long: `Info displays detailed information about your data file.

It shows:
- File path and size
- Data dimensions
- Column names
- Memory usage estimate
- Data preview (first few rows)

Example:
  complab-cli info -i data.csv`,
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Required flags
	infoCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input CSV file (required)")
	infoCmd.MarkFlagRequired("input")
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Get file info
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}

	fmt.Println("📄 File Information")
	fmt.Println("═══════════════════════════════════════════════════════")
	
	// File details
	fmt.Printf("Path:         %s\n", inputFile)
	fmt.Printf("Size:         %s\n", formatFileSize(fileInfo.Size()))
	fmt.Printf("Modified:     %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
	
	// Try to read the file
	data, headers, hasRowNames, err := detectAndLoadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	rows := len(data)
	cols := 0
	if rows > 0 {
		cols = len(data[0])
	}

	fmt.Printf("\n📊 Data Shape\n")
	fmt.Printf("Rows:         %d\n", rows)
	fmt.Printf("Columns:      %d", cols)
	if hasRowNames {
		fmt.Printf(" (excluding row names)")
	}
	fmt.Println()
	
	// Memory estimate (8 bytes per float64 + overhead)
	memoryBytes := rows * cols * 8
	fmt.Printf("Memory est:   %s (in-memory)\n", formatFileSize(int64(memoryBytes)))

	// Column information
	if len(headers) > 0 {
		fmt.Println("\n📋 Columns")
		for i, header := range headers {
			fmt.Printf("%2d. %s\n", i+1, header)
		}
	}

	// Data preview
	if rows > 0 && verbose {
		fmt.Println("\n👀 Data Preview (first 5 rows)")
		fmt.Println("─────────────────────────────────────────────────────────")
		
		// Print headers
		if len(headers) > 0 {
			for i, h := range headers {
				if i > 0 {
					fmt.Print("\t")
				}
				// Truncate long headers
				if len(h) > 12 {
					fmt.Printf("%s...", h[:9])
				} else {
					fmt.Printf("%-12s", h)
				}
			}
			fmt.Println()
			fmt.Println("─────────────────────────────────────────────────────────")
		}
		
		// Print first 5 rows
		maxRows := 5
		if rows < maxRows {
			maxRows = rows
		}
		
		for i := 0; i < maxRows; i++ {
			for j, val := range data[i] {
				if j > 0 {
					fmt.Print("\t")
				}
				fmt.Printf("%-12.4g", val)
			}
			fmt.Println()
		}
		
		if rows > 5 {
			fmt.Printf("... (%d more rows)\n", rows-5)
		}
	}

	// Quick stats
	fmt.Println("\n📈 Quick Stats")
	missingCount := 0
	for _, row := range data {
		for _, val := range row {
			if isNaN(val) {
				missingCount++
			}
		}
	}
	
	totalCells := rows * cols
	if totalCells > 0 {
		completeness := float64(totalCells-missingCount) * 100.0 / float64(totalCells)
		fmt.Printf("Completeness: %.1f%% (%d missing values)\n", completeness, missingCount)
	}
	
	fmt.Println("\n✅ File is ready for analysis")

	return nil
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func isNaN(f float64) bool {
	return f != f
}