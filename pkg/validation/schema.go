// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

// Package validation provides JSON schema validation for PCA models
package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
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

	// The regression block is optional: a model without it is a plain
	// decomposition, which is what every model produced before principal
	// component regression existed looks like.
	if regression, present := model["regression"]; present {
		if err := v.validateRegression(regression); err != nil {
			return fmt.Errorf("regression validation failed: %w", err)
		}
	}

	return nil
}

// validateRegression checks the regression half of a principal component
// regression model.
//
// The checks that matter are the ones a consumer would otherwise discover by
// producing wrong numbers: a component count that disagrees with the coefficients
// it indexes, and an original-scale form that claims to be usable while missing
// the coefficients it needs. Both would predict silently and incorrectly.
func (v *ModelValidator) validateRegression(data interface{}) error {
	regression, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("regression must be an object")
	}

	for _, field := range []string{"response", "components", "score_coefficients", "intercept"} {
		if _, present := regression[field]; !present {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	if _, ok := regression["response"].(string); !ok {
		return fmt.Errorf("response must be a string")
	}

	components, ok := regression["components"].(float64)
	if !ok {
		return fmt.Errorf("components must be a number")
	}
	if components < 0 {
		return fmt.Errorf("components must not be negative, got %v", components)
	}

	coefficients, ok := regression["score_coefficients"].([]interface{})
	if !ok {
		return fmt.Errorf("score_coefficients must be an array")
	}
	if len(coefficients) != int(components) {
		return fmt.Errorf(
			"the model declares %d components but carries %d score coefficients: "+
				"predicting from it would read the wrong number of directions",
			int(components), len(coefficients))
	}

	if _, ok := regression["intercept"].(float64); !ok {
		return fmt.Errorf("intercept must be a number")
	}

	// original_scale_valid is a promise about the fields beside it. If it claims
	// the collapsed form is usable, the coefficients must actually be there.
	if valid, present := regression["original_scale_valid"]; present {
		claimed, ok := valid.(bool)
		if !ok {
			return fmt.Errorf("original_scale_valid must be a boolean")
		}
		if claimed {
			original, present := regression["coefficients"]
			if !present {
				return fmt.Errorf(
					"original_scale_valid is true but coefficients are absent: " +
						"the model claims a collapsed form it does not carry")
			}
			if _, ok := original.([]interface{}); !ok {
				return fmt.Errorf("coefficients must be an array")
			}
		}
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
	// Note: version parameter is reserved for future schema versioning
	// Currently only v1 schemas are supported
	_ = version // Mark as intentionally unused

	// For now, just ensure valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}
