// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package security provides security utilities for input validation,
// path sanitization, and protection against common vulnerabilities.
package security

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Limits for various input types to prevent resource exhaustion
const (
	MaxFileSize         = 500 * 1024 * 1024 // 500MB max file size
	MaxCSVRows          = 1000000           // 1M rows max
	MaxCSVColumns       = 10000             // 10K columns max
	MaxFieldLength      = 100000            // 100K chars per field
	MaxStringLength     = 10000             // 10K chars for general strings
	MaxPathLength       = 4096              // Standard PATH_MAX
	MaxComponents       = 1000              // Max PCA components
	MinComponents       = 1                 // Min PCA components
	MaxKernelPCASamples = 10000             // Max samples for Kernel PCA (memory safety)
	MaxKernelGamma      = 1e6               // Max kernel gamma value
	MinKernelGamma      = 1e-6              // Min kernel gamma value
	MaxIterations       = 10000             // Max iterations for algorithms
	MaxMemoryUsageMB    = 2048              // 2GB max memory for operations
)

// ValidateNumericInput validates and sanitizes numeric input within bounds
func ValidateNumericInput(input string, min, max float64, paramName string) (float64, error) {
	// Remove whitespace
	input = strings.TrimSpace(input)

	// Check for empty input
	if input == "" {
		return 0, fmt.Errorf("%s: empty input", paramName)
	}

	// Check for invalid characters (prevent injection)
	for _, r := range input {
		if !unicode.IsDigit(r) && r != '.' && r != '-' && r != '+' && r != 'e' && r != 'E' {
			return 0, fmt.Errorf("%s: invalid character '%c' in numeric input", paramName, r)
		}
	}

	// Parse the number
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid numeric value: %w", paramName, err)
	}

	// Check for special values
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("%s: invalid numeric value (NaN or Inf)", paramName)
	}

	// Validate bounds
	if value < min || value > max {
		return 0, fmt.Errorf("%s: value %.6f out of range [%.6f, %.6f]", paramName, value, min, max)
	}

	return value, nil
}

// ValidateIntegerInput validates integer input within bounds
func ValidateIntegerInput(input string, min, max int, paramName string) (int, error) {
	// Remove whitespace
	input = strings.TrimSpace(input)

	// Check for empty input
	if input == "" {
		return 0, fmt.Errorf("%s: empty input", paramName)
	}

	// Check for invalid characters
	for i, r := range input {
		if i == 0 && (r == '-' || r == '+') {
			continue
		}
		if !unicode.IsDigit(r) {
			return 0, fmt.Errorf("%s: invalid character '%c' in integer input", paramName, r)
		}
	}

	// Parse the integer
	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer value: %w", paramName, err)
	}

	// Validate bounds
	if value < min || value > max {
		return 0, fmt.Errorf("%s: value %d out of range [%d, %d]", paramName, value, min, max)
	}

	return value, nil
}

// ValidateStringInput validates and sanitizes string input
func ValidateStringInput(input string, maxLength int, allowedChars string, paramName string) (string, error) {
	// Check UTF-8 validity
	if !utf8.ValidString(input) {
		return "", fmt.Errorf("%s: invalid UTF-8 encoding", paramName)
	}

	// Check length
	if len(input) > maxLength {
		return "", fmt.Errorf("%s: string too long (%d > %d)", paramName, len(input), maxLength)
	}

	// Remove null bytes and control characters
	cleaned := strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\t' && r != '\n' && r != '\r') {
			return -1 // Remove character
		}
		return r
	}, input)

	// Check allowed characters if specified
	if allowedChars != "" {
		for _, r := range cleaned {
			if !strings.ContainsRune(allowedChars, r) {
				return "", fmt.Errorf("%s: contains disallowed character '%c'", paramName, r)
			}
		}
	}

	return cleaned, nil
}

// ValidateComponentCount validates PCA component count
func ValidateComponentCount(components, maxFeatures int) error {
	if components < MinComponents {
		return fmt.Errorf("components must be at least %d", MinComponents)
	}

	if components > MaxComponents {
		return fmt.Errorf("components cannot exceed %d", MaxComponents)
	}

	if components > maxFeatures {
		return fmt.Errorf("components (%d) cannot exceed number of features (%d)", components, maxFeatures)
	}

	return nil
}

// ValidateKernelParameters validates kernel PCA parameters
func ValidateKernelParameters(kernelType string, gamma, degree float64, coef0 float64) error {
	validKernels := map[string]bool{
		"rbf":        true,
		"polynomial": true,
		"sigmoid":    true,
		"linear":     true,
	}

	if !validKernels[strings.ToLower(kernelType)] {
		return fmt.Errorf("invalid kernel type: %s", kernelType)
	}

	// Validate gamma for RBF and polynomial kernels
	if kernelType == "rbf" || kernelType == "polynomial" || kernelType == "sigmoid" {
		if gamma < MinKernelGamma || gamma > MaxKernelGamma {
			return fmt.Errorf("gamma %.6f out of range [%.6f, %.6f]", gamma, MinKernelGamma, MaxKernelGamma)
		}
	}

	// Validate degree for polynomial kernel
	if kernelType == "polynomial" {
		if degree < 1 || degree > 10 {
			return fmt.Errorf("polynomial degree %.0f out of range [1, 10]", degree)
		}
	}

	// Validate coef0 for polynomial and sigmoid kernels
	if kernelType == "polynomial" || kernelType == "sigmoid" {
		if math.Abs(coef0) > 1000 {
			return fmt.Errorf("coef0 %.2f out of range [-1000, 1000]", coef0)
		}
	}

	return nil
}

// ValidateDataDimensions validates data matrix dimensions
func ValidateDataDimensions(rows, cols int) error {
	if rows <= 0 || cols <= 0 {
		return fmt.Errorf("invalid dimensions: rows=%d, cols=%d", rows, cols)
	}

	if rows > MaxCSVRows {
		return fmt.Errorf("too many rows: %d (max %d)", rows, MaxCSVRows)
	}

	if cols > MaxCSVColumns {
		return fmt.Errorf("too many columns: %d (max %d)", cols, MaxCSVColumns)
	}

	// Check for potential memory issues
	estimatedMemoryMB := (rows * cols * 8) / (1024 * 1024) // Assuming 8 bytes per float64
	if estimatedMemoryMB > MaxMemoryUsageMB {
		return fmt.Errorf("dataset too large: estimated %dMB exceeds limit of %dMB",
			estimatedMemoryMB, MaxMemoryUsageMB)
	}

	return nil
}

// SanitizeFilename removes potentially dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	dangerous := []string{"/", "\\", "..", "~", "|", ">", "<", "&", "$", "`", ";", ":", "*", "?", "\"", "'"}

	result := filename
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Remove leading dots (hidden files)
	result = strings.TrimLeft(result, ".")

	// Limit length
	if len(result) > 255 {
		result = result[:255]
	}

	// Ensure non-empty
	if result == "" {
		result = "unnamed"
	}

	return result
}

// ValidateCSVDelimiter validates CSV delimiter character
func ValidateCSVDelimiter(delimiter string) (rune, error) {
	if len(delimiter) != 1 {
		return 0, fmt.Errorf("delimiter must be a single character")
	}

	r := rune(delimiter[0])

	// Allow common delimiters
	validDelimiters := []rune{',', ';', '\t', '|', ' '}
	valid := false
	for _, d := range validDelimiters {
		if r == d {
			valid = true
			break
		}
	}

	if !valid {
		return 0, fmt.Errorf("invalid delimiter: '%c'", r)
	}

	return r, nil
}

// IsValidEmail performs basic email validation
func IsValidEmail(email string) bool {
	// Basic validation - not meant to be exhaustive
	if len(email) > 254 { // RFC 5321
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	// Check for basic domain structure
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}
