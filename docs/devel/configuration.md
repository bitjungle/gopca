# GoPCA Configuration Guide

GoPCA provides configuration options to customize various aspects of the application behavior. While most users won't need to modify these settings, they are available for advanced use cases.

## Configuration Structure

The application uses sensible defaults for all configuration options. Currently, configuration is implemented through code constants and structures, with the foundation laid for future external configuration file support.

## CLI Configuration

### CSV Parsing
- **Type Detection Sample Size**: Number of rows sampled to determine column types (default: 10)
- **Default Null Values**: Strings recognized as missing values (default: "", "NA", "N/A", "null", "NULL", "NaN", "nan")

### Output
- **File Suffix**: Suffix added to output filenames (default: "_pca")
- **Create Output Directory**: Automatically create output directory if it doesn't exist (default: true)

### Analysis
- **Default Components**: Number of components when not specified (default: 0, auto-detect)
- **Show Preview**: Display preview of transformed data (default: true)
- **Preview Max Rows**: Maximum rows shown in preview (default: 10)

## GUI Configuration

### Visualization
- **Loadings Variable Threshold**: Percentage of variables to show in loadings plot (default: 50)
- **Correlation Threshold**: Minimum correlation to display in circle of correlations (default: 0.3)
- **Elbow Threshold**: Threshold for elbow detection in scree plot (default: 80%)
- **Mahalanobis Threshold**: Outlier detection threshold for Mahalanobis distance (default: 3.0)
- **RSS Threshold**: Outlier detection threshold for residual sum of squares (default: 0.03)
- **Default Confidence Level**: Confidence level for ellipses (default: 0.95)

### UI
- **Data Preview Max Rows**: Maximum rows in data table preview (default: 10)
- **Data Preview Max Columns**: Maximum columns in data table preview (default: 10)
- **Default Zoom Factor**: Zoom increment for plot controls (default: 0.8)

## Algorithm Parameters

These parameters are intentionally kept as internal constants as they represent well-tested values:

- **NIPALS Convergence Tolerance**: 1e-8
- **NIPALS Max Iterations**: 1000
- **Minimum Variance Threshold**: 1e-8

## Future Enhancements

The configuration infrastructure is designed to support future features:
- External configuration files (JSON/YAML)
- Environment variable overrides
- User-specific configuration
- Per-project configuration

## Programmatic Access

For developers, configuration structures are available in the `internal/config` package:

```go
import "github.com/bitjungle/gopca/internal/config"

// Get default CLI configuration
cliConfig := config.DefaultConfig()

// Get default GUI configuration
guiConfig := config.DefaultGUIConfig()
```