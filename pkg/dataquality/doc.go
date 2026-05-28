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
