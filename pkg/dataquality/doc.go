// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package dataquality provides data quality analysis and missing value handling
// for tabular datasets, independent of any UI or Wails framework.
//
// The package operates on plain [][]string data matrices and column metadata
// maps to avoid circular dependencies with UI-layer packages. Callers convert
// their domain types (e.g. FileData) into the AnalysisInput struct before
// calling package functions.
//
// Key capabilities:
//   - Missing value detection, statistics, and fill strategies (mean, median,
//     mode, forward-fill, backward-fill, custom value)
//   - Per-column statistics (mean, median, standard deviation, percentiles,
//     skewness, kurtosis, categorical frequency)
//   - Distribution analysis and histogram generation
//   - Outlier detection via IQR and Z-score methods
//   - Pairwise Pearson correlation matrix
//   - Data quality scoring and actionable issue/recommendation generation
package dataquality
