// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import {
    preprocessingPipeline,
    rowStagePipeline,
    columnStagePipeline,
    shouldShowPreview,
    PreprocessingSummaryConfig
} from './preprocessingPipeline';

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

describe('shouldShowPreview', () => {
    it('shows for row-wise normalization on its own', () => {
        // Reported from the running app: the preview appeared to need
        // Savitzky-Golay. Seeing what SNV did to the spectra is as reasonable a
        // question as seeing what a derivative did.
        expect(shouldShowPreview(true, config({ snv: true }))).toBe(true);
        expect(shouldShowPreview(true, config({ vectorNorm: true }))).toBe(true);
    });

    it('shows for Savitzky-Golay on its own', () => {
        expect(shouldShowPreview(true, config({ savgolWindow: 11 }))).toBe(true);
    });

    it('shows for both together', () => {
        expect(shouldShowPreview(true, config({ snv: true, savgolWindow: 11 }))).toBe(true);
    });

    it('hides when nothing happens to the data', () => {
        // A "preprocessed" view identical to the raw one says nothing.
        expect(shouldShowPreview(true, config())).toBe(false);
    });

    it('hides when the variables are not an axis worth drawing along', () => {
        // A line across unordered variables is a shape with no meaning.
        expect(shouldShowPreview(false, config({ snv: true, savgolWindow: 11 }))).toBe(false);
    });

    it('hides for Temporal PCA, which applies no row stage at all', () => {
        // Previewing SNV there would show a transformation that will not happen:
        // temporal PCA builds its own preprocessor and drops the row-wise
        // settings (#889).
        expect(shouldShowPreview(true, config({ snv: true }), 'temporal')).toBe(false);
        expect(shouldShowPreview(true, config({ savgolWindow: 11 }), 'Temporal')).toBe(false);
    });

    it('shows for every other method', () => {
        for (const method of ['SVD', 'NIPALS', 'kernel', undefined]) {
            expect(shouldShowPreview(true, config({ snv: true }), method)).toBe(true);
        }
    });

    it('ignores column-wise settings, which the preview does not apply', () => {
        expect(shouldShowPreview(true, config({ meanCenter: true, standardScale: true }))).toBe(false);
    });
});

describe('rowStagePipeline and columnStagePipeline', () => {
    it('split the pipeline exactly where the preview stops', () => {
        // The preview draws the row stage and not the column stage, so a caption
        // built from the full pipeline would name a step the curves never went
        // through.
        const full = config({
            snv: true,
            savgolWindow: 11, savgolPolyOrder: 2, savgolDeriv: 1,
            meanCenter: true, standardScale: true
        });
        expect(rowStagePipeline(full)).toEqual([
            'SNV',
            'Savitzky-Golay 1st derivative (window 11, order 2)'
        ]);
        expect(columnStagePipeline(full)).toEqual(['Standard scale']);
    });

    it('together they reproduce the full pipeline, in order', () => {
        // The summary line and the plot caption must not be able to disagree
        // about what happens or in what sequence.
        for (const c of [
            config(),
            config({ snv: true }),
            config({ meanCenter: true }),
            config({ vectorNorm: true, savgolWindow: 9, savgolDeriv: 2, robustScale: true }),
            config({ savgolWindow: 15, savgolPolyOrder: 3, scaleOnly: true })
        ]) {
            expect([...rowStagePipeline(c), ...columnStagePipeline(c)]).toEqual(preprocessingPipeline(c));
        }
    });

    it('reports no column stage when none is selected', () => {
        expect(columnStagePipeline(config({ snv: true }))).toEqual([]);
    });
});
