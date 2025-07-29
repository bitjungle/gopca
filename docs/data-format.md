# GoPCA Data Format Guide

## Overview

GoPCA accepts tabular data in CSV (Comma-Separated Values) format with specific conventions for organizing your data. This guide explains the expected format, supported features, and best practices for preparing your data for PCA analysis.

## Basic Structure

### Required Elements

1. **Header Row**: The first row must contain column names representing your variables/features
2. **Data Matrix**: Subsequent rows contain numeric data values for each sample

### Example Structure

```csv
Sample,Feature1,Feature2,Feature3,Category
Sample1,1.23,4.56,7.89,TypeA
Sample2,2.34,5.67,8.90,TypeB
Sample3,3.45,6.78,9.01,TypeA
```

## File Format Details

### CSV Separators

GoPCA automatically detects and supports multiple CSV formats:

- **Comma-separated (`,`)**: Standard CSV format
  - Uses period (`.`) as decimal separator
  - Example: `1.23,4.56,7.89`

- **Semicolon-separated (`;`)**: European CSV format
  - Uses comma (`,`) as decimal separator
  - Example: `1,23;4,56;7,89`

The parser automatically detects the format by analyzing the first few lines of your file.

### Row Names (Sample Identifiers)

The first column is automatically detected as row names if it contains non-numeric values. These identifiers:
- Appear as labels in scores plots
- Help identify outliers or interesting samples
- Should be unique for each sample
- Can contain any text (avoid special characters that might interfere with CSV parsing)

Example:
```csv
SampleID,Height,Weight,Age
Patient001,175.5,72.3,34
Patient002,168.2,65.1,28
Control001,172.0,70.5,31
```

## Column Types

GoPCA automatically detects and handles different types of columns:

### 1. Numeric Features (Default)

Standard numeric columns used in PCA calculation:
- Contain floating-point or integer values
- Can include scientific notation (e.g., `1.23e-4`)
- Form the main data matrix for PCA

### 2. Categorical Variables

Columns with string values are automatically detected as categorical:
- **Excluded from PCA calculation** (non-numeric data)
- **Available for plot coloring** using qualitative color palettes
- Useful for visualizing group membership or classifications

Example:
```csv
Sample,Gene1,Gene2,Gene3,Treatment,Batch
S1,12.3,45.6,78.9,Control,Batch1
S2,13.4,46.7,79.0,Treated,Batch1
S3,11.2,44.5,77.8,Control,Batch2
```

Both `Treatment` and `Batch` would be available as categorical coloring options.

### 3. Target Variables

Numeric columns marked as targets are treated specially:
- **Excluded from PCA calculation** (like dependent variables in regression)
- **Available for plot coloring** using sequential/gradient color palettes
- Perfect for visualizing continuous outcomes or responses

#### Marking Target Columns

Target columns are identified by the `#target` suffix in the column name:

```csv
Sample,Feature1,Feature2,Feature3,Response#target
S1,1.2,3.4,5.6,0.95
S2,2.3,4.5,6.7,0.87
S3,3.4,5.6,7.8,0.73
```

Alternative naming (with space):
```csv
Sample,Feature1,Feature2,Feature3,Response #target
```

Common use cases for target variables:
- Regression targets (y values)
- Continuous phenotypes
- Measurement outcomes
- Quality scores
- Time points

## Missing Values

GoPCA recognizes several representations of missing data:
- Empty cells
- `NA` or `na`
- `NaN` or `nan`
- `NULL` or `null`

Missing value handling strategies:
1. **Error** (default): Report an error if missing values are found
2. **Drop rows**: Remove samples with any missing values
3. **Impute mean**: Replace with column mean
4. **Impute median**: Replace with column median
5. **NIPALS native**: Use NIPALS algorithm's built-in missing value handling

## Special Values

The parser correctly handles:
- **Infinity**: `Inf`, `inf`, `+Inf`, `-Inf`
- **Scientific notation**: `1.23e-10`, `5.67E+5`
- **Very large/small numbers**: Within floating-point limits

## Data Preparation Best Practices

### 1. Variable Scaling

Consider your data scale:
- Variables with vastly different scales may dominate the PCA
- GoPCA offers preprocessing options (standardization, robust scaling)
- For spectroscopic data, consider SNV or vector normalization

### 2. Sample Size

- Minimum: More samples than variables (n > p)
- Recommended: At least 3-5 samples per variable
- For stable results: 10+ samples per variable

### 3. Column Naming

Use descriptive, valid column names:
- Avoid special characters: `<>:"/\|?*`
- Use underscores or camelCase: `Gene_Expression` or `geneExpression`
- Keep names reasonably short for better visualization
- Add `#target` suffix for target variables

### 4. Data Quality

Before analysis:
- Check for and handle outliers appropriately
- Verify measurement units are consistent
- Ensure proper decimal separator usage
- Remove or impute missing values as needed

## Example Files

### 1. Gene Expression Data
```csv
Sample,BRCA1,BRCA2,TP53,EGFR,Subtype,Survival#target
P001,5.23,3.45,7.89,2.34,Basal,24.5
P002,4.12,3.89,8.23,1.98,Luminal,48.2
P003,5.67,3.12,7.45,2.56,Basal,18.7
```

### 2. Spectroscopic Data
```csv
Wavelength,400nm,450nm,500nm,550nm,600nm,Concentration#target
Sample1,0.234,0.456,0.678,0.543,0.321,1.5
Sample2,0.245,0.467,0.689,0.554,0.332,1.8
Sample3,0.223,0.445,0.667,0.532,0.310,1.2
```

### 3. Mixed Data Types
```csv
ID,Height,Weight,Age,BMI,Gender,Group,Disease_Score#target
S001,175.5,72.3,34,23.5,M,Control,0.0
S002,162.3,58.7,28,22.3,F,Treatment,2.5
S003,180.2,85.1,45,26.2,M,Treatment,3.8
```

## Validation

Use the GoPCA CLI to validate your data format:

```bash
gopca-cli validate yourdata.csv
```

This will report:
- Data dimensions
- Detected column types
- Missing value locations
- Categorical columns (excluded from PCA)
- Target columns (excluded from PCA)
- Any format issues

## Summary

For successful PCA analysis with GoPCA:

1. ✅ Include a header row with column names
2. ✅ First column can be sample identifiers
3. ✅ Use consistent CSV format (comma or semicolon)
4. ✅ Ensure numeric data for PCA features
5. ✅ Mark target columns with `#target` suffix
6. ✅ Categorical columns are auto-detected for visualization
7. ✅ Handle missing values appropriately
8. ✅ Validate your data before analysis

Following these guidelines will ensure smooth data import and meaningful PCA results with full visualization capabilities in GoPCA.