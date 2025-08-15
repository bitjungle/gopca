// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

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
