// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package transform implements data transformations for tabular CSV data.
//
// All functions operate on [][]string data matrices with associated metadata
// (headers, column types). This design keeps the package free of any dependency
// on application-level types (such as FileData) so it can be shared across
// application entry points without circular imports.
//
// Supported transformations:
//   - Mathematical: log, sqrt, square (element-wise, numeric columns)
//   - Scaling: z-score standardization, min-max scaling (numeric columns)
//   - Discretization: equal-width binning (numeric → categorical)
//   - Encoding: one-hot encoding (categorical → multiple numeric columns)
//
// The primary entry points are:
//   - [Apply] — execute a transformation and return the modified data
//   - [GetTransformableColumns] — list columns eligible for a given transformation
package transform
