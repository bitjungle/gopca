# CSV Format Implementation Guide

## Overview

This document describes the internal CSV format handling in GoPCA, implemented through the unified `pkg/csv` package. This package consolidates all CSV parsing, writing, and validation logic across the monorepo.

## Architecture

### Package Structure

```
pkg/csv/
├── types.go       # Common interfaces and types
├── reader.go      # Unified CSV reading functionality
├── writer.go      # Unified CSV writing functionality
├── validation.go  # CSV validation and statistics
└── *_test.go      # Comprehensive test coverage
```

### Design Principles

1. **DRY (Don't Repeat Yourself)**: Single source of truth for CSV operations
2. **Flexibility**: Support multiple CSV formats and data types
3. **Backward Compatibility**: Existing code can use the package with minimal changes
4. **Performance**: Efficient handling of large datasets with streaming support

## Core Types

### Options

The `Options` struct provides unified configuration for all CSV operations:

```go
type Options struct {
    // Parsing options
    Delimiter        rune      // Field delimiter: ',', ';', '\t'
    DecimalSeparator rune      // Decimal separator: '.', ','
    HasHeaders       bool      // First row contains column names
    HasRowNames      bool      // First column contains row names
    NullValues       []string  // Strings to treat as missing values
    ParseMode        ParseMode // How to parse the data
    TargetSuffix     string    // Suffix to identify target columns (e.g., "#target")
    
    // Reading options (for large files)
    SkipRows      int   // Number of rows to skip at start
    MaxRows       int   // Maximum rows to read (0 for all)
    Columns       []int // Specific columns to read (empty for all)
    StreamingMode bool  // Enable streaming for large files
    
    // Writing options
    FloatFormat byte // Format for float output: 'g', 'f', 'e'
    Precision   int  // Decimal precision (-1 for auto)
}
```

### ParseMode

Different parsing modes for various use cases:

```go
const (
    ParseNumeric          // Treat all data as numeric values
    ParseString           // Treat all data as strings (for GoCSV)
    ParseMixed            // Automatically detect column types
    ParseMixedWithTargets // Detect columns and identify target columns
)
```

### Data Structure

The unified `Data` struct supports multiple data representations:

```go
type Data struct {
    // Core numeric data (always present for PCA)
    Matrix      types.Matrix // Numeric data matrix
    Headers     []string     // Column names
    RowNames    []string     // Row names
    MissingMask [][]bool     // Track missing values
    Rows        int          // Number of data rows
    Columns     int          // Number of data columns
    
    // Additional data types (optional)
    StringData           [][]string            // Raw string data (for GoCSV)
    CategoricalColumns   map[string][]string   // Categorical columns by name
    NumericTargetColumns map[string][]float64  // Numeric target columns
}
```

## Supported CSV Formats

### Standard Formats

1. **US/International (Default)**
   - Field delimiter: `,`
   - Decimal separator: `.`
   - Example: `1.23,4.56,7.89`

2. **European**
   - Field delimiter: `;`
   - Decimal separator: `,`
   - Example: `1,23;4,56;7,89`

3. **Tab-Delimited (TSV)**
   - Field delimiter: `\t`
   - Decimal separator: `.`
   - Example: `1.23	4.56	7.89`

### Format Detection

The package can automatically detect CSV format by trying multiple formats in sequence:

```go
formats := []Options{
    DefaultOptions(),        // Comma with dot decimal
    EuropeanOptions(),       // Semicolon with comma decimal
    TabDelimitedOptions(),   // Tab delimited
}
```

## Column Type Detection

### Numeric Columns
- Contain parseable floating-point or integer values
- Support scientific notation (e.g., `1.23e-4`)
- Special values: `Inf`, `-Inf`, `NaN`
- Used for PCA calculation

### Categorical Columns
- Contain non-numeric string values
- Automatically excluded from PCA calculation
- Available for visualization (plot coloring)
- Stored in `CategoricalColumns` map

### Target Columns
- Numeric columns with `#target` suffix in header
- Excluded from PCA calculation (like dependent variables)
- Available for continuous value visualization
- Stored in `NumericTargetColumns` map

### Missing Values

Recognized representations:
- Empty cells
- `NA`, `N/A`
- `NaN`, `nan`
- `NULL`, `null`
- `m` (legacy support)
- Custom values via `NullValues` option

## Usage Examples

### Reading CSV Files

```go
// Simple numeric parsing
opts := csv.DefaultOptions()
reader := csv.NewReader(opts)
data, err := reader.ReadFile("data.csv")

// Mixed data with target detection
opts := csv.DefaultOptions()
opts.ParseMode = csv.ParseMixedWithTargets
reader := csv.NewReader(opts)
data, err := reader.ReadFile("mixed_data.csv")

// European format
opts := csv.EuropeanOptions()
reader := csv.NewReader(opts)
data, err := reader.ReadFile("european.csv")
```

### Writing CSV Files

```go
// Write numeric matrix
opts := csv.DefaultOptions()
writer := csv.NewWriter(opts)
err := writer.WriteMatrixFile("output.csv", matrix, headers, rowNames)

// Write with European format
opts := csv.EuropeanOptions()
writer := csv.NewWriter(opts)
err := writer.WriteFile("output.csv", data)
```

### Validation

```go
// Validate CSV file
validator := csv.NewValidator(csv.DefaultOptions())
result := validator.Validate(data)

if !result.Valid {
    for _, err := range result.Errors {
        fmt.Println("Error:", err)
    }
}

// Get column statistics
for _, stats := range result.ColumnStats {
    fmt.Printf("Column %s: %.1f%% missing, mean=%.2f, std=%.2f\n",
        stats.Name, stats.MissingPercent, stats.Mean, stats.StdDev)
}
```

## Migration Guide

### From internal/cli/csv_parser.go

Before:
```go
import "github.com/bitjungle/gopca/internal/cli"

opts := cli.NewCSVParseOptions()
data, err := cli.ParseCSV(filename, opts)
```

After:
```go
import pkgcsv "github.com/bitjungle/gopca/pkg/csv"

opts := pkgcsv.DefaultOptions()
opts.ParseMode = pkgcsv.ParseMixed
reader := pkgcsv.NewReader(opts)
data, err := reader.ReadFile(filename)
```

### From internal/io/csv.go

Before:
```go
import "github.com/bitjungle/gopca/internal/io"

opts := io.DefaultCSVOptions()
matrix, headers, err := io.LoadCSV(filename, opts)
```

After:
```go
import pkgcsv "github.com/bitjungle/gopca/pkg/csv"

opts := pkgcsv.DefaultOptions()
reader := pkgcsv.NewReader(opts)
data, err := reader.ReadFile(filename)
matrix, headers := data.Matrix, data.Headers
```

### From cmd/gopca-desktop/parse_csv.go

Before:
```go
import "github.com/bitjungle/gopca/pkg/types"

format := types.DefaultCSVFormat()
data, catData, targetData, err := types.ParseCSVMixedWithTargets(reader, format, nil)
```

After:
```go
import pkgcsv "github.com/bitjungle/gopca/pkg/csv"

opts := pkgcsv.DefaultOptions()
opts.ParseMode = pkgcsv.ParseMixedWithTargets
reader := pkgcsv.NewReader(opts)
data, err := reader.Read(input)
// catData is now in data.CategoricalColumns
// targetData is now in data.NumericTargetColumns
```

## Performance Considerations

### Memory Optimization
- Use `StreamingMode` for large files
- Specify `Columns` to read only required columns
- Use `MaxRows` to limit data for preview/validation

### Parsing Performance
- Numeric parsing is optimized with minimal allocations
- Missing value detection uses map lookup for O(1) performance
- Decimal separator conversion is done in-place when possible

## Testing

The package includes comprehensive tests covering:
- All supported CSV formats
- Missing value handling
- Edge cases (empty files, inconsistent columns)
- Format auto-detection
- Large file handling
- Special values (Inf, NaN)

Run tests:
```bash
go test ./pkg/csv/...
```

## Future Enhancements

Planned improvements:
1. Full streaming support for very large files
2. Parallel parsing for multi-core systems
3. Automatic format detection improvements
4. Support for quoted fields with embedded delimiters
5. Memory-mapped file support for huge datasets

## API Stability

The `pkg/csv` package is considered stable with these guarantees:
- No breaking changes to exported types and functions
- New features added through additional options
- Deprecated functions kept for backward compatibility
- Clear migration paths for any necessary changes

## References

- [RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html) - Common Format and MIME Type for CSV Files
- [CSV on Wikipedia](https://en.wikipedia.org/wiki/Comma-separated_values)
- [Go encoding/csv package](https://pkg.go.dev/encoding/csv)