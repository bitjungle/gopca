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
