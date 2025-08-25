# GoPCA JSON Schemas

JSON Schema definitions for GoPCA's data export/import formats are located in the `schemas/` directory.

## Overview

GoPCA uses JSON for model serialization to ensure interoperability between different applications and third-party tools. These schemas provide formal definitions and validation for all JSON formats used in the GoPCA ecosystem.

## Schema Version

Current version: **v1**

The schemas follow JSON Schema Draft-07 specification.

## Available Schemas

### Core Schema (`v1/pca-output.schema.json`)
The main schema for complete PCA analysis results. This is the primary format used when:
- Exporting models from GoPCA Desktop
- Saving results from CLI with `-f json`
- Importing models for transformation

### Component Schemas

- **`common.schema.json`** - Shared type definitions (Matrix, enums)
- **`model-metadata.schema.json`** - Model metadata and configuration
- **`preprocessing.schema.json`** - Preprocessing settings and parameters
- **`model-components.schema.json`** - PCA components (loadings, variance)
- **`results.schema.json`** - Analysis results (scores, metrics)

## Usage Examples

### CLI Validation
When using the CLI transform command, models are automatically validated:
```bash
pca transform model.json new_data.csv
```

### Programmatic Validation (Go)
```go
import "github.com/bitjungle/gopca/pkg/validation"

validator, err := validation.NewModelValidator("v1")
if err != nil {
    // Handle error
}

jsonData, _ := os.ReadFile("model.json")
if err := validator.ValidateModel(jsonData); err != nil {
    // Validation failed
    fmt.Printf("Model validation error: %v\n", err)
}
```

### Third-Party Integration
The schemas can be used to:
- Generate TypeScript interfaces for web applications
- Create Python dataclasses for data science workflows
- Validate models in R or MATLAB
- Build API specifications

## Model Structure

A valid PCA model contains four required sections and an optional schema reference:

```json
{
  "$schema": "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
  "metadata": {
    "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
    "software_version": "0.9.0",
    "created_at": "2025-01-25T10:00:00Z",
    "software": "gopca",
    "config": { ... },
    "data_source": {
      "filename": "experiment_data.csv",
      "n_rows_original": 150,
      "n_cols_original": 4
    },
    "description": "PCA analysis of experimental data",
    "tags": ["experiment-1", "quality-control"]
  },
  "preprocessing": {
    "mean_center": true,
    "standard_scale": true,
    "parameters": { ... }
  },
  "model": {
    "loadings": [[...], [...]],
    "explained_variance": [...],
    "component_labels": ["PC1", "PC2", ...]
  },
  "results": {
    "samples": {
      "names": ["Sample1", "Sample2", ...],
      "scores": [[...], [...]]
    }
  }
}
```

Optional sections include:
- `diagnostics` - Statistical limits for outlier detection
- `eigencorrelations` - Correlations with metadata variables
- `preservedColumns` - Categorical/target data preserved from analysis

## Schema Reference

The `$schema` field is automatically included in exported models to indicate which schema version the document conforms to. This enables:
- Automatic schema detection by validators
- Version compatibility checking
- Tool integration without manual schema selection

Example:
```json
"$schema": "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json"
```

## Enhanced Metadata Fields

The schema now includes enhanced metadata for better traceability:

- **`analysis_id`** (required): UUID for unique identification
- **`data_source`** (optional): Information about input data
  - `filename`: Original data file name
  - `hash`: SHA-256 hash for data integrity
  - `n_rows_original`: Rows before exclusions
  - `n_cols_original`: Columns before exclusions
- **`description`** (optional): User notes about the analysis
- **`tags`** (optional): Array of user-defined tags

## Validation Rules

The schemas enforce:
- Required fields at each level
- Type constraints (numbers, strings, arrays)
- Value ranges (e.g., correlation coefficients between -1 and 1)
- Enumerated values (e.g., PCA methods: "svd", "nipals", "kernel")
- Array dimensions (e.g., loadings must be 2D array)
- UUID format for `analysis_id`

## Version Management

Schema versions follow semantic versioning:
- **Major**: Breaking changes to structure
- **Minor**: New optional fields
- **Patch**: Documentation or constraint updates

Future versions will be placed in new directories (e.g., `v2/`) with migration guides.

## Contributing

When modifying schemas:
1. Update the schema files in `schemas/v1/`
2. Update Go types in `pkg/types/pca.go` to match
3. Run validation tests: `go test ./pkg/validation`
4. Update this documentation if needed

## References

- [JSON Schema Specification](https://json-schema.org/draft/2020-12/json-schema-core.html)
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/)
- GoPCA Types: `pkg/types/pca.go`