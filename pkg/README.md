# GoPCA Shared Packages

This directory contains shared packages used by both GoPCA and GoCSV applications. These packages implement common functionality to ensure consistency and reduce code duplication.

## Package Structure

### `pkg/types/`
Core types and data structures shared across applications.

- **`json.go`**: Provides `JSONFloat64` type for safe JSON marshaling of special float values (NaN, Inf)
  - Marshals NaN and Inf as `null` for JavaScript compatibility
  - Unmarshals `null` as NaN
  - Includes helper methods: `Float64()`, `IsNaN()`, `IsInf()`

### `pkg/utils/`
Common utility functions for data processing.

- **`missing.go`**: Missing value detection and handling
  - `DefaultMissingValues()`: Returns standard missing value indicators
  - `IsMissingValue()`: Checks if a string represents missing data
  - `ContainsMissingValues()`: Checks if a slice contains missing values
  - `CountMissingValues()`: Counts missing values in a slice

- **`parsing.go`**: Numeric parsing utilities
  - `ParseNumericValue()`: Parses strings to float64 with decimal separator support
  - `ParseNumericValueWithMissing()`: Combines parsing with missing value detection
  - `IsNumericString()`: Checks if a string can be parsed as a number
  - `ParseFloatSlice()`: Parses a slice of strings to float64 values

## Usage Examples

### JSON-Safe Float Handling
```go
import "github.com/bitjungle/gopca/pkg/types"

// Create a JSON-safe float
value := types.JSONFloat64(math.NaN())

// Marshal to JSON (becomes "null")
data, _ := json.Marshal(value)

// Check if value is NaN
if value.IsNaN() {
    // Handle NaN case
}
```

### Missing Value Detection
```go
import "github.com/bitjungle/gopca/pkg/utils"

// Use default missing indicators
missingIndicators := utils.DefaultMissingValues()

// Check if a value is missing
if utils.IsMissingValue("NA", missingIndicators) {
    // Handle missing value
}

// Count missing values in data
count := utils.CountMissingValues(dataSlice, missingIndicators)
```

### Numeric Parsing
```go
import "github.com/bitjungle/gopca/pkg/utils"

// Parse with decimal separator
value, err := utils.ParseNumericValue("123,45", ',')

// Parse with missing value detection
val, isMissing, err := utils.ParseNumericValueWithMissing(
    "NA", '.', utils.DefaultMissingValues())
if isMissing {
    // Value is missing (val = NaN)
}
```

## Design Principles

1. **Consistency**: Both applications handle data the same way
2. **Robustness**: Comprehensive error handling and edge case coverage
3. **Performance**: Efficient implementations with benchmarks
4. **Testability**: High test coverage (>90%) with table-driven tests
5. **Simplicity**: Clear, single-purpose functions following KISS principle

## Testing

Run tests for shared packages:
```bash
go test ./pkg/...
```

Run with coverage:
```bash
go test -cover ./pkg/...
```

## Contributing

When modifying shared packages:
1. Ensure changes don't break existing functionality
2. Add comprehensive tests for new features
3. Update documentation as needed
4. Test both GoPCA and GoCSV after changes
5. Follow the project's coding standards from CLAUDE.md