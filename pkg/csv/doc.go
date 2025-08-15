// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package csv provides comprehensive CSV file handling for the GoPCA toolkit.
// It includes parsing, validation, and writing functionality with built-in
// security measures and support for various CSV formats.
//
// # Features
//
// The package supports:
//   - Multiple delimiters (comma, semicolon, tab)
//   - Different decimal separators (period, comma)
//   - Automatic column type detection
//   - Missing value handling
//   - Large file streaming
//   - Security validation against malicious inputs
//
// # Security
//
// All file operations include security validations:
//   - Path traversal prevention
//   - File size limits (500MB default)
//   - Field length limits (10,000 characters)
//   - Row and column count limits
//
// # Parse Modes
//
// The package supports four parsing modes:
//   - ParseNumeric: All data as floating-point numbers (for PCA)
//   - ParseString: All data as strings (for editing)
//   - ParseMixed: Automatic type detection
//   - ParseMixedWithTargets: Type detection with target column identification
//
// # Usage
//
// Basic usage:
//
//	opts := csv.DefaultOptions()
//	data, err := csv.ParseFile("data.csv", opts)
//
// European format:
//
//	opts := csv.EuropeanOptions()
//	data, err := csv.ParseFile("data.csv", opts)
//
// # Performance
//
// The package is optimized for both small and large datasets.
// Streaming mode is available for files that exceed memory constraints.
package csv
