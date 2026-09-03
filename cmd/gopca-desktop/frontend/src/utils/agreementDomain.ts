// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

/** A shared axis range, applied identically to x and y. */
export interface AgreementDomain {
    x: [number, number];
    y: [number, number];
}

/**
 * sharedDomainFor computes one range covering every value, for a plot whose two
 * axes carry the same quantity in the same units.
 *
 * Predicted against measured is the case in hand. Left to scale independently the
 * axes get different ranges, the line of perfect agreement stops being a
 * diagonal, and how far a point sits from it stops being readable, which is the
 * only thing such a plot is for.
 *
 * Returns undefined when there is nothing to scale, leaving the chart to its own
 * autoscaling rather than inventing a range around no data.
 */
export function sharedDomainFor(
    values: number[],
    marginFraction = 0.05
): AgreementDomain | undefined {
    const finite = values.filter(v => Number.isFinite(v));
    if (finite.length === 0) {
        return undefined;
    }

    const low = Math.min(...finite);
    const high = Math.max(...finite);

    // A response that never varies has no spread to pad, and a fraction of zero
    // would collapse the range onto a single value that Plotly cannot draw. Fall
    // back to the magnitude of the value, and to one when even that is zero.
    const spread = high - low;
    const pad = (spread || Math.abs(high) || 1) * marginFraction;

    const range: [number, number] = [low - pad, high + pad];
    return { x: range, y: range };
}
