// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { previewRowIndices, axisPositions, MAX_PREVIEW_ROWS } from './previewRows';

describe('previewRowIndices', () => {
    it('returns everything when the dataset is smaller than the cap', () => {
        expect(previewRowIndices(5, 30)).toEqual([0, 1, 2, 3, 4]);
    });

    it('always includes the first and last row', () => {
        // Spectral files are often ordered by batch or product, so the ends of
        // the dataset are frequently the most different from each other.
        const picked = previewRowIndices(881, 30);
        expect(picked[0]).toBe(0);
        expect(picked[picked.length - 1]).toBe(880);
    });

    it('never exceeds the cap', () => {
        expect(previewRowIndices(881, 30).length).toBeLessThanOrEqual(30);
        expect(previewRowIndices(1_000_000).length).toBeLessThanOrEqual(MAX_PREVIEW_ROWS);
    });

    it('spreads the choice across the dataset rather than taking a block', () => {
        // Taking the first thirty rows of an ordered file can mean thirty
        // samples of one material.
        const picked = previewRowIndices(900, 30);
        const spread = picked[picked.length - 1] - picked[0];
        expect(spread).toBeGreaterThan(800);
        const gaps = picked.slice(1).map((v, i) => v - picked[i]);
        expect(Math.max(...gaps) - Math.min(...gaps)).toBeLessThanOrEqual(1);
    });

    it('returns strictly increasing, unique indices', () => {
        for (const total of [7, 31, 80, 881, 5000]) {
            const picked = previewRowIndices(total, 30);
            expect(new Set(picked).size).toBe(picked.length);
            expect([...picked].sort((a, b) => a - b)).toEqual(picked);
            expect(picked.every(i => i >= 0 && i < total)).toBe(true);
        }
    });

    it('is deterministic, so the preview does not reshuffle as you type', () => {
        expect(previewRowIndices(881, 30)).toEqual(previewRowIndices(881, 30));
    });

    it('handles the degenerate sizes without throwing', () => {
        expect(previewRowIndices(0)).toEqual([]);
        expect(previewRowIndices(-3)).toEqual([]);
        expect(previewRowIndices(10, 0)).toEqual([]);
        expect(previewRowIndices(10, 1)).toEqual([0]);
    });
});

describe('axisPositions', () => {
    it('reads numeric column names as positions', () => {
        expect(axisPositions(['1100', '1102', '1104'])).toEqual([1100, 1102, 1104]);
    });

    it('accepts decimals and negatives', () => {
        expect(axisPositions(['-1.5', '0', '2.25'])).toEqual([-1.5, 0, 2.25]);
    });

    it('refuses a partially numeric axis rather than mixing two meanings', () => {
        // Plotting some points at wavelengths and others at ordinals would put
        // curves at positions that mean different things.
        expect(axisPositions(['1100', 'ratio', '1104'])).toBeNull();
        expect(axisPositions(['1100', '', '1104'])).toBeNull();
    });

    it('rejects non-finite values', () => {
        expect(axisPositions(['1100', 'Infinity', '1104'])).toBeNull();
        expect(axisPositions(['1100', 'NaN'])).toBeNull();
    });

    it('returns null for no columns', () => {
        expect(axisPositions([])).toBeNull();
    });
});
