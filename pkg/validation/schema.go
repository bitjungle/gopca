// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package validation provides JSON schema validation for PCA models
package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schemas/v1/*.json
var schemaFS embed.FS

// ModelValidator validates PCA model JSON data against schemas
type ModelValidator struct {
	mainSchema   string
	commonSchema string
	version      string
}

// NewModelValidator creates a new validator for the specified schema version
func NewModelValidator(version string) (*ModelValidator, error) {
	if version == "" {
		version = "v1"
	}

	// Load the main PCA output schema
	mainSchemaPath := fmt.Sprintf("schemas/%s/pca-output.schema.json", version)
	mainSchemaData, err := schemaFS.ReadFile(mainSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load main schema: %w", err)
	}

	// Load common definitions
	commonSchemaPath := fmt.Sprintf("schemas/%s/common.schema.json", version)
	commonSchemaData, err := schemaFS.ReadFile(commonSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load common schema: %w", err)
	}

	return &ModelValidator{
		mainSchema:   string(mainSchemaData),
		commonSchema: string(commonSchemaData),
		version:      version,
	}, nil
}

// ValidateModel validates PCA model JSON data against the schema
func (v *ModelValidator) ValidateModel(data []byte) error {
	// Parse JSON to check basic validity
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// For now, perform basic structural validation
	// Full schema validation would require resolving all $ref references
	var model map[string]interface{}
	if err := json.Unmarshal(data, &model); err != nil {
		return fmt.Errorf("failed to parse model: %w", err)
	}

	// Check for optional $schema field and validate if present
	if schema, ok := model["$schema"].(string); ok {
		// Validate that it points to a known schema version
		validSchemas := []string{
			"https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
			"../schemas/v1/pca-output.schema.json",
			"./schemas/v1/pca-output.schema.json",
		}
		schemaValid := false
		for _, valid := range validSchemas {
			if strings.HasSuffix(schema, valid) || schema == valid {
				schemaValid = true
				break
			}
		}
		if !schemaValid {
			return fmt.Errorf("unknown schema version: %s", schema)
		}
	}

	// Check required top-level fields
	requiredFields := []string{"metadata", "preprocessing", "model", "results"}
	for _, field := range requiredFields {
		if _, ok := model[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate metadata structure
	if err := v.validateMetadata(model["metadata"]); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Validate preprocessing structure
	if err := v.validatePreprocessing(model["preprocessing"]); err != nil {
		return fmt.Errorf("preprocessing validation failed: %w", err)
	}

	// Validate model components
	if err := v.validateModelComponents(model["model"]); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	// Validate results
	if err := v.validateResults(model["results"]); err != nil {
		return fmt.Errorf("results validation failed: %w", err)
	}

	return nil
}

// validateMetadata validates the metadata structure
func (v *ModelValidator) validateMetadata(data interface{}) error {
	metadata, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata must be an object")
	}

	// Check required fields
	requiredFields := []string{"analysis_id", "software_version", "created_at", "software", "config"}
	for _, field := range requiredFields {
		if _, ok := metadata[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate software field
	if software, ok := metadata["software"].(string); !ok || software != "gopca" {
		return fmt.Errorf("software must be 'gopca'")
	}

	// Validate config
	if config, ok := metadata["config"].(map[string]interface{}); ok {
		if _, ok := config["method"]; !ok {
			return fmt.Errorf("config.method is required")
		}
		if _, ok := config["n_components"]; !ok {
			return fmt.Errorf("config.n_components is required")
		}
	} else {
		return fmt.Errorf("config must be an object")
	}

	return nil
}

// validatePreprocessing validates the preprocessing structure
func (v *ModelValidator) validatePreprocessing(data interface{}) error {
	preprocessing, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("preprocessing must be an object")
	}

	// Check required boolean fields
	boolFields := []string{"mean_center", "standard_scale", "robust_scale", "scale_only", "snv", "vector_norm"}
	for _, field := range boolFields {
		if val, ok := preprocessing[field]; ok {
			if _, isBool := val.(bool); !isBool {
				return fmt.Errorf("%s must be a boolean", field)
			}
		} else {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Check parameters object exists
	if _, ok := preprocessing["parameters"]; !ok {
		return fmt.Errorf("missing required field: parameters")
	}

	return nil
}

// validateModelComponents validates the model components structure
func (v *ModelValidator) validateModelComponents(data interface{}) error {
	model, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("model must be an object")
	}

	// Check required fields
	requiredFields := []string{"loadings", "explained_variance", "explained_variance_ratio",
		"cumulative_variance", "component_labels", "feature_labels"}
	for _, field := range requiredFields {
		if _, ok := model[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate loadings is a 2D array
	if loadings, ok := model["loadings"].([]interface{}); ok {
		if len(loadings) > 0 {
			if _, ok := loadings[0].([]interface{}); !ok {
				return fmt.Errorf("loadings must be a 2D array")
			}
		}
	} else {
		return fmt.Errorf("loadings must be an array")
	}

	return nil
}

// validateResults validates the results structure
func (v *ModelValidator) validateResults(data interface{}) error {
	results, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("results must be an object")
	}

	// Check samples field exists
	samples, ok := results["samples"]
	if !ok {
		return fmt.Errorf("missing required field: samples")
	}

	// Validate samples structure
	samplesMap, ok := samples.(map[string]interface{})
	if !ok {
		return fmt.Errorf("samples must be an object")
	}

	// Check required fields in samples
	requiredFields := []string{"names", "scores"}
	for _, field := range requiredFields {
		if _, ok := samplesMap[field]; !ok {
			return fmt.Errorf("samples.%s is required", field)
		}
	}

	// Validate scores is a 2D array
	if scores, ok := samplesMap["scores"].([]interface{}); ok {
		if len(scores) > 0 {
			if _, ok := scores[0].([]interface{}); !ok {
				return fmt.Errorf("scores must be a 2D array")
			}
		}
	} else {
		return fmt.Errorf("scores must be an array")
	}

	return nil
}

// ValidateWithSchema validates JSON data against a specific schema file
// This is a simplified version for basic validation
func ValidateWithSchema(data []byte, schemaName string, version string) error {
	if version == "" {
		version = "v1"
	}

	// For now, just ensure valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// formatValidationErrors formats validation errors into a readable message
func formatValidationErrors(errors []gojsonschema.ResultError) error {
	if len(errors) == 0 {
		return nil
	}

	var msgs []string
	for _, err := range errors {
		// Format the error message with field context
		field := err.Field()
		if field == "(root)" {
			field = "model"
		}
		msgs = append(msgs, fmt.Sprintf("  - %s: %s", field, err.Description()))
	}

	return fmt.Errorf("validation failed:\n%s", strings.Join(msgs, "\n"))
}
