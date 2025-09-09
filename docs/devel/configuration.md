# GoPCA Suite Configuration Guide

GoPCA Suite provides configuration options to customize various aspects of the application behavior. While most users won't need to modify these settings, they are available for advanced use cases.

## Configuration Structure

The application uses sensible defaults for all configuration options. Currently, configuration is implemented through code constants and structures, with the foundation laid for future external configuration file support.

## pca CLI Configuration

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
- **Loadings Variable Threshold**: Maximum number of variables to show in loadings plot (default: 100)
- **Biplot Max Variables**: Maximum variables to display in biplot (default: 100)
- **Circle Max Variables**: Maximum variables in circle of correlations (default: 100)
- **Correlation Threshold**: Minimum correlation to display in circle of correlations (default: 0.3)
- **Elbow Threshold**: Threshold for elbow detection in scree plot (default: 80.0)
- **Mahalanobis Threshold**: Outlier detection threshold for Mahalanobis distance (default: 3.0)
- **RSS Threshold**: Outlier detection threshold for residual sum of squares (default: 0.03)
- **Default Confidence Level**: Confidence level for ellipses (default: 0.95)
- **Max Kernel Matrix Samples**: Maximum samples for kernel matrix visualization (default: 1000)

### UI
- **Data Preview Max Rows**: Maximum rows in data table preview (default: 10)
- **Data Preview Max Columns**: Maximum columns in data table preview (default: 10)
- **Default Zoom Factor**: Zoom increment for plot controls (default: 0.8)

## Security and Resource Limits

These limits are defined in `pkg/security/input_validation.go` to prevent resource exhaustion and ensure memory safety:

### File and Data Limits
- **Max File Size**: 500MB maximum file size for uploads
- **Max CSV Rows**: 1,000,000 rows maximum
- **Max CSV Columns**: 10,000 columns maximum
- **Max Field Length**: 100,000 characters per field
- **Max String Length**: 10,000 characters for general strings
- **Max Path Length**: 4096 characters (standard PATH_MAX)

### PCA Analysis Limits
- **Max Components**: 1000 maximum PCA components
- **Min Components**: 1 minimum PCA component
- **Max Kernel PCA Samples**: 10,000 maximum samples for Kernel PCA (memory safety)
- **Max Kernel Matrix Visualization**: 1000 maximum samples for kernel matrix heatmap visualization
- **Max Iterations**: 10,000 maximum iterations for iterative algorithms

### Kernel PCA Parameters
- **Max Kernel Gamma**: 1e6 maximum gamma value
- **Min Kernel Gamma**: 1e-6 minimum gamma value

### Memory Limits
- **Max Memory Usage**: 2GB maximum memory for operations

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

For developers, configuration structures and security constants are available:

### Configuration Structures

```go
import "github.com/bitjungle/gopca/internal/config"

// Get default CLI configuration
cliConfig := config.DefaultConfig()

// Get default GUI configuration
guiConfig := config.DefaultGUIConfig()
```

### Security Constants

```go
import "github.com/bitjungle/gopca/pkg/security"

// Access security limits
maxSamples := security.MaxKernelPCASamples
maxVisualizationSamples := security.MaxKernelMatrixVisualization
maxFileSize := security.MaxFileSize
```

### Example: Checking Kernel Matrix Visualization Limit

```go
import (
    "github.com/bitjungle/gopca/pkg/security"
    "github.com/bitjungle/gopca/internal/config"
)

// Using the security constant directly
if nSamples <= security.MaxKernelMatrixVisualization {
    // Include kernel matrix for visualization
}

// Or accessing through GUI config
guiConfig := config.DefaultGUIConfig()
maxSamples := guiConfig.Visualization.MaxKernelMatrixSamples
```