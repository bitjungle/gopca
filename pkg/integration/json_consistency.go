// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// JSONFieldConsistency checks for consistency between Go struct JSON tags and TypeScript interfaces
type JSONFieldConsistency struct {
	GoFieldName  string
	JSONTag      string
	TSFieldName  string
	Inconsistent bool
}

// CheckJSONConsistency validates that JSON tags follow consistent naming conventions
func CheckJSONConsistency(v interface{}) ([]JSONFieldConsistency, error) {
	var results []JSONFieldConsistency

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		// Skip fields without JSON tags or with "-" tag
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract field name from JSON tag (remove omitempty, etc.)
		tagParts := strings.Split(jsonTag, ",")
		jsonFieldName := tagParts[0]

		// Check naming convention consistency
		// TypeScript typically uses camelCase, Go JSON tags should match
		result := JSONFieldConsistency{
			GoFieldName: field.Name,
			JSONTag:     jsonFieldName,
			TSFieldName: toCamelCase(field.Name),
		}

		// Check if JSON tag matches expected TypeScript field name
		if jsonFieldName != result.TSFieldName && jsonFieldName != toSnakeCase(field.Name) {
			result.Inconsistent = true
		}

		results = append(results, result)
	}

	return results, nil
}

// toCamelCase converts a Go field name to TypeScript camelCase
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle acronyms
	s = handleAcronyms(s)

	// Convert first letter to lowercase
	return strings.ToLower(s[:1]) + s[1:]
}

// toSnakeCase converts a Go field name to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// handleAcronyms handles common acronyms in field names
func handleAcronyms(s string) string {
	// Common acronyms to handle
	acronyms := map[string]string{
		"ID":   "Id",
		"URL":  "Url",
		"API":  "Api",
		"HTTP": "Http",
		"JSON": "Json",
		"CSV":  "Csv",
		"PCA":  "Pca",
		"SNV":  "Snv",
	}

	for old, new := range acronyms {
		if strings.HasPrefix(s, old) {
			s = new + s[len(old):]
		}
	}

	return s
}

// ValidateJSONMarshaling validates that a struct can be marshaled/unmarshaled without data loss
func ValidateJSONMarshaling(v interface{}) error {
	// Marshal to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// Create a new instance of the same type
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	newV := reflect.New(t).Interface()

	// Unmarshal back
	if err := json.Unmarshal(data, newV); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	// Compare the two values (basic comparison)
	// In production, you'd want more sophisticated comparison
	original := fmt.Sprintf("%+v", v)
	reconstructed := fmt.Sprintf("%+v", newV)

	if original != reconstructed {
		return fmt.Errorf("data loss detected during JSON marshaling")
	}

	return nil
}

// StandardizeJSONTags returns recommended JSON tags for consistent serialization
func StandardizeJSONTags(fieldName string, omitempty bool) string {
	tag := toCamelCase(fieldName)
	if omitempty {
		tag += ",omitempty"
	}
	return tag
}
