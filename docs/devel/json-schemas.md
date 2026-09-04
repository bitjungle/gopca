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
    "software_version": "1.0.2",
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

### How enforcement works

`pkg/validation` compiles the schema graph with `gojsonschema` and validates
every model against it. Both `pca transform` and the Desktop export path run it.

Resolution is offline. Each schema has an absolute `$id` and refers to its
neighbours by relative filename, so a `$ref` resolves against the referrer's
`$id` to the `$id` of the target. All schemas in the embedded directory are
registered before compiling, so nothing is fetched over the network.
`TestEveryRefResolvesToAnEmbeddedFile` fails on a `$ref` naming a file that is
not embedded, because such a reference would silently turn validation into a
network operation.

### Two layers, and the boundary between them

| Layer | Enforces | Lives in |
|---|---|---|
| JSON Schema | Shape: types, required keys, ranges, enums, formats | `schemas/v1/*.json` |
| Semantic checks | Agreement between fields | `validateSemantics` in `pkg/validation/schema.go` |

Draft-07 constrains each value on its own. It cannot say that `regression.components`
must equal the length of `regression.score_coefficients`, or that a model claiming
`original_scale_valid: true` must actually carry `coefficients`. Those cross-field
agreements are what make a file self-contradictory rather than merely malformed, and
they are the only things the Go code still checks by hand.

Keep the boundary: anything expressible in the schema belongs there, so that there is
one place to look for each kind of rule.

### A naming trap for third-party consumers

`explained_variance_ratio` and `cumulative_variance` are **percentages, 0 to 100** —
corn's first component is `97.495`, not `0.97495`. The name says "ratio" for historical
reasons, and scikit-learn's `explained_variance_ratio_` is a fraction of 1, so code
comparing the two must divide by 100. The schema descriptions state this; the field
names cannot be changed without a `v2`.

## Version Management

Schema versions follow semantic versioning:
- **Major**: Breaking changes to structure
- **Minor**: New optional fields
- **Patch**: Documentation or constraint updates

Future versions will be placed in new directories (e.g., `v2/`) with migration guides.

## Contributing

When modifying schemas:
1. Update the schema files in `schemas/v1/` — this is the source copy, and the one
   the `$schema` URLs name
2. Run `make sync-schemas` to copy them over `pkg/validation/schemas/v1/`, which is
   what `//go:embed` picks up. The two directories exist because `//go:embed` cannot
   reach outside its own package; `TestSchemaCopiesAreIdentical` fails if they drift
3. Update Go types in `pkg/types/pca.go` to match
4. Run validation tests: `go test ./pkg/validation`
5. Update this documentation if needed

A schema change now has teeth: tightening a constraint will reject models that the
previous release wrote. Check the change against real output before committing it —
`pca analyze -f json -o out/ testdata/iris/iris.csv` followed by `pca transform` is
enough to catch the common case.

## References

- [JSON Schema Specification](https://json-schema.org/draft/2020-12/json-schema-core.html)
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/)
- GoPCA Types: `pkg/types/pca.go`