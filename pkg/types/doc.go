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

// Package types provides the core data structures and interfaces for the GoPCA toolkit.
// It defines the fundamental types used throughout the application for PCA analysis,
// data representation, and configuration.
//
// # Core Types
//
// The package defines several essential types:
//
//   - Matrix: 2D slice representation of numerical data
//   - PCAConfig: Configuration for PCA analysis including method selection and preprocessing
//   - PCAResult: Results from PCA analysis including scores, loadings, and variance metrics
//   - PCAEngine: Interface for different PCA algorithm implementations
//
// # Data Structures
//
// Matrix operations use row-major order where data[i][j] represents row i, column j.
// This aligns with standard CSV file structure and mathematical notation.
//
// # Configuration
//
// PCAConfig supports multiple PCA methods:
//   - SVD: Singular Value Decomposition (default, fast for complete data)
//   - NIPALS: Nonlinear Iterative Partial Least Squares (handles missing data)
//   - Kernel PCA: For non-linear relationships
//
// # Error Handling
//
// The package provides structured error types for consistent error handling
// across the application. All errors include context for debugging.
//
// # Thread Safety
//
// Types in this package are not thread-safe. Concurrent access to PCAEngine
// instances should be synchronized by the caller.
package types
