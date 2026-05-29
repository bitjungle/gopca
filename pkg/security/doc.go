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
