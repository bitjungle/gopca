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
