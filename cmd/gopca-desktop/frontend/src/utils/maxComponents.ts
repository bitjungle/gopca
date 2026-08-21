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

// Upper bound on the number of principal components a given PCA method can
// produce. These mirror the validation performed by the Go engine; keeping the
// spinner in step with the backend avoids offering the user a value the
// analysis will reject, and — more importantly — avoids hiding values the
// method genuinely supports.
//
// Backend references:
//   - Kernel PCA:   internal/core/kernel_pca.go   (Components > nSamples -> error)
//   - Temporal PCA: internal/core/temporal_pca.go (min(n - lags + 1, m * lags))

/**
 * Maximum number of components the given method can extract.
 *
 * Kernel PCA eigendecomposes the n×n kernel matrix rather than the p×p
 * covariance matrix, so it is bounded by the *sample* count, not the variable
 * count — it can legitimately return far more components than there are
 * variables. Temporal PCA (SSA) decomposes the trajectory matrix
 * [(n − L + 1) × (m · L)] and is bounded by the smaller of those dimensions.
 * SVD and NIPALS are bounded by min(variables, samples).
 *
 * @param method        PCA method ('SVD' | 'NIPALS' | 'kernel' | 'temporal')
 * @param nVariables    number of numeric analysis columns (target columns excluded)
 * @param nSamples      number of rows
 * @param temporalLags  window length L, only used when method is 'temporal'
 * @returns the maximum valid component count, never below 1
 */
export function maxComponentsFor(
    method: string,
    nVariables: number,
    nSamples: number,
    temporalLags?: number
): number {
    let max: number;

    switch (method) {
        case 'kernel':
            max = nSamples;
            break;
        case 'temporal': {
            const lags = temporalLags && temporalLags > 0 ? temporalLags : 1;
            max = Math.min(nSamples - lags + 1, nVariables * lags);
            break;
        }
        default:
            max = Math.min(nVariables, nSamples);
            break;
    }

    return Math.max(1, max);
}

/**
 * Clamp a component count typed into the spinner to a valid range.
 *
 * The `max` attribute on a number input constrains only the stepper arrows —
 * a typed value passes straight through — so the value has to be clamped here
 * or the backend rejects the analysis. Clearing the field yields NaN, in which
 * case the previous value is kept rather than substituting a constant that may
 * itself exceed the ceiling (e.g. defaulting to 5 on the 4-variable Iris set).
 *
 * @param raw       the raw input string
 * @param previous  the current component count, retained when input is empty
 * @param max       ceiling from maxComponentsFor() for the active method
 * @returns a component count within [1, max]
 */
export function clampComponentCount(raw: string, previous: number, max: number): number {
    const parsed = parseInt(raw, 10);
    if (!Number.isFinite(parsed)) {
        return previous;
    }
    return Math.min(Math.max(1, parsed), max);
}
