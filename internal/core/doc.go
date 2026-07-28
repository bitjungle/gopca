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

// Package core implements the unified Principal Component Analysis (PCA) engine
// that serves all GoPCA applications (the pca CLI, GoPCA Desktop, and GoCSV).
//
// It provides a single, bulletproof implementation of the core algorithms behind
// the [github.com/bitjungle/gopca/pkg/types.PCAEngine] interface:
//   - Standard PCA via SVD and NIPALS
//   - Kernel PCA (RBF, polynomial, linear)
//   - Temporal (time-delay) PCA
//
// Alongside the decompositions it supplies preprocessing (mean-centering,
// scaling, SNV), missing-value handling, and post-fit diagnostics and metrics
// (scores, loadings, explained variance, Hotelling's T², Q-residuals,
// confidence ellipses, eigencorrelations).
//
// The package is deliberately free of any UI or application dependencies so it
// can be shared unchanged across every interface. Algorithms are validated
// against reference implementations (scikit-learn, R); see the package tests and
// docs/devel/validation-methodology.md.
package core
