// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package security provides comprehensive security measures for the GoPCA toolkit.
// It implements defense-in-depth strategies to protect against common vulnerabilities
// including path traversal, command injection, and resource exhaustion attacks.
//
// # Input Validation
//
// The package provides validators for all user inputs:
//   - Numeric values with bounds checking
//   - String inputs with length and character restrictions
//   - File paths with traversal prevention
//   - Command arguments with injection prevention
//
// # Path Security
//
// File path operations include multiple layers of protection:
//   - Path traversal detection and prevention
//   - System directory write protection
//   - Jail/sandbox path enforcement
//   - Platform-specific validation (Windows reserved names, etc.)
//
// # Command Security
//
// External command execution is secured through:
//   - Command whitelisting
//   - Argument validation
//   - Special character escaping
//   - Environment variable sanitization
//
// # Resource Limits
//
// The package enforces limits to prevent resource exhaustion:
//   - Maximum file size: 500MB
//   - Maximum CSV rows: 1,000,000
//   - Maximum CSV columns: 10,000
//   - Maximum field length: 10,000 characters
//   - Maximum memory usage: 2GB for data matrices
//
// # Usage
//
// Input validation:
//
//	value, err := security.ValidateNumericInput(input, 0, 100, "parameter")
//
// Path validation:
//
//	err := security.ValidateInputPath(filePath)
//
// Command validation:
//
//	err := security.ValidateCommand(cmd, args)
//
// # Security Policy
//
// For vulnerability reporting and security policies, see SECURITY.md
// in the repository root.
package security
