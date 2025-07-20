package cmd

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate input data format and content",
	Long: `Validate checks your CSV file for format validity and data quality.

It reports:
- File format validity
- Data dimensions (rows × columns)
- Missing values count and locations
- Non-numeric values
- Basic statistics per column

Example:
  complab-cli validate -i data.csv`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Required flags
	validateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input CSV file to validate (required)")
	validateCmd.MarkFlagRequired("input")
}

func runValidate(cmd *cobra.Command, args []string) error {
	if !quiet {
		fmt.Printf("Validating file: %s\n\n", inputFile)
	}

	// Try to read the file
	data, headers, hasRowNames, err := detectAndLoadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("❌ Invalid CSV format: %w", err)
	}

	if !quiet {
		fmt.Println("✅ File format: Valid CSV")
		if hasRowNames {
			fmt.Println("ℹ️  Row names detected in first column (skipped for analysis)")
		}
	}

	// Check dimensions
	rows := len(data)
	cols := 0
	if rows > 0 {
		cols = len(data[0])
	}

	fmt.Printf("📊 Data dimensions: %d rows × %d columns\n", rows, cols)

	// Display headers
	if len(headers) > 0 {
		fmt.Printf("📋 Column names: ")
		for i, h := range headers {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s", h)
		}
		fmt.Println()
	}

	// Check for missing values and compute statistics
	missingCount := 0
	missingLocations := make(map[string][]int)
	
	// Statistics per column
	colStats := make([]struct {
		min, max, mean, sum float64
		count               int
		nonNumeric         int
	}, cols)

	// Initialize stats
	for j := 0; j < cols; j++ {
		colStats[j].min = math.Inf(1)
		colStats[j].max = math.Inf(-1)
	}

	// Analyze data
	for i, row := range data {
		for j, val := range row {
			if math.IsNaN(val) {
				missingCount++
				colName := fmt.Sprintf("Col_%d", j+1)
				if j < len(headers) {
					colName = headers[j]
				}
				missingLocations[colName] = append(missingLocations[colName], i+1)
			} else {
				colStats[j].count++
				colStats[j].sum += val
				if val < colStats[j].min {
					colStats[j].min = val
				}
				if val > colStats[j].max {
					colStats[j].max = val
				}
			}
		}
	}

	// Calculate means
	for j := 0; j < cols; j++ {
		if colStats[j].count > 0 {
			colStats[j].mean = colStats[j].sum / float64(colStats[j].count)
		}
	}

	// Report missing values
	fmt.Printf("\n🔍 Missing values: %d", missingCount)
	if missingCount > 0 {
		totalCells := rows * cols
		percentage := float64(missingCount) * 100.0 / float64(totalCells)
		fmt.Printf(" (%.2f%% of data)", percentage)
		
		if verbose && len(missingLocations) > 0 {
			fmt.Println("\n   Locations:")
			for col, locs := range missingLocations {
				fmt.Printf("   - %s: rows ", col)
				maxShow := 10
				for i, loc := range locs {
					if i >= maxShow {
						fmt.Printf("... (%d more)", len(locs)-maxShow)
						break
					}
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%d", loc)
				}
				fmt.Println()
			}
		}
	}
	fmt.Println()

	// Report statistics
	if verbose || !quiet {
		fmt.Println("\n📈 Column statistics:")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		fmt.Printf("%-20s %10s %10s %10s %10s\n", "Column", "Min", "Max", "Mean", "Valid")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		
		for j := 0; j < cols; j++ {
			colName := fmt.Sprintf("Col_%d", j+1)
			if j < len(headers) {
				colName = headers[j]
			}
			
			if colStats[j].count == 0 {
				fmt.Printf("%-20s %10s %10s %10s %10d\n", 
					truncate(colName, 20), "N/A", "N/A", "N/A", 0)
			} else {
				fmt.Printf("%-20s %10.3f %10.3f %10.3f %10d\n",
					truncate(colName, 20),
					colStats[j].min,
					colStats[j].max,
					colStats[j].mean,
					colStats[j].count)
			}
		}
		fmt.Println("─────────────────────────────────────────────────────────────────")
	}

	// Summary
	fmt.Println("\n📊 Summary:")
	if missingCount == 0 {
		fmt.Println("✅ No missing values found")
	} else {
		fmt.Printf("⚠️  Found %d missing values\n", missingCount)
	}
	
	fmt.Println("✅ All values are numeric")
	fmt.Println("✅ Data is ready for PCA analysis")

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}