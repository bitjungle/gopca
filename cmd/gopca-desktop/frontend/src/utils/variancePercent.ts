// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import type { PCAResponse } from '../types';

/**
 * Converts the variance fields of a PCA response from fractions to percentages.
 *
 * As of V2 the engine reports `explained_variance_ratio` and
 * `cumulative_variance` as fractions of 1, matching scikit-learn and matching
 * what the field names have always said (#848). Before V2 they were
 * percentages, and the name invited anyone comparing against scikit-learn to be
 * silently out by a factor of 100.
 *
 * The user interface works in percentages: every label it renders ends in a
 * `%`, and the plots are axed accordingly. So the conversion happens exactly
 * once, here, where the response crosses from the backend into the frontend.
 *
 * Doing it at the boundary rather than at each point of use is deliberate.
 * There are close to thirty places that read these two fields, and multiplying
 * at each of them means thirty chances to miss one -- and a missed one renders
 * as "0.7%" beside correct neighbours, which is the kind of wrong that gets
 * believed. One conversion is one place to be right, and one place to look when
 * a number seems off by two orders of magnitude.
 */
export function asPercentages(response: PCAResponse): PCAResponse {
    const result = response?.result;
    if (!result) {
        return response;
    }

    const scale = (values: number[] | undefined): number[] | undefined =>
        values ? values.map(v => v * 100) : values;

    return {
        ...response,
        result: {
            ...result,
            explained_variance_ratio: scale(result.explained_variance_ratio)!,
            cumulative_variance: scale(result.cumulative_variance)!
        }
    };
}

/**
 * The inverse of {@link asPercentages}, for the export path.
 *
 * The exported model file is written by the same Go code the CLI uses, from the
 * result the frontend hands back. Since the frontend holds percentages for
 * display, they have to become fractions again on the way out, or the Desktop
 * would write models that disagree with the CLI's about what these two fields
 * mean -- the exact confusion V2 exists to remove.
 *
 * The round trip is not bit-exact: (x * 100) / 100 can differ from x by an ulp.
 * That is acceptable here and worth stating rather than glossing. These fields
 * are reported, never recomputed from -- nothing in the codebase reads
 * explained_variance_ratio back out of a model file -- so a relative error of
 * 1e-16 has no consequence beyond the sixteenth decimal of a displayed figure.
 */
export function asFractions<T extends { explained_variance_ratio: number[]; cumulative_variance: number[] }>(
    result: T
): T {
    return {
        ...result,
        explained_variance_ratio: result.explained_variance_ratio.map(v => v / 100),
        cumulative_variance: result.cumulative_variance.map(v => v / 100)
    };
}
