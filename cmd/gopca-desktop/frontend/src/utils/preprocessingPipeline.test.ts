// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { preprocessingPipeline, PreprocessingSummaryConfig } from './preprocessingPipeline';

const config = (overrides: Partial<PreprocessingSummaryConfig> = {}): PreprocessingSummaryConfig => ({
    snv: false,
    vectorNorm: false,
    savgolWindow: 0,
    savgolPolyOrder: 2,
    savgolDeriv: 0,
    meanCenter: false,
    standardScale: false,
    robustScale: false,
    scaleOnly: false,
    ...overrides
});

describe('preprocessingPipeline', () => {
    it('says nothing when nothing is selected', () => {
        // Raw data into the decomposition is a legitimate choice; the caller
        // decides how to present an empty pipeline.
        expect(preprocessingPipeline(config())).toEqual([]);
    });

    it('lists the engine order, not the order the fields happen to be read in', () => {
        expect(preprocessingPipeline(config({
            snv: true,
            savgolWindow: 11, savgolPolyOrder: 2, savgolDeriv: 1,
            meanCenter: true, standardScale: true
        }))).toEqual([
            'SNV',
            'Savitzky-Golay 1st derivative (window 11, order 2)',
            'Standard scale'
        ]);
    });

    it('reports each step on its own', () => {
        expect(preprocessingPipeline(config({ snv: true }))).toEqual(['SNV']);
        expect(preprocessingPipeline(config({ vectorNorm: true }))).toEqual(['L2 normalization']);
        expect(preprocessingPipeline(config({ meanCenter: true }))).toEqual(['Mean center']);
        expect(preprocessingPipeline(config({ savgolWindow: 9, savgolPolyOrder: 3, savgolDeriv: 0 })))
            .toEqual(['Savitzky-Golay smoothing (window 9, order 3)']);
    });

    it('names each derivative order', () => {
        for (const [deriv, label] of [[0, 'smoothing'], [1, '1st derivative'], [2, '2nd derivative']] as const) {
            expect(preprocessingPipeline(config({ savgolWindow: 11, savgolDeriv: deriv }))[0])
                .toContain(label);
        }
    });

    it('omits the filter entirely when the window is zero', () => {
        // The order and derivative keep their values while the filter is off,
        // so reading them alone would announce a step that will not happen.
        expect(preprocessingPipeline(config({ savgolWindow: 0, savgolDeriv: 2 }))).toEqual([]);
    });

    it('follows the same precedence as the column-wise selector', () => {
        // The selector shows one value; the summary must not claim two, nor a
        // different one. scale-only wins, then robust, then standard, then
        // centre -- which is the order the control itself resolves them in.
        const all = config({ meanCenter: true, standardScale: true, robustScale: true, scaleOnly: true });
        expect(preprocessingPipeline(all)).toEqual(['Variance scale']);
        expect(preprocessingPipeline({ ...all, scaleOnly: false })).toEqual(['Robust scale']);
        expect(preprocessingPipeline({ ...all, scaleOnly: false, robustScale: false })).toEqual(['Standard scale']);
    });

    it('prefers SNV when both row-wise flags are somehow set', () => {
        // They are mutually exclusive in the panel, but the request type allows
        // both, and announcing two row-wise steps would describe a pipeline the
        // engine never runs.
        expect(preprocessingPipeline(config({ snv: true, vectorNorm: true }))).toEqual(['SNV']);
    });
});
