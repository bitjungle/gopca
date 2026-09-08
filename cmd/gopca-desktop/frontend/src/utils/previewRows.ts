// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

/** How many spectra the preview draws at most. */
export const MAX_PREVIEW_ROWS = 30;

/**
 * Chooses which rows to draw.
 *
 * Spread evenly across the dataset rather than taking the first N. Spectral
 * files are very often ordered — by batch, by product, by the day they were
 * measured — so the first thirty rows can be thirty samples of one material,
 * and a preview drawn from them would show a narrower spread of shapes than the
 * data actually contains. That is the wrong impression to give someone judging
 * whether a filter is behaving.
 *
 * The first and last rows are always included, so the extremes of whatever
 * ordering exists are visible. The choice is deterministic: a preview that
 * reshuffled itself on every keystroke would make it impossible to see what
 * changing the window actually did.
 */
export function previewRowIndices(total: number, max: number = MAX_PREVIEW_ROWS): number[] {
    if (total <= 0 || max <= 0) {
        return [];
    }
    if (total <= max) {
        return Array.from({ length: total }, (_, i) => i);
    }
    if (max === 1) {
        return [0];
    }

    const indices: number[] = [];
    for (let k = 0; k < max; k++) {
        // Spans 0..total-1 inclusive, so both ends are always drawn.
        const index = Math.round((k * (total - 1)) / (max - 1));
        if (indices[indices.length - 1] !== index) {
            indices.push(index);
        }
    }
    return indices;
}

/**
 * Parses column names as positions on a measured axis.
 *
 * Returns null unless every name is a number, in which case the caller plots
 * against variable index instead. Partial parsing is deliberately not attempted:
 * an axis where some points are wavelengths and others are ordinals would place
 * curves at positions that mean two different things, which is worse than an
 * honest index.
 */
export function axisPositions(headers: string[]): number[] | null {
    const values: number[] = [];
    for (const header of headers) {
        const value = Number(header);
        if (header.trim() === '' || !Number.isFinite(value)) {
            return null;
        }
        values.push(value);
    }
    return values.length > 0 ? values : null;
}
