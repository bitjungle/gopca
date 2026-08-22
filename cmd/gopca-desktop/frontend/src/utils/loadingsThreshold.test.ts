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

import { describe, it, expect } from 'vitest';
import { createLoadingsPlotConfig } from './plotlyDataTransform';

describe('loadings plot threshold', () => {
    // Regression test for #772: the threshold was hardcoded at 0.3. Loadings are
    // unit-norm, so with p variables an equal contribution is 1/sqrt(p). For the
    // 700-wavelength Corn spectra that is 0.038 — the fixed 0.3 sat an order of
    // magnitude above every loading, and because it is drawn as a y-axis shape it
    // dragged the autorange with it and squashed the curve into a sliver.
    const threshold = (n?: number) =>
        createLoadingsPlotConfig('line', false, undefined, undefined, n, 50).thresholdValue;

    it('scales as 1/sqrt(p)', () => {
        expect(threshold(4)).toBeCloseTo(0.5, 6);
        expect(threshold(700)).toBeCloseTo(0.0378, 4);
    });

    it('stays near the previous constant for mid-sized datasets', () => {
        expect(threshold(13)).toBeGreaterThan(0.25);
        expect(threshold(13)).toBeLessThan(0.3);
    });

    it('puts the Corn threshold within reach of real spectral loadings', () => {
        const t = threshold(700) as number;
        expect(t).toBeLessThan(0.05);
        expect(t).toBeGreaterThan(0.0);
    });

    it('falls back to 0.3 when the variable count is unknown', () => {
        expect(threshold(undefined)).toBe(0.3);
        expect(threshold(0)).toBe(0.3);
    });
});
