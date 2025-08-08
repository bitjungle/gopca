# GoPCA CLI Reference

Complete command-line interface documentation for pca.

## Overview

The GoPCA command-line interface (`pca`) provides powerful PCA analysis capabilities for automation, batch processing, and integration into data pipelines. It supports multiple PCA algorithms, comprehensive preprocessing options, and flexible output formats.

## Installation

Download the latest binary for your platform from the [GitHub Releases](https://github.com/bitjungle/gopca/releases) page:

```bash
# Linux/macOS
wget https://github.com/bitjungle/gopca/releases/latest/download/pca
chmod +x pca

# Add to PATH (optional)
sudo mv pca /usr/local/bin/
```

## Global Options

These options apply to all commands:

- `--help, -h` - Show help for any command
- `--version` - Display version information

## Commands

### `analyze` - Perform PCA Analysis

The main command for running PCA analysis on your data.

#### Basic Usage

```bash
pca analyze [OPTIONS] <input.csv>
```

**Important:** The input CSV file must be specified as the last argument. All options must come before the filename.

#### Options

##### General Options
- `--verbose, -v` - Enable verbose output with detailed progress
- `--quiet, -q` - Minimal output, suitable for scripting
- `--output-dir, -o <path>` - Output directory (default: same as input file)
- `--format, -f <format>` - Output format: `table` or `json` (default: `table`)

##### PCA Configuration
- `--components, -c <n>` - Number of principal components (default: 2)
- `--method <method>` - PCA algorithm: `svd`, `nipals`, or `kernel` (default: `svd`)
  - `svd` - Singular Value Decomposition (fastest, requires complete data)
  - `nipals` - Nonlinear Iterative Partial Least Squares (handles missing data)
  - `kernel` - Kernel PCA for non-linear relationships

##### Preprocessing Options
- `--no-mean-centering` - Disable mean centering
- `--scale <method>` - Scaling method:
  - `none` - No scaling (default)
  - `standard` - Standardize to unit variance
  - `robust` - Robust scaling using median and MAD
- `--scale-only` - Apply variance scaling without mean centering (useful for Kernel PCA)
- `--snv` - Apply Standard Normal Variate (row-wise normalization)
- `--vector-norm` - Apply L2 vector normalization (row-wise)

##### Kernel PCA Options
- `--kernel-type <type>` - Kernel type: `rbf`, `linear`, or `poly`
- `--kernel-gamma <value>` - Gamma parameter for RBF and polynomial kernels (default: 1)
- `--kernel-degree <n>` - Degree for polynomial kernel (default: 3)
- `--kernel-coef0 <value>` - Independent term for polynomial kernel (default: 0)

##### Data Format Options
- `--no-headers` - First row contains data, not column names
- `--no-index` - First column contains data, not row names
- `--delimiter <char>` - CSV delimiter: `comma`, `semicolon`, or `tab` (default: `comma`)
- `--decimal-separator <sep>` - Decimal separator: `dot` or `comma` (default: `dot`)
- `--na-values <list>` - Comma-separated strings representing missing values
  - Default: `"NA,N/A,nan,NaN,null,NULL"`

##### Missing Data Handling
- `--missing-strategy <strategy>` - How to handle missing values:
  - `error` - Fail if missing values found (default)
  - `drop` - Remove rows with missing values
  - `mean` - Replace with column mean
  - `median` - Replace with column median
  - `native` - Use NIPALS algorithm's native missing data handling

##### Data Selection
- `--exclude-rows <list>` - Exclude rows by index (1-based, e.g., '1,3,5-7')
- `--exclude-cols <list>` - Exclude columns by index (1-based, e.g., '2,4-6,8')

##### Group and Correlation Analysis
- `--group-column <name>` - Categorical column for grouping samples
- `--metadata-cols <list>` - Columns for eigencorrelation analysis
- `--target-columns <list>` - Target columns (auto-detected if ending with `#target`)
- `--eigencorrelations` - Calculate correlations between PCs and metadata/target

##### Output Control
- `--output-scores` - Include PC scores (default: true)
- `--output-loadings` - Include loadings (default: false)
- `--output-variance` - Include explained variance (default: false)
- `--output-all` - Output all results
- `--include-metrics` - Include diagnostic metrics (T², Mahalanobis, RSS)

#### Examples

##### Basic Analysis
```bash
# Simple 2-component PCA with default settings
pca analyze data.csv

# 3 components with standard scaling
pca analyze --components 3 --scale standard data.csv

# Save results to specific directory
pca analyze -o results/ data.csv
```

##### Advanced Preprocessing
```bash
# SNV preprocessing for spectroscopic data
pca analyze --snv --scale standard spectral_data.csv

# Robust scaling for data with outliers
pca analyze --scale robust --components 4 data.csv

# Vector normalization
pca analyze --vector-norm data.csv
```

##### Kernel PCA
```bash
# RBF kernel with custom gamma
pca analyze --method kernel --kernel-type rbf --kernel-gamma 0.5 data.csv

# Polynomial kernel of degree 3
pca analyze --method kernel --kernel-type poly --kernel-degree 3 data.csv

# Linear kernel PCA
pca analyze --method kernel --kernel-type linear data.csv
```

##### Missing Data
```bash
# Drop rows with missing values
pca analyze --missing-strategy drop data.csv

# Use NIPALS with native missing data handling
pca analyze --method nipals --missing-strategy native data.csv

# Replace missing with mean
pca analyze --missing-strategy mean data.csv
```

##### Group Analysis
```bash
# Specify grouping column
pca analyze --group-column sample_type data.csv

# Calculate eigencorrelations with metadata
pca analyze --metadata-cols age,weight --eigencorrelations -f json data.csv

# Include target columns
pca analyze --target-columns concentration,pH --eigencorrelations data.csv
```

##### Output Formats
```bash
# JSON output with all results
pca analyze -f json --output-all data.csv

# Table format with scores and variance
pca analyze --output-scores --output-variance data.csv

# Include diagnostic metrics
pca analyze --include-metrics -f json data.csv
```

### `validate` - Validate Input Data

Check your data for issues before running PCA analysis.

#### Basic Usage

```bash
pca validate [OPTIONS] <input.csv>
```

#### Options

- `--no-headers` - First row contains data, not column names
- `--no-index` - First column contains data, not row names
- `--delimiter <char>` - CSV delimiter (default: comma)
- `--na-values <list>` - Strings representing missing values
- `--strict` - Fail on warnings (not just errors)
- `--summary` - Show data summary statistics

#### Validation Checks

The validate command performs these checks:
- File format and structure validation
- Missing values detection and reporting
- Data type consistency verification
- Numerical range checks
- Low variance detection (constant columns)
- High missing value warnings (>50% missing)

#### Examples

```bash
# Basic validation
pca validate data.csv

# Show detailed summary statistics
pca validate --summary data.csv

# Strict mode - fail on any warnings
pca validate --strict data.csv

# Custom delimiter and missing values
pca validate --delimiter semicolon --na-values "?,unknown" data.csv
```

### `transform` - Apply PCA Model to New Data

Apply a previously trained PCA model to transform new data.

#### Basic Usage

```bash
pca transform [OPTIONS] <model.json> <input.csv>
```

**Note:** Both the model file and input CSV must be specified as the last two arguments.

#### Options

- `--output-dir, -o <path>` - Output directory for results
- `--format, -f <format>` - Output format: `table` or `json`
- `--verbose` - Show detailed progress
- `--quiet, -q` - Suppress output except errors
- `--no-headers` - Input CSV has no header row
- `--no-index` - Input CSV has no index column
- `--delimiter, -d <char>` - CSV delimiter
- `--decimal-separator <sep>` - Decimal separator
- `--na-values <list>` - Missing value strings
- `--exclude-rows <list>` - Row indices to exclude
- `--include-metrics` - Calculate diagnostic metrics

#### Requirements

- New data must have the same number of features as training data
- Column names should match for proper alignment
- Preprocessing from training is automatically applied
- Currently supports SVD and NIPALS models

#### Examples

```bash
# Basic transformation
pca transform model.json new_data.csv

# Save to specific file with JSON format
pca transform -f json -o results/ model.json new_data.csv

# Exclude specific rows
pca transform --exclude-rows 1,5-10 model.json new_data.csv

# Include diagnostic metrics
pca transform --include-metrics model.json new_data.csv
```

## Output Formats

### Table Format (Default)

Human-readable tabular output displayed in the terminal:

```
Principal Component Scores:
Sample      PC1        PC2        PC3
------      ---        ---        ---
Sample1     2.345     -0.123      1.456
Sample2    -1.234      0.987     -0.543
...

Explained Variance:
Component   Variance   Cumulative
---------   --------   ----------
PC1         45.23%     45.23%
PC2         23.45%     68.68%
PC3         12.34%     81.02%
```

### JSON Format

Machine-readable JSON output for integration with other tools:

```json
{
  "scores": [[2.345, -0.123, 1.456], [-1.234, 0.987, -0.543]],
  "loadings": [[0.234, -0.567], [0.123, 0.456]],
  "explainedVariance": [0.4523, 0.2345, 0.1234],
  "cumulativeVariance": [0.4523, 0.6868, 0.8102],
  "eigenvalues": [5.234, 2.456, 1.234],
  "rowNames": ["Sample1", "Sample2"],
  "columnNames": ["Feature1", "Feature2"],
  "metrics": {
    "hotellingT2": [1.234, 0.567],
    "mahalanobis": [2.345, 1.234],
    "rss": [0.012, 0.023]
  }
}
```

## Input Data Format

### CSV Requirements

- **Headers**: First row should contain column names (unless `--no-headers`)
- **Index**: First column can contain row names (unless `--no-index`)
- **Numeric Data**: All data columns must be numeric
- **Delimiters**: Comma (default), semicolon, or tab
- **Missing Values**: Use standard representations (NA, NaN, null, etc.)

### Example CSV Structure

```csv
Sample,Feature1,Feature2,Feature3,GroupLabel
Sample1,1.23,4.56,7.89,TypeA
Sample2,2.34,5.67,8.90,TypeB
Sample3,3.45,6.78,9.01,TypeA
```

### Special Columns

- **Group Columns**: Categorical columns for sample grouping
- **Target Columns**: Columns ending with `#target` are automatically detected
- **Metadata Columns**: Additional columns for correlation analysis

## Preprocessing Pipeline

The preprocessing steps are applied in this order:

1. **Row-wise preprocessing** (if enabled):
   - SNV (Standard Normal Variate)
   - L2 Vector Normalization

2. **Column-wise preprocessing**:
   - Mean Centering (unless disabled)
   - Scaling (standard, robust, or none)

3. **Algorithm-specific processing**:
   - SVD: Requires complete data after preprocessing
   - NIPALS: Can handle missing values natively
   - Kernel: Applies kernel transformation

## Best Practices

### Data Preparation
1. Validate your data first: `pca validate data.csv`
2. Handle missing values appropriately for your domain
3. Consider scaling for mixed-unit variables
4. Use SNV for spectroscopic data

### Algorithm Selection
- **SVD**: Fast and accurate for complete data
- **NIPALS**: When you have missing values
- **Kernel PCA**: For non-linear relationships

### Preprocessing Choices
- **No scaling**: When all variables are in same units
- **Standard scaling**: Mixed units or scales
- **Robust scaling**: Data contains outliers
- **SNV**: Spectroscopic or similar data

### Performance Tips
- Use `--quiet` for scripting and automation
- JSON format is faster to parse programmatically
- Exclude unnecessary columns to reduce memory usage
- Pre-filter rows if analyzing subsets

## Troubleshooting

### Common Issues

#### "Invalid CSV format"
- Check delimiter matches your file
- Ensure consistent column count across rows
- Verify decimal separator setting

#### "Insufficient numeric columns"
- Exclude non-numeric columns with `--exclude-cols`
- Check for columns with all missing values
- Ensure proper NA value detection

#### "Memory allocation failed"
- Reduce number of components
- Exclude unnecessary columns/rows
- Use NIPALS instead of SVD for large sparse data

#### "Convergence not achieved" (NIPALS)
- Increase max iterations (if option available)
- Check for extreme outliers
- Consider different preprocessing

## Integration Examples

### Bash Pipeline
```bash
#!/bin/bash
# Batch process multiple files
for file in data/*.csv; do
    pca analyze -f json -o results/ "$file"
done
```

### Python Integration
```python
import subprocess
import json

# Run PCA analysis
result = subprocess.run(
    ['pca', 'analyze', '-f', 'json', '--output-all', 'data.csv'],
    capture_output=True, text=True
)

# Parse JSON results
pca_results = json.loads(result.stdout)
scores = pca_results['scores']
variance = pca_results['explainedVariance']
```

### R Integration
```r
# Run pca from R
library(jsonlite)

output <- system2(
  "pca",
  args = c("analyze", "-f", "json", "--output-all", "data.csv"),
  stdout = TRUE
)

# Parse results
results <- fromJSON(paste(output, collapse = "\n"))
scores <- results$scores
```

## See Also

- [Introduction to PCA](intro_to_pca.md) - Understanding PCA theory
- [Data Format Guide](data-format.md) - Detailed CSV format specification
- [Data Preparation Guide](intro_to_data_prep.md) - Best practices for data preparation

## Getting Help

```bash
# General help
pca --help

# Command-specific help
pca analyze --help
pca validate --help
pca transform --help
```

For issues or questions, visit the [GitHub repository](https://github.com/bitjungle/gopca).